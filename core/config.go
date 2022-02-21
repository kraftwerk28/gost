package core

import (
	"errors"
	"fmt"
)

type BlockletConfig struct {
	Blocklet I3barBlocklet
}

func (bc *BlockletConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	nameOnly := struct {
		Name string `yaml:"name"`
	}{}
	if err := unmarshal(&nameOnly); err != nil {
		return err
	}
	// TODO: handle not defined block
	ctor, bExists := Builtin[nameOnly.Name]
	if !bExists {
		return errors.New(fmt.Sprintf(`Builtin blocklet "%s" is not defined`, nameOnly.Name))
	}
	blocklet := ctor()
	if c, ok := blocklet.(I3barBlockletConfigurable); ok {
		cfg := c.GetConfig()
		if err := unmarshal(cfg); err != nil {
			return err
		}
		bc.Blocklet = c
	}
	return nil
}

type AppConfig struct {
	Version string            `yaml:"version"`
	Blocks  []*BlockletConfig `yaml:"blocks"`
}
