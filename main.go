package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"path"
	"plugin"
	"syscall"

	_ "github.com/kraftwerk28/gost/blocks"
	"github.com/kraftwerk28/gost/core"
	"gopkg.in/yaml.v3"
)

const programName = "gost"
const configFileName = "config.yml"

func getConfigPath() string {
	xdg, hasXdg := os.LookupEnv("XDG_CONFIG_HOME")
	if !hasXdg {
		home, _ := os.UserHomeDir()
		xdg = path.Join(home, ".config")
	}
	return path.Join(xdg, programName, configFileName)
}

func main() {
	var cfgPath, logPath string
	flag.StringVar(&cfgPath, "config", "", "Path to config.yml")
	flag.StringVar(&logPath, "log", "", "Path to log file")
	flag.Parse()

	logOutput := os.Stderr
	if logPath != "" {
		var err error
		logOutput, err = os.OpenFile(
			logPath,
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			0644,
		)
		if err != nil {
			panic(err)
		}
	}
	core.InitializeLogger(logOutput)
	// Logger initialized at this point
	log := core.Log

	if cfgPath == "" {
		cfgPath = getConfigPath()
	}
	if _, err := os.Stat(cfgPath); err != nil {
		log.Fatal(err)
	}

	programCtx := context.Background()
	cancelCtx, _ := context.WithCancel(programCtx)
	// TODO: implement hot reloading by using context

	cfgFile, err := os.Open(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	defer cfgFile.Close()
	cfgDecoder := yaml.NewDecoder(cfgFile)
	cfg := &core.AppConfig{}
	if err := cfgDecoder.Decode(cfg); err != nil {
		log.Fatal(err)
	}

	managers := make([]*core.BlockletMgr, 0, len(cfg.Blocks))
	for _, c := range cfg.Blocks {
		var ctor core.I3barBlockletCtor
		if c.Name == "plugin" {
			handle, err := plugin.Open(c.Path)
			if err != nil {
				log.Println("Failed to load plugin:")
				log.Print(err)
			}
			sym, err := handle.Lookup("NewBlock")
			if err != nil {
				log.Println("Plugin must have `func NewBlock() I3barBlocklet`:")
				log.Print(err)
				continue
			}
			if c, ok := sym.(*core.I3barBlockletCtor); ok {
				ctor = *c
			} else {
				log.Println("Bad constructor")
				continue
			}
		} else if ct := core.GetBuiltin(c.Name); ct != nil {
			ctor = ct
		} else {
			log.Fatalf(`Unrecognized blocklet name: "%s"`, c.Name)
		}
		blocklet := ctor()
		if b, ok := blocklet.(core.I3barBlockletConfigurable); ok {
			cf, _ := yaml.Marshal(c)
			if err := yaml.Unmarshal(cf, b.GetConfig()); err != nil {
				log.Fatal(err)
			}
		}
		m := core.NewBlockletMgr(c.Name, blocklet, cancelCtx)
		managers = append(managers, m)
	}

	{
		header := core.I3barHeader{
			Version:     1,
			ClickEvents: true,
		}
		b, _ := json.Marshal(header)
		b = append(b, []byte("\n[\n")...)
		os.Stdout.Write(b)
	}

	listeners := make([]*core.BlockletMgr, 0, len(managers))
	updateChan := make(chan string)
	for _, m := range managers {
		m.Run(updateChan)
		if m.IsListener() {
			listeners = append(listeners, m)
		}
	}

	signalChan := make(chan os.Signal)
	signal.Notify(
		signalChan,
		syscall.SIGCONT, syscall.SIGSTOP,
		syscall.SIGUSR2, syscall.SIGTERM,
	)

	// Read events from stdin
	go func() {
		log := core.Log
		sc := bufio.NewScanner(os.Stdin)
		sc.Scan() // Strip `[`
		for sc.Scan() {
			raw := sc.Bytes()
			ev, err := core.NewEventFromRaw(raw)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("Click event: %+v\n", *ev)
			for _, m := range listeners {
				m.ProcessEvent(ev)
			}
		}
		if err := sc.Err(); err != nil {
			log.Fatal(err)
		}
	}()

loop:
	for {
		blocks := make([]core.I3barBlock, 0)
		for _, m := range managers {
			blocks = append(blocks, m.Render()...)
		}
		b, _ := json.Marshal(blocks)
		b = append(b, ',', '\n')
		os.Stdout.Write(b)
		select {
		case updateData := <-updateChan:
			for i := range managers {
				managers[i].TryInvalidate(updateData)
			}
		case signal := <-signalChan:
			switch signal {
			case syscall.SIGUSR2:
				// Reload bar
				break
			case syscall.SIGSTOP:
				// Stop emitting JSON
			case syscall.SIGCONT:
				// Continue emitting JSON
			case syscall.SIGTERM, syscall.SIGINT:
				break loop
			}
		}
	}
	core.Log.Println("\nBye.")
}
