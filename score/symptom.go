package score

import (
	"github.com/bitmark-inc/autonomy-api/schema"
)

func SymptomScore(weights schema.SymptomWeights, metric, oldMetric *schema.Metric) {
	details := &metric.Details.Symptoms
	for k, v := range schema.DefaultSymptomWeights {
		// in case key is missing in custom weights
		if cv, ok := weights[k]; ok {
			details.MaxScorePerPerson += cv
		} else {
			details.MaxScorePerPerson += v
		}
	}
	//	log.Info(fmt.Sprintf("SymptomScore : metrics:%v", len(details.Symptoms)))
	details.MaxScorePerPerson += details.CustomSymptomCount

	var w float64
	var ok bool
	for k, v := range details.Symptoms {
		if w, ok = weights[k]; !ok {
			w = schema.DefaultSymptomWeights[k]
		}
		details.SymptomTotal += w * float64(v)
		//log.Info(fmt.Sprintf("SymptomScore : k :%v w :%v , v:%v, symptomTotal:%v", k, w, v, details.SymptomTotal))
	}

	// update score
	if details.TotalPeople > 0 && (details.MaxScorePerPerson) > 0 {
		metric.SymptomCount = metric.SymptomCount
		details.Score = 100 - 100*(details.SymptomTotal/(details.TotalPeople*details.MaxScorePerPerson))
	}

	// update delta
	if oldMetric != nil && metric.SymptomCount > 0 {
		metric.SymptomDelta = (metric.SymptomCount - oldMetric.SymptomCount) / metric.SymptomCount
	}
}
