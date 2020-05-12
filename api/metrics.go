package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
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
		fmt.Println(err)
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}
	personalDelta := float64(1)
	if countYesterday > 0 {
		personalDelta = float64(countToday-countYesterday) / float64(countYesterday)
	}

	avgToday, avgYesterday, err := s.mongoStore.GetCommunityAvgReportCount(c.Param("reportType"), consts.CORHORT_DISTANCE_RANGE, *loc)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}
	communityDelta := float64(1)
	if avgYesterday > 0 {
		communityDelta = (avgToday - avgYesterday) / avgYesterday
	}

	c.JSON(http.StatusOK, gin.H{
		"me": gin.H{
			"total_today": countToday,
			"delta":       personalDelta,
		},
		"community": gin.H{
			"avg_today": avgToday,
			"delta":     communityDelta,
		},
	})
}
