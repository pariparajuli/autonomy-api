package score

import (
	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/store"
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

func CalculateMetric(mongo store.MongoStore, location schema.Location) (*schema.Metric, error) {
	behaviorScore, _, behaviorCount, behaviorDelta, err := behaviorScore(mongo, consts.CORHORT_DISTANCE_RANGE, location)
	if err != nil {
		return nil, err
	}

	symptomScore, _, symptomCount, symptomDelta, err := mongo.NearestSymptomScore(consts.CORHORT_DISTANCE_RANGE, location)
	if err != nil {
		return nil, err
	}

	confirmCount, confirmDelta, err := mongo.GetConfirm(location)
	if err != nil {
		return nil, err
	}

	confirmedScore, err := mongo.ConfirmScore(location)
	if err != nil {
		return nil, err
	}

	totalScore := DefaultTotalScore(symptomScore, behaviorScore, confirmedScore)

	return &schema.Metric{
		Symptoms:      float64(symptomCount),
		SymptomsDelta: float64(symptomDelta),
		SymptomScore:  float64(symptomScore),
		Behavior:      float64(behaviorCount),
		BehaviorDelta: float64(behaviorDelta),
		BehaviorScore: float64(behaviorScore),
		Confirm:       float64(confirmCount),
		ConfirmDelta:  float64(confirmDelta),
		ConfirmScore:  float64(confirmedScore),
		Score:         totalScore,
	}, nil
}
