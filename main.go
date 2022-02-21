package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"path"

	// "plugin"

	_ "github.com/kraftwerk28/gost/blocks"
	"github.com/kraftwerk28/gost/core"
	"gopkg.in/yaml.v2"
)

var log = core.Log

const PROGRAM_NAME = "gost"

func main() {
	core.InitializeLogger("/home/kraftwerk28/i3bar-attempt.log")

	cfgPath := flag.String("config", "", "Path to config.yml")
	flag.Parse()
	if *cfgPath == "" {
		xdg, hasXdg := os.LookupEnv("XDG_CONFIG_HOME")
		if !hasXdg {
			home, _ := os.UserHomeDir()
			xdg = path.Join(home, ".config")
		}
		*cfgPath = path.Join(xdg, PROGRAM_NAME, "config.yml")
	}
	if _, err := os.Stat(*cfgPath); err != nil {
		panic(errors.New("Could not load config"))
	}

	programCtx := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(programCtx)
	println(cancelCtx, cancelFunc)

	configContents, err := ioutil.ReadFile(*cfgPath)
	if err != nil {
		panic(err)
	}
	cfg := core.AppConfig{}
	if err := yaml.Unmarshal(configContents, &cfg); err != nil {
		panic(err)
	}

	header := core.I3barHeader{Version: 1, ClickEvents: true}
	b, _ := json.Marshal(header)
	b = append(b, []byte("\n[\n")...)
	os.Stdout.Write(b)

	blocklets := make([]core.I3barBlocklet, len(cfg.Blocks))
	// blocklets := []core.I3barBlocklet{
	// 	blocks.NewClickcountBlock(),
	// 	blocks.NewShellBlock(&blocks.ShellBlockConfig{
	// 		Command:        `while :; do date; sleep 1; done`,
	// 		OnClickCommand: `echo "button: $button; x: $x; y: $y"`,
	// 	}),
	// }
	for i, bc := range cfg.Blocks {
		blocklets[i] = bc.Blocklet
	}
	// homedir, _ := os.UserHomeDir()
	// pluginPath := path.Join(
	// 	homedir,
	// 	"projects/go/src/i3bar-attempt/contrib/build",
	// )
	// Load dynamic blocks
	// for _, pluginName := range pluggedBlocklets {
	// 	plug, err := plugin.Open(path.Join(pluginPath, pluginName+".so"))
	// 	if err != nil {
	// 		log.Println(err)
	// 		continue
	// 	}
	// 	sym, err := plug.Lookup("NewBlock")
	// 	if err != nil {
	// 		log.Println(err)
	// 		continue
	// 	}
	// 	if loadFunc, ok := sym.(core.PluginLoadFunc); ok {
	// 		blocklets = append(blocklets, loadFunc())
	// 	}
	// }

	updateChans := []core.UpdateChan{}
	for _, blocklet := range blocklets {
		if b, ok := blocklet.(core.I3barBlockletRunnable); ok {
			go b.Run()
		}
		if b, ok := blocklet.(core.I3barBlockletSelfUpdater); ok {
			updateChans = append(updateChans, b.UpdateChan())
		}
	}
	aggregateUpdateChan := core.CombineUpdateChans(updateChans)

	stdinCloseChan := make(chan struct{})

	// Read events from stdin
	go func() {
		log := core.Log
		sc := bufio.NewScanner(os.Stdin)
		sc.Scan()
		// log.Println(sc.Text())
		for sc.Scan() {
			raw := sc.Bytes()
			ev, err := core.NewEventFromRaw(raw)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("%+v\n", *ev)
			for _, blocklet := range blocklets {
				bl, ok := blocklet.(core.I3barBlockletListener)
				if !ok {
					continue
				}
				// FIXME: avoid using .Render() to check block name
				rendered := blocklet.Render()
				for _, b := range rendered {
					if ev.Name == b.Name {
						go bl.OnEvent(ev)
					}
				}
			}
		}
		if err := sc.Err(); err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Event stream exited")
		}
		stdinCloseChan <- struct{}{}
	}()

	for {
		blocks := make([]core.I3barBlock, 0, len(blocklets))
		for _, blocklet := range blocklets {
			b := blocklet.Render()
			blocks = append(blocks, b...)
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
