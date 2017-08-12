package bbox

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

const (
	TICK_DELAY = 2
)

type Render struct {
	beats   Beats
	closing chan struct{}
	msgs    <-chan Beats
	tick    int
	ticks   <-chan int
}

func InitRender(msgs <-chan Beats, ticks <-chan int) *Render {
	return &Render{
		closing: make(chan struct{}),
		msgs:    msgs,
		ticks:   ticks,
	}
}

func (r *Render) Draw() {
	oldTick := (r.tick + TICK_DELAY - 1) % TICKS
	newTick := (r.tick + TICK_DELAY) % TICKS
	termbox.SetCell(oldTick, 0, ' ', termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(newTick, 0, 'O', termbox.ColorBlack, termbox.ColorWhite)

	for i := 0; i < SOUNDS; i++ {
		// render all beats, slightly redundant with below
		for j := 0; j < BEATS; j++ {
			c := '-'
			back := termbox.ColorDefault
			fore := termbox.ColorDefault
			if r.beats[i][j] {
				c = 'X'
				back = termbox.ColorRed
				fore = termbox.ColorBlack
			}
			termbox.SetCell(j*TICKS_PER_BEAT, i+1, c, fore, back)
		}

		// render all runes in old and new columns
		oldRune := ' '
		oldBack := termbox.ColorDefault
		oldFore := termbox.ColorDefault

		newRune := '.'
		newBack := termbox.ColorWhite
		newFore := termbox.ColorBlack
		if oldTick%TICKS_PER_BEAT == 0 {
			// old tick is on a beat
			if r.beats[i][oldTick/TICKS_PER_BEAT] {
				// not ticked, activated
				oldRune = 'X'
				oldBack = termbox.ColorRed
				oldFore = termbox.ColorBlack
			} else {
				// not ticked, not activated
				oldRune = '-'
			}
		} else if newTick%TICKS_PER_BEAT == 0 {
			// new tick is on a beat
			if r.beats[i][newTick/TICKS_PER_BEAT] {
				// ticked, activated
				newRune = '8'
				newBack = termbox.ColorMagenta
			} else {
				// ticked, not activated
				newRune = '_'
			}
		}
		termbox.SetCell(oldTick, i+1, oldRune, oldFore, oldBack)
		termbox.SetCell(newTick, i+1, newRune, newFore, newBack)
	}

	termbox.Flush()
}

func (r *Render) Run() {
	// termbox.Init() called in InitKeyboard()
	defer termbox.Close()

	for {
		select {
		case _, more := <-r.closing:
			if !more {
				return
			}
		case tick := <-r.ticks:
			r.tick = tick
			r.Draw()
		case beats, more := <-r.msgs:
			if more {
				// incoming beat update from keyboard
				r.beats = beats
				r.Draw()
			} else {
				// closing
				fmt.Printf("Render closing\n")
				return
			}
		}
	}
}

func (r *Render) Close() {
	// TODO: this doesn't block?
	close(r.closing)
}
