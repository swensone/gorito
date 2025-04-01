package emulator

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/swensone/gorito/types"
)

// memory map
// 0x000-0x1FF   - Chip 8 interpreter (contains font set in emu)
// 0x050-0x0A0   - Used for the built in 4x5 pixel font set (0-F)
// 0x200-0xFFF   - Program ROM and work RAM
// 0x1000-0xFFFF - xo-chip high mem range

type EmulatorConfig struct {
	Savefile   string
	Mode       types.Mode
	Speed      uint32
	ColorMap   map[uint8]types.Color
	LogOpcodes bool
}

const (
	XRES   int32 = 128
	YRES   int32 = 64
	PLANES       = 2
)

func New(cfg EmulatorConfig, display Display, audio Audio, log *slog.Logger) (*Emulator, error) {
	storage, err := newStorage(cfg.Savefile, log)
	if err != nil {
		return nil, err
	}

	e := &Emulator{
		cfg:       cfg,
		registers: make([]uint8, 16),
		audio:     audio,
		display:   display,
		xres:      XRES,
		yres:      YRES,
		gfx:       make(map[int][]uint8),
		plane:     1,
		storage:   storage,
		log:       log,
	}

	for i := range PLANES {
		e.gfx[i] = make([]uint8, XRES*YRES)
		e.gfx[i] = make([]uint8, XRES*YRES)
	}
	e.Reset()

	return e, nil
}

type Emulator struct {
	// basic config
	cfg EmulatorConfig
	rom string

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
	gfx      map[int][]uint8
	plane    uint8
	xres     int32
	yres     int32
	hires    bool
	drawFlag bool

	// audio
	audio_pattern [16]uint8
	pitch         uint8

	// key tracking
	prevKeys [16]bool
	keys     [16]bool
	paused   bool
	finished bool

	// interfaces for graphics and sound functionality
	display Display
	audio   Audio

	// persistent storage for superchip and xo-chip
	storage *storage

	// logger
	log *slog.Logger
}

func (e *Emulator) LoadProgram(filepath string) error {
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

	e.log.Debug("loading program", "file", fp)
	var upper uint8
	var lower uint8
	for i, b := range data {
		e.memory[0x200+i] = b

		if i%2 == 0 {
			upper = b
		} else {
			lower = b
			e.log.Debug("loaded data", "location", i, "value", fmt.Sprintf("%02X%02X", upper, lower))
		}

	}
	e.log.Debug("done loading program")
	e.rom = RomName(filepath)

	return nil
}

func (e *Emulator) Run(rom string) error {
	if err := e.LoadProgram(rom); err != nil {
		return errors.Wrapf(err, "unable to open file %s", rom)
	}

	cyclesTimer := time.Now()
	lastDraw := time.Now()
	cycles := 0
	draws := 0
	for {
		start := time.Now()

		// check for keyboard events
		e.setKeys()

		// if we're paused, skip any cpu or graphics updates
		if e.paused && !e.finished {
			time.Sleep(time.Second / 100)
			continue
		}

		// fetch/decode/execute opcodes
		if err := e.execOpcode(); err != nil {
			return errors.Wrap(err, "failed during exec opcode")
		}

		// update the display at approx 60hz
		if time.Since(lastDraw) > time.Second/60 {
			if e.drawFlag {
				if err := e.display.Draw(e.getGfx()); err != nil {
					return errors.Wrap(err, "failed during draw")
				}
				e.drawFlag = false
			}
			lastDraw = time.Now()
			draws++

			// Update timers
			if e.delayTimer > 0 {
				e.delayTimer--
			}

			if e.soundTimer > 0 {
				e.audio.Play()
			} else {
				e.audio.Stop()
			}
			if e.soundTimer > 0 {
				e.soundTimer--
			}
		}

		// slow the emulator down to an approximately right speed
		time.Sleep(time.Until(start.Add(time.Second / time.Duration(e.cfg.Speed))))

		cycles++
		if time.Since(cyclesTimer) > time.Second {
			e.log.Debug("cycles per second", "cycles", cycles, "draws", draws)
			cyclesTimer = time.Now()
			cycles = 0
			draws = 0
		}

		if e.finished {
			return nil
		}
	}
}

func (e *Emulator) getGfx() []types.Color {
	res := make([]types.Color, XRES*YRES)
	for i := range XRES * YRES {
		var val uint8
		for j := range PLANES {
			val |= e.gfx[j][i] << j
		}
		res[i] = e.cfg.ColorMap[val]
	}
	return res
}

func (e *Emulator) Reset() {
	e.pc = 0x200       // Program counter starts at 0x200
	e.idx = 0          // Reset index register
	e.sp = 0           // Reset stack pointer
	e.delayTimer = 0   // Reset delay timer
	e.soundTimer = 0   // Reset sound timer
	e.timer = 0        // Reset timer counter
	e.counter = 0      // Reset counter
	e.paused = false   // Unpause if paused
	e.finished = false // Reset the finished flag

	// Clear graphics memory and reset graphics variables
	e.plane = 3
	e.clearDisplay()
	e.plane = 1
	e.hires = false

	// Clear stack
	for i := range e.stack {
		e.stack[i] = 0
	}

	// Clear registers
	for i := range e.registers {
		e.registers[i] = 0
	}

	// Clear memory
	for i := range e.memory {
		e.memory[i] = 0
	}

	// Clear audio
	// set values to approximately A4 as a default
	e.pitch = 247
	for i := range e.audio_pattern {
		if i < len(e.audio_pattern)/2 {
			e.audio_pattern[i] = 0x00
		} else {
			e.audio_pattern[i] = 0xff
		}
	}

	// Load the fonts
	for i, val := range fontSet {
		e.memory[FONT_OFFSET+i] = val
	}

	for i, val := range superchipFontSet {
		e.memory[SUPERCHIP_FONT_OFFSET+i] = val
	}
}
