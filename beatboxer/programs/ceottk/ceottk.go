package ceottk

import (
	"fmt"
	"time"

	"github.com/siggy/bbox/beatboxer"
	"github.com/siggy/bbox/beatboxer/color"
	"github.com/siggy/bbox/beatboxer/render"
)

const (
	SEQEUNCE_LENGTH = 123
	EARLY_PLAY      = 0.5
)

var (
	aliens = map[int]struct{}{
		6:   struct{}{},
		12:  struct{}{},
		26:  struct{}{},
		32:  struct{}{},
		45:  struct{}{},
		52:  struct{}{},
		57:  struct{}{},
		73:  struct{}{},
		76:  struct{}{},
		79:  struct{}{},
		100: struct{}{},
		123: struct{}{},
	}

	humanMods = map[int]int{
		7:  1,
		8:  2,
		9:  3,
		10: 4,
		11: 5,

		13: 1,
		14: 2,
		15: 3,
		16: 4,
		17: 5,

		18: 1,
		19: 2,
		20: 3,
		21: 4,
		22: 5,

		23: 1,
		24: 2,
		25: 3,

		27: 1,
		28: 2,
		29: 3,
		30: 4,
		31: 5,
	}
)

func locationToLed(location int) (int, int) {
	loc := location % (render.ROWS * render.COLUMNS)
	row := loc / render.COLUMNS
	col := loc - row*render.COLUMNS

	return row, col
}

type Ceottk struct {
	location int
	output   beatboxer.Output
	playing  bool
}

func (c *Ceottk) Init(output beatboxer.Output) {
	c.output = output
}

func (c *Ceottk) Pressed(row int, column int) {
	if c.playing {
		return
	}
	c.playing = true

	c.location++

	loc := c.location
	if location, ok := humanMods[c.location]; ok {
		loc = location
	}

	human := fmt.Sprintf("ceottk%03d_human.wav", loc)

	dur := c.output.Play(human)

	rs := render.RenderState{}
	row, col := locationToLed(c.location)
	rs.LEDs[row][col] = color.Make(127, 0, 0, 0)
	c.output.Render(rs)

	time.AfterFunc(time.Duration(float64(dur)*EARLY_PLAY), func() {
		if _, ok := aliens[c.location+1]; ok {
			c.location++
			alien := fmt.Sprintf("ceottk%03d_alien.wav", c.location)
			dur := c.output.Play(alien)

			rs := render.RenderState{}
			row, col := locationToLed(c.location)
			rs.LEDs[row][col] = color.Make(127, 127, 0, 127)
			c.output.Render(rs)

			time.AfterFunc(dur, func() {
				// this works because we know the last sound played is alien
				if c.location == SEQEUNCE_LENGTH {
					c.location = 0 // TODO: reset or have harness recreate program
					c.output.Yield()
				}
				c.playing = false
			})
		} else {
			c.playing = false
		}
	})
}
