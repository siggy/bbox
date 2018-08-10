package intro

import (
	"math/rand"
	"time"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer"
	"github.com/siggy/bbox/beatboxer/color"
	"github.com/siggy/bbox/beatboxer/render"
)

const (
	POWER_SOURCES = 4
	POWER_CYCLE   = 500 * time.Millisecond
	KEEP_ALIVE    = 14 * time.Minute
)

type Intro struct {
	// input
	amp      chan float64
	keyboard chan bbox.Coord
	close    chan struct{}

	// output
	play   chan string
	render chan render.State
	yield  chan struct{}
}

type power struct {
	row int
	col int
}

func (c *Intro) New(wavDurs map[string]time.Duration) beatboxer.Program {
	intro := &Intro{
		// input
		amp:      make(chan float64),
		keyboard: make(chan bbox.Coord),
		close:    make(chan struct{}),

		// output
		play:   make(chan string),
		render: make(chan render.State),
		yield:  make(chan struct{}),
	}

	go intro.run()

	return intro
}

// input
func (c *Intro) Amplitude() chan<- float64 {
	return c.amp
}
func (c *Intro) Keyboard() chan<- bbox.Coord {
	return c.keyboard
}
func (c *Intro) Close() chan<- struct{} {
	return c.close
}

// output
func (c *Intro) Play() <-chan string {
	return c.play
}
func (c *Intro) Render() <-chan render.State {
	return c.render
}
func (c *Intro) Yield() <-chan struct{} {
	return c.yield
}

func (c *Intro) run() {
	powerTicker := time.NewTicker(POWER_CYCLE)
	defer powerTicker.Stop()

	ticker := time.NewTicker(KEEP_ALIVE)
	defer ticker.Stop()

	rs := render.State{}

	powers := map[power]struct{}{}
	for i := 0; i < POWER_SOURCES; i++ {
		row := rand.Intn(render.ROWS)
		col := rand.Intn(render.COLUMNS)
		powers[power{
			row,
			col,
		}] = struct{}{}

		for p := range powers {
			rs.LEDs[p.row][p.col] = color.Colors[rand.Intn(len(color.Colors))]
		}
	}

	for {
		select {
		case <-powerTicker.C:
			powers = map[power]struct{}{}

			for i := 0; i < POWER_SOURCES; i++ {
				row := rand.Intn(render.ROWS)
				col := rand.Intn(render.COLUMNS)
				powers[power{
					row,
					col,
				}] = struct{}{}

				for p := range powers {
					rs.LEDs[p.row][p.col] = color.Colors[rand.Intn(len(color.Colors))]
				}
			}
		case <-ticker.C:
			c.play <- "kick-classic.wav"
		case <-c.keyboard:
			c.yield <- struct{}{}
			return
		case <-c.close:
			return
		default:
		}

		newRs := rs

		for row := 0; row < render.ROWS; row++ {
			for col := 0; col < render.COLUMNS; col++ {
				if _, ok := powers[power{row, col}]; ok {
					continue
				}
				rowPre := (row - 1 + render.ROWS) % render.ROWS
				rowPost := (row + 1) % render.ROWS
				colPre := (col - 1 + render.COLUMNS) % render.COLUMNS
				colPost := (col + 1) % render.COLUMNS

				// hack to prevent row wrapping
				if row == 0 {
					rowPre = 0
				} else if row == 3 {
					rowPost = 3
				}

				newColor := color.Color{}
				for _, r := range []int{rowPre, row, rowPost} {
					for _, c := range []int{colPre, col, colPost} {
						s := color.Split(rs.LEDs[r][c])
						newColor.R += s.R
						newColor.G += s.G
						newColor.B += s.B
						newColor.W += s.W
					}
				}

				newRs.LEDs[row][col] = color.Make(
					newColor.R/9,
					newColor.G/9,
					newColor.B/9,
					newColor.W/9,
				)
			}
		}

		rs = newRs
		c.render <- rs
	}
}
