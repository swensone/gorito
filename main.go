package main

// typedef unsigned char Uint8;
// void SineWave(void *userdata, Uint8 *stream, int len);
import "C"
import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/swensone/gorito/audio"
	"github.com/swensone/gorito/config"
	"github.com/swensone/gorito/emulator"
	"github.com/swensone/gorito/graphics"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		slog.Default().Error("error parsing config", slog.Any("error", err))
		os.Exit(1)
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     cfg.Level,
	})
	log := slog.New(handler)
	slog.SetDefault(log)
	log.Debug("configuration", "cfg", cfg)

	if err := initSDL(); err != nil {
		slog.Error("failed to init sdl", slog.Any("error", err))
		os.Exit(1)
	}
	defer sdl.Quit()

	// if the extension matches a mode, use it
	romext := filepath.Ext(cfg.ROM)
	if romext == ".xo8" {
		cfg.Mode = emulator.MODE_XOCHIP
	} else if romext == ".sc8" {
		cfg.Mode = emulator.MODE_SUPERCHIP
	}

	// create a title for the display window
	screenName := fmt.Sprintf("gorito - mode %s - %s", cfg.Mode.String(), emulator.RomName(cfg.ROM))

	// create our graphics service
	colormap := map[uint8]*graphics.RGB{
		0: cfg.BG,
		1: cfg.FG1,
		2: cfg.FG2,
		3: cfg.FG3,
	}
	display, err := graphics.New(screenName, cfg.Width, cfg.Height, cfg.Fullscreen, cfg.Mode, colormap)
	if err != nil {
		log.Error("failed to create graphics renderer", slog.Any("error", err))
	}
	defer display.Close()

	// create our audio service
	audio, err := audio.New()
	if err != nil {
		log.Error("failed to init sdl audio", slog.Any("error", err))
	}
	defer audio.Close()

	emu, err := emulator.New(log, cfg.Opcodes, cfg.Mode, display, audio)
	if err != nil {
		log.Error("failure while creating cpu emulator", "error", err)
		os.Exit(1)
	}

	if err := emu.Run(cfg.ROM); err != nil {
		log.Error("error returned from cpu run", "error", err)
	}
}

func initSDL() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}
	if _, err := sdl.ShowCursor(0); err != nil {
		return err
	}
	return nil
}
