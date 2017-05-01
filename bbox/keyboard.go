package bbox

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

func RunInput() {
	var current string
	var curev termbox.Event

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputAlt)

	data := make([]byte, 0, 64)
mainloop:
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
			current = fmt.Sprintf("%q", data)
			if current == `"q"` {
				panic(0)
				break mainloop
			}

			fmt.Println(data)
			fmt.Println(current)
			fmt.Println(curev)

			for {
				ev := termbox.ParseEvent(data)
				fmt.Printf("  data: %+v\n", data)
				fmt.Printf("  ev: %+v\n", ev)

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
