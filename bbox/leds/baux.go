package leds

import (
	"fmt"
	"math"
	"time"

	"github.com/siggy/bbox/bbox/color"
	"github.com/siggy/bbox/beatboxer/render/web"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	// TODO: bbox testing
	// 2x structure
	BAUX_STRAND_COUNT1 = 1
	BAUX_STRAND_LEN1   = 85

	// 1x globe
	BAUX_STRAND_COUNT2 = 4
	BAUX_STRAND_LEN2   = 60

	BAUX_LED_COUNT1 = BAUX_STRAND_COUNT1 * BAUX_STRAND_LEN1 // 5*30 // 30/m
	BAUX_LED_COUNT2 = BAUX_STRAND_COUNT2 * BAUX_STRAND_LEN2 // 4*60 // 60/m

	DEFAULT_INTERVAL_MS = 2000

	BAUX_STREAK_LENGTH = BAUX_LED_COUNT1 * 3 / 4
)

// LEDS
var globeLeds = []int{
	0,
	29,
	55,
	79,
	102,
	124,
	145,
	165,
	183,
	200,
	216,
	230,
}

type Baux struct {
	ampLevel float64
	closing  chan struct{}
	level    <-chan float64
	w        *web.Web
}

func InitBaux(level <-chan float64, w *web.Web) *Baux {
	InitLeds(DEFAULT_FREQ, BAUX_LED_COUNT1, BAUX_LED_COUNT2)

	return &Baux{
		closing: make(chan struct{}),
		level:   level,
		w:       w,
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

	strand1 := make([]uint32, BAUX_LED_COUNT1) // base
	strand2 := make([]uint32, BAUX_LED_COUNT2) // globe

	phoneR := uint32(200)
	phoneG := uint32(0)
	phoneB := uint32(100)
	webMotion := uint32(0)

	globeColor1 := color.TrueRed
	globeColor2 := color.Black

	last := time.Now()
	interval := 2 * time.Second

	for {
		select {
		case phone, more := <-c.w.Phone():
			if more {
				webMotion = color.ScaleMotion(
					phone.Motion.Acceleration.X,
					phone.Motion.Acceleration.Y,
					phone.Motion.Acceleration.Z,
				)
				phoneR = phone.R
				phoneG = phone.G
				phoneB = phone.B

				globeColor1 = color.Make(
					phoneR,
					phoneG,
					phoneB,
					0,
				)

				globeColor2 = color.Make(
					255-phoneR,
					255-phoneG,
					255-phoneB,
					0,
				)

			} else {
				return
			}
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
			interval = time.Duration(math.Max(
				DEFAULT_INTERVAL_MS-(DEFAULT_INTERVAL_MS*c.ampLevel),
				100,
			)) * time.Millisecond

			now := time.Now()
			loc := 1.0 - float64(now.Sub(last).Nanoseconds())/float64(interval.Nanoseconds())

			if loc < 0 {
				loc = 1
				last = now
			}

			// streaks
			sineMap := color.GetSineVals(BAUX_LED_COUNT1, loc*BAUX_LED_COUNT1, BAUX_STREAK_LENGTH)
			for led, value := range sineMap {
				mag := float64(value) / 254.0
				strand1[led] = color.Make(
					uint32(float64(phoneR)*mag),
					uint32(float64(phoneG)*mag),
					uint32(float64(phoneB)*mag),
					uint32(float64(webMotion)*mag),
				)
			}

			// globe
			for i := 0; i < BAUX_LED_COUNT2; i++ {
				strand2[i] = 0
			}
			for i := 0; i < len(globeLeds)-1; i++ {
				start := globeLeds[i]
				end := globeLeds[i+1] - 1
				peak1 := float64(start) + float64(end-start)*loc

				p2 := loc + 0.5 - math.Trunc(loc+0.5)
				peak2 := float64(start) + float64(end-start)*p2

				sineMap1 := color.GetSineVals(end-start, peak1, (end-start)/2)
				for led, value := range sineMap1 {
					mag := float64(value) / 254.0
					strand2[start+led] = color.MultiplyColor(globeColor1, mag)
				}

				sineMap2 := color.GetSineVals(end-start, peak2, (end-start)/2)
				for led, value := range sineMap2 {
					mag := float64(value) / 254.0
					strand2[start+led] = color.MultiplyColor(globeColor2, mag)
				}
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
