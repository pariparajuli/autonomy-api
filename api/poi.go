package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/store"
)

type userPOI struct {
	ID       string           `json:"id"`
	Alias    string           `json:"alias"`
	Address  string           `json:"address"`
	Location *schema.Location `json:"location"`
	Score    float64          `json:"score"`
}

func (s *Server) addPOI(c *gin.Context) {
	var body userPOI
	if err := c.BindJSON(&body); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	if body.Location == nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, fmt.Errorf("location not provided"))
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
	body.Score = poi.Metric.Score
	c.JSON(http.StatusOK, body)
}

func (s *Server) getPOI(c *gin.Context) {
	account, ok := c.MustGet("account").(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	pois, err := s.mongoStore.ListPOI(account.AccountNumber)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	c.JSON(http.StatusOK, pois)
}

func (s *Server) updatePOIAlias(c *gin.Context) {
	account, ok := c.MustGet("account").(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	var body userPOI
	if err := c.BindJSON(&body); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	poiID, err := primitive.ObjectIDFromHex(c.Param("poiID"))
	if err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, fmt.Errorf("invalid POI ID"))
		return
	}

	if err := s.mongoStore.UpdatePOIAlias(account.AccountNumber, body.Alias, poiID); err != nil {
		switch err {
		case store.ErrPOINotFound:
			abortWithEncoding(c, http.StatusBadRequest, errorUnknownPOI)
		default:
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK"})
}

func (s *Server) updatePOIOrder(c *gin.Context) {

	account, ok := c.MustGet("account").(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	var params struct {
		Order []string `json:"order"`
	}

	if err := c.BindJSON((&params)); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	if err := s.mongoStore.UpdatePOIOrder(account.AccountNumber, params.Order); err != nil {
		switch err {
		case store.ErrPOIListNotFound:
			abortWithEncoding(c, http.StatusInternalServerError, errorPOIListNotFound, err)
			return
		case store.ErrPOIListMismatch:
			abortWithEncoding(c, http.StatusInternalServerError, errorPOIListMissmatch, err)
			return
		default:
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK"})
}

func (s *Server) deletePOI(c *gin.Context) {
	account, ok := c.MustGet("account").(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	poiID, err := primitive.ObjectIDFromHex(c.Param("poiID"))
	if err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, fmt.Errorf("invalid POI ID"))
		return
	}

	if err := s.mongoStore.DeletePOI(account.AccountNumber, poiID); err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK"})
}
