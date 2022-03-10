package formatting

import (
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var rustFmtRe = regexp.MustCompile(`(?:^|[^{])\{(\w+)(?::(0)?(\d+))?(?:\^(\d+))?(?:;( )?(_)?([num1KMGT]))?(?:\*(_)?([\w%]+))?(?:#(\d+))?\}`)

// 2  3  name
// 4  5  min width zero
// 6  7  min width
// 8  9  max width
// 10 11 min prefix space
// 12 13 min prefix underscore
// 14 15 min prefix
// 16 17 unit underscore
// 18 19 unit
// 20 21 unit

// {<name>[:[0]<min width>][^<max width>][;[ ][_]<min prefix>][*[_]<unit>][#<bar max value>]}
// (?:^|[^{])\{(\w+)(?::(\d+))?(?:\^(\d+))?\}

type RustLikeFmt struct {
	rawParts     []string
	placeholders []fmtPlaceholder
}

type fmtPlaceholder struct {
	name                          string
	minWidthZero                  bool
	minWidth, maxWidth            int
	minPrefix                     string
	hideMinPrefix, minPrefixSpace bool
	unit                          string
	hideUnit                      bool
	barMaxValue                   int
}

func NewFromString(v string) *RustLikeFmt {
	parts, placeholders := parse(v)
	return &RustLikeFmt{parts, placeholders}
}

func (f *RustLikeFmt) UnmarshalYAML(node *yaml.Node) error {
	var raw string
	if err := node.Decode(&raw); err != nil {
		return err
	}
	f.rawParts, f.placeholders = parse(raw)
	return nil
}

func (f *RustLikeFmt) Expand(args NamedArgs) string {
	b := strings.Builder{}
	i := 0
	for ; i < len(f.placeholders); i++ {
		b.Write([]byte(f.rawParts[i]))
		b.Write([]byte(args[f.placeholders[i].name].(string)))
	}
	b.Write([]byte(f.rawParts[i]))
	return b.String()
}

func (p *fmtPlaceholder) format(value interface{}) string {
	r := value.(string)
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
	// TODO: check for rest specifiers...
	return r
}

func parse(fstr string) ([]string, []fmtPlaceholder) {
	rawParts := []string{}
	placeholders := []fmtPlaceholder{}
	lastIndex := 0
	for _, m := range rustFmtRe.FindAllStringSubmatchIndex(fstr, -1) {
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
		placeholders = append(placeholders, p)
		if m[0] == 0 {
			rawParts = append(rawParts, fstr[lastIndex:m[0]])
		} else {
			rawParts = append(rawParts, fstr[lastIndex:m[0]+1])
		}
		lastIndex = m[1]
	}
	rawParts = append(rawParts, fstr[lastIndex:])
	return rawParts, placeholders
}
