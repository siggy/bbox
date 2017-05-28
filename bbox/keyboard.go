package bbox

import (
	"fmt"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	DECAY = 2 * time.Second
)

type Button struct {
	beat  int
	tick  int
	reset bool
}

// normal operation:
//   beats -> emit -> msgs
// shtudown operation:
//   '`' -> closing<- {timers.Stop(), close(msgs), close(emit)} -> termbox.Close()
type Keyboard struct {
	wg      *sync.WaitGroup
	beats   Beats
	timers  [BEATS][TICKS]*time.Timer
	emit    chan Button
	msgs    []chan<- Beats
	closing chan struct{}
	debug   bool
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func InitKeyboard(wg *sync.WaitGroup, msgs []chan<- Beats, debug bool) *Keyboard {
	// termbox.Close() called when Render.Run() exits
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputAlt)

	wg.Add(1)

	return &Keyboard{
		wg:      wg,
		beats:   Beats{},
		msgs:    msgs,
		emit:    make(chan Button),
		closing: make(chan struct{}),
		debug:   debug,
	}
}

func (kb *Keyboard) Run() {
	defer kb.wg.Done()

	var current string
	var curev termbox.Event

	defer func() { kb.closing <- struct{}{} }()

	data := make([]byte, 0, 64)

	// starter beat
	go kb.emitter()
	kb.flip(1, 0)
	kb.flip(1, 8)

	for {
		if cap(data)-len(data) < 32 {
			newdata := make([]byte, len(data), len(data)+32)
			copy(newdata, data)
			data = newdata
		}
		beg := len(data)
		d := data[beg : beg+32]
		switch ev := termbox.PollRawEvent(d); ev.Type {
		case termbox.EventRaw:
			data = data[:beg+ev.N]
			current = fmt.Sprintf("%s", data)
			if current == "`" {
				return
			}

			key := keymaps[current]
			if key != nil {
				kb.flip(key[0], key[1])
			}

			for {
				// TODO: move kb.flip code to here
				ev := termbox.ParseEvent(data)
				if ev.N == 0 {
					break
				}
				curev = ev
				copy(data, data[curev.N:])
				data = data[:len(data)-curev.N]

				// for debugging output
				if kb.debug {
					tbprint(0, BEATS+1, termbox.ColorDefault, termbox.ColorDefault,
						fmt.Sprintf("EventKey: k: %5d, c: %c", ev.Key, ev.Ch))
					termbox.Flush()
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func (kb *Keyboard) flip(beat int, tick int) {
	kb.button(beat, tick, false)
}

func (kb *Keyboard) button(beat int, tick int, reset bool) {
	kb.emit <- Button{beat: beat, tick: tick, reset: reset}
}

func (kb *Keyboard) emitter() {
	for {
		select {
		case <-kb.closing:
			// ensure all timers are stopped before closing kb.emit
			for _, arr := range kb.timers {
				for _, t := range arr {
					if t != nil {
						t.Stop()
					}
				}
			}
			for _, msg := range kb.msgs {
				close(msg)
			}
			close(kb.emit)
			return
		case button, more := <-kb.emit:
			if more {
				// todo: consider re-using timers
				if kb.timers[button.beat][button.tick] != nil {
					kb.timers[button.beat][button.tick].Stop()
				}

				if button.reset {
					kb.beats[button.beat][button.tick] = false
				} else {
					kb.beats[button.beat][button.tick] = !kb.beats[button.beat][button.tick]

					// if we're turning this button on, set a decay timer
					if kb.beats[button.beat][button.tick] {
						kb.timers[button.beat][button.tick] = time.AfterFunc(DECAY, func() {
							kb.button(button.beat, button.tick, true)
						})
					}
				}

				// broadcast changes
				for _, msg := range kb.msgs {
					msg <- kb.beats
				}
			} else {
				// we should never get here
				fmt.Printf("closed on emit, invalid state")
				panic(1)
			}
		}
	}
}
