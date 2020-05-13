package score

import (
	"testing"
)

type changeRateTestCase struct {
	new                float64
	old                float64
	expectedChangeRate float64
}

func TestChangeRate(t *testing.T) {
	cases := []changeRateTestCase{
		{0, 0, 0},
		{10, 10, 0},
		{0, 10, -100},
		{10, 0, 100},
		{3, 5, -40},
		{3, 2, 50},
	}
	for _, c := range cases {
		if ChangeRate(c.new, c.old) != c.expectedChangeRate {
			t.Fatal()
		}
	}
}
