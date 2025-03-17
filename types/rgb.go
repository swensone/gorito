package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Color struct {
	R uint8
	G uint8
	B uint8
}

func (c *Color) ParseString(s string) error {
	s = strings.TrimPrefix(s, "0x")

	if len(s) != 6 {
		return fmt.Errorf("invalid color %s, must be 6 hex chars long", s)
	}

	r, err := strconv.ParseInt(s[0:2], 16, 16)
	if err != nil {
		return err
	}

	g, err := strconv.ParseInt(s[2:4], 16, 16)
	if err != nil {
		return err
	}

	b, err := strconv.ParseInt(s[4:6], 16, 16)
	if err != nil {
		return err
	}

	c.R = uint8(r)
	c.G = uint8(g)
	c.B = uint8(b)

	return nil
}

func (c *Color) String() string {
	return fmt.Sprintf("%02X%02X%02X", c.R, c.G, c.B)
}

func (c *Color) UnmarshalJSON(data []byte) error {
	sdata := string(data)
	// ignore null
	if sdata == "null" || sdata == `""` {
		return nil
	}

	if !strings.HasPrefix(sdata, "\"") || !strings.HasSuffix(sdata, "\"") {
		return errors.New("data must be formatted as a quoted string")
	}

	return c.ParseString(sdata[1 : len(sdata)-1])
}

func (c *Color) MarshalJSON() ([]byte, error) {
	return []byte("\"" + c.String() + "\""), nil
}
