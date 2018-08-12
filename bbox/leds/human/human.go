package human

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/siggy/bbox/bbox/color"
	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/bbox/beatboxer/render/web"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
	log "github.com/sirupsen/logrus"
)

const (
	// 1x heart, 1x globe
	STRAND_COUNT1 = 1
	STRAND_LEN1   = 30

	// 1 human
	STRAND_COUNT2 = 1
	STRAND_LEN2   = 30

	LED_COUNT1 = STRAND_COUNT1 * STRAND_LEN1 // 1*60 // 60/m
	LED_COUNT2 = STRAND_COUNT2 * STRAND_LEN2 // 1*60 // 60/m

	LIGHT_COUNT = 6 // 36 total deepPurple lights turned on at a time

	STREAK_LENGTH = LED_COUNT2 * 3 / 4

	// COOLING: How much does the air cool as it rises?
	// Less cooling = taller flames.  More cooling = shorter flames.
	// Default 55, suggested range 20-100
	COOLING = 55

	// SPARKING: What chance (out of 255) is there that a new spark will be lit?
	// Higher chance = more roaring fire.  Lower chance = more flickery fire.
	// Default 120, suggested range 50-200.
	SPARKING = 120
)

var LIGHT_COLOR = leds.MkColor(0, 123, 55, 0)

type Human struct {
	ampLevel float64
	closing  chan struct{}
	level    <-chan float64
	w        *web.Web
}

func Init(level <-chan float64, w *web.Web) *Human {
	leds.InitLeds(leds.DEFAULT_FREQ, LED_COUNT1, LED_COUNT2)

	return &Human{
		closing: make(chan struct{}),
		level:   level,
		w:       w,
	}
}

func scaleMotion(x, y, z float64) uint32 {
	prod := math.Abs(x) * math.Abs(y) * math.Abs(z)
	return uint32(math.Min(math.Log(prod)*10, 127))
}

func (h *Human) Run() {
	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	strand1 := make([]uint32, LED_COUNT1) // heart
	strand2 := make([]uint32, LED_COUNT2) // human

	heat := make([]uint32, LED_COUNT1) // heart heat

	r := uint32(255)
	g := uint32(0)
	b := uint32(0)

	webColor := uint32(0)
	webMotion := uint32(0)

	next := time.Now()

	for {
		select {
		case phone, more := <-h.w.Phone():
			if more {
				webMotion = scaleMotion(
					phone.Motion.Acceleration.X,
					phone.Motion.Acceleration.Y,
					phone.Motion.Acceleration.Z,
				)
				r = phone.R
				g = phone.G
				b = phone.B

				webColor = leds.MkColor(r, g, b, webMotion)
			} else {
				return
			}
		case level, more := <-h.level:
			if more {
				h.ampLevel = level
			} else {
				return
			}
		case _, more := <-h.closing:
			if !more {
				return
			}
		default:
		}
		_ = uint32(255.0 * h.ampLevel)

		if time.Now().After(next) {
			next = time.Now().Add(50 * time.Millisecond)

			// from https://learn.adafruit.com/led-campfire/the-code

			// Step 1.  Cool down every cell a little
			for i := 0; i < LED_COUNT1; i++ {
				nextHeat := int(heat[i]) - rand.Intn(((COOLING*10)/LED_COUNT1)+2)
				heat[i] = uint32(math.Max(0, float64(nextHeat)))
			}

			// Step 2.  Heat from each cell drifts 'up' and diffuses a little
			for i := LED_COUNT1 - 1; i > 2; i-- {
				heat[i] = (heat[i-1] + heat[i-2] + heat[i-3]) / 3
			}

			// Step 3.  Randomly ignite new 'sparks' of heat near the bottom
			if rand.Intn(255) < SPARKING {
				y := rand.Intn(7)
				sparcLoc := heat[y] + uint32(rand.Intn(95)) + 160
				heat[y] = uint32(math.Min(float64(sparcLoc), 255))
			}

			// Step 4.  Map from heat cells to LED colors
			for i := 0; i < LED_COUNT1; i++ {
				// Scale the heat value from 0-255 down to 0-240
				// for best results with color palettes.
				colorIndex := heat[i] * 255 / 240
				r, g, b := color.HeatColor(colorIndex)
				strand1[i] = leds.MkColor(r, g, b, 0)
			}

			// heartColor := leds.MkColorWeight(heartColor1, heartColor2, weight)
			// for i := 0; i < LED_COUNT1; i++ {
			// 	strand1[i] = heartColor
			// }

			// TODO: swap
			log.Infof("STRAND1:")
			for i := 0; i < LED_COUNT1; i++ {
				leds.PrintColor(strand1[i])
			}

			ws2811.SetBitmap(1, strand1)
		}

		// human/robot
		// human/robot
		for i := 0; i < LED_COUNT2; i++ {
			strand2[i] = webColor
		}

		ws2811.SetBitmap(0, strand2)

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

func (h *Human) Close() {
	// TODO: this doesn't block?
	close(h.closing)
}
