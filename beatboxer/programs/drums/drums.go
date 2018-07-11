package drums

import (
	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer"
)

type DrumMachine struct {
	kb     *Keyboard
	loop   *Loop
	r      *Render
	output beatboxer.Output
}

func (dm *DrumMachine) New(output beatboxer.Output) beatboxer.Program {
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
	kb := InitKeyboard(output.Yield, WriteonlyBeats(msgs), tempo, bbox.KeyMapsPC, false)
	loop := InitLoop(output.Play, msgs[0], tempo, WriteonlyInt(ticks), WriteonlyInterval(intervals))
	r := InitRender(msgs[1], ticks[0], intervals[0], output.Render)

	go loop.Run()
	go r.Run()

	return &DrumMachine{
		kb:   kb,
		loop: loop,
		r:    r,
	}
}

func (dm *DrumMachine) Amp(level float64) {}

func (dm *DrumMachine) Pressed(row int, column int) {
	dm.kb.Flip(row, column)
}

func (dm *DrumMachine) Close() {
	dm.kb.Close()
	dm.loop.Close()
	dm.r.Close()
}
