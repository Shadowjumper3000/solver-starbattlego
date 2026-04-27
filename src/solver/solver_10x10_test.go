package solver

import (
	"testing"
)

func TestSolve10x10(t *testing.T) {
	// 10x10 grid from a typical Star Battle (2 stars)
	// Example layout (simplified for testing):
	regions := [][]int{
		{0, 0, 0, 0, 0, 1, 1, 1, 1, 1},
		{0, 0, 0, 0, 0, 1, 1, 1, 1, 1},
		{2, 2, 2, 2, 2, 3, 3, 3, 3, 3},
		{2, 2, 2, 2, 2, 3, 3, 3, 3, 3},
		{4, 4, 4, 4, 4, 5, 5, 5, 5, 5},
		{4, 4, 4, 4, 4, 5, 5, 5, 5, 5},
		{6, 6, 6, 6, 6, 7, 7, 7, 7, 7},
		{6, 6, 6, 6, 6, 7, 7, 7, 7, 7},
		{8, 8, 8, 8, 8, 9, 9, 9, 9, 9},
		{8, 8, 8, 8, 8, 9, 9, 9, 9, 9},
	}
	grid := Grid{
		Size:    10,
		Regions: regions,
		Stars:   2,
	}

	solution, err := Solve(grid)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Verify row and column counts
	for r := 0; r < 10; r++ {
		rowStars := 0
		for c := 0; c < 10; c++ {
			if solution[r][c] == Star {
				rowStars++
			}
		}
		if rowStars != 2 {
			t.Errorf("Row %d has %d stars, expected 2", r, rowStars)
		}
	}
	for c := 0; c < 10; c++ {
		colStars := 0
		for r := 0; r < 10; r++ {
			if solution[r][c] == Star {
				colStars++
			}
		}
		if colStars != 2 {
			t.Errorf("Col %d has %d stars, expected 2", c, colStars)
		}
	}
}
