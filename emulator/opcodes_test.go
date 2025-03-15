package emulator

import (
	"fmt"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestDrawSprite(t *testing.T) {
	c := &CPU{
		xres:      128,
		yres:      64,
		gfxp1:     make([]uint8, 128*64),
		gfxp2:     make([]uint8, 128*64),
		plane:     1,
		hires:     false,
		display:   newTermDisplay(128, 64),
		registers: make([]uint8, 16),
	}
	c.Reset()

	tests := []struct {
		name    string
		hires   bool
		sprite1 uint16
		sprite2 uint16
		x       uint8
		y       uint8
		height  uint8
		vf      uint8
	}{
		{
			"no overlap",
			false,
			0x50,
			0x5F,
			5,
			0,
			5,
			0x00,
		},
		{
			"overlap",
			false,
			0x50,
			0x5F,
			0,
			0,
			5,
			0x01,
		},
		{
			"lowres font, hires screen",
			true,
			0x50,
			0x5F,
			5,
			0,
			5,
			0x00,
		},
		{
			"overlap lowres font, hires screen",
			true,
			0x50,
			0x5F,
			0,
			0,
			5,
			0x01,
		},
		{
			"large sprite no overlap",
			true,
			0x100,
			0x10A,
			10,
			0,
			10,
			0x00,
		},
		{
			"large sprite no overlap",
			true,
			0x100,
			0x10A,
			0,
			0,
			10,
			0x01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Reset()
			if tt.hires {
				c.enableHiRes()
			}

			// draw a 0 at 0,0
			c.idx = tt.sprite1
			c.registers[0] = 0
			c.registers[1] = 0
			fmt.Printf("draw sprite 0x%03x at 0,0\n", tt.sprite1)
			c.drawSprite(0, 1, tt.height)
			c.display.Draw(c.gfxp1)

			// draw a 3 at specific coordinates and check for overlap
			c.idx = tt.sprite2
			c.registers[0] = tt.x
			c.registers[1] = tt.y
			fmt.Printf("draw sprite 0x%03x at %d,%d\n", tt.sprite2, tt.x, tt.y)
			c.drawSprite(0, 1, tt.height)
			c.display.Draw(c.gfxp1)

			assert.Equal(t, c.registers[0x0f], tt.vf)
			assert.Equal(t, true, c.drawFlag)
		})
	}
}

type termDisplay struct {
	xres int
	yres int
}

func newTermDisplay(xres, yres int) *termDisplay {
	return &termDisplay{xres: xres, yres: yres}
}

func (t *termDisplay) Draw(gfx []uint8) error {
	for range t.xres + 2 {
		fmt.Print("*")
	}
	fmt.Println()
	for y := range t.yres {
		fmt.Print("*")
		for x := range t.xres {
			if gfx[y*t.xres+x] > 0 {
				fmt.Printf("%d", gfx[y*t.xres+x])
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println("*")
	}
	for range t.xres + 2 {
		fmt.Print("*")
	}
	fmt.Println()
	return nil
}
