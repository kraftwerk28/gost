package blocks

import (
	"fmt"
	"time"
)

type TimeBlock struct {
	ch chan int
}

func NewTimeBlock() I3barBlocklet {
	return &TimeBlock{make(chan int)}
}

func (t *TimeBlock) Run() {
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		t.ch <- 0
	}
}

func (t *TimeBlock) UpdateChan() UpdateChan {
	return UpdateChan(t.ch)
}

func (t *TimeBlock) Render() []I3barBlock {
	currentTime := time.Now()
	return []I3barBlock{{
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
