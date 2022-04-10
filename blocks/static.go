package blocks

import (
	"context"

	. "github.com/kraftwerk28/gost/core"
)

type StaticBlock struct {
	text  string
	color *ConfigColor
}

func NewStaticBlock(text string) I3barBlocklet {
	return &StaticBlock{text, FromRGB(255, 0, 0)}
}

func (t *StaticBlock) SetColor(color *ConfigColor) {
	t.color = color
}

func (t *StaticBlock) Run(ch UpdateChan, ctx context.Context) {}

func (b *StaticBlock) Render(cfg *AppConfig) []I3barBlock {
	return []I3barBlock{{
		FullText: b.text,
		Color:    b.color.String(),
		Markup:   MarkupPango,
	}}
}
