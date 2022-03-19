package core

import "strings"

func CountryFlagFromIsoCode(countryCode string) string {
	if len(countryCode) != 2 {
		return countryCode
	}
	b := []byte(strings.ToUpper(countryCode))
	// Each char is encoded as 1F1E6 to 1F1FF for A-Z
	c1, c2 := b[0]+0xa5, b[1]+0xa5
	// The last byte will always start with 101 (0xa0) and then the 5 least
	// significant bits from the previous result
	b1 := 0xa0 | (c1 & 0x1f)
	b2 := 0xa0 | (c2 & 0x1f)
	// Get the flag string from the UTF-8 representation of our Unicode characters.
	return string([]byte{0xf0, 0x9f, 0x87, b1, 0xf0, 0x9f, 0x87, b2})
}

func PercentageToHue(p int) int {
	return int((float64(p) / 100) * 120)
}
