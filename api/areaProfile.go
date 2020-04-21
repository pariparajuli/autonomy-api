package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bitmark-inc/autonomy-api/schema"
	scoreUtil "github.com/bitmark-inc/autonomy-api/score"
)

const (
	metricUpdateInterval = 5 * time.Minute
)

func (s *Server) singleAreaProfile(c *gin.Context) {
	poiID, err := primitive.ObjectIDFromHex(c.Param("poiID"))
	if err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, fmt.Errorf("invalid POI ID"))
		return
	}

	metric, err := s.mongoStore.GetPOIMetrics(poiID)
	if err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, fmt.Errorf("error getting poi"))
		return
	}

	c.JSON(http.StatusOK, metric)
}

func (s *Server) currentAreaProfile(c *gin.Context) {
	account, ok := c.MustGet("account").(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	profile, err := s.mongoStore.GetProfile(account.AccountNumber)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	metric := profile.Metric

	if profile.Location != nil {
		location := schema.Location{
			Latitude:  profile.Location.Coordinates[1],
			Longitude: profile.Location.Coordinates[0],
		}

		if time.Now().Sub(time.Unix(metric.LastUpdate, 0)) >= metricUpdateInterval {
			if m, err := scoreUtil.CalculateMetric(s.mongoStore, location); err != nil {
				c.Error(err)
			} else {
				m.LastUpdate = time.Now().UTC().Unix()
				if err := s.mongoStore.UpdateProfileMetric(account.AccountNumber, *m); err != nil {
					c.Error(err)
				} else {
					metric = *m
				}
			}
		}
	}

	c.JSON(http.StatusOK, metric)
}
