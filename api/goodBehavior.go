package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"github.com/bitmark-inc/autonomy-api/schema"
)

// BehaviorResp  respons struct of a good behavior
type BehaviorResp struct {
	ID   schema.GoodBehaviorType `json:"id"`
	Name string                  `json:"name"`
	Desc string                  `json:"desc"`
}

func (s *Server) goodBehaviors(c *gin.Context) {
	resp := []BehaviorResp{}
	for _, behavior := range schema.GoodBehaviors {
		respBehavior := BehaviorResp{ID: behavior.ID, Name: behavior.Name, Desc: behavior.Desc}
		resp = append(resp, respBehavior)
	}
	c.JSON(http.StatusOK, gin.H{"symptoms": resp})
}

func (s *Server) reportBehaviors(c *gin.Context) {
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
		GoodBehaviors []string `json:"behaviros"`
	}

	if err := c.BindJSON(&params); err != nil {
		abortWithEncoding(c, http.StatusBadRequest, errorInvalidParameters, err)
		return
	}
	schemaBehaviors := convertBehavior(params.GoodBehaviors)
	behaviorScore := behaviorScore(schemaBehaviors)
	data := schema.GoodBehaviorData{
		AccountNumber: account.Profile.AccountNumber,
		GoodBehaviors: schemaBehaviors,
		Location:      schema.GeoJSON{Type: "Point", Coordinates: []float64{loc.Latitude, loc.Longitude}},
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

func convertBehavior(ids []string) []schema.GoodBehavior {
	m := make(map[schema.GoodBehaviorType]schema.GoodBehavior)
	var ret []schema.GoodBehavior
	for _, s := range schema.GoodBehaviors {
		m[s.ID] = s
	}
	for _, id := range ids {
		st := schema.GoodBehaviorType(id)
		sy, ok := m[st]
		if ok {
			ret = append(ret, sy)
		}
	}
	return ret
}

func behaviorScore(behaviors []schema.GoodBehavior) float64 {
	return 0
}
