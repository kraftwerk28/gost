package formatting

import (
	"testing"
)

func TestSum(t *testing.T) {
	s := "hello, %{name} %{surname}, you're 0x%04{old#x}"
	r := Sprintf(s, NamedArgs{
		"name":    "John",
		"surname": "Doe",
		"old":     420,
		"bruh":    "kek",
	})
	t.Logf("[%s]\n", r)
}
