package leds

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	DEFAULT_FREQ = 800000
	GPIO_PIN1A   = 18  // PWM0, must be 18 or 12
	GPIO_PIN1B   = 12  // PWM0, must be 18 or 12
	GPIO_PIN2    = 13  // PWM1, must be 13 for rPI 3
	BRIGHTNESS   = 255 // 0-255
	PI_FACTOR    = math.Pi / 2.
)

const (
	PURPLE_STREAK = iota
	COLOR_STREAKS
	FAST_COLOR_STREAKS
	SOUND_COLOR_STREAKS
	FILL_RED
	SLOW_EQUALIZE
	FILL_EQUALIZE
	EQUALIZE
	STANDARD
	FLICKER
	AUDIO
	NUM_MODES
)

var (
	pink       = MkColor(159, 0, 159, 93)
	trueBlue   = MkColor(0, 0, 255, 0)
	TrueBlue   = trueBlue
	red        = MkColor(210, 0, 50, 40)
	lightGreen = MkColor(0, 181, 115, 43)
	trueRed    = MkColor(255, 0, 0, 0)
	trueWhite  = MkColor(0, 0, 0, 255)
	purple     = MkColor(82, 0, 197, 52)
	mint       = MkColor(0, 27, 0, 228)
	trueGreen  = MkColor(0, 255, 0, 0)
	deepPurple = MkColor(200, 0, 100, 0)

	Colors = []uint32{
		pink,
		trueBlue,
		red,
		lightGreen,
		trueRed,
		deepPurple,
		trueWhite,
		purple,
		mint,
		trueGreen,
	}

	redWhite = MkColor(255, 0, 0, 255)
	black    = MkColor(0, 0, 0, 0)
	Black    = black
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
	SINE_AMPLITUDE = 127
	SINE_SHIFT     = 127
)

// TODO: cache?
type sineKey struct {
	ledCount  int
	floatBeat float64
	period    int
}

var sineCache = make(map[sineKey]map[int]int)

func GetSineVals(ledCount int, floatBeat float64, period int) (sineVals map[int]int) {
	if sineCache[sineKey{ledCount, floatBeat, period}] != nil {
		return sineCache[sineKey{ledCount, floatBeat, period}]
	}

	halfPeriod := float64(period) / 2.0

	first := int(math.Ceil(floatBeat - halfPeriod)) // 12.7 - 1.5 => 11.2 => 12
	last := int(math.Floor(floatBeat + halfPeriod)) // 12.7 + 1.5 => 14.2 => 14

	sineFunc := func(x int) int {
		// y = a * sin((x-h)/b) + k
		h := floatBeat - float64(period)/4.0
		b := float64(period) / (2 * math.Pi)
		return int(
			SINE_AMPLITUDE*math.Sin((float64(x)-h)/b) +
				SINE_SHIFT,
		)
	}

	sineVals = make(map[int]int)

	for i := first; i <= last; i++ {
		y := sineFunc(i)
		if y != 0 {
			sineVals[(i+ledCount)%ledCount] = int(scale(uint32(sineFunc(i))))
		}
	}

	sineCache[sineKey{ledCount, floatBeat, period}] = sineVals

	return
}

// TODO: cache?
func SineScale(weight float64) float64 {
	return math.Sin(PI_FACTOR * weight)
}

func contains(s []uint32, e uint32) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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

type ColorWeight struct {
	color1 uint32
	color2 uint32
	weight float64
}

var (
	colorWeightCache = make(map[ColorWeight]uint32)
)

func PrintColor(color uint32) {
	b := color & 0x000000ff
	g := (color & 0x0000ff00) >> 8
	r := (color & 0x00ff0000) >> 16
	w := (color & 0xff000000) >> 24
	fmt.Printf("(%+v, %+v, %+v, %+v)\n", r, g, b, w)
}

func MultiplyColor(color uint32, multiplier float64) uint32 {
	b := color & 0x000000ff
	g := (color & 0x0000ff00) >> 8
	r := (color & 0x00ff0000) >> 16
	w := (color & 0xff000000) >> 24

	return MkColor(
		uint32(multiplier*float64(r)),
		uint32(multiplier*float64(g)),
		uint32(multiplier*float64(b)),
		uint32(multiplier*float64(w)),
	)
}

func MkColorWeight(color1 uint32, color2 uint32, weight float64) uint32 {
	cw := ColorWeight{
		color1: color1,
		color2: color2,
		weight: weight,
	}

	if val, ok := colorWeightCache[cw]; ok {
		return val
	}

	b1 := color1 & 0x000000ff
	g1 := (color1 & 0x0000ff00) >> 8
	r1 := (color1 & 0x00ff0000) >> 16
	w1 := (color1 & 0xff000000) >> 24

	b2 := color2 & 0x000000ff
	g2 := (color2 & 0x0000ff00) >> 8
	r2 := (color2 & 0x00ff0000) >> 16
	w2 := (color2 & 0xff000000) >> 24

	colorWeightCache[cw] = MkColor(
		scale(uint32(float64(r1)+float64(int32(r2)-int32(r1))*SineScale(weight))),
		scale(uint32(float64(g1)+float64(int32(g2)-int32(g1))*SineScale(weight))),
		scale(uint32(float64(b1)+float64(int32(b2)-int32(b1))*SineScale(weight))),
		scale(uint32(float64(w1)+float64(int32(w2)-int32(w1))*SineScale(weight))),
	)

	return colorWeightCache[cw]
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

func InitLeds(freq int, ledCount1 int, ledCount2 int) {
	// init once for each PIN1 (PWM0)
	fmt.Printf("ws2811.Init()\n")
	err := ws2811.Init(
		freq,
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
		DEFAULT_FREQ,
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
		DEFAULT_FREQ,
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
