package main

import (
	"fmt"
	"time"

	core "github.com/kraftwerk28/gost/core"
)

type TimeBlock struct {
	ch chan int
}

func NewBlock() core.I3barBlocklet {
	return &TimeBlock{make(chan int)}
}

func (t *TimeBlock) GetConfig() interface{}{
}

func (t *TimeBlock) Run() {
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		t.ch <- 0
	}
}

func (t *TimeBlock) UpdateChan() core.UpdateChan {
	return core.UpdateChan(t.ch)
}

func (t *TimeBlock) Render() []core.I3barBlock {
	currentTime := time.Now()
	return []core.I3barBlock{{
		FullText: fmt.Sprintf(
			"%d.%d %d:%d:%d",
			currentTime.Day(),
			currentTime.Month(),
			currentTime.Hour(),
			currentTime.Minute(),
			currentTime.Second(),
		),
		// TODO: auto-assigned name
		Name: "myclock",
	}}
}
