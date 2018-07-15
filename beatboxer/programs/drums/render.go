package drums

import (
	"github.com/siggy/bbox/beatboxer/color"
	"github.com/siggy/bbox/beatboxer/render"
	log "github.com/sirupsen/logrus"
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

	iv         Interval
	intervalCh <-chan Interval

	render chan<- render.RenderState
}

func InitRender(
	msgs <-chan Beats,
	ticks <-chan int,
	intervalCh <-chan Interval,
	render chan<- render.RenderState,
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
	renderState := render.RenderState{}

	newTick := (r.tick + TICK_DELAY) % r.iv.Ticks
	newLed := newTick / r.iv.TicksPerBeat

	transition := render.Transition{
		Color:    color.Make(0, 0, 0, 127),
		Location: float64(newTick-(newLed*r.iv.TicksPerBeat)) / float64(r.iv.TicksPerBeat),
		Length:   0.5,
	}

	renderState.Transitions[0][newLed] = transition
	renderState.Transitions[1][newLed] = transition
	renderState.Transitions[2][newLed] = transition
	renderState.Transitions[3][newLed] = transition

	for i := 0; i < SOUNDS; i++ {
		// render all beats, slightly redundant with below
		for j := 0; j < BEATS; j++ {
			if r.beats[i][j] {
				if j == newLed {
					renderState.LEDs[i][j] = color.Make(127, 127, 0, 127)
				} else {
					renderState.LEDs[i][j] = color.Make(127, 0, 0, 0)
				}
			}
		}
	}

	// log.Debugf("Render.Draw(): calling r.render() start")
	r.render <- renderState
	// log.Debugf("Render.Draw(): calling r.render() end")
}

func (r *Render) Run() {
	for {
		select {
		case _, more := <-r.closing:
			if !more {
				log.Debugf("Render.Run(): <-r.closing")
				return
			}
		case tick := <-r.ticks:
			r.tick = tick
			// log.Debugf("Render.Run(): <-r.ticks start")
			r.Draw()
			// log.Debugf("Render.Run(): <-r.ticks end")
		case beats, more := <-r.msgs:
			// log.Debugf("Render.Run(): <-r.msgs")
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
