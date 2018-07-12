package ceottk

import (
	"fmt"
	"math"
	"time"

	"github.com/siggy/bbox/beatboxer"
	"github.com/siggy/bbox/beatboxer/color"
	"github.com/siggy/bbox/beatboxer/render"
)

const (
	SEQUENCE_LENGTH      = 123
	EARLY_PLAY           = 0.5
	IMPATIENCE_THRESHOLD = 100
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

type Ceottk struct {
	location   int
	output     beatboxer.Output
	playing    bool
	impatience int

	closing     chan struct{}
	leds        chan [render.ROWS][render.COLUMNS]uint32
	transitions chan [render.ROWS][render.COLUMNS]render.Transition
}

func (c *Ceottk) New(output beatboxer.Output) beatboxer.Program {
	ceottk := &Ceottk{
		output:      output,
		closing:     make(chan struct{}),
		leds:        make(chan [render.ROWS][render.COLUMNS]uint32),
		transitions: make(chan [render.ROWS][render.COLUMNS]render.Transition),
	}

	go ceottk.run()

	return ceottk
}

func (c *Ceottk) Amp(level float64) {
	rs := render.RenderState{}
	amp := int(math.Min(level*4+1, 4))
	for row := render.ROWS - 1; row > (render.ROWS - 1 - amp); row-- {
		for col := 0; col < render.COLUMNS; col++ {
			rs.Transitions[row][col] = render.Transition{
				// blue -> purple -> red
				// 40, 32, 240 -> 160, 32, 34
				Color: color.Make(
					uint32(40*(4-row)),
					32,
					uint32(-12.875*float64(col)+240),
					0,
				),
				Location: level, // !!!!!
				Length:   level,
			}
		}
	}
	c.transitions <- rs.Transitions
}

func (c *Ceottk) Pressed(row int, col int) {
	if c.playing {
		c.impatience++
		if c.impatience > IMPATIENCE_THRESHOLD {
			c.output.Yield()
		}
		return
	}
	c.playing = true
	c.impatience = 0

	c.location++

	loc := c.location
	if location, ok := humanMods[c.location]; ok {
		loc = location
	}

	human := fmt.Sprintf("ceottk%03d_human.wav", loc)

	dur := c.output.Play(human)

	rs := render.RenderState{}
	rs.LEDs[row][col] = color.Make(127, 0, 0, 0)
	c.leds <- rs.LEDs

	time.AfterFunc(time.Duration(float64(dur)*EARLY_PLAY), func() {
		if _, ok := aliens[c.location+1]; ok {
			c.location++
			alien := fmt.Sprintf("ceottk%03d_alien.wav", c.location)
			dur := c.output.Play(alien)

			for aRow := int(math.Max(0, float64(row)-1)); aRow <= int(math.Min(render.ROWS-1, float64(row)+1)); aRow++ {
				for aCol := int(math.Max(0, float64(col)-1)); aCol <= int(math.Min(render.COLUMNS-1, float64(col)+1)); aCol++ {
					if aRow != row || aCol != col {
						rs.LEDs[aRow][aCol] = color.Make(250, 143, 94, 0)
					}
				}
			}

			c.leds <- rs.LEDs

			time.AfterFunc(dur, func() {
				// this works because we know the last sound played is alien
				if c.location == SEQUENCE_LENGTH {
					c.output.Yield()
				}
				c.playing = false
			})
		} else {
			c.playing = false
		}
	})
}

func (c *Ceottk) Close() {
	close(c.closing)
}

func (c *Ceottk) run() {
	rs := render.RenderState{}
	for {
		select {
		case leds := <-c.leds:
			rs.LEDs = leds
		case transitions := <-c.transitions:
			rs.Transitions = transitions
		case _, more := <-c.closing:
			if !more {
				return
			}
		}

		c.output.Render(rs)
	}
}
