package emulator

import "github.com/swensone/gorito/types"

type Audio interface {
	Beep(on bool) error
}

type Display interface {
	Draw(gfx []types.Color) error
}
