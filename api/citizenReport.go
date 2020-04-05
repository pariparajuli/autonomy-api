package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func (s *Server) report(c *gin.Context) {
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
	healthscore := score(params.Symptoms)
	data := schema.CitizenReportData{
		AccountNumber: account.Profile.AccountNumber,
		Symptoms:      convertSymptoms(params.Symptoms),
		Location:      schema.GeoJSON{Type: "Point", Coordinates: []float64{loc.Latitude, loc.Longitude}},
		HealthScore:   healthscore,
		Timestamp:     time.Now().Unix(),
	}

	err := s.mongoStore.CitizenReportSave(&data)
	if err != nil {
		c.Error(err)
		return
	}

	err = s.mongoStore.UpdateAccountScore(account.Profile.AccountNumber, healthscore)

	if err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorUpdateScore, err)
	}
	c.JSON(http.StatusOK, gin.H{"result": "OK"})

	return
}

func score(symotoms []string) float64 {
	if len(symotoms) < 10 {
		return float64(100 - len(symotoms)*10)
	}
	return 0
}

func convertSymptoms(ids []string) []schema.Symptom {
	m := make(map[schema.SymptomType]schema.Symptom)
	var ret []schema.Symptom
	for _, s := range schema.Symptoms {
		m[s.ID] = s
	}
	for _, id := range ids {
		st := schema.SymptomType(id)
		sy, ok := m[st]
		if ok {
			ret = append(ret, sy)
		}
	}
	return ret
}
