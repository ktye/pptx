package pptx

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	// "github.com/beevik/etree"
)

// Dimension is a EMU (english metric unit) used to position elements on a slide.
type Dimension uint

const MilliMeter Dimension = 36000
const Inch Dimension = 914400
const Dpi = 96 // Dpi is used to calculate the size of images.

// File stores the zip reader for the original file and
// a map of changed files.
// Objects stored in the map are of type *etree.Document (for changed xml files)
// or bytes.Buffer for png images (both implement io.WriterTo).
type File struct {
	fileName  string // filename
	tmpName   string // temporary file name
	r         *zip.ReadCloser
	m         map[string]io.WriterTo // Map of changed or new files.
	numSlides int
}

type dummyReadCloser zip.ReadCloser

func (z dummyReadCloser) Close() error {
	return nil
}

// Open unzips a pptx file, and stores the content in file.
func Open(filename string) (File, error) {
	f := File{
		fileName: filename,
		tmpName:  filename + "_",
	}
	if r, err := zip.OpenReader(filename); err != nil {
		return f, err
	} else {
		f.r = r
		return f, nil
	}
}

// Abort closes the input file without writing anything.
func (f File) Abort() {
	f.closeInput()
}

// Close writes to the tempfile, closes the original file
// and moves the new (temp file) over the original file.
func (f File) Close() error {
	if out, err := os.Create(f.tmpName); err != nil {
		return fmt.Errorf("Could not write to temporary file: %s", err)
	} else {

		// Create the new temporary zip file.
		zw := zip.NewWriter(out)

		// Write all files, which have not been modified or added.
		for _, v := range f.r.File {
			if _, ok := f.m[v.Name]; !ok {
				if w, err := zw.Create(v.Name); err != nil {
					zw.Close()
					return err
				} else {
					if rc, err := v.Open(); err != nil {
						zw.Close()
					} else {
						if _, err := io.Copy(w, rc); err != nil {
							rc.Close()
							zw.Close()
							return err
						} else {
							rc.Close()
						}
					}
				}
			}
		}

		// Write all new files.
		for name, v := range f.m {
			if w, err := zw.Create(name); err != nil {
				zw.Close()
				return err
			} else {
				if _, err := v.WriteTo(w); err != nil {
					zw.Close()
					return err
				}
			}
		}

		// Close the zip writer.
		if err := zw.Close(); err != nil {
			out.Close()
			return err
		}
		// Close the underlying file.
		if err := out.Close(); err != nil {
			return err
		}
	}
	// Close the original input file.
	if err := f.closeInput(); err != nil {
		return err
	}
	// Move temp file over the original file.
	if err := os.Rename(f.tmpName, f.fileName); err != nil {
		return fmt.Errorf("Could not overwrite original file with updated content: %s", err)
	}
	return nil
}

// closeInput closes the original input pptx file.
func (f File) closeInput() error {
	if err := f.r.Close(); err != nil {
		return fmt.Errorf("Could not close the original pptx file: %s", err)
	}
	return nil
}
