# Star Battle Solver - Implementation Plan (Pure Go Version)

## 1. Overview
The goal is to build an autonomous Star Battle ("Two Not Touch") solver application in Go. To ensure **universal installation** and portability, this version will use **Pure Go** libraries wherever possible, avoiding the heavy and complex OpenCV (CGO) dependency. The final result will be a single portable binary.

## 2. Technical Stack & Libraries (Universal/Pure Go)
*   **Language:** Go (Golang)
*   **Screen Capture:** `github.com/kbinani/screenshot`
    *   *Portability:* Cross-platform (Windows, macOS, Linux). Mostly cgo-free.
*   **Image Processing:** `image` (Std Lib) + `github.com/anthonynsimon/bild`
    *   *Portability:* **Pure Go.** No external C libraries required.
    *   *Usage:* `bild` will be used for grayscale conversion and edge detection to simplify the image before our custom grid-parsing logic runs.
*   **Solver Algorithm:** Custom Backtracking Search (Pure Go).
*   **Transparent Overlay GUI:** `github.com/hajimehoshi/ebiten/v2`
    *   *Portability:* While it uses some system libraries (OpenGL/DirectX), it is the industry standard for Go games and is significantly easier to install/bundle than OpenCV. It supports transparent, always-on-top windows.

## 3. Architecture & Data Flow

### Phase A: Capture and Vision (Pure Go Logic)
1.  **Capture:** Take a screenshot of the primary display using `screenshot`.
2.  **Preprocessing:** Convert to grayscale and apply a high-contrast threshold using `bild`.
3.  **Grid Locating (Custom Scan):**
    *   Scan for long vertical and horizontal lines (continuous runs of dark pixels).
    *   Identify the largest set of intersecting lines that form a square/rectangular grid.
    *   Determine grid bounds `(x, y, w, h)` and grid size $N$ (number of cells).
4.  **Region Mapping:**
    *   Calculate the center lines between cells.
    *   Check the pixel thickness of the lines between each cell. Thick lines = Region boundary.
    *   Use a Flood Fill algorithm on the cell boundaries to assign each cell a Region ID.

### Phase B: Solving (Pure Go)
1.  **Constraints:** 2 stars per row/column/region, no touching.
2.  **Algorithm:** Recursive backtracking with MRV (Minimum Remaining Values) and Forward Checking.

### Phase C: Overlay Rendering
1.  **Mapping:** Convert $(r, c)$ star positions to absolute screen pixels based on the detected grid bounds.
2.  **Ebiten Overlay:**
    *   Create a full-screen transparent window.
    *   Draw stars (simple circles or polygons) at the calculated positions.
    *   Set the window to "Always on Top" and "Click-Through" (using `ebiten.SetWindowMousePassthrough(true)` where supported).

## 4. Implementation Steps
1.  **Setup:** `go mod init`, fetch `screenshot`, `bild`, and `ebiten`.
2.  **Module 1 (Vision):** Implement the custom grid-finding logic using the `image` package. (Avoids OpenCV installation issues).
3.  **Module 2 (Solver):** Implement the backtracking solver with unit tests.
4.  **Module 3 (Overlay):** Create the Ebiten boilerplate for transparency.
5.  **Integration:** Connect the pipeline.

## 5. Benefits of this Workaround
*   **Single Binary:** No need for the user to install OpenCV, shared libraries, or complex C++ toolchains.
*   **Fast Compilation:** Standard `go build` works out of the box.
*   **Cross-Platform:** The same logic applies cleanly to Windows, Linux, and macOS.