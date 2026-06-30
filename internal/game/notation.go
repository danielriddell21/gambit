package game

import (
	"strings"

	"github.com/danielriddell21/gambit/pkg/chess"
)

func SAN(b *chess.Board, m chess.Move) string {
	switch m.Flag() {
	case chess.FlagCastleKingside:
		return "O-O" + checkSuffix(b, m)
	case chess.FlagCastleQueenside:
		return "O-O-O" + checkSuffix(b, m)
	}

	piece := b.PieceAt(m.From())
	pt := piece.Type()
	capture := !b.PieceAt(m.To()).IsEmpty() || m.Flag() == chess.FlagEnPassant

	var sb strings.Builder
	if pt == chess.Pawn {
		if capture {
			sb.WriteByte(byte('a' + m.From().File()))
		}
	} else {
		sb.WriteRune(chess.MakePiece(chess.White, pt).Symbol())
		sb.WriteString(disambiguation(b, m, pt))
	}

	if capture {
		sb.WriteByte('x')
	}
	sb.WriteString(m.To().String())

	if m.IsPromotion() {
		sb.WriteByte('=')
		sb.WriteRune(chess.MakePiece(chess.White, m.Promotion()).Symbol())
	}

	sb.WriteString(checkSuffix(b, m))
	return sb.String()
}

func disambiguation(b *chess.Board, m chess.Move, pt chess.PieceType) string {
	from := m.From()
	var others []chess.Square
	for _, mv := range b.LegalMoves() {
		if mv.To() == m.To() && mv.From() != from && b.PieceAt(mv.From()).Type() == pt {
			others = append(others, mv.From())
		}
	}
	if len(others) == 0 {
		return ""
	}

	sameFile, sameRank := false, false
	for _, o := range others {
		if o.File() == from.File() {
			sameFile = true
		}
		if o.Rank() == from.Rank() {
			sameRank = true
		}
	}

	switch {
	case !sameFile:
		return string(rune('a' + from.File()))
	case !sameRank:
		return string(rune('1' + from.Rank()))
	default:
		return string(rune('a'+from.File())) + string(rune('1'+from.Rank()))
	}
}

func checkSuffix(b *chess.Board, m chess.Move) string {
	nb := b.ApplyMove(m)
	if !nb.InCheck() {
		return ""
	}
	if nb.IsCheckmate() {
		return "#"
	}
	return "+"
}
