package drums

import (
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

	// if 50% of beats are active, yield to the next program
	YIELD_LIMIT = SOUNDS * BEATS / 2
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
	presses   <-chan bbox.Coord
	yield     chan<- struct{}
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
	presses <-chan bbox.Coord,
	yield chan<- struct{},
	msgs []chan<- Beats,
	tempo chan<- int,
	debug bool,
) *Keyboard {

	kb := Keyboard{
		presses: presses,
		yield:   yield,
		beats:   Beats{},
		msgs:    msgs,
		tempo:   tempo,
		emit:    make(chan Button),
		closing: make(chan struct{}),
		debug:   debug,
	}

	go kb.emitter()
	go func() {
		// starter beat
		kb.Flip(1, 0)
		kb.Flip(1, 8)
	}()

	return &kb
}

func (kb *Keyboard) Flip(beat int, tick int) {
	kb.button(beat, tick, false)
}

func (kb *Keyboard) button(beat int, tick int, decay bool) {
	kb.emit <- Button{beat: beat, tick: tick, decay: decay}
}

func (kb *Keyboard) activeButtons() int {
	active := 0

	for _, row := range kb.beats {
		for _, beat := range row {
			if beat {
				active++
			}
		}
	}

	return active
}

func (kb *Keyboard) allOff() bool {
	return kb.activeButtons() == 0
}

func (kb *Keyboard) emitter() {
	last := Button{}

	for {
		select {
		case coord, _ := <-kb.presses:
			go func() {
				kb.Flip(coord[0], coord[1])
			}()
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
			return
		case button, more := <-kb.emit:
			if !more {
				// we should never get here
				panic(1)
			}

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

				if kb.activeButtons() == YIELD_LIMIT {
					kb.yield <- struct{}{}
				}

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

		}
	}
}

func (kb *Keyboard) Close() {
	// TODO: this doesn't block?
	close(kb.closing)
}
