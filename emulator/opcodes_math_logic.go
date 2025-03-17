package emulator

import (
	"crypto/rand"
	"math/big"

	"github.com/swensone/gorito/types"
)

// setVXtoNN: 6XNN: Sets VX to NN
func (e *Emulator) setVXtoNN(X, NN uint8) {
	e.registers[X] = NN
}

// addNNtoVX: 7XNN: Adds NN to VX (carry flag is not changed)
func (e *Emulator) addNNtoVX(X, NN uint8) {
	e.registers[X] += NN
}

// setVXtoVY: 8XY0: Sets VX to the value of VY
func (e *Emulator) setVXtoVY(X, Y uint8) {
	e.registers[X] = e.registers[Y]
}

// setVXtoVXorVY: 8XY1: Sets VX to VX or VY (bitwise OR operation)
func (e *Emulator) setVXtoVXorVY(X, Y uint8) {
	e.registers[X] |= e.registers[Y]

	// handle VF reset quirk
	if e.cfg.Mode == types.MODE_CHIP8 {
		e.registers[0xF] = 0
	}
}

// setVXtoVXandVY: 8XY2: Sets VX to VX and VY (bitwise AND operation)
func (e *Emulator) setVXtoVXandVY(X, Y uint8) {
	e.registers[X] &= e.registers[Y]

	// handle VF reset quirk
	if e.cfg.Mode == types.MODE_CHIP8 {
		e.registers[0xF] = 0
	}
}

// setVXtoVXxorVY: 8XY3: Sets VX to VX xor VY (bitwise XOR operation)
func (e *Emulator) setVXtoVXxorVY(X, Y uint8) {
	e.registers[X] ^= e.registers[Y]

	// handle VF reset quirk
	if e.cfg.Mode == types.MODE_CHIP8 {
		e.registers[0xF] = 0
	}
}

// addVYtoVX: 8XY4: Adds VY to VX, VF is set to 1 when there's an overflow, and to 0 when there is not
func (e *Emulator) addVYtoVX(X, Y uint8) {
	VX := e.registers[X]
	VY := e.registers[Y]
	res := uint16(VX) + uint16(VY)
	e.registers[X] = uint8(res)

	e.registers[0xF] = uint8(res >> 8)
}

// subVYFromVX: 8XY5: VY is subtracted from VX, VF is set to 0 when there's an underflow, and 1 when there is not (i.e. VF
// set to 1 if VX >= VY and 0 if not)
func (e *Emulator) subVYFromVX(X, Y uint8) {
	VX := e.registers[X]
	VY := e.registers[Y]
	res := uint16(VX) - uint16(VY)
	e.registers[X] = uint8(res)

	if VX >= VY {
		e.registers[0xF] = 0x01
	} else {
		e.registers[0xF] = 0x00
	}
}

// shiftVXRight: 8XY6: Shifts VX to the right by 1, then stores the least significant bit of VX prior to the shift into VF
func (e *Emulator) shiftVXRight(X, Y uint8) {
	VX := e.registers[X]
	if e.cfg.Mode != types.MODE_SUPERCHIP {
		e.registers[X] = e.registers[Y]
	}
	e.registers[X] = e.registers[X] >> 1

	e.registers[0xF] = VX & 0x01
}

// subVXFromVY: 8XY7: Sets VX to VY minus VX. VF is set to 0 when there's an underflow, and 1 when there is not. (i.e. VF
// set to 1 if VY >= VX)
func (e *Emulator) subVXFromVY(X, Y uint8) {
	VX := e.registers[X]
	VY := e.registers[Y]
	res := uint16(VY) - uint16(VX)
	e.registers[X] = uint8(res)

	if VY >= VX {
		e.registers[0xF] = 0x01
	} else {
		e.registers[0xF] = 0x00
	}
}

// shiftVXLeft: 8XYE: Shifts VX to the left by 1, then sets VF to 1 if the most significant bit of VX prior to that shift
// was set, or to 0 if it was unset.
func (e *Emulator) shiftVXLeft(X, Y uint8) {
	VX := e.registers[X]
	if e.cfg.Mode != types.MODE_SUPERCHIP {
		e.registers[X] = e.registers[Y]
	}
	e.registers[X] = e.registers[X] << 1
	e.registers[0xF] = (0x80 & VX) >> 7
}

// setVXtoNNandRand CXNN: Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN
func (e *Emulator) setVXtoNNNandRand(X, NN uint8) {
	r, err := rand.Int(rand.Reader, big.NewInt(255))
	if err != nil {
		panic(err)
	}
	e.registers[X] = uint8(r.Int64()) & NN
}

// addVXtoI: FX1E: Adds VX to I. VF is not affected.
func (e *Emulator) addVXtoI(X uint8) {
	e.idx += uint16(e.registers[X])
}
