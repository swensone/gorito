package emulator

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/swensone/gorito/gmath"
)

type storage struct {
	filename string       `json:"-"`
	log      *slog.Logger `json:"-"`
	GameData []gameData   `json:"game_data"`
}

type gameData struct {
	Rom   string    `json:"rom"`
	Flags [16]uint8 `json:"flags"`
}

// remove the path and extension from the rom filename, leaving just (hopefully) the name of the game
func RomName(rompath string) string {
	rompath = filepath.Base(rompath)
	romext := filepath.Ext(rompath)
	rompath = strings.TrimSuffix(rompath, romext)
	return rompath
}

func newStorage(fpath string, log *slog.Logger) (*storage, error) {
	if fpath == "" {
		fpath = "~/.config/gorito-saves.json"
	}

	fpath = path.Clean(fpath)
	if strings.HasPrefix(fpath, "~/") {
		home, _ := os.UserHomeDir()
		fpath = filepath.Join(home, fpath[2:])
	}

	f, err := os.Open(fpath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &storage{filename: fpath, log: log}, nil
		}
		return nil, errors.Wrapf(err, "failed to open %s", fpath)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	s := storage{}
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	s.filename = fpath
	s.log = log
	return &s, nil
}

func (s *storage) Persist(rom string, flags []uint8) {
	for i, g := range s.GameData {
		if g.Rom == rom {
			copy(s.GameData[i].Flags[:], flags)
			return
		}
	}

	g := gameData{Rom: rom}
	copy(g.Flags[:], flags)

	s.GameData = append(s.GameData, g)

	if err := s.Save(); err != nil {
		s.log.Error("unable to save storage file", "error", err)
	}
}

func (s *storage) Load(rom string, len int, registers []uint8) {
	len = gmath.Min(len+1, 16)

	data := make([]uint8, len)
	for _, g := range s.GameData {
		if g.Rom == rom {
			copy(data, g.Flags[:len])
		}
	}
	copy(registers, data)
}

func (s *storage) Save() error {
	// if there's no data to save, don't save a file
	if s == nil || len(s.GameData) == 0 {
		return nil
	}

	data, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return nil
	}

	f, err := os.Create(s.filename)
	if err != nil {
		return err
	}

	if _, err := f.WriteString(string(data)); err != nil {
		return err
	}

	return nil
}
