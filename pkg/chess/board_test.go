package chess

import "testing"

func TestFENRoundTrip(t *testing.T) {
	fens := []string{
		StartingFEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2",
		"4k3/8/8/8/8/8/8/4K3 b - - 5 39",
	}
	for _, fen := range fens {
		b, err := ParseFEN(fen)
		if err != nil {
			t.Errorf("ParseFEN(%q): %v", fen, err)
			continue
		}
		if got := b.FEN(); got != fen {
			t.Errorf("FEN round-trip: got %q, want %q", got, fen)
		}
	}
}

func TestParseFENErrors(t *testing.T) {
	bad := []string{
		"",
		"too few fields",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq - 0 1",
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP w KQkq - 0 1", // 7 ranks
	}
	for _, fen := range bad {
		if _, err := ParseFEN(fen); err == nil {
			t.Errorf("ParseFEN(%q): expected error, got nil", fen)
		}
	}
}

// TestMakeUnmakeInvariance verifies that MakeMove followed by UnmakeMove
// restores the exact position, checked via the Zobrist hash and full FEN. This
// is the property alpha-beta search relies on.
func TestMakeUnmakeInvariance(t *testing.T) {
	fens := []string{
		StartingFEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2",
	}
	for _, fen := range fens {
		b, err := ParseFEN(fen)
		if err != nil {
			t.Fatalf("ParseFEN(%q): %v", fen, err)
		}
		before := b.FEN()
		hash := b.Hash()
		for _, m := range b.LegalMoves() {
			u := b.MakeMove(m)
			b.UnmakeMove(m, u)
			if got := b.FEN(); got != before {
				t.Errorf("move %s: FEN after make/unmake = %q, want %q", m, got, before)
			}
			if got := b.Hash(); got != hash {
				t.Errorf("move %s: hash changed after make/unmake", m)
			}
		}
	}
}

func TestCheckmate(t *testing.T) {
	// Fool's mate: 1. f3 e5 2. g4 Qh4#.
	b, err := ParseFEN("rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3")
	if err != nil {
		t.Fatal(err)
	}
	if !b.IsCheckmate() {
		t.Error("expected checkmate")
	}
	res, _ := b.Status()
	if res != BlackWins {
		t.Errorf("Status() = %v, want BlackWins", res)
	}
}

func TestStalemate(t *testing.T) {
	// Classic king+pawn stalemate, black to move.
	b, err := ParseFEN("7k/5Q2/6K1/8/8/8/8/8 b - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if !b.IsStalemate() {
		t.Error("expected stalemate")
	}
	res, reason := b.Status()
	if res != Draw || reason != Stalemate {
		t.Errorf("Status() = %v/%v, want Draw/Stalemate", res, reason)
	}
}

func TestInsufficientMaterial(t *testing.T) {
	cases := []struct {
		fen  string
		want bool
	}{
		{"4k3/8/8/8/8/8/8/4K3 w - - 0 1", true},    // K vs K
		{"4k3/8/8/8/8/8/8/3BK3 w - - 0 1", true},   // K+B vs K
		{"4k3/8/8/8/8/8/8/3NK3 w - - 0 1", true},   // K+N vs K
		{"4k3/8/8/8/8/8/4P3/4K3 w - - 0 1", false}, // K+P vs K
		{"4k3/8/8/8/8/8/8/R3K3 w - - 0 1", false},  // K+R vs K
		{"5bk1/8/8/8/8/8/8/2B1K3 w - - 0 1", true}, // bishops on same color
		{"3bk3/8/8/8/8/8/8/3BK3 w - - 0 1", false}, // bishops on opposite colors
	}
	for _, c := range cases {
		b, err := ParseFEN(c.fen)
		if err != nil {
			t.Fatalf("ParseFEN(%q): %v", c.fen, err)
		}
		if got := b.IsInsufficientMaterial(); got != c.want {
			t.Errorf("IsInsufficientMaterial(%q) = %v, want %v", c.fen, got, c.want)
		}
	}
}
