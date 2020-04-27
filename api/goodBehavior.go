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

func (s *Server) goodBehaviors(c *gin.Context) {
	a := c.MustGet("account")
	account, ok := a.(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}
	var loc *schema.Location
	gp := c.GetHeader("Geo-Position")

	if "" == gp {
		loc = account.Profile.State.LastLocation
		if nil == loc {
			abortWithEncoding(c, http.StatusBadRequest, errorUnknownAccountLocation)
			return
		}
	}

	if lat, long, err := parseGeoPosition(gp); err == nil {
		loc = &schema.Location{Latitude: lat, Longitude: long}
	}

	c.JSON(http.StatusOK, gin.H{"default_behaviors": schema.DefaultBehaviors})
}

func (s *Server) reportBehaviors(c *gin.Context) {
	a := c.MustGet("account")
	account, ok := a.(*schema.Account)
	if !ok {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}
	var loc *schema.Location
	gp := c.GetHeader("Geo-Position")

	if "" == gp {
		loc = account.Profile.State.LastLocation
		if nil == loc {
			abortWithEncoding(c, http.StatusBadRequest, errorUnknownAccountLocation)
			return
		}
	}

	if lat, long, err := parseGeoPosition(gp); err == nil {
		loc = &schema.Location{Latitude: lat, Longitude: long}
	}

	var params struct {
		DefaultBehaviors     []string                     `json:"default_behaviors"`
		SelfDefinedBehaviors []schema.SelfDefinedBehavior `json:"self_defined_behaviors"`
	}

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	defaultBehaviors, selfDefinedBehaviors := getGoodBehaviors(params.DefaultBehaviors, params.SelfDefinedBehaviors)
	behaviorWeight, selfDefinedWeight := behaviorWeight(defaultBehaviors, selfDefinedBehaviors)

	data := schema.BehaviorReportData{
		ProfileID:            account.Profile.ID.String(),
		AccountNumber:        account.Profile.AccountNumber,
		DefaultBehaviors:     defaultBehaviors,
		SelfDefinedBehaviors: selfDefinedBehaviors,
		DefaultWeight:        behaviorWeight,
		SelfDefinedWeight:    selfDefinedWeight,
		Location:             schema.GeoJSON{Type: "Point", Coordinates: []float64{loc.Longitude, loc.Latitude}},
		Timestamp:            time.Now().UTC().Unix(),
	}

	err := s.mongoStore.GoodBehaviorSave(&data)
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
	if err != nil {
		c.Error(err)
	}
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

func getGoodBehaviors(defaultBehaviors []string, selfDefinedBehaviors []schema.SelfDefinedBehavior) ([]schema.DefaultBehavior, []schema.SelfDefinedBehavior) {
	var retBehaviors []schema.DefaultBehavior
	var retSelfDedinedBehaviors []schema.SelfDefinedBehavior

	for _, behavior := range defaultBehaviors {
		st := schema.GoodBehaviorType(behavior)
		v, ok := schema.DefaultBehaviorMatrix[st]
		if ok {
			retBehaviors = append(retBehaviors, v)
		}
	}
	for _, defBehavior := range selfDefinedBehaviors {
		retSelfDedinedBehaviors = append(retSelfDedinedBehaviors, defBehavior)
	}
	return retBehaviors, retSelfDedinedBehaviors
}

func behaviorWeight(behaviors []schema.DefaultBehavior, selfDefined []schema.SelfDefinedBehavior) (float64, float64) {
	var sum float64
	for _, behavior := range behaviors {
		w, ok := schema.DefaultBehaviorWeightMatrix[behavior.ID]
		if ok {
			sum = sum + float64(w.Weight)
		}
	}
	return sum, float64(len(selfDefined))
}
