package main

import (
	"fmt"
	"image/color"
	"math/cmplx"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func main() {
	pixelgl.Run(func() {
		start()
	})
}

var (
	windowBounds     = pixel.R(0, 0, 1080, 1080)
	mandelbrotBounds = pixel.R(-2, -2, 2, 2)
	pixelData        = pixel.MakePictureData(windowBounds)

	mandelbrotSprite *pixel.Sprite
	mandelbrotMu     sync.RWMutex

	colourWhite = color.RGBA{0, 0, 0, 0}
)

const (
	iterations = 200
	contrast   = 20
)

func start() {
	// create window config
	cfg := pixelgl.WindowConfig{
		Title:     "Mandelbrot",
		Bounds:    windowBounds,
		VSync:     false,
		Resizable: true,
	}

	// create window
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		fmt.Printf("failed create new window: %s\n", err)
		return
	}

	// generate initial render
	generate()
	go func() {
		for {
			generate()
		}
	}()

	initialSize := mandelbrotBounds.Size()
	// limit update cycles to 30 FPS
	frameRateLimiter := time.Tick(time.Second / 30)

	// main game loop
	for !win.Closed() {
		scaleFactor := initialSize.ScaledXY(mandelbrotBounds.Size()).Scaled(0.001)

		// handle keyboard input
		switch {
		case win.JustPressed(pixelgl.KeyEscape):
			return

		case win.Pressed(pixelgl.KeyR):
			mandelbrotBounds = mandelbrotBounds.Resized(mandelbrotBounds.Center(), mandelbrotBounds.Size().Scaled(0.97))

		case win.Pressed(pixelgl.KeyF):
			mandelbrotBounds = mandelbrotBounds.Resized(mandelbrotBounds.Center(), mandelbrotBounds.Size().Scaled(1.03))

		case win.Pressed(pixelgl.KeyA):
			mandelbrotBounds = mandelbrotBounds.Moved(pixel.V(-scaleFactor.X, 0))

		case win.Pressed(pixelgl.KeyD):
			mandelbrotBounds = mandelbrotBounds.Moved(pixel.V(scaleFactor.X, 0))

		case win.Pressed(pixelgl.KeyS):
			mandelbrotBounds = mandelbrotBounds.Moved(pixel.V(0, -scaleFactor.Y))

		case win.Pressed(pixelgl.KeyW):
			mandelbrotBounds = mandelbrotBounds.Moved(pixel.V(0, scaleFactor.Y))
		}

		// draw
		win.Clear(colourWhite)

		mandelbrotMu.RLock()
		mandelbrotSprite.Draw(win, pixel.IM.Moved(win.Bounds().Size().Scaled(0.5)))
		mandelbrotMu.RUnlock()

		win.Update()

		<-frameRateLimiter
	}
}

func generate() {
	height := windowBounds.H()
	width := windowBounds.W()

	for py := 0.0; py < height; py++ {
		y := py/height*(mandelbrotBounds.Max.Y-mandelbrotBounds.Min.Y) + mandelbrotBounds.Min.Y

		for px := 0.0; px < width; px++ {
			x := px/width*(mandelbrotBounds.Max.X-mandelbrotBounds.Min.X) + mandelbrotBounds.Min.X
			z := complex(x, y)

			i := pixelData.Index(pixel.V(px, py))
			pixelData.Pix[i] = processPixel(z)
		}
	}

	newSprite := pixel.NewSprite(pixelData, pixelData.Bounds())
	mandelbrotMu.Lock()
	mandelbrotSprite = newSprite
	mandelbrotMu.Unlock()
}

func processPixel(c complex128) color.RGBA {
	var z complex128

	for n := uint8(0); n < iterations; n++ {
		z = z*z + c

		if cmplx.Abs(z) > 16 {
			return color.RGBA{
				R: 60 - contrast*n,
				G: 180 - contrast*n,
				B: contrast * n,
				A: 255,
			}
		}
	}
	return colourWhite
}