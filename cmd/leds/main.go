package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"

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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	strips, err := leds.New(stripLengths)
	if err != nil {
		log.Errorf("leds.New failed: %+v", err)
		os.Exit(1)
	}

	defer strips.Close()

	strips.Clear()

	prev := 0
	cur := 0
	for {
		select {
		case <-sig:
			return
		default:
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter LED: ")
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)

			if text == ";" {
				cur = int(math.Abs(float64(cur - 1)))
			} else if text == "'" {
				cur = cur + 1
			} else {
				cur, err = strconv.Atoi(text)
				if err != nil {
					fmt.Printf("strconv.Atoi failed: %+v\n", err)
					continue
				}
			}

			// strips.Clear()

			fmt.Printf("LED: %+v\n", cur)
			for strip := range stripLengths {
				err := strips.Set(strip, prev, leds.Color{R: 0, G: 0, B: 0, W: 0})
				if err != nil {
					log.Errorf("strips.Set failed: %+v\n", err)
				}
				err = strips.Set(strip, cur, leds.Color{R: 255, G: 0, B: 0, W: 0})
				if err != nil {
					log.Errorf("strips.Set failed: %+v\n", err)
				}
			}

			err := strips.Write()
			if err != nil {
				log.Errorf("strips.Write failed: %+v\n", err)
			}

			prev = cur
		}
	}
}
