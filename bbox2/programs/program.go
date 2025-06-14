package programs

type (
	Coord struct {
		Row int
		Col int
	}

	Color struct {
		R uint8 // 0-255
		G uint8 // 0-255
		B uint8 // 0-255
		W uint8 // 0-255, white channel for RGBW LEDs
	}

	LEDs [][]Color

	// Program defines the interface all Beatboxer programs must satisfy
	Program interface {
		// input
		Press(press Coord)

		// output
		Play() <-chan string
		Render() <-chan LEDs
		Yield() <-chan struct{}
	}
)

const (
	ChannelBuffer = 100
)

var (
	// should match scorpio/code.py
	StripLengths = []int{30, 30, 10, 10, 10, 10, 10, 10}
)

func NewLEDs(stripLengths []int) LEDs {
	leds := make(LEDs, len(stripLengths))
	for i, length := range stripLengths {
		leds[i] = make([]Color, length)
	}
	return leds
}

func (leds LEDs) String() string {
	var str string
	for strip, stripLEDs := range leds {
		str += "Strip " + string(strip) + ": "
		for _, led := range stripLEDs {
			str += "[" + string(led.R) + "," + string(led.G) + "," + string(led.B) + "," + string(led.W) + "] "
		}
		str += "\n"
	}
	return str
}

// byte(strip), pixel, g, r, b, w
// stripLengths = []int{30, 30, 10, 10, 10, 10, 10, 10}

// TODO:
// program interface {
// 	Press() Coord<-
//  Play() <-String
//  Render() <-LEDs
// }
