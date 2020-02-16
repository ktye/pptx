package pptx

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"testing"

	"github.com/beevik/etree"
)

// printXml prints an xml file from the map or archive indented.
func (f File) printXml(filePath string) error {
	if err := f.readXml(filePath); err != nil {
		return err
	}
	if xw, ok := f.m[filePath]; !ok {
		return fmt.Errorf("%s: file has not been added to the map!", filePath)
	} else {
		x := xw.(*etree.Document)
		x.Indent(2)
		fmt.Printf("%s:\n", filePath)
		x.WriteTo(os.Stdout)
		fmt.Println()
		return nil
	}
}

// printPath prints a file in the archive with a given path.
func (f File) printPath(filePath string) error {
	fmt.Printf("print %s:\n", filePath)
	for _, v := range f.r.File {
		if v.Name == filePath {
			if r, err := v.Open(); err != nil {
				return err
			} else {
				if _, err := io.Copy(os.Stdout, r); err != nil {
					return err
				}
				if err := r.Close(); err != nil {
					return err
				}
				fmt.Println()
				return nil
			}
		}
	}
	return fmt.Errorf("file %s does not exist", filePath)
}

type greyImage struct{ NoExchange }

func (g greyImage) Raster() (image.Image, error) {
	w, h := 500, 300
	im := image.NewGray(image.Rect(0, 0, w, h))
	draw.Draw(im, im.Bounds(), &image.Uniform{color.Gray{128}}, image.ZP, draw.Src)
	return im, nil
}

func TestAppend(t *testing.T) {
	if f, err := Open("minimal.pptx"); err != nil {
		t.Fatal(err)
	} else {
		s := Slide{
			Images: []Image{
				Image{
					X:     60 * MilliMeter,
					Y:     20 * MilliMeter,
					Image: greyImage{},
				},
			},
			TextBoxes: []TextBox{
				TextBox{
					X:     30 * MilliMeter,
					Y:     20 * MilliMeter,
					Lines: SimpleLines("alpha beta gamma"),
					Title: true,
				},
				TextBox{
					X:     30 * MilliMeter,
					Y:     60 * MilliMeter,
					Lines: SimpleLines("Das ist TextBox 2 in 22 pt"),
					Font:  Font{Size: 22},
				},
				TextBox{
					X:     30 * MilliMeter,
					Y:     90 * MilliMeter,
					Lines: SimpleLines("Das ist TextBox 3 in 22 pt Courier New\nund noch eine Zeile."),
					Font:  Font{Size: 22, Name: "Courier New"},
				},
			},
		}
		if err := f.Add(s); err != nil {
			t.Fatal(err)
		}
		if err := f.Add(s); err != nil {
			s.TextBoxes[0].Lines = SimpleLines("this is page 2")
			t.Fatal(err)
		}

		// Add another slide.
		s = Slide{
			Images: []Image{
				Image{
					X:     60 * MilliMeter,
					Y:     20 * MilliMeter,
					Image: greyImage{},
				},
			},
			TextBoxes: []TextBox{
				TextBox{
					X:     30 * MilliMeter,
					Y:     90 * MilliMeter,
					Lines: SimpleLines("another page has been added."),
					Font:  Font{Size: 22, Name: "Courier New"},
				},
			},
		}
		if err := f.Add(s); err != nil {
			t.Fatal(err)
		}
	}
}
