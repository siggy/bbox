package beatboxer

import (
	"errors"
	"sync"

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

	for _, renderer := range h.renderers {
		renderer.Render(render.State{})
	}

	active := 0
	cur := h.programs[active].New(h.wavs.Durations())

	for {
		err := h.runProgram(cur)
		go func(cur Program) {
			cur.Close() <- struct{}{}
		}(cur)
		if err != nil {
			break
		}

		h.wavs.StopAll()
		for _, renderer := range h.renderers {
			renderer.Render(render.State{})
		}

		active = (active + 1) % len(h.programs)
		cur = h.programs[active].New(h.wavs.Durations())
	}
}

func (h *Harness) runProgram(p Program) error {
	closing := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(5)

	go h.runAmp(p, &wg, closing)
	go h.runRender(p, &wg, closing)
	go h.runPlay(p, &wg, closing)
	go h.runYield(p, &wg, closing)
	err := h.runKB(p, &wg, closing)

	wg.Wait()

	return err
}

// input: amplitude
func (h *Harness) runAmp(p Program, wg *sync.WaitGroup, closing chan struct{}) {
	defer wg.Done()

	for {
		select {
		case p.Amplitude() <- <-h.amplitude.Level():
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

			p.Keyboard() <- coord
		case <-h.kb.Closing():
			close(closing)
			return errors.New("Exiting")
		case <-closing:
			return nil
		}
	}
}

// output: render
func (h *Harness) runRender(p Program, wg *sync.WaitGroup, closing chan struct{}) {
	defer wg.Done()

	for {
		select {
		case rs, _ := <-p.Render():
			for _, renderer := range h.renderers {
				renderer.Render(rs)
			}
		case <-closing:
			return
		}
	}
}

// output: play
func (h *Harness) runPlay(p Program, wg *sync.WaitGroup, closing chan struct{}) {
	defer wg.Done()

	for {
		select {
		case name, _ := <-p.Play():
			h.wavs.Play(name)
		case <-closing:
			return
		}
	}
}

// output: yield
func (h *Harness) runYield(p Program, wg *sync.WaitGroup, closing chan struct{}) {
	defer wg.Done()

	for {
		select {
		case <-p.Yield():
			close(closing)
			return
		case <-closing:
			return
		}
	}
}
