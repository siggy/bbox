package beatboxer

import (
	"github.com/siggy/bbox/beatboxer/render"
)

// Output defines the interface Beatboxer programs may use to send output
type Output interface {
	Play(name string)
	Render(rs render.RenderState)
}

// Program defines the interface all Beatboxer programs must satisfy
type Program interface {
	Init(output Output)
	Pressed(row int, column int)
}
