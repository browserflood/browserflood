package main

import (
	"fmt"
)

func init() {
	register("run", runCmd, "Runs a load test.")
}

func runCmd() {
	fmt.Printf("Run\n")
}
