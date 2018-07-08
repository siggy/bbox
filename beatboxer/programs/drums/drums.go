package drums

import (
	"github.com/siggy/bbox/bbox"
	"github.com/siggy/bbox/beatboxer/render"
)

type DrumMachine struct {
	kb   *Keyboard
	loop *Loop
	r    *Render
}

func (dm *DrumMachine) Init(renderCB func(render.RenderState)) {
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
	dm.kb = InitKeyboard(WriteonlyBeats(msgs), tempo, bbox.KeyMapsPC, false)
	dm.loop = InitLoop(msgs[0], tempo, WriteonlyInt(ticks), WriteonlyInterval(intervals))
	dm.r = InitRender(msgs[1], ticks[0], intervals[0], renderCB)

	go dm.loop.Run()
	go dm.r.Run()
}

func (dm *DrumMachine) Pressed(row int, column int) {
	dm.kb.Flip(row, column)
}

func (dm *DrumMachine) Close() {
	dm.kb.Close()
	dm.loop.Close()
	dm.r.Close()
}
