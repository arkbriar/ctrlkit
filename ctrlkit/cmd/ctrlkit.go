package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/imports"

	"github.com/arkbriar/ctrlkit/pkg/gen"
)

var (
	boilerplateFile string
	targetFile      string
	outputPath      string
	showHelp        bool
	verbose         bool
	ctrlKitPackage  string
)

func init() {
	flag.StringVar(&boilerplateFile, "b", "", "boilerplate file (optional)")
	flag.StringVar(&outputPath, "o", "", "output path (optional)")
	flag.BoolVar(&showHelp, "h", false, "show help")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.StringVar(&ctrlKitPackage, "p", "", "replace ctrlkit package")
}

func parseFlags() {
	flag.Parse()
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	targetFile = flag.Arg(0)

	if len(ctrlKitPackage) > 0 {
		gen.CtrlKitPackage = ctrlKitPackage
	}
}

func parseDoc() *gen.ControllerManagerDocument {
	f, err := os.Open(targetFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	doc, err := gen.ParseDoc(bufio.NewReader(f))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if verbose {
		fmt.Println("================= PARSED DOCUMENT =================")
		b, _ := json.MarshalIndent(doc, "", "  ")
		fmt.Println(string(b))
		fmt.Println("===================================================")
	}

	return doc
}

func newGoFileName(f string) string {
	if strings.HasSuffix(f, ".cm") {
		return f[:len(f)-3] + ".go"
	}
	return f + ".go"
}

const (
	goBuildIgnoreComments = `//go:build !ignore_autogenerated
// +build !ignore_autogenerated

`

	doNotEditComment = `// Code generated by ctrlkit. DO NOT EDIT.

`
)

func main() {
	parseFlags()

	fileName := newGoFileName(filepath.Base(targetFile))
	packageName := "manager"
	if len(outputPath) > 0 {
		packageName = filepath.Base(outputPath)
	}

	s, err := gen.GenerateStubCodes(parseDoc(), packageName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if verbose {
		fmt.Println("================= UNFORMATTED CODE =================")
		fmt.Println(s)
		fmt.Println("====================================================")
	}

	formatted, err := imports.Process("", []byte(s), nil)
	if err != nil {
		fmt.Println(fmt.Errorf("format error: %w", err))
		os.Exit(1)
	}
	s = string(formatted)

	var w io.Writer = os.Stdout
	if len(outputPath) > 0 {
		f, err := os.OpenFile(filepath.Join(outputPath, fileName), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()

		w = f
	}
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	// Write ignore header.
	if _, err := io.WriteString(w, goBuildIgnoreComments+doNotEditComment); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Write bolierplate first (if provided).
	if len(boilerplateFile) > 0 {
		func() {
			f, err := os.Open(boilerplateFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if _, err = io.Copy(w, f); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}()
		if _, err := io.WriteString(w, "\n"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if _, err := io.WriteString(w, s); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
