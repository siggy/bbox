package keyboard

import (
	"fmt"
	"sync"

	"github.com/nsf/termbox-go"
	"github.com/siggy/bbox/bbox"
	log "github.com/sirupsen/logrus"
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
	// cell    chan tbcell
	// flush   chan struct{}
	closing chan struct{}
	wg      sync.WaitGroup
}

func Init(
	// pressed chan bbox.Coord,
	keyMap map[bbox.Key]*bbox.Coord,
) *Keyboard {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputAlt)

	return &Keyboard{
		keyMap:  keyMap,
		pressed: make(chan bbox.Coord),
		// cell:    make(chan tbcell),
		// flush:   make(chan struct{}),
		closing: make(chan struct{}),
		wg:      sync.WaitGroup{},
	}
}

func (kb *Keyboard) Run() {
	// err := termbox.Init()
	// if err != nil {
	// 	panic(err)
	// }

	// // defer termbox.Close()

	// termbox.SetInputMode(termbox.InputAlt)

	// wg := sync.WaitGroup{}

	kb.input()
	// kb.output()
}

func (kb *Keyboard) Pressed() <-chan bbox.Coord {
	return kb.pressed
}

// func (kb *Keyboard) SetCell(x, y int, ch rune, fg, bg termbox.Attribute) {
// 	kb.cell <- tbcell{
// 		x: x,
// 		y: y,
// 		Cell: termbox.Cell{
// 			Ch: ch,
// 			Fg: fg,
// 			Bg: bg,
// 		},
// 	}
// }

func (kb *Keyboard) Closing() <-chan struct{} {
	return kb.closing
}

// func (kb *Keyboard) Flush() {
// 	kb.flush <- struct{}{}
// }

func (kb *Keyboard) Close() {
	log.Debugf("kb.Close: close(kb.pressed)")
	close(kb.pressed)

	// log.Debugf("kb.Close: close(kb.cell)")
	// close(kb.cell)

	// log.Debugf("kb.Close: close(kb.flush)")
	// close(kb.flush)

	// waits for input and output to finish
	log.Debugf("kb.Close: kb.wg.Wait()")
	kb.wg.Wait()

	log.Debugf("kb.Close: termbox.Close()")
	termbox.Close()
}

func (kb *Keyboard) input() {
	kb.wg.Add(1)
	defer kb.wg.Done()

	defer close(kb.closing)

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
			if fmt.Sprintf("%s", data) == "`" {
				log.Debugf("Keyboard received backtick, closing")
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
			log.Debugf("termbox.EventInterrupt received, closing")
			return
		case termbox.EventError:
			log.Debugf("termbox.EventError received, closing: %+v", ev.Err)
			return
		}
	}
}

// func (kb *Keyboard) output() {
// 	kb.wg.Add(1)
// 	defer kb.wg.Done()

// 	for {
// 		select {
// 		case cell, more := <-kb.cell:
// 			if !more {
// 				log.Debugf("output: cell channel closed, closing")
// 				return
// 			}
// 			termbox.SetCell(cell.x, cell.y, cell.Ch, cell.Fg, cell.Bg)
// 			// case _, more := <-kb.flush:
// 			// 	if !more {
// 			// 		log.Debugf("output: flush channel closed, closing")
// 			// 		return
// 			// 	}
// 			// 	termbox.Flush()
// 		}
// 	}
// }
