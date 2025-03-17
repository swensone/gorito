package emulator

// setVXToDelay: FX07: Sets VX to the value of the delay timer.
func (e *Emulator) setVXToDelay(X uint8) {
	e.registers[X] = e.delayTimer
}

// setDelayTimerToVX: FX15: Sets the delay timer to VX.
func (e *Emulator) setDelayTimerToVX(X uint8) {
	e.delayTimer = e.registers[X]
}

// setSoundTimerToVX: FX18: Sets the sound timer to VX.
func (e *Emulator) setSoundTimerToVX(X uint8) {
	e.soundTimer = e.registers[X]
}
