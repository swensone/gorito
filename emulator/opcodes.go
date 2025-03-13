package emulator

import (
	"crypto/rand"
	"math/big"
)

// returnFromSubroutine: 00EE: Return from subroutine
// Return from subroutine. Set the PC to the address at the top of the stack and subtract 1 from the SP.
func (c *CPU) returnFromSubroutine() {
	c.pc = c.stack[c.sp]
	c.sp--
}

// scrollDown: 00CN: Scroll the display down by 0 to 15 pixels
func (c *CPU) scrollDown(N uint8) {
	scrollLen := int(c.xres) * int(N)
	scroll := make([]bool, scrollLen)
	c.gfx = append(scroll, c.gfx[0:len(c.gfx)-scrollLen]...)
}

// scrollRight: 00FB: Scroll the display right by 4 pixels
func (c *CPU) scrollRight() {
	newGfx := []bool{}
	scroll := make([]bool, 4)
	for i := range c.yres {
		newGfx = append(newGfx, scroll...)
		newGfx = append(newGfx, c.gfx[i*c.xres:(i+1)*c.xres-4]...)
	}
	c.gfx = newGfx
}

// scrollLeft: 00FC: Scroll the display left by 4 pixels.
func (c *CPU) scrollLeft() {
	newGfx := []bool{}
	scroll := make([]bool, 4)
	for i := range c.yres {
		newGfx = append(newGfx, c.gfx[i*c.xres+4:(i+1)*c.xres]...)
		newGfx = append(newGfx, scroll...)
	}
	c.gfx = newGfx
}

// exitInterpreter: 00FD: Exits the interpreter (superchip extension)
func (c *CPU) exitInterpreter() {
	c.finished = true
}

// disableHiRes: 00FE: Disable high-resolution mode
func (c *CPU) disableHiRes() {
	c.hires = false
}

// enableHiRes: 00FF: Enable high-resolution mode
func (c *CPU) enableHiRes() {
	c.hires = true
}

// jumpToNNN: 1NNN: Jumps to address NNN
// set PC to NNN
func (c *CPU) jumpToNNN(NNN uint16) {
	c.pc = NNN
	c.pc -= 2
}

// callNNN: 2NNN: Call subroutine at NNN
// Increment the SP and put the current PC value on the top of the stack. Then set the PC to NNN. Generally there is a limit of 16 successive calls.
func (c *CPU) callNNN(NNN uint16) {
	c.sp++
	if len(c.stack) <= int(c.sp) {
		panic("stack overflow")
	}

	c.stack[c.sp] = c.pc
	c.pc = NNN
	c.pc -= 2
}

// skipIfVXEqual: 3XNN: Skips the next instruction if VX equals NN
func (c *CPU) skipIfVXEqualsNN(X, NN uint8) {
	if c.registers[X] == NN {
		c.pc += 2
	}
}

// skipIfVXNotEqual: 4XNN: Skips the next instruction if VX does not equal NN
func (c *CPU) skipIfVXNotEqualsNN(X, NN uint8) {
	if c.registers[X] != NN {
		c.pc += 2
	}
}

// skipIfVXEqualsVY: 5XY0: Skips the next instruction if VX equals VY
func (c *CPU) skipIfVXEqualsVY(X, Y uint8) {
	if c.registers[X] == c.registers[Y] {
		c.pc += 2
	}
}

// setVXtoNN: 6XNN: Sets VX to NN
func (c *CPU) setVXtoNN(X, NN uint8) {
	c.registers[X] = NN
}

// addNNtoVX: 7XNN: Adds NN to VX (carry flag is not changed)
func (c *CPU) addNNtoVX(X, NN uint8) {
	c.registers[X] += NN
}

// setVXtoVY: 8XY0: Sets VX to the value of VY
func (c *CPU) setVXtoVY(X, Y uint8) {
	c.registers[X] = c.registers[Y]
}

// setVXtoVXorVY: 8XY1: Sets VX to VX or VY (bitwise OR operation)
func (c *CPU) setVXtoVXorVY(X, Y uint8) {
	c.registers[X] |= c.registers[Y]

	// handle VF reset quirk
	if c.mode == MODE_CHIP8 {
		c.registers[0xF] = 0
	}
}

// setVXtoVXandVY: 8XY2: Sets VX to VX and VY (bitwise AND operation)
func (c *CPU) setVXtoVXandVY(X, Y uint8) {
	c.registers[X] &= c.registers[Y]

	// handle VF reset quirk
	if c.mode == MODE_CHIP8 {
		c.registers[0xF] = 0
	}
}

// setVXtoVXxorVY: 8XY3: Sets VX to VX xor VY (bitwise XOR operation)
func (c *CPU) setVXtoVXxorVY(X, Y uint8) {
	c.registers[X] ^= c.registers[Y]

	// handle VF reset quirk
	if c.mode == MODE_CHIP8 {
		c.registers[0xF] = 0
	}
}

// addVYtoVX: 8XY4: Adds VY to VX, VF is set to 1 when there's an overflow, and to 0 when there is not
func (c *CPU) addVYtoVX(X, Y uint8) {
	VX := c.registers[X]
	VY := c.registers[Y]
	res := uint16(VX) + uint16(VY)
	c.registers[X] = uint8(res)

	c.registers[0xF] = uint8(res >> 8)
}

// subVYFromVX: 8XY5: VY is subtracted from VX, VF is set to 0 when there's an underflow, and 1 when there is not (i.e. VF
// set to 1 if VX >= VY and 0 if not)
func (c *CPU) subVYFromVX(X, Y uint8) {
	VX := c.registers[X]
	VY := c.registers[Y]
	res := uint16(VX) - uint16(VY)
	c.registers[X] = uint8(res)

	if VX >= VY {
		c.registers[0xF] = 0x01
	} else {
		c.registers[0xF] = 0x00
	}
}

// shiftVXRight: 8XY6: Shifts VX to the right by 1, then stores the least significant bit of VX prior to the shift into VF
func (c *CPU) shiftVXRight(X, Y uint8) {
	VX := c.registers[X]
	if c.mode != MODE_SUPERCHIP {
		c.registers[X] = c.registers[Y]
	}
	c.registers[X] = c.registers[X] >> 1

	c.registers[0xF] = VX & 0x01
}

// subVXFromVY: 8XY7: Sets VX to VY minus VX. VF is set to 0 when there's an underflow, and 1 when there is not. (i.e. VF
// set to 1 if VY >= VX)
func (c *CPU) subVXFromVY(X, Y uint8) {
	VX := c.registers[X]
	VY := c.registers[Y]
	res := uint16(VY) - uint16(VX)
	c.registers[X] = uint8(res)

	if VY >= VX {
		c.registers[0xF] = 0x01
	} else {
		c.registers[0xF] = 0x00
	}
}

// shiftVXLeft: 8XYE: Shifts VX to the left by 1, then sets VF to 1 if the most significant bit of VX prior to that shift
// was set, or to 0 if it was unset.
func (c *CPU) shiftVXLeft(X, Y uint8) {
	VX := c.registers[X]
	if c.mode != MODE_SUPERCHIP {
		c.registers[X] = c.registers[Y]
	}
	c.registers[X] = c.registers[X] << 1
	c.registers[0xF] = (0x80 & VX) >> 7
}

// skipIfVXnotEqualsVY: 9XY0: Skips the next instruction if VX does not equal VY. (Usually the next instruction is a jump to skip a
// code block).
func (c *CPU) skipIfVXnotEqualsVY(X, Y uint8) {
	if c.registers[X] != c.registers[Y] {
		c.pc += 2
	}
}

// setItoNNN: ANNN: Sets I to the address NNN
func (c *CPU) setItoNNN(NNN uint16) {
	c.idx = NNN
}

// jumpToNNNplusV0: BNNN: Jumps to the address NNN plus V0
// superChip works as BXNN: It will jump to the address XNN, plus the value in the register VX
func (c *CPU) jumpToNNNplusV0(X uint8, NNN uint16) {
	if c.mode != MODE_SUPERCHIP {
		X = 0
	}
	c.pc = uint16(c.registers[X]) + NNN
	c.pc -= 2
}

// setVXtoNNandRand CXNN: Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN
func (c *CPU) setVXtoNNNandRand(X, NN uint8) {
	r, err := rand.Int(rand.Reader, big.NewInt(255))
	if err != nil {
		panic(err)
	}
	c.registers[X] = uint8(r.Int64()) & NN
}

// drawSprite: DXYN: Draws a sprite at coordinate (VX, VY) that has a width of 8 pixels and a height of N pixels.
// Each row of 8 pixels is read as bit-coded starting from memory location I; I value does not change after the
// execution of this instruction. As described above, VF is set to 1 if any screen pixels are flipped from set to
// unset when the sprite is drawn, and to 0 if that does not happen.
func (c *CPU) drawSprite(X, Y, N uint8) {
	xres := int(c.xres)
	yres := int(c.yres)
	// handle display wait quirk
	if c.mode == MODE_CHIP8 {
		if c.counter%4 != 0 {
			c.pc -= 2
			return
		}
	}

	spriteWidth := 8
	spriteHeight := int(N)
	if c.hires && N == 0 {
		spriteWidth = 16
		spriteHeight = 16
	}

	scaleFactor := 2
	if c.hires {
		scaleFactor = 1
	}

	VX := int(c.registers[X]) * scaleFactor
	VY := int(c.registers[Y]) * scaleFactor
	if c.mode != MODE_XOCHIP {
		if VX >= xres {
			VX = VX % xres
		}
		if VY >= yres {
			VY = VY % yres
		}
	}

	offset := c.idx
	c.registers[0xF] = 0x00
	for i := range spriteHeight {
		spriteData := uint16(c.memory[offset])
		offset++
		if spriteWidth == 16 {
			spriteData = spriteData<<8 | uint16(c.memory[offset])
			offset++
		}

		for bit := range spriteWidth {
			posX := VX + bit*scaleFactor
			posY := VY + i*scaleFactor

			if c.mode == MODE_XOCHIP {
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

			set := (spriteData >> (spriteWidth - 1 - bit) & 0x01) == 1
			if c.drawAt(posX, posY, scaleFactor, set) {
				c.registers[0xF] = 0x01
			}
		}
	}
	c.drawFlag = true
}

func (c *CPU) drawAt(x, y, scaleFactor int, set bool) bool {
	flip := false
	for xoffset := range scaleFactor {
		for yoffset := range scaleFactor {
			gfxIdx := (y+yoffset)*int(c.xres) + x + xoffset

			prevSet := c.gfx[gfxIdx]
			c.gfx[gfxIdx] = prevSet != set
			if prevSet && set {
				flip = true
			}
		}
	}
	return flip
}

// skipIfPressed: EX9E: Skips the next instruction if the key stored in VX(only consider the lowest nibble) is pressed
// (usually the next instruction is a jump to skip a code block)
func (c *CPU) skipIfPressed(X uint8) {
	if c.keys[c.registers[X]] {
		c.pc += 2
	}
}

// skipIfNotPressed EXA1: Skips the next instruction if the key stored in VX(only consider the lowest nibble) is not pressed
// (usually the next instruction is a jump to skip a code block)
func (c *CPU) skipIfNotPressed(X uint8) {
	if !c.keys[c.registers[X]] {
		c.pc += 2
	}
}

// waitKeyPress: FX0A: A key press is awaited, and then stored in VX (blocking operation, all instruction halted until next
// key event, delay and sound timers should continue processing)
func (c *CPU) waitKeyPress(X uint8) {
	for i := range c.keys {
		if c.prevKeys[i] && !c.keys[i] {
			c.registers[X] = uint8(i)
			return
		}
	}
	c.pc -= 2
}

// setVXToDelay: FX07: Sets VX to the value of the delay timer.
func (c *CPU) setVXToDelay(X uint8) {
	c.registers[X] = c.delayTimer
}

// setDelayTimerToVX: FX15: Sets the delay timer to VX.
func (c *CPU) setDelayTimerToVX(X uint8) {
	c.delayTimer = c.registers[X]
}

// setSoundTimerToVX: FX18: Sets the sound timer to VX.
func (c *CPU) setSoundTimerToVX(X uint8) {
	c.soundTimer = c.registers[X]
}

// addVXtoI: FX1E: Adds VX to I. VF is not affected.
func (c *CPU) addVXtoI(X uint8) {
	c.idx += uint16(c.registers[X])
}

// setItoChar: FX29: Sets I to the location of the sprite for the character in VX (only consider the lowest nibble).
// Characters 0-F (in hexadecimal) are represented by a 4x5 font.
func (c *CPU) setItoChar(X uint8) {
	c.idx = FONT_OFFSET + uint16(c.registers[X])*5
}

// setItoHiresChar: FX30: Sets I to the location of the sprite for the character in VX (only consider the lowest nibble).
// Characters 0-9 are represented by a 8x10 font.
func (c *CPU) setItoHiresChar(X uint8) {
	c.idx = SUPERCHIP_FONT_OFFSET + uint16(c.registers[X])*10
}

// storeVXatIinBCD: FX33: Stores the binary-coded decimal representation of VX, with the hundreds digit in memory at location
// in I, the tens digit at location I+1, and the ones digit at location I+2.
func (c *CPU) storeVXatIinBCD(X uint8) {
	vx := c.registers[X]
	for i := range 3 {
		c.memory[c.idx+2-uint16(i)] = vx % 10
		vx = vx / 10
	}
}

// storeRegistersInMemory: FX55: Stores from V0 to VX (including VX) in memory, starting at address I. The offset from I is increased by 1
// for each value written, but I itself is left unmodified.[d][24]
func (c *CPU) storeRegistersInMemory(X uint8) {
	for i := range X + 1 {
		c.memory[c.idx+uint16(i)] = c.registers[i]
	}

	// handle memory quirk
	if c.mode != MODE_SUPERCHIP {
		c.idx += (uint16(X) + 1)
	}
}

// storeMemInRegisters: FX65: Fills from V0 to VX (including VX) with values from memory, starting at address I. The offset from I
// is increased by 1 for each value read, but I itself is left unmodified.
func (c *CPU) storeMemInRegisters(X uint8) {
	for i := range X + 1 {
		c.registers[i] = c.memory[c.idx+uint16(i)]
	}

	// handle memory quirk
	if c.mode != MODE_SUPERCHIP {
		c.idx += (uint16(X) + 1)
	}
}

// storeRegistersToStorage: FX75: Store V0..VX in RPL user flags (X <= 7 if superchip, X <= 16 if xo-chip)
func (c *CPU) storeRegistersToStorage(X uint8) {
	if c.mode == MODE_CHIP8 {
		return
	} else if c.mode == MODE_SUPERCHIP && X > 7 {
		X = 7
	}
	registersToPersist := uint16(X) + 1
	c.storage.Persist(c.rom, c.registers[:registersToPersist])
}

// loadRegistersFromStorage: FX85: Read V0..VX from RPL user flags (X <= 7 if superchip, X <= 16 if xo-chip)
func (c *CPU) loadRegistersFromStorage(X uint8) {
	if c.mode == MODE_CHIP8 {
		return
	} else if c.mode == MODE_SUPERCHIP && X > 7 {
		X = 7
	}
	registersToLoad := uint16(X) + 1

	data := c.storage.Load(c.rom, uint16(X)+1)
	copy(c.registers[:registersToLoad], data)
}
