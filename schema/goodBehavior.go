package schema

import (
	"encoding/json"
)

type GoodBehaviorType string

// DefaultBehaviorMatrix is a map which key is GoodBehavior.ID and value is a object of GoodBehavior
var DefaultBehaviorMatrix = map[GoodBehaviorType]DefaultBehavior{
	GoodBehaviorType(CleanHand):        DefaultBehaviors[0],
	GoodBehaviorType(SocialDistancing): DefaultBehaviors[1],
	GoodBehaviorType(TouchFace):        DefaultBehaviors[2],
	GoodBehaviorType(WearMask):         DefaultBehaviors[3],
	GoodBehaviorType(CoveringCough):    DefaultBehaviors[4],
	GoodBehaviorType(CleanSurface):     DefaultBehaviors[5],
}

// DefaultBehaviorWeightMatrix is a map which key is GoodBehavior.ID and value is a object of GoodBehavior
var DefaultBehaviorWeightMatrix = map[GoodBehaviorType]BehaviorWeight{
	GoodBehaviorType(CleanHand):        BehaviorWeight{ID: GoodBehaviorType(CleanHand), Weight: 1},
	GoodBehaviorType(SocialDistancing): BehaviorWeight{ID: GoodBehaviorType(SocialDistancing), Weight: 1},
	GoodBehaviorType(TouchFace):        BehaviorWeight{ID: GoodBehaviorType(TouchFace), Weight: 1},
	GoodBehaviorType(WearMask):         BehaviorWeight{ID: GoodBehaviorType(WearMask), Weight: 1},
	GoodBehaviorType(CoveringCough):    BehaviorWeight{ID: GoodBehaviorType(CoveringCough), Weight: 1},
	GoodBehaviorType(CleanSurface):     BehaviorWeight{ID: GoodBehaviorType(CleanSurface), Weight: 1},
}

const (
	GoodBehaviorCollection  = "goodBehavior"
	TotalGoodBehaviorWeight = 6
)

const (
	CleanHand        GoodBehaviorType = "clean_hand"
	SocialDistancing GoodBehaviorType = "social_distancing"
	TouchFace        GoodBehaviorType = "touch_face"
	WearMask         GoodBehaviorType = "wear_mask"
	CoveringCough    GoodBehaviorType = "covering_coughs"
	CleanSurface     GoodBehaviorType = "clean_surface"
)

// DefaultBehavior a struct to define a good behavior
type DefaultBehavior struct {
	ID   GoodBehaviorType `json:"id"`
	Name string           `json:"name"`
	Desc string           `json:"desc"`
}

// SelfDefinedBehavior a struct to define a self-defined good behavior
type SelfDefinedBehavior struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// BehaviorWeight a struct to define a good behavior weight
type BehaviorWeight struct {
	ID     GoodBehaviorType `json:"id"`
	Weight float64          `json:"weight"`
}

// DefaultBehaviors return a slice that contains all GoodBehavior
var DefaultBehaviors = []DefaultBehavior{
	{CleanHand, "Frequent hand cleaning", "Washing hands thoroughly with soap and water for at least 20 seconds or applying an alcohol-based hand sanitizer"},
	{SocialDistancing, "Social & physical distancing", "Avoiding crowds, working from home, and maintaining at least 6 feet of distance from others whenever possibl"},
	{TouchFace, "Avoiding touching face", "Restraining from touching your eyes, nose, or mouth, especially with unwashed hands."},
	{WearMask, "Wearing a face mask or covering", "Covering your nose and mouth when in public or whenever social distancing measures are difficult to maintain."},
	{CoveringCough, "Covering coughs and sneezes", "Covering your mouth with the inside of your elbow or a tissue whenever you cough or sneeze."},
	{CleanSurface, "Cleaing and disinfecting surfaces", "Cleaning and disinfecting frequently touched surfaces daily, such as doorknobs, tables, light switches, and keyboards."},
}

// GoodBehaviorData the struct to store citizen data and score
type GoodBehaviorData struct {
	ProfileID            string                `json:"profile_id" bson:"profile_id"`
	AccountNumber        string                `json:"account_number" bson:"account_number"`
	DefaultBehaviors     []DefaultBehavior     `json:"default_behaviors" bson:"default_behaviors"`
	SelfDefinedBehaviors []SelfDefinedBehavior `json:"self_defined_behaviors" bson:"self_defined_behaviors"`
	Location             GeoJSON               `json:"location" bson:"location"`
	DefaultWeight        float64               `json:"default_weight" bson:"default_weight"`
	SelfDefinedWeight    float64               `json:"self_defined_weight" bson:"self_defined_weight"`
	Timestamp            int64                 `json:"ts" bson:"ts"`
}

func (b *GoodBehaviorData) MarshalJSON() ([]byte, error) {
	behaviors := b.GoodBehaviors
	if b.GoodBehaviors == nil {
		behaviors = make([]string, 0)
	}
	return json.Marshal(&struct {
		GoodBehaviors []string `json:"behaviors"`
		Location      Location `json:"location"`
		Timestamp     int64    `json:"timestamp"`
	}{
		GoodBehaviors: behaviors,
		Location:      Location{Longitude: b.Location.Coordinates[0], Latitude: b.Location.Coordinates[1]},
		Timestamp:     b.Timestamp,
	})
}
