package drums

import (
	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer"
	"github.com/siggy/bbox/beatboxer/render"
)

type DrumMachine struct {
	kb   *Keyboard
	loop *Loop
	r    *Render

	amp      chan float64
	keyboard chan bbox.Coord
	close    chan struct{}

	// output
	play   chan string
	render chan render.RenderState
	yield  chan struct{}
}

func (dm *DrumMachine) New() beatboxer.Program {
	// input
	keyboard := make(chan bbox.Coord)

	// output
	play := make(chan string)
	render := make(chan render.RenderState)
	yield := make(chan struct{})

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
	kb := InitKeyboard(keyboard, yield, WriteonlyBeats(msgs), tempo, bbox.KeyMapsPC, false)
	loop := InitLoop(play, msgs[0], tempo, WriteonlyInt(ticks), WriteonlyInterval(intervals))
	r := InitRender(msgs[1], ticks[0], intervals[0], render)

	go loop.Run()
	go r.Run()

	return &DrumMachine{
		kb:   kb,
		loop: loop,
		r:    r,

		// input
		amp:      make(chan float64),
		keyboard: keyboard,
		close:    make(chan struct{}),

		// output
		play:   play,
		render: render,
		yield:  yield,
	}
}

// input
func (dm *DrumMachine) Amplitude() chan<- float64 {
	return dm.amp
}
func (dm *DrumMachine) Keyboard() chan<- bbox.Coord {
	return dm.keyboard
}
func (dm *DrumMachine) Close() chan<- struct{} {
	return dm.close
}

// output
func (dm *DrumMachine) Play() <-chan string {
	return dm.play
}
func (dm *DrumMachine) Render() <-chan render.RenderState {
	return dm.render
}
func (dm *DrumMachine) Yield() <-chan struct{} {
	return dm.yield
}

// func (dm *DrumMachine) Amp(level float64) {}

// func (dm *DrumMachine) Pressed(row int, col int) {
// 	log.Debugf("dm.Pressed start: %02d, %02d", row, col)
// 	dm.kb.Flip(row, col)
// 	log.Debugf("dm.Pressed end: %02d, %02d", row, col)
// }

// func (dm *DrumMachine) Close() {
// 	dm.kb.Close()
// 	dm.loop.Close()
// 	dm.r.Close()
// }
