package chess

import (
	"fmt"
	"strconv"
	"strings"
)

// StartingFEN is the FEN of the standard initial position.
const StartingFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// fenPieceType maps a FEN letter (case-insensitive) to a piece type.
var fenPieceType = map[byte]PieceType{
	'p': Pawn, 'n': Knight, 'b': Bishop, 'r': Rook, 'q': Queen, 'k': King,
}

// ParseFEN parses a FEN string into a Board.
func ParseFEN(fen string) (*Board, error) {
	fields := strings.Fields(fen)
	if len(fields) < 4 {
		return nil, fmt.Errorf("chess: FEN needs at least 4 fields, got %d", len(fields))
	}

	b := &Board{enPassant: NoSquare}

	if err := parseFENPlacement(b, fields[0]); err != nil {
		return nil, err
	}

	switch fields[1] {
	case "w":
		b.sideToMove = White
	case "b":
		b.sideToMove = Black
	default:
		return nil, fmt.Errorf("chess: invalid side to move %q", fields[1])
	}

	b.castling = parseFENCastling(fields[2])

	if fields[3] != "-" {
		sq, err := ParseSquare(fields[3])
		if err != nil {
			return nil, fmt.Errorf("chess: invalid en-passant square: %w", err)
		}
		b.enPassant = sq
	}

	b.halfMove = 0
	if len(fields) >= 5 {
		n, err := strconv.Atoi(fields[4])
		if err != nil {
			return nil, fmt.Errorf("chess: invalid halfmove clock %q", fields[4])
		}
		b.halfMove = n
	}

	b.fullMove = 1
	if len(fields) >= 6 {
		n, err := strconv.Atoi(fields[5])
		if err != nil {
			return nil, fmt.Errorf("chess: invalid fullmove number %q", fields[5])
		}
		b.fullMove = n
	}

	return b, nil
}

func parseFENPlacement(b *Board, placement string) error {
	ranks := strings.Split(placement, "/")
	if len(ranks) != 8 {
		return fmt.Errorf("chess: FEN placement needs 8 ranks, got %d", len(ranks))
	}
	// FEN lists ranks from 8 down to 1.
	for i, rankStr := range ranks {
		if err := parseFENRank(b, rankStr, 7-i); err != nil {
			return err
		}
	}
	return nil
}

// parseFENRank places the pieces of one FEN rank string onto the board.
func parseFENRank(b *Board, rankStr string, rank int) error {
	file := 0
	for j := 0; j < len(rankStr); j++ {
		ch := rankStr[j]
		if ch >= '1' && ch <= '8' {
			file += int(ch - '0')
			continue
		}
		pt, ok := fenPieceType[lower(ch)]
		if !ok {
			return fmt.Errorf("chess: invalid FEN piece %q", string(ch))
		}
		if file > 7 {
			return fmt.Errorf("chess: too many squares in rank %q", rankStr)
		}
		color := White
		if ch >= 'a' && ch <= 'z' {
			color = Black
		}
		b.setPiece(NewSquare(file, rank), MakePiece(color, pt))
		file++
	}
	if file != 8 {
		return fmt.Errorf("chess: rank %q does not fill 8 files", rankStr)
	}
	return nil
}

func parseFENCastling(s string) CastleRights {
	var c CastleRights
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case 'K':
			c |= WhiteKingside
		case 'Q':
			c |= WhiteQueenside
		case 'k':
			c |= BlackKingside
		case 'q':
			c |= BlackQueenside
		}
	}
	return c
}

func lower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// FEN returns the board's position as a FEN string.
func (b *Board) FEN() string {
	var sb strings.Builder

	for rank := 7; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < 8; file++ {
			p := b.squares[NewSquare(file, rank)]
			if p.IsEmpty() {
				empty++
				continue
			}
			if empty > 0 {
				sb.WriteByte(byte('0' + empty))
				empty = 0
			}
			sb.WriteRune(p.Symbol())
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
		if rank > 0 {
			sb.WriteByte('/')
		}
	}

	sb.WriteByte(' ')
	if b.sideToMove == White {
		sb.WriteByte('w')
	} else {
		sb.WriteByte('b')
	}

	sb.WriteByte(' ')
	sb.WriteString(castlingString(b.castling))

	sb.WriteByte(' ')
	sb.WriteString(b.enPassant.String())

	sb.WriteByte(' ')
	sb.WriteString(strconv.Itoa(b.halfMove))
	sb.WriteByte(' ')
	sb.WriteString(strconv.Itoa(b.fullMove))

	return sb.String()
}

func castlingString(c CastleRights) string {
	if c == 0 {
		return "-"
	}
	var sb strings.Builder
	if c.Has(WhiteKingside) {
		sb.WriteByte('K')
	}
	if c.Has(WhiteQueenside) {
		sb.WriteByte('Q')
	}
	if c.Has(BlackKingside) {
		sb.WriteByte('k')
	}
	if c.Has(BlackQueenside) {
		sb.WriteByte('q')
	}
	return sb.String()
}
