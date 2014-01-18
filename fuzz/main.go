package main

import (
	"fmt"
	"github.com/hawx/img/blend"
	"github.com/hawx/img/blur"
	"github.com/hawx/img/contrast"
	"github.com/hawx/img/gamma"
	"github.com/hawx/img/greyscale"
	"github.com/hawx/img/pixelate"
	"github.com/hawx/img/sharpen"
	"github.com/hawx/img/utils"
	"image"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"time"

	_ "image/png"

	termcolor "github.com/daviddengcn/go-colortext"
	"github.com/jessevdk/go-flags"
	diff "github.com/sergi/go-diff/diffmatchpatch"
	"github.com/zimmski/gosseract"
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
	File string `long:"file" required:"true"`
	Text string `long:"text" required:"true"`
}

func main() {
	var err error

	p := flags.NewNamedParser("fuzz", flags.HelpFlag)
	p.ShortDescription = "A fuzzer for the img filters"
	p.AddGroup("Fuzzer arguments", "", &opts)

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

	by, err := ioutil.ReadFile(opts.Text)
	if err != nil {
		panic(err)
	}

	perfect := string(by)

	best := math.MaxInt64
	var filters = []string{
		"blend",
		"blur",
		"channel",
		"contrast",
		"gamma",
		"greyscale",
		"levels",
		"pixelate",
		"sharpen",
	}
	om := make(map[string]struct{})

	rand.Seed(time.Now().UnixNano())

	for {
		t := img

		// preprocessing
		fm := make(map[int]struct{})
		filterCount := rand.Intn(len(filters))
		selectedFilters := ""

		for i := 0; i < filterCount; i++ {
			filterID := rand.Intn(len(filters))

			if _, ok := fm[filterID]; !ok {
				fm[filterID] = struct{}{}

				filterName := filters[filterID]
				if selectedFilters != "" {
					selectedFilters += "," + filterName
				} else {
					selectedFilters = filterName
				}

				switch filterName {
				case "blend":
					v := rand.Float64()
					selectedFilters += fmt.Sprintf("!%f", v)
					t = blend.Fade(t, v)
				case "blur":
					v := rand.Intn(4) + 1
					selectedFilters += fmt.Sprintf("!%d", v)
					t = blur.Box(t, v, blur.IGNORE)
				case "channel":
				case "contrast":
					v := rand.Float64()
					selectedFilters += fmt.Sprintf("!%f", v)
					t = contrast.Linear(t, v)
				case "gamma":
					t = gamma.Auto(t)
				case "greyscale":
					t = greyscale.Maximal(t)
				case "levels":
				case "pixelate":
					t = pixelate.Pixelate(t, utils.Dimension{2, 1}, pixelate.FITTED)
				case "sharpen":
					v1 := rand.Intn(4) + 1
					selectedFilters += fmt.Sprintf("!%d", v1)
					v2 := rand.Float64()
					selectedFilters += fmt.Sprintf("!%f", v2)
					t = sharpen.Sharpen(t, v1, v2)
				}
			}
		}

		if _, ok := om[selectedFilters]; ok {
			continue
		}
		om[selectedFilters] = struct{}{}

		// ocr the heck out of the image
		ocr := gosseract.SummonServant()

		err = ocr.LangUse("deu")
		if err != nil {
			panic(err)
		}

		ocr.Eat(t)

		text, err := ocr.Out()
		if err != nil {
			panic(err)
		}

		// count errors
		errorCount := 0

		if text == "" {
			errorCount = len(perfect)
		} else {
			ds := diff.New().DiffMain(text, perfect, true)

			for _, d := range ds {
				if d.Type == DiffDelete {
					errorCount += len(d.Text)
				}
			}
		}

		if errorCount < best {
			best = errorCount

			termcolor.ChangeColor(termcolor.Green, false, termcolor.None, false)
			fmt.Printf("%d;%s\n", errorCount, selectedFilters)
			termcolor.ResetColor()
		} else {
			fmt.Printf("%d;%s\n", errorCount, selectedFilters)
		}
	}
}
