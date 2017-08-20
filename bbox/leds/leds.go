package leds

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	GPIO_PIN1A = 18  // PWM0, must be 18 or 12
	GPIO_PIN1B = 12  // PWM0, must be 18 or 12
	GPIO_PIN2  = 13  // PWM1, must be 13 for rPI 3
	BRIGHTNESS = 255 // 0-255
	PI_FACTOR  = math.Pi / 2.
)

const (
	EQUALIZE = iota
	STANDARD
	FLICKER
	AUDIO
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

	Colors = []uint32{
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

	redWhite = MkColor(255, 0, 0, 255)
)

var (
	Red    = binary.LittleEndian.Uint32([]byte{0x00, 0x00, 0x20, 0x00})
	redw   = binary.LittleEndian.Uint32([]byte{0x00, 0x00, 0x20, 0x10})
	green  = binary.LittleEndian.Uint32([]byte{0x00, 0x20, 0x00, 0x00})
	greenw = binary.LittleEndian.Uint32([]byte{0x00, 0x20, 0x00, 0x10})
	blue   = binary.LittleEndian.Uint32([]byte{0x20, 0x00, 0x00, 0x00})
	bluew  = binary.LittleEndian.Uint32([]byte{0x20, 0x00, 0x00, 0x10})
	white  = binary.LittleEndian.Uint32([]byte{0x20, 0x20, 0x20, 0x00})
	whitew = binary.LittleEndian.Uint32([]byte{0x20, 0x20, 0x20, 0x10})

	testColors = []uint32{Red, redw, green, greenw, blue, bluew, white, whitew}
)

const (
	SINE_AMPLITUDE   = 127
	SINE_SHIFT       = 127
	SINE_PERIOD      = 3
	SINE_HALF_PERIOD = float64(SINE_PERIOD) / 2.0
)

// TODO: cache?
func getSineVals(ledCount int, floatBeat float64) (sineVals map[int]int) {
	first := int(math.Ceil(floatBeat - SINE_HALF_PERIOD)) // 12.7 - 1.5 => 11.2 => 12
	last := int(math.Floor(floatBeat + SINE_HALF_PERIOD)) // 12.7 + 1.5 => 14.2 => 14

	sineFunc := func(x int) int {
		// y = a * sin((x-h)/b) + k
		h := floatBeat - SINE_PERIOD/4.0
		b := SINE_PERIOD / (2 * math.Pi)
		return int(
			SINE_AMPLITUDE*math.Sin((float64(x)-h)/b) +
				SINE_SHIFT,
		)
	}

	for i := first; i <= last; i++ {
		y := sineFunc(i)
		if y != 0 {
			sineVals[(i+ledCount)%ledCount] = sineFunc(i)
		}
	}

	return
}

// TODO: cache?
func SineScale(weight float64) float64 {
	return math.Sin(PI_FACTOR * weight)
}

// maps midpoint 128 => 32 for brightness
func scale(x uint32) uint32 {
	// y = 1000*(0.005333 * 4002473^(x/1000)-0.005333)
	return uint32(1000 * (0.005333*math.Pow(4002473., float64(x)/1000.) - 0.005333))
}

// expects 0 <= [r,g,b,w] <= 255
func MkColor(r uint32, g uint32, b uint32, w uint32) uint32 {
	return uint32(b + g<<8 + r<<16 + w<<24)
}

func MkColorWeight(color1 uint32, color2 uint32, weight float64) uint32 {
	b1 := color1 & 0x000000ff
	g1 := (color1 & 0x0000ff00) >> 8
	r1 := (color1 & 0x00ff0000) >> 16
	w1 := (color1 & 0xff000000) >> 24

	b2 := color2 & 0x000000ff
	g2 := (color2 & 0x0000ff00) >> 8
	r2 := (color2 & 0x00ff0000) >> 16
	w2 := (color2 & 0xff000000) >> 24

	return MkColor(
		scale(uint32(float64(r1)+float64(int32(r2)-int32(r1))*SineScale(weight))),
		scale(uint32(float64(g1)+float64(int32(g2)-int32(g1))*SineScale(weight))),
		scale(uint32(float64(b1)+float64(int32(b2)-int32(b1))*SineScale(weight))),
		scale(uint32(float64(w1)+float64(int32(w2)-int32(w1))*SineScale(weight))),
	)
}

func AmpColor(color uint32, ampLevel uint32) uint32 {
	b := color & 0x000000ff
	g := (color & 0x0000ff00) >> 8
	r := (color & 0x00ff0000) >> 16
	w := (color & 0xff000000) >> 24

	return MkColor(
		uint32(math.Min(float64(r+ampLevel), 255)),
		g,
		b,
		w,
	)
}

/*
 * Standalone functions to test all LEDs
 */

func InitLeds(ledCount1 int, ledCount2 int) {
	// init once for each PIN1 (PWM0)
	fmt.Printf("ws2811.Init()\n")
	err := ws2811.Init(
		GPIO_PIN1A, ledCount1, BRIGHTNESS,
		GPIO_PIN2, ledCount2, BRIGHTNESS,
	)
	if err != nil {
		fmt.Printf("ws2811.Init failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}

	ws2811.Fini()

	err = ws2811.Init(
		GPIO_PIN1B, ledCount1, BRIGHTNESS,
		GPIO_PIN2, ledCount2, BRIGHTNESS,
	)
	if err != nil {
		fmt.Printf("ws2811.Init failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.Render()\n")
	err = ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}

	// warm up
	for i := 0; i < ledCount1; i += 30 {
		fmt.Printf("warmup GPIO1: %+v of %+v\n", i, ledCount1)
		for j := 0; j < i; j++ {
			ws2811.SetLed(0, j, Red)
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
	for i := 0; i < ledCount2; i += 30 {
		fmt.Printf("warmup GPIO2: %+v of %+v\n", i, ledCount2)
		for j := 0; j < i; j++ {
			ws2811.SetLed(1, j, Red)
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

func SetLed(channel int, led int) {
	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.SetLed(%+v)\n", led)
	ws2811.SetLed(channel, led, Red)

	fmt.Printf("ws2811.Render()\n")
	err := ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}
}

func Shutdown() {
	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.Render()\n")
	err := ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Fini()\n")
	ws2811.Fini()
}

// Turn off all LEDs
func Clear(ledCount1 int, ledCount2 int) {
	fmt.Printf("ws2811.Init()\n")
	err := ws2811.Init(
		GPIO_PIN1A, ledCount1, BRIGHTNESS,
		GPIO_PIN2, ledCount2, BRIGHTNESS,
	)
	if err != nil {
		fmt.Printf("ws2811.Init failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.Render()\n")
	err = ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Fini()\n")
	ws2811.Fini()
}
