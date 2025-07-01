package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/siggy/bbox/bbox2/leds"
	log "github.com/sirupsen/logrus"
)

var (
	// should match scorpio/code.py
	// StripLengths = []int{30, 30, 10, 10, 10, 10, 10, 10}
	stripLengths = []int{30}
)

func main() {
	logLevel := flag.String("log-level", "debug", "set log level (debug, info, warn, error, fatal, panic)")
	flag.Parse()

	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	log.SetLevel(lvl)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ledStrips, err := leds.New(ctx, stripLengths, false)
	if err != nil {
		log.Errorf("leds.New failed: %+v", err)
		os.Exit(1)
	}

	defer ledStrips.Close()

	ledStrips.Clear()

	prev := 0
	cur := 0
	for {
		select {
		case <-ctx.Done():
			log.Info("context done")
			return
		default:
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter LED: ")
			text, _ := reader.ReadString('\n')
			text = strings.ReplaceAll(text, "\n", "")

			switch text {
			case ";":
				cur = int(math.Abs(float64(cur - 1)))
			case "'":
				cur = cur + 1
			default:
				cur, err = strconv.Atoi(text)
				if err != nil {
					fmt.Printf("strconv.Atoi failed: %+v\n", err)
					continue
				}
			}

			fmt.Printf("LED: %+v\n", cur)
			state := leds.State{}
			for strip := range stripLengths {
				state.Set(strip, prev, leds.Color{R: 0, G: 0, B: 0, W: 0})
				state.Set(strip, cur, leds.Color{R: 255, G: 0, B: 0, W: 0})
			}

			ledStrips.Set(state)
			if err != nil {
				log.Errorf("ledStrips.Write failed: %+v\n", err)
			}

			prev = cur
		}
	}
}
