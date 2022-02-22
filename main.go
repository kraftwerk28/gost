package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path"

	// "context"
	// "plugin"

	_ "github.com/kraftwerk28/gost/blocks"
	"github.com/kraftwerk28/gost/core"
	"gopkg.in/yaml.v2"
)

var log = core.Log

const PROGRAM_NAME = "gost"

func main() {
	var cfgPath, logPath string
	flag.StringVar(&cfgPath, "config", "", "Path to config.yml")
	flag.StringVar(&logPath, "log", "", "Path to log file")
	flag.Parse()

	logOutput := io.Discard
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

	if cfgPath == "" {
		xdg, hasXdg := os.LookupEnv("XDG_CONFIG_HOME")
		if !hasXdg {
			home, _ := os.UserHomeDir()
			xdg = path.Join(home, ".config")
		}
		cfgPath = path.Join(xdg, PROGRAM_NAME, "config.yml")
	}
	if _, err := os.Stat(cfgPath); err != nil {
		log.Fatal(errors.New("Could not load config"))
	}

	// programCtx := context.Background()
	// cancelCtx, cancelFunc := context.WithCancel(programCtx)
	// println(cancelCtx, cancelFunc)

	configContents, err := ioutil.ReadFile(cfgPath)
	configContentsStr := os.ExpandEnv(string(configContents))
	if err != nil {
		log.Fatal(err)
	}
	cfg := core.AppConfig{}
	if err := yaml.Unmarshal([]byte(configContentsStr), &cfg); err != nil {
		log.Fatal(err)
	}

	managers := make([]core.BlockletMgr, len(cfg.Blocks))
	for i, b := range cfg.Blocks {
		managers[i] = core.MakeBlockletMgr(b)
	}

	header := core.I3barHeader{Version: 1, ClickEvents: true}
	b, _ := json.Marshal(header)
	b = append(b, []byte("\n[\n")...)
	os.Stdout.Write(b)

	updateChans := []core.UpdateChan{}
	for _, m := range managers {
		m.Run()
		updateChans = append(updateChans, m.UpdateChan())
	}
	aggregateUpdateChan := core.CombineUpdateChans(updateChans)
	stdinCloseChan := make(chan struct{})

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
			log.Printf("%+v\n", *ev)
			for _, m := range managers {
				m.ProcessEvent(ev)
			}
		}
		if err := sc.Err(); err != nil {
			log.Fatal(err)
		}
		stdinCloseChan <- struct{}{}
	}()

	for {
		blocks := make([]core.I3barBlock, 0)
		for _, m := range managers {
			blocks = append(blocks, m.Render()...)
		}
		b, _ := json.Marshal(blocks)
		b = append(b, []byte(",\n")...)
		os.Stdout.Write(b)
		select {
		case <-aggregateUpdateChan:
			continue
		case <-stdinCloseChan:
			break
		}
		break
	}
}
