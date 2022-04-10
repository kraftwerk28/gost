package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/kraftwerk28/gost/blocks"
	"github.com/kraftwerk28/gost/core"
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

func readEvents(ch chan *core.I3barClickEvent) {
	log := core.Log
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan() // Skip "["
	for sc.Scan() {
		raw := sc.Bytes()
		ev, err := core.NewEventFromRaw(raw)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Click event: %+v\n", *ev)
		ch <- ev
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}

func feedBlocks(o io.Writer, blocks []core.I3barBlock) (err error) {
	buf := new(bytes.Buffer)
	e := json.NewEncoder(buf)
	e.SetEscapeHTML(false)
	if err = e.Encode(blocks); err != nil {
		return
	}
	if err = buf.WriteByte('\n'); err != nil {
		return
	}
	b := buf.Bytes()
	b[len(b)-2] = ','
	_, err = os.Stdout.Write(b)
	return
}

func setupWatcher(path string) (w *fsnotify.Watcher, err error) {
	if w, err = fsnotify.NewWatcher(); err != nil {
		return
	}
	w.Add(path)
	return
}

func main() {
	var cfgPath, logPath string
	var err error
	flag.StringVar(&cfgPath, "config", "", "Path to config.yml")
	flag.StringVar(&logPath, "log", "", "Path to log file")
	flag.Parse()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGUSR2,                 // reload config
		syscall.SIGTERM, syscall.SIGINT, // exit
	)

	logWriter := os.Stderr
	if logPath != "" {
		if logWriter, err = os.OpenFile(
			logPath,
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			0644,
		); err != nil {
			panic(err)
		}
	}
	core.InitializeLogger(logWriter)
	// Logger initialized at this point
	log := core.Log

	{
		header := core.I3barHeader{
			Version:     1,
			ClickEvents: true,
		}
		b, _ := json.Marshal(header)
		b = append(b, []byte("\n[\n")...)
		os.Stdout.Write(b)
	}

	eventChan := make(chan *core.I3barClickEvent)
	go readEvents(eventChan)
	// var fileWatchCh chan fsnotify.Event

outerLoop:
	for {
		ctx, ctxCancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		if cfgPath == "" {
			cfgPath = getConfigPath()
		}
		var cfg *core.AppConfig
		var managers []*core.BlockletMgr
		if cfg, err = core.LoadConfigFromFile(cfgPath); err != nil {
			log.Println(err)
			b := blocks.NewStaticBlock("Error loading the config: " + err.Error())
			managers = []*core.BlockletMgr{core.NewBlockletMgr("error", b, nil)}
			// if cfg.Watch == nil || *cfg.Watch == true {
			// 	w, err := setupWatcher(cfgPath)
			// 	if err == nil {
			// 		fileWatchCh = w.Events
			// 	}
			// }
		} else {
			managers = cfg.CreateManagers(ctx)
		}
		wg.Add(len(managers))
		updateChan := make(chan string)
		for _, m := range managers {
			go m.Run(updateChan, ctx, wg)
		}
		go func() {
		loop:
			for {
				select {
				case e := <-eventChan:
					for _, m := range managers {
						if m.MatchesEvent(e) {
							wg.Add(1)
							go m.ProcessEvent(e, ctx, wg)
						}
					}
				case <-ctx.Done():
					break loop
				}
			}
		}()
	renderLoop:
		for {
			blocks := make([]core.I3barBlock, 0, len(managers))
			for _, m := range managers {
				blocks = append(blocks, m.Render()...)
			}
			if err := feedBlocks(os.Stdout, blocks); err != nil {
				log.Print(err)
			}
			select {
			case updateData := <-updateChan:
				for i := range managers {
					managers[i].TryInvalidate(updateData)
				}
			case signal := <-signalChan:
				switch signal {
				case syscall.SIGUSR2:
					ctxCancel()
					log.Println("Waiting for blocklets to finish...")
					wg.Wait()
					break renderLoop
				case syscall.SIGTERM, syscall.SIGINT:
					ctxCancel()
					log.Println("Waiting for blocklets to finish...")
					c := make(chan struct{})
					go func() {
						wg.Wait()
						c <- struct{}{}
					}()
					for {
						select {
						case <-c:
							break outerLoop
						case s := <-signalChan:
							if s == syscall.SIGINT || s == syscall.SIGTERM {
								log.Println(
									"Blocks didn't finish. Exiting anyway...",
								)
								break outerLoop
							}
						}
					}
				}
				// case e := <-fileWatchCh:
				// 	if e.Op == fsnotify.Create {
				// 		ctxCancel()
				// 		log.Println("Waiting for blocklets to finish...")
				// 		wg.Wait()
				// 		break renderLoop
				// 	}
			}
		}
	}
	log.Println("Auf Wiedersehen.")
}
