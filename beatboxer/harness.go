package beatboxer

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer/render"
	"github.com/siggy/bbox/beatboxer/wavs"
)

type Harness struct {
	pressed  chan bbox.Coord
	kb       *Keyboard
	programs []Program // TODO: designate one as active
	wavs     *wavs.Wavs
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
	program.Init(h.wavs, render.Render)
	h.programs = append(h.programs, program)
}

func (h *Harness) Run() {
	defer termbox.Close()
	defer h.wavs.Close()

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
