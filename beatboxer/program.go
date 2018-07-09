package beatboxer

import (
	"github.com/siggy/bbox/beatboxer/render"
	"github.com/siggy/bbox/beatboxer/wavs"
)

type Program interface {
	Init(player wavs.Player, render func(render.RenderState))
	Pressed(row int, column int)
}
