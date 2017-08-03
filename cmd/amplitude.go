package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/siggy/bbox/bbox"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	level := make(chan float64)

	amplitude := bbox.InitAmplitude(level)

	go amplitude.Run()
	defer amplitude.Close()

	for {
		select {
		case i, more := <-level:
			if more {
				// fmt.Printf("\r%s", strings.Repeat(" ", 100))
				// fmt.Printf("\r%s", strings.Repeat("#", int(100*i)))
				fmt.Printf("%s\n", strings.Repeat("#", int(100*i)))
			} else {
				return
			}
		case <-sig:
			return
		default:
		}
	}
}
