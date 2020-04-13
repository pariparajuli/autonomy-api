package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func (s *Server) getSymptoms(c *gin.Context) {
	resp := []SymptomResp{}
	for _, symptom := range schema.Symptoms {
		respSymptom := SymptomResp{ID: symptom.ID, Name: symptom.Name, Desc: symptom.Desc}
		resp = append(resp, respSymptom)
	}
	c.JSON(http.StatusOK, gin.H{"symptoms": resp})
}

// SymptomResp  respons struct of a symptom
type SymptomResp struct {
	ID   schema.SymptomType `json:"id"`
	Name string             `json:"name"`
	Desc string             `json:"desc"`
}
