package main

import (
	"flag"
	"fmt"
	"image/color"
	"math/cmplx"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var (
	iterations       uint
	windowSize       float64
	windowBounds     pixel.Rect
	mandelbrotBounds = pixel.R(-2, -2, 2, 2)

	pixelData        *pixel.PictureData
	mandelbrotSprite *pixel.Sprite
	// mutex serialises access to the drawable pixel data
	mandelbrotMu sync.RWMutex

	colourBlack = color.RGBA{0, 0, 0, 0}
)

const (
	colourContrast = 20
)

func main() {
	// process flags
	flag.UintVar(&iterations, "iterations", 200, "the number of mandelbrot iterations")
	flag.Float64Var(&windowSize, "size", 720, "the window size")
	flag.Parse()

	fmt.Printf("Generating Mandelbrot for %d iterations at %dx%d\n", iterations, int(windowSize), int(windowSize))

	pixelgl.Run(func() {
		start()
	})
}

func start() {
	windowBounds = pixel.R(0, 0, windowSize, windowSize)

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

	pixelData = pixel.MakePictureData(windowBounds)

	// generate initial mandelbrot and continue to generate a fresh copy independent of the main thread
	generate()
	go func() {
		for {
			generate()
		}
	}()

	// limit update cycles to 30 FPS
	frameRateLimiter := time.Tick(time.Second / 120)
	initialBoundsSize := mandelbrotBounds.Size()

	// main game loop
	for !win.Closed() {
		scaleFactor := initialBoundsSize.ScaledXY(mandelbrotBounds.Size()).Scaled(0.001)

		// handle keyboard input
		if win.JustPressed(pixelgl.KeyEscape) {
			return
		}
		if win.Pressed(pixelgl.KeyR) {
			mandelbrotBounds = mandelbrotBounds.Resized(mandelbrotBounds.Center(), mandelbrotBounds.Size().Scaled(0.997))
		}
		if win.Pressed(pixelgl.KeyF) {
			mandelbrotBounds = mandelbrotBounds.Resized(mandelbrotBounds.Center(), mandelbrotBounds.Size().Scaled(1.003))
		}
		if win.Pressed(pixelgl.KeyA) {
			mandelbrotBounds = mandelbrotBounds.Moved(pixel.V(-scaleFactor.X, 0))
		}
		if win.Pressed(pixelgl.KeyD) {
			mandelbrotBounds = mandelbrotBounds.Moved(pixel.V(scaleFactor.X, 0))
		}
		if win.Pressed(pixelgl.KeyS) {
			mandelbrotBounds = mandelbrotBounds.Moved(pixel.V(0, -scaleFactor.Y))
		}
		if win.Pressed(pixelgl.KeyW) {
			mandelbrotBounds = mandelbrotBounds.Moved(pixel.V(0, scaleFactor.Y))
		}

		// draw window and mandelbrot
		win.Clear(colourBlack)

		mandelbrotMu.RLock()
		tempMandelbrotSprite := mandelbrotSprite
		mandelbrotMu.RUnlock()
		tempMandelbrotSprite.Draw(win, pixel.IM.Moved(win.Bounds().Size().Scaled(0.5)))

		win.Update()

		<-frameRateLimiter
	}
}

// generates a fresh mandelbrot represented in pixel.Sprite form
func generate() {
	for py := 0.0; py < windowSize; py++ {
		y := py/windowSize*(mandelbrotBounds.Max.Y-mandelbrotBounds.Min.Y) + mandelbrotBounds.Min.Y

		for px := 0.0; px < windowSize; px++ {
			x := px/windowSize*(mandelbrotBounds.Max.X-mandelbrotBounds.Min.X) + mandelbrotBounds.Min.X
			z := complex(x, y)

			// set individual pixel image data
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

	for n := uint8(0); n < uint8(iterations); n++ {
		z = z*z + c

		if cmplx.Abs(z) > 16 {
			return color.RGBA{
				R: 60 - colourContrast*n,
				G: 180 - colourContrast*n,
				B: colourContrast * n,
				A: 255,
			}
		}
	}
	return colourBlack
}
