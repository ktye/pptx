package pptx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

// ItemBox is a textbox with nested items.
type ItemBox struct {
	X, Y, Width, Height Dimension
	// Font  Font
	Items []Item
	xml   *etree.Document
}

// Item is a line of text with an indentation level.
type Item struct {
	Level int
	Text  string
}

// SimpleItems splits a string at newlines into items.
// The item level is the number of leading dashes.
func SimpleItems(s string) []Item {
	lines := strings.Split(s, "\n")
	items := make([]Item, len(lines))
	for i, line := range lines {
		items[i].Level, items[i].Text = splitLevelText(line)
	}
	return items
}

func splitLevelText(s string) (int, string) {
	level := 0
	for _, c := range s {
		if c == '-' {
			level++
		} else {
			break
		}
	}
	return level, s[level:]
}

// addItemBox adds an ItemBox to the slide's xml tree.
func (s *Slide) addItemBox(ib ItemBox, ibNum int) error {
	if err := ib.build(ibNum); err != nil {
		return err
	}
	root := s.xml.Root()
	if root == nil {
		return fmt.Errorf("cannot find root element")
	}

	var spTree *etree.Element
	if spTree = root.FindElement("p:cSld/p:spTree"); spTree == nil {
		return fmt.Errorf("Cannot find spTree")
	}
	ibRoot := ib.xml.Root()
	if ibRoot == nil {
		return fmt.Errorf("Cannot find ibRoot element")
	}
	spTree.Child = append(spTree.Child, ibRoot)
	return nil
}

// build creates the xml tree of an item box.
func (ib *ItemBox) build(ibNum int) error {
	numStr := strconv.Itoa(ibNum + 1)
	x := strconv.FormatUint(uint64(ib.X), 10)
	y := strconv.FormatUint(uint64(ib.Y), 10)
	w := strconv.FormatUint(uint64(ib.Width), 10)
	h := strconv.FormatUint(uint64(ib.Height), 10)
	template := `<p:sp>
<p:nvSpPr>
<p:cNvPr id="` + strconv.Itoa(ibNum+2) + `" name="ItemBox ` + numStr + `"/>
<p:cNvSpPr/>
<p:nvPr>
<p:ph idx="1"/>
</p:nvPr>
</p:nvSpPr>
<p:spPr>
<a:xfrm>
<a:off x="` + x + `" y="` + y + `"/>
<a:ext cx="` + w + `" cy="` + h + `"/>
</a:xfrm>
</p:spPr>
<p:txBody>
<a:bodyPr/>
<a:lstStyle/>
</p:txBody>
</p:sp>`
	ib.xml = etree.NewDocument()
	if err := ib.xml.ReadFromString(template); err != nil {
		return err
	}
	txBody := ib.xml.FindElement("p:sp/p:txBody")
	if txBody == nil {
		return fmt.Errorf("cannot find txBody")
	}
	for _, line := range ib.Items {
		txBody.Child = append(txBody.Child, ib.buildLine(line).Root())
	}
	return nil
}

// buildLine returns the xml tree for a single item line.
func (ib ItemBox) buildLine(item Item) *etree.Document {
	doc := etree.NewDocument()
	ap := doc.CreateElement("a:p")
	apPr := ap.CreateElement("a:pPr")
	apPr.Attr = []etree.Attr{etree.Attr{Key: "lvl", Value: strconv.Itoa(item.Level)}}
	ar := ap.CreateElement("a:r")
	at := ar.CreateElement("a:t")
	at.CreateCharData(item.Text)
	return doc
}
