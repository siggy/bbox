package beatboxer

import (
	"time"

	"github.com/siggy/bbox/beatboxer/render"
)

// Output defines the interface Beatboxer programs may use to send output
type Output interface {
	Play(name string) time.Duration
	Render(rs render.RenderState)
	Yield()
}

// Program defines the interface all Beatboxer programs must satisfy
type Program interface {
	Init(output Output)
	Pressed(row int, column int)
}
