package emulator

// waitKeyPress: FX0A: A key press is awaited, and then stored in VX (blocking operation, all instruction halted until next
// key event, delay and sound timers should continue processing)
func (e *Emulator) waitKeyPress(X uint8) {
	for i := range e.keys {
		if e.prevKeys[i] && !e.keys[i] {
			e.registers[X] = uint8(i)
			return
		}
	}
	e.pc -= 2
}
