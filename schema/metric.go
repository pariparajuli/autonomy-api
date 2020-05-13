package schema

import "time"

type ConfirmDetail struct {
	Yesterday float64 `json:"yesterday" bson:"yesterday"`
	Today     float64 `json:"today" bson:"today"`
	Score     float64 `json:"score" bson:"score"`
}

type BehaviorDetail struct {
	BehaviorTotal           float64 `json:"behavior_total" bson:"behavior_total"`
	TotalPeople             float64 `json:"total_people" bson:"total_people"`
	MaxScorePerPerson       float64 `json:"max_score_per_person" bson:"max_score_per_person"`
	CustomizedBehaviorTotal float64 `json:"behavior_customized_total" bson:"behavior_customized_total"`
	Score                   float64 `json:"score" bson:"score"`
}

type SymptomDetail struct {
	SymptomTotal       float64             `json:"total_weight" bson:"total_weight"`
	TotalPeople        float64             `json:"total_people" bson:"total_people"`
	MaxScorePerPerson  float64             `json:"max_weight" bson:"max_weight"`
	CustomizedWeight   float64             `json:"customized_weight" bson:"customized_weight"`
	Score              float64             `json:"score" bson:"score"`
	Symptoms           SymptomDistribution `json:"-" bson:"symptoms"`
	LastSpikeUpdate    time.Time           `json:"last_spike_update" bson:"last_spike_update"`
	LastSpikeList      []SymptomType       `json:"last_spike_types" bson:"last_spike_types"`
	CustomSymptomCount float64             `json:"-" bson:"custom_symptom_count"`
	TodayData          NearestSymptomData  `json:"today_data"  bson:"today_data"`
	YesterdayData      NearestSymptomData  `json:"yesterday_data"  bson:"-"`
}
type NearestSymptomData struct {
	UserCount          float64             `json:"user_count" bson:"user_count"`
	OfficialCount      float64             `json:"official_count" bson:"official_count"`
	CustomizedCount    float64             `json:"customized_count" bson:"customized_count"`
	WeightDistribution SymptomDistribution `json:"weight_distribution" bson:"weight_distribution"`
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
	LastUpdate     int64   `json:"last_update" bson:"last_update"`
	Details        Details `json:"details" bson:"details"`
}
