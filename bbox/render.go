package bbox

import (
	"github.com/nsf/termbox-go"
)

type Render struct {
	beats Beats
	msgs  <-chan Beats
	ticks <-chan int
}

func InitRender(msgs <-chan Beats, ticks <-chan int) *Render {
	return &Render{
		msgs:  msgs,
		ticks: ticks,
	}
}

func (r *Render) Draw() {
	for i := 0; i < BEATS; i++ {
		for j := 0; j < TICKS; j++ {
			c := '-'
			if r.beats[i][j] {
				c = 'X'
			}
			termbox.SetCell(j*2, i+1, c, termbox.ColorDefault, termbox.ColorDefault)
			termbox.SetCell(j*2+1, i+1, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	termbox.Flush()
}

func (r *Render) Run() {
	for i := 0; i < BEATS+1; i++ {
		for j := 0; j < TICKS*2; j++ {
			termbox.SetCell(j, i, '-', termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	termbox.Flush()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for {
		select {
		case tick := <-r.ticks:
			termbox.SetCell((tick+TICKS-1)%TICKS*2, 0, ' ', termbox.ColorDefault, termbox.ColorDefault)
			termbox.SetCell(tick*2, 0, 'O', termbox.ColorDefault, termbox.ColorDefault)
			termbox.Flush()
		case beats, more := <-r.msgs:
			if more {
				// incoming beat update from keyboard
				r.beats = beats
				r.Draw()
			} else {
				// closing
				return
			}
		}
	}
}
