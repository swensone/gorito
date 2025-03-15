package emulator

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/cockroachdb/errors"
)

// memory map
// 0x000-0x1FF - Chip 8 interpreter (contains font set in emu)
// 0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
// 0x200-0xFFF - Program ROM and work RAM

func New(log *slog.Logger, logOpcodes bool, mode Mode, display Display, audio Audio) (*CPU, error) {
	storage, err := NewStorage("~/.config/gorito-storage.json", log)
	if err != nil {
		return nil, err
	}

	xres, yres := GetRes(mode)
	c := &CPU{
		registers:  make([]uint8, 16),
		audio:      audio,
		display:    display,
		mode:       mode,
		xres:       xres,
		yres:       yres,
		gfxp1:      make([]uint8, xres*yres),
		gfxp2:      make([]uint8, xres*yres),
		plane:      1,
		storage:    storage,
		log:        log,
		logOpcodes: logOpcodes,
	}
	c.Reset()

	return c, nil
}

type CPU struct {
	// core emulator functionality
	registers  []uint8
	stack      [16]uint16
	sp         uint8
	memory     [64 * 1024]uint8
	idx        uint16
	pc         uint16
	timer      uint32
	delayTimer uint8
	soundTimer uint8
	counter    uint64

	// graphics
	gfxp1    []uint8
	gfxp2    []uint8
	plane    uint8
	xres     int32
	yres     int32
	hires    bool
	drawFlag bool

	// emulator mode for quirk modeling
	mode Mode

	// key tracking
	prevKeys [16]bool
	keys     [16]bool
	paused   bool
	finished bool

	// interfaces for graphics and sound functionality
	display Display
	audio   Audio

	// persistent storage for superchip and xo-chip
	storage *Storage
	rom     string

	// logger
	log        *slog.Logger
	logOpcodes bool
}

func (c *CPU) LoadProgram(filepath string) error {
	if filepath == "" {
		return errors.New("rom path must be specified")
	}
	fp := path.Clean(filepath)
	f, err := os.Open(fp)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	c.log.Debug("loading program", "file", fp)
	var upper uint8
	var lower uint8
	for i, b := range data {
		c.memory[0x200+i] = b

		if i%2 == 0 {
			upper = b
		} else {
			lower = b
			c.log.Debug("loaded data", "location", i, "value", fmt.Sprintf("%02X%02X", upper, lower))
		}

	}
	c.log.Debug("done loading program")
	c.rom = RomName(filepath)

	return nil
}

func (c *CPU) Run(rom string) error {
	if err := c.LoadProgram(rom); err != nil {
		return errors.Wrapf(err, "unable to open file %s", rom)
	}

	cyclesTimer := time.Now()
	lastDraw := time.Now().Add(-2 * time.Minute)
	cycles := 0
	for {
		start := time.Now()

		// check for keyboard events
		c.setKeys()

		// if we're paused, skip any cpu or graphics updates
		if c.paused && !c.finished {
			time.Sleep(time.Second / 100)
			continue
		}

		// fetch/decode/execute opcodes
		if err := c.execOpcode(); err != nil {
			return errors.Wrap(err, "failed during exec opcode")
		}

		// update the display at approx 60hz
		if time.Since(lastDraw) > time.Second/60 {
			if c.drawFlag {
				if err := c.display.Draw(c.getGfx()); err != nil {
					return errors.Wrap(err, "failed during draw")
				}
				c.drawFlag = false
			}
			lastDraw = time.Now()
		}

		// slow the emulator down to an approximately right speed
		time.Sleep(time.Until(start.Add(time.Second / 600)))

		cycles++
		if time.Since(cyclesTimer) > time.Second {
			c.log.Debug("cycles per second", "cycles", cycles)
			cyclesTimer = time.Now()
			cycles = 0
		}

		if c.finished {
			return nil
		}
	}
}

func (c *CPU) getGfx() []uint8 {
	res := make([]uint8, len(c.gfxp1))
	for i := range c.gfxp1 {
		res[i] = c.gfxp2[i]<<1 | c.gfxp1[i]
	}
	return res
}

func (c *CPU) Reset() {
	c.pc = 0x200       // Program counter starts at 0x200
	c.idx = 0          // Reset index register
	c.sp = 0           // Reset stack pointer
	c.delayTimer = 0   // Reset delay timer
	c.soundTimer = 0   // Reset sound timer
	c.timer = 0        // Reset timer counter
	c.counter = 0      // Reset counter
	c.paused = false   // Unpause if paused
	c.finished = false // Reset the finished flag

	// Clear graphics memory
	c.plane = 3
	c.clearDisplay()
	c.plane = 1
	c.hires = false

	// Clear stack
	for i := range c.stack {
		c.stack[i] = 0
	}

	// Clear registers
	for i := range c.registers {
		c.registers[i] = 0
	}

	// Clear memory
	for i := range c.memory {
		c.memory[i] = 0
	}

	// Load the fonts
	for i, val := range fontSet {
		c.memory[FONT_OFFSET+i] = val
	}

	for i, val := range superchipFontSet {
		c.memory[SUPERCHIP_FONT_OFFSET+i] = val
	}
}

func (c *CPU) clearDisplay() {
	if c.plane&0x01 > 0 {
		for i := range c.gfxp1 {
			c.gfxp1[i] = 0
		}
	}
	if c.plane&0x02 > 0 {
		for i := range c.gfxp1 {
			c.gfxp1[i] = 0
		}
	}
	c.drawFlag = true
}

func (c *CPU) execOpcode() error {
	// fetch the opcode
	opcode := c.opcodeAt()
	B1 := uint8(opcode >> 8)
	B2 := uint8(opcode)
	N1 := B1 & 0xF0 >> 4
	N2 := B1 & 0x0F
	N3 := B2 & 0xF0 >> 4
	N4 := B2 & 0x0F
	NNN := opcode & 0x0FFF

	if c.logOpcodes {
		c.log.Debug("running opcode",
			"opcode", fmt.Sprintf("%04X", opcode),
			"vx", fmt.Sprintf("%02d", c.registers[N2]),
			"vy", fmt.Sprintf("%02d", c.registers[N3]),
			"vf", fmt.Sprintf("%02d", c.registers[0xF]),
			"pc", fmt.Sprintf("%02X", c.pc),
			"idx", fmt.Sprintf("%02X", c.idx),
			"sp", fmt.Sprintf("%02X", c.sp),
			"timer", c.timer)
	}

	if B1 == 0x00 && N3 == 0xc {
		// scrollDown: 00CN: Scroll the display down by 0 to 15 pixels
		c.scrollDown(N4)
	} else if B1 == 0x00 && N3 == 0xd {
		// scrollUp: 00DN: Scroll the display up by 0 to 15 pixels
		c.scrollUp(N4)
	} else if opcode == 0x00FB {
		// scrollRight: 00FB Scroll the display right by 4 pixels
		c.scrollRight()
	} else if opcode == 0x00FC {
		// scrollLeft: 00FC: Scroll the display left by 4 pixels.
		c.scrollLeft()
	} else if opcode == 0x00e0 {
		// 00E0: clear display
		c.clearDisplay()
	} else if opcode == 0x00ee {
		// 00EE: Return from subroutine
		c.returnFromSubroutine()
	} else if opcode == 0x00fd {
		// 00FD: Exit interpreter (superchip extension)
		c.exitInterpreter()
	} else if opcode == 0x00fe {
		// 00FE: Disable high-resolution mode (superchip extension)
		c.disableHiRes()
	} else if opcode == 0x00ff {
		// 00FF: Enable high-resolution mode (superchip extension)
		c.enableHiRes()
	} else if N1 == 0x1 {
		// 1NNN: Jumps to address NNN
		c.jumpToNNN(NNN)
	} else if N1 == 0x2 {
		// 2NNN: Calls subroutine at NNN
		c.callNNN(NNN)
	} else if N1 == 0x3 {
		// 3XNN: Skips the next instruction if VX equals NN
		c.skipIfVXEqualsNN(N2, B2)
	} else if N1 == 0x4 {
		// 4XNN: Skips the next instruction if VX does not equal NN
		c.skipIfVXNotEqualsNN(N2, B2)
	} else if N1 == 0x5 && N4 == 0x0 {
		// 5XY0: Skips the next instruction if VX equals VY
		c.skipIfVXEqualsVY(N2, N3)
	} else if N1 == 0x5 && N4 == 0x2 {
		// saveVXthroughVY: 5XY2: Save an inclusive range of registers VX to VY in memory starting at I
		c.saveVXthroughVY(N2, N3)
	} else if N1 == 0x5 && N4 == 0x3 {
		// loadVXthroughVY: 5XY3: Load an inclusive range of registers VX to VY from memory starting at I
		c.loadVXthroughVY(N2, N3)
	} else if N1 == 0x6 {
		// 6XNN: Sets VX to NN
		c.setVXtoNN(N2, B2)
	} else if N1 == 0x7 {
		// 7XNN: Adds NN to VX (carry flag is not changed)
		c.addNNtoVX(N2, B2)
	} else if N1 == 0x8 && N4 == 0x0 {
		// 8XY0: Sets VX to the value of VY
		c.setVXtoVY(N2, N3)
	} else if N1 == 0x8 && N4 == 0x1 {
		// 8XY1: Sets VX to VX or VY (bitwise OR operation)
		c.setVXtoVXorVY(N2, N3)
	} else if N1 == 0x8 && N4 == 0x2 {
		// 8XY2: Sets VX to VX and VY (bitwise AND operation)
		c.setVXtoVXandVY(N2, N3)
	} else if N1 == 0x8 && N4 == 0x3 {
		// 8XY3: Sets VX to VX xor VY (bitwise XOR operation)
		c.setVXtoVXxorVY(N2, N3)
	} else if N1 == 0x8 && N4 == 0x4 {
		// 8XY4: Adds VY to VX, VF is set to 1 when there's an overflow, and to 0 when there is not
		c.addVYtoVX(N2, N3)
	} else if N1 == 0x8 && N4 == 0x5 {
		// 8XY5: VY is subtracted from VX, VF is set to 0 when there's an underflow, and 1 when there is not (i.e. VF
		// set to 1 if VX >= VY and 0 if not)
		c.subVYFromVX(N2, N3)
	} else if N1 == 0x8 && N4 == 0x6 {
		// 8XY6: Shifts VX to the right by 1, then stores the least significant bit of VX prior to the shift into VF
		c.shiftVXRight(N2, N3)
	} else if N1 == 0x8 && N4 == 0x7 {
		// 8XY7: Sets VX to VY minus VX. VF is set to 0 when there's an underflow, and 1 when there is not. (i.e. VF
		// set to 1 if VY >= VX)
		c.subVXFromVY(N2, N3)
	} else if N1 == 0x8 && N4 == 0xe {
		// 8XYE: Shifts VX to the left by 1, then sets VF to 1 if the most significant bit of VX prior to that shift
		// was set, or to 0 if it was unset.
		c.shiftVXLeft(N2, N3)
	} else if N1 == 0x9 && N4 == 0x0 {
		// 9XY0: Skips the next instruction if VX does not equal VY. (Usually the next instruction is a jump to skip a
		// code block).
		c.skipIfVXnotEqualsVY(N2, N3)
	} else if N1 == 0xA {
		// ANNN: Sets IDX to the address NNN
		c.setItoNNN(NNN)
	} else if N1 == 0xB {
		// BNNN: Jumps to the address NNN plus V0
		c.jumpToNNNplusV0(N2, NNN)
	} else if N1 == 0xC {
		// CXNN: Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN
		c.setVXtoNNNandRand(N2, B2)
	} else if N1 == 0xD {
		// DXYN: Draws a sprite at coordinate (VX, VN3) that has a width of 8 pixels and a height of N pixels.
		// Each row of 8 pixels is read as bit-coded starting from memory location IDX ; IDX  value does not change
		// after the execution of this instruction. As described above, VF is set to 1 if any screen pixels are
		// flipped from set to unset when the sprite is drawn, and to 0 if that does not happen.
		c.drawSprite(N2, N3, N4)
	} else if N1 == 0xE && B2 == 0x9E {
		// EX9E: Skips the next instruction if the key stored in VX(only consider the lowest nibble) is pressed
		// (usually the next instruction is a jump to skip a code block)
		c.skipIfPressed(N2)
	} else if N1 == 0xE && B2 == 0xA1 {
		// EXA1: Skips the next instruction if the key stored in VX(only consider the lowest nibble) is not pressed
		// (usually the next instruction is a jump to skip a code block)
		c.skipIfNotPressed(N2)
	} else if opcode == 0xF000 {
		c.pc += 2
		c.loadHiMem(c.opcodeAt())
	} else if N1 == 0xF && B2 == 0x01 {
		// FX01: Select bit planes to draw on
		c.SelectPlane(N2)
	} else if N1 == 0xF && B2 == 0x07 {
		// FX07: Sets VX to the value of the delay timer.
		c.setVXToDelay(N2)
	} else if N1 == 0xF && B2 == 0x0A {
		// FX0A: A key press is awaited, and then stored in VX (blocking operation, all instruction halted until next
		// key event, delay and sound timers should continue processing)
		c.waitKeyPress(N2)
	} else if N1 == 0xF && B2 == 0x15 {
		// FX15: Sets the delay timer to VX.
		c.setDelayTimerToVX(N2)
	} else if N1 == 0xF && B2 == 0x18 {
		// FX18: Sets the sound timer to VX.
		c.setSoundTimerToVX(N2)
	} else if N1 == 0xF && B2 == 0x1E {
		// FX1E: Adds VX to IDX. VF is not affected.
		c.addVXtoI(N2)
	} else if N1 == 0xF && B2 == 0x29 {
		// FX29: Sets IDX to the location of the sprite for the character in VX (only consider the lowest nibble).
		// Characters 0-F (in hexadecimal) are represented by a 4x5 font.
		c.setItoChar(N2)
	} else if N1 == 0xF && B2 == 0x30 {
		// setItoHiresChar: FX30: Sets IDX to the location of the sprite for the character in VX (only consider the lowest nibble).
		// Characters 0-9 are represented by a 8x10 font.
		c.setItoHiresChar(N2)
	} else if N1 == 0xF && B2 == 0x33 {
		// FX33: Stores the binary-coded decimal representation of VX, with the hundreds digit in memory at location
		// in IDX , the tens digit at location IDX+1, and the ones digit at location IDX+2.
		c.storeVXatIinBCD(N2)
	} else if N1 == 0xF && B2 == 0x55 {
		// FX55: Stores from V0 to VX (including VX) in memory, starting at address IDX. The offset from IDX  is increased by 1
		// for each value written, but IDX  itself is left unmodified.[d][24]
		c.storeRegistersInMemory(N2)
	} else if N1 == 0xF && B2 == 0x65 {
		// FX65: Fills from V0 to VX (including VX) with values from memory, starting at address IDX. The offset from IDX
		// is increased by 1 for each value read, but IDX  itself is left unmodified.
		c.storeMemInRegisters(N2)
	} else if N1 == 0xF && B2 == 0x75 {
		// FX75: Store V0..VX in RPL user flags (X <= 7 if superchip, X <= 16 if xo-chip)
		c.storeRegistersToStorage(N2)
	} else if N1 == 0xF && B2 == 0x85 {
		// FX85: Read V0..VX from RPL user flags (X <= 7 if superchip, X <= 16 if xo-chip)
		c.loadRegistersFromStorage(N2)
	} else {
		c.log.Error("bad opcode: unable to interpret opcode", "opcode", fmt.Sprintf("%04X", opcode))
	}

	// increment the program counter by two bytes
	c.pc += 2

	// Update timers
	c.timer++
	if c.timer >= 10 {
		if c.delayTimer > 0 {
			c.delayTimer--
		}

		c.audio.Beep(c.soundTimer >= 1)
		if c.soundTimer > 0 {
			c.soundTimer--
		}
		c.timer = 0
	}

	c.counter++
	return nil
}

func (c *CPU) opcodeAt() uint16 {
	opcode := uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])
	return opcode
}
