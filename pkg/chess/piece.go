package chess

// PieceType is the kind of a piece, independent of color.
type PieceType uint8

// Piece types. NoPieceType marks an empty square's type.
const (
	NoPieceType PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

// Piece encodes a colored piece in a single byte. The low three bits hold the
// PieceType and the next bit holds the Color. NoPiece (zero) is an empty square.
type Piece uint8

// NoPiece is the empty-square sentinel.
const NoPiece Piece = 0

const (
	pieceTypeMask Piece = 0b0111
	pieceColorBit Piece = 0b1000
)

// MakePiece builds a Piece from a color and type.
func MakePiece(c Color, t PieceType) Piece {
	p := Piece(t)
	if c == Black {
		p |= pieceColorBit
	}
	return p
}

// IsEmpty reports whether the piece is the empty-square sentinel.
func (p Piece) IsEmpty() bool {
	return p == NoPiece
}

// Color returns the owning color. Meaningless for NoPiece.
func (p Piece) Color() Color {
	if p&pieceColorBit != 0 {
		return Black
	}
	return White
}

// Type returns the piece type (NoPieceType for an empty square).
func (p Piece) Type() PieceType {
	return PieceType(p & pieceTypeMask)
}

// pieceLetters maps a PieceType to its uppercase (white) FEN letter.
var pieceLetters = [...]byte{
	Pawn:   'P',
	Knight: 'N',
	Bishop: 'B',
	Rook:   'R',
	Queen:  'Q',
	King:   'K',
}

// Symbol returns the FEN letter for the piece: uppercase for white, lowercase
// for black. Returns a space for an empty square.
func (p Piece) Symbol() rune {
	if p.IsEmpty() {
		return ' '
	}
	letter := pieceLetters[p.Type()]
	if p.Color() == Black {
		letter += 'a' - 'A'
	}
	return rune(letter)
}
