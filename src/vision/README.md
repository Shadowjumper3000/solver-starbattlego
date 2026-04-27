Vision module — data contract

This document describes the minimal data contract and expectations for `src/vision`.
The goal is to keep the vision step a single-responsibility module that accepts
an image and a selection rectangle, then returns a structured `solver.Grid`
and the bounding rectangle for the detected grid inside the provided image.

API (Go types)

- Vision interface
  - `ProcessSelection(fullImg image.Image, selection image.Rectangle) (solver.Grid, image.Rectangle, error)`
    - `fullImg`: The full-screen or source image. Coordinates in `selection` are in the same image space.
    - `selection`: The rectangle that was selected by the user (inclusive of Min, exclusive of Max in Go image.Rect semantics).
    - Returns:
      - `solver.Grid`: See `src/solver` for type. Important fields:
        - `Size` (int): the grid size (n for n x n)
        - `Regions` ([][]int): n x n array where each entry is a region identifier (0..R-1). Regions must cover all cells.
        - `Stars` (int): number of stars per row/column/region
      - `image.Rectangle`: bounding rectangle of the detected puzzle grid in the same coordinate space as `fullImg` (i.e., global screen coordinates when `fullImg` is a screenshot).
      - `error`: non-nil on detection failure

Coordinate rules

- All rectangles returned must use the same coordinate space as the provided `fullImg`.
- Cells are addressed row-major as `Regions[row][col]` with `row` in `[0, Size)` and `col` in `[0, Size)`.

Implementation guidance

- Keep `ProcessSelection` single-responsibility: detection only. Do not spawn overlay/UI actions from within vision.
- Implementations may save debug artifacts when requested by a higher-level config; avoid writing to global locations unless explicitly configured.
- For testing, provide a deterministic template implementation that returns a default `10x10` grid so downstream wiring can be validated.

Examples

- Template implementation: `NewTemplateVision()` returns a `Vision` that crops the `fullImg` to `selection` and returns a default `10x10` grid with a single region id (0) for all cells and `Stars=2`.

- Production implementation: perform image processing to detect grid lines, compute regions, and return precise bounds.

Notes

- The orchestrator (in `main/`) is responsible for calling `CaptureScreen()` and passing the correct `fullImg` and `selection` to the `Vision` implementation.
- The solver expects `Regions` to be dense and consistent; inconsistent region maps will cause solver failures.
