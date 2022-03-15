// This is an example blocklet that doesn't have much sense.
// Use it as a template.
package blocks

import (
	"context"
	"fmt"

	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
)

type ClickcountConfig struct {
	Format *ConfigFormat `yaml:"format"`
}

type Clickcount struct {
	ClickcountConfig
	clicks int
	ch     UpdateChan
}

func NewClickcountBlock() I3barBlocklet {
	b := Clickcount{}
	b.Format = NewConfigFormatFromString("{clicks}")
	return &b
}

func (c *Clickcount) Run(ch UpdateChan, ctx context.Context) {
	c.ch = ch
}

func (c *Clickcount) GetConfig() interface{} {
	return &c.ClickcountConfig
}

func (t *Clickcount) Render() []I3barBlock {
	txt := t.Format.Expand(formatting.NamedArgs{
		"clicks": fmt.Sprintf("%d", t.clicks),
	})
	return []I3barBlock{{FullText: txt}}
}

func (t *Clickcount) OnEvent(e *I3barClickEvent, ctx context.Context) {
	if e.Button == ButtonScrollDown {
		t.clicks--
	} else {
		t.clicks++
	}
	if t.clicks < 0 {
		t.clicks = 0
	}
	t.ch.SendUpdate()
}

func init() {
	RegisterBlocklet("clicks", NewClickcountBlock)
}
