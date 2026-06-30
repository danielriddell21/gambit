binary := "gambit"

# list available recipes
default:
    @just --list

# build the CLI binary (terminal-only, no Ebiten linked)
[group('build')]
build:
    go build -o bin/{{binary}} ./cmd/gambit

# run the tests
[group('test')]
test:
    go test ./...

# run the linter
[group('dev')]
lint:
    golangci-lint run

# format the code
[group('dev')]
fmt:
    golangci-lint fmt

# tidy module dependencies
[group('dev')]
tidy:
    go mod tidy

# full gate: lint + test + build. all must pass before committing
[group('dev')]
ci: lint test build

# build the GUI binary (Ebiten window; on Linux this needs the OpenGL/X11 libs)
[group('build')]
build-gui:
    go build -tags ebiten -o bin/{{binary}}-gui ./cmd/gambit

# run in the terminal (minimax vs random by default; pass extra flags after --)
[group('run')]
run *ARGS:
    go run ./cmd/gambit {{ARGS}}

# run the GUI window (Ebiten)
[group('run')]
gui *ARGS:
    go run -tags ebiten ./cmd/gambit {{ARGS}}

# run the tests with the race detector
[group('test')]
test-race:
    go test -race ./...

# run move-generation perft tests verbosely
[group('test')]
perft:
    go test ./pkg/chess -run TestPerft -v

# regenerate the demo GIFs under docs/demos by recording the GUI
[group('run')]
demos:
    go run -tags ebiten ./cmd/gambit --white minimax --black random --seed 7 --delay 1ms --square 48 --record docs/demos/minimax-vs-random.gif
    go run -tags ebiten ./cmd/gambit --white minimax --black minimax --depth 3 --seed 1 --delay 1ms --square 48 --record docs/demos/minimax-vs-minimax.gif
