package main

import (
	"fmt"
	"os"
)

var (
	// public commands, internal ones are special cased in switch below
	commands = make(map[string]func() error, 0)
	usage    = make(map[string]string, 0)
)

func main() {
	if len(os.Args) < 2 {
		help()
		os.Exit(0)
	}
	var cmd func() error
	switch cmdName := os.Args[1]; cmdName {
	case "slave":
		cmd = slaveCmd
	default:
		var ok bool
		cmd, ok = commands[cmdName]
		if !ok {
			fatal("Unknown command: %s", cmdName)
		}
	}
	if err := cmd(); err != nil {
		fatal("%s", err)
	}
}

func help() {
	fmt.Printf("browserflood <command>\n")
	fmt.Printf("\n")
	fmt.Printf("commands:\n")
	for name, use := range usage {
		fmt.Printf("  %s %s\n", name, use)
	}
}

func register(name string, function func() error, use string) {
	commands[name] = function
	usage[name] = use
}

func fatal(reason string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+reason+"\n", args...)
	os.Exit(1)
}
