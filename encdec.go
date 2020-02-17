package pptx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"io/ioutil"
	"strings"

	"github.com/ktye/pptx/pptxt"
)

// EncodeSlides encodes Slides to a text form.
func EncodeSlides(slides []Slide, w io.Writer) (e error) {
	for _, s := range slides {
		e = encodeSlide(s, w, e)
	}
	return e
}

// DecodeSlides decodes a flat text description of a presentation.
// It does _not_ decode a pptx file.
func DecodeSlides(r io.Reader) ([]Slide, error) {
	b, e := ioutil.ReadAll(r)
	if e != nil {
		return nil, e
	}
	lr := &lineReader{Buffer: bytes.NewBuffer(b)}
	var slides []Slide
	for {
		s, e := decodeSlide(lr)
		if e == nil {
			slides = append(slides, s)
		} else if e == io.EOF {
			return append(slides, s), nil
		} else {
			return nil, e
		}
	}
}

type encoder interface {
	encode(io.Writer, error) error
}

func encodeSlide(s Slide, w io.Writer, e error) error {
	if e != nil {
		return e
	}
	fmt.Fprintf(w, "Slide\n Master %d\n", s.Master)
	for _, x := range s.TextBoxes {
		e = x.encode(w, e)
	}
	for _, x := range s.ItemBoxes {
		e = x.encode(w, e)
	}
	for _, x := range s.Images {
		e = x.encode(w, e)
	}
	return e
}
func decodeSlide(r pptxt.LineReader) (s Slide, e error) {
	e = expect(r, "Slide", e)
	if e != nil {
		return s, e
	}
	for {
		b, err := r.Peek()
		if err != nil { // maybe EOF
			return s, err
		}
		l := string(b)
		if strings.HasPrefix(l, "TextBox") {
			if tb, err := decodeTextBox(r); err != nil {
				return s, err
			} else {
				s.TextBoxes = append(s.TextBoxes, tb)
			}
		} else if strings.HasPrefix(l, "ItemBox") {
			if ib, err := decodeItemBox(r); err != nil {
				return s, err
			} else {
				s.ItemBoxes = append(s.ItemBoxes, ib)
			}
		} else if strings.HasPrefix(l, "Image") {
			if m, err := decodeImage(r); err != nil {
				return s, err
			} else {
				s.Images = append(s.Images, m)
			}
		} else if strings.HasPrefix(l, "Master") {
			if b, err := r.ReadLine(); err != nil {
				return s, err
			} else {
				if _, err := fmt.Sscanf(string(b), "Master %d", &s.Master); err != nil {
					return s, err
				}
			}
		} else if strings.HasPrefix(l, "Slide") {
			return s, nil
		}
	}
}

func (t TextBox) encode(w io.Writer, e error) error {
	if e != nil {
		return e
	}
	fmt.Fprintln(w, " TextBox")
	fmt.Fprintf(w, "  Position [%d, %d]\n", t.X, t.Y)
	for _, l := range t.Lines {
		fmt.Fprintf(w, "Line")
		for _, le := range l {
			r, g, b := uint32(0), uint32(0), uint32(0)
			if le.Color != nil {
				r, g, b, _ = le.Color.RGBA()
			}
			fmt.Fprintf(w, " %02x%02x%02x %q", r>>8, g>>8, b>>8, le.Text)
		}
		w.Write([]byte{'\n'})
	}
	e = js(w, "  Title", t.Title, e)
	e = js(w, "  Font", t.Font, e)
	return e
}
func decodeTextBox(r pptxt.LineReader) (t TextBox, e error) {
	e = expect(r, "TextBox", e)
	if e != nil {
		return t, e
	}
	var xy [2]Dimension
	e = sj(r, "Position", &xy, e)
	t.X = xy[0]
	t.Y = xy[1]
	var l []byte
	for {
		l, e = r.Peek()
		if e != nil {
			return t, e
		} else if bytes.HasPrefix(l, []byte("Line")) {
			if l, err := decodeLine(r); err != nil {
				return t, err
			} else {
				t.Lines = append(t.Lines, l)
			}
		} else {
			break
		}
	}
	e = sj(r, "Title", &t.Title, e)
	e = sj(r, "Font", &t.Font, e)
	return t, e
}
func decodeLine(r pptxt.LineReader) (l Line, e error) {
	b, err := r.ReadLine()
	if err != nil {
		return l, err
	}
	b = bytes.TrimPrefix(b, []byte("Line "))
	f := bytes.NewReader(b)
	for {
		var le LineElement
		var r, g, b uint32
		_, e := fmt.Fscanf(f, "%02x%02x%02x %q", &r, &g, &b, &le.Text)
		if e == io.EOF {
			return l, nil
		} else if e != nil {
			return l, e
		}
		le.Color = color.RGBA{uint8(r), uint8(g), uint8(b), 0}
		l = append(l, le)
	}
}

func (b ItemBox) encode(w io.Writer, e error) error {
	if e != nil {
		return e
	}
	fmt.Fprintln(w, " ItemBox")
	fmt.Fprintf(w, "  Position [%d, %d, %d, %d]\n", b.X, b.Y, b.Width, b.Height)
	for _, it := range b.Items {
		e = js(w, "  Item", it, e)
	}
	return e
}
func decodeItemBox(r pptxt.LineReader) (t ItemBox, e error) {
	e = expect(r, "ItemBox", e)
	var xywh [4]Dimension
	e = sj(r, "Position", &xywh, e)
	t.X = xywh[0]
	t.Y = xywh[1]
	t.Width = xywh[2]
	t.Height = xywh[3]
	for {
		if l, err := r.Peek(); err != nil {
			return t, err
		} else if bytes.HasPrefix(l, []byte("Item")) {
			var it Item
			e = sj(r, "Item", &it, e)
			if e != nil {
				return t, e
			} else {
				t.Items = append(t.Items, it)
			}
		} else {
			return t, nil
		}
	}
}

func (b Image) encode(w io.Writer, e error) error {
	if e != nil {
		return e
	}
	fmt.Fprintf(w, " Image\n  Position [%d, %d]\n", b.X, b.Y)
	var buf bytes.Buffer
	e = b.Image.Encode(&buf)
	if e != nil {
		return e
	}
	lines := bytes.Split(buf.Bytes(), []byte{'\n'}) // Indent
	for i := range lines {
		if _, e = w.Write(append([]byte{' ', ' '}, lines[i]...)); e != nil {
			return e
		}
	}
	w.Write([]byte{'\n'})
	return e
}
func decodeImage(r pptxt.LineReader) (im Image, e error) {
	var xy [2]Dimension
	e = expect(r, "Image", e)
	e = sj(r, "Position", &xy, e)
	im.X = xy[0]
	im.Y = xy[1]
	if e != nil {
		return im, e
	}
	if l, err := r.Peek(); err != nil {
		return im, err
	} else if imageDecoders == nil {
		return im, fmt.Errorf("no image decoders are registered")
	} else {
		s := string(l)
		for _, d := range imageDecoders {
			if mag := d.Magic(); len(mag) == 0 {
				return im, fmt.Errorf("registered image decoder has no magic")
			} else if strings.HasPrefix(s, mag) {
				im.Image, e = d.Decode(r)
				if e != nil {
					return im, fmt.Errorf("line %d: %s", r.LineNumber(), e)
				}
				return im, e
			}
		}
		return im, fmt.Errorf("unknown image decoder: %s", s)
	}
}
func js(w io.Writer, name string, v interface{}, e error) error {
	if e != nil {
		return e
	}
	var b []byte
	b, e = json.Marshal(v)
	if e != nil {
		return e
	}
	fmt.Fprintf(w, "%s %s\n", name, string(b))
	return nil
}
func sj(r pptxt.LineReader, name string, v interface{}, e error) error {
	if e != nil {
		return e
	}
	b, e := r.ReadLine()
	if e != nil {
		return e
	}
	if bytes.HasPrefix(b, []byte(name)) == false {
		return fmt.Errorf("line %d: expected %q", r.LineNumber(), name)
	} else {
		b = b[len(name):]
	}
	e = json.Unmarshal(b, v)
	if e != nil {
		return fmt.Errorf("line %d: %s", r.LineNumber(), e)
	}
	return nil
}
func expect(r pptxt.LineReader, s string, e error) error {
	if e != nil {
		return e
	}
	if b, e := r.ReadLine(); e != nil {
		return e
	} else if string(b) != s {
		return fmt.Errorf("line %d: expected: %q got %q", r.LineNumber(), s, string(b))
	}
	return nil
}

type lineReader struct {
	*bytes.Buffer
	peek []byte
	lino int
}

func (l *lineReader) ReadLine() (c []byte, e error) {
	if l.peek != nil {
		c = l.peek
		l.peek = nil
		return c, nil
	}
	l.lino++
	c, e = l.Buffer.ReadBytes('\n')
	if e == io.EOF {
		return c, io.EOF
	} else if e != nil {
		return c, fmt.Errorf("line %d: %s", l.lino, e)
	}
	return bytes.Trim(c, "\r\n\t "), e
}
func (l *lineReader) Peek() (c []byte, e error) {
	c, e = l.ReadLine()
	l.peek = c
	return c, e
}
func (l *lineReader) LineNumber() int {
	return l.lino
}

var imageDecoders []pptxt.Raster

// RegisterImageDecoder registers a custom Raster image serialization format.
func RegisterImageDecoder(r pptxt.Raster) {
	imageDecoders = append(imageDecoders, r)
}
func init() {
	RegisterImageDecoder(GoImage{})
	RegisterImageDecoder(PngFile{})
}
