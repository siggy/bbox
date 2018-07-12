package beatboxer

import (
	"fmt"
	"time"

	termbox "github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer/render"
	"github.com/siggy/bbox/beatboxer/wavs"
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
	r.harness.render(r.id, rs)
}
func (r *registered) Yield() {
	r.harness.yield(r.id)
}

type Harness struct {
	renderer  render.Renderer
	renderFn  func(render.RenderState)
	pressed   chan bbox.Coord
	kb        *Keyboard
	wavs      *wavs.Wavs
	amplitude *Amplitude
	programs  []Program
	active    int
	level     chan float64
}

func InitHarness(
	renderer render.Renderer,
	renderFn func(render.RenderState),
	keyMap map[bbox.Key]*bbox.Coord,
) *Harness {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	level := make(chan float64)
	amplitude := InitAmplitude(level)
	pressed := make(chan bbox.Coord)

	return &Harness{
		renderer:  renderer,
		renderFn:  renderFn,
		pressed:   pressed,
		wavs:      wavs.InitWavs(),
		kb:        InitKeyboard(pressed, keyMap),
		amplitude: amplitude,
		level:     level,
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
	h.renderFn(render.RenderState{})

	reg := registered{
		harness: h,
		id:      h.active,
	}
	h.programs[h.active] = h.programs[h.active].New(&reg)
}

func (h *Harness) Run() {
	defer termbox.Close()
	defer h.wavs.Close()
	defer h.amplitude.Close()

	go h.amplitude.Run()
	go h.kb.Run()

	// make the first program active
	reg := registered{
		harness: h,
		id:      0,
	}
	h.programs[0] = h.programs[0].New(&reg)

	switcherCount := 0
	for {
		select {
		case level, more := <-h.level:
			if !more {
				fmt.Printf("amplitude.level closed\n")
				return
			}
			h.programs[h.active].Amp(level)
		case coord, more := <-h.pressed:
			if !more {
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
		}
	}
}

func (h *Harness) play(id int, name string) time.Duration {
	if id != h.active {
		fmt.Printf("play called by invalid program %d: %s", id, name)
		return time.Duration(0)
	}

	return h.wavs.Play(name)
}

func (h *Harness) render(id int, rs render.RenderState) {
	if id != h.active {
		fmt.Printf("render called by invalid program %d: %+v", id, rs)
		return
	}

	h.renderFn(rs)

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
		fmt.Printf("yield called by invalid program %d", id)
		return
	}

	h.NextProgram()
}
