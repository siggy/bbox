package color

import (
	"math"

	"github.com/siggy/bbox/bbox/leds"
	log "github.com/sirupsen/logrus"
)

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
	return leds.MkColorWeight(c.colors[index1], c.colors[index2], intraWeight)
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
