package core

import (
	"context"
	"log"
)

type UpdateChan struct {
	ch   chan string
	name string
}

func (u *UpdateChan) SendUpdate() {
	u.ch <- u.name
}

type I3barBlockletCtor func() I3barBlocklet

var builtin = map[string]I3barBlockletCtor{}
var blockletCounters = map[string]int{}

func RegisterBlocklet(name string, ctor I3barBlockletCtor) {
	builtin[name] = ctor
}

func GetBuiltin(name string) I3barBlockletCtor {
	return builtin[name]
}

type I3barBlocklet interface {
	Run(ch UpdateChan, ctx context.Context)
	Render(cfg *AppConfig) []I3barBlock
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
