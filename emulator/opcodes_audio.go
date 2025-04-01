package emulator

// F002: Store 16 bytes starting at idx in the audio pattern buffer
func (e *Emulator) loadAudioPattern() {
	for i := range 16 {
		e.audio_pattern[i] = e.memory[int(e.idx)+i]
	}
}

// FX3A: Set the audio pattern playback rate to 4000*2^((vx-64)/48)Hz
func (e *Emulator) setAudioPitch(X uint8) {
	e.audio.SetPitch(e.registers[X])
}
