package chess

// CastleRights is a bitset of the four castling possibilities.
type CastleRights uint8

// Castling-rights bits.
const (
	WhiteKingside CastleRights = 1 << iota
	WhiteQueenside
	BlackKingside
	BlackQueenside
)

// Has reports whether all of the given rights are present.
func (c CastleRights) Has(r CastleRights) bool {
	return c&r == r
}

// Board is a single chess position using an 8x8 mailbox representation. It is
// deliberately representation-agnostic at the API level so the internals could
// be swapped for bitboards later without breaking callers.
//
// [Board.MakeMove] and [Board.UnmakeMove] mutate the receiver in place and are
// meant to be paired for fast search; [Board.ApplyMove] instead returns a new
// board and leaves the receiver unchanged, for callers that want an immutable
// step.
type Board struct {
	squares    [64]Piece
	sideToMove Color
	castling   CastleRights
	enPassant  Square // NoSquare when no en-passant capture is available
	halfMove   int    // halfmove clock for the fifty-move rule
	fullMove   int    // starts at 1, increments after every black move
	kings      [2]Square
}

// NewStartingBoard returns a board in the standard initial position.
func NewStartingBoard() *Board {
	b, err := ParseFEN(StartingFEN)
	if err != nil {
		// StartingFEN is a constant and always valid.
		panic("chess: invalid StartingFEN: " + err.Error())
	}
	return b
}

// Clone returns an independent copy of the board.
func (b *Board) Clone() *Board {
	cp := *b
	return &cp
}

// PieceAt returns the piece on the given square (NoPiece if empty).
func (b *Board) PieceAt(s Square) Piece {
	return b.squares[s]
}

// SideToMove returns the color to move.
func (b *Board) SideToMove() Color {
	return b.sideToMove
}

// EnPassantSquare returns the current en-passant target, or NoSquare.
func (b *Board) EnPassantSquare() Square {
	return b.enPassant
}

// CastleRights returns the current castling rights.
func (b *Board) CastleRights() CastleRights {
	return b.castling
}

// HalfMoveClock returns the halfmove clock (for the fifty-move rule).
func (b *Board) HalfMoveClock() int {
	return b.halfMove
}

// FullMoveNumber returns the full-move counter.
func (b *Board) FullMoveNumber() int {
	return b.fullMove
}

// KingSquare returns the square of the given color's king.
func (b *Board) KingSquare(c Color) Square {
	return b.kings[c]
}

// Each calls fn for every square on the board in index order. Useful for
// rendering without exposing the internal array.
func (b *Board) Each(fn func(s Square, p Piece)) {
	for s := range Square(64) {
		fn(s, b.squares[s])
	}
}

// setPiece places a piece on a square, keeping derived state (king squares)
// consistent. Used only during construction and move application.
func (b *Board) setPiece(s Square, p Piece) {
	b.squares[s] = p
	if !p.IsEmpty() && p.Type() == King {
		b.kings[p.Color()] = s
	}
}
