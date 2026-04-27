package overlay

import "image"

// Overlayer describes essential overlay interactions used by the orchestrator
// code: selecting an area, updating displayed stars, and toggling prompt mode.
// This lets the main orchestration depend on an interface for testability.
type Overlayer interface {
	SelectAreaWithCallback(screenWidth, screenHeight int, callback func(image.Rectangle))
	SetOverlayStars(stars []StarPos)
	SetPromptMode(enabled bool)
}

// DefaultOverlay is a thin adapter over the package-level functions so the
// rest of the code can depend on the Overlayer interface.
type DefaultOverlay struct{}

func (DefaultOverlay) SelectAreaWithCallback(screenWidth, screenHeight int, callback func(image.Rectangle)) {
	SelectAreaWithCallback(screenWidth, screenHeight, callback)
}

func (DefaultOverlay) SetOverlayStars(stars []StarPos) {
	SetOverlayStars(stars)
}

func (DefaultOverlay) SetPromptMode(enabled bool) {
	SetPromptMode(enabled)
}
