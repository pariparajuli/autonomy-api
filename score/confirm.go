package score

import (
	"math"

	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
)

func CalculateConfirmScore(metric *schema.Metric) {
	details := &metric.Details.Confirm
	dataset := details.ContinuousData
	score := float64(0)
	sizeOfConfirmData := len(dataset)
	if 0 == len(dataset) {
		metric.Details.Confirm.Score = 0
		return
	} else if len(dataset) < consts.ConfirmScoreWindowSize {
		zeroDay := []schema.CDSScoreDataSet{schema.CDSScoreDataSet{Name: dataset[0].Name, Cases: 0}}
		for idx := 0; idx < consts.ConfirmScoreWindowSize-sizeOfConfirmData; idx++ {
			dataset = append(zeroDay, dataset...)
		}
		details.ContinuousData = dataset
	}
	numerator := float64(0)
	denominator := float64(0)
	for idx, val := range dataset {
		power := (float64(idx) + 1) / 2
		numerator = numerator + math.Exp(power)*val.Cases
		denominator = denominator + math.Exp(power)*(val.Cases+1)
	}

	if denominator > 0 {
		score = 1 - numerator/denominator
	}
	metric.Details.Confirm.Score = score * 100
}
