package bbox

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

// mapping from keyboard box
// 2 x 21 = [volume down]
// 2 x 24 = [mute]
// 3 x 19 = `
// 3 x 20 = 1
// 3 x 21 = q
// 3 x 22 = KeyTab
// 3 x 23 = a
// 3 x 24 = z
// 4 x 19 = KeyF1
// 4 x 20 = 2
// 4 x 21 = w
// 4 x 23 = S
// 4 x 24 = §
// 4 x 25 = x
// 5 x 19 = KeyF2
// 5 x 20 = 3
// 5 x 21 = e
// 5 x 22 = d
// 5 x 23 = c
// 5 x 24 = KeyF4
// 6 x 19 = 5
// 6 x 20 = 4
// 6 x 21 = r
// 6 x 22 = t
// 6 x 23 = f
// 6 x 24 = g
// 6 x 25 = v
// 6 x 26 = b
// 7 x 19 = 6
// 7 x 20 = 7
// 7 x 21 = u
// 7 x 22 = y
// 7 x 23 = j
// 7 x 24 = h
// 7 x 25 = m
// 7 x 26 = n
// 8 x 19 = =
// 8 x 20 = 8
// 8 x 21 = i
// 8 x 22 = ]
// 8 x 23 = K
// 8 x 24 = KeyF6
// 8 x 25 = ,
// 9 x 19 = KeyF8
// 9 x 20 = 9
// 9 x 21 = o
// 9 x 23 = l
// 9 x 25 = .
// 10 x 19 = -
// 10 x 20 = 0
// 10 x 21 = p
// 10 x 22 = [
// 10 x 23 = ;
// 10 x 24 = '
// 10 x 25 = \
// 10 x 26 = /
// 11 x 19 = KeyF9
// 11 x 20 = KeyF10
// 11 x 22 = KeyBackspace2
// 11 x 23 = \ ***
// 11 x 24 = KeyF5
// 11 x 25 = KeyEnter
// 11 x 26 = KeySpace
// 12 x 20 = KeyF12
// 12 x 21 = 8 ***
// 12 x 22 = 5 ***
// 12 x 23 = 2 ***
// 12 x 24 = 0 ***
// 12 x 25 = / ***
// 12 x 26 = KeyArrowRight
// 13 x 19 = KeyDelete
// 13 x 20 = [fn f11]
// 13 x 21 = 7 ***
// 13 x 22 = 4 ***
// 13 x 23 = 1 ***
// 13 x 26 = KeyArrowDown
// 14 x 19 = KeyPgup
// 14 x 20 = KeyPgdn
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

var keymaps_rpi = map[string][]int{
	"1": []int{0, 0},  // 3 x 20
	"q": []int{0, 1},  // 3 x 21
	"a": []int{0, 2},  // 3 x 23
	"z": []int{0, 4},  // 3 x 24
	"2": []int{0, 5},  // 4 x 20
	"w": []int{0, 6},  // 4 x 21
	"S": []int{0, 7},  // 4 x 23
	"§": []int{0, 8},  // 4 x 24
	"x": []int{0, 9},  // 4 x 25
	"3": []int{0, 10}, // 5 x 20
	"e": []int{0, 11}, // 5 x 21
	"d": []int{0, 12}, // 5 x 22
	"c": []int{0, 13}, // 5 x 23
	"5": []int{0, 14}, // 6 x 19
	"4": []int{0, 15}, // 6 x 20

	"r": []int{1, 0},  // 6 x 21
	"t": []int{1, 1},  // 6 x 22
	"f": []int{1, 2},  // 6 x 23
	"g": []int{1, 3},  // 6 x 24
	"v": []int{1, 4},  // 6 x 25
	"b": []int{1, 5},  // 6 x 26
	"6": []int{1, 6},  // 7 x 19
	"7": []int{1, 7},  // 7 x 20
	"u": []int{1, 8},  // 7 x 21
	"y": []int{1, 9},  // 7 x 22
	"j": []int{1, 10}, // 7 x 23
	"h": []int{1, 11}, // 7 x 24
	"m": []int{1, 12}, // 7 x 25
	"n": []int{1, 13}, // 7 x 26
	"=": []int{1, 14}, // 8 x 19
	"8": []int{1, 15}, // 8 x 20

	"i":  []int{2, 0},  // 8 x 21
	"]":  []int{2, 1},  // 8 x 22
	"K":  []int{2, 2},  // 8 x 23
	",":  []int{2, 3},  // 8 x 25
	"9":  []int{2, 4},  // 9 x 20
	"o":  []int{2, 5},  // 9 x 21
	"l":  []int{2, 6},  // 9 x 23
	".":  []int{2, 7},  // 9 x 23
	"-":  []int{2, 8},  // 10 x 19
	"0":  []int{2, 9},  // 10 x 20
	"p":  []int{2, 10}, // 10 x 21
	"[":  []int{2, 11}, // 10 x 22
	";":  []int{2, 12}, // 10 x 23
	"'":  []int{2, 13}, // 10 x 24
	"\\": []int{2, 14}, // 10 x 25
	"/":  []int{2, 15}, // 10 x 26

	// ".": []int{3, 8}, // 14 x 24

	"*": []int{3, 9}, // 14 x 25

	// "-": []int{3, 10}, // 14 x 26

	"+": []int{3, 13}, // 15 x 21
}

var keymaps_rpi_keys = map[termbox.Key][]int{
	termbox.KeyTab: []int{0, 3}, // 3 x 22

	termbox.KeyBackspace:  []int{3, 0},  // 11 x 22
	termbox.KeyEnter:      []int{3, 1},  // 11 x 25
	termbox.KeySpace:      []int{3, 2},  // 11 x 26
	termbox.KeyArrowRight: []int{3, 3},  // 12 x 26
	termbox.KeyDelete:     []int{3, 4},  // 13 x 19
	termbox.KeyArrowDown:  []int{3, 5},  // 13 x 26
	termbox.KeyPgup:       []int{3, 6},  // 14 x 19
	termbox.KeyPgdn:       []int{3, 7},  // 14 x 20
	termbox.KeyHome:       []int{3, 11}, // 15 x 19
	termbox.KeyEnd:        []int{3, 12}, // 15 x 20
	termbox.KeyArrowUp:    []int{3, 14}, // 15 x 24
	termbox.KeyF2:         []int{3, 15}, // 15 x 25
}

// normal operation:
//   beats -> emit -> msgs
// shtudown operation:
//   q -> close(emit) -> close(msgs) -> termbox.Close()
type Keyboard struct {
	beats Beats
	emit  chan Beats
	msgs  []chan<- Beats
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
		beats: Beats{},
		emit:  make(chan Beats),
		msgs:  msgs,
	}
}

func (kb *Keyboard) Run() {
	var current string
	var curev termbox.Event

	defer close(kb.emit)

	data := make([]byte, 0, 64)

	// starter beat
	go kb.Emitter()
	kb.beats[1][0] = true
	kb.beats[1][8] = true
	kb.Emit()

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
				kb.beats[key[0]][key[1]] = !kb.beats[key[0]][key[1]]
				kb.Emit()
			}

			for {
				// TODO: move kb.beats code to here
				ev := termbox.ParseEvent(data)
				if ev.N == 0 {
					break
				}
				curev = ev
				copy(data, data[curev.N:])
				data = data[:len(data)-curev.N]

				tbprint(0, BEATS+1, termbox.ColorDefault, termbox.ColorDefault,
					fmt.Sprintf("EventKey: k: %5d, c: %c", ev.Key, ev.Ch))
				termbox.Flush()
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func (kb *Keyboard) Emit() {
	beats := kb.beats
	kb.emit <- beats
}

func (kb *Keyboard) Emitter() {
	for {
		select {
		case beats, more := <-kb.emit:
			if more {
				for _, msg := range kb.msgs {
					msg <- beats
				}
			} else {
				for _, msg := range kb.msgs {
					close(msg)
				}
				return
			}
		}
	}
}
