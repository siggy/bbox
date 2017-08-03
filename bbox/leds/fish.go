package leds

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/siggy/rpi_ws281x/golang/ws2811"
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
)

const (
	STANDARD = iota
	FLICKER
	NUM_MODES
)

var (
	pink       = MkColor(159, 0, 159, 93)
	trueBlue   = MkColor(0, 0, 255, 0)
	red        = MkColor(210, 0, 50, 40)
	lightGreen = MkColor(0, 181, 115, 43)
	trueRed    = MkColor(255, 0, 0, 0)
	trueWhite  = MkColor(0, 0, 0, 255)
	purple     = MkColor(82, 0, 197, 52)
	mint       = MkColor(0, 27, 0, 228)
	trueGreen  = MkColor(0, 255, 0, 0)

	colors = []uint32{
		pink,
		trueBlue,
		red,
		lightGreen,
		trueRed,
		trueWhite,
		purple,
		mint,
		trueGreen,
	}
)

type Fish struct {
	press <-chan struct{}
	wg    *sync.WaitGroup
}

func InitFish(wg *sync.WaitGroup, press <-chan struct{}) *Fish {
	wg.Add(1)

	InitLeds(LED_COUNT1, LED_COUNT2)

	return &Fish{
		press: press,
		wg:    wg,
	}
}

func (f *Fish) Run() {
	defer f.wg.Done()

	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	ws2811.Clear()

	strand1 := make([]uint32, LED_COUNT1)
	strand2 := make([]uint32, LED_COUNT2)

	mode := STANDARD

	// STANDARD mode
	iter := 0
	weight := float64(0)

	// FLICKER mode
	flickerIter := 0

	// precompute random color rotation
	randColors := make([]uint32, LED_COUNT1)
	for i := 0; i < LED_COUNT1; i++ {
		randColors[i] = MkColor(0, uint32(rand.Int31n(256)), uint32(rand.Int31n(256)), uint32(rand.Int31n(256)))
	}

	for {
		select {
		case _, more := <-f.press:
			if more {
				mode = (mode + 1) % NUM_MODES
			} else {
				return
			}
		default:
			switch mode {
			case STANDARD:
				for i := 0; i < STRAND_COUNT1; i++ {
					color1 := colors[(iter+i)%len(colors)]
					color2 := colors[(iter+i+1)%len(colors)]
					color := MkColorWeight(color1, color2, weight)

					for j := 0; j < STRAND_LEN1; j++ {
						strand1[i*STRAND_LEN1+j] = color
					}
				}

				for i := 0; i < STRAND_COUNT2; i++ {
					color1 := colors[(iter+i)%len(colors)]
					color2 := colors[(iter+i+1)%len(colors)]
					color := MkColorWeight(color1, color2, weight)

					for j := 0; j < STRAND_LEN2; j++ {
						strand2[i*STRAND_LEN2+j] = color
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

				if weight < 1 {
					weight += 0.01
				} else {
					weight = 0
					iter = (iter + 1) % len(colors)
				}

			case FLICKER:
				for i := 0; i < LED_COUNT1; i++ {
					ws2811.SetLed(0, i, randColors[(i+flickerIter)%LED_COUNT1])
				}

				for i := 0; i < LED_COUNT2; i++ {
					ws2811.SetLed(1, i, randColors[(i+flickerIter)%LED_COUNT1])
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

				// time.Sleep(1 * time.Millisecond)

				flickerIter++
			}
		}
	}
}
