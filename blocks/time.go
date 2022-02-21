package blocks

import (
	"fmt"
	"time"

	. "github.com/kraftwerk28/gost/core"
)

type TimeBlockConfig struct {
	Zone     string `yaml:"zone"`
	interval int
}

type TimeBlock struct {
	*TimeBlockConfig
	ch chan int
}

func NewTimeBlock() I3barBlocklet {
	return &TimeBlock{new(TimeBlockConfig), make(chan int)}
}

func (t *TimeBlock) GetConfig() interface{} {
	return t.TimeBlockConfig
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

func init() {
	RegisterBlocklet("time", NewTimeBlock)
}
