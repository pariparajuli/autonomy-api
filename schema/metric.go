package schema

type Metric struct {
	Confirm       float64 `json:"confirm" bson:"confirm"`
	ConfirmDelta  float64 `json:"confirm_delta" bson:"confirm_delta"`
	ConfirmScore  float64 `json:"confirm_score" bson:"confirm_score"`
	Symptoms      float64 `json:"symptoms" bson:"symptoms"`
	SymptomsDelta float64 `json:"symptoms_delta" bson:"symptoms_delta"`
	SymptomScore  float64 `json:"symptom_score" bson:"symptom_score"`
	Behavior      float64 `json:"behavior" bson:"behavior"`
	BehaviorDelta float64 `json:"behavior_delta" bson:"behavior_delta"`
	BehaviorScore float64 `json:"behavior_score" bson:"behavior_score"`
	Score         float64 `json:"score" bson:"score"`
	LastUpdate    int64   `bson:"last_update"`
}
