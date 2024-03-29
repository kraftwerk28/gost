package core

import (
	"context"
	"fmt"
	"log"
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

func MakeBlockletMgr(
	name string,
	b I3barBlocklet,
	cfg *AppConfig,
) BlockletMgr {
	bmName := fmt.Sprintf("%s:%d", name, blockletCounters[name])
	blockletCounters[name]++
	return BlockletMgr{name: bmName, blocklet: b, appConfig: cfg}
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
	blocks := bm.blocklet.Render(bm.appConfig)
	for i := range blocks {
		b := &blocks[i]
		if b.Name == "" {
			b.Name = fmt.Sprintf("%s:%d", bm.name, i)
		} else {
			b.Name = fmt.Sprintf("%s:%s", bm.name, b.Name)
		}
		if bm.appConfig != nil {
			if w := bm.appConfig.SeparatorWidth; w > 0 {
				b.SeparatorBlockWidth = w
			}
			if bm.appConfig.Markup == MarkupPango {
				b.Markup = bm.appConfig.Markup
			}
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

func (bm *BlockletMgr) MatchesEvent(e *I3barClickEvent) bool {
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

func (bm *BlockletMgr) ProcessEvent(
	e *I3barClickEvent,
	ctx context.Context,
	wg *sync.WaitGroup,
) bool {
	defer wg.Done()
	if cfg := bm.getBaseConfig(); cfg != nil {
		if cfg.OnClick != nil {
			cmd := e.ShellCommand(*cfg.OnClick, ctx)
			if err := cmd.Run(); err != nil {
				Log.Print(err)
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
