package leds

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	// 2x structure
	CRANE_STRAND_COUNT1 = 10
	CRANE_STRAND_LEN1   = 30

	// 1x heart
	CRANE_STRAND_COUNT2 = 4 + 4
	CRANE_STRAND_LEN2   = 60

	CRANE_LED_COUNT1 = CRANE_STRAND_COUNT1 * CRANE_STRAND_LEN1 // 10*30 // * 2x(10) // 30/m
	CRANE_LED_COUNT2 = CRANE_STRAND_COUNT2 * CRANE_STRAND_LEN2 // 8*60 // * 1x(4 + 4) // 60/m

	BPM      = 36
	INTERVAL = 60 * time.Second / BPM / 2 // (36 beats/min) / 2 color transitions/beat

	LIGHT_COUNT = 18 // 2 x 18 == 36 total trueWhite lights turned on at a time

	// TODO: unused?
	CRANE_COLOR_WEIGHT = 0.01
)

var (
	black = MkColor(0, 0, 0, 0)
)

func contains(s []uint32, e uint32) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type Crane struct {
	ampLevel float64
	closing  chan struct{}
	level    <-chan float64
}

func InitCrane(level <-chan float64) *Crane {
	InitLeds(CRANE_LED_COUNT1, CRANE_LED_COUNT2)

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

	strand1 := make([]uint32, LED_COUNT1) // structure
	strand2 := make([]uint32, LED_COUNT2) // heart

	// 18 lights at trueWhite (36 total between two strands)
	lights := make([]uint32, LIGHT_COUNT)
	for i, _ := range lights {
		r := uint32(rand.Int31n(LED_COUNT1))
		for contains(lights, r) {
			r = uint32(rand.Int31n(LED_COUNT1))
		}
		lights[i] = r
	}

	lightIter := 0
	nextLight := uint32(rand.Int31n(LED_COUNT1))
	for contains(lights, nextLight) {
		nextLight = uint32(rand.Int31n(LED_COUNT1))
	}

	heartColor1 := trueRed
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

				nextLight = uint32(rand.Int31n(LED_COUNT1))
				for contains(lights, nextLight) {
					nextLight = uint32(rand.Int31n(LED_COUNT1))
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

			// heart
			heartColor := MkColorWeight(heartColor1, heartColor2, weight)
			for i := 0; i < LED_COUNT2; i++ {
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
