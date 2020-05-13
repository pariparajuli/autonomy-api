package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/score"
)

func (s *Server) getMetrics(c *gin.Context) {
	a := c.MustGet("account")
	account, ok := a.(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	loc := account.Profile.State.LastLocation
	if nil == loc {
		abortWithEncoding(c, http.StatusBadRequest, errorUnknownAccountLocation)
		return
	}

	countToday, countYesterday, err := s.mongoStore.GetPersonalReportCount(c.Param("reportType"), account.AccountNumber)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	avgToday, avgYesterday, err := s.mongoStore.GetCommunityAvgReportCount(c.Param("reportType"), consts.NEARBY_DISTANCE_RANGE, *loc)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"me": gin.H{
			"total_today": countToday,
			"delta":       score.ChangeRate(float64(countToday), float64(countYesterday)),
		},
		"community": gin.H{
			"avg_today": avgToday,
			"delta":     score.ChangeRate(float64(avgToday), float64(avgYesterday)),
		},
	})
}
