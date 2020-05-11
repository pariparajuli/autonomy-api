package score

import (
	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	DefaultScoreV1SymptomCoefficient  = 0.25
	DefaultScoreV1BehaviorCoefficient = 0.25
	DefaultScoreV1ConfirmCoefficient  = 0.5
)

func DefaultTotalScore(symptomScore, behaviorScore, confirmedScore float64) float64 {
	return TotalScoreV1(schema.ScoreCoefficient{
		Symptoms:  DefaultScoreV1SymptomCoefficient,
		Behaviors: DefaultScoreV1BehaviorCoefficient,
		Confirms:  DefaultScoreV1ConfirmCoefficient,
	},
		symptomScore,
		behaviorScore,
		confirmedScore)
}

func TotalScoreV1(c schema.ScoreCoefficient, symptomScore, behaviorScore, confirmedScore float64) float64 {
	return c.Symptoms*symptomScore + c.Behaviors*behaviorScore + c.Confirms*confirmedScore
}

// CheckScoreColorChange check if the color of a score need to be changed.
// Currently,
// Red:     0 ~ 33
// Yellow: 34 ~ 66
// Green:  67 ~ 100
func CheckScoreColorChange(oldScore, newScore float64) bool {
	oldScoreMod := (int(oldScore) - 1) / 33
	newScoreMod := (int(newScore) - 1) / 33

	// for case score is 100, set the value to 2
	if oldScoreMod == 3 {
		oldScoreMod = 2
	}
	if newScoreMod == 3 {
		newScoreMod = 2
	}
	return oldScoreMod != newScoreMod
}

// CalculateMetric will calculate, summarize and return a metric based on collected raw metrics
func CalculateMetric(rawMetrics schema.Metric, coefficient *schema.ScoreCoefficient) schema.Metric {
	metric := rawMetrics
	CalculateConfirmScore(&metric)

	if coefficient != nil {
		metric = CalculateSymptomScore(coefficient.SymptomWeights, metric)
		metric.Score = TotalScoreV1(*coefficient, metric.Details.Symptoms.Score, metric.Details.Behaviors.Score, metric.Details.Confirm.Score)
	} else {
		metric = CalculateSymptomScore(schema.DefaultSymptomWeights, metric)
		metric.Score = DefaultTotalScore(metric.Details.Symptoms.Score, metric.Details.Behaviors.Score, metric.Details.Confirm.Score)
	}

	return metric
}
