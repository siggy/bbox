package leds

import "math"

const (
	sineAmplitude = 127
	sineShift     = 127
)

func GetSineVals(ledCount int, floatBeat float64, period int) (sineVals map[int]int) {
	halfPeriod := float64(period) / 2.0

	first := int(math.Ceil(floatBeat - halfPeriod)) // 12.7 - 1.5 => 11.2 => 12
	last := int(math.Floor(floatBeat + halfPeriod)) // 12.7 + 1.5 => 14.2 => 14

	sineFunc := func(x int) int {
		// y = a * sin((x-h)/b) + k
		h := floatBeat - float64(period)/4.0
		b := float64(period) / (2 * math.Pi)
		return int(
			sineAmplitude*math.Sin((float64(x)-h)/b) +
				sineShift,
		)
	}

	sineVals = make(map[int]int)

	for i := first; i <= last; i++ {
		y := sineFunc(i)
		if y != 0 {
			sineVals[(i+ledCount)%ledCount] = int(scale(uint32(sineFunc(i))))
		}
	}

	return
}

// maps midpoint 128 => 32 for brightness
func scale(x uint32) uint32 {
	// y = 1000*(0.005333 * 4002473^(x/1000)-0.005333)
	return uint32(1000 * (0.005333*math.Pow(4002473., float64(x)/1000.) - 0.005333))
}
