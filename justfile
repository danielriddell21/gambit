binary := "gambit"

# list available recipes
default:
    @just --list

# build the CLI binary (terminal-only, no Ebiten linked)
build:
    go build -o bin/{{binary}} ./cmd/gambit

# build the GUI binary (Ebiten window; on Linux this needs the OpenGL/X11 libs)
build-gui:
    go build -tags ebiten -o bin/{{binary}}-gui ./cmd/gambit

# run in the terminal (minimax vs random by default; pass extra flags after --)
run *ARGS:
    go run ./cmd/gambit {{ARGS}}

# run the GUI window (Ebiten)
gui *ARGS:
    go run -tags ebiten ./cmd/gambit {{ARGS}}

# run the tests
test:
    go test ./...

# run the tests with the race detector
test-race:
    go test -race ./...

# run move-generation perft tests verbosely
perft:
    go test ./pkg/chess -run TestPerft -v

# run the linter
lint:
    golangci-lint run

# format the code
fmt:
    gofmt -w .

# tidy module dependencies
tidy:
    go mod tidy

# regenerate the demo GIFs under docs/demos by recording the GUI
demos:
    go run -tags ebiten ./cmd/gambit -white minimax -black random -seed 7 -delay 1ms -square 48 -record docs/demos/minimax-vs-random.gif
    go run -tags ebiten ./cmd/gambit -white minimax -black minimax -depth 3 -seed 1 -delay 1ms -square 48 -record docs/demos/minimax-vs-minimax.gif
