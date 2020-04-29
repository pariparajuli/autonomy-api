package schema

type GoodBehaviorType string

// OfficialBehaviorMatrix is a map which key is GoodBehavior.ID and value is a object of GoodBehavior
var OfficialBehaviorMatrix = map[GoodBehaviorType]Behavior{
	GoodBehaviorType(CleanHand):        OfficialBehaviors[0],
	GoodBehaviorType(SocialDistancing): OfficialBehaviors[1],
	GoodBehaviorType(TouchFace):        OfficialBehaviors[2],
	GoodBehaviorType(WearMask):         OfficialBehaviors[3],
	GoodBehaviorType(CoveringCough):    OfficialBehaviors[4],
	GoodBehaviorType(CleanSurface):     OfficialBehaviors[5],
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
	BehaviorCollection          = "behaviors"
	BehaviorReportCollection    = "behaviorReport"
	TotalOfficialBehaviorWeight = float64(6)
)

type BehaviorSource string

const (
	OfficialBehavior   BehaviorSource = "official"
	CustomizedBehavior BehaviorSource = "customized"
)

const (
	CleanHand        GoodBehaviorType = "clean_hand"
	SocialDistancing GoodBehaviorType = "social_distancing"
	TouchFace        GoodBehaviorType = "touch_face"
	WearMask         GoodBehaviorType = "wear_mask"
	CoveringCough    GoodBehaviorType = "covering_coughs"
	CleanSurface     GoodBehaviorType = "clean_surface"
)

// Behavior a struct to define a good behavior
type Behavior struct {
	ID     GoodBehaviorType `json:"id" bson:"_id"`
	Name   string           `json:"name"  bson:"name"`
	Desc   string           `json:"desc"  bson:"desc"`
	Source BehaviorSource   `json:"-" bson:"source"`
}

// BehaviorWeight a struct to define a good behavior weight
type BehaviorWeight struct {
	ID     GoodBehaviorType `json:"id"`
	Weight float64          `json:"weight"`
}

// OfficialBehaviors return a slice that contains all GoodBehavior
var OfficialBehaviors = []Behavior{
	{CleanHand, "Frequent hand cleaning", "Washing hands thoroughly with soap and water for at least 20 seconds or applying an alcohol-based hand sanitizer", OfficialBehavior},
	{SocialDistancing, "Social & physical distancing", "Avoiding crowds, working from home, and maintaining at least 6 feet of distance from others whenever possibl", OfficialBehavior},
	{TouchFace, "Avoiding touching face", "Restraining from touching your eyes, nose, or mouth, especially with unwashed hands.", OfficialBehavior},
	{WearMask, "Wearing a face mask or covering", "Covering your nose and mouth when in public or whenever social distancing measures are difficult to maintain.", OfficialBehavior},
	{CoveringCough, "Covering coughs and sneezes", "Covering your mouth with the inside of your elbow or a tissue whenever you cough or sneeze.", OfficialBehavior},
	{CleanSurface, "Cleaing and disinfecting surfaces", "Cleaning and disinfecting frequently touched surfaces daily, such as doorknobs, tables, light switches, and keyboards.", OfficialBehavior},
}

// BehaviorReportData the struct to store citizen data and score
type BehaviorReportData struct {
	ProfileID           string     `json:"profile_id" bson:"profile_id"`
	AccountNumber       string     `json:"account_number" bson:"account_number"`
	OfficialBehaviors   []Behavior `json:"official_behaviors" bson:"official_behaviors"`
	CustomizedBehaviors []Behavior `json:"customized_behaviors" bson:"customized_behaviors"`
	Location            GeoJSON    `json:"location" bson:"location"`
	OfficialWeight      float64    `json:"official_weight" bson:"official_weight"`
	CustomizedWeight    float64    `json:"customized_weight" bson:"customized_weight"`
	Timestamp           int64      `json:"ts" bson:"ts"`
}
