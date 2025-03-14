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

type Storage struct {
	filename string       `json:"-"`
	log      *slog.Logger `json:"-"`
	GameData []GameData   `json:"game_data"`
}

type GameData struct {
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

func NewStorage(fpath string, log *slog.Logger) (*Storage, error) {
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
			return &Storage{filename: fpath, log: log}, nil
		}
		return nil, errors.Wrapf(err, "failed to open %s", fpath)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	s := Storage{}
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	s.filename = fpath
	s.log = log
	return &s, nil
}

func (s *Storage) Persist(rom string, flags []uint8) {
	for i, g := range s.GameData {
		if g.Rom == rom {
			copy(s.GameData[i].Flags[:], flags)
			return
		}
	}

	g := GameData{Rom: rom}
	copy(g.Flags[:], flags)

	s.GameData = append(s.GameData, g)

	if err := s.Save(); err != nil {
		s.log.Error("unable to save storage file", "error", err)
	}
}

func (s *Storage) Load(rom string, len uint16) []uint8 {
	len = gmath.Min(len, 16)

	data := make([]uint8, len)
	for _, g := range s.GameData {
		if g.Rom == rom {
			copy(data, g.Flags[:len])
		}
	}

	return data
}

func (s *Storage) Save() error {
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
