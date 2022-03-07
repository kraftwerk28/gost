package formatting

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type RustLikeFmt struct{}

type fmtPlaceholder struct {
	name                          string
	minWidth, maxWidth            int
	minPrefix                     string
	hideMinPrefix, minPrefixSpace bool
	unit                          string
	hideUnit                      bool
	barMaxValue                   int
}

var rustFmtRe = regexp.MustCompile(`(?:^|[^{])\{(\w+)(?::(\d+))?(?:\^(\d+))?(?:;( )?(_)?([num1KMGT]))?(?:\*(_)?([\w%]+))?(?:#(\d+))?\}`)

// 2  3  name
// 4  5  min width
// 6  7  max width
// 8  9  min prefix space
// 10 11 min prefix underscore
// 12 13 min prefix
// 14 15 unit underscore
// 16 17 unit
// 18 19 unit

// {<name>[:[0]<min width>][^<max width>][;[ ][_]<min prefix>][*[_]<unit>][#<bar max value>]}
// (?:^|[^{])\{(\w+)(?::(\d+))?(?:\^(\d+))?\}

func parse(fstr string) ([]string, []fmtPlaceholder) {
	rawParts := []string{}
	placeholders := []fmtPlaceholder{}
	lastIndex := 0
	for _, m := range rustFmtRe.FindAllStringSubmatchIndex(fstr, -1) {
		name := fstr[m[2]:m[3]]
		var minWidth, maxWidth int
		if m[4] == -1 {
			minWidth = -1
		} else {
			minWidth, _ = strconv.Atoi(fstr[m[4]:m[5]])
		}
		if m[6] == -1 {
			maxWidth = -1
		} else {
			maxWidth, _ = strconv.Atoi(fstr[m[6]:m[7]])
		}
		minPrefix := "1"
		if m[12] != -1 {
			minPrefix = fstr[m[12]:m[13]]
		}
		hideMinPrefix, minPrefixSpace := false, false
		if m[10] != -1 {
			hideMinPrefix = true
		}
		if m[8] != -1 {
			minPrefixSpace = true
		}
		unit := ""
		if m[16] != -1 {
			unit = fstr[m[16]:m[17]]
		}
		hideUnit := false
		if m[14] != -1 {
			hideUnit = true
		}
		barMaxValue := -1
		if m[18] != -1 {
			barMaxValue, _ = strconv.Atoi(fstr[m[18]:m[19]])
		}
		p := fmtPlaceholder{
			name,
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

func (f *RustLikeFmt) Sprintf(fstr string, args NamedArgs) string {
	b := strings.Builder{}
	rawParts, placeholders := parse(fstr)
	for _, p := range placeholders {
		fmt.Printf("%+v\n", p)
	}
	i := 0
	for ; i < len(placeholders); i++ {
		b.Write([]byte(rawParts[i]))
		b.Write([]byte(args[placeholders[i].name].(string)))
	}
	b.Write([]byte(rawParts[i]))
	return b.String()
}
