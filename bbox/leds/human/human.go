package human

import (
	"fmt"
	"math"
	"math/rand"
	//"time"

	"github.com/siggy/bbox/bbox/color"
	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/bbox/beatboxer/render/web"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
	//log "github.com/sirupsen/logrus"
)

const (
	// 1x heart, 1x globe
	STRAND_COUNT1 = 4
	STRAND_LEN1   = 60

	// 1 human
	STRAND_COUNT2 = 8
	STRAND_LEN2   = 60

	LED_COUNT1 = STRAND_COUNT1 * STRAND_LEN1 // 4*60 // 60/m
	LED_COUNT2 = STRAND_COUNT2 * STRAND_LEN2 // 8*60 // 60/m

	LIGHT_COUNT = 6 // 36 total deepPurple lights turned on at a time

	STREAK_LENGTH = LED_COUNT2 / 3 // 8*60/3 == 160

	// COOLING: How much does the air cool as it rises?
	// Less cooling = taller flames.  More cooling = shorter flames.
	// Default 55, suggested range 20-100
	COOLING = 100

	// SPARKING: What chance (out of 255) is there that a new spark will be lit?
	// Higher chance = more roaring fire.  Lower chance = more flickery fire.
	// Default 120, suggested range 50-200.
	SPARKING = 50
)

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
	return uint32(math.Min(math.Log(prod)*10, 255))
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

	phoneR := uint32(200)
	phoneG := uint32(0)
	phoneB := uint32(100)
	webMotion := uint32(0)

	cp := color.Init([]uint32{
		color.Black,
		color.Make(127, 0, 0, 0),
		color.Make(127, 127, 0, 0),
		color.Make(0, 0, 0, 64),
	})

	streakLoc2 := 0.0

	//next := time.Now()

	for {
		select {
		case phone, more := <-h.w.Phone():
			if more {
				webMotion = scaleMotion(
					phone.Motion.Acceleration.X,
					phone.Motion.Acceleration.Y,
					phone.Motion.Acceleration.Z,
				)
				phoneR = phone.R
				phoneG = phone.G
				phoneB = phone.B

				cp = color.Init([]uint32{
					color.Black,
					color.Make(phoneR/2, phoneG/2, phoneB/2, webMotion/4),
					color.Make(127, 127, 0, 0),
					color.Make(0, 0, 0, 64),
				})

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

		if true { // time.Now().After(next) {
			//next = time.Now().Add(50 * time.Millisecond)

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
				strand1[i] = cp.Get(float64(colorIndex) / float64(255))
				//rHeat, gHeat, bHeat := color.HeatColor(heat[i])
				//strand1[i] = color.Make(rHeat, gHeat, bHeat, 0)
			}

			// heartColor := color.MkColorWeight(heartColor1, heartColor2, weight)
			// for i := 0; i < LED_COUNT1; i++ {
			// 	strand1[i] = heartColor
			// }

			// TODO: swap
			//log.Infof("STRAND1:")
			//for i := 0; i < LED_COUNT1; i++ {
			//	color.PrintColor(strand1[i])
			//}

			ws2811.SetBitmap(0, strand1)
		}

		// human/robot
		amped2 := int(h.ampLevel * LED_COUNT2)
		// fmt.Printf("AMPED: %+v\n", amped2)
		for i := 0; i < amped2; i++ {
			strand2[i] = color.Red
		}
		for i := amped2; i < LED_COUNT2; i++ {
			strand2[i] = color.Black
		}

		sineMap := color.GetSineVals(LED_COUNT2, streakLoc2, STREAK_LENGTH)
		for led, value := range sineMap {
			multiplier := float64(value) / 255.0
			strand2[led] = color.Make(
				uint32(multiplier*float64(phoneR/2)),
				uint32(multiplier*float64(phoneG/2)),
				uint32(multiplier*float64(phoneB/2)),
				uint32(multiplier*float64(webMotion/4)),
			)
		}

		speed := math.Max(LED_COUNT2/36, 1)
		speed = math.Max(float64(webMotion/2), speed)

		streakLoc2 += speed
		if streakLoc2 >= LED_COUNT2 {
			streakLoc2 = 0
		}

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

func (h *Human) Close() {
	// TODO: this doesn't block?
	close(h.closing)
}
