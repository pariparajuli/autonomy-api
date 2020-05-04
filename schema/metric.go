package schema

type ConfirmDetail struct {
	Yesterday float64 `json:"yesterday" bson:"yesterday"`
	Today     float64 `json:"today" bson:"today"`
	Score     float64 `json:"score" bson:"score"`
}

type BehaviorDetail struct {
	BehaviorTotal     float64 `json:"behavior_total" bson:"behavior_total"`
	TotalPeople       float64 `json:"total_people" bson:"total_people"`
	MaxScorePerPerson float64 `json:"max_score_per_person" bson:"max_score_per_person"`
	Score             float64 `json:"score" bson:"score"`
}

type SymptomDetail struct {
	SymptomTotal       float64             `json:"symptom_total" bson:"symptom_total"`
	TotalPeople        float64             `json:"total_people" bson:"total_people"`
	MaxScorePerPerson  float64             `json:"max_score_per_person" bson:"max_score_per_person"`
	Score              float64             `json:"score" bson:"score"`
	Symptoms           SymptomDistribution `json:"-" bson:"symptoms"`
	CustomSymptomCount float64             `json:"-" bson:"custom_symptom_count"`
}

type Details struct {
	Confirm   ConfirmDetail  `json:"confirm" bson:"confirm"`
	Behaviors BehaviorDetail `json:"behaviors" bson:"behaviors"`
	Symptoms  SymptomDetail  `json:"symptoms" bson:"symptoms"`
}

type Metric struct {
	ConfirmedCount float64 `json:"confirm" bson:"confirm"`
	ConfirmedDelta float64 `json:"confirm_delta" bson:"confirm_delta"`
	SymptomCount   float64 `json:"symptoms" bson:"symptoms"`
	SymptomDelta   float64 `json:"symptoms_delta" bson:"symptoms_delta"`
	BehaviorCount  float64 `json:"behavior" bson:"behavior"`
	BehaviorDelta  float64 `json:"behavior_delta" bson:"behavior_delta"`
	Score          float64 `json:"score" bson:"score"`
	LastUpdate     int64   `bson:"last_update"`
	Details        Details `json:"details" bson:"details"`
}
