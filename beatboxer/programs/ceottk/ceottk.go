package ceottk

import (
	"fmt"
	"time"

	"github.com/siggy/bbox/beatboxer/render"
	"github.com/siggy/bbox/beatboxer/wavs"
)

const (
	SEQEUNCE_LENGTH = 123
)

var (
	aliens = map[int]struct{}{
		6:   struct{}{},
		12:  struct{}{},
		26:  struct{}{},
		32:  struct{}{},
		45:  struct{}{},
		52:  struct{}{},
		57:  struct{}{},
		73:  struct{}{},
		76:  struct{}{},
		79:  struct{}{},
		100: struct{}{},
		123: struct{}{},
	}

	humanMods = map[int]int{
		7:  1,
		8:  2,
		9:  3,
		10: 4,
		11: 5,

		13: 1,
		14: 2,
		15: 3,
		16: 4,
		17: 5,

		18: 1,
		19: 2,
		20: 3,
		21: 4,
		22: 5,

		23: 1,
		24: 2,
		25: 3,

		27: 1,
		28: 2,
		29: 3,
		30: 4,
		31: 5,
	}
)

type Ceottk struct {
	location int
	player   wavs.Player
}

func (c *Ceottk) Init(player wavs.Player, render func(render.RenderState)) {
	c.player = player

	// player.Play("ceottk001_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk002_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk003_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk004_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk005_human.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk006_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk001_human.wav") // 7
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk002_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk003_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk004_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk005_human.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk012_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// time.Sleep(3000 * time.Millisecond)

	// player.Play("ceottk001_human.wav") // 13
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk002_human.wav") // 14
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk003_human.wav") // 15
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk004_human.wav") // 16
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk005_human.wav") // 17
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk001_human.wav") // 18
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk002_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk003_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk004_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk005_human.wav") // 22
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk001_human.wav") // 23
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk002_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk003_human.wav") // 25
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk026_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// time.Sleep(3000 * time.Millisecond)

	// player.Play("ceottk001_human.wav") // 27
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk002_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk003_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk004_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk005_human.wav") // 31
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk032_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// time.Sleep(14000 * time.Millisecond)

	// player.Play("ceottk033_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk034_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk035_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk036_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk037_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk038_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk039_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk040_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk041_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk042_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk043_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk044_human.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk045_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// time.Sleep(3000 * time.Millisecond)

	// player.Play("ceottk046_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk047_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk048_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk049_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk050_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk051_human.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk052_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// time.Sleep(1000 * time.Millisecond)

	// player.Play("ceottk053_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk054_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk055_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk056_human.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk057_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// time.Sleep(4000 * time.Millisecond)

	// player.Play("ceottk058_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk059_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk060_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk061_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk062_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk063_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk064_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk065_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk066_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk067_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk068_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk069_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk070_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk071_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk072_human.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk073_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk074_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk075_human.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk076_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk077_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk078_human.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk079_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// time.Sleep(2000 * time.Millisecond)

	// player.Play("ceottk080_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk081_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk082_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk083_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk084_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk085_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk086_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk087_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk088_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk089_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk090_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk091_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk092_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk093_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk094_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk095_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk096_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk097_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk098_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk099_human.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk100_alien.wav")
	// time.Sleep(500 * time.Millisecond)

	// time.Sleep(2000 * time.Millisecond)

	// player.Play("ceottk101_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk102_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk103_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk104_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk105_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk106_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk107_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk108_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk109_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk110_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk111_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk112_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk113_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk114_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk115_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk116_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk117_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk118_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk119_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk120_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk121_human.wav")
	// time.Sleep(500 * time.Millisecond)
	// player.Play("ceottk122_human.wav")
	// time.Sleep(500 * time.Millisecond)

	// player.Play("ceottk123_alien.wav")
	// time.Sleep(500 * time.Millisecond)
}

func (c *Ceottk) Pressed(row int, column int) {
	c.location++

	loc := c.location
	if location, ok := humanMods[c.location]; ok {
		loc = location
	}

	human := fmt.Sprintf("ceottk%03d_human.wav", loc)
	c.player.Play(human)

	if _, ok := aliens[c.location+1]; ok {
		time.Sleep(500 * time.Millisecond)
		c.location++
		alien := fmt.Sprintf("ceottk%03d_alien.wav", c.location)
		c.player.Play(alien)
	}
}
