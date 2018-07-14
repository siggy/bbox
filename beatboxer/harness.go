package beatboxer

import (
	"time"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer/keyboard"
	"github.com/siggy/bbox/beatboxer/render"
	"github.com/siggy/bbox/beatboxer/wavs"
	log "github.com/sirupsen/logrus"
)

const (
	SWITCH_COUNT = 5
)

var (
	switcher = bbox.Coord{1, 15}
)

type registered struct {
	harness *Harness
	id      int
}

// satisfy Output interface
func (r *registered) Play(name string) time.Duration {
	return r.harness.play(r.id, name)
}
func (r *registered) Render(rs render.RenderState) {
	log.Debugf("(r *registered) Render start: %+d", r.id)
	r.harness.render(r.id, rs)
	log.Debugf("(r *registered) Render start: %+d", r.id)
}
func (r *registered) Yield() {
	r.harness.yield(r.id)
}

type Harness struct {
	renderer   render.Renderer
	terminal   *render.Terminal
	termRender chan render.RenderState
	// pressed   chan bbox.Coord
	kb *keyboard.Keyboard
	// flush     chan struct{}
	wavs      *wavs.Wavs
	keyMap    map[bbox.Key]*bbox.Coord
	amplitude *Amplitude
	programs  []Program
	active    int
	level     chan float64
}

func InitHarness(
	renderer render.Renderer,
	keyMap map[bbox.Key]*bbox.Coord,
) *Harness {
	// err := termbox.Init()
	// if err != nil {
	// 	panic(err)
	// }

	level := make(chan float64)
	amplitude := InitAmplitude(level)

	return &Harness{
		renderer: renderer,
		// pressed:   make(chan bbox.Coord),
		termRender: make(chan render.RenderState),
		wavs:       wavs.InitWavs(),
		keyMap:     keyMap,
		amplitude:  amplitude,
		level:      level,
		// flush:     make(chan struct{}),
	}
}

func (h *Harness) Register(program Program) {
	h.programs = append(h.programs, program)
}

func (h *Harness) NextProgram() {
	prev := h.programs[h.active]
	h.active = (h.active + 1) % len(h.programs)
	prev.Close()

	// clear the display
	// h.renderFn(render.RenderState{})
	h.terminal.Render(render.RenderState{})

	reg := registered{
		harness: h,
		id:      h.active,
	}
	h.programs[h.active] = h.programs[h.active].New(&reg)
}

func (h *Harness) Run() {
	// err := termbox.Init()
	// if err != nil {
	// 	panic(err)
	// }

	h.kb = keyboard.Init(h.keyMap)

	go h.amplitude.Run()
	go h.kb.Run()

	h.terminal = render.InitTerminal(h.kb)

	defer func() {
		log.Debugf("h.Run() defer func() 0")

		// termbox.Interrupt()
		// termbox.Close()
		prev := h.programs[h.active]
		// don't actually start the next program
		h.active = (h.active + 1) % len(h.programs)
		prev.Close()

		log.Debugf("h.Run() defer func() 1")

		// ensure nested shutdown for portaudio even though it shouldn't be necessary?
		go func() {
			h.amplitude.Close()
			h.wavs.Close()
		}()

		go h.kb.Close()

		log.Debugf("h.Run() defer func() 2")
	}()

	// make the first program active
	reg := registered{
		harness: h,
		id:      0,
	}
	h.programs[0] = h.programs[0].New(&reg)

	switcherCount := 0
	for {
		select {
		// case _, more := <-h.flush:
		// 	if !more {
		// 		log.Debugf("flush channel closed")
		// 		return
		// 	}
		// 	termbox.Flush()
		case level, more := <-h.level:
			log.Debugf("h.level")
			if !more {
				log.Debugf("harness: amplitude.level channel closed")
				return
			}

			// log.Debugf("h.programs[h.active].Amp(level) start")

			h.programs[h.active].Amp(level)

			// log.Debugf("h.programs[h.active].Amp(level) end")
		case coord, more := <-h.kb.Pressed():
			log.Debugf("h.kb.Pressed(): %+v", coord)
			if !more {
				log.Debugf("harness: pressed channel closed")
				return
			}
			if coord == switcher {
				switcherCount++

				if switcherCount >= SWITCH_COUNT {
					h.NextProgram()
					switcherCount = 0
					continue
				}
			} else {
				switcherCount = 0
			}

			h.programs[h.active].Pressed(coord[0], coord[1])
		case _, more := <-h.kb.Closing():
			log.Debugf("<-h.kb.Closing()")
			if !more {
				log.Debugf("harness: keyboard closing")
				return
			}
		case rs, more := <-h.termRender:
			log.Debugf("<-h.termRender")
			if !more {
				log.Debugf("harness: termRender channel closed")
				return
			}
			log.Debugf("h.terminal.Render(rs) start")
			h.terminal.Render(rs)
			log.Debugf("h.terminal.Render(rs) end")
		}
	}
}

func (h *Harness) play(id int, name string) time.Duration {
	if id != h.active {
		log.Debugf("play called by invalid program %d: %s", id, name)
		return time.Duration(0)
	}

	return h.wavs.Play(name)
}

func (h *Harness) render(id int, rs render.RenderState) {
	if id != h.active {
		log.Debugf("render called by invalid program %d: %+v", id, rs)
		return
	}

	// TODO: led renderer, which eventually call renderer.SetLed
	// h.led.Render(rs)

	log.Debugf("(h *Harness) render start: %+d", id)
	h.termRender <- rs
	log.Debugf("(h *Harness) render end: %+d", id)
	// h.renderFn(rs)

	// TODO: should renderFn() do this?
	// also TODO: should this be synch?
	// h.kb.Flush()

	// TODO: decide if a web renderer is performant enough
	// h.toRenderer(rs)
}

// temporary until all the "68, 64, 60, 56" foo is moved over
func (h *Harness) toRenderer(rs render.RenderState) {
	for col := 0; col < render.COLUMNS; col++ {
		for row := 0; row < render.ROWS-2; row++ {
			h.renderer.SetLed(0, col, rs.LEDs[row][col])
		}
		for row := render.ROWS - 2; row < render.ROWS; row++ {
			h.renderer.SetLed(1, col, rs.LEDs[row][col])
		}
	}
}

func (h *Harness) yield(id int) {
	if id != h.active {
		log.Debugf("yield called by invalid program %d", id)
		return
	}

	h.NextProgram()
}
