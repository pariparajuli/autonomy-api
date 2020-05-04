package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	cadenceClient "go.uber.org/cadence/client"

	scoreWorker "github.com/bitmark-inc/autonomy-api/background/score"
	"github.com/bitmark-inc/autonomy-api/schema"
	scoreUtil "github.com/bitmark-inc/autonomy-api/score"
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

// getProfileFormula returns customized formula saved by a user
func (s *Server) getProfileFormula(c *gin.Context) {
	var isDefaultFormula bool
	accountNumber := c.GetString("requester")

	coefficient, err := s.mongoStore.GetProfileCoefficient(accountNumber)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	if coefficient == nil {
		isDefaultFormula = true
		coefficient = &schema.ScoreCoefficient{
			Symptoms:       scoreUtil.DefaultScoreV1SymptomCoefficient,
			Behaviors:      scoreUtil.DefaultScoreV1BehaviorCoefficient,
			Confirms:       scoreUtil.DefaultScoreV1ConfirmCoefficient,
			SymptomWeights: schema.DefaultSymptomWeights,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"is_default":  isDefaultFormula,
		"coefficient": coefficient,
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

	profile.Metric.Score = scoreUtil.TotalScoreV1(params.Coefficient,
		profile.Metric.SymptomScore,
		profile.Metric.BehaviorScore,
		profile.Metric.ConfirmedScore,
	)

	if err := s.mongoStore.UpdateProfileMetric(accountNumber, &profile.Metric); err != nil {
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
		metric.Score = scoreUtil.TotalScoreV1(params.Coefficient, metric.SymptomScore, metric.BehaviorScore, metric.ConfirmedScore)

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

	profile.Metric.Score = scoreUtil.DefaultTotalScore(
		profile.Metric.SymptomScore,
		profile.Metric.BehaviorScore,
		profile.Metric.ConfirmedScore,
	)

	if err := s.mongoStore.UpdateProfileMetric(accountNumber, &profile.Metric); err != nil {
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
