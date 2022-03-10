package core

import (
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/kraftwerk28/gost/core/formatting"
	"gopkg.in/yaml.v3"
)

var dr = regexp.MustCompile(`^(\d+)([smh]|ms)?$`)

type ConfigInterval time.Duration

func NewFromString(v string) (*ConfigInterval, error) {
	m := dr.FindStringSubmatch(v)
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

type AppConfig struct {
	Version string           `yaml:"version"`
	Blocks  []BlockletConfig `yaml:"blocks"`
}

// TODO: use different formatters?
type ConfigFormat struct {
	formatting.RustLikeFmt
}

func NewConfigFormatFromString(s string) *ConfigFormat {
	return &ConfigFormat{*formatting.NewFromString(s)}
}
