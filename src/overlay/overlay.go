package overlay

import (
	"image"
	"image/color"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type StarPos struct {
	X, Y float32
}

var globalOverlay *Overlay
var overlayPromptMode atomic.Bool

type Overlay struct {
	Stars               []StarPos
	Size                image.Point
	Selecting           bool
	WindowHidden        bool
	SelectStart         image.Point
	SelectEnd           image.Point
	SelectionResult     image.Rectangle
	SelectionDone       chan image.Rectangle
	SelectionSwitchTime time.Time
	SelectionCallback   func(image.Rectangle)
}

func (o *Overlay) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		os.Exit(0)
	}

	if overlayPromptMode.Load() {
		if !o.WindowHidden {
			ebiten.SetWindowSize(1, 1)
			ebiten.SetWindowPosition(-32000, -32000)
			o.WindowHidden = true
		}
		return nil
	}

	if o.WindowHidden {
		ebiten.SetWindowPosition(0, 0)
		ebiten.SetWindowSize(o.Size.X, o.Size.Y)
		ebiten.SetWindowMousePassthrough(true)
		o.WindowHidden = false
	}

	if o.Selecting {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			o.SelectStart = image.Pt(x, y)
			o.SelectEnd = o.SelectStart
		} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			o.SelectEnd = image.Pt(x, y)
		} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			rect := image.Rectangle{Min: o.SelectStart, Max: o.SelectEnd}.Canon()
			log.Printf("Selection confirmed: %v\n", rect)
			if rect.Dx() > 10 && rect.Dy() > 10 {
				o.Selecting = false
				o.SelectionResult = rect
				o.SelectionSwitchTime = time.Now().Add(100 * time.Millisecond)

				// Enable mouse passthrough for star display
				ebiten.SetWindowMousePassthrough(true)

				// Call the selection callback to process and display stars
				if o.SelectionCallback != nil {
					go o.SelectionCallback(rect)
				}

				if o.SelectionDone != nil {
					select {
					case o.SelectionDone <- rect:
					default:
					}
				}
			}
		}
	}

	return nil
}

func (o *Overlay) Draw(screen *ebiten.Image) {
	// Clear previous frame so selection dimming does not persist after selection.
	screen.Clear()
	if o.WindowHidden {
		return
	}

	if o.Selecting {
		// Draw selection box
		rect := image.Rectangle{Min: o.SelectStart, Max: o.SelectEnd}.Canon()
		vector.StrokeRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), 2, color.RGBA{0, 255, 0, 255}, true)
		// Dim the rest of the screen
		vector.DrawFilledRect(screen, 0, 0, float32(o.Size.X), float32(rect.Min.Y), color.RGBA{0, 0, 0, 100}, true)
		vector.DrawFilledRect(screen, 0, float32(rect.Max.Y), float32(o.Size.X), float32(o.Size.Y-rect.Max.Y), color.RGBA{0, 0, 0, 100}, true)
		vector.DrawFilledRect(screen, 0, float32(rect.Min.Y), float32(rect.Min.X), float32(rect.Dy()), color.RGBA{0, 0, 0, 100}, true)
		vector.DrawFilledRect(screen, float32(rect.Max.X), float32(rect.Min.Y), float32(o.Size.X-rect.Max.X), float32(rect.Dy()), color.RGBA{0, 0, 0, 100}, true)
	} else {
		for _, s := range o.Stars {
			vector.DrawFilledCircle(screen, s.X, s.Y, 10, color.RGBA{255, 215, 0, 255}, true)
		}
	}
}

func (o *Overlay) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func Run(stars []StarPos, screenWidth, screenHeight int) {
	ebiten.SetWindowDecorated(false)
	ebiten.SetScreenTransparent(true)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowMousePassthrough(true)
	ebiten.SetWindowSize(screenWidth, screenHeight)

	ov := &Overlay{
		Stars: stars,
		Size:  image.Pt(screenWidth, screenHeight),
	}

	if err := ebiten.RunGame(ov); err != nil {
		log.Fatal(err)
	}
}

// SelectAreaWithCallback opens a selection overlay, calls the callback when selection is done,
// then displays stars. This runs a single Ebiten game loop for the entire workflow.
func SelectAreaWithCallback(screenWidth, screenHeight int, callback func(image.Rectangle)) {
	ebiten.SetWindowDecorated(false)
	ebiten.SetScreenTransparent(true)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowMousePassthrough(false)
	ebiten.SetWindowSize(screenWidth, screenHeight)

	ov := &Overlay{
		Size:              image.Pt(screenWidth, screenHeight),
		Selecting:         true,
		SelectionCallback: callback,
	}
	globalOverlay = ov

	if err := ebiten.RunGame(ov); err != nil {
		log.Fatal(err)
	}
}

// SetOverlayStars updates the global overlay's stars (called from the selection callback)
func SetOverlayStars(stars []StarPos) {
	if globalOverlay != nil {
		globalOverlay.Stars = stars
	}
}

// SetPromptMode hides the overlay window while terminal input is requested.
func SetPromptMode(enabled bool) {
	overlayPromptMode.Store(enabled)
}

// SelectArea opens a selection overlay and returns the selected rectangle.
// IMPORTANT: Only one ebiten game loop can run per application, so this function
// should only be called before Run(). For a unified selection + display workflow,
// use SelectAndDisplay instead.
func SelectArea(screenWidth, screenHeight int) image.Rectangle {
	ebiten.SetWindowDecorated(false)
	ebiten.SetScreenTransparent(true)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowMousePassthrough(false) // We need to catch mouse events for selection
	ebiten.SetWindowSize(screenWidth, screenHeight)

	ov := &Overlay{
		Size:      image.Pt(screenWidth, screenHeight),
		Selecting: true,
	}

	if err := ebiten.RunGame(ov); err != nil {
		log.Fatalf("SelectArea: ebiten run error: %v", err)
	}

	return ov.SelectionResult
}
