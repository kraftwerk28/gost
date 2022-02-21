package blocks

import (
	"fmt"

	. "github.com/kraftwerk28/gost/core"
)

type ClickcountConfig struct {
	Fmt string `yaml:"fmt"`
}

type Clickcount struct {
	*ClickcountConfig
	ch     chan int
	clicks uint
}

func NewClickcountBlock() I3barBlocklet {
	return &Clickcount{new(ClickcountConfig), make(chan int), 0}
}

func (c *Clickcount) GetConfig() interface{} {
	return c.ClickcountConfig
}

func (t *Clickcount) UpdateChan() UpdateChan {
	return UpdateChan(t.ch)
}

func (t *Clickcount) Render() []I3barBlock {
	return []I3barBlock{{
		FullText: fmt.Sprintf(t.Fmt, t.clicks),
		Name:     "myclickcount",
	}}
}

func (t *Clickcount) OnEvent(e *I3barClickEvent) {
	t.clicks++
	t.ch <- 0
}

func init() {
	RegisterBlocklet("clicks", NewClickcountBlock)
}
