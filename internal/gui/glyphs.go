//go:build ebiten

package gui

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/danielriddell21/gambit/internal/assets"
	"github.com/danielriddell21/gambit/pkg/chess"
)

// solidGlyph maps a piece type to its filled Unicode chess glyph (the U+265A..F
// block). We use the solid glyphs for both colors and distinguish color by
// tint, since the outlined "white" glyphs render poorly on light squares.
var solidGlyph = map[chess.PieceType]rune{
	chess.King:   '♚',
	chess.Queen:  '♛',
	chess.Rook:   '♜',
	chess.Bishop: '♝',
	chess.Knight: '♞',
	chess.Pawn:   '♟',
}

// newFace builds a text face from the embedded font at the given pixel size.
func newFace(size float64) (*text.GoTextFace, error) {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(assets.Font))
	if err != nil {
		return nil, fmt.Errorf("ui: load font: %w", err)
	}
	return &text.GoTextFace{Source: src, Size: size}, nil
}
