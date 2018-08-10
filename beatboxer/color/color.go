package color

type Color struct {
	R, G, B, W uint32
}

var (
	pink       = Make(159, 0, 159, 93)
	trueBlue   = Make(0, 0, 255, 0)
	TrueBlue   = trueBlue
	red        = Make(210, 0, 50, 40)
	lightGreen = Make(0, 181, 115, 43)
	TrueRed    = Make(255, 0, 0, 0)
	trueWhite  = Make(0, 0, 0, 255)
	purple     = Make(82, 0, 197, 52)
	mint       = Make(0, 27, 0, 228)
	trueGreen  = Make(0, 255, 0, 0)
	deepPurple = Make(200, 0, 100, 0)

	Colors = []uint32{
		pink,
		trueBlue,
		red,
		lightGreen,
		TrueRed,
		deepPurple,
		trueWhite,
		purple,
		mint,
		trueGreen,
	}

	redWhite = Make(255, 0, 0, 255)
	black    = Make(0, 0, 0, 0)
	Black    = black
)

func Make(r uint32, g uint32, b uint32, w uint32) uint32 {
	return uint32(b + g<<8 + r<<16 + w<<24)
}

func Split(color uint32) Color {
	return Color{
		B: color & 0x000000ff,
		G: (color & 0x0000ff00) >> 8,
		R: (color & 0x00ff0000) >> 16,
		W: (color & 0xff000000) >> 24,
	}
}
