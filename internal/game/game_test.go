package game

import (
	"context"
	"testing"

	"github.com/danielriddell21/gambit/internal/agent"
	"github.com/danielriddell21/gambit/pkg/chess"
)

// TestRandomGamesTerminate runs many random-vs-random games and asserts every
// one ends with a valid result and that no illegal move is ever applied. This
// fuzzes the whole engine + game stack without the GUI.
func TestRandomGamesTerminate(t *testing.T) {
	for seed := int64(0); seed < 40; seed++ {
		white, err := agent.New("random", agent.Options{Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		black, err := agent.New("random", agent.Options{Seed: seed + 1000})
		if err != nil {
			t.Fatal(err)
		}
		g := New(Players{White: white, Black: black}, nil)
		if err := g.Run(context.Background(), nil); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if !g.Over() {
			t.Fatalf("seed %d: game did not terminate", seed)
		}
		if g.Result() == chess.InProgress {
			t.Fatalf("seed %d: result still InProgress", seed)
		}
	}
}

// TestMinimaxFindsMateInOne checks the alpha-beta agent plays a forced mate.
func TestMinimaxFindsMateInOne(t *testing.T) {
	// White to move: Qd8# is mate (back-rank).
	b, err := chess.ParseFEN("6k1/5ppp/8/8/8/8/8/3Q2K1 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	mm, err := agent.New("minimax", agent.Options{Depth: 3})
	if err != nil {
		t.Fatal(err)
	}
	m, err := mm.SelectMove(context.Background(), b)
	if err != nil {
		t.Fatal(err)
	}
	if SAN(b, m) != "Qd8#" {
		t.Errorf("minimax played %s, want Qd8#", SAN(b, m))
	}
}

// TestSAN spot-checks notation rendering.
func TestSAN(t *testing.T) {
	cases := []struct {
		fen  string
		move chess.Move
		want string
	}{
		{chess.StartingFEN, chess.NewMove(mustSq(t, "e2"), mustSq(t, "e4"), chess.FlagDoublePawnPush), "e4"},
		{chess.StartingFEN, chess.NewMove(mustSq(t, "g1"), mustSq(t, "f3"), chess.FlagNormal), "Nf3"},
	}
	for _, c := range cases {
		b, err := chess.ParseFEN(c.fen)
		if err != nil {
			t.Fatal(err)
		}
		if got := SAN(b, c.move); got != c.want {
			t.Errorf("SAN = %q, want %q", got, c.want)
		}
	}
}

func mustSq(t *testing.T, s string) chess.Square {
	t.Helper()
	sq, err := chess.ParseSquare(s)
	if err != nil {
		t.Fatal(err)
	}
	return sq
}
