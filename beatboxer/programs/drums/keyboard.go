package drums

import (
	"fmt"
	"time"

	"github.com/siggy/bbox/bbox"
)

const (
	TEMPO_TICK = 15

	// test
	// DECAY      = 2 * time.Second
	// KEEP_ALIVE = 5 * time.Second

	// prod
	DECAY      = 3 * time.Minute
	KEEP_ALIVE = 14 * time.Minute
)

type Button struct {
	beat  int
	tick  int
	decay bool
}

// normal operation:
//   beats -> emit -> msgs
// shtudown operation:
//   '`' -> closing<- {timers.Stop(), close(msgs), close(emit)} -> termbox.Close()
type Keyboard struct {
	keyMap    map[bbox.Key]*bbox.Coord
	beats     Beats
	timers    [SOUNDS][BEATS]*time.Timer
	keepAlive *time.Timer    // ensure at least one beat is sent periodically to keep speaker alive
	emit      chan Button    // single button press, keyboard->emitter
	msgs      []chan<- Beats // all beats, emitter->msgs
	tempo     chan<- int     // tempo changes
	closing   chan struct{}
	debug     bool
}

func InitKeyboard(
	msgs []chan<- Beats,
	tempo chan<- int,
	keyMap map[bbox.Key]*bbox.Coord,
	debug bool,
) *Keyboard {

	fmt.Printf("InitKeyboard00\n")

	kb := Keyboard{
		keyMap:  keyMap,
		beats:   Beats{},
		msgs:    msgs,
		tempo:   tempo,
		emit:    make(chan Button),
		closing: make(chan struct{}),
		debug:   debug,
	}

	go func() {
		kb.emitter()
		// starter beat
		fmt.Printf("InitKeyboard0\n")
		kb.Flip(1, 0)
		fmt.Printf("InitKeyboard1\n")
		kb.Flip(1, 8)
		fmt.Printf("InitKeyboard2\n")
	}()

	return &kb
}

func (kb *Keyboard) Flip(beat int, tick int) {
	kb.button(beat, tick, false)
}

func (kb *Keyboard) button(beat int, tick int, decay bool) {
	kb.emit <- Button{beat: beat, tick: tick, decay: decay}
}

func (kb *Keyboard) allOff() bool {
	for _, row := range kb.beats {
		for _, beat := range row {
			if beat {
				return false
			}
		}
	}

	return true
}

func (kb *Keyboard) emitter() {
	last := Button{}

	for {
		select {
		case <-kb.closing:
			// ensure all timers are stopped before closing kb.emit
			if kb.keepAlive != nil {
				kb.keepAlive.Stop()
			}

			for _, arr := range kb.timers {
				for _, t := range arr {
					if t != nil {
						t.Stop()
					}
				}
			}
			for _, msg := range kb.msgs {
				close(msg) // signals to other processes to quit
			}
			close(kb.emit)
			fmt.Printf("Keyboard emitter closing\n")
			return
		case button, more := <-kb.emit:
			if more {
				// TODO: consider re-using timers
				if kb.timers[button.beat][button.tick] != nil {
					kb.timers[button.beat][button.tick].Stop()
				}

				if button.decay || kb.beats[button.beat][button.tick] {
					// turning off
					kb.beats[button.beat][button.tick] = false

					// check for all beats off, if so, set a keepalive timer
					if kb.allOff() {
						kb.keepAlive = time.AfterFunc(KEEP_ALIVE, func() {
							// enable a beat to keep the speaker alive
							kb.Flip(1, 0)
						})
					}
				} else {
					// turning on
					kb.beats[button.beat][button.tick] = true

					// set a decay timer
					kb.timers[button.beat][button.tick] = time.AfterFunc(DECAY, func() {
						kb.button(button.beat, button.tick, true)
					})

					// we've enabled a beat, kill keepAlive
					if kb.keepAlive != nil {
						kb.keepAlive.Stop()
					}
				}

				// broadcast changes
				for _, msg := range kb.msgs {
					msg <- kb.beats
				}

				// if tempo change, broadcast
				if button.tick == TEMPO_TICK && last.tick == TEMPO_TICK {
					if button.beat == 0 && last.beat == 0 {
						kb.tempo <- 4
					} else if button.beat == 3 && last.beat == 3 {
						kb.tempo <- -4
					}
				}

				last = button
			} else {
				// we should never get here
				fmt.Printf("closed on emit, invalid state")
				panic(1)
			}
		}
	}
}

func (kb *Keyboard) Close() {
	// TODO: this doesn't block?
	close(kb.closing)
}
