package leds

import (
	"encoding/binary"
	"fmt"

	"github.com/siggy/rpi_ws281x/golang/ws2811"
)

const (
	GPIO_PIN1  = 18  // PWM0, must be 18 or 12
	GPIO_PIN2  = 13  // PWM1, must be 13 for rPI 3
	LED_COUNT1 = 30  // 144 * 5 // * 5 // * (1 + 5 + 5) // 30/m
	LED_COUNT2 = 30  // 144 * 5 // * 5 // * (1 + 5 + 5) // 30/m
	BRIGHTNESS = 255 // 0-255
)

var (
	Red    = binary.LittleEndian.Uint32([]byte{0x00, 0x00, 0x20, 0x00})
	redw   = binary.LittleEndian.Uint32([]byte{0x00, 0x00, 0x20, 0x10})
	green  = binary.LittleEndian.Uint32([]byte{0x00, 0x20, 0x00, 0x00})
	greenw = binary.LittleEndian.Uint32([]byte{0x00, 0x20, 0x00, 0x10})
	blue   = binary.LittleEndian.Uint32([]byte{0x20, 0x00, 0x00, 0x00})
	bluew  = binary.LittleEndian.Uint32([]byte{0x20, 0x00, 0x00, 0x10})
	white  = binary.LittleEndian.Uint32([]byte{0x20, 0x20, 0x20, 0x00})
	whitew = binary.LittleEndian.Uint32([]byte{0x20, 0x20, 0x20, 0x10})

	Colors = []uint32{Red, redw, green, greenw, blue, bluew, white, whitew}
)

/*
 * Standalone functions to test all LEDs
 */
func Init() {
	fmt.Printf("ws2811.Init()\n")
	err := ws2811.Init(
		GPIO_PIN1, LED_COUNT1, BRIGHTNESS,
		GPIO_PIN2, LED_COUNT2, BRIGHTNESS,
	)
	if err != nil {
		fmt.Printf("ws2811.Init failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.Render()\n")
	err = ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}
}

func SetLed(channel int, led int) {
	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.SetLed(%+v)\n", led)
	ws2811.SetLed(channel, led, Red)

	fmt.Printf("ws2811.Render()\n")
	err := ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}
}

func SetLeds(led int) {
	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	for i, color := range Colors {
		for j := 0; j < LED_COUNT1; j++ {
			index := (led + i + len(Colors)*j) % LED_COUNT1
			// fmt.Printf("ws2811.SetLed(%+v, %+v)\n", index, color)
			ws2811.SetLed(0, index, color)
		}
		for j := 0; j < LED_COUNT2; j++ {
			index := (led + i + len(Colors)*j) % LED_COUNT2
			// fmt.Printf("ws2811.SetLed(%+v, %+v)\n", index, color)
			ws2811.SetLed(1, index, color)
		}
	}

	fmt.Printf("ws2811.Render()\n")
	err := ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}
}

func Shutdown() {
	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.Render()\n")
	err := ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Fini()\n")
	ws2811.Fini()
}

// Turn off all LEDs
func Clear(gpioPin1 int, ledCount1 int, gpioPin2 int, ledCount2 int) {
	fmt.Printf("ws2811.Init()\n")
	err := ws2811.Init(
		gpioPin1, ledCount1, BRIGHTNESS,
		gpioPin2, ledCount2, BRIGHTNESS,
	)
	if err != nil {
		fmt.Printf("ws2811.Init failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Clear()\n")
	ws2811.Clear()

	fmt.Printf("ws2811.Render()\n")
	err = ws2811.Render()
	if err != nil {
		fmt.Printf("ws2811.Render failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Wait()\n")
	err = ws2811.Wait()
	if err != nil {
		fmt.Printf("ws2811.Wait failed: %+v\n", err)
		panic(err)
	}

	fmt.Printf("ws2811.Fini()\n")
	ws2811.Fini()
}
