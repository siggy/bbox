package beatboxer

import (
	"fmt"

	"github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox"
)

// normal operation:
//   keyboard -> emit
type Keyboard struct {
	keyMap  map[bbox.Key]*bbox.Coord
	pressed chan<- bbox.Coord // single button press
}

func tbprint(x, y int, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}

func InitKeyboard(
	pressed chan<- bbox.Coord,
	keyMap map[bbox.Key]*bbox.Coord,
) *Keyboard {
	// termbox.Close() called when Render.Run() exits
	// err := termbox.Init()
	// if err != nil {
	// 	panic(err)
	// }
	termbox.SetInputMode(termbox.InputAlt)

	return &Keyboard{
		keyMap:  keyMap,
		pressed: pressed,
	}
}

func (kb *Keyboard) Run() {
	var current string
	var curev termbox.Event
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
			if current == "`" {
				// triggers a deferred kb.closing, which closes the emitter
				fmt.Printf("Keyboard received backtick, closing\n")
				close(kb.pressed)
				return
			}

			for {
				ev := termbox.ParseEvent(data)
				if ev.N == 0 {
					break
				}
				curev = ev
				copy(data, data[curev.N:])
				data = data[:len(data)-curev.N]

				key := kb.keyMap[bbox.Key{ev.Ch, ev.Key}]
				if key != nil {
					kb.pressed <- *key
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}
