package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	cadenceClient "go.uber.org/cadence/client"

	scoreWorker "github.com/bitmark-inc/autonomy-api/background/score"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/score"
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

	if _, err := s.cadenceClient.StartWorkflow(c, cadenceClient.StartWorkflowOptions{
		ID:                           fmt.Sprintf("account-state-%s", accountNumber),
		TaskList:                     scoreWorker.TaskListName,
		ExecutionStartToCloseTimeout: time.Hour,
		WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
	}, "AccountStateUpdateWorkflow", accountNumber); err != nil {
		c.Error(err)
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

// accountHere is an api to acking for an account
func (s *Server) accountHere(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// getProfileFormula returns customized formula saved by a user
func (s *Server) getProfileFormula(c *gin.Context) {
	var isDefaultFormula bool
	accountNumber := c.GetString("requester")

	var params struct {
		Language string `form:"lang"`
	}

	if err := c.Bind(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	lang := "en"
	if params.Language != "" {
		lang = params.Language
	}

	coefficient, err := s.mongoStore.GetProfileCoefficient(accountNumber)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	symptoms, err := s.mongoStore.ListOfficialSymptoms(lang)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	if coefficient == nil {
		isDefaultFormula = true
		coefficient = &schema.ScoreCoefficient{
			Symptoms:       score.DefaultScoreV1SymptomCoefficient,
			Behaviors:      score.DefaultScoreV1BehaviorCoefficient,
			Confirms:       score.DefaultScoreV1ConfirmCoefficient,
			SymptomWeights: schema.DefaultSymptomWeights,
		}
	}

	type SymptomWeightsRepresentation struct {
		Symptom schema.Symptom `json:"symptom"`
		Weight  float64        `json:"weight"`
	}

	SymptomWeightsRepresentationList := make([]SymptomWeightsRepresentation, 0)

	for _, s := range symptoms {
		if weight, ok := coefficient.SymptomWeights[s.ID]; ok {
			SymptomWeightsRepresentationList = append(SymptomWeightsRepresentationList, SymptomWeightsRepresentation{
				Symptom: s,
				Weight:  weight,
			})
		}

	}

	c.JSON(http.StatusOK, gin.H{
		"is_default": isDefaultFormula,
		"coefficient": map[string]interface{}{
			"symptoms":        coefficient.Symptoms,
			"behaviors":       coefficient.Behaviors,
			"confirms":        coefficient.Confirms,
			"symptom_weights": SymptomWeightsRepresentationList,
		},
	})
}

// updateProfileFormula will update a customized formula submitted by a user
func (s *Server) updateProfileFormula(c *gin.Context) {
	accountNumber := c.GetString("requester")

	var params struct {
		Coefficient schema.ScoreCoefficient
	}

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	params.Coefficient.UpdatedAt = time.Now().UTC()

	if err := s.mongoStore.UpdateProfileCoefficient(accountNumber, params.Coefficient); err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	profile, err := s.mongoStore.GetProfile(accountNumber)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	profile.Metric.Score = score.TotalScoreV1(params.Coefficient,
		profile.Metric.Details.Symptoms.Score,
		profile.Metric.Details.Behaviors.Score,
		profile.Metric.Details.Confirm.Score,
	)

	if err := s.mongoStore.UpdateProfileMetric(accountNumber, profile.Metric); err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	for _, profilePOI := range profile.PointsOfInterest {
		poi, err := s.mongoStore.GetPOI(profilePOI.ID)
		if err != nil {
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		}

		metric := poi.Metric
		metric.Score = score.TotalScoreV1(params.Coefficient, metric.Details.Symptoms.Score, metric.Details.Behaviors.Score, metric.Details.Confirm.Score)

		if err := s.mongoStore.UpdateProfilePOIMetric(profile.AccountNumber, poi.ID, metric); err != nil {
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK"})
}

// resetProfileFormula cleans up existing customized formula for a user
func (s *Server) resetProfileFormula(c *gin.Context) {
	accountNumber := c.GetString("requester")

	if err := s.mongoStore.ResetProfileCoefficient(accountNumber); err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	profile, err := s.mongoStore.GetProfile(accountNumber)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	profile.Metric.Score = score.DefaultTotalScore(
		profile.Metric.Details.Symptoms.Score,
		profile.Metric.Details.Behaviors.Score,
		profile.Metric.Details.Confirm.Score,
	)

	if err := s.mongoStore.UpdateProfileMetric(accountNumber, profile.Metric); err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	for _, profilePOI := range profile.PointsOfInterest {
		poi, err := s.mongoStore.GetPOI(profilePOI.ID)
		if err != nil {
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		}

		metric := poi.Metric

		if err := s.mongoStore.UpdateProfilePOIMetric(profile.AccountNumber, poi.ID, metric); err != nil {
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK"})
}

// profile returns personal profile includes both individual and neighbor data
func (s *Server) profile(c *gin.Context) {
	account, ok := c.MustGet("account").(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}

	profile, err := s.mongoStore.GetProfile(account.AccountNumber)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	individualMetric := profile.IndividualMetric
	metric := profile.Metric

	if profile.Location != nil {
		location := schema.Location{
			Latitude:  profile.Location.Coordinates[1],
			Longitude: profile.Location.Coordinates[0],
		}

		// FIXME: return cached result directly if possible; otherwise get coefficient and run SyncAccountMetrics

		metricLastUpdate := time.Unix(metric.LastUpdate, 0)
		var coefficient *schema.ScoreCoefficient

		if time.Since(metricLastUpdate) >= metricUpdateInterval {
			// will sync with coefficient = nil
		} else if coefficient = profile.ScoreCoefficient; coefficient != nil && coefficient.UpdatedAt.Sub(metricLastUpdate) > 0 {
			// will sync with coefficient = profile.ScoreCoefficient
		} else {
			c.JSON(http.StatusOK, gin.H{
				"individual": individualMetric,
				"neighbor":   metric,
			})
			return
		}

		m, err := s.mongoStore.SyncAccountMetrics(account.AccountNumber, coefficient, location)
		if err != nil {
			c.Error(err)
			abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
			return
		} else {
			metric = *m
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"individual": individualMetric,
		"neighbor":   metric,
	})
}
