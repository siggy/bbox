package color

import "fmt"

// color => [r,g,b,w]
func ParseColor(color uint32) (uint32, uint32, uint32, uint32) {
	b := color & 0x000000ff
	g := (color & 0x0000ff00) >> 8
	r := (color & 0x00ff0000) >> 16
	w := (color & 0xff000000) >> 24

	return r, g, b, w
}

func ColorStr(color uint32) string {
	r, g, b, w := ParseColor(color)
	return fmt.Sprintf("(%+v, %+v, %+v, %+v)", r, g, b, w)
}

func PrintColor(color uint32) {
	fmt.Printf("%s\n", ColorStr(color))
}
