package main

import (
	"bufio"
	"encoding/json"
	logLib "log"
	"os"
	"path"
	"plugin"

	// "gopkg.in/yaml.v2"
	blocks "i3bar-attempt/blocks"
)

var log = createLogger()

func createLogger() *logLib.Logger {
	f, err := os.Create("i3barattempt.log")
	if err != nil {
		panic(err)
	}
	return logLib.New(f, "", logLib.Ldate|logLib.Ltime)
}

var pluggedBlocklets = []string{}

func main() {
	header := blocks.I3barHeader{Version: 1, ClickEvents: true}
	b, _ := json.Marshal(header)
	b = append(b, []byte("\n[\n")...)
	os.Stdout.Write(b)

	blocklets := []blocks.I3barBlocklet{
		blocks.NewClickcountBlock(),
		blocks.NewShellBlock(&blocks.ShellBlockConfig{
			Command:        "while sleep 1; do date; done",
			OnClickCommand: `echo "button: $button; x: $x; y: $y" >> $HOME/clicks.log`,
		}),
	}
	pluginPath := "/home/kraftwerk28/projects/go/src/i3bar-attempt/contrib/build"
	// Load dynamic blocks
	for _, pluginName := range pluggedBlocklets {
		plug, err := plugin.Open(path.Join(pluginPath, pluginName+".so"))
		if err != nil {
			log.Println(err)
			continue
		}
		sym, err := plug.Lookup("NewBlock")
		if err != nil {
			log.Println(err)
			continue
		}
		if newfunc, ok := sym.(func() blocks.I3barBlocklet); ok {
			blocklets = append(blocklets, newfunc())
		}
	}

	updateChans := []blocks.UpdateChan{}
	for _, blocklet := range blocklets {
		if b, ok := blocklet.(blocks.I3barBlockletRun); ok {
			go b.Run()
		}
		if b, ok := blocklet.(blocks.I3barBlockletAutoUpdater); ok {
			updateChans = append(updateChans, b.UpdateChan())
		}
	}
	aggregateUpdateChan := blocks.CombineUpdateChans(updateChans)

	stdinCloseChan := make(chan struct{})

	// Read events from stdin
	go func() {
		sc := bufio.NewScanner(os.Stdin)
		sc.Scan()
		log.Println(sc.Text())
		for sc.Scan() {
			raw := sc.Bytes()
			ev, err := blocks.NewEventFromRaw(raw)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("%+v\n", *ev)
			for _, blocklet := range blocklets {
				bl, ok := blocklet.(blocks.I3barBlockletListener)
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
		blocks := make([]blocks.I3barBlock, 0, len(blocklets))
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
