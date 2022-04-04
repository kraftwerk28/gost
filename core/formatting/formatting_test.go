package formatting

import (
	"testing"
)

func RustLikeFormatting1(t *testing.T) {
	s := "{foo:03} hello, {bar^6} {baz; G*_f#420$}, you're"
	f := RustLikeFmt(Parse(s))
	res := f.Expand(NamedArgs{
		"foo": 3,
		"bar": "________",
		"baz": 800,
	})
	exp := "003 hello, ______ 420 , you're"
	if res != exp {
		t.Errorf(`Expected "%s" to be "%s"\n`, res, exp)
	}
}
