package schema

type GoodBehaviorType string

// GoodBehaviorFromID is a map which key is GoodBehavior.ID and value is a object of GoodBehavior
var GoodBehaviorFromID = map[GoodBehaviorType]GoodBehavior{
	GoodBehaviorType(CleanHand):        GoodBehaviors[0],
	GoodBehaviorType(SocialDistancing): GoodBehaviors[1],
	GoodBehaviorType(TouchFace):        GoodBehaviors[2],
	GoodBehaviorType(WearMask):         GoodBehaviors[3],
	GoodBehaviorType(CoveringCough):    GoodBehaviors[4],
	GoodBehaviorType(CleanSurface):     GoodBehaviors[5],
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

// GoodBehavior a struct to define a good behavior
type GoodBehavior struct {
	ID     GoodBehaviorType `json:"id"`
	Name   string           `json:"name"`
	Desc   string           `json:"desc"`
	Weight float64          `json:"-"`
}

// GoodBehaviors return a slice that contains all GoodBehavior
var GoodBehaviors = []GoodBehavior{
	{CleanHand, "Frequent hand cleaning", "Washing hands thoroughly with soap and water for at least 20 seconds or applying an alcohol-based hand sanitizer", 1},
	{SocialDistancing, "Social & physical distancing", "Washing hands thoroughly with soap and water for at least 20 seconds or applying an alcohol-based hand sanitizer.", 1},
	{TouchFace, "Avoiding touching face", "Avoiding crowds, working from home, and maintaining at least 6 feet of distance from others whenever possible.", 1},
	{WearMask, "Wearing a face mask or covering", "Covering your nose and mouth when in public or whenever social distancing measures are difficult to maintain.", 1},
	{CoveringCough, "Covering coughs and sneezes", "Covering your mouth with the inside of your elbow or a tissue whenever you cough or sneeze.", 1},
	{CleanSurface, "Cleaing and disinfecting surfaces", "Cleaing and disinfecting surfaces", 1},
}

// GoodBehaviorData the struct to store citizen data and score
type GoodBehaviorData struct {
	ProfileID     string   `json:"profile_id" bson:"profile_id"`
	AccountNumber string   `json:"account_number" bson:"account_number"`
	GoodBehaviors []string `json:"good_behaviors" bson:"good_behaviors"`
	Location      GeoJSON  `json:"location" bson:"location"`
	BehaviorScore float64  `json:"behavior_score" bson:"behavior_score"`
	Timestamp     int64    `json:"ts" bson:"ts"`
}
