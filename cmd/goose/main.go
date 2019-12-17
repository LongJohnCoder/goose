package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"

	"github.com/tchajed/goose"
)

//noinspection GoUnhandledErrorResult
func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: goose [options] <path to go package>")

		flag.PrintDefaults()
	}
	var config goose.Config
	flag.BoolVar(&config.AddSourceFileComments, "source-comments", false,
		"add comments indicating Go source code location for each top-level declaration")
	flag.BoolVar(&config.TypeCheck, "typecheck", false,
		"add type-checking theorems")

	var outFile string
	flag.StringVar(&outFile, "out", "-",
		"file to output to (use '-' for stdout)")

	var ignoreErrors bool
	flag.BoolVar(&ignoreErrors, "ignore-errors", false,
		"output partial translation even if there are errors")

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	srcDir := flag.Arg(0)

	red := color.New(color.FgRed).SprintFunc()
	f, err := config.TranslatePackage(srcDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, red(err.Error()))
		if !ignoreErrors {
			os.Exit(1)
		}
	}
	if outFile == "-" {
		f.Write(os.Stdout)
	} else {
		out, err := os.Create(outFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			fmt.Fprintln(os.Stderr, red("could not write output"))
			os.Exit(1)
		}
		defer out.Close()
		f.Write(out)
	}
	if err != nil {
		os.Exit(1)
	}
}
