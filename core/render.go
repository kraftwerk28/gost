package core

type Render struct {
	blocks []I3barBlock
	markup I3barMarkup
}

func MakeRender() Render {
	return Render{
		blocks: []I3barBlock{},
	}
}

func (r *Render) Block(b I3barBlock) {
	r.blocks = append(r.blocks, b)
}

func (r *Render) Markup(m I3barMarkup) {
	r.markup = m
}

func (r *Render) Build() []I3barBlock {
	return r.blocks
}
