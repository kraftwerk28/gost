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
	Run(updateChan UpdateChan)
	Render() []I3barBlock
}

type I3barBlockletConfigurable interface {
	I3barBlocklet
	GetConfig() interface{}
}

type I3barBlockletListener interface {
	I3barBlocklet
	OnEvent(*I3barClickEvent)
}

// A helper wrapper around a blocklet
type BlockletMgr struct {
	Name     string
	Blocklet I3barBlocklet
}

func NewBlockletMgr(name string, b I3barBlocklet) *BlockletMgr {
	bm := BlockletMgr{fmt.Sprintf("%s:%d", name, blockletCounters[name]), b}
	blockletCounters[name]++
	return &bm
}

func (bm *BlockletMgr) Render() []I3barBlock {
	blocks := bm.Blocklet.Render()
	for i := range blocks {
		blocks[i].Name = fmt.Sprintf("%s:%d", bm.Name, i)
	}
	return blocks
}

func (bm *BlockletMgr) Run(ch UpdateChan) {
	go bm.Blocklet.Run(ch)
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
