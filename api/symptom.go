package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func (s *Server) getSymptoms(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"symptoms": schema.Symptoms})
}

func (s *Server) reportSymptoms(c *gin.Context) {
	a := c.MustGet("account")
	account, ok := a.(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	loc := account.Profile.State.LastLocation
	if nil == loc {
		abortWithEncoding(c, http.StatusBadRequest, errorUnknownAccountLocation)
		return
	}

	var params struct {
		Symptoms []string `json:"symptoms"`
	}

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}
	symptoms, symptomIDs := getSymptoms(params.Symptoms)

	symptomScore := score(symptoms)
	data := schema.SymptomReportData{
		AccountNumber: account.Profile.AccountNumber,
		Symptoms:      symptomIDs,
		Location:      schema.GeoJSON{Type: "Point", Coordinates: []float64{loc.Longitude, loc.Latitude}},
		SymptomScore:  symptomScore,
		Timestamp:     time.Now().Unix(),
	}

	err := s.mongoStore.SymptomReportSave(&data)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "OK"})

	return
}

func score(symptoms []schema.Symptom) float64 {
	var sum float64 = 0
	for _, symptom := range symptoms {
		sum = sum + symptom.Weight
	}
	if len(symptoms) >= 3 {
		return 2 * sum
	}
	return sum
}

func getSymptoms(ids []string) ([]schema.Symptom, []string) {
	var symptoms []schema.Symptom
	var syIDs []string
	for _, id := range ids {
		st := schema.SymptomType(id)
		sy, ok := schema.SymptomFromID[st]
		if ok {
			symptoms = append(symptoms, sy)
			syIDs = append(syIDs, string(sy.ID))
		}
	}
	return symptoms, syIDs
}
