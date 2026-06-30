package agent

import (
	"context"
	"fmt"
	"sort"

	"github.com/danielriddell21/gambit/pkg/chess"
)

type Agent interface {
	Name() string

	SelectMove(ctx context.Context, b *chess.Board) (chess.Move, error)
}

type Options struct {
	Seed       int64
	Depth      int
	Width      int
	Iterations int
}

type Factory func(opts Options) (Agent, error)

var registry = map[string]Factory{}

func Register(name string, f Factory) {
	if _, exists := registry[name]; exists {
		panic("agent: duplicate registration for " + name)
	}
	registry[name] = f
}

func New(name string, opts Options) (Agent, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("agent: unknown strategy %q (available: %v)", name, Available())
	}
	return f(opts)
}

func Available() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
