package utils

import (
	"fmt"
	"strings"
	"time"
)

var locations map[string]*time.Location = map[string]*time.Location{}

func init() {
	for i := time.Duration(-12); i < 15; i++ {
		name := fmt.Sprintf("GMT%+d", i)
		locations[name] = time.FixedZone(name, int((i * time.Hour).Seconds()))
	}
}

// GetLocation returns a location of a GMT-X format timezone from a pre-defined locations map.
func GetLocation(timezone string) *time.Location {
	if tz, ok := locations[strings.ToUpper(timezone)]; ok {
		return tz
	}
	return nil
}
