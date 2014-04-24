package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(0)
	}
	switch cmd := os.Args[1]; cmd {
	case "init":
		initCmd()
	default:
		fatal("Unknown command: %s", cmd)
	}
}

func usage() {
	fmt.Printf("browserflood <command>\n")
	fmt.Printf("\n")
	fmt.Printf("commands:\n")
	fmt.Printf("  init  Initializes a browserflood project in the current directory.\n")
}

func fatal(reason string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+reason+"\n", args...)
	os.Exit(1)
}
