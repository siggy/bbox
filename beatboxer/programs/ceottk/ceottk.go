package ceottk

import (
	"fmt"
	"math"
	"time"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/bbox/color"
	"github.com/siggy/bbox/beatboxer"
	"github.com/siggy/bbox/beatboxer/render"
	log "github.com/sirupsen/logrus"
)

const (
	SEQUENCE_LENGTH      = 123
	IMPATIENCE_THRESHOLD = 20
	TIMEOUT_THRESHOLD    = 180 * time.Second
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
	wavDurs map[string]time.Duration

	setLocation chan int
	getLocation chan int

	setPlayCh chan bool
	getPlayCh chan bool

	leds        chan [render.ROWS][render.COLUMNS]uint32
	transitions chan [render.ROWS][render.COLUMNS]render.Transition

	// input
	amp      chan float64
	keyboard chan bbox.Coord
	close    chan struct{}

	// output
	play   chan string
	render chan render.State
	yield  chan struct{}
}

func (c *Ceottk) New(wavDurs map[string]time.Duration) beatboxer.Program {
	ceottk := &Ceottk{
		wavDurs: wavDurs,

		setLocation: make(chan int),
		getLocation: make(chan int),

		setPlayCh: make(chan bool),
		getPlayCh: make(chan bool),

		leds:        make(chan [render.ROWS][render.COLUMNS]uint32),
		transitions: make(chan [render.ROWS][render.COLUMNS]render.Transition),

		// input
		amp:      make(chan float64),
		keyboard: make(chan bbox.Coord),
		close:    make(chan struct{}),

		// output
		play:   make(chan string),
		render: make(chan render.State),
		yield:  make(chan struct{}),
	}

	go ceottk.run()

	return ceottk
}

// input
func (c *Ceottk) Amplitude() chan<- float64 {
	return c.amp
}
func (c *Ceottk) Keyboard() chan<- bbox.Coord {
	return c.keyboard
}
func (c *Ceottk) Close() chan<- struct{} {
	return c.close
}

// output
func (c *Ceottk) Play() <-chan string {
	return c.play
}
func (c *Ceottk) Render() <-chan render.State {
	return c.render
}
func (c *Ceottk) Yield() <-chan struct{} {
	return c.yield
}

func (c *Ceottk) run() {
	go c.runLocation()
	go c.runPlaying()
	go c.runKB()
	go c.runAmp()

	rs := render.State{}
	for {
		select {
		case leds := <-c.leds:
			rs.LEDs = leds
		case transitions := <-c.transitions:
			rs.Transitions = transitions
		case <-c.close:
			return
		}

		c.render <- rs
	}
}

func (c *Ceottk) runAmp() {
	for {
		select {
		case level, _ := <-c.amp:
			log.Debugf("Ceottk:amp: %f", level)
			rs := render.State{}
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
		case <-c.close:
			return
		}
	}
}

func (c *Ceottk) runLocation() {
	location := 0
	for {
		select {
		case location = <-c.setLocation:
		case c.getLocation <- location:
		case <-c.close:
			return
		}
	}
}

func (c *Ceottk) incLoc() int {
	location := <-c.getLocation
	newLoc := (location + 1) % (SEQUENCE_LENGTH + 1) ////
	c.setLocation <- newLoc

	return newLoc
}

func (c *Ceottk) runPlaying() {
	playing := false
	for {
		select {
		case playing = <-c.setPlayCh:
		case c.getPlayCh <- playing:
		case <-c.close:
			return
		}
	}
}

func (c *Ceottk) getPlaying() bool {
	return <-c.getPlayCh
}

func (c *Ceottk) setPlaying(playing bool) {
	c.setPlayCh <- playing
}

func (c *Ceottk) runKB() {
	impatience := 0

	for {
		select {
		case coord, _ := <-c.keyboard:
			row := coord[0]
			col := coord[1]

			if c.getPlaying() {
				impatience++
				if impatience > IMPATIENCE_THRESHOLD {
					c.yield <- struct{}{}
					return
				}
				break
			}
			c.setPlaying(true)
			impatience = 0

			loc := c.incLoc()
			humanLoc := loc
			if mod, ok := humanMods[loc]; ok {
				humanLoc = mod
			}

			human := fmt.Sprintf("ceottk%03d_human.wav", humanLoc)

			c.play <- human

			rs := render.State{}
			rs.LEDs[row][col] = color.Make(127, 0, 0, 0)
			c.leds <- rs.LEDs

			go func() {
				if _, ok := aliens[loc+1]; !ok {
					c.setPlaying(false)
					return
				}

				location := c.incLoc()
				alien := fmt.Sprintf("ceottk%03d_alien.wav", location)

				c.play <- alien

				rowStart := int(math.Max(0, float64(row)-1))
				rowEnd := int(math.Min(render.ROWS-1, float64(row)+1))
				colStart := col - 1
				colEnd := col + 1

				for aRow := rowStart; aRow <= rowEnd; aRow++ {
					for aCol := colStart; aCol <= colEnd; aCol++ {
						c := (aCol + render.COLUMNS) % render.COLUMNS
						if aRow != row || c != col {
							rs.LEDs[aRow][c] = color.Make(250, 143, 94, 0)
						}
					}
				}

				c.leds <- rs.LEDs

				time.AfterFunc(c.wavDurs[alien], func() {
					// this works because we know the last sound played is alien
					if location == SEQUENCE_LENGTH {
						c.incLoc()
						c.yield <- struct{}{}
					}
					c.setPlaying(false)
				})
			}()

		case <-time.After(TIMEOUT_THRESHOLD):
			c.yield <- struct{}{}

		case <-c.close:
			return
		}
	}
}
