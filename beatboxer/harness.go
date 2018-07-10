package beatboxer

import (
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
	program Program
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
	pressed chan bbox.Coord
	kb      *Keyboard
	wavs    *wavs.Wavs

	programs []registered
	active   int
}

func InitHarness() *Harness {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	pressed := make(chan bbox.Coord)

	return &Harness{
		pressed: pressed,
		wavs:    wavs.InitWavs(),
		kb:      InitKeyboard(pressed, bbox.KeyMapsPC), // TODO: parameterize for KeyMapsPI
	}
}

func (h *Harness) Register(program Program) {
	id := len(h.programs)

	r := registered{
		harness: h,
		program: program,
		id:      id,
	}
	h.programs = append(h.programs, r)

	program.Init(&r)
}

func (h *Harness) Run() {
	defer termbox.Close()
	defer h.wavs.Close()

	go h.kb.Run()

	switcherCount := 0
	for {
		select {
		case coord, more := <-h.pressed:
			if !more {
				return
			}
			if coord == switcher {
				switcherCount++
				if switcherCount >= SWITCH_COUNT {
					h.active = (h.active + 1) % len(h.programs)
					switcherCount = 0
				}
			} else {
				switcherCount = 0
			}

			h.programs[h.active].program.Pressed(coord[0], coord[1])
		}
	}
}

func (h *Harness) play(id int, name string) time.Duration {
	if id == h.active {
		return h.wavs.Play(name)
	}

	return time.Duration(0)
}

func (h *Harness) render(id int, rs render.RenderState) {
	if id == h.active {
		render.Render(rs)
	}
}

func (h *Harness) yield(id int) {
	if id == h.active {
		h.active = (h.active + 1) % len(h.programs)
	}
}
