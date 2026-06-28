# gambit
*n.* an opening that sacrifices material for position. Also: two bots arguing in algebraic notation.


[![CI](https://github.com/danielriddell21/gambit/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/gambit/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/gambit/graph/badge.svg)](https://codecov.io/gh/danielriddell21/gambit)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_gambit&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_gambit)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

Two chess agents play each other, every move logged to the terminal in algebraic
notation — headless by default, or in an optional [Ebiten](https://ebitengine.org/)
window. Built as a testbed for experimenting with search strategies — see
[docs/demos.md](docs/demos.md).

## Install

### Homebrew
```sh
brew install danielriddell21/tap/gambit         # CLI (all platforms)
brew install --cask danielriddell21/tap/gambit  # native macOS GUI window
```

## Layout

- `cmd/gambit` — entry point.
- `pkg/chess` — reusable chess engine (board, moves, legality, FEN, results). The only public package.
- `internal/agent` — pluggable `Agent` interface + the `random` and `minimax` strategies.
- `internal/game` — turn orchestration, draw detection, SAN.
- `internal/ui` — Ebiten rendering.

## Usage

Needs Go 1.26.3, [`just`](https://github.com/casey/just), and (for the GUI)
Ebiten's [system deps](https://ebitengine.org/en/documents/install.html).

```sh
just run                       # terminal: minimax (white) vs random (black)
just gui                       # Ebiten window (opt-in, -tags ebiten)
just test                      # tests        just perft   # move-gen correctness
just lint                      # golangci-lint just demos   # regenerate demo GIFs
```

Flags: `--white`, `--black`, `--depth`, `--seed`, `--fen`, `--delay`.

<details>
<summary>Linux: OpenGL/X11 libraries</summary>

The window needs OpenGL/X11. On Debian/Ubuntu:

```sh
sudo apt install libgl1-mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev
```
</details>

## Adding a strategy

Implement `agent.Agent` and `agent.Register` it in an `init` — the game loop and
GUI need no changes. `Board.MakeMove`/`UnmakeMove` (for alpha-beta) and
`Board.ApplyMove` (clone, for keeping many positions alive) are both available,
and the headless build batches games with no display.

## Documentation

- [Demos](docs/demos.md)
