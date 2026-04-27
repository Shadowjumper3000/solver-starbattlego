package solver

import "fmt"

const (
	Empty = 0
	Star  = 1
)

type Point struct {
	R, C int
}

type Grid struct {
	Size    int
	Regions [][]int
	Stars   int
}

type State struct {
	Cells        [][]int
	RowCounts    []int
	ColCounts    []int
	RegionCounts map[int]int
}

func NewState(g Grid) *State {
	cells := make([][]int, g.Size)
	for i := range cells {
		cells[i] = make([]int, g.Size)
	}

	return &State{
		Cells:        cells,
		RowCounts:    make([]int, g.Size),
		ColCounts:    make([]int, g.Size),
		RegionCounts: make(map[int]int),
	}
}

func Solve(g Grid) ([][]int, error) {
	s := NewState(g)

	if solve(g, s, 0, 0) {
		return s.Cells, nil
	}
	return nil, fmt.Errorf("no solution found")
}

func solve(g Grid, s *State, r, c int) bool {
	if r == g.Size {
		return true
	}

	nr, nc := r, c+1
	if nc == g.Size {
		nr, nc = r+1, 0
	}

	if canPlace(g, s, r, c) {
		place(g, s, r, c)
		if solve(g, s, nr, nc) {
			return true
		}
		unplace(g, s, r, c)
	}

	if canSkip(g, s, r, c) {
		if solve(g, s, nr, nc) {
			return true
		}
	}

	return false
}

func canPlace(g Grid, s *State, r, c int) bool {
	if s.RowCounts[r] >= g.Stars {
		return false
	}
	if s.ColCounts[c] >= g.Stars {
		return false
	}

	region := g.Regions[r][c]
	if s.RegionCounts[region] >= g.Stars {
		return false
	}

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

func place(g Grid, s *State, r, c int) {
	s.Cells[r][c] = Star
	s.RowCounts[r]++
	s.ColCounts[c]++
	region := g.Regions[r][c]
	s.RegionCounts[region]++
}

func unplace(g Grid, s *State, r, c int) {
	s.Cells[r][c] = Empty
	s.RowCounts[r]--
	s.ColCounts[c]--
	region := g.Regions[r][c]
	s.RegionCounts[region]--
}

func canSkip(g Grid, s *State, r, c int) bool {
	// Row check
	remainingRow := g.Size - c - 1
	if s.RowCounts[r]+remainingRow < g.Stars {
		return false
	}

	// Column check
	remainingCol := g.Size - r - 1
	if s.ColCounts[c]+remainingCol < g.Stars {
		return false
	}

	// Region check (weak but cheap)
	region := g.Regions[r][c]
	if s.RegionCounts[region] > g.Stars {
		return false
	}

	return true
}