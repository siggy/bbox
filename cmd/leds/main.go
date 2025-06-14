package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/siggy/bbox/bbox/color"
	// "github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/bbox/bbox2/leds"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	KEYS_LED_COUNT1 = 5 * 60
	KEYS_LED_COUNT2 = 5 * 60
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	leds.Run()
	os.Exit(0)

	// leds.InitLeds(leds.DEFAULT_FREQ, KEYS_LED_COUNT1, KEYS_LED_COUNT2)

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
			//i, err := strconv.Atoi(text)
			//if err != nil {
			//	fmt.Printf("strconv.Atoi failed: %+v\n", err)
			//}

			if text == ";" {
				cur = int(math.Abs(float64(cur - 1)))
			} else if text == "'" {
				cur = cur + 1
			}

			ws2811.Clear()

			fmt.Printf("LED: %+v\n", cur)
			ws2811.SetLed(0, cur, color.Red)
			ws2811.SetLed(1, cur, color.Red)

			err := ws2811.Render()
			if err != nil {
				fmt.Printf("ws2811.Render failed: %+v\n", err)
				panic(err)
			}

			err = ws2811.Wait()
			if err != nil {
				fmt.Printf("ws2811.Wait failed: %+v\n", err)
				panic(err)
			}

			time.Sleep(time.Millisecond)
		}
	}
}
