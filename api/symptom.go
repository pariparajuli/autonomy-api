package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func (s *Server) getSymptoms(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"symptoms": schema.Symptoms})
}
