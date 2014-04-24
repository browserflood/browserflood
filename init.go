package main

import (
	"os"
)

func initCmd() {
	if err := os.Mkdir("deps", 0777); err != nil {
		fatal("%s", err)
	}
	if _, err := os.OpenFile("config.toml", os.O_CREATE, 0x666); err != nil {
		fatal("%s", err)
	}
	if _, err := os.OpenFile("hosts.toml", os.O_CREATE, 0x666); err != nil {
		fatal("%s", err)
	}
}
