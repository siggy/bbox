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
