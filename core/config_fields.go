package core

import (
	"errors"
	"regexp"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

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

func (c *ConfigInterval) UnmarshalYAML(value *yaml.Node) (err error) {
	var v string
	if err = value.Decode(&v); err != nil {
		return
	}
	result := new(ConfigInterval)
	if result, err = NewFromString(v); err != nil {
		return
	}
	*c = *result
	return
}

type ConfigColor struct {
	c string
}

var hexColorRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)

func (c *ConfigColor) UnmarshalYAML(node *yaml.Node) (err error) {
	var v string
	if err = node.Decode(&v); err != nil {
		return
	}
	if hexColorRe.FindString(v) == "" {
		return errors.New("Invalid hex color")
	}
	c.c = v
	return
}

func (c *ConfigColor) String() string {
	return c.c
}
