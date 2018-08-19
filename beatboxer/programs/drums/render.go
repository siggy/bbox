package drums

import (
	"github.com/siggy/bbox/bbox/color"
	"github.com/siggy/bbox/beatboxer/render"
	log "github.com/sirupsen/logrus"
)

const (
	TICK_DELAY = -6
)

type Render struct {
	beats   Beats
	closing chan struct{}
	msgs    <-chan Beats
	tick    int
	ticks   <-chan int

	iv         Interval
	intervalCh <-chan Interval

	render chan<- render.State
}

func InitRender(
	msgs <-chan Beats,
	ticks <-chan int,
	intervalCh <-chan Interval,
	render chan<- render.State,
) *Render {
	return &Render{
		closing: make(chan struct{}),
		msgs:    msgs,
		ticks:   ticks,

		iv: Interval{
			TicksPerBeat: DEFAULT_TICKS_PER_BEAT,
			Ticks:        DEFAULT_TICKS,
		},
		intervalCh: intervalCh,
		render:     render,
	}
}

func (r *Render) Draw() {
	state := render.State{}

	newTick := (r.tick + r.iv.Ticks + TICK_DELAY) % r.iv.Ticks
	newLed := newTick / r.iv.TicksPerBeat

	transition := render.Transition{
		Color:    color.Make(0, 0, 0, 127),
		Location: float64(newTick-(newLed*r.iv.TicksPerBeat)) / float64(r.iv.TicksPerBeat),
		Length:   0.5,
	}

	state.Transitions[0][newLed] = transition
	state.Transitions[1][newLed] = transition
	state.Transitions[2][newLed] = transition
	state.Transitions[3][newLed] = transition

	for i := 0; i < SOUNDS; i++ {
		// render all beats, slightly redundant with below
		for j := 0; j < BEATS; j++ {
			if r.beats[i][j] {
				if j == newLed {
					state.LEDs[i][j] = color.ActiveBeatPurple
				} else {
					state.LEDs[i][j] = color.Make(127, 0, 0, 0)
				}
			}
		}
	}

	r.render <- state
}

func (r *Render) Run() {
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
				log.Debugf("Render.msgs closed")
				return
			}
		case iv, more := <-r.intervalCh:
			if more {
				// incoming interval update from loop
				r.iv = iv
			} else {
				// we should never get here
				log.Debugf("unexpected: intervalCh return no more")
				return
			}
		}
	}
}

func (r *Render) Close() {
	// TODO: this doesn't block?
	close(r.closing)
}
