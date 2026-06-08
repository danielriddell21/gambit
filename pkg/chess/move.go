package chess

// MoveFlag describes the special nature of a move, if any.
type MoveFlag uint8

// Move flags.
const (
	FlagNormal MoveFlag = iota
	FlagDoublePawnPush
	FlagEnPassant
	FlagCastleKingside
	FlagCastleQueenside
	FlagPromoKnight
	FlagPromoBishop
	FlagPromoRook
	FlagPromoQueen
)

// Move packs a from-square, to-square and flag into a single uint16:
// bits 0..5 = from, bits 6..11 = to, bits 12..15 = flag. This keeps move
// lists cheap to copy and sort during search.
type Move uint16

// NoMove is the zero value, distinguishable because a real move never has
// equal from/to squares.
const NoMove Move = 0

// NewMove constructs a move from its components.
func NewMove(from, to Square, flag MoveFlag) Move {
	return Move(uint16(from) | uint16(to)<<6 | uint16(flag)<<12)
}

// From returns the origin square.
func (m Move) From() Square {
	return Square(m & 0x3f)
}

// To returns the destination square.
func (m Move) To() Square {
	return Square((m >> 6) & 0x3f)
}

// Flag returns the move flag.
func (m Move) Flag() MoveFlag {
	return MoveFlag(m >> 12)
}

// IsPromotion reports whether the move promotes a pawn.
func (m Move) IsPromotion() bool {
	return m.Flag() >= FlagPromoKnight
}

// Promotion returns the piece type a pawn promotes to, or NoPieceType.
func (m Move) Promotion() PieceType {
	switch m.Flag() {
	case FlagPromoKnight:
		return Knight
	case FlagPromoBishop:
		return Bishop
	case FlagPromoRook:
		return Rook
	case FlagPromoQueen:
		return Queen
	default:
		return NoPieceType
	}
}

// IsNull reports whether the move is the zero value.
func (m Move) IsNull() bool {
	return m.From() == m.To()
}

// String returns long algebraic notation, e.g. "e2e4" or "e7e8q".
func (m Move) String() string {
	if m.IsNull() {
		return "0000"
	}
	s := m.From().String() + m.To().String()
	if m.IsPromotion() {
		s += string(Piece(m.Promotion()).Symbol())
	}
	return s
}
