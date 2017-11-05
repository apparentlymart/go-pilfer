package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/apparentlymart/go-pilfer/pilfer"
	flag "github.com/ogier/pflag"
)

var outPath = flag.StringP("output", "o", "", "output filename")
var outPkg = flag.String("package", "", "package name for generated file")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] SOURCE\n", os.Args[0])
		flag.PrintDefaults()
		os.Stderr.WriteString("\n")
	}
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	srcPkg, wantType := parseSourceArg(args[0])
	if srcPkg == "" || wantType == "" {
		fmt.Fprintln(os.Stderr, "SOURCE argument must be package path and type name separated by colon\n")
		os.Exit(1)
	}

	if *outPath == "" {
		outV := fmt.Sprintf("%s.go", strings.ToLower(wantType))
		outPath = &outV
	}

	outAbs, err := filepath.Abs(*outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error with output file: %s\n", err)
		os.Exit(1)
	}

	// If we don't have a user-supplied package name then we'll try to guess
	// one based on existing files in the output directory.
	if *outPkg == "" {
		outDir := filepath.Dir(outAbs)
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, outDir, nil, parser.PackageClauseOnly)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error inferring package name: %s\n", err)
			os.Exit(1)
		}
		if len(pkgs) != 1 {
			fmt.Fprintf(os.Stderr, "failed to infer a single package name for output directory %s\n", outDir)
			os.Exit(1)
		}
		for _, pkg := range pkgs {
			name := pkg.Name
			outPkg = &name
			break
		}
	}

	outF, err := os.Create(outAbs)
	defer outF.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output file: %s\n", err)
		os.Exit(1)
	}

	err = pilfer.Pilfer(srcPkg, wantType, outF, *outPkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func parseSourceArg(arg string) (string, string) {
	sepCount := strings.Count(arg, ":")
	if sepCount != 1 {
		return "", ""
	}

	sepPos := strings.Index(arg, ":")
	return arg[:sepPos], arg[sepPos+1:]
}
