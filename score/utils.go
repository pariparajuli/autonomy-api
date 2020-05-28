package score

func ChangeRate(new, old float64) float64 {
	if old == 0 {
		if new == 0 {
			return float64(0)
		} else {
			return float64(100)
		}
	}

	return (new - old) / old * 100
}

func DivOrDefault(numerator, denominator, def float64) float64 {
	if denominator == 0 {
		return def
	}
	return numerator / denominator
}
