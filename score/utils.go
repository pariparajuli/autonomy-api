package score

func ChangeRate(new, old float64) float64 {
	rate := float64(1)
	if old > 0 {
		rate = (new - old) / old
	}
	return rate
}
