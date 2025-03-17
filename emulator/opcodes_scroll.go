package emulator

// scrollDown: 00CN: Scroll the display down by 0 to 15 pixels
func (e *Emulator) scrollDown(N uint8) {
	scrollLen := int(e.xres) * int(N)
	scroll := make([]uint8, scrollLen)

	for i := range PLANES {
		if e.plane>>i&0x01 == 0x01 {
			e.gfx[i] = append(scroll, e.gfx[i][0:len(e.gfx[i])-scrollLen]...)
		}
	}
}

// scrollUp: 00DN: Scroll the display up by 0 to 15 pixels
func (e *Emulator) scrollUp(N uint8) {
	scrollLen := int(e.xres) * int(N)
	scroll := make([]uint8, scrollLen)

	for i := range PLANES {
		if e.plane>>i&0x01 == 0x01 {
			e.gfx[i] = append(e.gfx[i][len(e.gfx[i])-scrollLen:], scroll...)
		}
	}
}

// scrollRight: 00FB: Scroll the display right by 4 pixels
func (e *Emulator) scrollRight() {
	for i := range PLANES {
		if e.plane>>i&0x01 == 0x01 {
			newGfx := []uint8{}
			scroll := make([]uint8, 4)
			for j := range e.yres {
				newGfx = append(newGfx, scroll...)
				newGfx = append(newGfx, e.gfx[i][j*e.xres:(j+1)*e.xres-4]...)
			}
			e.gfx[i] = newGfx
		}
	}
}

// scrollLeft: 00FC: Scroll the display left by 4 pixels.
func (e *Emulator) scrollLeft() {
	for i := range PLANES {
		if e.plane>>i&0x01 == 0x01 {
			newGfx := []uint8{}
			scroll := make([]uint8, 4)
			for j := range e.yres {
				newGfx = append(newGfx, e.gfx[i][j*e.xres+4:(j+1)*e.xres]...)
				newGfx = append(newGfx, scroll...)
			}
			e.gfx[i] = newGfx
		}
	}
}
