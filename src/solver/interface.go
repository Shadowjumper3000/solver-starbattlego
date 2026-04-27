package solver

// Solver defines the abstraction for a solver implementation. The existing
// package-level Solve function is wrapped by DefaultSolver to satisfy this
// interface and make it easy to swap in alternative implementations for
// testing or extension.
type Solver interface {
	Solve(g Grid) ([][]int, error)
}

// DefaultSolver delegates to the package-level Solve function.
type DefaultSolver struct{}

func (DefaultSolver) Solve(g Grid) ([][]int, error) {
	return Solve(g)
}
