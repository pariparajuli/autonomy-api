package utils

import (
	"fmt"
	"strings"

	"github.com/bitmark-inc/autonomy-api/consts"
)

// TwCountyKey - convert chinese tw county name into key
func TwCountyKey(county string) (string, error) {
	if zh, ok := consts.TwCountyEnglish[county]; !ok {
		return county, fmt.Errorf("%s not exist", county)
	} else {
		return EnNameToKey(zh), nil
	}
}

// EnNameToKey - normalize *english* county name into all small case with
// underscore
func EnNameToKey(str string) string {
	return strings.Replace(strings.ToLower(str), " ", "_", -1)
}
