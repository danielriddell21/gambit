# Contributing to gambit

## Requirements

* [Go](https://go.dev) (stable — version from `go.mod`)
* [just](https://github.com/casey/just)
* [golangci-lint](https://golangci-lint.run/welcome/install/) (for `just lint`)
* [gremlins](https://github.com/go-gremlins/gremlins) (for mutation testing)

A C toolchain is only needed for the optional Ebiten GUI (`just build-gui`).

## Development workflow

```
just build      # build the headless CLI
just build-gui  # build with the Ebiten GUI (-tags ebiten, needs cgo)
just test       # run unit tests
just test-race  # run tests with the race detector
just perft      # run the move-generation perft checks
just lint       # golangci-lint
just fmt        # gofumpt
just tidy       # go mod tidy
just demos      # regenerate demo assets
```

Run `just --list` to see every recipe. Run `just lint` and `just test` before each commit. CI runs lint + test + build on every push to `trunk` and every pull request targeting `trunk`.

## Project layout

gambit is both a CLI and a reusable chess library.

```
pkg/chess/       public chess library
cmd/gambit/      CLI entry point (headless + optional Ebiten GUI)
internal/        implementation packages (agents, game, ui)
docs/            documentation
```

## Commit style

```
type(scope): short imperative description
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`. No period at the end of the subject line; keep it under 72 characters.

## Releases

Releases are triggered by pushing a semver tag — maintainers only. A GitHub Actions workflow runs GoReleaser to build the binaries (and the macOS GUI cask) and update the Homebrew tap; it requires the tap app credentials configured as repository secrets.
