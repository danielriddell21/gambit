package chess

import "testing"

// Perft counts the leaf nodes of the move tree to the given depth. It is the
// standard correctness check for move generation.
func Perft(b *Board, depth int) uint64 {
	if depth == 0 {
		return 1
	}
	moves := b.GenerateMoves(nil)
	if depth == 1 {
		return uint64(len(moves))
	}
	var nodes uint64
	for _, m := range moves {
		u := b.MakeMove(m)
		nodes += Perft(b, depth-1)
		b.UnmakeMove(m, u)
	}
	return nodes
}

func perftFromFEN(t *testing.T, fen string, depth int, want uint64) {
	t.Helper()
	b, err := ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN(%q): %v", fen, err)
	}
	if got := Perft(b, depth); got != want {
		t.Errorf("perft(%d) for %q = %d, want %d", depth, fen, got, want)
	}
}

func TestPerftStartpos(t *testing.T) {
	cases := []struct {
		depth int
		want  uint64
	}{
		{1, 20},
		{2, 400},
		{3, 8902},
		{4, 197281},
	}
	for _, c := range cases {
		perftFromFEN(t, StartingFEN, c.depth, c.want)
	}
	if testing.Short() {
		return
	}
	perftFromFEN(t, StartingFEN, 5, 4865609)
}

// TestPerftKiwipete exercises castling, en passant, promotions and pins.
func TestPerftKiwipete(t *testing.T) {
	const fen = "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	perftFromFEN(t, fen, 1, 48)
	perftFromFEN(t, fen, 2, 2039)
	perftFromFEN(t, fen, 3, 97862)
	if testing.Short() {
		return
	}
	perftFromFEN(t, fen, 4, 4085603)
}

// TestPerftPosition3 is an endgame-heavy position from the CPW perft suite.
func TestPerftPosition3(t *testing.T) {
	const fen = "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"
	perftFromFEN(t, fen, 1, 14)
	perftFromFEN(t, fen, 2, 191)
	perftFromFEN(t, fen, 3, 2812)
	perftFromFEN(t, fen, 4, 43238)
}

// TestPerftPosition4 covers promotions and underpromotions with checks.
func TestPerftPosition4(t *testing.T) {
	const fen = "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"
	perftFromFEN(t, fen, 1, 6)
	perftFromFEN(t, fen, 2, 264)
	perftFromFEN(t, fen, 3, 9467)
}

// TestPerftPosition5 is a tricky position that catches en-passant edge cases.
func TestPerftPosition5(t *testing.T) {
	const fen = "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8"
	perftFromFEN(t, fen, 1, 44)
	perftFromFEN(t, fen, 2, 1486)
	perftFromFEN(t, fen, 3, 62379)
}
