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

func CalculateMetric(rawMetrics schema.Metric) (*schema.Metric, error) {
	metric := rawMetrics
	totalScore := DefaultTotalScore(metric.SymptomScore, metric.BehaviorScore, metric.ConfirmedScore)
	metric.Score = totalScore

	return &metric, nil
}
