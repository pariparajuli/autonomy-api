package api

import (
	"net/http"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/gin-gonic/gin"
)

// adminExpireRequests is an internal only api to trigger the task to
// check expired help requests
func (s *Server) adminExpireRequests(c *gin.Context) {
	if _, err := s.background.SendTask(&tasks.Signature{
		Name: "expire_help_requests",
	}); err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	c.JSON(200, gin.H{"result": "OK"})
}
