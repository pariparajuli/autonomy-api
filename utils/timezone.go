package utils

import (
	"fmt"
	"strings"
	"time"
)

var locations map[string]*time.Location = map[string]*time.Location{}

var utcBaseTime = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)

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

	unknowTimezone := strings.Replace(timezone, "GMT", "", -1) // Turn GMT+12:45 to +12:45
	if len(unknowTimezone) == 5 {
		b := []byte(unknowTimezone)
		unknowTimezone = fmt.Sprintf("%s0%s", string(b[0]), string(b[1:]))
	}

	t, err := time.Parse("Z07:00", unknowTimezone) // Parse +12:45 as a time struct
	if err != nil {
		return nil
	}
	tz := time.FixedZone(timezone, int(utcBaseTime.Sub(t).Seconds()))
	locations[timezone] = tz
	return tz
}
