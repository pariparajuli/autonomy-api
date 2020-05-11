package score

import "github.com/bitmark-inc/autonomy-api/schema"

const (
	confirmCriteria = 10
)

func CalculateConfirmScore(metric *schema.Metric) {
	details := &metric.Details.Confirm
	delta := details.Today - details.Yesterday

	// equal or more than that criteria get 0 score
	if delta >= confirmCriteria {
		details.Score = 0
		return
	}

	// in between 0 - 10 people, each increased confirm case deduct 5 points
	details.Score = 100 - 5*delta
}
