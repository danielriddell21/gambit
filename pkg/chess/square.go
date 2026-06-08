package chess

import (
	"fmt"
)

// Square is a board square index in the range 0..63, where A1=0, B1=1, ...,
// H8=63. NoSquare (64) is the "no such square" sentinel.
type Square uint8

// NoSquare marks the absence of a square (e.g. no en-passant target).
const NoSquare Square = 64

// NewSquare builds a square from zero-based file (0=a..7=h) and rank
// (0=rank 1..7=rank 8).
func NewSquare(file, rank int) Square {
	return Square(rank*8 + file)
}

// File returns the zero-based file (0=a..7=h).
func (s Square) File() int {
	return int(s) & 7
}

// Rank returns the zero-based rank (0=rank 1..7=rank 8).
func (s Square) Rank() int {
	return int(s) >> 3
}

// Valid reports whether the square is on the board.
func (s Square) Valid() bool {
	return s < 64
}

// String returns the square in algebraic form, e.g. "e4", or "-" for NoSquare.
func (s Square) String() string {
	if !s.Valid() {
		return "-"
	}
	return fmt.Sprintf("%c%c", 'a'+s.File(), '1'+s.Rank())
}

// ParseSquare parses an algebraic square such as "e4".
func ParseSquare(str string) (Square, error) {
	if len(str) != 2 {
		return NoSquare, fmt.Errorf("chess: invalid square %q", str)
	}
	file := int(str[0] - 'a')
	rank := int(str[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return NoSquare, fmt.Errorf("chess: invalid square %q", str)
	}
	return NewSquare(file, rank), nil
}
