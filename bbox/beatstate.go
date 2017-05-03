package bbox

// TODO: make this not racey

type BeatState struct {
	beats [][]bool
}

func InitBeatState(beats int) *BeatState {
	bs := BeatState{
		beats: make([][]bool, beats),
	}
	for i, _ := range bs.beats {
		bs.beats[i] = make([]bool, TICKS)
	}

	return &bs
}

func (bs *BeatState) Toggle(beat int, tick int) {
	bs.beats[beat][tick] = !bs.beats[beat][tick]
}

func (bs *BeatState) Get(beat int, tick int) bool {
	return bs.beats[beat][tick]
}

func (bs *BeatState) BeatCount() int {
	return len(bs.beats)
}
