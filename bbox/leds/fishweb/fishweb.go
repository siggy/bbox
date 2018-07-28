package fishweb

import (
	"fmt"
	"math"
	"math/rand"
	// "time"

	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/bbox/beatboxer/render/web"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
	log "github.com/sirupsen/logrus"
)

const (
	// 2x side fins
	STRAND_COUNT1 = 5
	STRAND_LEN1   = 144

	// 1x top and back fins
	STRAND_COUNT2 = 10
	STRAND_LEN2   = 60

	LED_COUNT1 = STRAND_COUNT1 * STRAND_LEN1 // 5*144 // * 2x(5) // 144/m
	LED_COUNT2 = STRAND_COUNT2 * STRAND_LEN2 // 10*60 // * 1x(4 + 2 + 4) // 60/m

	AMPLITUDE_FACTOR = 0.75

	WEBBY = leds.NUM_MODES
)

type Fish struct {
	ampLevel float64
	closing  chan struct{}
	level    <-chan float64
	press    <-chan struct{}
	w        *web.Web
}

func InitFish(level <-chan float64, press <-chan struct{}, w *web.Web) *Fish {
	leds.InitLeds(leds.DEFAULT_FREQ, LED_COUNT1, LED_COUNT2)

	return &Fish{
		closing: make(chan struct{}),
		level:   level,
		press:   press,
		w:       w,
	}
}

func scaleMotion(x, y, z float64) uint32 {
	prod := math.Abs(x) * math.Abs(y) * math.Abs(z)
	return uint32(math.Min(math.Log(prod)*10, 127))
}

func (f *Fish) Run() {
	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	ws2811.Clear()

	strand1 := make([]uint32, LED_COUNT1)
	strand2 := make([]uint32, LED_COUNT2)

	mode := leds.PURPLE_STREAK

	// PURPLE_STREAK mode
	streakLoc1 := 0.0
	streakLoc2 := 0.0
	length := LED_COUNT1 / 3.6

	// STANDARD mode
	iter := 0
	weight := float64(0)

	// FLICKER mode
	flickerIter := 0

	// precompute random color rotation
	randColors := make([]uint32, LED_COUNT1)
	for i := 0; i < LED_COUNT1; i++ {
		randColors[i] = leds.MkColor(0, uint32(rand.Int31n(256)), uint32(rand.Int31n(256)), uint32(rand.Int31n(128)))
	}

	r := uint32(200)
	g := uint32(0)
	b := uint32(100)
	w := uint32(0)

	webMotion := uint32(0)

	for {
		select {
		case phone, more := <-f.w.Phone():
			if more {
				webMotion = scaleMotion(
					phone.Motion.Acceleration.X,
					phone.Motion.Acceleration.Y,
					phone.Motion.Acceleration.Z,
				)
				r = phone.R
				g = phone.G
				b = phone.B
			} else {
				return
			}
		case _, more := <-f.press:
			if more {
				mode = (mode + 1) % (WEBBY + 1)
				log.Infof("press: %+v", mode)
			} else {
				return
			}
		case level, more := <-f.level:
			if more {
				f.ampLevel = level
				log.Debugf("level: %+v", level)
			} else {
				return
			}
		case _, more := <-f.closing:
			if !more {
				return
			}
		}

		ampLevel := uint32(255.0 * f.ampLevel * AMPLITUDE_FACTOR)

		switch mode {
		case WEBBY:
			webColor := leds.MkColor(r, g, b, webMotion)
			for i := range strand1 {
				strand1[i] = webColor
			}
			ws2811.SetBitmap(0, strand1)
			for i := range strand2 {
				strand2[i] = webColor
			}
			ws2811.SetBitmap(1, strand2)

			err := ws2811.Render()
			if err != nil {
				fmt.Printf("ws2811.Render failed: %+v\n", err)
				panic(err)
			}

		case leds.PURPLE_STREAK:
			amped1 := int(f.ampLevel * LED_COUNT1)
			for i := 0; i < amped1; i++ {
				strand1[i] = leds.Red
			}
			for i := amped1; i < LED_COUNT1; i++ {
				strand1[i] = leds.Black
			}

			sineMap := leds.GetSineVals(LED_COUNT1, streakLoc1, int(length))
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand1[led] = leds.MkColor(
					uint32(multiplier*float64(r)),
					uint32(multiplier*float64(g)),
					uint32(multiplier*float64(b)),
					uint32(multiplier*float64(w)),
				)
			}

			speed := math.Max(LED_COUNT1/36, 1)
			speed = math.Max(float64(webMotion), speed)

			speed = 0

			streakLoc1 += speed
			if streakLoc1 >= LED_COUNT1 {
				streakLoc1 = 0
			}
			ws2811.SetBitmap(0, strand1)

			amped2 := int(f.ampLevel * LED_COUNT2)
			for i := 0; i < amped2; i++ {
				strand2[i] = leds.Red
			}
			for i := amped2; i < LED_COUNT2; i++ {
				strand2[i] = leds.Black
			}

			sineMap = leds.GetSineVals(LED_COUNT2, streakLoc2, int(length))
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand2[led] = leds.MkColor(
					uint32(multiplier*float64(r)),
					uint32(multiplier*float64(g)),
					uint32(multiplier*float64(b)),
					uint32(multiplier*float64(w)),
				)
			}

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

		case leds.COLOR_STREAKS:
			speed := 10.0
			color := leds.Colors[(iter)%len(leds.Colors)]

			sineMap := leds.GetSineVals(LED_COUNT1, streakLoc1, int(length))
			for i := range strand1 {
				strand1[i] = leds.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand1[led] = leds.MultiplyColor(color, multiplier)
			}

			streakLoc1 += speed
			if streakLoc1 >= LED_COUNT1 {
				streakLoc1 = 0
				iter = (iter + 1) % len(leds.Colors)
			}

			ws2811.SetBitmap(0, strand1)

			sineMap = leds.GetSineVals(LED_COUNT2, streakLoc2, int(length))
			for i := range strand2 {
				strand2[i] = leds.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand2[led] = leds.MultiplyColor(color, multiplier)
			}

			streakLoc2 += speed
			if streakLoc2 >= LED_COUNT2 {
				streakLoc2 = 0
			}

			ws2811.SetBitmap(1, strand2)
		case leds.FAST_COLOR_STREAKS:
			speed := 100.0
			color := leds.Colors[(iter)%len(leds.Colors)]

			sineMap := leds.GetSineVals(LED_COUNT1, streakLoc1, int(length))
			for i := range strand1 {
				strand1[i] = leds.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand1[led] = leds.MultiplyColor(color, multiplier)
			}

			streakLoc1 += speed
			if streakLoc1 >= LED_COUNT1 {
				streakLoc1 = 0
				iter = (iter + 1) % len(leds.Colors)
			}

			ws2811.SetBitmap(0, strand1)

			sineMap = leds.GetSineVals(LED_COUNT2, streakLoc2, int(length))
			for i := range strand2 {
				strand2[i] = leds.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand2[led] = leds.MultiplyColor(color, multiplier)
			}

			streakLoc2 += speed
			if streakLoc2 >= LED_COUNT2 {
				streakLoc2 = 0
			}

			ws2811.SetBitmap(1, strand2)
		case leds.SOUND_COLOR_STREAKS:
			speed := 100.0*f.ampLevel + 10.0
			color := leds.Colors[(iter)%len(leds.Colors)]

			sineMap := leds.GetSineVals(LED_COUNT1, streakLoc1, int(length))
			for i := range strand1 {
				strand1[i] = leds.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand1[led] = leds.MultiplyColor(color, multiplier)
			}

			streakLoc1 += speed
			if streakLoc1 >= LED_COUNT1 {
				streakLoc1 = 0
				iter = (iter + 1) % len(leds.Colors)
			}

			ws2811.SetBitmap(0, strand1)

			sineMap = leds.GetSineVals(LED_COUNT2, streakLoc2, int(length))
			for i := range strand2 {
				strand2[i] = leds.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand2[led] = leds.MultiplyColor(color, multiplier)
			}

			streakLoc2 += speed
			if streakLoc2 >= LED_COUNT2 {
				streakLoc2 = 0
			}

			ws2811.SetBitmap(1, strand2)
		case leds.FILL_RED:
			for i := 0; i < LED_COUNT1; i += 30 {
				for j := 0; j < i; j++ {
					ws2811.SetLed(0, j, leds.Red)
				}

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
			for i := 0; i < LED_COUNT2; i += 30 {
				for j := 0; j < i; j++ {
					ws2811.SetLed(1, j, leds.Red)
				}

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
			for i := 0; i < LED_COUNT1; i += 30 {
				for j := 0; j < i; j++ {
					ws2811.SetLed(0, j, leds.MkColor(0, 0, 0, 0))
				}

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
			for i := 0; i < LED_COUNT2; i += 30 {
				for j := 0; j < i; j++ {
					ws2811.SetLed(1, j, leds.MkColor(0, 0, 0, 0))
				}

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
		case leds.SLOW_EQUALIZE:
			for i := 0; i < STRAND_COUNT1; i++ {
				color := leds.Colors[(iter+i)%len(leds.Colors)]

				for j := 0; j < STRAND_LEN1; j++ {
					strand1[i*STRAND_LEN1+j] = color
				}
			}

			for i := 0; i < STRAND_COUNT2; i++ {
				color := leds.Colors[(iter+i)%len(leds.Colors)]

				for j := 0; j < STRAND_LEN2; j++ {
					strand2[i*STRAND_LEN2+j] = color
				}
			}

			// for i, color := range strand1 {
			// 	// if i == 0 {
			// 	// 	PrintColor(color)
			// 	// }
			// 	ws2811.SetLed(0, i, color)
			// 	if i%10 == 0 {
			// 		ws2811.Render()
			// 	}
			// }
			// time.Sleep(1 * time.Second)

			ws2811.SetBitmap(0, strand1)
			ws2811.SetBitmap(1, strand2)

			if weight < 1 {
				weight += 0.01
			} else {
				weight = 0
				iter = (iter + 1) % len(leds.Colors)
			}

			// time.Sleep(1 * time.Second)
		case leds.FILL_EQUALIZE:
			for i := 0; i < STRAND_COUNT1; i++ {
				color1 := leds.Colors[(iter+i)%len(leds.Colors)]
				color2 := leds.Colors[(iter+i+1)%len(leds.Colors)]
				color := leds.MkColorWeight(color1, color2, weight)

				for j := 0; j < STRAND_LEN1; j++ {
					strand1[i*STRAND_LEN1+j] = color
				}
			}

			for i := 0; i < STRAND_COUNT2; i++ {
				color1 := leds.Colors[(iter+i)%len(leds.Colors)]
				color2 := leds.Colors[(iter+i+1)%len(leds.Colors)]
				color := leds.MkColorWeight(color1, color2, weight)

				for j := 0; j < STRAND_LEN2; j++ {
					strand2[i*STRAND_LEN2+j] = color
				}
			}

			for i, color := range strand1 {
				// if i == 0 {
				// 	PrintColor(color)
				// }
				ws2811.SetLed(0, i, color)
				if i%10 == 0 {
					ws2811.Render()
				}
			}
			for i, color := range strand2 {
				// if i == 0 {
				// 	PrintColor(color)
				// }
				ws2811.SetLed(1, i, color)
				if i%10 == 0 {
					ws2811.Render()
				}
			}

			iter = (iter + 1) % len(leds.Colors)

			// time.Sleep(1 * time.Second)
		case leds.EQUALIZE:
			for i := 0; i < STRAND_COUNT1; i++ {
				color1 := leds.Colors[(iter+i)%len(leds.Colors)]
				color2 := leds.Colors[(iter+i+1)%len(leds.Colors)]
				color := leds.MkColorWeight(color1, color2, weight)

				for j := 0; j < STRAND_LEN1; j++ {
					strand1[i*STRAND_LEN1+j] = color
				}
			}

			for i := 0; i < STRAND_COUNT2; i++ {
				color1 := leds.Colors[(iter+i)%len(leds.Colors)]
				color2 := leds.Colors[(iter+i+1)%len(leds.Colors)]
				color := leds.MkColorWeight(color1, color2, weight)

				for j := 0; j < STRAND_LEN2; j++ {
					strand2[i*STRAND_LEN2+j] = color
				}
			}

			for i, color := range strand1 {
				// if i == 0 {
				// 	PrintColor(color)
				// }
				ws2811.SetLed(0, i, color)
				if i%10 == 0 {
					ws2811.Render()
				}
			}
			// time.Sleep(1 * time.Second)

			// ws2811.SetBitmap(0, strand1)
			ws2811.SetBitmap(1, strand2)

			if weight < 1 {
				weight += 0.01
			} else {
				weight = 0
				iter = (iter + 1) % len(leds.Colors)
			}

			// time.Sleep(1 * time.Second)

		case leds.STANDARD:
			for i := 0; i < STRAND_COUNT1; i++ {
				color1 := leds.Colors[(iter+i)%len(leds.Colors)]
				color2 := leds.Colors[(iter+i+1)%len(leds.Colors)]
				color := leds.MkColorWeight(color1, color2, weight)
				ampColor := leds.AmpColor(color, ampLevel)

				for j := 0; j < STRAND_LEN1; j++ {
					strand1[i*STRAND_LEN1+j] = ampColor
				}
			}

			for i := 0; i < STRAND_COUNT2; i++ {
				color1 := leds.Colors[(iter+i)%len(leds.Colors)]
				color2 := leds.Colors[(iter+i+1)%len(leds.Colors)]
				color := leds.MkColorWeight(color1, color2, weight)
				ampColor := leds.AmpColor(color, ampLevel)

				for j := 0; j < STRAND_LEN2; j++ {
					strand2[i*STRAND_LEN2+j] = ampColor
				}
			}

			ws2811.SetBitmap(0, strand1)
			ws2811.SetBitmap(1, strand2)

			if weight < 1 {
				weight += 0.01
			} else {
				weight = 0
				iter = (iter + 1) % len(leds.Colors)
			}

		case leds.FLICKER:
			for i := 0; i < LED_COUNT1; i++ {
				ws2811.SetLed(0, i, leds.AmpColor(randColors[(i+flickerIter)%LED_COUNT1], ampLevel))
			}

			for i := 0; i < LED_COUNT2; i++ {
				ws2811.SetLed(1, i, leds.AmpColor(randColors[(i+flickerIter)%LED_COUNT1], ampLevel))
			}

			flickerIter++
		case leds.AUDIO:
			ampColor := leds.AmpColor(leds.TrueBlue, ampLevel)
			for i := 0; i < LED_COUNT1; i++ {
				ws2811.SetLed(0, i, ampColor)
			}

			for i := 0; i < LED_COUNT2; i++ {
				ws2811.SetLed(1, i, ampColor)
			}
		}

		err := ws2811.Render()
		if err != nil {
			fmt.Printf("ws2811.Render failed: %+v\n", err)
			panic(err)
		}

		// err = ws2811.Wait()
		// if err != nil {
		// 	fmt.Printf("ws2811.Wait failed: %+v\n", err)
		// 	panic(err)
		// }
	}
}

func (f *Fish) Close() {
	// TODO: this doesn't block?
	close(f.closing)
}
