package core

import (
	"context"
	"log"
)

type UpdateChan chan int

type I3barBlockletCtor func() I3barBlocklet

var Builtin = map[string]I3barBlockletCtor{}
var blockletCounters = map[string]int{}

func RegisterBlocklet(name string, ctor I3barBlockletCtor) {
	Builtin[name] = ctor
}

func (u UpdateChan) SendUpdate() {
	// TODO: send some more information, not just integer
	u <- 0
}

type I3barBlocklet interface {
	Run(ch UpdateChan, ctx context.Context)
	Render() []I3barBlock
}

type I3barBlockletConfigurable interface {
	I3barBlocklet
	GetConfig() interface{}
}

type I3barBlockletListener interface {
	I3barBlocklet
	OnEvent(*I3barClickEvent, context.Context)
}

type I3barBlockletLogger interface {
	I3barBlocklet
	GetLogger() *log.Logger
}
