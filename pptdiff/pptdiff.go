// Program pptdiff prints differences in ppt xml files.
//
//
// Pptdiff unpacks 2 pptx files in memory, compares the directory structure,
// indents all xml files and prints xml file differences.
//
// It is used for development of pptx.
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/beevik/etree"
	"github.com/ktye/pptx/pptdiff/difflib"
)

// file holds the opended zip archive z and the map m of files to
// bytes.Buffer (for media files) or *etree.Document (for all other files).
type file struct {
	name string
	z    *zip.ReadCloser
	m    map[string]io.WriterTo
}

func open(filename string) file {
	var f file
	f.name = filename
	f.m = make(map[string]io.WriterTo)
	if r, err := zip.OpenReader(filename); err != nil {
		log.Fatalf("cannot unzip %s", filename)
	} else {
		f.z = r
	}
	for _, v := range f.z.File {
		isXml := true
		if strings.HasPrefix(v.Name, "ppt/media/") || strings.HasSuffix(v.Name, ".jpeg") {
			isXml = false
		}
		if rc, err := v.Open(); err != nil {
			fmt.Errorf("%s:%s: %s", filename, v.Name, err)
		} else {
			if isXml {
				d := etree.NewDocument()
				if _, err := d.ReadFrom(rc); err != nil {
					log.Fatalf("%s:%s: %s", filename, v.Name, err)
				}
				d.Indent(2)
				f.m[v.Name] = d
			} else {
				var b bytes.Buffer
				if _, err := b.ReadFrom(rc); err != nil {
					log.Fatalf("%s:%s: %s", filename, v.Name, err)
				} else {
					f.m[v.Name] = &b
				}
			}
			rc.Close()
		}
	}
	return f
}

// sortedKeys returns the keys of the map f.m alphabetically sorted.
func sortedKeys(f file) (keys []string) {
	keys = make([]string, len(f.m))
	i := 0
	for key := range f.m {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// difftree compares the directory structure.
// It prints only differences.
func difftree(f1, f2 file) {
	files1 := sortedKeys(f1)
	files2 := sortedKeys(f2)

	// Files that are in f1 but not in f2 are indicated by '-'
	header := false
	for _, name := range files1 {
		if _, ok := f2.m[name]; !ok {
			if header == false {
				fmt.Printf("# files present in %s but not in %s\n", f1.name, f2.name)
				header = true
			}
			fmt.Printf("-%s\n", name)
		}
	}
	if header {
		fmt.Println()
	}

	// Files that are in f2 but not in f1 are indicated by '+'
	header = false
	for _, name := range files2 {
		if _, ok := f1.m[name]; !ok {
			if header == false {
				fmt.Printf("# files present in %s but not in %s\n", f2.name, f1.name)
				header = true
			}
			fmt.Printf("+%s\n", name)
		}
	}
	if header {
		fmt.Println()
	}
}

// difffile compares xml files which are present in both pptx files.
// It prints only differences.
func difffiles(f1, f2 file) {
	files := sortedKeys(f1)
	for _, name := range files {
		a := f1.m[name]
		if b, ok := f2.m[name]; ok {
			printDiff([2]io.WriterTo{a, b}, f1.name+":"+name, f2.name+":"+name)
		}
	}
}

// printDiff prints the difference of 2 files.
func printDiff(r [2]io.WriterTo, aName, bName string) {
	var buf [2]bytes.Buffer
	for i := 0; i < 2; i++ {
		if _, err := r[i].WriteTo(&buf[i]); err != nil {
			log.Fatal(err)
		}
	}

	switch r[1].(type) {
	case *bytes.Buffer:
		compareBinary(buf, bName)
		return
	}

	diff := difflib.UnifiedDiff{
		A:        strings.Split(string(buf[0].Bytes()), "\n"),
		B:        strings.Split(string(buf[1].Bytes()), "\n"),
		FromFile: aName,
		ToFile:   bName,
		Eol:      "",
	}
	/*
		if err := difflib.WriteUnifiedDiff(os.Stdout, diff); err != nil {
			log.Fatal(err)
		}
	*/
	if text, err := difflib.GetUnifiedDiffString(diff); err != nil {
		log.Fatal(err)
	} else {
		if len(text) > 0 {
			fmt.Println(text)
		}
	}
}

func compareBinary(buf [2]bytes.Buffer, name string) {
	a := buf[0].Bytes()
	b := buf[1].Bytes()
	differs := false
	if len(a) != len(b) {
		differs = true
	} else {
		for i, v := range a {
			if v != b[i] {
				differs = true
				break
			}
		}
	}
	if differs {
		fmt.Printf("# %s: binary files differ\n", name)
	}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("pptdiff needs 2 input files")
	}
	f1 := open(os.Args[1])
	f2 := open(os.Args[2])
	difftree(f1, f2)
	difffiles(f1, f2)
}
