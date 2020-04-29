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

func (s *Server) createBehavior(c *gin.Context) {
	var params schema.Behavior

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}

	id, err := s.mongoStore.CreateBehavior(params)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return
}

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
	official, err := s.mongoStore.ListOfficialBehavior()
	if err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorUnknownAccountLocation)
	}
	customerized, err := s.areaBehaviorInfection(consts.NEARBY_DISTANCE_RANGE, *loc)
	if err != nil {
		c.Error(err)
	}

	if customerized != nil {
		official = append(official, customerized...)
	}

	c.JSON(http.StatusOK, gin.H{"behaviors": official})
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
		Behaviors []string `json:"behaviors"`
	}

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}
	official, customerized, err := s.getBehaviors(params.Behaviors)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}
	behaviorWeight, selfDefinedWeight := behaviorWeight(official, customerized)

	data := schema.BehaviorReportData{
		ProfileID:             account.Profile.ID.String(),
		AccountNumber:         account.Profile.AccountNumber,
		OfficialBehaviors:     official,
		CustomerizedBehaviors: customerized,
		OfficialWeight:        behaviorWeight,
		CustomerizedWeight:    selfDefinedWeight,
		Location:              schema.GeoJSON{Type: "Point", Coordinates: []float64{loc.Longitude, loc.Latitude}},
		Timestamp:             time.Now().UTC().Unix(),
	}

	err = s.mongoStore.GoodBehaviorSave(&data)
	if err != nil {
		abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
		return
	}
	err = s.userBehaviorInfection(&data, *loc)
	if err != nil { // do nothing
		c.Error(err)
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

func (s *Server) getBehaviors(ids []string) ([]schema.Behavior, []schema.Behavior, error) {
	var behviorIDs []schema.GoodBehaviorType
	for _, id := range ids {
		behviorIDs = append(behviorIDs, schema.GoodBehaviorType(id))
	}
	official, customeried, _, err := s.mongoStore.IDToBehaviors(behviorIDs)
	return official, customeried, err
}

func behaviorWeight(official []schema.Behavior, customerized []schema.Behavior) (float64, float64) {
	var sum float64
	for _, behavior := range official {
		w, ok := schema.DefaultBehaviorWeightMatrix[behavior.ID]
		if ok {
			sum = sum + float64(w.Weight)
		}
	}
	return sum, float64(len(customerized))
}

func (s *Server) userBehaviorInfection(infectedUser *schema.BehaviorReportData, loc schema.Location) error {
	err := s.mongoStore.UpdateAreaProfileBehavior(infectedUser.CustomerizedBehaviors, loc)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) areaBehaviorInfection(distance int, loc schema.Location) ([]schema.Behavior, error) {
	list, err := s.mongoStore.AreaCustomerizedBehaviorList(distance, loc)
	if err != nil {
		return nil, err
	}
	return list, nil
}
