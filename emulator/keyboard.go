package emulator

import (
	"github.com/veandco/go-sdl2/sdl"
)

/*
CHIP-8 has an odd hex-based keypad layout:
1	2	3	C
4	5	6	D
7	8	9	E
A	0	B	F

We remap the first four columns of the standard ASCII keyboard in order to support this as best as possible.
*/

var keymap = map[int]int{
	sdl.SCANCODE_1: 0x1, // map key 1 to 1
	sdl.SCANCODE_2: 0x2, // map key 2 to 2
	sdl.SCANCODE_3: 0x3, // map key 3 to 3
	sdl.SCANCODE_4: 0xC, // map key 4 to C
	sdl.SCANCODE_Q: 0x4, // map key Q to 4
	sdl.SCANCODE_W: 0x5, // map key W to 5
	sdl.SCANCODE_E: 0x6, // map key E to 6
	sdl.SCANCODE_R: 0xD, // map key R to D
	sdl.SCANCODE_A: 0x7, // map key A to 7
	sdl.SCANCODE_S: 0x8, // map key S to 8
	sdl.SCANCODE_D: 0x9, // map key D to 9
	sdl.SCANCODE_F: 0xE, // map key F to E
	sdl.SCANCODE_Z: 0xA, // map key Z to A
	sdl.SCANCODE_X: 0x0, // map key X to 0
	sdl.SCANCODE_C: 0xB, // map key C to B
	sdl.SCANCODE_V: 0xF, // map key V to F
}

func (c *CPU) setKeys() {
	copy(c.prevKeys[:], c.keys[:])

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch ke := event.(type) {
		case *sdl.KeyboardEvent:
			if ke.Type == sdl.KEYUP && ke.Keysym.Scancode == sdl.SCANCODE_P {
				c.paused = !c.paused
			}
		case *sdl.QuitEvent:
			c.finished = true
		}
	}

	keyState := sdl.GetKeyboardState()

	for key, mapped := range keymap {
		c.keys[mapped] = keyState[key] == 1
	}

	if keyState[sdl.SCANCODE_ESCAPE] == 1 {
		c.finished = true
	}
}
