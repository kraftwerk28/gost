package core

import (
	"context"
	"fmt"
	"strings"
)

// A helper wrapper around a blocklet
type BlockletMgr struct {
	Name     string
	Blocklet I3barBlocklet
	Ctx      context.Context
}

func NewBlockletMgr(name string, b I3barBlocklet, ctx context.Context) *BlockletMgr {
	bm := BlockletMgr{
		Name:     fmt.Sprintf("%s:%d", name, blockletCounters[name]),
		Blocklet: b,
		Ctx:      ctx,
	}
	blockletCounters[name]++
	return &bm
}

func (bm *BlockletMgr) Render() []I3barBlock {
	blocks := bm.Blocklet.Render()
	for i := range blocks {
		if blocks[i].Name == "" {
			blocks[i].Name = fmt.Sprintf("%s:%d", bm.Name, i)
		} else {
			blocks[i].Name = fmt.Sprintf("%s:%s", bm.Name, blocks[i].Name)
		}
	}
	return blocks
}

func (bm *BlockletMgr) Run(ch UpdateChan) {
	go bm.Blocklet.Run(ch, bm.Ctx)
}

func (bm *BlockletMgr) IsListener() bool {
	if _, ok := bm.Blocklet.(I3barBlockletListener); ok {
		return true
	}
	return false
}

func (bm *BlockletMgr) MatchesEvent(e *I3barClickEvent) bool {
	return strings.HasPrefix(e.Name, bm.Name)
}

func (bm *BlockletMgr) ProcessEvent(e *I3barClickEvent) bool {
	if bm.MatchesEvent(e) {
		if b, ok := bm.Blocklet.(I3barBlockletListener); ok {
			b.OnEvent(e)
			return true
		}
	}
	return false
}
