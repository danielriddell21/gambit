package assets

import _ "embed"

// Font is the DejaVu Sans TTF, which includes the Unicode chess glyphs
// (U+2654..U+265F). DejaVu Sans is freely redistributable.
//
//go:embed DejaVuSans.ttf
var Font []byte
