package api

import (
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/getsentry/sentry-go"

	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/utils"

)

func (s *Server) createSymptom(c *gin.Context) {
	var params schema.Symptom

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	id, err := s.mongoStore.CreateSymptom(params)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return
}

func (s *Server) getSymptoms(c *gin.Context) {
	symptoms, err := s.mongoStore.ListSymptoms()
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"symptoms": symptoms})
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
	official, customerized, err := s.findSymptomsInDB(params.Symptoms)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}
	totalSymptoms := append(official, customerized...)
	symptomScore := score(totalSymptoms)
	data := schema.SymptomReportData{
		ProfileID:          account.Profile.ID.String(),
		AccountNumber:      account.Profile.AccountNumber,
		OfficialSymptoms:   official,
		Location:           schema.GeoJSON{Type: "Point", Coordinates: []float64{loc.Longitude, loc.Latitude}},
		CustomizedSymptoms: customerized,
		SymptomScore:       symptomScore,
		Timestamp:          time.Now().UTC().Unix(),
	}

	err = s.mongoStore.SymptomReportSave(&data)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	accts, err := s.mongoStore.NearestDistance(consts.NEARBY_DISTANCE_RANGE, *loc)
	if nil == err {
		go func() {
			if err := utils.TriggerAccountUpdate(*s.cadenceClient, c, accts); err != nil {
				sentry.CaptureException(err)
			}
		}()
	} else {
		c.Error(err)
	}
	pois, err := s.mongoStore.NearestPOI(consts.NEARBY_DISTANCE_RANGE, *loc)
	if nil == err {
		go func() {
			if err := utils.TriggerPOIUpdate(*s.cadenceClient, c, pois); err != nil {
				sentry.CaptureException(err)
			}
		}()
	} else {
		c.Error(err)
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

func (s *Server) findSymptomsInDB(ids []string) ([]schema.Symptom, []schema.Symptom, error) {
	var syIDs []schema.SymptomType
	for _, id := range ids {
		syIDs = append(syIDs, schema.SymptomType(id))
	}
	official, customeried, _, err := s.mongoStore.QuerySymptoms(syIDs)
	return official, customeried, err
}
