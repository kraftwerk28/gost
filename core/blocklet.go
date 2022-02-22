package core

import (
	"fmt"
	"strings"
)

type UpdateChan chan int

type I3barBlockletCtor func() I3barBlocklet

var Builtin = map[string]I3barBlockletCtor{}
var blockletCounters = map[string]int{}

func RegisterBlocklet(name string, ctor I3barBlockletCtor) {
	Builtin[name] = ctor
}

// TODO: send some more information, not just integer
func (u *UpdateChan) SendUpdate() {
	*u <- 0
}

func CombineUpdateChans(chans []UpdateChan) UpdateChan {
	ch := UpdateChan(make(chan int))
	for i := range chans {
		go func(c UpdateChan) {
			for {
				v := <-c
				ch <- v
			}
		}(chans[i])
	}
	return ch
}

type I3barBlocklet interface {
	Render() []I3barBlock
}

type I3barBlockletConfigurable interface {
	I3barBlocklet
	GetConfig() interface{}
}

type I3barBlockletListener interface {
	I3barBlocklet
	OnEvent(event *I3barClickEvent)
}

type I3barBlockletRunnable interface {
	I3barBlocklet
	Run()
}

type I3barBlockletSelfUpdater interface {
	I3barBlocklet
	UpdateChan() UpdateChan
}

type PluginLoadFunc func() I3barBlocklet

// A helper wrapper around a blocklet
type BlockletMgr struct {
	name string
	b    I3barBlocklet
}

func NewBlockletMgr(cfg *BlockletConfig) *BlockletMgr {
	// Increment global names
	name := cfg.Name
	bm := BlockletMgr{
		fmt.Sprintf("%s:%d", name, blockletCounters[name]),
		cfg.Blocklet,
	}
	blockletCounters[name]++
	return &bm
}

func (bm *BlockletMgr) Render() []I3barBlock {
	blocks := bm.b.Render()
	for i := range blocks {
		blocks[i].Name = fmt.Sprintf("%s:%d", bm.name, i)
	}
	return blocks
}

func (bm *BlockletMgr) Run() {
	if b, ok := bm.b.(I3barBlockletRunnable); ok {
		go b.Run()
	}
}

func (bm *BlockletMgr) UpdateChan() UpdateChan {
	if b, ok := bm.b.(I3barBlockletSelfUpdater); ok {
		return b.UpdateChan()
	}
	return nil
}

func (bm *BlockletMgr) IsListener() bool {
	if _, ok := bm.b.(I3barBlockletListener); ok {
		return true
	}
	return false
}

func (bm *BlockletMgr) MatchesEvent(e *I3barClickEvent) bool {
	return strings.HasPrefix(e.Name, bm.name)
}

func (bm *BlockletMgr) ProcessEvent(e *I3barClickEvent) bool {
	if bm.MatchesEvent(e) {
		if b, ok := bm.b.(I3barBlockletListener); ok {
			b.OnEvent(e)
			return true
		}
	}
	return false
}
