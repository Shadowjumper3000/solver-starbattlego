package vision

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/dahoe/starbattlego-solver/src/solver"
	"github.com/kbinani/screenshot"
)

// Vision defines the abstraction for the vision step (selection -> grid).
// Implementations should focus on a single responsibility: produce a
// `solver.Grid` and the bounding rectangle for the detected grid inside the
// provided image.
type Vision interface {
	ProcessSelection(fullImg image.Image, selection image.Rectangle) (solver.Grid, image.Rectangle, error)
}

// NewTemplateVision returns a minimal, well-documented template implementation
// of Vision that is intended as a starting point for implementing more
// advanced processors. It purposefully performs only basic cropping and
// returns a default 10x10 empty grid so callers can focus on wiring and
// SOLID structure first.
func NewTemplateVision() Vision {
	return &templateVision{}
}

type templateVision struct{}

func (v *templateVision) ProcessSelection(fullImg image.Image, selection image.Rectangle) (solver.Grid, image.Rectangle, error) {
	if fullImg == nil {
		return solver.Grid{}, image.Rectangle{}, fmt.Errorf("nil image")
	}
	if selection.Empty() {
		return solver.Grid{}, image.Rectangle{}, fmt.Errorf("empty selection")
	}

	// Crop to selection (caller-provided coordinates must match fullImg)
	cropped := image.NewRGBA(image.Rect(0, 0, selection.Dx(), selection.Dy()))
	draw.Draw(cropped, cropped.Bounds(), fullImg, selection.Min, draw.Src)

	// This template does not attempt computer vision. It returns a simple
	// default grid so that the rest of the application (solver, overlay)
	// can be developed and tested. Replace this with a real implementation
	// when ready.
	n := 10
	regions := make([][]int, n)
	for r := 0; r < n; r++ {
		regions[r] = make([]int, n)
		for c := 0; c < n; c++ {
			regions[r][c] = 0
		}
	}

	grid := solver.Grid{
		Size:    n,
		Regions: regions,
		Stars:   2,
	}

	// Return bounds relative to fullImg: here we simply echo the selection.
	return grid, selection, nil
}

// CaptureScreen is a convenience wrapper for taking a screenshot of the
// primary display. It is intentionally small so callers can keep orchestration
// simple; more advanced capture strategies can be provided via an interface
// or separate package.
func CaptureScreen() (image.Image, error) {
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		return nil, fmt.Errorf("no active displays found")
	}
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// TemplateProcessSelection is a minimal, SOLID-friendly template showing how
// the vision package can accept a pre-selected screen region and return a
// `solver.Grid` plus the grid bounds in screen-local coordinates. This is
// intended as a clear, tiny example of the data flow (selection -> vision -> solver).
func TemplateProcessSelection(fullImg image.Image, selection image.Rectangle) (solver.Grid, image.Rectangle, error) {
	// Crop the provided full image to the selection rectangle. The caller
	// is responsible for providing coordinates in the same image space.
	if fullImg == nil {
		return solver.Grid{}, image.Rectangle{}, fmt.Errorf("nil image provided")
	}

	if selection.Empty() {
		return solver.Grid{}, image.Rectangle{}, fmt.Errorf("empty selection")
	}

	cropped := image.NewRGBA(image.Rect(0, 0, selection.Dx(), selection.Dy()))
	draw.Draw(cropped, cropped.Bounds(), fullImg, selection.Min, draw.Src)

	// For now reuse the existing ParseGrid implementation to detect grid
	// properties from the cropped region. In a pure-template usage this
	// could be replaced with a lightweight stub or a more explicit
	// interface-based implementation.
	_, grid, bounds, err := ParseGrid(cropped)
	if err != nil {
		return grid, bounds, err
	}

	// Translate bounds back into the coordinate space of the original
	// full image by adding selection.Min offsets.
	bounds.Min.X += selection.Min.X
	bounds.Max.X += selection.Min.X
	bounds.Min.Y += selection.Min.Y
	bounds.Max.Y += selection.Min.Y

	return grid, bounds, nil
}
