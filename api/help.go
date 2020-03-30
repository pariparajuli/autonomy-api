package api

import (
	"net/http"

	"github.com/bitmark-inc/autonomy-api/store"
	"github.com/gin-gonic/gin"
)

// askForHelp is the API for asking help from others
func (s *Server) askForHelp(c *gin.Context) {
	requester := c.GetString("requester")

	var params struct {
		Subject string `json:"subject"`
		Text    string `json:"text"`
	}

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	req, err := s.store.RequestHelp(requester, params.Subject, params.Text)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	// TODO: broadcast a notification to surrounding users
	c.JSON(http.StatusOK, req)
	return
}

// answerHelp is the API for answer a help
func (s *Server) answerHelp(c *gin.Context) {
	id := c.Param("helpID")
	requester := c.GetString("requester")

	if err := s.store.AnswerHelp(requester, id); err != nil {
		if err == store.ErrRequestNotExist {
			abortWithEncoding(c, http.StatusNotFound, errorRequestNotExist, err)
		} else {
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK"})
}
