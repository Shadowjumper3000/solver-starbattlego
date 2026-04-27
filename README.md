# starbattlego-solver

Star Battle grid detector and solver with an on-screen overlay.

## Project layout

- `src/config`: Runtime config loading from flags and `.env`.
- `src/vision`: Vision interfaces and template processor (see `src/vision/README.md`).
- `src/solver`: Backtracking solver and `Solver` interface adapter.
- `src/overlay`: Overlay rendering and selection; also provides an `Overlayer` adapter.
- `main/`: Application entrypoint wiring modules together (replaces old `main.go`).

## Debug images

Debug images are written only when debug mode is enabled.

Enable debug with one of these options:
- Command flag: `-debug`
- `.env` value: `DEBUG=true`

Optional output folder:
- Command flag: `-debug-dir <folder>`
- `.env` value: `DEBUG_IMAGES_DIR=<folder>`

If debug is enabled and no folder is set, images go to `debug_images/`.

## Build & Run

Build the project (produces no tracked binary):

```bash
cd /home/dahoe/dev/solver-starbattlego
go build ./...
```

Run the app directly in-place (use `main/` entrypoint):

```bash
go run main
```

With debug enabled:

```bash
go run main -debug
```

## Vision module

See `src/vision/README.md` for the vision data contract and expectations for implementations.