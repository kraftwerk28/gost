package formatting

import (
	"testing"
)

func TestSum(t *testing.T) {
	// s := "hello, %{name} %{surname}, you're 0x%04{old#x}"
	s := "{bruh:3} hello, {name^6} {surname; G*_f#420}, you're"
	var f Formatter
	f = &RustLikeFmt{}
	r := f.Sprintf(s, NamedArgs{
		"name":    "John",
		"surname": "Doe",
		"old":     420,
		"bruh":    "kek",
	})
	t.Logf("[%s]\n", r)
}
