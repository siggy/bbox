package color

func Make(r uint32, g uint32, b uint32, w uint32) uint32 {
	return uint32(b + g<<8 + r<<16 + w<<24)
}
