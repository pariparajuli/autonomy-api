package schema

import (
	"time"
)

type ConfirmDetail struct {
	ContinuousData []CDSScoreDataSet `json:"-" bson:"data"`
	Score          float64           `json:"score" bson:"score"`
}

type BehaviorDetail struct {
	Score                 float64        `json:"score" bson:"score"`
	ReportTimes           int            `json:"-" bson:"-"`
	TodayDistribution     map[string]int `json:"-" bson:"-"`
	YesterdayDistribution map[string]int `json:"-" bson:"-"`
}

type SymptomDetail struct {
	Score           float64            `json:"score" bson:"score"`
	TotalPeople     float64            `json:"-" bson:"-"`
	TodayData       NearestSymptomData `json:"-"  bson:"-"`
	YesterdayData   NearestSymptomData `json:"-"  bson:"-"`
	LastSpikeUpdate time.Time          `json:"-" bson:"last_spike_update"`
	LastSpikeList   []string           `json:"-" bson:"last_spike_types"`
}

type NearestSymptomData struct {
	WeightDistribution SymptomDistribution `json:"weight_distribution" bson:"weight_distribution"`
}

type Details struct {
	Confirm   ConfirmDetail  `json:"confirm" bson:"confirm"`
	Behaviors BehaviorDetail `json:"behaviors" bson:"behaviors"`
	Symptoms  SymptomDetail  `json:"symptoms" bson:"symptoms"`
}

type IndividualMetric struct {
	SymptomCount  float64 `json:"symptoms" bson:"symptoms"`
	SymptomDelta  float64 `json:"symptoms_delta" bson:"symptoms_delta"`
	BehaviorCount float64 `json:"behavior" bson:"behavior"`
	BehaviorDelta float64 `json:"behavior_delta" bson:"behavior_delta"`
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
