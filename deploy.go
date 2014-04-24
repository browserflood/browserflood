package main

import (
	"fmt"
)

func init() {
	register("deploy", deployCmd, "Deploys deps to all hosts.")
}

func deployCmd() {
	fmt.Printf("Run\n")
}
