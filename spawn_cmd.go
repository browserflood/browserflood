package main

import (
	"fmt"
)

func init() {
	register("spawn", spawnCmd, "Launch [n] servers on [provider].")
}

func spawnCmd() error {
	fmt.Printf("Spawning!\n")
	return nil
}
