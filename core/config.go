package core

import (
	"context"
	"errors"
	"log"
	"os"
	"plugin"
	"regexp"
	"strconv"
	"time"

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
			handle, err := plugin.Open(c.Path)
			if err != nil {
				log.Println("Failed to load plugin:")
				log.Print(err)
				continue
			}
			sym, err := handle.Lookup("NewBlock")
			if err != nil {
				log.Println("Plugin must have `func NewBlock() I3barBlocklet`:")
				log.Print(err)
				continue
			}
			if c, ok := sym.(*I3barBlockletCtor); ok {
				ctor = *c
			} else {
				log.Println("Bad constructor")
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

var durationRegexp = regexp.MustCompile(`^(\d+)([smh]|ms)?$`)

type ConfigInterval time.Duration

func NewFromString(v string) (*ConfigInterval, error) {
	m := durationRegexp.FindStringSubmatch(v)
	if m == nil {
		return nil, errors.New("Invalid value for `interval`")
	}
	var base time.Duration
	mul, _ := strconv.Atoi(m[1])
	switch m[2] {
	case "s", "":
		base = time.Second
	case "m":
		base = time.Minute
	case "h":
		base = time.Hour
	case "ms":
		base = time.Millisecond
	}
	res := ConfigInterval(base * time.Duration(mul))
	return &res, nil
}

func (c *ConfigInterval) UnmarshalYAML(value *yaml.Node) error {
	var v string
	if err := value.Decode(&v); err != nil {
		return err
	}
	result, err := NewFromString(v)
	if err != nil {
		return err
	}
	*c = *result
	return nil
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
