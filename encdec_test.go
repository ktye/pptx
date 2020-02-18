package pptx

import (
	"bytes"
	"fmt"
	"testing"
)

func TestEncDec(t *testing.T) {
	slds := []Slide{exampleSlide(1), exampleSlide(2)}
	b1 := testEncode(t, slds)
	if string(b1) != twoSlides {
		fmt.Println(string(b1))
		fmt.Println(twoSlides)
		t.Fatalf("text encoding mismatches")
	}
	// os.Stdout.Write(b1)
	b2 := testEncode(t, testDecode(t, []byte(twoSlides)))
	if string(b1) != string(b2) {
		t.Fatalf("encodings differ:\n%s\n%s", b1, b2)
	}
}
func testEncode(t *testing.T, slds []Slide) []byte {
	var b bytes.Buffer
	if e := EncodeSlides(slds, &b); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func testDecode(t *testing.T, b []byte) []Slide {
	s, e := DecodeSlides(bytes.NewReader(b))
	if e != nil {
		t.Fatal(e)
	}
	return s
}

const twoSlides = `Slide
 Master 0
 TextBox
  Position [1080000, 720000]
   Line 000000 "Slide 1: alpha beta gamma"
  Title true
  Font {"Name":"","Size":0}
 TextBox
  Position [1080000, 2160000]
   Line 000000 "Das ist TextBox 2 in 22 pt"
  Title false
  Font {"Name":"","Size":22}
 TextBox
  Position [1080000, 3240000]
   Line 000000 "Das ist TextBox 3 in 22 pt Courier New"
   Line 000000 "und noch eine Zeile."
  Title false
  Font {"Name":"Courier New","Size":22}
 Image
  Position [2160000, 720000]
  GoImage iVBORw0KGgoAAAANSUhEUgAAAfQAAAEsCAAAAAAQ0B4UAAACxUlEQVR4nOzRUQkAIBTAQBGDG90Wvo/dJRjs3EXNng7gP9ODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA96AQAA///4wgLbSdbvDwAAAABJRU5ErkJggg==
Slide
 Master 0
 TextBox
  Position [1080000, 720000]
   Line 000000 "Slide 2: alpha beta gamma"
  Title true
  Font {"Name":"","Size":0}
 TextBox
  Position [1080000, 2160000]
   Line 000000 "Das ist TextBox 2 in 22 pt"
  Title false
  Font {"Name":"","Size":22}
 TextBox
  Position [1080000, 3240000]
   Line 000000 "Das ist TextBox 3 in 22 pt Courier New"
   Line 000000 "und noch eine Zeile."
  Title false
  Font {"Name":"Courier New","Size":22}
 Image
  Position [2160000, 720000]
  GoImage iVBORw0KGgoAAAANSUhEUgAAAfQAAAEsCAAAAAAQ0B4UAAACxUlEQVR4nOzRUQkAIBTAQBGDG90Wvo/dJRjs3EXNng7gP9ODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA8yPcj0INODTA96AQAA///4wgLbSdbvDwAAAABJRU5ErkJggg==
`
