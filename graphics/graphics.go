package graphics

import (
	"github.com/fstanis/screenresolution"
	"github.com/hashicorp/go-multierror"
	"github.com/veandco/go-sdl2/sdl"

	"github.com/swensone/gorito/emulator"
	"github.com/swensone/gorito/util"
)

type Graphics struct {
	window   *sdl.Window
	renderer *sdl.Renderer

	screenWidth  int32
	screenHeight int32
	windowWidth  int32
	windowHeight int32
	pixelSize    int32
	xOffset      int32
	yOffset      int32
	bgColor      *RGB
	fgColor      *RGB
}

func New(name string, windowwidth, windowheight int32, fullscreen bool, mode emulator.Mode, bg *RGB, fg *RGB) (*Graphics, error) {
	flags := uint32(sdl.WINDOW_SHOWN)
	if fullscreen {
		res := screenresolution.GetPrimary()
		windowwidth = int32(res.Width)
		windowheight = int32(res.Height)

		flags = flags | sdl.WINDOW_FULLSCREEN
	}

	window, err := sdl.CreateWindow(name, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, windowwidth, windowheight, flags)
	if err != nil {
		return nil, err
	}
	w, h := window.GetSize()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, err
	}
	renderer.Clear()

	// determine the pixel size based on the max square that will fit in both
	// screen directions, and center it in both directions
	screenWidth, screenHeight := emulator.GetRes(mode)
	pixelSize := util.Min(w/screenWidth, h/screenHeight)
	xoffset := (w - pixelSize*screenWidth) / 2
	yoffset := (h - pixelSize*screenHeight) / 2

	g := &Graphics{
		window:       window,
		renderer:     renderer,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		windowWidth:  w,
		windowHeight: h,
		bgColor:      bg,
		fgColor:      fg,
		pixelSize:    pixelSize,
		xOffset:      xoffset,
		yOffset:      yoffset,
	}

	return g, nil
}

func (g *Graphics) Close() error {
	var merr *multierror.Error
	if err := g.window.Destroy(); err != nil {
		merr = multierror.Append(merr, err)
	}
	if err := g.renderer.Destroy(); err != nil {
		merr = multierror.Append(merr, err)
	}
	return merr.ErrorOrNil()
}

func (g *Graphics) Draw(gfx []bool) error {
	// clear the screen with the background color
	if err := g.renderer.SetDrawColor(g.bgColor.R, g.bgColor.G, g.bgColor.B, 255); err != nil {
		return err
	}

	if err := g.renderer.Clear(); err != nil {
		return err
	}

	// draw each pixel
	if err := g.renderer.SetDrawColor(g.fgColor.R, g.fgColor.G, g.fgColor.B, 255); err != nil {
		return err
	}
	idx := 0
	for y := range g.screenHeight {
		for x := range g.screenWidth {
			if gfx[idx] {
				if err := g.renderer.FillRect(&sdl.Rect{
					X: g.xOffset + x*g.pixelSize,
					Y: g.yOffset + y*g.pixelSize,
					W: g.pixelSize,
					H: g.pixelSize,
				}); err != nil {
					return err
				}
			}
			idx++
		}
	}

	// and show the screen
	g.renderer.Present()
	return nil
}
