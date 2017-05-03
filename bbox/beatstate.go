package bbox

import (
	"sync/atomic"
)

const (
	DISABLED uint32 = 0
	ENABLED  uint32 = 1
)

type BeatState struct {
	beats [][]*uint32
}

func InitBeatState(beats int) *BeatState {
	bs := BeatState{
		beats: make([][]*uint32, beats),
	}
	for i, _ := range bs.beats {
		bs.beats[i] = make([]*uint32, TICKS)
		for j, _ := range bs.beats[i] {
			num := DISABLED
			bs.beats[i][j] = &num
		}
	}

	return &bs
}

func (bs *BeatState) Toggle(beat int, tick int) {
	val := ENABLED
	if bs.Enabled(beat, tick) {
		val = DISABLED
	}
	atomic.StoreUint32(bs.beats[beat][tick], val)
}

func (bs *BeatState) Enabled(beat int, tick int) bool {
	return atomic.LoadUint32(bs.beats[beat][tick]) == ENABLED
}

func (bs *BeatState) BeatCount() int {
	return len(bs.beats)
}
