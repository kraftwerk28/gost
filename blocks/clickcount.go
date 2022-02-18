package blocks

import (
	"fmt"
)

type Clickcount struct {
	ch     chan int
	clicks uint
}

func NewClickcountBlock() I3barBlocklet {
	return &Clickcount{make(chan int), 0}
}

func (t *Clickcount) UpdateChan() UpdateChan {
	return UpdateChan(t.ch)
}

func (t *Clickcount) Render() []I3barBlock {
	return []I3barBlock{{
		FullText: fmt.Sprintf("clicks: %d", t.clicks),
		Name:     "myclickcount",
	}}
}

func (t *Clickcount) OnEvent(e *I3barClickEvent) {
	t.clicks++
	t.ch <- 0
}
