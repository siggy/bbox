package beatboxer

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer/keyboard"
	"github.com/siggy/bbox/beatboxer/render"
	"github.com/siggy/bbox/beatboxer/wavs"
)

const (
	SWITCH_COUNT = 5
)

var (
	switcher = bbox.Coord{1, 15}
)

type Harness struct {
	renderers []render.Renderer
	kb        *keyboard.Keyboard
	wavs      *wavs.Wavs
	amplitude *Amplitude
	programs  []Program
	active    uint32
}

func InitHarness(
	renderers []render.Renderer,
	kb *keyboard.Keyboard,
) *Harness {

	return &Harness{
		renderers: renderers,
		wavs:      wavs.InitWavs(),
		amplitude: InitAmplitude(),
		kb:        kb,
	}
}

func (h *Harness) Register(program Program) {
	h.programs = append(h.programs, program)
}

func (h *Harness) Run() {
	go h.amplitude.Run()
	go h.kb.Run()

	defer func() {
		h.amplitude.Close()
		h.wavs.Close()
	}()
	defer h.kb.Close()

	active := atomic.LoadUint32(&h.active)
	cur := h.programs[active].New(active, h, h.wavs.Durations())

	for {
		err := h.runProgram(cur)
		go cur.Close()
		if err != nil {
			break
		}

		h.wavs.StopAll()
		for _, renderer := range h.renderers {
			renderer.Render(render.State{})
		}

		newID := atomic.AddUint32(&h.active, 1)
		h.programs[newID%uint32(len(h.programs))].New(newID, h, h.wavs.Durations())
	}
}

func (h *Harness) runProgram(p Program) error {
	closing := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(2)

	go h.runAmp(p, &wg, closing)
	err := h.runKB(p, &wg, closing)

	wg.Wait()

	return err
}

// input: amplitude
func (h *Harness) runAmp(p Program, wg *sync.WaitGroup, closing chan struct{}) {
	defer wg.Done()

	for {
		select {
		case amp, _ := <-h.amplitude.Level():
			p.Amplitude(amp)
		case <-closing:
			return
		}
	}
}

// input: keyboard
func (h *Harness) runKB(p Program, wg *sync.WaitGroup, closing chan struct{}) error {
	defer wg.Done()

	switcherCount := 0

	for {
		select {
		case coord, _ := <-h.kb.Pressed():
			if coord == switcher {
				switcherCount++
				if switcherCount >= SWITCH_COUNT {
					close(closing)
					return nil
				}
			} else {
				switcherCount = 0
			}

			p.Keyboard(coord)
		case <-h.kb.Closing():
			close(closing)
			return errors.New("Exiting")
		case <-closing:
			return nil
		}
	}
}

func (h *Harness) Play(id uint32, name string) {
	if id != atomic.LoadUint32(&h.active) {
		return
	}

	h.wavs.Play(name)
}
func (h *Harness) Render(id uint32, rs render.State) {
	if id != atomic.LoadUint32(&h.active) {
		return
	}
	for _, renderer := range h.renderers {
		renderer.Render(rs)
	}
}
func (h *Harness) Yield(id uint32) {
	active := atomic.LoadUint32(&h.active)
	if id != active {
		return
	}

	cur := h.programs[active%uint32(len(h.programs))]
	go func(cur Program) {
		cur.Close()
	}(cur)

	h.wavs.StopAll()
	for _, renderer := range h.renderers {
		renderer.Render(render.State{})
	}

	newID := atomic.AddUint32(&h.active, 1)
	h.programs[newID%uint32(len(h.programs))].New(newID, h, h.wavs.Durations())
}
