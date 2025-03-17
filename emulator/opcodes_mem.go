package emulator

import (
	"github.com/swensone/gorito/gmath"
	"github.com/swensone/gorito/types"
)

// saveVXthroughVY: 5XY2: Save an inclusive range of registers VX to VY in memory starting at I
func (e *Emulator) saveVXthroughVY(X, Y uint8) {
	registers := gmath.Abs(int(X) - int(Y))
	if X > Y {
		for i := 0; i <= registers; i++ {
			e.memory[e.idx+uint16(i)] = e.registers[X-uint8(i)]
		}
	} else {
		for i := 0; i <= registers; i++ {
			e.memory[e.idx+uint16(i)] = e.registers[X+uint8(i)]
		}
	}
}

// loadVXthroughVY: 5XY3: Load an inclusive range of registers VX to VY from memory starting at I
func (e *Emulator) loadVXthroughVY(X, Y uint8) {
	registers := gmath.Abs(int(X) - int(Y))
	if X > Y {
		for i := 0; i <= registers; i++ {
			e.registers[X-uint8(i)] = e.memory[e.idx+uint16(i)]
		}
	} else {
		for i := 0; i <= registers; i++ {
			e.registers[X+uint8(i)] = e.memory[e.idx+uint16(i)]
		}
	}
}

// setItoNNN: ANNN: Sets I to the address NNN
func (e *Emulator) setItoNNN(NNN uint16) {
	e.idx = NNN
}

// loadHiMem: F000 NNNN: Load I with 16-bit address NNNN
func (e *Emulator) loadHiMem(NNNN uint16) {
	e.pc += 2
	e.idx = NNNN
}

// storeRegistersInMemory: FX55: Stores from V0 to VX (including VX) in memory, starting at address I. The offset from I is increased by 1
// for each value written, but I itself is left unmodified.[d][24]
func (e *Emulator) storeRegistersInMemory(X uint8) {
	for i := range X + 1 {
		e.memory[e.idx+uint16(i)] = e.registers[i]
	}

	// handle memory quirk
	if e.cfg.Mode != types.MODE_SUPERCHIP {
		e.idx += (uint16(X) + 1)
	}
}

// storeMemInRegisters: FX65: Fills from V0 to VX (including VX) with values from memory, starting at address I. The offset from I
// is increased by 1 for each value read, but I itself is left unmodified.
func (e *Emulator) storeMemInRegisters(X uint8) {
	for i := range X + 1 {
		e.registers[i] = e.memory[e.idx+uint16(i)]
	}

	// handle memory quirk
	if e.cfg.Mode != types.MODE_SUPERCHIP {
		e.idx += (uint16(X) + 1)
	}
}

// storeRegistersToStorage: FX75: Store V0..VX in RPL user flags (X <= 7 if superchip, X <= 16 if xo-chip)
func (e *Emulator) storeRegistersToStorage(X uint8) {
	e.log.Debug("storeRegistersToStorag", "X", X)
	if e.cfg.Mode == types.MODE_CHIP8 {
		return
	} else if e.cfg.Mode == types.MODE_SUPERCHIP && X > 7 {
		X = 7
	}

	e.storage.Persist(e.rom, e.registers[:int(X)+1])
}

// loadRegistersFromStorage: FX85: Read V0..VX from RPL user flags (X <= 7 if superchip, X <= 16 if xo-chip)
func (e *Emulator) loadRegistersFromStorage(X uint8) {
	e.log.Debug("loadRegistersFromStorage", "X", X)
	if e.cfg.Mode == types.MODE_CHIP8 {
		return
	} else if e.cfg.Mode == types.MODE_SUPERCHIP && X > 7 {
		X = 7
	}

	e.storage.Load(e.rom, int(X), e.registers)
}
