package main

import (
	"bufio"
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/dahoe/starbattlego-solver/src/config"
	"github.com/dahoe/starbattlego-solver/src/overlay"
	"github.com/dahoe/starbattlego-solver/src/solver"
	"github.com/dahoe/starbattlego-solver/src/vision"
	"github.com/kbinani/screenshot"
)

func promptYesNo(reader *bufio.Reader, prompt string, defaultYes bool) bool {
	for {
		fmt.Print(prompt)
		line, err := reader.ReadString('\n')
		if err != nil {
			return defaultYes
		}
		answer := strings.ToLower(strings.TrimSpace(line))
		if answer == "" {
			return defaultYes
		}
		if answer == "y" || answer == "yes" {
			return true
		}
		if answer == "n" || answer == "no" {
			return false
		}
		fmt.Println("Please answer y or n.")
	}
}

func promptInt(reader *bufio.Reader, prompt string, defaultValue int, min int, max int) int {
	for {
		fmt.Print(prompt)
		line, err := reader.ReadString('\n')
		if err != nil {
			return defaultValue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return defaultValue
		}
		v, err := strconv.Atoi(line)
		if err != nil || v < min || v > max {
			fmt.Printf("Enter a number between %d and %d.\n", min, max)
			continue
		}
		return v
	}
}

func main() {
	cfg := config.Load()
	vision.SetDebugOutput(cfg.Debug, cfg.DebugImagesDir)
	if cfg.Debug {
		fmt.Printf("Debug output enabled. Writing images to %s\n", cfg.DebugImagesDir)
	} else {
		fmt.Println("Debug output disabled. Set DEBUG=true in .env/.env.example or run with -debug.")
	}

	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		log.Fatal("no active displays found")
	}
	mainDisplay := screenshot.GetDisplayBounds(0)

	// Use overlay to select area. Pass empty stars to trigger selection mode.
	fmt.Println("Please draw a box around the puzzle grid (in the overlay window)...")
	overlay.SelectAreaWithCallback(mainDisplay.Dx(), mainDisplay.Dy(), func(selectedArea image.Rectangle) {
		fmt.Printf("Area selected: %v\n", selectedArea)

		// Capture screen
		fmt.Println("Capturing screen...")
		fullImg, err := vision.CaptureScreen()
		if err != nil {
			log.Fatalf("Failed to capture screen: %v", err)
		}
		if err := vision.SaveDebugImage(fullImg, "debug_full.png"); err != nil {
			log.Printf("Failed to save debug_full.png: %v", err)
		} else if cfg.Debug {
			fmt.Println("Saved debug_full.png for inspection")
		}

		// Crop to selection
		croppedImg := image.NewRGBA(image.Rect(0, 0, selectedArea.Dx(), selectedArea.Dy()))
		for y := 0; y < selectedArea.Dy(); y++ {
			for x := 0; x < selectedArea.Dx(); x++ {
				croppedImg.Set(x, y, fullImg.At(selectedArea.Min.X+x, selectedArea.Min.Y+y))
			}
		}

		fmt.Println("Parsing grid in selected area...")
		debugImg, grid, bounds, err := vision.ParseGrid(croppedImg)
		if err != nil {
			if debugImg != nil {
				_ = vision.SaveDebugImage(debugImg, "debug_grid.png")
			}
			log.Fatalf("Failed to parse grid: %v", err)
		}
		_ = vision.SaveDebugImage(debugImg, "debug_grid.png")

		// Adjust bounds back to global screen coordinates
		bounds.Min.X += selectedArea.Min.X
		bounds.Max.X += selectedArea.Min.X
		bounds.Min.Y += selectedArea.Min.Y
		bounds.Max.Y += selectedArea.Min.Y

		fmt.Printf("Grid found at %v, size %dx%d\n", bounds, grid.Size, grid.Size)

		reader := bufio.NewReader(os.Stdin)
		overlay.SetPromptMode(true)
		defer overlay.SetPromptMode(false)

		confirmed := promptYesNo(reader, "Is detected grid correct? [Y/n]: ", true)
		if !confirmed {
			fmt.Println("Grid rejected. Re-run and select puzzle again.")
			return
		}

		starsPerUnit := promptInt(reader, "How many stars per row/column? [2]: ", 2, 1, 9)
		grid.Stars = starsPerUnit
		fmt.Printf("Using %d stars per row/column/region\n", grid.Stars)

		fmt.Println("Solving...")
		solution, err := solver.Solve(grid)
		if err != nil {
			log.Fatalf("Failed to solve: %v", err)
		}

		// Map solution to global screen coordinates
		var stars []overlay.StarPos
		cellW := float32(bounds.Dx()) / float32(grid.Size)
		cellH := float32(bounds.Dy()) / float32(grid.Size)

		for r := 0; r < grid.Size; r++ {
			for c := 0; c < grid.Size; c++ {
				if solution[r][c] == solver.Star {
					x := float32(bounds.Min.X) + (float32(c)+0.5)*cellW
					y := float32(bounds.Min.Y) + (float32(r)+0.5)*cellH
					stars = append(stars, overlay.StarPos{X: x, Y: y})
				}
			}
		}

		fmt.Printf("Found %d stars. Press 'Q' to exit.\n", len(stars))
		// Update overlay with stars
		overlay.SetOverlayStars(stars)
	})
}
