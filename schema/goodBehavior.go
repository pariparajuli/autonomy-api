package schema

type GoodBehaviorType string

const (
	GoodBehaviorCollectionName = "goodBehavior"
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
	Weight float64          `json:"weight"`
}

// GoodBehaviors return a slice that contains all GoodBehavior
var GoodBehaviors = []GoodBehavior{
	{CleanHand, "Frequent hand cleaning", "Washing hands thoroughly with soap and water for at least 20 seconds or applying an alcohol-based hand sanitizer", 2},
	{SocialDistancing, "Social & physical distancing", "Washing hands thoroughly with soap and water for at least 20 seconds or applying an alcohol-based hand sanitizer.", 1},
	{TouchFace, "Avoiding touching face", "Avoiding crowds, working from home, and maintaining at least 6 feet of distance from others whenever possible.", 2},
	{WearMask, "Wearing a face mask or covering", "Covering your nose and mouth when in public or whenever social distancing measures are difficult to maintain.", 1},
	{CoveringCough, "Covering coughs and sneezes", "Covering your mouth with the inside of your elbow or a tissue whenever you cough or sneeze.", 1},
	{CleanSurface, "Cleaing and disinfecting surfaces", "TBD", 1},
}

// GoodBehaviorData the struct to store citizen data and score
type GoodBehaviorData struct {
	AccountNumber string         `json:"account_number" bson:"account_number"`
	GoodBehaviors []GoodBehavior `json:"symptoms" bson:"symptoms"`
	Location      GeoJSON        `json:"location" bson:"location"`
	BehaviorScore float64        `json:"health_score" bson:"health_score"`
	Timestamp     int64          `json:"ts" bson:"ts"`
}
