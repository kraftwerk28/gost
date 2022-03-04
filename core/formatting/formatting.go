package formatting

import (
	"fmt"
	"regexp"
)

type NamedArgs map[string]interface{}

type Formatter interface {
	Sprintf(string, map[string]interface{})
}

type GoFmt struct{}

// Sprintf support named format
func (f *GoFmt) Sprintf(format string, params map[string]interface{}) string {
	s, p := f.formatPlaceholder(format, params)
	return fmt.Sprintf(s, p...)
}

var plRe = regexp.MustCompile(
	`(?:[^{]|^)\{([\w_-]+)(#[xsvTtbcdoqXxUeEfFgGp])?\}`,
)

func (fmter *GoFmt) formatPlaceholder(
	f string,
	args map[string]interface{},
) (string, []interface{}) {
	fmtArgs := []interface{}{}
	lastRawIndex := 0
	out := ""
	for _, idx := range plRe.FindAllStringSubmatchIndex(f, -1) {
		mStart, mEnd := idx[0], idx[1]
		keyStart, keyEnd := idx[2], idx[3]
		specStart, specEnd := idx[4], idx[5]
		if mStart > 0 && f[mStart-1] == '{' &&
			mEnd < len(f) && f[mEnd+1] == '}' {
			out += "%" + f[mStart:mEnd]
			continue
		}
		mapKey := f[keyStart:keyEnd]
		if mapVal, ok := args[mapKey]; ok {
			fmtArgs = append(fmtArgs, mapVal)
		} else {
			fmtArgs = append(fmtArgs, "N/A")
		}
		out += f[lastRawIndex : keyStart-1]
		if specStart != -1 {
			// for example %{mykey#f} - float format
			lastRawIndex = specEnd + 1
			out += f[specStart+1 : specEnd]
		} else {
			// for example %{mykey} - no format specifier, defaulting to `v`
			out += "v"
			lastRawIndex = keyEnd + 1
		}
	}
	out += f[lastRawIndex:]
	return out, fmtArgs
}
