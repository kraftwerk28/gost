package core

import (
	"context"
	"log"
	"os"
	"plugin"

	"github.com/kraftwerk28/gost/core/formatting"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Version        string           `yaml:"version"`
	SeparatorWidth int              `yaml:"separator_width"`
	Blocks         []BlockletConfig `yaml:"blocks"`
}

func LoadConfigFromFile(filename string) (*AppConfig, error) {
	cfgFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer cfgFile.Close()
	cfgDecoder := yaml.NewDecoder(cfgFile)
	cfg := &AppConfig{}
	if err := cfgDecoder.Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *AppConfig) CreateManagers(ctx context.Context) []*BlockletMgr {
	managers := make([]*BlockletMgr, 0, len(cfg.Blocks))
	for _, c := range cfg.Blocks {
		var ctor I3barBlockletCtor
		if c.Name == "plugin" {
			var err error
			var handle *plugin.Plugin
			var sym interface{}
			if handle, err = plugin.Open(c.Path); err != nil {
				log.Println("Failed to load plugin:", err)
				continue
			}
			if sym, err = handle.Lookup("NewBlock"); err != nil {
				log.Println(
					"Plugin must have `func NewBlock() I3barBlocklet`:",
					err,
				)
				continue
			}
			if c, ok := sym.(*I3barBlockletCtor); ok {
				ctor = *c
			} else {
				log.Println("Bad constructor. The plugin must have the following line:")
				log.Println("`var NewBlock core.I3barBlockletCtor = newFooBlock`,")
				log.Println("where `newFooBlock` is your blocklet constructor.")
				continue
			}
		} else if ct := GetBuiltin(c.Name); ct != nil {
			ctor = ct
		} else {
			log.Fatalf(`Unrecognized blocklet name: "%s"`, c.Name)
		}
		blocklet := ctor()
		if b, ok := blocklet.(I3barBlockletConfigurable); ok {
			cf, _ := yaml.Marshal(c)
			if err := yaml.Unmarshal(cf, b.GetConfig()); err != nil {
				log.Fatal(err)
			}
		}
		m := NewBlockletMgr(c.Name, blocklet, cfg)
		managers = append(managers, m)
	}
	return managers
}

type BlockletConfig struct {
	Name string `yaml:"name"`
	// Path to plugin, if `name == "plugin"`
	Path string                 `yaml:"path"`
	Rest map[string]interface{} `yaml:",inline"`
}

// TODO: use different formatters?
type ConfigFormat struct {
	formatting.RustLikeFmt
}

func NewConfigFormatFromString(s string) *ConfigFormat {
	return &ConfigFormat{formatting.NewFromString(s)}
}

type BaseBlockletConfig struct {
	Color   *ConfigColor `yaml:"color,omitempty"`
	OnClick *string      `yaml:"on_click"`
}

type BaseBlockletConfigIface interface {
	Get() *BaseBlockletConfig
}

func (c *BaseBlockletConfig) Get() *BaseBlockletConfig { return c }
