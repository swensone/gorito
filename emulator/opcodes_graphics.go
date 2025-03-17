package emulator

import "github.com/swensone/gorito/types"

// clearDisplay: 00E0: clears display. on xo-chip, clears the selected display plane.
func (e *Emulator) clearDisplay() {
	for i := range PLANES {
		if e.plane>>i&0x01 == 0x01 {
			for j := range XRES * YRES {
				e.gfx[i][j] = 0
			}
		}
	}

	e.drawFlag = true
}

// disableHiRes: 00FE: Disable high-resolution mode
func (e *Emulator) disableHiRes() {
	e.hires = false
}

// enableHiRes: 00FF: Enable high-resolution mode
func (e *Emulator) enableHiRes() {
	e.hires = true
}

// drawSprite: DXYN: Draws a sprite at coordinate (VX, VY) that has a width of 8 pixels and a height of N pixels.
// Each row of 8 pixels is read as bit-coded starting from memory location I; I value does not change after the
// execution of this instruction. As described above, VF is set to 1 if any screen pixels are flipped from set to
// unset when the sprite is drawn, and to 0 if that does not happen.
func (e *Emulator) drawSprite(X, Y, N uint8) {
	xres := int(e.xres)
	yres := int(e.yres)
	// handle display wait quirk
	if e.cfg.Mode == types.MODE_CHIP8 {
		if e.counter%4 != 0 {
			e.pc -= 2
			return
		}
	}

	spriteWidth := 8
	spriteHeight := int(N)
	if e.hires && N == 0 {
		spriteWidth = 16
		spriteHeight = 16
	}

	scaleFactor := 2
	if e.hires {
		scaleFactor = 1
	}

	VX := int(e.registers[X]) * scaleFactor
	VY := int(e.registers[Y]) * scaleFactor
	if e.cfg.Mode != types.MODE_XOCHIP {
		if VX >= xres {
			VX = VX % xres
		}
		if VY >= yres {
			VY = VY % yres
		}
	}

	offset := e.idx
	e.registers[0xF] = 0x00
	for i := range spriteHeight {
		spriteData := uint16(e.memory[offset])
		offset++
		if spriteWidth == 16 {
			spriteData = spriteData<<8 | uint16(e.memory[offset])
			offset++
		}

		for bit := range spriteWidth {
			posX := VX + bit*scaleFactor
			posY := VY + i*scaleFactor

			if e.cfg.Mode == types.MODE_XOCHIP {
				posX = posX % int(xres)
				posY = posY % int(yres)
			} else {
				if posX < 0 || posX >= int(xres) {
					continue
				}
				if posY < 0 || posY >= int(yres) {
					continue
				}
			}

			set := uint8(spriteData >> (spriteWidth - 1 - bit) & 0x01)
			if e.drawAt(posX, posY, scaleFactor, set) {
				e.registers[0xF] = 0x01
			}
		}
	}
	e.drawFlag = true
}

func (e *Emulator) drawAt(x, y, scaleFactor int, set uint8) bool {
	flip := false
	for xoffset := range scaleFactor {
		for yoffset := range scaleFactor {
			gfxIdx := (y+yoffset)*int(e.xres) + x + xoffset

			for i := range PLANES {
				if e.plane>>i&0x01 == 0x01 {
					prevSet := e.gfx[i][gfxIdx]
					e.gfx[i][gfxIdx] = prevSet ^ set
					if prevSet&set == 001 {
						flip = true
					}
				}
			}
		}
	}
	return flip
}

// selectPlane: FX01: Select bit planes to draw on
func (e *Emulator) selectPlane(X uint8) {
	if X > 3 {
		e.log.Error("SelectPlane passed invalid value for X (must be 0-3)", "X", X)
		return
	}
	e.plane = X
}

// setItoChar: FX29: Sets I to the location of the sprite for the character in VX (only consider the lowest nibble).
// Characters 0-F (in hexadecimal) are represented by a 4x5 font.
func (e *Emulator) setItoChar(X uint8) {
	e.idx = FONT_OFFSET + uint16(e.registers[X])*5
}

// setItoHiresChar: FX30: Sets I to the location of the sprite for the character in VX (only consider the lowest nibble).
// Characters 0-9 are represented by a 8x10 font.
func (e *Emulator) setItoHiresChar(X uint8) {
	e.idx = SUPERCHIP_FONT_OFFSET + uint16(e.registers[X])*10
}

// storeVXatIinBCD: FX33: Stores the binary-coded decimal representation of VX, with the hundreds digit in memory at location
// in I, the tens digit at location I+1, and the ones digit at location I+2.
func (e *Emulator) storeVXatIinBCD(X uint8) {
	vx := e.registers[X]
	for i := range 3 {
		e.memory[e.idx+2-uint16(i)] = vx % 10
		vx = vx / 10
	}
}
