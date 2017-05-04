package bbox

import (
	"fmt"

	"github.com/nsf/termbox-go"
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

type Keyboard struct {
	beats Beats
	msgs  chan<- Beats
}

func InitKeyboard(msgs chan<- Beats) *Keyboard {
	beats := Beats{}

	// starter beat
	beats[0][0] = true
	beats[0][8] = true
	go func() { msgs <- beats }()

	return &Keyboard{
		beats: beats,
		msgs:  msgs,
	}
}

func (kb *Keyboard) Draw() {
	for i := 0; i < len(kb.beats); i++ {
		for j := 0; j < TICKS; j++ {
			c := '-'
			if kb.beats[i][j] {
				c = 'X'
			}
			termbox.SetCell(j*2, i, c, termbox.ColorDefault, termbox.ColorDefault)
			termbox.SetCell(j*2+1, i, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	termbox.Flush()
}

func (kb *Keyboard) Run() {
	var current string
	var curev termbox.Event

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer func() {
		termbox.Close()
		close(kb.msgs)
	}()
	termbox.SetInputMode(termbox.InputAlt)

	kb.Draw()

	data := make([]byte, 0, 64)

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
			if current == "q" {
				// trigger the deferred termbox.Close and quit<-
				return
			}

			key := keymaps[current]
			if key != nil {
				kb.beats[key[0]][key[1]] = !kb.beats[key[0]][key[1]]
				beats := kb.beats // make a copy before going asynch
				go func() { kb.msgs <- beats }()
			}
			kb.Draw()

			for {
				ev := termbox.ParseEvent(data)
				// fmt.Printf("  data: %+v\n", data)
				// fmt.Printf("  ev: %+v\n", ev)

				if ev.N == 0 {
					break
				}
				curev = ev
				copy(data, data[curev.N:])
				data = data[:len(data)-curev.N]
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}
