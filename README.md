# starbattlego-solver

Star Battle grid detector and solver with on-screen overlay.

## Project layout

- src/config: Runtime config loading from flags and .env
- src/vision: Screen capture and grid detection pipeline
- src/solver: Backtracking solver
- src/overlay: Overlay rendering and area selection
- main.go: App entry point wiring modules together

## Debug images

Debug images are now only saved when debug mode is enabled.

Enable debug with one of these options:
- Command flag: -debug
- .env value: DEBUG=true

Optional output folder:
- Command flag: -debug-dir <folder>
- .env value: DEBUG_IMAGES_DIR=<folder>

If debug is enabled and no folder is set, images go to debug_images.

## Run

With debug disabled:
- go run .

With debug enabled:
- go run . -debug