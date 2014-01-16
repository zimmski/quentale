package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/otiai10/gosseract"
)

const (
	ReturnOk = iota
	ReturnHelp
)

var opts struct {
	File string `short:"f" long:"file" description:"File that should be OCRed" required:"true"`
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

	ocr := gosseract.SummonServant()

	err = ocr.LangUse("deu")
	if err != nil {
		panic(err)
	}

	ocr.Target(opts.File)

	text, err := ocr.Out()
	if err != nil {
		panic(err)
	}

	fmt.Println(text)

	os.Exit(ReturnOk)
}
