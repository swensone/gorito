package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	kjson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"

	"github.com/swensone/gorito/emulator"
	"github.com/swensone/gorito/graphics"
)

type Config struct {
	Mode       emulator.Mode `yaml:"mode,omitempty"`
	ROM        string        `yaml:"rom,omitempty"`
	Width      int32         `yaml:"width,omitempty"`
	Height     int32         `yaml:"height,omitempty"`
	Fullscreen bool          `yaml:"fullscreen,omitempty"`
	Opcodes    bool          `yaml:"opcodes,omitempty"`
	Level      *slog.Level   `yaml:"level,omitempty"`
	BG         *graphics.RGB `yaml:"bg,omitempty"`
	FG         *graphics.RGB `yaml:"fg,omitempty"`
}

func Parse() (*Config, error) {
	k := koanf.New(".")
	// set up some decent defaults
	k.Load(confmap.Provider(map[string]interface{}{
		"config":     "~/.config/gorito.yaml",
		"mode":       "superchip",
		"width":      1280,
		"height":     640,
		"fullscreen": false,
		"opcodes":    false,
		"level":      "INFO",
		"bg":         "080808",
		"fg":         "52a6c5",
	}, "."), nil)

	// Parse command line flags
	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}

	f.StringP("config", "c", "", "path to one or more .yaml config files")
	f.StringP("mode", "m", "", fmt.Sprintf("emulator mode, possible values: %s", strings.Join(emulator.SupportedModes(), ", ")))
	f.StringP("rom", "r", "", "path to the rom you want to load")
	f.IntP("width", "x", 0, "window width")
	f.IntP("height", "y", 0, "window height")
	f.BoolP("fullscreen", "f", false, "display full screen")
	f.BoolP("opcodes", "o", false, "log opcodes, extremely noisy")
	f.StringP("level", "l", "", "log level")
	f.String("bg", "", "background color in hex")
	f.String("fg", "", "foreground color in hex")
	if err := f.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	// clean up the yaml config file path and load
	configFile := path.Clean(k.String("config"))
	if strings.HasPrefix(configFile, "~/") {
		home, _ := os.UserHomeDir()
		configFile = filepath.Join(home, configFile[2:])
	}

	if err := k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
		// allow not exists errors for the config file, we can run with defaults or command line flags
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	// load flags and merge into default
	if err := k.Load(posflag.Provider(f, ".", k), nil); err != nil {
		return nil, err
	}

	data, err := k.Marshal(kjson.Parser())
	if err != nil {
		return nil, err
	}

	c := Config{}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
