package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	reportTypeSymptoms  = "symptoms"
	reportTypeBehaviors = "behaviors"

	defaultLimit = int64(100)
)

type historyQueryParams struct {
	Before int64 `form:"before"`
	Limit  int64 `form:"limit"`
}

func (s *Server) getHistory(c *gin.Context) {
	account, ok := c.MustGet("account").(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	var params historyQueryParams
	if err := c.Bind(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}
	limit := defaultLimit
	if params.Limit > 0 {
		limit = params.Limit
	}

	switch c.Param("reportType") {
	case reportTypeSymptoms:
		records, err := s.mongoStore.GetReportedSymptoms(account.AccountNumber, params.Before, limit)
		if err != nil {
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"symptoms_history": records})
	case reportTypeBehaviors:
		records, err := s.mongoStore.GetReportedBehaviors(account.AccountNumber, params.Before, limit)
		if err != nil {
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"behaviors_history": records})
	default:
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters)
	}
}
