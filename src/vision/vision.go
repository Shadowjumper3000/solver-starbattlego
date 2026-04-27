package vision

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/anthonynsimon/bild/effect"
	"github.com/dahoe/starbattlego-solver/src/solver"
	"github.com/kbinani/screenshot"
)

var debugOutputEnabled bool
var debugOutputDir = "debug_images"

func SetDebugOutput(enabled bool, outputDir string) {
	debugOutputEnabled = enabled
	if strings.TrimSpace(outputDir) == "" {
		debugOutputDir = "debug_images"
		return
	}
	debugOutputDir = outputDir
}

func SaveDebugImage(img image.Image, filename string) error {
	if !debugOutputEnabled || img == nil {
		return nil
	}
	if err := os.MkdirAll(debugOutputDir, 0o755); err != nil {
		return err
	}
	name := filepath.Base(strings.TrimSpace(filename))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return fmt.Errorf("invalid debug filename")
	}
	return SaveImage(img, filepath.Join(debugOutputDir, name))
}

func saveDebugImage(img image.Image, filename string) {
	if !debugOutputEnabled {
		return
	}
	if err := SaveDebugImage(img, filename); err != nil {
		fmt.Printf("Vision: failed saving %s: %v\n", filename, err)
		return
	}
	fmt.Printf("Vision: saved %s\n", filename)
}

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

func SaveImage(img image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func ParseGrid(img image.Image) (image.Image, solver.Grid, image.Rectangle, error) {
	// Crop a small margin (2%) to avoid window decorations or selection artifacts
	// at the edges showing up in edge detection.
	b0 := img.Bounds()
	minDim := math.Min(float64(b0.Dx()), float64(b0.Dy()))
	margin := int(minDim * 0.02)
	if margin > 0 {
		cropped := image.NewRGBA(image.Rect(0, 0, b0.Dx()-2*margin, b0.Dy()-2*margin))
		draw.Draw(cropped, cropped.Bounds(), img, image.Pt(b0.Min.X+margin, b0.Min.Y+margin), draw.Src)
		img = cropped
		fmt.Printf("Vision: cropped margin %d -> new size %dx%d\n", margin, cropped.Bounds().Dx(), cropped.Bounds().Dy())
	}

	// Pre-process: Alternative approach - skip Sobel, look for contrast transitions
	// Grid lines are typically darker than cell backgrounds, so find dark-to-light transitions
	grayImg := effect.Grayscale(img)
	saveDebugImage(grayImg, "debug_gray.png")

	// Detect vertical and horizontal contrast edges directly
	b := grayImg.Bounds()

	// Vertical edges: look for horizontal transitions (dark to light or vice versa)
	verticalEdges := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X + 1; x < b.Max.X; x++ {
			left := getGrayComponent(grayImg.At(x-1, y))
			right := getGrayComponent(grayImg.At(x, y))
			diff := int(left) - int(right)
			if diff < 0 {
				diff = -diff
			}
			// Store edge strength
			if uint8(diff) > 255 {
				verticalEdges.SetGray(x, y, color.Gray{Y: 255})
			} else {
				verticalEdges.SetGray(x, y, color.Gray{Y: uint8(diff)})
			}
		}
	}

	// Horizontal edges: look for vertical transitions
	horizontalEdges := image.NewGray(b)
	for y := b.Min.Y + 1; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			top := getGrayComponent(grayImg.At(x, y-1))
			bottom := getGrayComponent(grayImg.At(x, y))
			diff := int(top) - int(bottom)
			if diff < 0 {
				diff = -diff
			}
			if uint8(diff) > 255 {
				horizontalEdges.SetGray(x, y, color.Gray{Y: 255})
			} else {
				horizontalEdges.SetGray(x, y, color.Gray{Y: uint8(diff)})
			}
		}
	}

	// Combine edges
	edges := image.NewGray(b)
	var maxEdgeVal uint8
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			v := verticalEdges.GrayAt(x, y).Y
			h := horizontalEdges.GrayAt(x, y).Y
			if h > v {
				v = h
			}
			edges.SetGray(x, y, color.Gray{Y: v})
			if v > maxEdgeVal {
				maxEdgeVal = v
			}
		}
	}

	saveDebugImage(verticalEdges, "debug_vertical_edges.png")
	saveDebugImage(horizontalEdges, "debug_horizontal_edges.png")
	saveDebugImage(edges, "debug_edges.png")

	fmt.Printf("Vision: Contrast edges max value = %d\n", maxEdgeVal)

	be := edges.Bounds()

	// Threshold: use a reasonable fraction of max edge value
	// Use adaptive thresholding so low-contrast captures still produce lines.
	thresh := int(float64(maxEdgeVal) * 0.40) // 40% of max
	if thresh < 3 {
		thresh = 3
	}
	maxAdaptive := int(float64(maxEdgeVal) * 0.90)
	if maxAdaptive < 3 {
		maxAdaptive = 3
	}
	if thresh > maxAdaptive {
		thresh = maxAdaptive
	}
	fmt.Printf("Vision: Contrast threshold = %d (40%% of max %d, adaptive)\n", thresh, maxEdgeVal)

	binarized := image.NewGray(edges.Bounds())
	for y := be.Min.Y; y < be.Max.Y; y++ {
		for x := be.Min.X; x < be.Max.X; x++ {
			if edges.GrayAt(x, y).Y >= uint8(thresh) {
				binarized.SetGray(x, y, color.Gray{Y: 255})
			} else {
				binarized.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}

	// Grid lines should be sparse in the binarized image. If most pixels are
	// white, the foreground/background polarity is likely inverted.
	bb := binarized.Bounds()
	whitePixels := 0
	blackPixels := 0
	total := bb.Dx() * bb.Dy()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			v := binarized.GrayAt(x, y).Y
			if v > 200 {
				whitePixels++
			} else if v < 50 {
				blackPixels++
			}
		}
	}
	whiteRatio := float64(whitePixels) / float64(total)
	fmt.Printf("Vision: Binarized white pixels~%d, black pixels~%d, ratio=%.3f\n", whitePixels, blackPixels, whiteRatio)

	if whiteRatio > 0.70 {
		// invert
		fmt.Println("Vision: Inverting binarized image (white-dominant image detected)")
		inv := image.NewGray(b)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				inv.SetGray(x, y, color.Gray{Y: 255 - binarized.GrayAt(x, y).Y})
			}
		}
		binarized = inv
		saveDebugImage(binarized, "debug_binarized_inverted.png")
	}

	// Save intermediate debug images to help diagnose failures
	saveDebugImage(binarized, "debug_binarized.png")

	// Keep a separate mask for tile-size detection where thin grid lines are preserved.
	lineMask := buildLineMask(grayImg, edges)
	saveDebugImage(lineMask, "debug_binarized_lines.png")

	// Apply a small morphological closing to remove thin border artifacts
	// that may appear along the selection edge.
	closed := closeBinary(binarized, 1)
	saveDebugImage(closed, "debug_binarized_closed.png")

	// Step 1: Find grid size from thin-line mask, then fall back to closed mask if needed.
	bounds, n, err := findGridBounds(lineMask)
	if err != nil {
		fmt.Printf("Vision: thin-line grid detection failed: %v\n", err)
		bounds, n, err = findGridBounds(closed)
		if err != nil {
			fmt.Printf("Vision: closed-mask grid detection failed: %v\n", err)
			fmt.Println("Vision: Falling back to selection bounds and default 10x10 grid size")
			bounds = closed.Bounds()
			n = 10
		}
	}

	// Step 2: Build region map from the closed mask where puzzle shapes are most stable.
	regions := buildRegions(closed, bounds, n)

	grid := solver.Grid{
		Size:    n,
		Regions: regions,
		Stars:   2, // Default to 2
	}

	return closed, grid, bounds, nil
}

func findGridBounds(img *image.Gray) (image.Rectangle, int, error) {
	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()

	rowDensities := make([]int, height)
	colDensities := make([]int, width)

	maxRowD, maxColD := 0, 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			val := img.GrayAt(b.Min.X+x, b.Min.Y+y).Y
			if val > 128 {
				rowDensities[y]++
				colDensities[x]++
			}
		}
	}

	for _, d := range rowDensities {
		if d > maxRowD {
			maxRowD = d
		}
	}
	for _, d := range colDensities {
		if d > maxColD {
			maxColD = d
		}
	}

	fmt.Printf("Vision: Scanned area %dx%d. Max Row Density: %d, Max Col Density: %d\n", width, height, maxRowD, maxColD)

	if maxRowD == 0 || maxColD == 0 {
		return image.Rectangle{}, 0, fmt.Errorf("no white pixels found in binarized image")
	}

	// Threshold densities to detect candidate line columns/rows
	// Make thresholds much lower to catch even weak lines
	colThr := max(1, maxColD*10/100) // 10% instead of 25%
	rowThr := max(1, maxRowD*10/100)

	fmt.Printf("Vision: Using density thresholds: col=%d, row=%d\n", colThr, rowThr)

	type seg struct{ start, end int }
	var colSegs []seg
	var rowSegs []seg

	// collect column segments
	in := false
	s := 0
	for x := 0; x < width; x++ {
		if colDensities[x] > colThr {
			if !in {
				in = true
				s = x
			}
		} else {
			if in {
				colSegs = append(colSegs, seg{start: s, end: x - 1})
				in = false
			}
		}
	}
	if in {
		colSegs = append(colSegs, seg{start: s, end: width - 1})
	}

	// collect row segments
	in = false
	s = 0
	for y := 0; y < height; y++ {
		if rowDensities[y] > rowThr {
			if !in {
				in = true
				s = y
			}
		} else {
			if in {
				rowSegs = append(rowSegs, seg{start: s, end: y - 1})
				in = false
			}
		}
	}
	if in {
		rowSegs = append(rowSegs, seg{start: s, end: height - 1})
	}

	fmt.Printf("Vision: detected %d vertical segments, %d horizontal segments\n", len(colSegs), len(rowSegs))

	if len(colSegs) < 2 || len(rowSegs) < 2 {
		return image.Rectangle{}, 0, fmt.Errorf("insufficient line segments (%d vertical, %d horizontal)", len(colSegs), len(rowSegs))
	}

	// compute approximate grid size as number of cells between lines
	nX := len(colSegs) - 1
	nY := len(rowSegs) - 1
	n := nX
	if nY > n {
		n = nY
	}
	if n < 2 {
		return image.Rectangle{}, 0, fmt.Errorf("invalid detected grid size %d", n)
	}

	// Use outermost segment edges as bounding box
	left := b.Min.X + colSegs[0].start
	right := b.Min.X + colSegs[len(colSegs)-1].end
	top := b.Min.Y + rowSegs[0].start
	bottom := b.Min.Y + rowSegs[len(rowSegs)-1].end

	// If the outer segments are significantly thicker than median, prefer their exact edges.
	colWidths := make([]int, len(colSegs))
	for i, s := range colSegs {
		colWidths[i] = s.end - s.start + 1
	}
	rowWidths := make([]int, len(rowSegs))
	for i, s := range rowSegs {
		rowWidths[i] = s.end - s.start + 1
	}
	sort.Ints(colWidths)
	sort.Ints(rowWidths)
	medColW := colWidths[len(colWidths)/2]
	medRowW := rowWidths[len(rowWidths)/2]

	if (colSegs[0].end - colSegs[0].start + 1) >= medColW*3/2 {
		left = b.Min.X + colSegs[0].start
	}
	if (colSegs[len(colSegs)-1].end - colSegs[len(colSegs)-1].start + 1) >= medColW*3/2 {
		right = b.Min.X + colSegs[len(colSegs)-1].end
	}
	if (rowSegs[0].end - rowSegs[0].start + 1) >= medRowW*3/2 {
		top = b.Min.Y + rowSegs[0].start
	}
	if (rowSegs[len(rowSegs)-1].end - rowSegs[len(rowSegs)-1].start + 1) >= medRowW*3/2 {
		bottom = b.Min.Y + rowSegs[len(rowSegs)-1].end
	}

	rect := image.Rect(left, top, right, bottom)
	return rect, n, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func buildRegions(img *image.Gray, bounds image.Rectangle, n int) [][]int {
	cellW := float64(bounds.Dx()) / float64(n)
	cellH := float64(bounds.Dy()) / float64(n)

	adj := make([][]bool, n*n)
	for i := range adj {
		adj[i] = make([]bool, n*n)
	}

	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			// Check right border
			if c < n-1 {
				x := bounds.Min.X + int(float64(c+1)*cellW)
				yStart := bounds.Min.Y + int(float64(r)*cellH) + int(cellH/4)
				yEnd := bounds.Min.Y + int(float64(r+1)*cellH) - int(cellH/4)
				if isThinLine(img, x, yStart, yEnd, true) {
					adj[r*n+c][r*n+c+1] = true
					adj[r*n+c+1][r*n+c] = true
				}
			}
			// Check bottom border
			if r < n-1 {
				y := bounds.Min.Y + int(float64(r+1)*cellH)
				xStart := bounds.Min.X + int(float64(c)*cellW) + int(cellW/4)
				xEnd := bounds.Min.X + int(float64(c+1)*cellW) - int(cellW/4)
				if isThinLine(img, y, xStart, xEnd, false) {
					adj[r*n+c][(r+1)*n+c] = true
					adj[(r+1)*n+c][r*n+c] = true
				}
			}
		}
	}

	regions := make([][]int, n)
	for i := range regions {
		regions[i] = make([]int, n)
		for j := range regions[i] {
			regions[i][j] = -1
		}
	}

	nextID := 0
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			if regions[r][c] == -1 {
				floodFill(r, c, nextID, regions, adj, n)
				nextID++
			}
		}
	}

	return regions
}

func floodFill(r, c, id int, regions [][]int, adj [][]bool, n int) {
	q := []int{r*n + c}
	regions[r][c] = id
	for len(q) > 0 {
		curr := q[0]
		q = q[1:]
		for i := 0; i < n*n; i++ {
			if adj[curr][i] {
				ir, ic := i/n, i%n
				if regions[ir][ic] == -1 {
					regions[ir][ic] = id
					q = append(q, i)
				}
			}
		}
	}
}

func isThinLine(img *image.Gray, pos, start, end int, vertical bool) bool {
	energy := 0
	for i := start; i < end; i++ {
		for d := -3; d <= 3; d++ {
			var val uint8
			if vertical {
				val = img.GrayAt(pos+d, i).Y
			} else {
				val = img.GrayAt(i, pos+d).Y
			}
			if val > 128 {
				energy++
			}
		}
	}

	avgEnergy := float64(energy) / float64(end-start)
	// Experimental: Thick borders have double edge energy (two lines close together)
	return avgEnergy < 3.0
}

// closeBinary performs a simple morphological closing (dilation then erosion)
// on a binary grayscale image (values 0 or 255). Radius should be small (1-2).
func closeBinary(img *image.Gray, radius int) *image.Gray {
	b := img.Bounds()
	dil := image.NewGray(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			white := uint8(0)
			found := false
			for dy := -radius; dy <= radius && !found; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					xx := x + dx
					yy := y + dy
					if xx >= b.Min.X && xx < b.Max.X && yy >= b.Min.Y && yy < b.Max.Y {
						if img.GrayAt(xx, yy).Y > 128 {
							white = 255
							found = true
							break
						}
					}
				}
			}
			dil.SetGray(x, y, color.Gray{Y: white})
		}
	}

	ero := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			keep := uint8(255)
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					xx := x + dx
					yy := y + dy
					if xx >= b.Min.X && xx < b.Max.X && yy >= b.Min.Y && yy < b.Max.Y {
						if dil.GrayAt(xx, yy).Y == 0 {
							keep = 0
							dx = radius + 1
							dy = radius + 1
						}
					}
				}
			}
			ero.SetGray(x, y, color.Gray{Y: keep})
		}
	}
	return ero
}

// getGrayComponent extracts the gray value from an RGBA color
func getGrayComponent(c color.Color) uint8 {
	r, _, _, _ := c.RGBA()
	// Use red channel (all should be identical for grayscale)
	return uint8(r >> 8)
}

// buildLineMask extracts candidate tile lines by projecting grayscale darkness
// and edge strength across columns/rows, then selecting strong local peaks.
func buildLineMask(grayImg image.Image, edgeMap *image.Gray) *image.Gray {
	b := edgeMap.Bounds()
	w := b.Dx()
	h := b.Dy()

	colScores := make([]float64, w)
	rowScores := make([]float64, h)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			g := float64(getGrayComponent(grayImg.At(x, y)))
			dark := (255.0 - g) / 255.0
			e := float64(edgeMap.GrayAt(x, y).Y) / 255.0

			score := dark*0.35 + e*2.0
			colScores[x-b.Min.X] += score
			rowScores[y-b.Min.Y] += score
		}
	}

	colScores = smooth1D(colScores, 2)
	rowScores = smooth1D(rowScores, 2)

	colMax := maxFloat(colScores)
	rowMax := maxFloat(rowScores)
	colMean := meanFloat(colScores)
	rowMean := meanFloat(rowScores)

	colThr := maxFloat2(colMax*0.58, colMean*1.18)
	rowThr := maxFloat2(rowMax*0.58, rowMean*1.18)

	minGap := max(3, min(w, h)/70)
	colPeaks := pickPeaks(colScores, colThr, minGap)
	rowPeaks := pickPeaks(rowScores, rowThr, minGap)

	fmt.Printf("Vision: Line peaks detected: %d vertical, %d horizontal\n", len(colPeaks), len(rowPeaks))

	lineMask := image.NewGray(b)
	for _, cx := range colPeaks {
		x := b.Min.X + cx
		for y := b.Min.Y; y < b.Max.Y; y++ {
			lineMask.SetGray(x, y, color.Gray{Y: 255})
			if x+1 < b.Max.X {
				lineMask.SetGray(x+1, y, color.Gray{Y: 255})
			}
		}
	}

	for _, ry := range rowPeaks {
		y := b.Min.Y + ry
		for x := b.Min.X; x < b.Max.X; x++ {
			lineMask.SetGray(x, y, color.Gray{Y: 255})
			if y+1 < b.Max.Y {
				lineMask.SetGray(x, y+1, color.Gray{Y: 255})
			}
		}
	}

	return lineMask
}

func smooth1D(values []float64, radius int) []float64 {
	out := make([]float64, len(values))
	for i := range values {
		start := i - radius
		if start < 0 {
			start = 0
		}
		end := i + radius
		if end >= len(values) {
			end = len(values) - 1
		}
		sum := 0.0
		count := 0.0
		for j := start; j <= end; j++ {
			sum += values[j]
			count += 1.0
		}
		out[i] = sum / count
	}
	return out
}

func pickPeaks(values []float64, threshold float64, minGap int) []int {
	peaks := make([]int, 0)
	last := -minGap * 2
	for i := 1; i < len(values)-1; i++ {
		if values[i] < threshold {
			continue
		}
		if values[i] < values[i-1] || values[i] < values[i+1] {
			continue
		}
		if i-last < minGap {
			if len(peaks) > 0 && values[i] > values[peaks[len(peaks)-1]] {
				peaks[len(peaks)-1] = i
				last = i
			}
			continue
		}
		peaks = append(peaks, i)
		last = i
	}
	return peaks
}

func maxFloat(values []float64) float64 {
	m := 0.0
	for _, v := range values {
		if v > m {
			m = v
		}
	}
	return m
}

func meanFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range values {
		s += v
	}
	return s / float64(len(values))
}

func maxFloat2(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// dilateBinary performs morphological dilation on a binary grayscale image.
// Dilation expands white regions, useful for thickening edges and connecting
// nearby edge segments. Radius should be small (1-2).
func dilateBinary(img *image.Gray, radius int) *image.Gray {
	b := img.Bounds()
	result := image.NewGray(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			white := uint8(0)
			// If this pixel or any nearby pixel is white, output is white
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					xx := x + dx
					yy := y + dy
					if xx >= b.Min.X && xx < b.Max.X && yy >= b.Min.Y && yy < b.Max.Y {
						if img.GrayAt(xx, yy).Y > 128 {
							white = 255
							goto next_pixel
						}
					}
				}
			}
		next_pixel:
			result.SetGray(x, y, color.Gray{Y: white})
		}
	}
	return result
}
