package main

import (
	"fmt"
	"github.com/hawx/img/greyscale"
	"github.com/hawx/img/sharpen"
	"image"
	"io/ioutil"
	"os"

	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	termcolor "github.com/daviddengcn/go-colortext"
	"github.com/jessevdk/go-flags"
	"github.com/otiai10/gosseract"
	diff "github.com/sergi/go-diff/diffmatchpatch"
)

const (
	ReturnOk = iota
	ReturnHelp
)

const (
	DiffDelete = -1
	DiffInsert = 1
	DiffEqual  = 0
)

var opts struct {
	ComparePerfect   string `long:"compare-perfect" description:"Compare the OCRed text with the content of a file"`
	File             string `short:"f" long:"file" description:"File that should be OCRed" required:"true"`
	PreprocessingOut string `long:"preprocessing-out" description:"Save the image afert preprocessing"`
}

func main() {
	var err error

	p := flags.NewNamedParser("quentale", flags.HelpFlag)
	p.ShortDescription = "Extract informations of receipts"
	p.AddGroup("Quentale arguments", "", &opts)

	_, err = p.ParseArgs(os.Args)
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			panic(err)
		} else {
			p.WriteHelp(os.Stdout)

			os.Exit(ReturnHelp)
		}
	}

	// read file
	reader, err := os.Open(opts.File)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}

	// preprocessing
	img = greyscale.Maximal(img)
	img = sharpen.Sharpen(img, 1, 0.5)

	if opts.PreprocessingOut != "" {
		imageOut, err := os.Create(opts.PreprocessingOut)
		if err != nil {
			panic(err)
		}
		defer imageOut.Close()

		png.Encode(imageOut, img)
	}

	// ocr the heck out of the image
	ocr := gosseract.SummonServant()

	err = ocr.LangUse("deu")
	if err != nil {
		panic(err)
	}

	ocr.Eat(img)

	text, err := ocr.Out()
	if err != nil {
		panic(err)
	}

	// postprocessing

	// output
	if opts.ComparePerfect != "" {
		errorCount := 0

		by, err := ioutil.ReadFile(opts.ComparePerfect)
		if err != nil {
			panic(err)
		}

		perfect := string(by)

		ds := diff.New().DiffMain(text, perfect, true)

		for _, d := range ds {
			switch d.Type {
			case DiffDelete:
				termcolor.ChangeColor(termcolor.Red, false, termcolor.None, false)
				fmt.Print(d.Text)
				termcolor.ResetColor()

				errorCount += len(d.Text)
			case DiffEqual:
				fmt.Print(d.Text)
			}
		}

		fmt.Printf("\nGot %d errors\n", errorCount)
	} else {
		fmt.Println(text)
	}

	os.Exit(ReturnOk)
}
