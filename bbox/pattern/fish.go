package pattern

import (
	"fmt"
	"math/rand"
	// "time"

	"github.com/siggy/bbox/bbox/renderer"
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
)

type Fish struct {
	ampLevel float64
	closing  chan struct{}
	level    <-chan float64
	press    <-chan struct{}
	r        renderer.Renderer
}

func InitFish(r renderer.Renderer, level <-chan float64, press <-chan struct{}) *Fish {
	InitLeds(r, DEFAULT_FREQ, LED_COUNT1, LED_COUNT2)

	return &Fish{
		closing: make(chan struct{}),
		level:   level,
		press:   press,
		r:       r,
	}
}

func (f *Fish) Run() {
	defer func() {
		f.r.Clear()
		f.r.Render()
		f.r.Wait()
		f.r.Fini()
	}()

	f.r.Clear()

	strand1 := make([]uint32, LED_COUNT1)
	strand2 := make([]uint32, LED_COUNT2)

	mode := FILL_EQUALIZE

	// PURPLE_STREAK mode
	streakLoc1 := 0.0
	streakLoc2 := 0.0
	length := 200
	speed := 10.0
	r := 200
	g := 0
	b := 100
	w := 0

	// STANDARD mode
	iter := 0
	weight := float64(0)

	// FLICKER mode
	flickerIter := 0

	// precompute random color rotation
	randColors := make([]uint32, LED_COUNT1)
	for i := 0; i < LED_COUNT1; i++ {
		randColors[i] = MkColor(0, uint32(rand.Int31n(256)), uint32(rand.Int31n(256)), uint32(rand.Int31n(128)))
	}

	for {
		select {
		case _, more := <-f.press:
			if more {
				mode = (mode + 1) % NUM_MODES
			} else {
				return
			}
		case level, more := <-f.level:
			if more {
				f.ampLevel = level
			} else {
				return
			}
		case _, more := <-f.closing:
			if !more {
				return
			}
		default:
			ampLevel := uint32(255.0 * f.ampLevel * AMPLITUDE_FACTOR)

			switch mode {
			case PURPLE_STREAK:
				speed = 10.0
				sineMap := GetSineVals(LED_COUNT1, streakLoc1, int(length))
				for i, _ := range strand1 {
					strand1[i] = black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand1[led] = MkColor(
						uint32(multiplier*float64(r)),
						uint32(multiplier*float64(g)),
						uint32(multiplier*float64(b)),
						uint32(multiplier*float64(w)),
					)
				}

				streakLoc1 += speed
				if streakLoc1 >= LED_COUNT1 {
					streakLoc1 = 0
				}

				f.r.SetBitmap(0, strand1)

				sineMap = GetSineVals(LED_COUNT2, streakLoc2, int(length))
				for i, _ := range strand2 {
					strand2[i] = black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand2[led] = MkColor(
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

				f.r.SetBitmap(1, strand2)
			case COLOR_STREAKS:
				speed = 10.0
				color := Colors[(iter)%len(Colors)]

				sineMap := GetSineVals(LED_COUNT1, streakLoc1, int(length))
				for i, _ := range strand1 {
					strand1[i] = black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand1[led] = MultiplyColor(color, multiplier)
				}

				streakLoc1 += speed
				if streakLoc1 >= LED_COUNT1 {
					streakLoc1 = 0
					iter = (iter + 1) % len(Colors)
				}

				f.r.SetBitmap(0, strand1)

				sineMap = GetSineVals(LED_COUNT2, streakLoc2, int(length))
				for i, _ := range strand2 {
					strand2[i] = black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand2[led] = MultiplyColor(color, multiplier)
				}

				streakLoc2 += speed
				if streakLoc2 >= LED_COUNT2 {
					streakLoc2 = 0
				}

				f.r.SetBitmap(1, strand2)
			case FAST_COLOR_STREAKS:
				speed = 100.0
				color := Colors[(iter)%len(Colors)]

				sineMap := GetSineVals(LED_COUNT1, streakLoc1, int(length))
				for i, _ := range strand1 {
					strand1[i] = black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand1[led] = MultiplyColor(color, multiplier)
				}

				streakLoc1 += speed
				if streakLoc1 >= LED_COUNT1 {
					streakLoc1 = 0
					iter = (iter + 1) % len(Colors)
				}

				f.r.SetBitmap(0, strand1)

				sineMap = GetSineVals(LED_COUNT2, streakLoc2, int(length))
				for i, _ := range strand2 {
					strand2[i] = black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand2[led] = MultiplyColor(color, multiplier)
				}

				streakLoc2 += speed
				if streakLoc2 >= LED_COUNT2 {
					streakLoc2 = 0
				}

				f.r.SetBitmap(1, strand2)
			case SOUND_COLOR_STREAKS:
				speed = 100.0*f.ampLevel + 10.0
				color := Colors[(iter)%len(Colors)]

				sineMap := GetSineVals(LED_COUNT1, streakLoc1, int(length))
				for i, _ := range strand1 {
					strand1[i] = black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand1[led] = MultiplyColor(color, multiplier)
				}

				streakLoc1 += speed
				if streakLoc1 >= LED_COUNT1 {
					streakLoc1 = 0
					iter = (iter + 1) % len(Colors)
				}

				f.r.SetBitmap(0, strand1)

				sineMap = GetSineVals(LED_COUNT2, streakLoc2, int(length))
				for i, _ := range strand2 {
					strand2[i] = black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand2[led] = MultiplyColor(color, multiplier)
				}

				streakLoc2 += speed
				if streakLoc2 >= LED_COUNT2 {
					streakLoc2 = 0
				}

				f.r.SetBitmap(1, strand2)
			case FILL_RED:
				for i := 0; i < LED_COUNT1; i += 30 {
					for j := 0; j < i; j++ {
						f.r.SetLed(0, j, Red)
					}

					err := f.r.Render()
					if err != nil {
						fmt.Printf("f.r.Render failed: %+v\n", err)
						panic(err)
					}
					err = f.r.Wait()
					if err != nil {
						fmt.Printf("f.r.Wait failed: %+v\n", err)
						panic(err)
					}
				}
				for i := 0; i < LED_COUNT2; i += 30 {
					for j := 0; j < i; j++ {
						f.r.SetLed(1, j, Red)
					}

					err := f.r.Render()
					if err != nil {
						fmt.Printf("f.r.Render failed: %+v\n", err)
						panic(err)
					}
					err = f.r.Wait()
					if err != nil {
						fmt.Printf("f.r.Wait failed: %+v\n", err)
						panic(err)
					}
				}
				for i := 0; i < LED_COUNT1; i += 30 {
					for j := 0; j < i; j++ {
						f.r.SetLed(0, j, MkColor(0, 0, 0, 0))
					}

					err := f.r.Render()
					if err != nil {
						fmt.Printf("f.r.Render failed: %+v\n", err)
						panic(err)
					}
					err = f.r.Wait()
					if err != nil {
						fmt.Printf("f.r.Wait failed: %+v\n", err)
						panic(err)
					}
				}
				for i := 0; i < LED_COUNT2; i += 30 {
					for j := 0; j < i; j++ {
						f.r.SetLed(1, j, MkColor(0, 0, 0, 0))
					}

					err := f.r.Render()
					if err != nil {
						fmt.Printf("f.r.Render failed: %+v\n", err)
						panic(err)
					}
					err = f.r.Wait()
					if err != nil {
						fmt.Printf("f.r.Wait failed: %+v\n", err)
						panic(err)
					}
				}
			case SLOW_EQUALIZE:
				for i := 0; i < STRAND_COUNT1; i++ {
					color := Colors[(iter+i)%len(Colors)]

					for j := 0; j < STRAND_LEN1; j++ {
						strand1[i*STRAND_LEN1+j] = color
					}
				}

				for i := 0; i < STRAND_COUNT2; i++ {
					color := Colors[(iter+i)%len(Colors)]

					for j := 0; j < STRAND_LEN2; j++ {
						strand2[i*STRAND_LEN2+j] = color
					}
				}

				// for i, color := range strand1 {
				// 	// if i == 0 {
				// 	// 	PrintColor(color)
				// 	// }
				// 	f.r.SetLed(0, i, color)
				// 	if i%10 == 0 {
				// 		f.r.Render()
				// 	}
				// }
				// time.Sleep(1 * time.Second)

				f.r.SetBitmap(0, strand1)
				f.r.SetBitmap(1, strand2)

				if weight < 1 {
					weight += 0.01
				} else {
					weight = 0
					iter = (iter + 1) % len(Colors)
				}

				// time.Sleep(1 * time.Second)
			case FILL_EQUALIZE:
				for i := 0; i < STRAND_COUNT1; i++ {
					color1 := Colors[(iter+i)%len(Colors)]
					color2 := Colors[(iter+i+1)%len(Colors)]
					color := MkColorWeight(color1, color2, weight)

					for j := 0; j < STRAND_LEN1; j++ {
						strand1[i*STRAND_LEN1+j] = color
					}
				}

				for i := 0; i < STRAND_COUNT2; i++ {
					color1 := Colors[(iter+i)%len(Colors)]
					color2 := Colors[(iter+i+1)%len(Colors)]
					color := MkColorWeight(color1, color2, weight)

					for j := 0; j < STRAND_LEN2; j++ {
						strand2[i*STRAND_LEN2+j] = color
					}
				}

				for i, color := range strand1 {
					// if i == 0 {
					// 	PrintColor(color)
					// }
					f.r.SetLed(0, i, color)
					if i%10 == 0 {
						f.r.Render()
					}
				}
				for i, color := range strand2 {
					// if i == 0 {
					// 	PrintColor(color)
					// }
					f.r.SetLed(1, i, color)
					if i%10 == 0 {
						f.r.Render()
					}
				}

				iter = (iter + 1) % len(Colors)

				// time.Sleep(1 * time.Second)
			case EQUALIZE:
				for i := 0; i < STRAND_COUNT1; i++ {
					color1 := Colors[(iter+i)%len(Colors)]
					color2 := Colors[(iter+i+1)%len(Colors)]
					color := MkColorWeight(color1, color2, weight)

					for j := 0; j < STRAND_LEN1; j++ {
						strand1[i*STRAND_LEN1+j] = color
					}
				}

				for i := 0; i < STRAND_COUNT2; i++ {
					color1 := Colors[(iter+i)%len(Colors)]
					color2 := Colors[(iter+i+1)%len(Colors)]
					color := MkColorWeight(color1, color2, weight)

					for j := 0; j < STRAND_LEN2; j++ {
						strand2[i*STRAND_LEN2+j] = color
					}
				}

				for i, color := range strand1 {
					// if i == 0 {
					// 	PrintColor(color)
					// }
					f.r.SetLed(0, i, color)
					if i%10 == 0 {
						f.r.Render()
					}
				}
				// time.Sleep(1 * time.Second)

				// f.r.SetBitmap(0, strand1)
				f.r.SetBitmap(1, strand2)

				if weight < 1 {
					weight += 0.01
				} else {
					weight = 0
					iter = (iter + 1) % len(Colors)
				}

				// time.Sleep(1 * time.Second)

			case STANDARD:
				for i := 0; i < STRAND_COUNT1; i++ {
					color1 := Colors[(iter+i)%len(Colors)]
					color2 := Colors[(iter+i+1)%len(Colors)]
					color := MkColorWeight(color1, color2, weight)
					ampColor := AmpColor(color, ampLevel)

					for j := 0; j < STRAND_LEN1; j++ {
						strand1[i*STRAND_LEN1+j] = ampColor
					}
				}

				for i := 0; i < STRAND_COUNT2; i++ {
					color1 := Colors[(iter+i)%len(Colors)]
					color2 := Colors[(iter+i+1)%len(Colors)]
					color := MkColorWeight(color1, color2, weight)
					ampColor := AmpColor(color, ampLevel)

					for j := 0; j < STRAND_LEN2; j++ {
						strand2[i*STRAND_LEN2+j] = ampColor
					}
				}

				f.r.SetBitmap(0, strand1)
				f.r.SetBitmap(1, strand2)

				if weight < 1 {
					weight += 0.01
				} else {
					weight = 0
					iter = (iter + 1) % len(Colors)
				}

			case FLICKER:
				for i := 0; i < LED_COUNT1; i++ {
					f.r.SetLed(0, i, AmpColor(randColors[(i+flickerIter)%LED_COUNT1], ampLevel))
				}

				for i := 0; i < LED_COUNT2; i++ {
					f.r.SetLed(1, i, AmpColor(randColors[(i+flickerIter)%LED_COUNT1], ampLevel))
				}

				flickerIter++
			case AUDIO:
				ampColor := AmpColor(trueBlue, ampLevel)
				for i := 0; i < LED_COUNT1; i++ {
					f.r.SetLed(0, i, ampColor)
				}

				for i := 0; i < LED_COUNT2; i++ {
					f.r.SetLed(1, i, ampColor)
				}
			}

			err := f.r.Render()
			if err != nil {
				fmt.Printf("f.r.Render failed: %+v\n", err)
				panic(err)
			}

			err = f.r.Wait()
			if err != nil {
				fmt.Printf("f.r.Wait failed: %+v\n", err)
				panic(err)
			}
		}
	}
}

func (f *Fish) Close() {
	// TODO: this doesn't block?
	close(f.closing)
}
