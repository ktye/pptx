package pptx

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/ktye/pptx/pptxt"
)

// What should be the image's dimensions?
// When inserting directly, it seems to use 96 dpi, or maybe depending on the presentation setup.
// However optimal resolution should depend on the screen resolution, which is normally unknown.

// An Image can be added to a Slide.
type Image struct {
	X, Y  Dimension
	Image pptxt.Raster
}

// PngFile is the file path of a raster image
// It implements a serializable Raster.
type PngFile struct {
	Path string
	im   image.Image
}

func (p PngFile) Raster() (m image.Image, e error) {
	if p.im == nil {
		p, e = p.load()
	}
	return p.im, e
}
func (p PngFile) Encode(w io.Writer) error { _, e := fmt.Fprintf(w, "File %s\n", p.Path); return e }
func (p PngFile) Magic() string            { return "File" }
func (p PngFile) Decode(r pptxt.LineReader) (pptxt.Raster, error) {
	l, e := r.ReadLine()
	if e != nil {
		return nil, e
	}
	s := string(l)
	if strings.HasPrefix(s, "File ") == false {
		return nil, fmt.Errorf(`expected: File "path/to/file"`)
	}
	s = strings.TrimPrefix(s, "File ")
	p.Path = s
	return p.load()
}
func (p PngFile) load() (PngFile, error) {
	if f, err := os.Open(p.Path); err != nil {
		return p, err
	} else {
		defer f.Close()
		p.im, err = png.Decode(f)
		return p, err
	}
}

// GoImage is a serializable Raster image that can be any image.Image.
type GoImage struct {
	image.Image
}

func (g GoImage) Raster() (image.Image, error) { return g.Image, nil }
func (g GoImage) Encode(w io.Writer) error {
	w.Write([]byte("GoImage "))
	enc := base64.NewEncoder(base64.StdEncoding, w)
	if e := png.Encode(enc, g.Image); e != nil {
		return e
	}
	enc.Close()
	return nil
}
func (g GoImage) Magic() string { return "GoImage" }
func (g GoImage) Decode(r pptxt.LineReader) (pptxt.Raster, error) {
	l, e := r.ReadLine()
	if e != nil {
		return nil, e
	}
	var m image.Image
	m, e = png.Decode(base64.NewDecoder(base64.StdEncoding, bytes.NewReader(bytes.TrimPrefix(l, []byte("GoImage ")))))
	return GoImage{m}, e
}

// NoExchange can be embedded in a Raster that does not support serialization.
type NoExchange struct{}

func (n NoExchange) Encode(w io.Writer) error                        { return n }
func (n NoExchange) Magic() string                                   { return n.Error() }
func (n NoExchange) Decode(r pptxt.LineReader) (pptxt.Raster, error) { return nil, n }
func (n NoExchange) Error() string                                   { return "this image type is not serializable" }

// addImageRef adds the image reference to the the slide's xml tree.
// The image reference is appended to the slide at the path:
// <p:sld...><p:cSld><p:spTree>
func (s *Slide) addImageRef(im Image, imageNum int) (image.Image, error) {
	xml, m, err := im.build(imageNum)
	if err != nil {
		return m, err
	}
	root := s.xml.Root()
	if root == nil {
		return m, fmt.Errorf("Cannot find root element")
	}
	var spTree *etree.Element
	if spTree = root.FindElement("p:cSld/p:spTree"); spTree == nil {
		return m, fmt.Errorf("Cannot find spTree")
	}
	imRoot := xml.Root()
	if imRoot == nil {
		return m, fmt.Errorf("Cannot find imRoot element")
	}
	spTree.Child = append(spTree.Child, imRoot)
	return m, nil
}

// addImageFile adds the png file to the presentation.
func (f *File) addImageFile(m image.Image, imageNum, slideNum int) error {
	imagePath := fmt.Sprintf("ppt/media/slide%dimage%d.png", slideNum, imageNum)
	var b bytes.Buffer
	if err := png.Encode(&b, m); err != nil {
		return fmt.Errorf("slide %d image %d: %s", slideNum, imageNum, err)
	}
	f.m[imagePath] = &b
	return nil
}

// build create the xml tree of the image reference.
func (im *Image) build(imNum int) (*etree.Document, image.Image, error) {
	m, err := im.Image.Raster()
	if err != nil {
		return nil, nil, err
	}
	cxDim := Dimension(m.Bounds().Dx()) * Inch / Dpi
	cyDim := Dimension(m.Bounds().Dy()) * Inch / Dpi
	x := strconv.FormatUint(uint64(im.X), 10)
	y := strconv.FormatUint(uint64(im.Y), 10)
	cx := strconv.FormatUint(uint64(cxDim), 10)
	cy := strconv.FormatUint(uint64(cyDim), 10)
	template := `<p:pic>
<p:nvPicPr>
<p:cNvPr id="` + strconv.Itoa(1026+imNum) + `" name="Picture ` + strconv.Itoa(imNum+1) + `"/>
<p:cNvPicPr/>
<p:nvPr/>
</p:nvPicPr>
<p:blipFill>
<a:blip r:embed="rId` + strconv.Itoa(imNum+2) + `">
<a:extLst>
<a:ext uri="{28A0092B-C50C-407E-A947-70E740481C1C}">
<a14:useLocalDpi xmlns:a14="http://schemas.microsoft.com/office/drawing/2010/main" val="0"/>
</a:ext>
</a:extLst>
</a:blip>
<a:srcRect/>
<a:stretch>
<a:fillRect/>
</a:stretch>
</p:blipFill>
<p:spPr bwMode="auto">
<a:xfrm>
<a:off x="` + x + `" y="` + y + `"/>
<a:ext cx="` + cx + `" cy="` + cy + `"/>
</a:xfrm>
<a:prstGeom prst="rect">
<a:avLst/>
</a:prstGeom>
<a:noFill/>
<a:extLst>
<a:ext uri="{909E8E84-426E-40DD-AFC4-6F175D3DCCD1}">
<a14:hiddenFill xmlns:a14="http://schemas.microsoft.com/office/drawing/2010/main">
<a:solidFill>
<a:srgbClr val="FFFFFF"/>
</a:solidFill>
</a14:hiddenFill>
</a:ext>
</a:extLst>
</p:spPr>
</p:pic>`
	doc := etree.NewDocument()
	err = doc.ReadFromString(template)
	return doc, m, err
}

/* This original image was 192x107
   <p:pic>
     <p:nvPicPr>
       <p:cNvPr id="1026" name="Picture 2"/>
       <p:cNvPicPr/>
       <p:nvPr/>
     </p:nvPicPr>
     <p:blipFill>
       <a:blip r:embed="rId2">
         <a:extLst>
           <a:ext uri="{28A0092B-C50C-407E-A947-70E740481C1C}">
             <a14:useLocalDpi xmlns:a14="http://schemas.microsoft.com/office/drawing/2010/main" val="0"/>
           </a:ext>
         </a:extLst>
       </a:blip>
       <a:srcRect/>
       <a:stretch>
         <a:fillRect/>
       </a:stretch>
     </p:blipFill>
     <p:spPr bwMode="auto">
       <a:xfrm>
         <a:off x="755576" y="692696"/>
         <a:ext cx="974725" cy="542925"/>
       </a:xfrm>
       <a:prstGeom prst="rect">
         <a:avLst/>
       </a:prstGeom>
       <a:noFill/>
       <a:extLst>
         <a:ext uri="{909E8E84-426E-40DD-AFC4-6F175D3DCCD1}">
           <a14:hiddenFill xmlns:a14="http://schemas.microsoft.com/office/drawing/2010/main">
             <a:solidFill>
               <a:srgbClr val="FFFFFF"/>
             </a:solidFill>
           </a14:hiddenFill>
         </a:ext>
       </a:extLst>
     </p:spPr>
   </p:pic>
*/
