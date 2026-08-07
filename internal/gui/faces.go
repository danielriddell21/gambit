package gui

import (
	"fmt"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/danielriddell21/gambit/internal/assets"
	"github.com/danielriddell21/gambit/pkg/chess"
)

// solidGlyph maps a piece to the filled figurine that stands for it.
var solidGlyph = map[chess.PieceType]rune{
	chess.King:   '♚',
	chess.Queen:  '♛',
	chess.Rook:   '♜',
	chess.Bishop: '♝',
	chess.Knight: '♞',
	chess.Pawn:   '♟',
}

// Faces are the three sizes the board draws in: the piece figurines, the info
// bar, and the end-of-game banner.
type Faces struct {
	Piece  font.Face
	Bar    font.Face
	Banner font.Face
}

// LoadFaces renders the bundled font at the sizes a board of the given square
// size needs. It returns display-free faces, so the same typography serves a
// live window and a headless recording.
func LoadFaces(squareSize, barHeight int) (Faces, error) {
	src, err := opentype.Parse(assets.Font)
	if err != nil {
		return Faces{}, fmt.Errorf("ui: parse font: %w", err)
	}
	sq := float64(squareSize)
	at := func(size float64) (font.Face, error) {
		f, err := opentype.NewFace(src, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
		if err != nil {
			return nil, fmt.Errorf("ui: font face at %.1f: %w", size, err)
		}
		return f, nil
	}
	var f Faces
	if f.Piece, err = at(sq * 0.8); err != nil {
		return Faces{}, err
	}
	if f.Bar, err = at(float64(barHeight) * 0.3); err != nil {
		return Faces{}, err
	}
	if f.Banner, err = at(sq * 0.42); err != nil {
		return Faces{}, err
	}
	return f, nil
}

// BarHeight is the info bar's height for a given square size, floored so the
// two lines of text always fit.
func BarHeight(squareSize int) int {
	return max(squareSize*7/10, 46)
}
