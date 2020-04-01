package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	cohortCount = 450
)

func (s *Server) score(c *gin.Context) {
	a := c.MustGet("account")
	account, ok := a.(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	loc := account.Profile.State.LastLocation
	ids, err := s.mongoStore.NearestCount(cohortCount, *loc)

	if nil != err {
		abortWithEncoding(c, http.StatusInternalServerError, errorScore, err)
		return
	}

	score, err := s.mongoStore.Health(ids)
	if nil != err {
		abortWithEncoding(c, http.StatusInternalServerError, errorScore, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"score": score,
	})
}
