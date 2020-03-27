package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type symptom struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

var symptoms = []symptom{
	{"Fever", ""},
	{"Cough", ""},
	{"Fatigue", ""},
	{"Difficulty breathing", ""},
	{"Nasal congestion", ""},
	{"Sore throat", ""},
	{"Chest pain or pressure", ""},
	{"Bluish lips or face", ""},
}

func (s *Server) getSymptoms(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"symptoms": symptoms})
}
