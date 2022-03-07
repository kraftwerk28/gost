package core

import "context"

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

// Deprecated
// func CombineUpdateChans(chans []UpdateChan) UpdateChan {
// 	ch := UpdateChan(make(chan int))
// 	for i := range chans {
// 		go func(c UpdateChan) {
// 			for {
// 				v := <-c
// 				ch <- v
// 			}
// 		}(chans[i])
// 	}
// 	return ch
// }

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
	OnEvent(*I3barClickEvent)
}
