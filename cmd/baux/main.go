package main

import (
	"context"
	"flag"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/siggy/bbox/bbox2/amplitude"
	"github.com/siggy/bbox/bbox2/leds"
	log "github.com/sirupsen/logrus"
)

const (
	baseLEDStrip   = 0
	globalLEDStrip = 1

	baseLEDCount  = 85
	globeLEDCount = 240

	bauxStreakLength = baseLEDCount * 3 / 4

	fps               = 30
	defaultIntervalMS = 2000
)

var (
	// should match scorpio/code.py
	stripLengths = []int{baseLEDCount, globeLEDCount}

	globeLeds = []int{
		0,
		29,
		55,
		79,
		102,
		124,
		145,
		165,
		183,
		200,
		216,
		230,
	}
)

func main() {
	logLevel := flag.String("log-level", "debug", "set log level (debug, info, warn, error, fatal, panic)")
	fakeLEDs := flag.Bool("fake-leds", false, "enable fake LEDs")
	macDevice := flag.Bool("mac-device", false, "connect to scorpio from a macbook (default is Raspberry Pi)")
	flag.Parse()

	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	log.SetLevel(lvl)
	log := log.WithField("baux", "main")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// init
	var ledStrips leds.LEDs
	if *fakeLEDs {
		ledStrips, err = leds.NewFake(stripLengths)
		if err != nil {
			log.Errorf("leds.NewFake failed: %+v", err)
			os.Exit(1)
		}
	} else {
		ledStrips, err = leds.New(ctx, stripLengths, *macDevice)
		if err != nil {
			log.Errorf("leds.New failed: %+v", err)
			os.Exit(1)
		}
	}
	defer ledStrips.Close()
	ledStrips.Clear()

	amp, err := amplitude.New()
	if err != nil {
		log.Fatalf("amplitude.New failed: %v", err)
	}
	defer amp.Close()

	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	last := time.Now()
	interval := 2 * time.Second

	ampLevel := 0.0

	for {
		select {
		case _ = <-ticker.C:
			ledsState := leds.State{}

			interval = time.Duration(math.Max(
				defaultIntervalMS-(defaultIntervalMS*ampLevel),
				100,
			)) * time.Millisecond

			now := time.Now()
			loc := 1.0 - float64(now.Sub(last).Nanoseconds())/float64(interval.Nanoseconds())

			if loc < 0 {
				loc = 1
				last = now
			}

			// streaks
			sineMap := leds.GetSineVals(baseLEDCount, loc*baseLEDCount, bauxStreakLength)
			for led, value := range sineMap {
				mag := float64(value) / 254.0
				ledsState.Set(baseLEDStrip, led, leds.Brightness(leds.DeepPurple, mag))
			}

			// globe
			for i := 0; i < len(globeLeds)-1; i++ {
				start := globeLeds[i]
				end := globeLeds[i+1]
				length := end - start

				peak := float64(length) * loc

				sineMap := leds.GetSineVals(length, peak, length/2)
				for led, value := range sineMap {
					mag := (float64(value) / 254.0)
					ledsState.Set(globalLEDStrip, start+led, leds.Brightness(leds.Red, mag))
				}
			}

			ledStrips.Clear()
			ledStrips.Set(ledsState)

		case level := <-amp.Level():
			ampLevel = level

		case <-ctx.Done():
			log.Info("context done")
			return
		}
	}
}
