package main

import (
	"fmt"
)

func init() {
	register("run", runCmd, "Runs a load test.")
}

func runCmd() error {
	fmt.Printf("todo\n")
	return nil
}
