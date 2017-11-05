package main

import (
	"fmt"
	"os"
	"strings"

	flag "github.com/ogier/pflag"
)

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
		fmt.Fprintln(os.Stderr, "SOURCE argument must be package path and type name separated by period\n")
		os.Exit(1)
	}

	fmt.Printf("should extract %s from %s\n", wantType, srcPkg)
}

func parseSourceArg(arg string) (string, string) {
	dotCount := strings.Count(arg, ".")
	if dotCount != 1 {
		return "", ""
	}

	dotPos := strings.Index(arg, ".")
	return arg[:dotPos], arg[dotPos+1:]
}
