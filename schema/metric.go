package schema

type ConfirmDetail struct {
	Yesterday float64 `json:"yesterday" bson:"yesterday"`
	Today     float64 `json:"today" bson:"today"`
	Score     float64 `json:"score" bson:"score"`
}

type BehaviorDetail struct {
	BehaviorTotal           float64 `json:"behavior_total" bson:"behavior_total"`
	TotalPeople             float64 `json:"total_people" bson:"total_people"`
	MaxScorePerPerson       float64 `json:"max_score_per_person" bson:"max_score_per_person"`
	CustomizedBehaviorTotal float64 `json:"behavior_customized_total" bson:"behavior_customized_total`
	Score                   float64 `json:"score" bson:"score"`
}

type SymptomDetail struct {
	SymptomTotal       float64             `json:"total_weight" bson:"total_weight"`
	TotalPeople        float64             `json:"total_people" bson:"total_people"`
	MaxScorePerPerson  float64             `json:"max_weight" bson:"max_weight"`
	CustomizedWeight   float64             `json:"customized_weight" bson:"customized_weight"`
	Score              float64             `json:"score" bson:"score"`
	Symptoms           SymptomDistribution `json:"-" bson:"symptoms"`
	CustomSymptomCount float64             `json:"custom_symptom_count" bson:"custom_symptom_count"`
	TodayData          NearestSymptomData  `json:"-"  bson:"-"`
	YesterdayData      NearestSymptomData  `json:"-"  bson:"-"`
}
type NearestSymptomData struct {
	UserCount          float64             `json:"userCount" bson:"userCount`
	OfficialCount      float64             `json:"officialCount" bson:"officialCount"`
	CustomizedCount    float64             `json:"customizedCount" bson:"customizedCount"`
	WeightDistribution SymptomDistribution `json:"weight_distribution" beson:"weight_distribution"`
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
