package main

import (
	"fmt"
)

func init() {
	register("spawn", spawnCmd, "Launch [n] servers on [provider].")
}

func spawnCmd() {
	fmt.Printf("Spawning!\n")
}
