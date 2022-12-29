package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kraftwerk28/gost/blocks"
	"github.com/kraftwerk28/gost/core"
)

const programName = "gost"

func getConfigPath(flag string) string {
	if len(flag) > 0 {
		return flag
	}
	xdg, hasXdg := os.LookupEnv("XDG_CONFIG_HOME")
	if !hasXdg {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		xdg = path.Join(home, ".config")
	}
	p := path.Join(xdg, programName, "config.yml")
	if _, err := os.Stat(p); err != nil {
		p = path.Join(xdg, programName, "config.yaml")
	}
	if _, err := os.Stat(p); err != nil {
		return ""
	}
	return p
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

var isFirstFrame = true

func feedBlocks(o io.Writer, blocks []core.I3barBlock) (err error) {
	if isFirstFrame {
		isFirstFrame = false
	} else {
		if _, err = os.Stdout.WriteString(","); err != nil {
			return
		}
	}
	e := json.NewEncoder(os.Stdout)
	e.SetEscapeHTML(false)
	if err = e.Encode(blocks); err != nil {
		return
	}
	return
}

func setupWatcher(path string) (w *fsnotify.Watcher, err error) {
	if w, err = fsnotify.NewWatcher(); err != nil {
		return
	}
	w.Add(path)
	return
}

func processEvents(
	ctx context.Context,
	managers []core.BlockletMgr,
	wg *sync.WaitGroup,
	events chan *core.I3barClickEvent,
) {
	for {
		select {
		case e := <-events:
			for i := range managers {
				m := &managers[i]
				if m.MatchesEvent(e) {
					wg.Add(1)
					go m.ProcessEvent(e, ctx, wg)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func stopWithTimeout(cancel context.CancelFunc, wg *sync.WaitGroup) bool {
	wgDoneCh := make(chan struct{})
	go func() {
		wg.Wait()
		wgDoneCh <- struct{}{}
	}()
	cancel()
	tout := time.NewTimer(time.Second * 3)
	select {
	case <-wgDoneCh:
		tout.Stop()
		return true
	case <-tout.C:
		return false
	}
}

func main() {
	var cfgPathFlag, logPath string
	var err error
	flag.StringVar(&cfgPathFlag, "config", "", "Path to config.yml")
	flag.StringVar(&logPath, "log", "", "Path to log file")
	flag.Parse()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,                  // reload config
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

outerLoop:
	for {
		ctx, ctxCancel := context.WithCancel(context.Background())
		wg := sync.WaitGroup{}
		cfgPath := getConfigPath(cfgPathFlag)
		if cfgPath == "" {
			log.Fatalln("No config file found")
		}
		var cfg *core.AppConfig
		var configWatcher *fsnotify.Watcher
		var fileWatchChan chan fsnotify.Event
		var managers []core.BlockletMgr
		if cfg, err = core.LoadConfigFromFile(cfgPath); err != nil {
			log.Println(err)
			b := blocks.NewStaticBlock("Error loading the config: " + err.Error())
			managers = []core.BlockletMgr{core.MakeBlockletMgr("error", b, nil)}
		} else {
			managers = cfg.CreateManagers(ctx)
			if cfg.WatchConfig == nil || *cfg.WatchConfig {
				configWatcher, err = setupWatcher(cfgPath)
				if err == nil {
					fileWatchChan = configWatcher.Events
					log.Println("Watching config for changes")
				} else {
					log.Print(err)
				}
			}
		}
		wg.Add(len(managers))
		updateChan := make(chan string)
		for i := range managers {
			go managers[i].Run(updateChan, ctx, &wg)
		}
		go processEvents(ctx, managers, &wg, eventChan)
	renderLoop:
		for {
			blocks := make([]core.I3barBlock, 0, len(managers))
			for i := range managers {
				blocks = append(blocks, managers[i].Render()...)
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
				case syscall.SIGHUP:
					log.Println("Waiting for blocklets to stop")
					if !stopWithTimeout(ctxCancel, &wg) {
						log.Println("Blocklets failed to stop")
					}
					break renderLoop
				case syscall.SIGTERM, syscall.SIGINT:
					log.Println("Waiting for blocklets to finish")
					if !stopWithTimeout(ctxCancel, &wg) {
						log.Println("Blocklets failed to stop")
					}
					break outerLoop
				}
			case e := <-fileWatchChan:
				if e.Op == fsnotify.Write {
					log.Println("Config change detected")
					log.Println("Waiting for blocklets to finish")
					if !stopWithTimeout(ctxCancel, &wg) {
						log.Println("Blocklets failed to stop")
						break outerLoop
					} else {
						configWatcher.Close()
						break renderLoop
					}
				}
			}
		}
	}
	log.Println("Auf Wiedersehen")
}
