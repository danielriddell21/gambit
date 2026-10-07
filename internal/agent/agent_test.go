package agent

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/danielriddell21/gambit/pkg/chess"
)

var strategies = []struct {
	name string
	opts Options
}{
	{"beam", Options{Seed: 1, Depth: 2, Width: 4}},
	{"greedy", Options{Seed: 1}},
	{"iterative", Options{Seed: 1, Depth: 2}},
	{"mcts", Options{Seed: 1, Iterations: 50}},
	{"minimax", Options{Seed: 1, Depth: 2}},
	{"random", Options{Seed: 1}},
}

func mustFEN(t *testing.T, fen string) *chess.Board {
	t.Helper()
	b, err := chess.ParseFEN(fen)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestAvailableListsEveryStrategy(t *testing.T) {
	got := Available()
	if !slices.IsSorted(got) {
		t.Errorf("Available() = %v, want sorted", got)
	}
	for _, s := range strategies {
		if !slices.Contains(got, s.name) {
			t.Errorf("Available() = %v, missing %q", got, s.name)
		}
	}
}

func TestNewRejectsUnknownStrategies(t *testing.T) {
	if _, err := New("nope", Options{}); err == nil {
		t.Fatal("New(nope) succeeded, want an error")
	}
}

func TestEveryStrategyPlaysALegalMove(t *testing.T) {
	for _, s := range strategies {
		t.Run(s.name, func(t *testing.T) {
			a, err := New(s.name, s.opts)
			if err != nil {
				t.Fatal(err)
			}
			if a.Name() != s.name {
				t.Errorf("Name() = %q, want %q", a.Name(), s.name)
			}
			b := chess.NewStartingBoard()
			m, err := a.SelectMove(context.Background(), b)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(b.LegalMoves(), m) {
				t.Errorf("%s played %v, which is not legal", s.name, m)
			}
		})
	}
}

func TestEveryStrategyFailsWithNoLegalMoves(t *testing.T) {
	// Black is checkmated.
	b := mustFEN(t, "3Q2k1/5ppp/8/8/8/8/8/6K1 b - - 0 1")
	for _, s := range strategies {
		t.Run(s.name, func(t *testing.T) {
			a, err := New(s.name, s.opts)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := a.SelectMove(context.Background(), b); !errors.Is(err, errNoMoves) {
				t.Errorf("SelectMove on a finished game: err = %v, want errNoMoves", err)
			}
		})
	}
}

func TestSearchingStrategiesTakeAHangingQueen(t *testing.T) {
	// The black queen on d5 is undefended and white's queen can take it.
	b := mustFEN(t, "4k3/8/8/3q4/8/8/8/3QK3 w - - 0 1")
	d1, _ := chess.ParseSquare("d1")
	d5, _ := chess.ParseSquare("d5")
	for _, s := range strategies {
		if s.name == "random" || s.name == "mcts" {
			continue
		}
		t.Run(s.name, func(t *testing.T) {
			a, err := New(s.name, s.opts)
			if err != nil {
				t.Fatal(err)
			}
			m, err := a.SelectMove(context.Background(), b)
			if err != nil {
				t.Fatal(err)
			}
			if m.From() != d1 || m.To() != d5 {
				t.Errorf("%s played %v, want Qxd5", s.name, m)
			}
		})
	}
}

func TestSeededStrategiesAreDeterministic(t *testing.T) {
	for _, s := range strategies {
		t.Run(s.name, func(t *testing.T) {
			first, _ := New(s.name, s.opts)
			second, _ := New(s.name, s.opts)
			b := chess.NewStartingBoard()
			m1, err1 := first.SelectMove(context.Background(), b)
			m2, err2 := second.SelectMove(context.Background(), b.Clone())
			if err1 != nil || err2 != nil {
				t.Fatal(err1, err2)
			}
			if m1 != m2 {
				t.Errorf("two %s agents with one seed played %v and %v", s.name, m1, m2)
			}
		})
	}
}
