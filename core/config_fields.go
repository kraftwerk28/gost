package core

import (
	"errors"
	"fmt"
	"math"
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
		return nil, errors.New("invalid value for `interval`")
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
	if c, err = NewFromString(v); err != nil {
		return
	}
	*c = *result
	return
}

type ConfigColor struct {
	r, g, b, a uint8
}

var hexColorRe = regexp.MustCompile(`^#([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})?$`)

func FromHSV(h, s, v int) *ConfigColor {
	var ss, vv, c, x, m, r, g, b float64
	ss, vv = float64(s)/100, float64(v)/100
	c = ss * vv
	x = c * (1 - math.Abs(math.Mod(float64(h)/60, 2)-1))
	m = float64(v) - c
	if h < 60 {
		r, g, b = c, x, 0
	} else if h < 120 {
		r, g, b = x, c, 0
	} else if h < 180 {
		r, g, b = 0, c, x
	} else if h < 240 {
		r, g, b = 0, x, c
	} else if h < 300 {
		r, g, b = x, 0, c
	} else {
		r, g, b = c, 0, x
	}
	return &ConfigColor{
		uint8((r - m) * 255),
		uint8((g - m) * 255),
		uint8((b - m) * 255),
		0xff,
	}
}

func FromRGB(r, g, b uint8) *ConfigColor {
	return &ConfigColor{r, g, b, 0xff}
}

func FromRGBA(r, g, b, a uint8) *ConfigColor {
	return &ConfigColor{r, g, b, a}
}

func (c *ConfigColor) UnmarshalYAML(node *yaml.Node) (err error) {
	var v string
	if err = node.Decode(&v); err != nil {
		return
	}
	m := hexColorRe.FindStringSubmatch(v)
	if m == nil {
		return errors.New("invalid hex color")
	}
	r, _ := strconv.ParseUint(m[1], 16, 8)
	g, _ := strconv.ParseUint(m[2], 16, 8)
	b, _ := strconv.ParseUint(m[3], 16, 8)
	a := uint64(0xff)
	if m[4] != "" {
		a, _ = strconv.ParseUint(m[4], 16, 8)
	}
	c.r, c.g, c.b, c.a = uint8(r), uint8(g), uint8(b), uint8(a)
	return
}

func (c *ConfigColor) String() string {
	s := fmt.Sprintf("#%02x%02x%02x", c.r, c.g, c.b)
	if c.a < 0xff {
		s += fmt.Sprintf("%02x", c.a)
	}
	return s
}
