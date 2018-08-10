package fishweb

import (
	"fmt"
	"math"
	"math/rand"
	// "time"

	"github.com/siggy/bbox/bbox/color"
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

	WEBBY = color.NUM_MODES
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

	mode := color.PURPLE_STREAK

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
		randColors[i] = color.Make(0, uint32(rand.Int31n(256)), uint32(rand.Int31n(256)), uint32(rand.Int31n(128)))
	}

	phoneR := uint32(200)
	phoneG := uint32(0)
	phoneB := uint32(100)

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
				phoneR = phone.R
				phoneG = phone.G
				phoneB = phone.B
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
		default:
		}

		ampLevel := uint32(255.0 * f.ampLevel * AMPLITUDE_FACTOR)

		switch mode {
		case WEBBY:
			webcolor := color.Make(phoneR, phoneG, phoneB, webMotion)
			for i := range strand1 {
				strand1[i] = webcolor
			}
			ws2811.SetBitmap(0, strand1)
			for i := range strand2 {
				strand2[i] = webcolor
			}
			ws2811.SetBitmap(1, strand2)

			err := ws2811.Render()
			if err != nil {
				fmt.Printf("ws2811.Render failed: %+v\n", err)
				panic(err)
			}

		case color.PURPLE_STREAK:
			amped1 := int(f.ampLevel * LED_COUNT1)
			for i := 0; i < amped1; i++ {
				strand1[i] = color.Red
			}
			for i := amped1; i < LED_COUNT1; i++ {
				strand1[i] = color.Black
			}

			sineMap := color.GetSineVals(LED_COUNT1, streakLoc1, int(length))
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand1[led] = color.Make(
					uint32(multiplier*float64(phoneR)),
					uint32(multiplier*float64(phoneG)),
					uint32(multiplier*float64(phoneB)),
					uint32(multiplier*float64(webMotion)),
				)
			}

			speed := math.Max(LED_COUNT1/36, 1)
			speed = math.Max(float64(webMotion), speed)

			streakLoc1 += speed
			if streakLoc1 >= LED_COUNT1 {
				streakLoc1 = 0
			}
			ws2811.SetBitmap(0, strand1)

			amped2 := int(f.ampLevel * LED_COUNT2)
			for i := 0; i < amped2; i++ {
				strand2[i] = color.Red
			}
			for i := amped2; i < LED_COUNT2; i++ {
				strand2[i] = color.Black
			}

			sineMap = color.GetSineVals(LED_COUNT2, streakLoc2, int(length))
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand2[led] = color.Make(
					uint32(multiplier*float64(phoneR)),
					uint32(multiplier*float64(phoneG)),
					uint32(multiplier*float64(phoneB)),
					uint32(multiplier*float64(webMotion)),
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

		case color.COLOR_STREAKS:
			speed := 10.0
			c := color.Colors[(iter)%len(color.Colors)]

			sineMap := color.GetSineVals(LED_COUNT1, streakLoc1, int(length))
			for i := range strand1 {
				strand1[i] = color.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand1[led] = color.MultiplyColor(c, multiplier)
			}

			streakLoc1 += speed
			if streakLoc1 >= LED_COUNT1 {
				streakLoc1 = 0
				iter = (iter + 1) % len(color.Colors)
			}

			ws2811.SetBitmap(0, strand1)

			sineMap = color.GetSineVals(LED_COUNT2, streakLoc2, int(length))
			for i := range strand2 {
				strand2[i] = color.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand2[led] = color.MultiplyColor(c, multiplier)
			}

			streakLoc2 += speed
			if streakLoc2 >= LED_COUNT2 {
				streakLoc2 = 0
			}

			ws2811.SetBitmap(1, strand2)
		case color.FAST_COLOR_STREAKS:
			speed := 100.0
			c := color.Colors[(iter)%len(color.Colors)]

			sineMap := color.GetSineVals(LED_COUNT1, streakLoc1, int(length))
			for i := range strand1 {
				strand1[i] = color.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand1[led] = color.MultiplyColor(c, multiplier)
			}

			streakLoc1 += speed
			if streakLoc1 >= LED_COUNT1 {
				streakLoc1 = 0
				iter = (iter + 1) % len(color.Colors)
			}

			ws2811.SetBitmap(0, strand1)

			sineMap = color.GetSineVals(LED_COUNT2, streakLoc2, int(length))
			for i := range strand2 {
				strand2[i] = color.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand2[led] = color.MultiplyColor(c, multiplier)
			}

			streakLoc2 += speed
			if streakLoc2 >= LED_COUNT2 {
				streakLoc2 = 0
			}

			ws2811.SetBitmap(1, strand2)
		case color.SOUND_COLOR_STREAKS:
			speed := 100.0*f.ampLevel + 10.0
			c := color.Colors[(iter)%len(color.Colors)]

			sineMap := color.GetSineVals(LED_COUNT1, streakLoc1, int(length))
			for i := range strand1 {
				strand1[i] = color.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand1[led] = color.MultiplyColor(c, multiplier)
			}

			streakLoc1 += speed
			if streakLoc1 >= LED_COUNT1 {
				streakLoc1 = 0
				iter = (iter + 1) % len(color.Colors)
			}

			ws2811.SetBitmap(0, strand1)

			sineMap = color.GetSineVals(LED_COUNT2, streakLoc2, int(length))
			for i := range strand2 {
				strand2[i] = color.Black
			}
			for led, value := range sineMap {
				multiplier := float64(value) / 255.0
				strand2[led] = color.MultiplyColor(c, multiplier)
			}

			streakLoc2 += speed
			if streakLoc2 >= LED_COUNT2 {
				streakLoc2 = 0
			}

			ws2811.SetBitmap(1, strand2)
		case color.FILL_RED:
			for i := 0; i < LED_COUNT1; i += 30 {
				for j := 0; j < i; j++ {
					ws2811.SetLed(0, j, color.Red)
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
					ws2811.SetLed(1, j, color.Red)
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
					ws2811.SetLed(0, j, color.Make(0, 0, 0, 0))
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
					ws2811.SetLed(1, j, color.Make(0, 0, 0, 0))
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
		case color.SLOW_EQUALIZE:
			for i := 0; i < STRAND_COUNT1; i++ {
				c := color.Colors[(iter+i)%len(color.Colors)]

				for j := 0; j < STRAND_LEN1; j++ {
					strand1[i*STRAND_LEN1+j] = c
				}
			}

			for i := 0; i < STRAND_COUNT2; i++ {
				c := color.Colors[(iter+i)%len(color.Colors)]

				for j := 0; j < STRAND_LEN2; j++ {
					strand2[i*STRAND_LEN2+j] = c
				}
			}

			// for i, c := range strand1 {
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
				iter = (iter + 1) % len(color.Colors)
			}

			// time.Sleep(1 * time.Second)
		case color.FILL_EQUALIZE:
			for i := 0; i < STRAND_COUNT1; i++ {
				color1 := color.Colors[(iter+i)%len(color.Colors)]
				color2 := color.Colors[(iter+i+1)%len(color.Colors)]
				c := color.MkColorWeight(color1, color2, weight)

				for j := 0; j < STRAND_LEN1; j++ {
					strand1[i*STRAND_LEN1+j] = c
				}
			}

			for i := 0; i < STRAND_COUNT2; i++ {
				color1 := color.Colors[(iter+i)%len(color.Colors)]
				color2 := color.Colors[(iter+i+1)%len(color.Colors)]
				c := color.MkColorWeight(color1, color2, weight)

				for j := 0; j < STRAND_LEN2; j++ {
					strand2[i*STRAND_LEN2+j] = c
				}
			}

			for i, c := range strand1 {
				// if i == 0 {
				// 	PrintColor(c)
				// }
				ws2811.SetLed(0, i, c)
				if i%10 == 0 {
					ws2811.Render()
				}
			}
			for i, c := range strand2 {
				// if i == 0 {
				// 	PrintColor(c)
				// }
				ws2811.SetLed(1, i, c)
				if i%10 == 0 {
					ws2811.Render()
				}
			}

			iter = (iter + 1) % len(color.Colors)

			// time.Sleep(1 * time.Second)
		case color.EQUALIZE:
			for i := 0; i < STRAND_COUNT1; i++ {
				color1 := color.Colors[(iter+i)%len(color.Colors)]
				color2 := color.Colors[(iter+i+1)%len(color.Colors)]
				c := color.MkColorWeight(color1, color2, weight)

				for j := 0; j < STRAND_LEN1; j++ {
					strand1[i*STRAND_LEN1+j] = c
				}
			}

			for i := 0; i < STRAND_COUNT2; i++ {
				color1 := color.Colors[(iter+i)%len(color.Colors)]
				color2 := color.Colors[(iter+i+1)%len(color.Colors)]
				c := color.MkColorWeight(color1, color2, weight)

				for j := 0; j < STRAND_LEN2; j++ {
					strand2[i*STRAND_LEN2+j] = c
				}
			}

			for i, c := range strand1 {
				// if i == 0 {
				// 	PrintColor(c)
				// }
				ws2811.SetLed(0, i, c)
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
				iter = (iter + 1) % len(color.Colors)
			}

			// time.Sleep(1 * time.Second)

		case color.STANDARD:
			for i := 0; i < STRAND_COUNT1; i++ {
				color1 := color.Colors[(iter+i)%len(color.Colors)]
				color2 := color.Colors[(iter+i+1)%len(color.Colors)]
				c := color.MkColorWeight(color1, color2, weight)
				ampColor := color.AmpColor(c, ampLevel)

				for j := 0; j < STRAND_LEN1; j++ {
					strand1[i*STRAND_LEN1+j] = ampColor
				}
			}

			for i := 0; i < STRAND_COUNT2; i++ {
				color1 := color.Colors[(iter+i)%len(color.Colors)]
				color2 := color.Colors[(iter+i+1)%len(color.Colors)]
				c := color.MkColorWeight(color1, color2, weight)
				ampColor := color.AmpColor(c, ampLevel)

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
				iter = (iter + 1) % len(color.Colors)
			}

		case color.FLICKER:
			for i := 0; i < LED_COUNT1; i++ {
				ws2811.SetLed(0, i, color.AmpColor(randColors[(i+flickerIter)%LED_COUNT1], ampLevel))
			}

			for i := 0; i < LED_COUNT2; i++ {
				ws2811.SetLed(1, i, color.AmpColor(randColors[(i+flickerIter)%LED_COUNT1], ampLevel))
			}

			flickerIter++
		case color.AUDIO:
			ampColor := color.AmpColor(color.TrueBlue, ampLevel)
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
