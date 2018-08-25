package leds

import (
	"fmt"
	"math/rand"

	"github.com/siggy/bbox/bbox/color"
	"github.com/siggy/bbox/beatboxer/render/web"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	// 2x undercarriage strands
	CRAWLER_STRAND_COUNT1 = 4
	CRAWLER_STRAND_LEN1   = 60

	CRAWLER_LED_COUNT1 = CRAWLER_STRAND_COUNT1 * CRAWLER_STRAND_LEN1 // 4*60 // * 2x(4) // 60/m

	// TODO: make transition time-based rather than fast as possible?
	// lower weight == slower transitions
	// crawler has fewer LEDs than fish == faster iterations
	CRAWLER_COLOR_WEIGHT = 0.005
)

type Crawler struct {
	ampLevel float64
	closing  chan struct{}
	level    <-chan float64
	press    <-chan struct{}
	w        *web.Web
}

func InitCrawler(level <-chan float64, press <-chan struct{}, w *web.Web) *Crawler {
	InitLeds(DEFAULT_FREQ, CRAWLER_LED_COUNT1, 0)

	return &Crawler{
		closing: make(chan struct{}),
		level:   level,
		press:   press,
		w:       w,
	}
}

func (c *Crawler) Run() {
	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	ws2811.Clear()

	strand1 := make([]uint32, CRAWLER_LED_COUNT1)

	mode := color.PURPLE_STREAK

	// PURPLE_STREAK mode
	streakLoc1 := 0.0
	length := 200
	speed := 0.1

	// STANDARD mode
	iter := 0
	weight := float64(0)

	// FLICKER mode
	flickerIter := 0

	// precompute random color rotation
	randColors := make([]uint32, CRAWLER_LED_COUNT1)
	for i := 0; i < CRAWLER_LED_COUNT1; i++ {
		randColors[i] = color.Make(0, uint32(rand.Int31n(256)), uint32(rand.Int31n(256)), uint32(rand.Int31n(256)))
	}

	phoneR := uint32(200)
	phoneG := uint32(0)
	phoneB := uint32(100)

	webMotion := uint32(0)

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
			} else {
				return
			}
		case _, more := <-c.press:
			if more {
				mode = (mode + 1) % color.NUM_MODES
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
			ampLevel := uint32(255.0 * c.ampLevel)
			switch mode {
			case color.PURPLE_STREAK:
				speed = 0.5
				sineMap := color.GetSineVals(CRAWLER_LED_COUNT1, streakLoc1, int(length))
				for i, _ := range strand1 {
					strand1[i] = color.Black
				}

				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand1[led] = color.Make(
						uint32(multiplier*float64(phoneR)),
						uint32(multiplier*float64(phoneG)),
						uint32(multiplier*float64(phoneB)),
						uint32(multiplier*float64(webMotion)),
					)
				}

				streakLoc1 += speed
				if streakLoc1 >= CRAWLER_LED_COUNT1 {
					streakLoc1 = 0
				}

				ws2811.SetBitmap(0, strand1)
			case color.COLOR_STREAKS:
				speed = 10.0
				c := color.Colors[(iter)%len(color.Colors)]

				sineMap := color.GetSineVals(CRAWLER_LED_COUNT1, streakLoc1, int(length))
				for i, _ := range strand1 {
					strand1[i] = color.Black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand1[led] = color.MultiplyColor(c, multiplier)
				}

				streakLoc1 += speed
				if streakLoc1 >= CRAWLER_LED_COUNT1 {
					streakLoc1 = 0
					iter = (iter + 1) % len(color.Colors)
				}

				ws2811.SetBitmap(0, strand1)
			case color.FAST_COLOR_STREAKS:
				speed = 100.0
				c := color.Colors[(iter)%len(color.Colors)]

				sineMap := color.GetSineVals(CRAWLER_LED_COUNT1, streakLoc1, int(length))
				for i, _ := range strand1 {
					strand1[i] = color.Black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand1[led] = color.MultiplyColor(c, multiplier)
				}

				streakLoc1 += speed
				if streakLoc1 >= CRAWLER_LED_COUNT1 {
					streakLoc1 = 0
					iter = (iter + 1) % len(color.Colors)
				}

				ws2811.SetBitmap(0, strand1)
			case color.SOUND_COLOR_STREAKS:
				speed = 100.0*c.ampLevel + 10.0
				c := color.Colors[(iter)%len(color.Colors)]

				sineMap := color.GetSineVals(CRAWLER_LED_COUNT1, streakLoc1, int(length))
				for i, _ := range strand1 {
					strand1[i] = color.Black
				}
				for led, value := range sineMap {
					multiplier := float64(value) / 255.0
					strand1[led] = color.MultiplyColor(c, multiplier)
				}

				streakLoc1 += speed
				if streakLoc1 >= CRAWLER_LED_COUNT1 {
					streakLoc1 = 0
					iter = (iter + 1) % len(color.Colors)
				}

				ws2811.SetBitmap(0, strand1)
			case color.FILL_RED:
				for i := 0; i < CRAWLER_LED_COUNT1; i += 30 {
					for j := 0; j < i; j++ {
						ws2811.SetLed(0, j, color.Red)
					}
				}
				for i := 0; i < CRAWLER_LED_COUNT1; i += 30 {
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

			case color.SLOW_EQUALIZE:
				for i := 0; i < CRAWLER_STRAND_COUNT1; i++ {
					c := color.Colors[(iter+i)%len(color.Colors)]

					for j := 0; j < CRAWLER_STRAND_LEN1; j++ {
						strand1[i*CRAWLER_STRAND_LEN1+j] = c
					}
				}

				// for i, c := range strand1 {
				// 	// if i == 0 {
				// 	// 	PrintColor(c)
				// 	// }
				// 	ws2811.SetLed(0, i, c)
				// 	if i%10 == 0 {
				// 		ws2811.Render()
				// 	}
				// }
				// time.Sleep(1 * time.Second)

				ws2811.SetBitmap(0, strand1)

				if weight < 1 {
					weight += 0.01
				} else {
					weight = 0
					iter = (iter + 1) % len(color.Colors)
				}

				// time.Sleep(1 * time.Second)
			case color.FILL_EQUALIZE:
				for i := 0; i < CRAWLER_STRAND_COUNT1; i++ {
					color1 := color.Colors[(iter+i)%len(color.Colors)]
					color2 := color.Colors[(iter+i+1)%len(color.Colors)]
					c := color.MkColorWeight(color1, color2, weight)

					for j := 0; j < CRAWLER_STRAND_LEN1; j++ {
						strand1[i*CRAWLER_STRAND_LEN1+j] = c
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
				iter = (iter + 1) % len(color.Colors)

				// time.Sleep(1 * time.Second)
			case color.EQUALIZE:
				ampLevel1 := int(CRAWLER_LED_COUNT1 * c.ampLevel)

				for i := 0; i < CRAWLER_STRAND_COUNT1; i++ {
					color1 := color.Colors[(iter+i)%len(color.Colors)]
					color2 := color.Colors[(iter+i+1)%len(color.Colors)]
					c := color.MkColorWeight(color1, color2, weight)
					ampColor := color.AmpColor(c, 255)

					for j := 0; j < CRAWLER_STRAND_LEN1; j++ {
						idx := i*CRAWLER_STRAND_LEN1 + j
						if idx < ampLevel1 {
							strand1[idx] = ampColor
						} else {
							strand1[idx] = c
						}
					}
				}

				ws2811.SetBitmap(0, strand1)

				err := ws2811.Render()
				if err != nil {
					fmt.Printf("ws2811.Render failed: %+v\n", err)
					panic(err)
				}

				if weight < 1 {
					weight += CRAWLER_COLOR_WEIGHT
				} else {
					weight = 0
					iter = (iter + 1) % len(color.Colors)
				}
			case color.STANDARD:
				for i := 0; i < CRAWLER_STRAND_COUNT1; i++ {
					color1 := color.Colors[(iter+i)%len(color.Colors)]
					color2 := color.Colors[(iter+i+1)%len(color.Colors)]
					c := color.MkColorWeight(color1, color2, weight)
					ampColor := color.AmpColor(c, ampLevel)

					for j := 0; j < CRAWLER_STRAND_LEN1; j++ {
						strand1[i*CRAWLER_STRAND_LEN1+j] = ampColor
					}
				}

				ws2811.SetBitmap(0, strand1)

				if weight < 1 {
					weight += CRAWLER_COLOR_WEIGHT
				} else {
					weight = 0
					iter = (iter + 1) % len(color.Colors)
				}

			case color.FLICKER:
				for i := 0; i < CRAWLER_LED_COUNT1; i++ {
					ws2811.SetLed(0, i, color.AmpColor(randColors[(i+flickerIter)%CRAWLER_LED_COUNT1], ampLevel))
				}

				flickerIter++
			case color.AUDIO:
				ampColor := color.AmpColor(color.TrueBlue, ampLevel)
				for i := 0; i < CRAWLER_LED_COUNT1; i++ {
					ws2811.SetLed(0, i, ampColor)
				}
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
	}
}

func (c *Crawler) Close() {
	// TODO: this doesn't block?
	close(c.closing)
}
