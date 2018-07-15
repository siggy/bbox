package beatboxer

import (
	"errors"
	"sync"

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

// type registered struct {
// 	harness *Harness
// 	id      int
// }

// satisfy Output interface
// func (r *registered) Play(name string) time.Duration {
// 	return r.harness.play(r.id, name)
// }
// func (r *registered) Render(rs render.RenderState) {
// 	log.Debugf("(r *registered) Render start: %+d", r.id)
// 	r.harness.render(r.id, rs)
// 	log.Debugf("(r *registered) Render start: %+d", r.id)
// }
// func (r *registered) Yield() {
// 	r.harness.yield(r.id)
// }

type harness struct {
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
	// active    int
}

func InitHarness(
	renderer render.Renderer,
	keyMap map[bbox.Key]*bbox.Coord,
) *harness {
	// err := termbox.Init()
	// if err != nil {
	// 	panic(err)
	// }

	kb := keyboard.Init(keyMap)

	return &harness{
		renderer: renderer,
		// pressed:   make(chan bbox.Coord),
		termRender: make(chan render.RenderState),
		wavs:       wavs.InitWavs(),
		keyMap:     keyMap,
		amplitude:  InitAmplitude(),
		kb:         kb,
		terminal:   render.InitTerminal(kb),
		// flush:     make(chan struct{}),
	}
}

func (h *harness) Register(program Program) {
	h.programs = append(h.programs, program)
}

// func (h *harness) NextProgram() {
// 	prev := h.programs[h.active]
// 	h.active = (h.active + 1) % len(h.programs)
// 	prev.Close()

// 	// clear the display
// 	// h.renderFn(render.RenderState{})
// 	h.terminal.Render(render.RenderState{})

// 	reg := registered{
// 		harness: h,
// 		id:      h.active,
// 	}
// 	h.programs[h.active] = h.programs[h.active].New(&reg)
// }

// func (h *harness) Run() {
// 	// err := termbox.Init()
// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	// h.kb = keyboard.Init(h.keyMap)

// 	go h.amplitude.Run()
// 	go h.kb.Run()

// 	// h.terminal = render.InitTerminal(h.kb)

// 	defer func() {
// 		log.Debugf("h.Run() defer func() 0")

// 		// termbox.Interrupt()
// 		// termbox.Close()
// 		prev := h.programs[h.active]
// 		// don't actually start the next program
// 		h.active = (h.active + 1) % len(h.programs)
// 		prev.Close()

// 		log.Debugf("h.Run() defer func() 1")

// 		// ensure nested shutdown for portaudio even though it shouldn't be necessary?
// 		go func() {
// 			h.amplitude.Close()
// 			h.wavs.Close()
// 		}()

// 		go h.kb.Close()

// 		log.Debugf("h.Run() defer func() 2")
// 	}()

// 	// make the first program active
// 	reg := registered{
// 		harness: h,
// 		id:      0,
// 	}
// 	h.programs[0] = h.programs[0].New(&reg)

// 	switcherCount := 0
// 	for {
// 		select {
// 		// case _, more := <-h.flush:
// 		// 	if !more {
// 		// 		log.Debugf("flush channel closed")
// 		// 		return
// 		// 	}
// 		// 	termbox.Flush()
// 		case level, more := <-h.level:
// 			log.Debugf("h.level")
// 			if !more {
// 				log.Debugf("harness: amplitude.level channel closed")
// 				return
// 			}

// 			// log.Debugf("h.programs[h.active].Amp(level) start")

// 			h.programs[h.active].Amp(level)

// 			// log.Debugf("h.programs[h.active].Amp(level) end")
// 		case coord, more := <-h.kb.Pressed():
// 			log.Debugf("h.kb.Pressed(): %+v", coord)
// 			if !more {
// 				log.Debugf("harness: pressed channel closed")
// 				return
// 			}
// 			if coord == switcher {
// 				switcherCount++

// 				if switcherCount >= SWITCH_COUNT {
// 					h.NextProgram()
// 					switcherCount = 0
// 					continue
// 				}
// 			} else {
// 				switcherCount = 0
// 			}

// 			h.programs[h.active].Pressed(coord[0], coord[1])
// 		case _, more := <-h.kb.Closing():
// 			log.Debugf("<-h.kb.Closing()")
// 			if !more {
// 				log.Debugf("harness: keyboard closing")
// 				return
// 			}
// 		case rs, more := <-h.termRender:
// 			log.Debugf("<-h.termRender")
// 			if !more {
// 				log.Debugf("harness: termRender channel closed")
// 				return
// 			}
// 			log.Debugf("h.terminal.Render(rs) start")
// 			h.terminal.Render(rs)
// 			log.Debugf("h.terminal.Render(rs) end")
// 		}
// 	}
// }

// func (h *harness) play(id int, name string) time.Duration {
// 	if id != h.active {
// 		log.Debugf("play called by invalid program %d: %s", id, name)
// 		return time.Duration(0)
// 	}

// 	return h.wavs.Play(name)
// }

// func (h *harness) render(id int, rs render.RenderState) {
// 	if id != h.active {
// 		log.Debugf("render called by invalid program %d: %+v", id, rs)
// 		return
// 	}

// 	// TODO: led renderer, which eventually call renderer.SetLed
// 	// h.led.Render(rs)

// 	log.Debugf("(h *Harness) render start: %+d", id)
// 	h.termRender <- rs
// 	log.Debugf("(h *Harness) render end: %+d", id)
// 	// h.renderFn(rs)

// 	// TODO: should renderFn() do this?
// 	// also TODO: should this be synch?
// 	// h.kb.Flush()

// 	// TODO: decide if a web renderer is performant enough
// 	// h.toRenderer(rs)
// }

// temporary until all the "68, 64, 60, 56" foo is moved over
func (h *harness) toRenderer(rs render.RenderState) {
	for col := 0; col < render.COLUMNS; col++ {
		for row := 0; row < render.ROWS-2; row++ {
			h.renderer.SetLed(0, col, rs.LEDs[row][col])
		}
		for row := render.ROWS - 2; row < render.ROWS; row++ {
			h.renderer.SetLed(1, col, rs.LEDs[row][col])
		}
	}
}

// func (h *harness) yield(id int) {
// 	if id != h.active {
// 		log.Debugf("yield called by invalid program %d", id)
// 		return
// 	}

// 	h.NextProgram()
// }

func (h *harness) Run() {
	go h.amplitude.Run()
	go h.kb.Run()

	defer func() {
		h.amplitude.Close()
		h.wavs.Close()
	}()
	defer h.kb.Close()

	active := 0
	cur := h.programs[active].New()

	for {
		err := h.RunProgram(cur)
		go func(cur Program) {
			cur.Close()
		}(cur)
		if err != nil {
			break
		}

		active = (active + 1) % len(h.programs)
		cur = h.programs[active].New()
	}
}

func (h *harness) RunProgram(p Program) error {
	var err error
	yielding := make(chan struct{})
	exiting := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(5)

	// input: amplitude
	go func() {
		log.Debugf("h.RunProgram AMP")
		// wg.Add(1)
		defer wg.Done()

		for {
			log.Debugf("h.RunProgram AMP 1")
			select {
			case p.Amplitude() <- <-h.amplitude.Level():
				// log.Debugf("h.RunProgram AMP 2")
				// p.Amplitude() <- a
				// log.Debugf("h.RunProgram AMP 3")
			case _, more := <-yielding:
				if !more {
					return
				}
			case _, more := <-exiting:
				if !more {
					log.Debugf("h.RunProgram AMP 4")
					return
				}
			}
		}
	}()

	// input: keyboard
	go func() {
		log.Debugf("h.RunProgram KB")

		// wg.Add(1)
		defer wg.Done()

		for {
			log.Debugf("h.RunProgram KB 1")
			select {
			case coord, _ := <-h.kb.Pressed():
				log.Debugf("h.RunProgram KB 2")
				p.Keyboard() <- coord
				log.Debugf("h.RunProgram KB 3")
			case _, more := <-h.kb.Closing():
				log.Debugf("h.RunProgram KB 4")
				if !more {
					close(exiting)
					err = errors.New("Exiting")
					log.Debugf("h.RunProgram KB 5")
					return
				}
			case _, more := <-yielding:
				if !more {
					return
				}
			}
		}
	}()

	// output: render
	go func() {
		log.Debugf("h.RunProgram RENDER")

		// wg.Add(1)
		defer wg.Done()

		for {
			// log.Debugf("h.RunProgram RENDER 1")
			select {
			case rs, _ := <-p.Render():
				// log.Debugf("h.RunProgram RENDER 2")
				h.terminal.Render(rs)
			case _, more := <-yielding:
				if !more {
					return
				}
			case _, more := <-exiting:
				if !more {
					log.Debugf("h.RunProgram RENDER 3")
					return
				}
			}
		}
	}()

	// output: play
	go func() {
		log.Debugf("h.RunProgram PLAY")

		// wg.Add(1)
		defer wg.Done()

		for {
			log.Debugf("h.RunProgram PLAY 1")
			select {
			case name, _ := <-p.Play():
				log.Debugf("h.RunProgram PLAY 2")
				h.wavs.Play(name)
			case _, more := <-yielding:
				if !more {
					return
				}
			case _, more := <-exiting:
				if !more {
					log.Debugf("h.RunProgram PLAY 3")
					return
				}
			}
		}
	}()

	// output: yield
	go func() {
		log.Debugf("h.RunProgram YIELD")

		// wg.Add(1)
		defer wg.Done()

		for {
			log.Debugf("h.RunProgram YIELD 1")
			select {
			case <-p.Yield():
				log.Debugf("h.RunProgram YIELD 2")
				close(yielding)
				return
				// go func() {
				// 	p.Close() <- struct{}{}
				// }()

				// h.terminal.Render(render.RenderState{})

				// active = (active + 1) % len(h.programs)
				// cur = h.programs[active].New()
			case _, more := <-exiting:
				if !more {
					log.Debugf("h.RunProgram YIELD 3")
					return
				}
			}
		}
	}()

	log.Debugf("h.RunProgram wg.Wait1")

	wg.Wait()

	log.Debugf("h.RunProgram wg.Wait2")

	return err
}
