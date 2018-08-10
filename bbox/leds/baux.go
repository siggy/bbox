package leds

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/siggy/bbox/bbox/color"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	// TODO: bbox testing
	// 2x structure
	BAUX_STRAND_COUNT1 = 1
	BAUX_STRAND_LEN1   = 85

	// 1x heart
	BAUX_STRAND_COUNT2 = 4
	BAUX_STRAND_LEN2   = 60

	BAUX_LED_COUNT1 = BAUX_STRAND_COUNT1 * BAUX_STRAND_LEN1 // 5*30 // 30/m
	BAUX_LED_COUNT2 = BAUX_STRAND_COUNT2 * BAUX_STRAND_LEN2 // 4*60 // 60/m

	BAUX_BPM      = 15
	BAUX_INTERVAL = 60 * time.Second / BAUX_BPM / 2 // (30 beats/min) / 2 color transitions/beat

	BAUX_LIGHT_COUNT = 6 // 36 total deepPurple lights turned on at a time

	BAUX_STREAK_LENGTH = BAUX_LED_COUNT1 * 3 / 4
)

var BAUX_LIGHT_COLOR = color.Make(0, 123, 55, 0)

type Baux struct {
	ampLevel float64
	closing  chan struct{}
	level    <-chan float64
}

func InitBaux(level <-chan float64) *Baux {
	InitLeds(DEFAULT_FREQ, BAUX_LED_COUNT1, BAUX_LED_COUNT2)

	return &Baux{
		closing: make(chan struct{}),
		level:   level,
	}
}

func (c *Baux) Run() {
	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	strand1 := make([]uint32, BAUX_LED_COUNT1) // structure
	strand2 := make([]uint32, BAUX_LED_COUNT2) // heart

	// 36 lights at deepPurple
	lights := make([]uint32, BAUX_LIGHT_COUNT)
	for i, _ := range lights {
		r := uint32(rand.Int31n(BAUX_LED_COUNT1))
		for color.Contains(lights, r) {
			r = uint32(rand.Int31n(BAUX_LED_COUNT1))
		}
		lights[i] = r
	}

	lightIter := 0
	nextLight := uint32(rand.Int31n(BAUX_LED_COUNT1))
	for color.Contains(lights, nextLight) {
		nextLight = uint32(rand.Int31n(BAUX_LED_COUNT1))
	}

	heartColor1 := color.TrueRed
	heartColor2 := color.Black

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
			weight := 1.0 - float64(now.Sub(last).Nanoseconds())/float64(BAUX_INTERVAL.Nanoseconds())

			if weight < 0 {
				weight = 1
				last = now

				// structure iters
				lights[lightIter] = nextLight
				lightIter = (lightIter + BAUX_LIGHT_COUNT - 1) % BAUX_LIGHT_COUNT

				nextLight = uint32(rand.Int31n(BAUX_LED_COUNT1))
				for color.Contains(lights, nextLight) {
					nextLight = uint32(rand.Int31n(BAUX_LED_COUNT1))
				}

				// heart iters
				tmp := heartColor1
				heartColor1 = heartColor2
				heartColor2 = tmp
			}

			// structure
			for i, light := range lights {
				if i == lightIter {
					strand1[light] = color.MkColorWeight(BAUX_LIGHT_COLOR, color.Black, weight)
					strand1[nextLight] = color.MkColorWeight(color.Black, BAUX_LIGHT_COLOR, weight)
				} else {
					strand1[light] = BAUX_LIGHT_COLOR
				}
			}

			// streaks
			sineMap := color.GetSineVals(BAUX_LED_COUNT1, weight*BAUX_LED_COUNT1, BAUX_STREAK_LENGTH)
			for led, value := range sineMap {
				if !color.Contains(lights, uint32(led)) {
					mag := float64(value) / 254.0
					strand1[led] = color.Make(uint32(float64(200)*mag), 0, uint32(float64(100)*mag), 0)
				}
			}

			// heart
			heartColor := color.MkColorWeight(heartColor1, heartColor2, weight)
			for i := 0; i < BAUX_LED_COUNT2; i++ {
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
		}
	}
}

func (c *Baux) Close() {
	// TODO: this doesn't block?
	close(c.closing)
}
