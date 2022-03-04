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
	TimeBlockConfig
}

func NewTimeBlock() I3barBlocklet {
	return &TimeBlock{}
}

func (t *TimeBlock) GetConfig() interface{} {
	return &t.TimeBlockConfig
}

func (t *TimeBlock) Run(ch UpdateChan) {
	ti := time.Tick(time.Second)
	for {
		<-ti
		ch.SendUpdate()
	}
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
