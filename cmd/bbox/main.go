package main

import (
	"fmt"
	"github.com/siggy/bbox/bbox"
	"time"
)

const (
	BPM      = 120
	TICKS    = 16
	INTERVAL = 60 * time.Second / BPM / (TICKS / 4)
)

func main() {
	fmt.Printf("bbox.Init()1\n")
	audio := bbox.Init()
	fmt.Printf("audio %+v\n", audio)
	fmt.Printf("bbox.Init()2\n")

	lastTick := TICKS - 1
	curTick := 0

	// interval := 100 * time.Millisecond
	ticker := time.NewTicker(INTERVAL)
	quit := make(chan struct{})
	last := time.Now().UnixNano()
	go func() {
		for {
			select {
			case <-ticker.C:
				// fmt.Println("Tick at", t)
				now := time.Now().UnixNano()
				fmt.Println("DIFF1: ", now-last)
				fmt.Println("DIFF2: ", float64(now-last-INTERVAL.Nanoseconds())/float64(time.Second.Nanoseconds()))
				last = now

				// fmt.Println("DIFF: ", t.UnixNano()-last-interval.Nanoseconds())
				if (curTick % 2) == 0 {
					audio.PlayZero(3)
				}

				lastTick = curTick
				curTick = (curTick + 1) % TICKS

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// Tickers can be stopped like timers. Once a ticker
	// is stopped it won't receive any more values on its
	// channel. We'll stop ours after 1600ms.
	time.Sleep(time.Millisecond * 160000000)
	// ticker.Stop()
	// fmt.Println("Ticker stopped")

	// go audio.Play(0)
	// go audio.Play(1)
	// go audio.Play(2)
	// go audio.Play(3)

	// go bbox.RunInput()
	// go bbox.RunAudio()
	// audio.Play(0)
	// fmt.Printf("audio.PlayZero(3) 1\n")
	// audio.PlayZero(3)
	// fmt.Printf("audio.PlayZero(3) 2\n")

	// i := 0
	// for {
	// 	go audio.PlayZero(3)
	// 	time.Sleep(1000 * time.Millisecond)
	// 	go audio.PlayZero(3)
	// 	time.Sleep(100 * time.Millisecond)
	// 	go audio.PlayZero(3)
	// 	time.Sleep(10 * time.Millisecond)
	// 	// audio.Play(0)
	// 	// audio.Play(0)
	// 	// audio.Play(0)
	// 	time.Sleep(2000 * time.Millisecond)
	// 	i++
	// }
}
