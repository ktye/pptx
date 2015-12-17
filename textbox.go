package pptx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

// A TextBox can be added to a slide.
type TextBox struct {
	X, Y  Dimension // Position
	Text  string    // Lines are split at "\n".
	Title bool      // Mark this textbox as slide title.
	Font  Font      // Can be unspecified for defaults.
	xml   *etree.Document
}

// Font specifies the font used in the text box.
type Font struct {
	Name string  // E.g. "Courier New"
	Size float64 // Font size.
}

// addTextBox adds a textbox the the slide's xml tree.
// The textbox is appended to the slide at the path:
// <p:sld...><p:cSld><p:spTree>
func (s *Slide) addTextBox(tb TextBox, tbNum int) error {
	if err := tb.build(tbNum); err != nil {
		return err
	}
	root := s.xml.Root()
	if root == nil {
		return fmt.Errorf("Cannot find root element")
	}
	var spTree *etree.Element
	if spTree = root.FindElement("p:cSld/p:spTree"); spTree == nil {
		return fmt.Errorf("Cannot find spTree")
	}
	tbRoot := tb.xml.Root()
	if tbRoot == nil {
		return fmt.Errorf("Cannot find tbRoot element")
	}
	spTree.Child = append(spTree.Child, tbRoot)
	return nil
}

// build creates the xml tree of a text box.
func (tb *TextBox) build(tbNum int) error {
	numStr := strconv.Itoa(tbNum + 1)
	x := strconv.FormatUint(uint64(tb.X), 10)
	y := strconv.FormatUint(uint64(tb.Y), 10)
	template := `<p:sp>
<p:nvSpPr>
<p:cNvPr id="` + strconv.Itoa(tbNum+2) + `" name="TextBox ` + numStr + `"/>
<p:cNvSpPr txBox="1"/>
<p:nvPr/>
</p:nvSpPr>
<p:spPr>
<a:xfrm>
<a:off x="` + x + `" y="` + y + `"/>
<a:ext cx="360000" cy="360000"/>
</a:xfrm>
</p:spPr>
<p:txBody>
<a:bodyPr wrap="none" rtlCol="0">
</a:bodyPr>
</p:txBody>
</p:sp>`
	tb.xml = etree.NewDocument()
	if err := tb.xml.ReadFromString(template); err != nil {
		return err
	}
	txBody := tb.xml.FindElement("p:sp/p:txBody")
	if txBody == nil {
		return fmt.Errorf("cannot find txBody")
	}
	if tb.Title {
		nvPr := tb.xml.FindElement("p:sp/p:nvSpPr/p:nvPr")
		if nvPr == nil {
			return fmt.Errorf("cannot find p:nvPr")
		}
		ph := nvPr.CreateElement("p:ph")
		ph.CreateAttr("type", "title")
	}
	lines := strings.Split(tb.Text, "\n")
	for _, line := range lines {
		txBody.Child = append(txBody.Child, tb.buildLine(line).Root())
	}
	return nil
}

// buildLine returns the xml tree for a text box line.
func (tb TextBox) buildLine(line string) *etree.Document {
	doc := etree.NewDocument()
	ap := doc.CreateElement("a:p")
	ar := ap.CreateElement("a:r")
	if tb.Font.Size > 0 || tb.Font.Name != "" {
		arPr := ar.CreateElement("a:rPr")
		if s := int(tb.Font.Size * 100); s > 0 {
			arPr.Attr = []etree.Attr{
				etree.Attr{Key: "sz", Value: strconv.Itoa(s)},
			}
		}
		if tb.Font.Name != "" {
			alatin := arPr.CreateElement("a:latin")
			alatin.Attr = []etree.Attr{
				etree.Attr{Key: "typeface", Value: tb.Font.Name},
			}
			acs := arPr.CreateElement("a:cs")
			acs.Attr = []etree.Attr{
				etree.Attr{Key: "typeface", Value: tb.Font.Name},
			}
		}
	}
	at := ar.CreateElement("a:t")
	at.CreateCharData(line)
	return doc
}

// A textbox only needs an addition to ppt/slides/slideN.xml
// The node <p:sp> should be inserted to the path:
// <p:sld...><p:cSld><p:spTree> after <p:grpSpPr>

/*
// Textbox with default font:
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="2" name="TextBox 1"/>
          <p:cNvSpPr txBox="1"/>
          <p:nvPr/>
        </p:nvSpPr>
        <p:spPr>
          <a:xfrm>
            <a:off x="1115616" y="980728"/>
            <a:ext cx="702436" cy="369332"/>
          </a:xfrm>
        </p:spPr>
        <p:txBody>
          <a:bodyPr wrap="none" rtlCol="0">
          </a:bodyPr>
          <a:p>
            <a:r>
              <a:t>alpha beta gamma</a:t>
            </a:r>
          </a:p>
        </p:txBody>
      </p:sp>

// Textbox with "Courier New" size 20: 2 lines:
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="5" name="TextBox 4"/>
          <p:cNvSpPr txBox="1"/>
          <p:nvPr/>
        </p:nvSpPr>
        <p:spPr>
          <a:xfrm>
            <a:off x="1466834" y="4437112"/>
            <a:ext cx="2114681" cy="646331"/>
          </a:xfrm>
        </p:spPr>
        <p:txBody>
          <a:bodyPr wrap="none" rtlCol="0">
          </a:bodyPr>
          <a:p>
            <a:r>
              <a:rPr sz="2000">
                <a:latin typeface="Courier New"/>
                <a:cs typeface="Courier New"/>
              </a:rPr>
              <a:t>textboxCouriermitäüßere</a:t>
            </a:r>
          </a:p>
          <a:p>
            <a:r>
              <a:rPr sz="2000">
                <a:latin typeface="Courier New"/>
                <a:cs typeface="Courier New"/>
              </a:rPr>
              <a:t>zeilezwei</a:t>
            </a:r>
          </a:p>
        </p:txBody>
      </p:sp>
*/
