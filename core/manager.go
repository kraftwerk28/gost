package core

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
)

// A helper wrapper around a blocklet
// TODO: utilize `instance` field, as defined in the i3bar protocol
type BlockletMgr struct {
	name        string
	blocklet    I3barBlocklet
	renderCache []I3barBlock
	appConfig   *AppConfig
	isError     bool
}

func NewBlockletMgr(
	name string,
	b I3barBlocklet,
	cfg *AppConfig,
) *BlockletMgr {
	bmName := fmt.Sprintf("%s:%d", name, blockletCounters[name])
	blockletCounters[name]++
	return &BlockletMgr{name: bmName, blocklet: b, appConfig: cfg}
}

func (bm *BlockletMgr) invalidateCache() {
	if bm.isError {
		if len(bm.renderCache) > 0 {
			bm.renderCache = []I3barBlock{{
				FullText: fmt.Sprintf("E [%s]", bm.name),
				Color:    "#ff0000",
			}}
		} else {
			for i := range bm.renderCache {
				bm.renderCache[i].Color = "#ff0000"
			}
		}
		return
	}
	blocks := bm.blocklet.Render()
	for i := range blocks {
		if blocks[i].Name == "" {
			blocks[i].Name = fmt.Sprintf("%s:%d", bm.name, i)
		} else {
			blocks[i].Name = fmt.Sprintf("%s:%s", bm.name, blocks[i].Name)
		}
		if w := bm.appConfig.SeparatorWidth; w > 0 {
			blocks[i].SeparatorBlockWidth = w
		}
	}
	bm.renderCache = blocks
}

func (bm *BlockletMgr) Render() []I3barBlock {
	if bm.renderCache == nil {
		bm.invalidateCache()
	}
	return bm.renderCache
}

func (bm *BlockletMgr) initLogger() {
	if logb, ok := bm.blocklet.(I3barBlockletLogger); ok {
		prefix := Log.Prefix() + ":" + bm.name
		*logb.GetLogger() = *log.New(Log.Writer(), prefix, Log.Flags())
	}
}

func (bm *BlockletMgr) Run(ch chan string, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	bm.initLogger()
	uc := UpdateChan{ch, bm.name}
	defer func() {
		if r := recover(); r != nil {
			Log.Printf("error in blocklet %s:", bm.name)
			Log.Print(r)
			bm.isError = true
			bm.invalidateCache()
			uc.SendUpdate()
		}
	}()
	bm.blocklet.Run(uc, ctx)
}

func (bm *BlockletMgr) IsListener() bool {
	if _, ok := bm.blocklet.(I3barBlockletListener); ok {
		return true
	}
	return false
}

func (bm *BlockletMgr) matchesEvent(e *I3barClickEvent) bool {
	return strings.HasPrefix(e.Name, bm.name)
}

func (bm *BlockletMgr) getBaseConfig() *BaseBlockletConfig {
	if bc, ok := bm.blocklet.(I3barBlockletConfigurable); ok {
		if c, ok := bc.GetConfig().(BaseBlockletConfigIface); ok {
			return c.Get()
		}
	}
	return nil
}

func (bm *BlockletMgr) ProcessEvent(e *I3barClickEvent, ctx context.Context) bool {
	if !bm.matchesEvent(e) {
		return false
	}
	if cfg := bm.getBaseConfig(); cfg != nil {
		if cfg.OnClick != nil {
			cmd := e.ShellCommand(*cfg.OnClick, ctx)
			cerr := cmd.Run()
			if err, ok := cerr.(*exec.ExitError); ok {
				Log.Printf("%s\n", err.Stderr)
			}
		}
	}
	if b, ok := bm.blocklet.(I3barBlockletListener); ok {
		b.OnEvent(e, ctx)
		return true
	}
	return false
}

// If name matches blocklet manager name, re-render blocklets
func (bm *BlockletMgr) TryInvalidate(name string) {
	if name == bm.name {
		bm.invalidateCache()
	}
}
