package formatting

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// var rustFmtRe = regexp.MustCompile(`(?:^|[^{])\{(\w+)(?::(0)?(\d+))?(?:\^(\d+))?(?:;( )?(_)?([num1KMGT]))?(?:\*(_)?([\w%]+))?(?:#(\d+))?\}`)
var rustFmtRe = regexp.MustCompile(`\{(\w+)(?::(0)?(\d+))?(?:\^(\d+))?(?:;( )?(_)?([num1KMGT]))?(?:\*(_)?([\w%]+))?(?:#(\d+))?\}`)

// 2  3  name
// 4  5  min width zero
// 6  7  min width
// 8  9  max width
// 10 11 min prefix space
// 12 13 min prefix underscore
// 14 15 min prefix
// 16 17 unit underscore
// 18 19 unit
// 20 21 bar max value

// {<name>[:[0]<min width>][^<max width>][;[ ][_]<min prefix>][*[_]<unit>][#<bar max value>]}
// (?:^|[^{])\{(\w+)(?::(\d+))?(?:\^(\d+))?\}

type RustLikeFmt []fmtPart

type fmtPart struct {
	Placeholder *fmtPlaceholder
	Raw         string
}

type fmtPlaceholder struct {
	name               string
	minWidthZero       bool
	minWidth, maxWidth int
	// Engineering suffix, i.e. 1.0m, 4.3K etc
	minPrefix                     string
	hideMinPrefix, minPrefixSpace bool
	unit                          string
	hideUnit                      bool
	barMaxValue                   int
}

func NewFromString(v string) RustLikeFmt {
	parts := Parse(v)
	return RustLikeFmt(parts)
}

func (f *RustLikeFmt) UnmarshalYAML(node *yaml.Node) error {
	var raw string
	if err := node.Decode(&raw); err != nil {
		return err
	}
	*f = Parse(raw)
	return nil
}

func (f RustLikeFmt) Expand(args NamedArgs) string {
	b := strings.Builder{}
	for _, part := range f {
		if p := part.Placeholder; p != nil {
			if value, ok := args[p.name]; ok {
				b.WriteString(p.format(value))
			}
		} else {
			b.WriteString(part.Raw)
		}
	}
	return b.String()
}

func (p *fmtPlaceholder) format(value interface{}) string {
	var r string
	vof := reflect.ValueOf(value)
	switch vof.Kind() {
	case reflect.Float32, reflect.Float64:
		n := vof.Float()
		if max := float64(p.barMaxValue); p.barMaxValue > -1 && n > max {
			n = max
		}
		const floatPrecision = 8
		r = strconv.FormatFloat(n, 'f', floatPrecision, 64)
	case reflect.Int, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint32, reflect.Uint64:
		n := vof.Int()
		if max := int64(p.barMaxValue); p.barMaxValue != -1 && n > max {
			n = max
		}
		r = strconv.FormatInt(n, 10)
	case reflect.String:
		r = vof.String()
	}
	if p.minWidth > -1 && len(r) < p.minWidth {
		fill := " "
		if p.minWidthZero {
			fill = "0"
		}
		r = strings.Repeat(fill, p.minWidth-len(r)) + r
	}
	if p.maxWidth > -1 && len(r) > p.maxWidth {
		r = r[:p.maxWidth]
	}
	return r
}

func Parse(fstr string) (parts []fmtPart) {
	parts = make([]fmtPart, 0)
	lastIndex := 0
	for _, m := range rustFmtRe.FindAllStringSubmatchIndex(fstr, -1) {
		if m[0] > 0 && fstr[m[0]-1] == '{' {
			parts = append(parts, fmtPart{Raw: fstr[lastIndex:m[1]]})
			lastIndex = m[1]
			continue
		}
		name := fstr[m[2]:m[3]]
		var minWidth, maxWidth int
		minWidthZero := false
		if m[4] != -1 {
			minWidthZero = true
		}
		if m[6] == -1 {
			minWidth = -1
		} else {
			minWidth, _ = strconv.Atoi(fstr[m[6]:m[7]])
		}
		if m[8] == -1 {
			maxWidth = -1
		} else {
			maxWidth, _ = strconv.Atoi(fstr[m[8]:m[9]])
		}
		minPrefix := "1"
		if m[14] != -1 {
			minPrefix = fstr[m[14]:m[15]]
		}
		hideMinPrefix, minPrefixSpace := false, false
		if m[12] != -1 {
			hideMinPrefix = true
		}
		if m[10] != -1 {
			minPrefixSpace = true
		}
		unit := ""
		if m[18] != -1 {
			unit = fstr[m[18]:m[19]]
		}
		hideUnit := false
		if m[16] != -1 {
			hideUnit = true
		}
		barMaxValue := -1
		if m[20] != -1 {
			barMaxValue, _ = strconv.Atoi(fstr[m[20]:m[21]])
		}
		p := fmtPlaceholder{
			name,
			minWidthZero,
			minWidth, maxWidth,
			minPrefix,
			hideMinPrefix, minPrefixSpace,
			unit,
			hideUnit,
			barMaxValue,
		}
		parts = append(
			parts,
			fmtPart{Raw: fstr[lastIndex:m[0]]},
			fmtPart{Placeholder: &p},
		)
		lastIndex = m[1]
	}
	parts = append(parts, fmtPart{Raw: fstr[lastIndex:]})
	return
}
