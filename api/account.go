package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/schema"
)

// accountRegister is the API for register a new account
func (s *Server) accountRegister(c *gin.Context) {
	accountNumber := c.GetString("requester")

	var params struct {
		EncPubKey string                 `json:"enc_pub_key"`
		Metadata  map[string]interface{} `json:"metadata"`
	}

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	a, err := s.store.CreateAccount(accountNumber, params.EncPubKey, params.Metadata)
	if err != nil {
		abortWithEncoding(c, http.StatusForbidden, errorAccountTaken, err)
		return
	}

	err = s.mongoStore.CreateAccount(a)
	if err != nil {
		abortWithEncoding(c, http.StatusForbidden, errorAccountTaken, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": a.Profile,
	})
}

// accountDetail is the API to query an account
func (s *Server) accountDetail(c *gin.Context) {
	a := c.MustGet("account")
	account, ok := a.(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": account.Profile,
	})
}

// accountUpdateMetadata is the API to update metadata for a user
func (s *Server) accountUpdateMetadata(c *gin.Context) {
	accountNumber := c.GetString("requester")

	var params struct {
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorCannotParseRequest, err)
		return
	}

	if err := s.store.UpdateAccountMetadata(accountNumber, params.Metadata); err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK"})
}

// accountDelete is the API to remove an account from our service
func (s *Server) accountDelete(c *gin.Context) {
	accountNumber := c.GetString("requester")

	if err := s.store.DeleteAccount(accountNumber); err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	if err := s.mongoStore.DeleteAccount(accountNumber); err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK"})
}
