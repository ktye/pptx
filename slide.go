package pptx

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

// Slide holds the content of a slide which can be added to the presentation.
// It supports TextBoxes and Images.
type Slide struct {
	TextBoxes []TextBox       // TextBoxes.
	Images    []Image         // Images will be encoded as png.
	Master    int             // Slide layout master id. Default is 1
	n         int             // Slide number
	name      string          // slide file name, e.g.: slide5.xml, if n is 5.
	rId       string          // relationship id of the slide, e.g. "rId9"
	id        string          // slice id in ppt/presentation.xml slide list, e.g. "256"
	xml       *etree.Document // slide xml tree.
}

// Add appends a slide to the presentation.
func (f *File) Add(s Slide) error {
	if f.numSlides == 0 {
		// Only count the first time the file is read.
		f.numSlides = f.slideCount()
	}
	f.numSlides++
	s.n = f.numSlides
	s.name = fmt.Sprintf("slide%d.xml", s.n)

	if err := f.addToContentTypes(s); err != nil {
		return err
	}

	if err := f.addToRelationships(&s); err != nil {
		return err
	}

	/*
		if err := f.addPngImage(s); err != nil {
			return err
		}
	*/
	if err := s.build(f); err != nil {
		return err
	}

	if err := f.addSlideFile(s); err != nil {
		return err
	}

	if err := f.addSlideRels(s); err != nil {
		return err
	}

	if err := f.addToPresentation(&s); err != nil {
		return err
	}

	// deb.Println("TODO: (f pptx.File) Add(s slide) is not finished.")

	return nil
}

// build builds the slide xml tree.
func (s *Slide) build(f *File) error {
	s.xml = minimalSlide()
	for i, tb := range s.TextBoxes {
		if err := s.addTextBox(tb, i); err != nil {
			return err
		}
	}
	for i, im := range s.Images {
		if err := s.addImageRef(im, i); err != nil {
			return err
		}
		if err := f.addImageFile(im, i, s.n); err != nil {
			return err
		}
	}
	return nil
}

// addToContentTypes adds the new slide reference to [Conent_Types].xml.
func (f *File) addToContentTypes(slide Slide) error {
	contentTypes := "[Content_Types].xml"
	if err := f.readXml(contentTypes); err != nil {
		return err
	}

	// Add an element to Types.Override with PartName="/ptt/slides/" + slideName and ContentType=...
	if xw, ok := f.m[contentTypes]; !ok {
		return fmt.Errorf("cannot find: %s", contentTypes)
	} else {
		x := xw.(*etree.Document)
		partName := "/ppt/slides/" + slide.name
		typesElement := x.SelectElement("Types")
		if typesElement == nil {
			return fmt.Errorf("[Content_Types].xml: Element does not exist: <Types...")
		}
		override := typesElement.CreateElement("Override")
		override.Attr = []etree.Attr{
			etree.Attr{Key: "PartName", Value: partName},
			etree.Attr{Key: "ContentType", Value: "application/vnd.openxmlformats-officedocument.presentationml.slide+xml"},
		}
	}
	return nil
}

// addToRelationships adds the new slide entry to ppt/_rels/presentation.xml.rels.
// The file has the form:
// 	<Relationships ...
//		<Relationship Id=rId3 Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/presProps" ...
//		<Relationship Id=rId7 Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/presProps" ...
// Create a new ID and add a relationship for the slide:
//		<Relationship Id=rId? Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/presProps" Target="slides/slide3.xml"/>
func (f *File) addToRelationships(slide *Slide) error {
	relFile := "ppt/_rels/presentation.xml.rels"
	if err := f.readXml(relFile); err != nil {
		return err
	}

	if xw, ok := f.m[relFile]; !ok {
		return fmt.Errorf("cannot find: %s", relFile)
	} else {
		x := xw.(*etree.Document)

		// Create a new relationship ID.
		m := make(map[string]bool)
		rootElement := x.SelectElement("Relationships")
		if rootElement == nil {
			return fmt.Errorf("%s: Cannot find <Relationships...", relFile)
		}
		n := 1
		for _, e := range rootElement.ChildElements() {
			n++
			if s := e.SelectAttrValue("Id", ""); s == "" {
				return fmt.Errorf("%s: unknown relationship entry: %s", relFile, s)
			} else {
				m[s] = true
			}
		}
		id := ""
		for i := 0; i < 10000; i++ {
			id = fmt.Sprintf("rId%d", slide.n+i)
			if _, ok := m[id]; !ok {
				break
			}
		}
		if id == "" {
			return fmt.Errorf("%s: Failed to create a unique ID", relFile)
		}
		// Store id
		slide.rId = id
		// Add the relation to the new slide.
		e := rootElement.CreateElement("Relationship")
		e.CreateAttr("Id", id)
		e.CreateAttr("Type", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide")
		e.CreateAttr("Target", "slides/"+slide.name)
	}
	return nil
}

// addSlideFile adds the slideN.xml file to the presentation.
func (f *File) addSlideFile(slide Slide) error {
	slideFile := "ppt/slides/" + slide.name
	if slide.xml == nil {
		return fmt.Errorf("there is no slide to be added.")
	}
	f.m[slideFile] = slide.xml
	return nil
}

// addSlideRels add the slide relation file ppt/slides/_rels/slide?.xml.rels
func (f *File) addSlideRels(slide Slide) error {
	relFile := fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", slide.n)
	var d etree.Document
	err := d.ReadFromString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
</Relationships>`)
	if err != nil {
		return fmt.Errorf("%s: Malformed template.", relFile)
	}

	// Add inside <Relationships>:
	// e.g.: <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
	rootElement := d.SelectElement("Relationships")
	if rootElement == nil {
		return fmt.Errorf("%s: Cannot find <Relationships...", relFile)
	}
	if slide.Master == 0 {
		slide.Master = 1
	}
	e := rootElement.CreateElement("Relationship")
	e.CreateAttr("Id", "rId1") // Is this always the id of layout 1?
	e.CreateAttr("Type", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout")
	e.CreateAttr("Target", fmt.Sprintf("../slideLayouts/slideLayout%d.xml", slide.Master))
	// Add relations for each image in the slide.
	for i := range slide.Images {
		e := rootElement.CreateElement("Relationship")
		e.CreateAttr("Id", fmt.Sprintf("rId%d", i+2)) // rId2...
		e.CreateAttr("Type", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image")
		e.CreateAttr("Target", fmt.Sprintf("../media/slide%dimage%d.png", slide.n, i))
	}
	// Add the file to the map.
	f.m[relFile] = &d
	return nil
}

// addToPresentation adds the slide reference to ppt/presentation.xml.
// <p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/ma
//	<p:sldIdLst>
//		<p:sldId id="256" r:id="rId7"/> // firt slide starts at id=256
//		<p:sldId id="257" r:id="rId8"/>
//		<p:sldId id="258" r:id="rId?"/>
//	</p:sldIdLst>
func (f *File) addToPresentation(slide *Slide) error {
	if slide.rId == "" {
		return fmt.Errorf("New slide has no rId.")
	}
	presentationFile := "ppt/presentation.xml"
	if err := f.readXml(presentationFile); err != nil {
		return err
	}

	if xw, ok := f.m[presentationFile]; !ok {
		return fmt.Errorf("cannot find: %s", presentationFile)
	} else {
		x := xw.(*etree.Document)
		rootElement := x.SelectElement("p:presentation")
		if rootElement == nil {
			return fmt.Errorf("%s: Cannot find <p:presentation...", presentationFile)
		}
		listElement := rootElement.SelectElement("p:sldIdLst")
		// Create slide id list if it does not exist (for the first slide).
		if listElement == nil {
			listElement = rootElement.CreateElement("p:sldIdLst")
		}
		// Get the next free id, starting at 256
		id := 256
		for _, e := range listElement.ChildElements() {
			if s := e.SelectAttrValue("id", ""); s == "" {
				return fmt.Errorf("%s: A slide in the list has no id.", presentationFile)
			} else {
				if idNum, err := strconv.Atoi(s); err != nil {
					return fmt.Errorf("%s: Cannot convert slide id to integer: %s", presentationFile, s)
				} else {
					if idNum >= id {
						id = idNum + 1 // Always have id more than the max id of the previous slides.
					}
				}
			}
		}
		// Store slice id and add it to the xml tree.
		slide.id = strconv.Itoa(id)
		e := listElement.CreateElement("p:sldId")
		e.CreateAttr("id", slide.id)
		e.CreateAttr("r:id", slide.rId)
	}
	return nil
}

// slideCount returns the number of slides in the presentation.
func (f File) slideCount() int {
	// The directory /ppt/slides contains a file called
	// _rels and slide files with the name slide1.xml, ...
	n := 0
	for _, v := range f.r.File {
		if strings.HasPrefix(v.Name, "ppt/slides/slide") {
			n++
		}
	}
	return n
}

// readXml populates the map f.m with an xml-etree from in zip input.
// f must already be loaded with Open() or New().
func (f *File) readXml(filePath string) error {
	if _, ok := f.m[filePath]; ok {
		return nil // File is already read.
	}
	for _, v := range f.r.File {
		if v.Name == filePath {
			if r, err := v.Open(); err != nil {
				return err
			} else {
				d := etree.NewDocument()
				if _, err := d.ReadFrom(r); err != nil {
					r.Close()
					return fmt.Errorf("%s: %s", filePath, err)
				} else {
					if err := r.Close(); err != nil {
						return fmt.Errorf("%s: %s", filePath, err)
					}
					if f.m == nil {
						f.m = make(map[string]io.WriterTo)
					}
					f.m[filePath] = d
					return nil
				}
			}
		}
	}
	return fmt.Errorf("%s: file does not exist in input pptx.", filePath)
}

// minimalSlide returns the xml tree of a minimal slide without content.
func minimalSlide() *etree.Document {
	var d etree.Document

	err := d.ReadFromString(`
		<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
		<p:cSld>
		<p:spTree>
		<p:nvGrpSpPr>
		<p:cNvPr id="1" name=""/>
		<p:cNvGrpSpPr/>
		<p:nvPr/>
		</p:nvGrpSpPr>
		<p:grpSpPr>
		<p:grpSpPr>
		<a:xfrm>
		<a:off x="0" y="0"/>
		<a:ext cx="0" cy="0"/>
		<a:chOff x="0" y="0"/>
		<a:chExt cx="0" cy="0"/>
		</a:xfrm>
		</p:grpSpPr>
		</p:spTree>
		</p:cSld>
		<p:clrMapOvr>
		<a:masterClrMapping/>
		</p:clrMapOvr>
		</p:sld>
	   	`)

	// This is a bigger version // TODO: remove.
	/*
			err := d.ReadFromString(`
		<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
		<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
		  <p:cSld>
		    <p:spTree>
		      <p:nvGrpSpPr>
		        <p:cNvPr id="1" name=""/>
		        <p:cNvGrpSpPr/>
		        <p:nvPr/>
		      </p:nvGrpSpPr>
		      <p:grpSpPr>
		        <a:xfrm>
		          <a:off x="0" y="0"/>
		          <a:ext cx="0" cy="0"/>
		          <a:chOff x="0" y="0"/>
		          <a:chExt cx="0" cy="0"/>
		        </a:xfrm>
		      </p:grpSpPr>
		    </p:spTree>
		    <p:extLst>
		      <p:ext uri="{BB962C8B-B14F-4D97-AF65-F5344CB8AC3E}">
		        <p14:creationId xmlns:p14="http://schemas.microsoft.com/office/powerpoint/2010/main" val="235730354"/>
		      </p:ext>
		    </p:extLst>
		  </p:cSld>
		  <p:clrMapOvr>
		    <a:masterClrMapping/>
		  </p:clrMapOvr>
		</p:sld>
		`)
	*/
	if err != nil {
		panic(err) // This should parse without error.
	}
	return &d
}
