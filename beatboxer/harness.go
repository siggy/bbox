package beatboxer

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer/render"
)

type Harness struct {
	pressed  chan bbox.Coord
	kb       *Keyboard
	programs []Program
}

func InitHarness() *Harness {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	h := Harness{}

	h.pressed = make(chan bbox.Coord)
	h.kb = InitKeyboard(h.pressed, bbox.KeyMapsPC)

	return &h
}

func (h *Harness) Register(program Program) {
	program.Init(render.Render)
	h.programs = append(h.programs, program)
}

func (h *Harness) Run() {
	defer termbox.Close()

	go h.kb.Run()

	for {
		select {
		case coord, more := <-h.pressed:
			if !more {
				return
			}

			for _, program := range h.programs {
				program.Pressed(coord[0], coord[1])
			}
		}
	}
}
