package emulator

import (
	"strings"

	"github.com/cockroachdb/errors"
)

type Mode int

const (
	MODE_CHIP8 = iota
	MODE_SUPERCHIP
	MODE_XOCHIP
)

var modemap = map[Mode]string{
	MODE_CHIP8:     "chip-8",
	MODE_SUPERCHIP: "superchip",
	MODE_XOCHIP:    "xo-chip",
}

func ModeFromString(s string) (Mode, error) {
	for mode, modestr := range modemap {
		if modestr == s {
			return mode, nil
		}
	}
	return 0, errors.Errorf("unknown mode: %s", s)
}

func (m *Mode) String() string {
	return modemap[*m]
}

func SupportedModes() []string {
	var modes []string
	for _, m := range modemap {
		modes = append(modes, m)
	}
	return modes
}

func (m *Mode) UnmarshalJSON(data []byte) error {
	sdata := string(data)
	// ignore null
	if sdata == "null" || sdata == `""` {
		return nil
	}

	if !strings.HasPrefix(sdata, "\"") || !strings.HasSuffix(sdata, "\"") {
		return errors.New("data must be formatted as a quoted string")
	}

	ms, err := ModeFromString(sdata[1 : len(sdata)-1])
	if err != nil {
		return err
	}

	*m = ms
	return nil
}

func (m *Mode) MarshalJSON() ([]byte, error) {
	return []byte("\"" + m.String() + "\""), nil
}

func GetRes(m Mode) (int32, int32) {
	return 128, 64
}
