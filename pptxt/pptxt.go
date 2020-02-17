// Pptxt defines the interfaces for text decoding/encoding of pptx.Raster images.
// It is it's own package to unbundle the implementor from importing all of pptx.
// e.g. ktye/plot.
package pptxt

import (
	"image"
	"io"
)

// Raster defines methods for a raster image that is able to encode/decode itself from text form.
type Raster interface {
	Raster() (image.Image, error)
	// The following may be implemented to support textual serialization (pptadd)
	Encode(w io.Writer) error
	Magic() string
	Decode(r LineReader) (Raster, error)
}

// LineReader can read line by line.
// Lines are trimmed of whitespace: "\r\n\t "
type LineReader interface {
	ReadLine() ([]byte, error)
	Peek() ([]byte, error)
	LineNumber() int
}
