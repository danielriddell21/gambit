# gambit

Two chess agents play each other in an [Ebiten](https://ebitengine.org/) window;
every move is logged to the terminal in algebraic notation. Built as a testbed
for experimenting with search strategies — see [docs/demos.md](docs/demos.md).

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
just run                       # GUI: minimax (white) vs random (black)
just headless                  # terminal only, no window
just test                      # tests        just perft   # move-gen correctness
just lint                      # golangci-lint just demos   # regenerate demo GIFs
```

Flags: `-white`, `-black`, `-depth`, `-seed`, `-fen`, `-delay`.

## Adding a strategy

Implement `agent.Agent` and `agent.Register` it in an `init` — the game loop and
GUI need no changes. `Board.MakeMove`/`UnmakeMove` (for alpha-beta) and
`Board.ApplyMove` (clone, for keeping many positions alive) are both available,
and the headless build batches games with no display.
