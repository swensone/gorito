package emulator

import "github.com/swensone/gorito/types"

// returnFromSubroutine: 00EE: Return from subroutine
// Return from subroutine. Set the PC to the address at the top of the stack and subtract 1 from the SP.
func (e *Emulator) returnFromSubroutine() {
	e.pc = e.stack[e.sp]
	e.sp--
}

// exitInterpreter: 00FD: Exits the interpreter (superchip extension)
func (e *Emulator) exitInterpreter() {
	e.finished = true
}

// jumpToNNN: 1NNN: Jumps to address NNN
// set PC to NNN
func (e *Emulator) jumpToNNN(NNN uint16) {
	e.pc = NNN
	e.pc -= 2
}

// callNNN: 2NNN: Call subroutine at NNN
// Increment the SP and put the current PC value on the top of the stack. Then set the PC to NNN. Generally there is a limit of 16 successive calls.
func (e *Emulator) callNNN(NNN uint16) {
	e.sp++
	if len(e.stack) <= int(e.sp) {
		panic("stack overflow")
	}

	e.stack[e.sp] = e.pc
	e.pc = NNN
	e.pc -= 2
}

// jumpToNNNplusV0: BNNN: Jumps to the address NNN plus V0
// superChip works as BXNN: It will jump to the address XNN, plus the value in the register VX
func (e *Emulator) jumpToNNNplusV0(X uint8, NNN uint16) {
	if e.cfg.Mode != types.MODE_SUPERCHIP {
		X = 0
	}
	e.pc = uint16(e.registers[X]) + NNN
	e.pc -= 2
}
