package score

func TotalScore(symptomScore, behaviorScore, confirmedScore float64) float64 {
	return 0.25*symptomScore + 0.25*behaviorScore + 0.5*confirmedScore
}
