package schema

type Metric struct {
	Confirm       float64 `json:"confirm" bson:"confirm"`
	ConfirmDelta  float64 `json:"confirm_delta" bson:"confirm_delta"`
	Symptoms      float64 `json:"symptoms" bson:"symptoms"`
	SymptomsDelta float64 `json:"symptoms_delta" bson:"symptoms_delta"`
	Behavior      float64 `json:"behavior" bson:"behavior"`
	BehaviorDelta float64 `json:"behavior_delta" bson:"behavior_delta"`
	Score         float64 `json:"score" bson:"score"`
	LastUpdate    int64   `bson:"last_update"`
}
