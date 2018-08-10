package color

import (
	"encoding/binary"
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
)

const (
	PI_FACTOR = math.Pi / 2.

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

const (
	SINE_AMPLITUDE = 127
	SINE_SHIFT     = 127
)

type Color struct {
	R, G, B, W uint32
}

var (
	pink       = Make(159, 0, 159, 93)
	trueBlue   = Make(0, 0, 255, 0)
	TrueBlue   = trueBlue
	red        = Make(210, 0, 50, 40)
	lightGreen = Make(0, 181, 115, 43)
	TrueRed    = Make(255, 0, 0, 0)
	TrueWhite  = Make(0, 0, 0, 255)
	purple     = Make(82, 0, 197, 52)
	mint       = Make(0, 27, 0, 228)
	trueGreen  = Make(0, 255, 0, 0)
	deepPurple = Make(200, 0, 100, 0)

	Colors = []uint32{
		pink,
		trueBlue,
		red,
		lightGreen,
		TrueRed,
		deepPurple,
		TrueWhite,
		purple,
		mint,
		trueGreen,
	}

	redWhite = Make(255, 0, 0, 255)
	black    = Make(0, 0, 0, 0)
	Black    = black
	Purple   = purple

	Red    = binary.LittleEndian.Uint32([]byte{0x00, 0x00, 0x20, 0x00})
	redw   = binary.LittleEndian.Uint32([]byte{0x00, 0x00, 0x20, 0x10})
	green  = binary.LittleEndian.Uint32([]byte{0x00, 0x20, 0x00, 0x00})
	greenw = binary.LittleEndian.Uint32([]byte{0x00, 0x20, 0x00, 0x10})
	blue   = binary.LittleEndian.Uint32([]byte{0x20, 0x00, 0x00, 0x00})
	bluew  = binary.LittleEndian.Uint32([]byte{0x20, 0x00, 0x00, 0x10})
	white  = binary.LittleEndian.Uint32([]byte{0x20, 0x20, 0x20, 0x00})
	whitew = binary.LittleEndian.Uint32([]byte{0x20, 0x20, 0x20, 0x10})

	testColors = []uint32{Red, redw, green, greenw, blue, bluew, white, whitew}

	sineCache = make(map[sineKey]map[int]int)
)

// TODO: cache?
type sineKey struct {
	ledCount  int
	floatBeat float64
	period    int
}

type ColorPalette struct {
	colors []uint32
	length int
}

func Init(colors []uint32) *ColorPalette {
	return &ColorPalette{
		colors: colors,
		length: len(colors),
	}
}

func (c *ColorPalette) Get(weight float64) uint32 {
	if weight < 0 {
		log.Errorf("Weight out of range: %f", weight)
		weight = 0
	} else if weight > 1 {
		// log.Errorf("Weight out of range: %f", weight)
		weight = 1
	}
	scaled := float64(c.length-1) * weight

	floor := math.Floor(scaled)
	ceil := math.Ceil(scaled)

	index1 := int(math.Floor(scaled))
	index2 := int(math.Ceil(scaled))
	intraWeight := scaled - float64(index1)

	if floor == 0 {
		index2 = (index1 + 1) % c.length
	} else if ceil == float64(c.length-1) {
		index1 = (index2 - 1) % c.length
	}

	if weight == 1 {
		intraWeight = 1
	}
	// fmt.Printf("ColorPalette:\n  c: %+v\n  weight: %f\n  i1: %d\n  i2: %d\n  intraWeight: %f\n", c, weight, index1, index2, intraWeight)
	return MkColorWeight(c.colors[index1], c.colors[index2], intraWeight)
}

func HeatColor(temperature uint32) (uint32, uint32, uint32) {
	var r, g, b uint32

	// Scale 'heat' down from 0-255 to 0-191,
	// which can then be easily divided into three
	// equal 'thirds' of 64 units each.
	t192 := temperature * 191 / 255

	// calculate a value that ramps up from
	// zero to 255 in each 'third' of the scale.
	heatramp := t192 & 0x3F // 0..63
	heatramp <<= 2          // scale up to 0..252

	// now figure out which third of the spectrum we're in:
	if t192&0x80 != 0 {
		// we're in the hottest third
		r = 255      // full red
		g = 255      // full green
		b = heatramp // ramp up blue

	} else if t192&0x40 != 0 {
		// we're in the middle third
		r = 255      // full red
		g = heatramp // ramp up green
		b = 0        // no blue

	} else {
		// we're in the coolest third
		r = heatramp // ramp up red
		g = 0        // no green
		b = 0        // no blue
	}

	return r, g, b
}

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

func Contains(s []uint32, e uint32) bool {
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

	return Make(
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

	colorWeightCache[cw] = Make(
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

	return Make(
		uint32(math.Min(float64(r+ampLevel), 255)),
		g,
		b,
		w,
	)
}

func Make(r uint32, g uint32, b uint32, w uint32) uint32 {
	return uint32(b + g<<8 + r<<16 + w<<24)
}

func Split(color uint32) Color {
	return Color{
		B: color & 0x000000ff,
		G: (color & 0x0000ff00) >> 8,
		R: (color & 0x00ff0000) >> 16,
		W: (color & 0xff000000) >> 24,
	}
}
