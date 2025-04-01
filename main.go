package main

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
	"github.com/swensone/gorito/types"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		slog.Default().Error("error parsing config", slog.Any("error", err))
		os.Exit(1)
	}

	log := getLogger(cfg.Level)
	log.Debug("configuration", "cfg", cfg)

	log.Debug("initializing sdl")
	if err := initSDL(); err != nil {
		slog.Error("failed to init sdl", slog.Any("error", err))
		os.Exit(1)
	}
	defer sdl.Quit()

	// if the extension matches a mode, use it
	romext := filepath.Ext(cfg.ROM)
	if romext == ".xo8" {
		cfg.Mode = types.MODE_XOCHIP
	} else if romext == ".sc8" {
		cfg.Mode = types.MODE_SUPERCHIP
	}

	// create a title for the display window
	screenName := fmt.Sprintf("gorito - mode %s - %s", cfg.Mode.String(), emulator.RomName(cfg.ROM))

	// create our graphics service
	log.Debug("initializing graphics")
	colorMap := map[uint8]types.Color{
		0: cfg.BG,
		1: cfg.FG1,
		2: cfg.FG2,
		3: cfg.FG3,
	}
	display, err := graphics.New(screenName, cfg.Width, cfg.Height, cfg.Fullscreen, cfg.BG)
	if err != nil {
		log.Error("failed to create graphics renderer", slog.Any("error", err))
	}
	defer display.Close()

	// create our audio service
	log.Debug("initializing audio")
	audio := audio.New(20)
	defer audio.Close()

	emu, err := emulator.New(
		emulator.EmulatorConfig{
			Savefile:   cfg.Savefile,
			Mode:       cfg.Mode,
			Speed:      cfg.Speed,
			ColorMap:   colorMap,
			LogOpcodes: cfg.Opcodes,
		},
		display,
		audio,
		log,
	)
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

func getLogger(level slog.Level) *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})
	log := slog.New(handler)
	slog.SetDefault(log)
	return log
}
