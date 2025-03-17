package emulator

import "fmt"

func (e *Emulator) execOpcode() error {
	// fetch the opcode
	opcode := e.opcodeAt(e.pc)
	B1 := uint8(opcode >> 8)
	B2 := uint8(opcode)
	N1 := B1 & 0xF0 >> 4
	N2 := B1 & 0x0F
	N3 := B2 & 0xF0 >> 4
	N4 := B2 & 0x0F
	NNN := opcode & 0x0FFF

	if e.cfg.LogOpcodes {
		e.log.Debug("running opcode",
			"opcode", fmt.Sprintf("%04X", opcode),
			"vx", fmt.Sprintf("%02d", e.registers[N2]),
			"vy", fmt.Sprintf("%02d", e.registers[N3]),
			"vf", fmt.Sprintf("%02d", e.registers[0xF]),
			"pc", fmt.Sprintf("%02X", e.pc),
			"idx", fmt.Sprintf("%02X", e.idx),
			"sp", fmt.Sprintf("%02X", e.sp),
			"timer", e.timer)
	}

	if B1 == 0x00 && N3 == 0xc {
		// 00CN: Scroll the display down by 0 to 15 pixels
		e.scrollDown(N4)
	} else if B1 == 0x00 && N3 == 0xd {
		// 00DN: Scroll the display up by 0 to 15 pixels
		e.scrollUp(N4)
	} else if opcode == 0x00FB {
		// 00FB Scroll the display right by 4 pixels
		e.scrollRight()
	} else if opcode == 0x00FC {
		// 00FC: Scroll the display left by 4 pixels.
		e.scrollLeft()
	} else if opcode == 0x00e0 {
		// 00E0: clear display
		e.clearDisplay()
	} else if opcode == 0x00ee {
		// 00EE: Return from subroutine
		e.returnFromSubroutine()
	} else if opcode == 0x00fd {
		// 00FD: Exit interpreter (superchip extension)
		e.exitInterpreter()
	} else if opcode == 0x00fe {
		// 00FE: Disable high-resolution mode (superchip extension)
		e.disableHiRes()
	} else if opcode == 0x00ff {
		// 00FF: Enable high-resolution mode (superchip extension)
		e.enableHiRes()
	} else if N1 == 0x1 {
		// 1NNN: Jumps to address NNN
		e.jumpToNNN(NNN)
	} else if N1 == 0x2 {
		// 2NNN: Calls subroutine at NNN
		e.callNNN(NNN)
	} else if N1 == 0x3 {
		// 3XNN: Skips the next instruction if VX equals NN
		e.skipIfVXEqualsNN(N2, B2)
	} else if N1 == 0x4 {
		// 4XNN: Skips the next instruction if VX does not equal NN
		e.skipIfVXNotEqualsNN(N2, B2)
	} else if N1 == 0x5 && N4 == 0x0 {
		// 5XY0: Skips the next instruction if VX equals VY
		e.skipIfVXEqualsVY(N2, N3)
	} else if N1 == 0x5 && N4 == 0x2 {
		// 5XY2: Save an inclusive range of registers VX to VY in memory starting at idx
		e.saveVXthroughVY(N2, N3)
	} else if N1 == 0x5 && N4 == 0x3 {
		// 5XY3: Load an inclusive range of registers VX to VY from memory starting at idx
		e.loadVXthroughVY(N2, N3)
	} else if N1 == 0x6 {
		// 6XNN: Sets VX to NN
		e.setVXtoNN(N2, B2)
	} else if N1 == 0x7 {
		// 7XNN: Adds NN to VX (carry flag is not changed)
		e.addNNtoVX(N2, B2)
	} else if N1 == 0x8 && N4 == 0x0 {
		// 8XY0: Sets VX to the value of VY
		e.setVXtoVY(N2, N3)
	} else if N1 == 0x8 && N4 == 0x1 {
		// 8XY1: Sets VX to VX or VY (bitwise OR operation)
		e.setVXtoVXorVY(N2, N3)
	} else if N1 == 0x8 && N4 == 0x2 {
		// 8XY2: Sets VX to VX and VY (bitwise AND operation)
		e.setVXtoVXandVY(N2, N3)
	} else if N1 == 0x8 && N4 == 0x3 {
		// 8XY3: Sets VX to VX xor VY (bitwise XOR operation)
		e.setVXtoVXxorVY(N2, N3)
	} else if N1 == 0x8 && N4 == 0x4 {
		// 8XY4: Adds VY to VX, VF is set to 1 when there's an overflow, and to 0 when there is not
		e.addVYtoVX(N2, N3)
	} else if N1 == 0x8 && N4 == 0x5 {
		// 8XY5: VY is subtracted from VX, VF is set to 0 when there's an underflow, and 1 when there is not (i.e. VF
		// set to 1 if VX >= VY and 0 if not)
		e.subVYFromVX(N2, N3)
	} else if N1 == 0x8 && N4 == 0x6 {
		// 8XY6: Shifts VX to the right by 1, then stores the least significant bit of VX prior to the shift into VF
		e.shiftVXRight(N2, N3)
	} else if N1 == 0x8 && N4 == 0x7 {
		// 8XY7: Sets VX to VY minus VX. VF is set to 0 when there's an underflow, and 1 when there is not. (i.e. VF
		// set to 1 if VY >= VX)
		e.subVXFromVY(N2, N3)
	} else if N1 == 0x8 && N4 == 0xe {
		// 8XYE: Shifts VX to the left by 1, then sets VF to 1 if the most significant bit of VX prior to that shift
		// was set, or to 0 if it was unset.
		e.shiftVXLeft(N2, N3)
	} else if N1 == 0x9 && N4 == 0x0 {
		// 9XY0: Skips the next instruction if VX does not equal VY. (Usually the next instruction is a jump to skip a
		// code block).
		e.skipIfVXnotEqualsVY(N2, N3)
	} else if N1 == 0xA {
		// ANNN: Sets idx to the address NNN
		e.setItoNNN(NNN)
	} else if N1 == 0xB {
		// BNNN: Jumps to the address NNN plus V0
		e.jumpToNNNplusV0(N2, NNN)
	} else if N1 == 0xC {
		// CXNN: Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN
		e.setVXtoNNNandRand(N2, B2)
	} else if N1 == 0xD {
		// DXYN: Draws a sprite at coordinate (VX, VN3) that has a width of 8 pixels and a height of N pixels.
		// Each row of 8 pixels is read as bit-coded starting from memory location idx ; idx  value does not change
		// after the execution of this instruction. As described above, VF is set to 1 if any screen pixels are
		// flipped from set to unset when the sprite is drawn, and to 0 if that does not happen.
		e.drawSprite(N2, N3, N4)
	} else if N1 == 0xE && B2 == 0x9E {
		// EX9E: Skips the next instruction if the key stored in VX(only consider the lowest nibble) is pressed
		// (usually the next instruction is a jump to skip a code block)
		e.skipIfPressed(N2)
	} else if N1 == 0xE && B2 == 0xA1 {
		// EXA1: Skips the next instruction if the key stored in VX(only consider the lowest nibble) is not pressed
		// (usually the next instruction is a jump to skip a code block)
		e.skipIfNotPressed(N2)
	} else if opcode == 0xF000 {
		// F000: Load the next two bytes into idx
		e.loadHiMem(e.opcodeAt(e.pc + 2))
	} else if opcode == 0xF002 {
		// F002: Store 16 bytes starting at idx in the audio pattern buffer
		e.loadAudioPattern()
	} else if N1 == 0xF && B2 == 0x01 {
		// FX01: Select bit planes to draw on
		e.selectPlane(N2)
	} else if N1 == 0xF && B2 == 0x07 {
		// FX07: Sets VX to the value of the delay timer.
		e.setVXToDelay(N2)
	} else if N1 == 0xF && B2 == 0x0A {
		// FX0A: A key press is awaited, and then stored in VX (blocking operation, all instruction halted until next
		// key event, delay and sound timers should continue processing)
		e.waitKeyPress(N2)
	} else if N1 == 0xF && B2 == 0x15 {
		// FX15: Sets the delay timer to VX.
		e.setDelayTimerToVX(N2)
	} else if N1 == 0xF && B2 == 0x18 {
		// FX18: Sets the sound timer to VX.
		e.setSoundTimerToVX(N2)
	} else if N1 == 0xF && B2 == 0x1E {
		// FX1E: Adds VX to idx. VF is not affected.
		e.addVXtoI(N2)
	} else if N1 == 0xF && B2 == 0x29 {
		// FX29: Sets idx to the location of the sprite for the character in VX (only consider the lowest nibble).
		// Characters 0-F (in hexadecimal) are represented by a 4x5 font.
		e.setItoChar(N2)
	} else if N1 == 0xF && B2 == 0x3A {
		// FX3A: Set the audio pattern playback rate to 4000*2^((vx-64)/48)Hz
		e.setAudioPitch(N2)
	} else if N1 == 0xF && B2 == 0x30 {
		// FX30: Sets idx to the location of the sprite for the character in VX (only consider the lowest nibble).
		// Characters 0-9 are represented by a 8x10 font.
		e.setItoHiresChar(N2)
	} else if N1 == 0xF && B2 == 0x33 {
		// FX33: Stores the binary-coded decimal representation of VX, with the hundreds digit in memory at location
		// in idx , the tens digit at location idx+1, and the ones digit at location idx+2.
		e.storeVXatIinBCD(N2)
	} else if N1 == 0xF && B2 == 0x55 {
		// FX55: Stores from V0 to VX (including VX) in memory, starting at address idx. The offset from idx  is increased by 1
		// for each value written, but idx  itself is left unmodified.[d][24]
		e.storeRegistersInMemory(N2)
	} else if N1 == 0xF && B2 == 0x65 {
		// FX65: Fills from V0 to VX (including VX) with values from memory, starting at address idx. The offset from idx
		// is increased by 1 for each value read, but idx  itself is left unmodified.
		e.storeMemInRegisters(N2)
	} else if N1 == 0xF && B2 == 0x75 {
		// FX75: Store V0..VX in RPL user flags (X <= 7 if superchip, X <= 16 if xo-chip)
		e.storeRegistersToStorage(N2)
	} else if N1 == 0xF && B2 == 0x85 {
		// FX85: Read V0..VX from RPL user flags (X <= 7 if superchip, X <= 16 if xo-chip)
		e.loadRegistersFromStorage(N2)
	} else {
		e.log.Error("bad opcode: unable to interpret opcode", "opcode", fmt.Sprintf("%04X", opcode))
	}

	// increment the program counter by two bytes
	e.pc += 2
	e.counter++
	return nil
}

func (e *Emulator) opcodeAt(pc uint16) uint16 {
	opcode := uint16(e.memory[pc])<<8 | uint16(e.memory[pc+1])
	return opcode
}
