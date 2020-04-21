package score

import (
	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/store"
)

func TotalScore(symptomScore, behaviorScore, confirmedScore float64) float64 {
	return 0.25*symptomScore + 0.25*behaviorScore + 0.5*confirmedScore
}

func CalculateMetric(mongo store.MongoStore, location schema.Location) (*schema.Metric, error) {
	behaviorScore, _, behaviorCount, behaviorDelta, err := mongo.NearestGoodBehaviorScore(consts.CORHORT_DISTANCE_RANGE, location)
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

	totalScore := TotalScore(symptomScore, behaviorScore, confirmedScore)

	return &schema.Metric{
		Symptoms:      float64(symptomCount),
		SymptomsDelta: float64(symptomDelta),
		Behavior:      float64(behaviorCount),
		BehaviorDelta: float64(behaviorDelta),
		Confirm:       float64(confirmCount),
		ConfirmDelta:  float64(confirmDelta),
		Score:         totalScore,
	}, nil
}
