# gambit

> *n.* an opening that sacrifices material for position. Also: two bots arguing in algebraic notation.

[![CI](https://github.com/danielriddell21/gambit/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/gambit/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/gambit/graph/badge.svg)](https://codecov.io/gh/danielriddell21/gambit)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_gambit&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_gambit)
[![Go Reference](https://pkg.go.dev/badge/github.com/danielriddell21/gambit/pkg/chess.svg)](https://pkg.go.dev/github.com/danielriddell21/gambit/pkg/chess)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

Two chess agents play each other, every move logged to the terminal in algebraic notation — headless by default, or in an optional [Ebiten](https://ebitengine.org/) window. Built as a testbed for experimenting with search strategies.

## Agents

| Name | Description |
|---|---|
| `random` | Picks uniformly at random from the legal moves |
| `greedy` | Best immediate material gain, no lookahead |
| `minimax` | Alpha-beta search to a fixed depth |
| `iterative` | Iterative-deepening search |
| `beam` | Beam search over a bounded frontier |
| `mcts` | Monte Carlo tree search |
| `human` | You, in the GUI build |

## Install

### Homebrew
```bash
# CLI (all platforms)
brew install danielriddell21/tap/gambit

# native macOS GUI window
brew install --cask danielriddell21/tap/gambit
```

### Go install
```bash
go install github.com/danielriddell21/gambit/cmd/gambit@latest
```

### From source
```bash
git clone https://github.com/danielriddell21/gambit
cd gambit
just run
```

A `go install` build is headless — the window sits behind the `ebiten` build tag.

## Library

`pkg/chess` is the public package — board, moves, legality, FEN and results — and is documented on [pkg.go.dev](https://pkg.go.dev/github.com/danielriddell21/gambit/pkg/chess).

## Documentation

Full documentation lives in the [gambit wiki](https://github.com/danielriddell21/gambit/wiki).
