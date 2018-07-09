package drums

// can't pass a slice of non-direction channels as a slice of directional
// channels, so we have to convert the whole slice to directional first.
func WriteonlyBeats(channels []chan Beats) []chan<- Beats {
	ret := make([]chan<- Beats, len(channels))
	for n, ch := range channels {
		ret[n] = ch
	}
	return ret
}

func WriteonlyInt(channels []chan int) []chan<- int {
	ret := make([]chan<- int, len(channels))
	for n, ch := range channels {
		ret[n] = ch
	}
	return ret
}

func WriteonlyInterval(channels []chan Interval) []chan<- Interval {
	ret := make([]chan<- Interval, len(channels))
	for n, ch := range channels {
		ret[n] = ch
	}
	return ret
}
