package blocks

import (
	"context"
	"fmt"
	"time"

	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
)

type TimeBlockConfig struct {
	Format   *ConfigFormat   `yaml:"format"`
	Layout   string          `yaml:"layout"`
	Interval *ConfigInterval `yaml:"interval"`
}

type TimeBlock struct {
	TimeBlockConfig
}

func NewTimeBlock() I3barBlocklet {
	defInterval := ConfigInterval(time.Second)
	defaultFmt := NewConfigFormatFromString("{layout}")
	return &TimeBlock{TimeBlockConfig{
		Interval: &defInterval,
		Layout:   time.RFC1123,
		Format:   defaultFmt,
	}}
}

func (t *TimeBlock) GetConfig() interface{} {
	return &t.TimeBlockConfig
}

func (t *TimeBlock) Run(ch UpdateChan, ctx context.Context) {
	ti := time.Tick(time.Duration(*t.Interval))
	for {
		select {
		case <-ti:
			ch.SendUpdate()
		case <-ctx.Done():
			break
		}
	}
}

func (t *TimeBlock) Render() []I3barBlock {
	currentTime := time.Now()
	return []I3barBlock{{
		FullText: t.Format.Expand(formatting.NamedArgs{
			"time": fmt.Sprintf(
				"%d.%d %d:%d:%d",
				currentTime.Day(),
				currentTime.Month(),
				currentTime.Hour(),
				currentTime.Minute(),
				currentTime.Second(),
			),
			"layout": currentTime.Format(t.Layout),
		}),
	}}
}

func init() {
	RegisterBlocklet("time", NewTimeBlock)
}
