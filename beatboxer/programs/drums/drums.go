package drums

import (
	"time"

	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer"
)

type DrumMachine struct {
	id uint32

	kb   *Keyboard
	loop *Loop
	r    *Render

	// input
	amp      chan float64
	keyboard chan bbox.Coord
	close    chan struct{}

	output beatboxer.ProgramOutput
}

func (dm *DrumMachine) New(
	id uint32,
	output beatboxer.ProgramOutput,
	wavDurs map[string]time.Duration,
) beatboxer.Program {
	// input
	close := make(chan struct{})
	keyboard := make(chan bbox.Coord)

	// beat changes
	//   keyboard => loop
	//   keyboard => render
	msgs := []chan Beats{
		make(chan Beats),
		make(chan Beats),
	}

	// tempo changes
	//	 keyboard => loop
	tempo := make(chan int)

	// ticks
	//   loop => render
	ticks := []chan int{
		make(chan int),
	}

	// interval changes
	//   loop => render
	intervals := []chan Interval{
		make(chan Interval),
	}
	// keyboard broadcasts quit with close(msgs)
	kb := InitKeyboard(id, keyboard, output.Yield, WriteonlyBeats(msgs), tempo, false)
	loop := InitLoop(id, output.Play, msgs[0], tempo, WriteonlyInt(ticks), WriteonlyInterval(intervals))
	r := InitRender(id, msgs[1], ticks[0], intervals[0], output.Render)

	go loop.Run()
	go r.Run()

	// DrumMachine shutdown
	go func() {
		<-close

		kb.Close()
		loop.Close()
		r.Close()
	}()

	return &DrumMachine{
		id:   id,
		kb:   kb,
		loop: loop,
		r:    r,

		// input
		amp:      make(chan float64),
		keyboard: keyboard,
		close:    close,

		// output
		output: output,
	}
}

// input
func (dm *DrumMachine) Amplitude(amp float64) {

}
func (dm *DrumMachine) Keyboard(coord bbox.Coord) {
	dm.keyboard <- coord
}
func (dm *DrumMachine) Close() {

}
