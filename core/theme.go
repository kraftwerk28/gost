package core

type ThemeConfig struct {
	Saturation int `yaml:"saturation"`
	Value      int `yaml:"value"`
}

type Theme struct{}

func (t *Theme) Apply(b *I3barBlock) {}
