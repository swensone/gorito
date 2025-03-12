package emulator

type Audio interface {
	Beep(on bool) error
}

type Display interface {
	Draw(gfx []bool) error
}
