//go:build ebiten

package gui

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/danielriddell21/gambit/internal/assets"
	"github.com/danielriddell21/gambit/pkg/chess"
)

var solidGlyph = map[chess.PieceType]rune{
	chess.King:   '♚',
	chess.Queen:  '♛',
	chess.Rook:   '♜',
	chess.Bishop: '♝',
	chess.Knight: '♞',
	chess.Pawn:   '♟',
}

func newFace(size float64) (*text.GoTextFace, error) {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(assets.Font))
	if err != nil {
		return nil, fmt.Errorf("ui: load font: %w", err)
	}
	return &text.GoTextFace{Source: src, Size: size}, nil
}
