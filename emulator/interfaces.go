package emulator

import "github.com/swensone/gorito/types"

type Audio interface {
	Play()
	Stop()
	LoadPattern(pattern [16]uint8)
	SetPitch(pitch uint8)
}

type Display interface {
	Draw(gfx []types.Color) error
}
