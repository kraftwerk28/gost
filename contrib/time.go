package main

import (
	"fmt"
	"time"

	core "github.com/kraftwerk28/gost/core"
)

type TimeBlock struct{}

func NewBlock() core.I3barBlocklet {
	return new(TimeBlock)
}

func (t *TimeBlock) Run(ch core.UpdateChan) {
	ticker := time.Tick(time.Second * 4)
	for {
		<-ticker
		ch <- 0
	}
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
	}}
}
