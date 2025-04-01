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

	"github.com/swensone/gorito/types"
)

type Config struct {
	Savefile   string      `yaml:"savefile,omitempty"`
	Level      slog.Level  `yaml:"level,omitempty"`
	Opcodes    bool        `yaml:"opcodes,omitempty"`
	Mode       types.Mode  `yaml:"mode,omitempty"`
	Speed      uint32      `yaml:"speed,omitempty"`
	ROM        string      `yaml:"rom,omitempty"`
	Width      int32       `yaml:"width,omitempty"`
	Height     int32       `yaml:"height,omitempty"`
	Fullscreen bool        `yaml:"fullscreen,omitempty"`
	BG         types.Color `yaml:"bg,omitempty"`
	FG1        types.Color `yaml:"fg1,omitempty"`
	FG2        types.Color `yaml:"fg2,omitempty"`
	FG3        types.Color `yaml:"fg3,omitempty"`
}

func Parse() (*Config, error) {
	k := koanf.New(".")
	// set up some decent defaults
	k.Load(confmap.Provider(map[string]interface{}{
		"savefile":   "~/.config/gorito-storage.json",
		"config":     "~/.config/gorito.yaml",
		"level":      "INFO",
		"opcodes":    false,
		"mode":       "superchip",
		"speed":      600,
		"width":      1280,
		"height":     640,
		"fullscreen": false,
		"bg":         "080808",
		"fg1":        "1e81b0",
		"fg2":        "eab676",
		"fg3":        "873e23",
	}, "."), nil)

	// Parse command line flags
	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}

	f.StringP("config", "c", "", "path to one or more .yaml config files")
	f.String("savefile", "", "save file location")
	f.StringP("level", "l", "", "log level")
	f.BoolP("opcodes", "o", false, "log opcodes, extremely noisy")
	f.StringP("mode", "m", "", fmt.Sprintf("emulator mode, possible values: %s", strings.Join(types.SupportedModes(), ", ")))
	f.Uint16P("speed", "s", 0, "speed in cycles per seond")
	f.StringP("rom", "r", "", "path to the rom you want to load")
	f.IntP("width", "x", 0, "window width")
	f.IntP("height", "y", 0, "window height")
	f.BoolP("fullscreen", "f", false, "display full screen")
	f.String("bg", "", "background color in hex")
	f.String("fg1", "", "foreground 1 color in hex")
	f.String("fg2", "", "foreground 2 color in hex, only used in xo-chip")
	f.String("fg3", "", "foreground 3 color in hex, only used in xo-chip")
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
