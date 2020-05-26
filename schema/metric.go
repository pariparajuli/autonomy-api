package schema

import (
	"time"
)

type ConfirmDetail struct {
	ContinuousData []CDSScoreDataSet `json:"-" bson:"data"`
	Score          float64           `json:"score" bson:"score"`
}

type BehaviorDetail struct {
	// FIXME: deprecate these fields
	BehaviorTotal           float64 `json:"behavior_total" bson:"behavior_total"`
	TotalPeople             float64 `json:"total_people" bson:"total_people"`
	MaxScorePerPerson       float64 `json:"max_score_per_person" bson:"max_score_per_person"`
	CustomizedBehaviorTotal float64 `json:"behavior_customized_total" bson:"behavior_customized_total"`

	Score                 float64        `json:"score" bson:"score"`
	ReportTimes           int            `json:"x" bson:"-"`
	TodayDistribution     map[string]int `json:"y" bson:"-"`
	YesterdayDistribution map[string]int `json:"-" bson:"-"`
}

type SymptomDetail struct {
	Score             float64            `json:"score" bson:"score"`
	SymptomTotal      float64            `json:"total_weight" bson:"total_weight"`
	TotalPeople       float64            `json:"total_people" bson:"total_people"`
	MaxScorePerPerson float64            `json:"max_weight" bson:"max_weight"`
	CustomizedWeight  float64            `json:"customized_weight" bson:"customized_weight"`
	TodayData         NearestSymptomData `json:"today_data"  bson:"today_data"`
	YesterdayData     NearestSymptomData `json:"-"  bson:"-"`
	LastSpikeUpdate   time.Time          `json:"-" bson:"last_spike_update"`
	LastSpikeList     []string           `json:"-" bson:"last_spike_types"`
}

type NearestSymptomData struct {
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
