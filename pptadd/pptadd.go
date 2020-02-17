package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ktye/plot"
	"github.com/ktye/pptx"
)

func init() {
	pptx.RegisterImageDecoder(plot.PptPlot{})
}

func main() {
	var r io.Reader
	args := os.Args[1:]
	if n := len(args); n < 1 || n > 2 || strings.HasSuffix(args[0], ".pptx") == false {
		fatal(fmt.Errorf("pptadd file.pptx [input.txt]"))
	} else if n == 1 {
		r = os.Stdin
	} else {
		f, err := os.Open(args[1])
		fatal(err)
		defer f.Close()
		r = f
	}

	slides, err := pptx.DecodeSlides(r)
	fatal(err)
	if len(slides) == 0 {
		fatal(fmt.Errorf("no slides added"))
	}

	f, err := pptx.Open(args[0])
	fatal(err)
	for _, s := range slides {
		if err := f.Add(s); err != nil {
			f.Abort()
			fatal(err)
		}
	}
	fatal(f.Close())
}
func fatal(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
