package formatting

import (
	"testing"
)

func TestSum(t *testing.T) {
	// s := "hello, %{name} %{surname}, you're 0x%04{old#x}"
	s := "{bruh:03} hello, {name^6} {surname; G*_f#420$}, you're"
	f := RustLikeFmt(Parse(s))
	for _, p := range f {
		t.Logf("%+v\n", p.Placeholder)
	}
	r := f.Expand(NamedArgs{
		"name":    "John",
		"surname": "Doe",
		"old":     420,
		"bruh":    "kek",
	})
	t.Logf("[%s]\n", r)
}
