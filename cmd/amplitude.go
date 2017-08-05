package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/siggy/bbox/bbox"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	var wg sync.WaitGroup

	level := make(chan float64)

	amplitude := bbox.InitAmplitude(&wg, level)

	go amplitude.Run()

	for {
		select {
		case i, more := <-level:
			if more {
				fmt.Printf("\r%s", strings.Repeat(" ", 100))
				fmt.Printf("\r%s", strings.Repeat("#", int(100*i)))
			} else {
				return
			}
		case <-sig:
			return
		default:
		}
	}

	wg.Wait()
}
