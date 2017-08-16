package leds

import (
	"fmt"
	"math/rand"

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

	AMPLITUDE_FACTOR = 0.75
)

type Fish struct {
	ampLevel float64
	closing  chan struct{}
	level    <-chan float64
	press    <-chan struct{}
}

func InitFish(level <-chan float64, press <-chan struct{}) *Fish {
	InitLeds(LED_COUNT1, LED_COUNT2)

	return &Fish{
		closing: make(chan struct{}),
		level:   level,
		press:   press,
	}
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

	mode := EQUALIZE

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
			case EQUALIZE:
				ampLevel1 := int(LED_COUNT1 * f.ampLevel)
				ampLevel2 := int(LED_COUNT2 * f.ampLevel)

				for i := 0; i < STRAND_COUNT1; i++ {
					color1 := Colors[(iter+i)%len(Colors)]
					color2 := Colors[(iter+i+1)%len(Colors)]
					color := MkColorWeight(color1, color2, weight)
					ampColor := AmpColor(color, 255)

					for j := 0; j < STRAND_LEN1; j++ {
						idx := i*STRAND_LEN1 + j
						if idx < ampLevel1 {
							strand1[idx] = ampColor
						} else {
							strand1[idx] = color
						}
					}
				}

				for i := 0; i < STRAND_COUNT2; i++ {
					color1 := Colors[(iter+i)%len(Colors)]
					color2 := Colors[(iter+i+1)%len(Colors)]
					color := MkColorWeight(color1, color2, weight)
					ampColor := AmpColor(color, 255)

					for j := 0; j < STRAND_LEN2; j++ {
						idx := i*STRAND_LEN2 + j
						if idx < ampLevel2 {
							strand2[idx] = ampColor
						} else {
							strand2[idx] = color
						}
					}
				}

				ws2811.SetBitmap(0, strand1)
				ws2811.SetBitmap(1, strand2)

				err := ws2811.Render()
				if err != nil {
					fmt.Printf("ws2811.Render failed: %+v\n", err)
					panic(err)
				}

				if weight < 1 {
					weight += 0.01
				} else {
					weight = 0
					iter = (iter + 1) % len(Colors)
				}
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

				ws2811.SetBitmap(0, strand1)
				ws2811.SetBitmap(1, strand2)

				err := ws2811.Render()
				if err != nil {
					fmt.Printf("ws2811.Render failed: %+v\n", err)
					panic(err)
				}

				if weight < 1 {
					weight += 0.01
				} else {
					weight = 0
					iter = (iter + 1) % len(Colors)
				}

			case FLICKER:
				for i := 0; i < LED_COUNT1; i++ {
					ws2811.SetLed(0, i, AmpColor(randColors[(i+flickerIter)%LED_COUNT1], ampLevel))
				}

				for i := 0; i < LED_COUNT2; i++ {
					ws2811.SetLed(1, i, AmpColor(randColors[(i+flickerIter)%LED_COUNT1], ampLevel))
				}

				err := ws2811.Render()
				if err != nil {
					fmt.Printf("ws2811.Render failed: %+v\n", err)
					panic(err)
				}

				// time.Sleep(1 * time.Millisecond)

				flickerIter++
			case AUDIO:
				ampColor := AmpColor(trueBlue, ampLevel)
				for i := 0; i < LED_COUNT1; i++ {
					ws2811.SetLed(0, i, ampColor)
				}

				for i := 0; i < LED_COUNT2; i++ {
					ws2811.SetLed(1, i, ampColor)
				}

				err := ws2811.Render()
				if err != nil {
					fmt.Printf("ws2811.Render failed: %+v\n", err)
					panic(err)
				}
			}
		}
	}
}

func (f *Fish) Close() {
	// TODO: this doesn't block?
	close(f.closing)
}
