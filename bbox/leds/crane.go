package leds

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	// TODO: bbox testing
	// 2x structure
	CRANE_STRAND_COUNT1 = 5
	CRANE_STRAND_LEN1   = 30

	// 1x heart
	CRANE_STRAND_COUNT2 = 4
	CRANE_STRAND_LEN2   = 60

	CRANE_LED_COUNT1 = CRANE_STRAND_COUNT1 * CRANE_STRAND_LEN1 // 10*30 // * 2x(10) // 30/m
	CRANE_LED_COUNT2 = CRANE_STRAND_COUNT2 * CRANE_STRAND_LEN2 // 8*60 // * 1x(4 + 4) // 60/m

	BPM      = 36
	INTERVAL = 60 * time.Second / BPM / 2 // (36 beats/min) / 2 color transitions/beat

	LIGHT_COUNT = 18 // 2 x 18 == 36 total trueWhite lights turned on at a time

	STREAK_LENGTH = 30
	STREAK_STEP   = 0.1

	// TODO: unused?
	CRANE_COLOR_WEIGHT = 0.01
)

var (
	RED_LIGHT_RANGE = []uint32{0, 1, 2, 3, 4}
)

type Crane struct {
	ampLevel float64
	closing  chan struct{}
	level    <-chan float64
}

func InitCrane(level <-chan float64) *Crane {
	InitLeds(DEFAULT_FREQ, CRANE_LED_COUNT1, CRANE_LED_COUNT2)

	return &Crane{
		closing: make(chan struct{}),
		level:   level,
	}
}

func (c *Crane) Run() {
	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	strand1 := make([]uint32, CRANE_LED_COUNT1) // structure
	strand2 := make([]uint32, CRANE_LED_COUNT2) // heart

	// 18 lights at trueWhite (36 total between two strands)
	lights := make([]uint32, LIGHT_COUNT)
	for i, _ := range lights {
		r := uint32(rand.Int31n(CRANE_LED_COUNT1))
		for Contains(append(lights, RED_LIGHT_RANGE...), r) {
			r = uint32(rand.Int31n(CRANE_LED_COUNT1))
		}
		lights[i] = r
	}

	lightIter := 0
	nextLight := uint32(rand.Int31n(CRANE_LED_COUNT1))
	for Contains(append(lights, RED_LIGHT_RANGE...), nextLight) {
		nextLight = uint32(rand.Int31n(CRANE_LED_COUNT1))
	}

	streakLoc := 0.0

	heartColor1 := TrueRed
	heartColor2 := trueWhite

	last := time.Now()

	for {
		select {
		case level, more := <-c.level:
			if more {
				c.ampLevel = level
			} else {
				return
			}
		case _, more := <-c.closing:
			if !more {
				return
			}
		default:
			_ = uint32(255.0 * c.ampLevel)

			now := time.Now()
			weight := float64(now.Sub(last).Nanoseconds()) / float64(INTERVAL.Nanoseconds())
			if weight > 1 {
				weight = 0
				last = now

				// structure iters
				lights[lightIter] = nextLight
				lightIter = (lightIter + 1) % len(lights)

				nextLight = uint32(rand.Int31n(CRANE_LED_COUNT1))
				for Contains(append(lights, RED_LIGHT_RANGE...), nextLight) {
					nextLight = uint32(rand.Int31n(CRANE_LED_COUNT1))
				}

				// heart iters
				tmp := heartColor1
				heartColor1 = heartColor2
				heartColor2 = tmp
			}

			// structure
			for i, light := range lights {
				if i == lightIter {
					strand1[light] = MkColorWeight(trueWhite, black, weight)
					strand1[nextLight] = MkColorWeight(black, trueWhite, weight)
				} else {
					strand1[light] = trueWhite
				}
			}

			// streaks
			sineMap := GetSineVals(CRANE_LED_COUNT1, streakLoc, STREAK_LENGTH)
			for led, value := range sineMap {
				if !Contains(append(lights, RED_LIGHT_RANGE...), uint32(led)) {
					mag := float64(value) / 254.0
					strand1[led] = MkColor(0, uint32(float64(123)*mag), uint32(float64(55)*mag), 0)
				}
			}

			streakLoc += STREAK_STEP
			if streakLoc >= CRANE_LED_COUNT1 {
				streakLoc = 0
			}

			// heart
			heartColor := MkColorWeight(heartColor1, heartColor2, weight)
			for i := 0; i < CRANE_LED_COUNT2; i++ {
				strand2[i] = heartColor
			}

			ws2811.SetBitmap(0, strand1)
			ws2811.SetBitmap(1, strand2)

			err := ws2811.Render()
			if err != nil {
				fmt.Printf("ws2811.Render failed: %+v\n", err)
				panic(err)
			}

			err = ws2811.Wait()
			if err != nil {
				fmt.Printf("ws2811.Wait failed: %+v\n", err)
				panic(err)
			}

			// if weight < 1 {
			// 	weight += CRANE_COLOR_WEIGHT
			// } else {
			// 	weight = 0
			// 	iter = (iter + 1) % len(colors)
			// }
		}
	}
}

func (c *Crane) Close() {
	// TODO: this doesn't block?
	close(c.closing)
}
