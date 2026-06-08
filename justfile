binary := "gambit"
nongui := "./pkg/... ./internal/agent/... ./internal/game/... ./internal/log/..."

# list available recipes
default:
    @just --list

# build the GUI binary
build:
    go build -o bin/{{binary}} ./cmd/gambit

# run the GUI (minimax vs random by default; pass extra flags after --)
run *ARGS:
    go run ./cmd/gambit {{ARGS}}

# run a game in the terminal only (no window, no Ebiten linked)
headless *ARGS:
    go run -tags headless ./cmd/gambit {{ARGS}}

# run the non-GUI tests
test:
    go test {{nongui}}

# run the non-GUI tests with the race detector
test-race:
    go test -race {{nongui}}

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
    go run ./cmd/gambit -white minimax -black random -seed 7 -delay 1ms -square 48 -record docs/demos/minimax-vs-random.gif
    go run ./cmd/gambit -white minimax -black minimax -depth 3 -seed 1 -delay 1ms -square 48 -record docs/demos/minimax-vs-minimax.gif
