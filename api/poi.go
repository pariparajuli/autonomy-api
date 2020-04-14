package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/schema"
)

type userPOI struct {
	ID       string           `json:"id"`
	Alias    string           `json:"alias"`
	Address  string           `json:"address"`
	Location *schema.Location `json:"location"`
	Score    int              `json:"score"`
}

func (s *Server) addPOI(c *gin.Context) {
	var body userPOI
	if err := c.BindJSON(&body); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	account, ok := c.MustGet("account").(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	poi, err := s.mongoStore.AddPOI(account.AccountNumber, body.Alias, body.Address,
		body.Location.Longitude, body.Location.Latitude)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	body.ID = poi.ID.Hex()
	body.Score = poi.Score
	c.JSON(http.StatusOK, body)
}
