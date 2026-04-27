package solver

import (
	"testing"
)

func TestSolve5x5(t *testing.T) {
	// Simple 5x5 grid, 1 star per unit
	// Regions:
	// 0 0 1 1 1
	// 0 0 1 1 1
	// 2 2 2 3 3
	// 4 4 4 3 3
	// 4 4 4 3 3
	regions := [][]int{
		{0, 0, 1, 1, 1},
		{0, 0, 1, 1, 1},
		{2, 2, 2, 3, 3},
		{4, 4, 4, 3, 3},
		{4, 4, 4, 3, 3},
	}
	grid := Grid{
		Size:    5,
		Regions: regions,
		Stars:   1,
	}

	solution, err := Solve(grid)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Verify solution
	for r := 0; r < 5; r++ {
		rowStars := 0
		for c := 0; c < 5; c++ {
			if solution[r][c] == Star {
				rowStars++
				// Check adjacency
				for dr := -1; dr <= 1; dr++ {
					for dc := -1; dc <= 1; dc++ {
						if dr == 0 && dc == 0 {
							continue
						}
						nr, nc := r+dr, c+dc
						if nr >= 0 && nr < 5 && nc >= 0 && nc < 5 {
							if solution[nr][nc] == Star {
								t.Errorf("Adjacent stars at (%d,%d) and (%d,%d)", r, c, nr, nc)
							}
						}
					}
				}
			}
		}
		if rowStars != 1 {
			t.Errorf("Row %d has %d stars, expected 1", r, rowStars)
		}
	}
}
