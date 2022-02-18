package main

import (
	"fmt"
	blocks "i3bar-attempt/blocks"
	"time"
)

type TimeBlock struct {
	ch chan int
}

func NewBlock() blocks.I3barBlocklet {
	return &TimeBlock{make(chan int)}
}

func (t *TimeBlock) Run() {
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		t.ch <- 0
	}
}

func (t *TimeBlock) UpdateChan() blocks.UpdateChan {
	return blocks.UpdateChan(t.ch)
}

func (t *TimeBlock) Render() []blocks.I3barBlock {
	currentTime := time.Now()
	return []blocks.I3barBlock{{
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
