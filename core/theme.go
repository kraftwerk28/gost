package core

type ThemeConfig struct {
	Saturation int `yaml:"saturation"`
	Value      int `yaml:"value"`
}

func (t *ThemeConfig) HSVColor(hue int) *ConfigColor {
	return FromHSV(hue, t.Saturation, t.Value)
}
