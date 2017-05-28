package bbox

import (
	"fmt"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	DECAY = 2 * time.Second
)

var keymaps = map[string][]int{
	"1": []int{0, 0},
	"2": []int{0, 1},
	"3": []int{0, 2},
	"4": []int{0, 3},
	"5": []int{0, 4},
	"6": []int{0, 5},
	"7": []int{0, 6},
	"8": []int{0, 7},
	"!": []int{0, 8},
	"@": []int{0, 9},
	"#": []int{0, 10},
	"$": []int{0, 11},
	"%": []int{0, 12},
	"^": []int{0, 13},
	"&": []int{0, 14},
	"*": []int{0, 15},

	"w": []int{1, 0},
	"e": []int{1, 1},
	"r": []int{1, 2},
	"t": []int{1, 3},
	"y": []int{1, 4},
	"u": []int{1, 5},
	"i": []int{1, 6},
	"o": []int{1, 7},
	"W": []int{1, 8},
	"E": []int{1, 9},
	"R": []int{1, 10},
	"T": []int{1, 11},
	"Y": []int{1, 12},
	"U": []int{1, 13},
	"I": []int{1, 14},
	"O": []int{1, 15},

	"a": []int{2, 0},
	"s": []int{2, 1},
	"d": []int{2, 2},
	"f": []int{2, 3},
	"g": []int{2, 4},
	"h": []int{2, 5},
	"j": []int{2, 6},
	"k": []int{2, 7},
	"A": []int{2, 8},
	"S": []int{2, 9},
	"D": []int{2, 10},
	"F": []int{2, 11},
	"G": []int{2, 12},
	"H": []int{2, 13},
	"J": []int{2, 14},
	"K": []int{2, 15},

	"z": []int{3, 0},
	"x": []int{3, 1},
	"c": []int{3, 2},
	"v": []int{3, 3},
	"b": []int{3, 4},
	"n": []int{3, 5},
	"m": []int{3, 6},
	",": []int{3, 7},
	"Z": []int{3, 8},
	"X": []int{3, 9},
	"C": []int{3, 10},
	"V": []int{3, 11},
	"B": []int{3, 12},
	"N": []int{3, 13},
	"M": []int{3, 14},
	"<": []int{3, 15},
}

type Key struct {
	Ch  rune        // a unicode character
	Key termbox.Key // one of Key* constants, invalid if 'Ch' is not 0
}

type Button struct {
	beat  int
	tick  int
	reset bool
}

// mapping from keyboard box
var keymaps_rpi = map[Key][]int{
	// 2 x 21 = [volume down]
	// 2 x 24 = [mute]
	// 3 x 19 = ` (quit)

	{'1', 0}:            []int{0, 0}, // 3 x 20
	{'q', 0}:            []int{0, 1}, // 3 x 21
	{0, termbox.KeyTab}: []int{0, 2}, // 3 x 22
	{'a', 0}:            []int{0, 3}, // 3 x 23
	{'z', 0}:            []int{0, 4}, // 3 x 24
	{0, termbox.KeyF1}:  []int{0, 5}, // 4 x 19
	{'2', 0}:            []int{0, 6}, // 4 x 20
	{'w', 0}:            []int{0, 7}, // 4 x 21
	{'S', 0}:            []int{0, 8}, // 4 x 23
	// 4 x 24 = §
	{'x', 0}:           []int{0, 9},  // 4 x 25
	{0, termbox.KeyF2}: []int{0, 10}, // 5 x 19
	{'3', 0}:           []int{0, 11}, // 5 x 20
	{'e', 0}:           []int{0, 12}, // 5 x 21
	{'d', 0}:           []int{0, 13}, // 5 x 22
	{'c', 0}:           []int{0, 14}, // 5 x 23
	{0, termbox.KeyF4}: []int{0, 15}, // 5 x 24

	{'5', 0}: []int{1, 0},  // 6 x 19
	{'4', 0}: []int{1, 1},  // 6 x 20
	{'r', 0}: []int{1, 2},  // 6 x 21
	{'t', 0}: []int{1, 3},  // 6 x 22
	{'f', 0}: []int{1, 4},  // 6 x 23
	{'g', 0}: []int{1, 5},  // 6 x 24
	{'v', 0}: []int{1, 6},  // 6 x 25
	{'b', 0}: []int{1, 7},  // 6 x 26
	{'6', 0}: []int{1, 8},  // 7 x 19
	{'7', 0}: []int{1, 9},  // 7 x 20
	{'u', 0}: []int{1, 10}, // 7 x 21
	{'y', 0}: []int{1, 11}, // 7 x 22
	{'j', 0}: []int{1, 12}, // 7 x 23
	{'h', 0}: []int{1, 13}, // 7 x 24
	{'m', 0}: []int{1, 14}, // 7 x 25
	{'n', 0}: []int{1, 15}, // 7 x 26

	{'=', 0}:           []int{2, 0},  // 8 x 19
	{'8', 0}:           []int{2, 1},  // 8 x 20
	{'i', 0}:           []int{2, 2},  // 8 x 21
	{']', 0}:           []int{2, 3},  // 8 x 22
	{'K', 0}:           []int{2, 4},  // 8 x 23
	{0, termbox.KeyF6}: []int{2, 5},  // 8 x 24
	{',', 0}:           []int{2, 6},  // 8 x 25
	{0, termbox.KeyF8}: []int{2, 7},  // 9 x 19
	{'9', 0}:           []int{2, 8},  // 9 x 20
	{'o', 0}:           []int{2, 9},  // 9 x 21
	{'l', 0}:           []int{2, 10}, // 9 x 23
	{'.', 0}:           []int{2, 11}, // 9 x 25
	{'-', 0}:           []int{2, 12}, // 10 x 19
	{'0', 0}:           []int{2, 13}, // 10 x 20
	{'p', 0}:           []int{2, 14}, // 10 x 21
	{'[', 0}:           []int{2, 15}, // 10 x 22

	{';', 0}:                   []int{3, 0}, // 10 x 23
	{'\'', 0}:                  []int{3, 1}, // 10 x 24
	{'\\', 0}:                  []int{3, 2}, // 10 x 25
	{'/', 0}:                   []int{3, 3}, // 10 x 26
	{0, termbox.KeyF9}:         []int{3, 4}, // 11 x 19
	{0, termbox.KeyF10}:        []int{3, 5}, // 11 x 20
	{0, termbox.KeyBackspace2}: []int{3, 6}, // 11 x 22
	// 11 x 23 = \ ***
	{0, termbox.KeyF5}:    []int{3, 7},  // 11 x 24
	{0, termbox.KeyEnter}: []int{3, 8},  // 11 x 25
	{0, termbox.KeySpace}: []int{3, 9},  // 11 x 26
	{0, termbox.KeyF12}:   []int{3, 10}, // 12 x 20
	// 12 x 21 = 8 ***
	// 12 x 22 = 5 ***
	// 12 x 23 = 2 ***
	// 12 x 24 = 0 ***
	// 12 x 25 = / ***
	{0, termbox.KeyArrowRight}: []int{3, 11}, // 12 x 26
	{0, termbox.KeyDelete}:     []int{3, 12}, // 13 x 19
	// 13 x 20 = [fn f11]
	// 13 x 21 = 7 ***
	// 13 x 22 = 4 ***
	// 13 x 23 = 1 ***
	{0, termbox.KeyArrowDown}: []int{3, 13}, // 13 x 26
	{0, termbox.KeyPgup}:      []int{3, 14}, // 14 x 19
	{0, termbox.KeyPgdn}:      []int{3, 15}, // 14 x 20
	// 14 x 21 = 9 ***
	// 14 x 22 = 6 ***
	// 14 x 23 = 3 ***
	// 14 x 24 = . ***
	// 14 x 25 = *
	// 14 x 26 = - ***
	// 15 x 19 = KeyHome
	// 15 x 20 = KeyEnd
	// 15 x 21 = +
	// 15 x 23 = KeyEnter ***
	// 15 x 24 = KeyArrowUp
	// 15 x 25 = [brightness up]
	// 15 x 26 = KeyArrowLeft
	// 16 x 21 = [brightness down]
	// 17 x 24 = [launch itunes?]
	// 18 x 22 = [volume up]
}

// normal operation:
//   beats -> emit -> msgs
// shtudown operation:
//   '`' -> closing<- {timers.Stop(), close(msgs), close(emit)} -> termbox.Close()
type Keyboard struct {
	beats   Beats
	timers  [BEATS][TICKS]*time.Timer
	emit    chan Button
	msgs    []chan<- Beats
	closing chan struct{}
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func InitKeyboard(msgs []chan<- Beats) *Keyboard {
	// termbox.Close() called when Render.Run() exits
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputAlt)

	return &Keyboard{
		beats:   Beats{},
		msgs:    msgs,
		emit:    make(chan Button),
		closing: make(chan struct{}),
	}
}

func (kb *Keyboard) Run() {
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
				// tbprint(0, BEATS+1, termbox.ColorDefault, termbox.ColorDefault,
				// 	fmt.Sprintf("EventKey: k: %5d, c: %c", ev.Key, ev.Ch))
				// termbox.Flush()
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
