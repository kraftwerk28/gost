package core

type BlockletConfig struct {
	Name string `yaml:"name"`
	// Path to plugin, if `name == "plugin"`
	Path string                 `yaml:"path"`
	Rest map[string]interface{} `yaml:",inline"`
}

type AppConfig struct {
	Version string           `yaml:"version"`
	Blocks  []BlockletConfig `yaml:"blocks"`
}
