package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	reportTypeSymptoms  = "symptoms"
	reportTypeBehaviors = "behaviors"
	reportTypeLocations = "locations"

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

	var before, limit int64

	switch {
	case params.Before > 0:
		before = params.Before
	case params.Before == 0:
		before = time.Now().Unix()
	default:
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, fmt.Errorf("negative before"))
		return
	}

	switch {
	case params.Limit > 0:
		limit = params.Limit
	case params.Limit == 0:
		limit = defaultLimit
	default:
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, fmt.Errorf("negative limit"))
		return
	}

	switch c.Param("reportType") {
	case reportTypeSymptoms:
		records, err := s.mongoStore.GetReportedSymptoms(account.AccountNumber, before, limit)
		if err != nil {
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"symptoms_history": records})
	case reportTypeBehaviors:
		records, err := s.mongoStore.GetReportedBehaviors(account.AccountNumber, before, limit)
		if err != nil {
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"behaviors_history": records})
	case reportTypeLocations:
		records, err := s.mongoStore.GetReportedLocations(account.AccountNumber, before, limit)
		if err != nil {
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"locations_history": records})
	default:
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters)
	}
}
