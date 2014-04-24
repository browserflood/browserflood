package main

import (
	"fmt"
	"os"
)

var (
	commands = make(map[string]func(), 0)
	usage    = make(map[string]string, 0)
)

func main() {
	if len(os.Args) < 2 {
		help()
		os.Exit(0)
	}
	cmd, ok := commands[os.Args[1]]
	if !ok {
		fatal("Unknown command: %s", cmd)
	}
	cmd()
}

func help() {
	fmt.Printf("browserflood <command>\n")
	fmt.Printf("\n")
	fmt.Printf("commands:\n")
	for name, use := range usage {
		fmt.Printf("  %s %s\n", name, use)
	}
}

func register(name string, function func(), use string) {
	commands[name] = function
	usage[name] = use
}

func fatal(reason string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+reason+"\n", args...)
	os.Exit(1)
}
