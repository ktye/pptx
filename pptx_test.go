package pptx

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"testing"
)

func greyImage() image.Image {
	w, h := 500, 300
	im := image.NewGray(image.Rect(0, 0, w, h))
	draw.Draw(im, im.Bounds(), &image.Uniform{color.Gray{128}}, image.ZP, draw.Src)
	return im
}

// TestNew creates a new file out.pptx with one slide.
// In a second pass it opens the file again and adds 2 more slides.
func TestNew(t *testing.T) {
	var f File
	var err error

	// Create a new prentation using the template internal/data/minimal.pptx.
	f, err = New("out.pptx")
	if err != nil {
		t.Fatal(err)
	}

	// Add a single slide.
	s := Slide{
		TextBoxes: []TextBox{
			TextBox{
				X:    20 * MilliMeter,
				Y:    50 * MilliMeter,
				Text: "This is a text box.",
			},
		},
	}
	if err := f.Add(s); err != nil {
		t.Fatal(err)
	}

	// Close the file and open again.
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	// Open a second time and add a new slide.
	f, err = Open("out.pptx")
	if err != nil {
		t.Fatal("err")
	}

	// Add a slide with an image and 2 multiline textboxes
	// with custom fonts.
	s = Slide{
		Images: []Image{
			Image{
				X:     20 * MilliMeter,
				Y:     10 * MilliMeter,
				Image: greyImage(),
			},
		},
		TextBoxes: []TextBox{
			TextBox{
				X:    20 * MilliMeter,
				Y:    100 * MilliMeter,
				Text: "Normal textbox",
			},
			TextBox{
				X:    20 * MilliMeter,
				Y:    120 * MilliMeter,
				Text: "This is a multiline\nTextbox with\nCourier New at size 22.",
				Font: Font{
					Name: "Courier New",
					Size: 22,
				},
			},
		},
	}
	err = f.Add(s)
	if err != nil {
		t.Fatal(err)
	}

	// Add another slide to the open file.
	s.TextBoxes[0].Text += " at 10 pt."
	s.TextBoxes[0].Font.Size = 10
	err = f.Add(s)
	if err != nil {
		t.Fatal(err)
	}

	// Close the file.
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	// Delete the file, uncomment to keep it.
	if err := os.Remove("out.pptx"); err != nil {
		t.Fatal(err)
	}
}
