package emulator

// skipIfVXEqual: 3XNN: Skips the next instruction if VX equals NN
func (e *Emulator) skipIfVXEqualsNN(X, NN uint8) {
	if e.registers[X] == NN {
		e.pc += 2
		if e.opcodeAt(e.pc) == 0xF000 {
			e.pc += 2
		}
	}
}

// skipIfVXNotEqual: 4XNN: Skips the next instruction if VX does not equal NN
func (e *Emulator) skipIfVXNotEqualsNN(X, NN uint8) {
	if e.registers[X] != NN {
		e.pc += 2
		if e.opcodeAt(e.pc) == 0xF000 {
			e.pc += 2
		}
	}
}

// skipIfVXEqualsVY: 5XY0: Skips the next instruction if VX equals VY
func (e *Emulator) skipIfVXEqualsVY(X, Y uint8) {
	if e.registers[X] == e.registers[Y] {
		e.pc += 2
		if e.opcodeAt(e.pc) == 0xF000 {
			e.pc += 2
		}
	}
}

// skipIfVXnotEqualsVY: 9XY0: Skips the next instruction if VX does not equal VY. (Usually the next instruction is a jump to skip a
// code block).
func (e *Emulator) skipIfVXnotEqualsVY(X, Y uint8) {
	if e.registers[X] != e.registers[Y] {
		e.pc += 2
		if e.opcodeAt(e.pc) == 0xF000 {
			e.pc += 2
		}
	}
}

// skipIfPressed: EX9E: Skips the next instruction if the key stored in VX (only consider the lowest nibble) is pressed
// (usually the next instruction is a jump to skip a code block)
func (e *Emulator) skipIfPressed(X uint8) {
	if e.keys[e.registers[X]] {
		e.pc += 2
		if e.opcodeAt(e.pc) == 0xF000 {
			e.pc += 2
		}
	}
}

// skipIfNotPressed EXA1: Skips the next instruction if the key stored in VX (only consider the lowest nibble) is not pressed
// (usually the next instruction is a jump to skip a code block)
func (e *Emulator) skipIfNotPressed(X uint8) {
	if !e.keys[e.registers[X]] {
		e.pc += 2
		if e.opcodeAt(e.pc) == 0xF000 {
			e.pc += 2
		}
	}
}
