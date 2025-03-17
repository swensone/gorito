package emulator

import (
	"fmt"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/swensone/gorito/types"
)

func TestDrawSprite(t *testing.T) {
	tests := []struct {
		name    string
		hires   bool
		mode    types.Mode
		sprite1 uint16
		sprite2 uint16
		plane   uint8
		x       uint8
		y       uint8
		height  uint8
		vf      uint8
	}{
		{
			"no overlap",
			false,
			types.MODE_CHIP8,
			0x50,
			0x5F,
			1,
			5,
			0,
			5,
			0x00,
		},
		{
			"overlap",
			false,
			types.MODE_CHIP8,
			0x50,
			0x5F,
			1,
			0,
			0,
			5,
			0x01,
		},
		{
			"lowres font, hires screen",
			true,
			types.MODE_SUPERCHIP,
			0x50,
			0x5F,
			1,
			5,
			0,
			5,
			0x00,
		},
		{
			"overlap lowres font, hires screen",
			true,
			types.MODE_SUPERCHIP,
			0x50,
			0x5F,
			1,
			0,
			0,
			5,
			0x01,
		},
		{
			"large sprite no overlap",
			true,
			types.MODE_SUPERCHIP,
			0x100,
			0x10A,
			1,
			10,
			0,
			10,
			0x00,
		},
		{
			"large sprite overlap",
			true,
			types.MODE_SUPERCHIP,
			0x100,
			0x10A,
			1,
			0,
			0,
			10,
			0x01,
		},
		{
			"large sprite second plane no overlap",
			true,
			types.MODE_XOCHIP,
			0x100,
			0x10A,
			2,
			0,
			0,
			10,
			0x00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			e := &Emulator{
				xres:      XRES,
				yres:      YRES,
				gfx:       make(map[int][]uint8),
				plane:     1,
				display:   newTermDisplay(XRES, YRES),
				registers: make([]uint8, 16),
			}
			e.gfx[0] = make([]uint8, XRES*YRES)
			e.gfx[1] = make([]uint8, XRES*YRES)
			e.cfg.ColorMap = make(map[uint8]types.Color)
			e.cfg.ColorMap[0] = types.Color{R: 0}
			e.cfg.ColorMap[1] = types.Color{R: 1}
			e.cfg.ColorMap[2] = types.Color{R: 2}
			e.cfg.ColorMap[3] = types.Color{R: 3}
			e.cfg.Mode = tt.mode
			e.Reset()
			if tt.hires {
				e.enableHiRes()
			}

			// draw a 0 at 0,0
			e.idx = tt.sprite1
			e.registers[0] = 0
			e.registers[1] = 0
			fmt.Printf("draw sprite 0x%03x at 0,0\n", tt.sprite1)
			e.drawSprite(0, 1, tt.height)
			e.display.Draw(e.getGfx())

			// draw a 3 at specific coordinates and check for overlap
			e.idx = tt.sprite2
			e.registers[0] = tt.x
			e.registers[1] = tt.y
			fmt.Printf("draw sprite 0x%03x at %d,%d\n", tt.sprite2, tt.x, tt.y)
			e.selectPlane(tt.plane)
			e.drawSprite(0, 1, tt.height)
			e.display.Draw(e.getGfx())

			assert.Equal(t, e.registers[0x0f], tt.vf)
			assert.Equal(t, true, e.drawFlag)
		})
	}
}

type termDisplay struct {
	xres int32
	yres int32
}

func newTermDisplay(xres, yres int32) *termDisplay {
	return &termDisplay{xres: xres, yres: yres}
}

func (t *termDisplay) Draw(gfx []types.Color) error {
	for range t.xres + 2 {
		fmt.Print("*")
	}
	fmt.Println()
	for y := range t.yres {
		fmt.Print("*")
		for x := range t.xres {
			if gfx[y*t.xres+x].R != 0 {
				fmt.Printf("%d", gfx[y*t.xres+x].R)
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
