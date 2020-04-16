package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func (s *Server) goodBehaviors(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"good_behaviors": schema.GoodBehaviors})
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
		GoodBehaviors []string `json:"good_behaviors"`
	}

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}
	behaviors, IDs := getGoodBehavior(params.GoodBehaviors)
	behaviorScore := behaviorScore(behaviors)

	data := schema.GoodBehaviorData{
		ProfileID:     account.Profile.ID.String(),
		AccountNumber: account.Profile.AccountNumber,
		GoodBehaviors: IDs,
		Location:      schema.GeoJSON{Type: "Point", Coordinates: []float64{loc.Longitude, loc.Latitude}},
		BehaviorScore: behaviorScore,
		Timestamp:     time.Now().Unix(),
	}

	err := s.mongoStore.GoodBehaviorSave(&data)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK"})

	return
}

func getGoodBehavior(behaviors []string) ([]schema.GoodBehavior, []string) {
	var retBehaviors []schema.GoodBehavior
	var reBehaviorsID []string
	for _, behavior := range behaviors {
		st := schema.GoodBehaviorType(behavior)
		v, ok := schema.GoodBehaviorFromID[st]
		if ok {
			retBehaviors = append(retBehaviors, v)
			reBehaviorsID = append(reBehaviorsID, string(v.ID))

		}
	}
	return retBehaviors, reBehaviorsID
}

func behaviorScore(behaviors []schema.GoodBehavior) float64 {
	var sum float64
	for _, behavior := range behaviors {
		sum = sum + float64(behavior.Weight)
	}
	return sum
}
