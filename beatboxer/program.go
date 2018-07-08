package beatboxer

import "github.com/siggy/bbox/beatboxer/render"

type Program interface {
	Init(render func(render.RenderState))
	Pressed(row int, column int)
}
