package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/siggy/bbox/bbox/leds"
	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	LED_COUNT1 = 600
	LED_COUNT2 = 600
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	leds.InitLeds(LED_COUNT1, LED_COUNT2)

	defer func() {
		ws2811.Clear()
		ws2811.Render()
		ws2811.Wait()
		ws2811.Fini()
	}()

	for {
		select {
		case <-sig:
			return
		default:
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter LED: ")
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)
			i, err := strconv.Atoi(text)
			if err != nil {
				fmt.Printf("strconv.Atoi failed: %+v\n", err)
			}

			ws2811.Clear()

			fmt.Printf("LED: %+v\n", i)
			ws2811.SetLed(0, i, leds.Red)
			ws2811.SetLed(1, i, leds.Red)

			err = ws2811.Render()
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
