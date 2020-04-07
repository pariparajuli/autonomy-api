package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// parseGeoPosition will parse latitude and longitude from the geo-position string
func parseGeoPosition(geoPosition string) (float64, float64, error) {
	positions := strings.Split(geoPosition, ";")

	if len(positions) != 2 {
		return 0, 0, fmt.Errorf("invalid geo-position value")
	}

	lat, err := strconv.ParseFloat(positions[0], 64)
	if err != nil {

		return 0, 0, err
	}

	long, err := strconv.ParseFloat(positions[1], 64)
	if err != nil {

		return 0, 0, err
	}

	return lat, long, nil
}

// updateGeoPositionMiddleware is a middleware to store geo-position for every
// api requests from users
func (s *Server) updateGeoPositionMiddleware(c *gin.Context) {
	gp := c.GetHeader("Geo-Position")
	accountNumber := c.GetString("requester")

	if gp != "" && accountNumber != "" {
		if lat, long, err := parseGeoPosition(gp); err == nil {
			if err := s.store.UpdateAccountGeoPosition(accountNumber, lat, long); err != nil {
				c.Error(err)
			}
		} else {
			c.Error(err)
		}
	}
	c.Next()
}
