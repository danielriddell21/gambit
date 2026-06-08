// Package agent defines the pluggable AI-player interface and a registry of
// strategies. Adding a new search strategy means adding one file that registers
// a factory; the game loop and GUI never change.
package agent

import (
	"context"
	"fmt"
	"sort"

	"github.com/danielriddell21/gambit/pkg/chess"
)

// Agent selects a move for the side to move in a given position.
type Agent interface {
	// Name identifies the strategy (used for logging and selection).
	Name() string
	// SelectMove returns the chosen move. The context carries optional time or
	// iteration budgets and allows cancellation so the GUI stays responsive.
	SelectMove(ctx context.Context, b *chess.Board) (chess.Move, error)
}

// Options carries tunable knobs for constructing an agent. New strategies add
// fields here as needed (e.g. beam width).
type Options struct {
	Seed  int64 // RNG seed for stochastic strategies
	Depth int   // search depth for depth-limited strategies
}

// Factory builds an agent from options.
type Factory func(opts Options) (Agent, error)

var registry = map[string]Factory{}

// Register makes a strategy available by name. It is intended to be called from
// an init function and panics on a duplicate name.
func Register(name string, f Factory) {
	if _, exists := registry[name]; exists {
		panic("agent: duplicate registration for " + name)
	}
	registry[name] = f
}

// New constructs a registered agent by name.
func New(name string, opts Options) (Agent, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("agent: unknown strategy %q (available: %v)", name, Available())
	}
	return f(opts)
}

// Available returns the sorted names of all registered strategies.
func Available() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
