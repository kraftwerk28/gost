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
		Layout:   "Mon 02.01.2006 15:04:05",
		Format:   defaultFmt,
	}}
}

func (t *TimeBlock) GetConfig() interface{} {
	return &t.TimeBlockConfig
}

func (t *TimeBlock) Run(ch UpdateChan, ctx context.Context) {
	tickTimer := time.NewTicker(time.Duration(*t.Interval))
	for {
		select {
		case <-tickTimer.C:
			ch.SendUpdate()
		case <-ctx.Done():
			return
		}
	}
}

func (t *TimeBlock) Render(cfg *AppConfig) []I3barBlock {
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
