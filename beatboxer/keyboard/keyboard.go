package keyboard

import (
	"fmt"

	"github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox"
)

type tbcell struct {
	x int
	y int
	termbox.Cell
}

// normal operation:
//   keyboard -> emit
type Keyboard struct {
	keyMap  map[bbox.Key]*bbox.Coord
	pressed chan bbox.Coord // single button press
	cell    chan tbcell
	flush   chan struct{}
}

func Init(
	// pressed chan bbox.Coord,
	keyMap map[bbox.Key]*bbox.Coord,
) *Keyboard {
	// termbox.Close() called when Render.Run() exits
	// err := termbox.Init()
	// if err != nil {
	// 	panic(err)
	// }
	// termbox.SetInputMode(termbox.InputAlt)

	return &Keyboard{
		keyMap:  keyMap,
		pressed: make(chan bbox.Coord),
		cell:    make(chan tbcell),
		flush:   make(chan struct{}),
	}
}

func (kb *Keyboard) Run() {

	outputDone := make(chan struct{})

	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	// this defer kicks off once the input returns
	defer func(outputDone chan<- struct{}) {
		outputDone <- struct{}{}
		close(kb.pressed)
		close(kb.cell)
		close(kb.flush)
		// this only executes after both input and output are done
		termbox.Close()
	}(outputDone)

	termbox.SetInputMode(termbox.InputAlt)

	// output
	go func(outputDone <-chan struct{}) {
		// defer func() {
		// 	outputDone <- struct{}{}
		// }()
		for {
			select {
			case cell, more := <-kb.cell:
				if !more {
					fmt.Printf("cell channel closed, closing\n")
					return
				}
				termbox.SetCell(cell.x, cell.y, cell.Ch, cell.Fg, cell.Bg)
			case _, more := <-kb.flush:
				if !more {
					fmt.Printf("flush channel closed, closing\n")
					return
				}
				termbox.Flush()
			case _, more := <-outputDone:
				if !more {
					fmt.Printf("outputDone channel closed, closing\n")
					return
				}
			}
		}
	}(outputDone)

	var curev termbox.Event
	data := make([]byte, 0, 64)

	// input
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
			if fmt.Sprintf("%s", data) == "`" {
				fmt.Printf("Keyboard received backtick, closing\n")
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

				key := kb.keyMap[bbox.Key{
					Ch:  ev.Ch,
					Key: ev.Key,
				}]

				if key != nil {
					kb.pressed <- *key
				}
			}
		case termbox.EventInterrupt:
			fmt.Printf("termbox.EventInterrupt received, closing\n")
			return
		case termbox.EventError:
			fmt.Printf("termbox.EventError received, closing: %+v\n", ev.Err)
			return
		}
	}
}

func (kb *Keyboard) Pressed() <-chan bbox.Coord {
	return kb.pressed
}

func (kb *Keyboard) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	kb.cell <- tbcell{
		x: x,
		y: y,
		Cell: termbox.Cell{
			Ch: ch,
			Fg: fg,
			Bg: bg,
		},
	}
}

func (kb *Keyboard) Flush() {
	kb.flush <- struct{}{}
}
