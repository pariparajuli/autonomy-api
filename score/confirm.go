package score

import (
	"fmt"
	"math"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	confirmCriteria = 10
)

func CalculateConfirmScore(metric *schema.Metric) {
	details := &metric.Details.Confirm
	dataset := details.ContinuousData
	score := float64(0)
	if len(dataset) < 14 {
		metric.Details.Confirm.Score = score
		return
	}
	fraction := float64(0)
	denominator := float64(0)
	for idx, val := range dataset {
		epx := (float64(idx) + 1) / 2
		fraction = fraction + math.Exp(epx)*val.Cases
		denominator = denominator + math.Exp(epx)*(val.Cases+1)
	}

	if denominator > 0 {
		score = 1 - fraction/denominator
	}
	metric.Details.Confirm.Score = score * 100
	fmt.Println(fmt.Sprintf("score in percentage:%f     fraction:%f denominator:%f score in fraction:%f", metric.Details.Confirm.Score, fraction, denominator, score))
}
