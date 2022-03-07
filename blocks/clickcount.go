package blocks

import (
	"context"
	"fmt"

	. "github.com/kraftwerk28/gost/core"
)

type ClickcountConfig struct {
	Fmt string `yaml:"fmt"`
}

type Clickcount struct {
	*ClickcountConfig
	ch     UpdateChan
	clicks uint
}

func NewClickcountBlock() I3barBlocklet {
	return &Clickcount{nil, make(chan int), 0}
}

func (c *Clickcount) Run(ch UpdateChan, ctx context.Context) {
	c.ch = ch
}

func (c *Clickcount) GetConfig() interface{} {
	return &c.ClickcountConfig
}

func (t *Clickcount) Render() []I3barBlock {
	return []I3barBlock{{FullText: fmt.Sprintf(t.Fmt, t.clicks)}}
}

func (t *Clickcount) OnEvent(e *I3barClickEvent) {
	if e.Button == ButtonScrollDown {
		t.clicks--
	} else {
		t.clicks++
	}
	if t.clicks == 0 {
		t.clicks = 0
	}
	t.ch.SendUpdate()
}

func init() {
	RegisterBlocklet("clicks", NewClickcountBlock)
}
