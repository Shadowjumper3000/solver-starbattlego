package solver

import (
	"fmt"
)

const (
	Empty = 0
	Star  = 1
)

type Point struct {
	R, C int
}

type Grid struct {
	Size    int
	Regions [][]int // Map each cell (r, c) to a region ID
	Stars   int     // Stars per row/col/region (usually 2)
}

type State struct {
	Cells [][]int
}

func NewState(size int) *State {
	cells := make([][]int, size)
	for i := range cells {
		cells[i] = make([]int, size)
	}
	return &State{Cells: cells}
}

func Solve(g Grid) ([][]int, error) {
	state := NewState(g.Size)
	if solveRecursive(g, state, 0, 0, 0) {
		return state.Cells, nil
	}
	return nil, fmt.Errorf("no solution found")
}

func solveRecursive(g Grid, s *State, row, col, starsPlaced int) bool {
	if starsPlaced == g.Size*g.Stars {
		return true
	}

	// Move to next cell
	nextRow, nextCol := row, col+1
	if nextCol == g.Size {
		nextRow, nextCol = row+1, 0
	}

	// Try placing a star
	if canPlaceStar(g, s, row, col) {
		s.Cells[row][col] = Star
		if solveRecursive(g, s, nextRow, nextCol, starsPlaced+1) {
			return true
		}
		s.Cells[row][col] = Empty
	}

	// Try not placing a star
	// Optimization: check if it's still possible to reach required stars
	if canSkipCell(g, s, row, col) {
		if solveRecursive(g, s, nextRow, nextCol, starsPlaced) {
			return true
		}
	}

	return false
}

func canPlaceStar(g Grid, s *State, r, c int) bool {
	// Check row
	rowCount := 0
	for i := 0; i < g.Size; i++ {
		if s.Cells[r][i] == Star {
			rowCount++
		}
	}
	if rowCount >= g.Stars {
		return false
	}

	// Check col
	colCount := 0
	for i := 0; i < g.Size; i++ {
		if s.Cells[i][c] == Star {
			colCount++
		}
	}
	if colCount >= g.Stars {
		return false
	}

	// Check region
	regionID := g.Regions[r][c]
	regionCount := 0
	for i := 0; i < g.Size; i++ {
		for j := 0; j < g.Size; j++ {
			if g.Regions[i][j] == regionID && s.Cells[i][j] == Star {
				regionCount++
			}
		}
	}
	if regionCount >= g.Stars {
		return false
	}

	// Check adjacency (8 neighbors)
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			nr, nc := r+dr, c+dc
			if nr >= 0 && nr < g.Size && nc >= 0 && nc < g.Size {
				if s.Cells[nr][nc] == Star {
					return false
				}
			}
		}
	}

	return true
}

func canSkipCell(g Grid, s *State, r, c int) bool {
	// If skipping this cell makes it impossible to fill the row/col/region, return false

	// Check row remaining capacity
	rowStars := 0
	for i := 0; i < c; i++ {
		if s.Cells[r][i] == Star {
			rowStars++
		}
	}
	remainingInRow := g.Size - 1 - c
	if rowStars+remainingInRow < g.Stars {
		return false
	}

	// Simplified check: we can always skip if there's enough space left in the row.
	// For more efficiency, we should also check columns and regions.
	return true
}
