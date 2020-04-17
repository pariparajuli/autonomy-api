package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func (s *Server) areaProfile(c *gin.Context) {
	p := c.Param("poiID")
	if p == "" {
		s.currentAreaProfile(c)
	} else {
		s.singleAreaProfile(c)
	}
}

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

	metric, err := s.mongoStore.ProfileMetric(account.AccountNumber)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	c.JSON(http.StatusOK, metric)
}
