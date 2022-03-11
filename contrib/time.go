package main

import (
	"context"
	"fmt"
	"time"

	core "github.com/kraftwerk28/gost/core"
)

type TimeBlock struct{}

func NewBlock() core.I3barBlocklet {
	return &TimeBlock{}
}

func (t *TimeBlock) Run(ch core.UpdateChan, ctx context.Context) {
	ticker := time.Tick(time.Second)
	for {
		select {
		case <-ticker:
			ch.SendUpdate()
		case <-ctx.Done():
			return
		}
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
