package eval

import (
	"testing"

	"github.com/danielriddell21/gambit/pkg/chess"
)

func mustFEN(t *testing.T, fen string) *chess.Board {
	t.Helper()
	b, err := chess.ParseFEN(fen)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestMaterialIsZeroAtTheStart(t *testing.T) {
	if got := Material(chess.NewStartingBoard()); got != 0 {
		t.Errorf("Material(start) = %d, want 0", got)
	}
}

func TestMaterialIsFromTheSideToMove(t *testing.T) {
	// White has an extra queen.
	white := Material(mustFEN(t, "4k3/8/8/8/8/8/8/3QK3 w - - 0 1"))
	black := Material(mustFEN(t, "4k3/8/8/8/8/8/8/3QK3 b - - 0 1"))
	if white <= 0 {
		t.Errorf("white to move, a queen up: %d, want > 0", white)
	}
	if black != -white {
		t.Errorf("black to move: %d, want %d", black, -white)
	}
}

func TestMaterialValuesPieces(t *testing.T) {
	queen := Material(mustFEN(t, "4k3/8/8/8/8/8/8/3QK3 w - - 0 1"))
	rook := Material(mustFEN(t, "4k3/8/8/8/8/8/8/3RK3 w - - 0 1"))
	pawn := Material(mustFEN(t, "4k3/8/8/8/8/8/3P4/4K3 w - - 0 1"))
	if queen <= rook || rook <= pawn || pawn <= 0 {
		t.Errorf("queen %d, rook %d, pawn %d: want queen > rook > pawn > 0", queen, rook, pawn)
	}
}

func TestMaterialPrefersCentralKnights(t *testing.T) {
	centre := Material(mustFEN(t, "4k3/8/8/8/3N4/8/8/4K3 w - - 0 1"))
	corner := Material(mustFEN(t, "4k3/8/8/8/8/8/8/N3K3 w - - 0 1"))
	if centre <= corner {
		t.Errorf("knight on d4 %d, on a1 %d: want the centre to score higher", centre, corner)
	}
}

func TestMaterialIsSymmetricBetweenColours(t *testing.T) {
	// The same knight, mirrored for black, is worth the same to its owner.
	white := Material(mustFEN(t, "4k3/8/8/8/3N4/8/8/4K3 w - - 0 1"))
	black := Material(mustFEN(t, "4k3/8/8/3n4/8/8/8/4K3 b - - 0 1"))
	if white != black {
		t.Errorf("white knight %d, mirrored black knight %d: want equal", white, black)
	}
}
