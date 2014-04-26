package main

import (
	"fmt"
	"os"
)

func init() {
	register("init", initCmd, "Initializes a browserflood project in the current directory.")
}

func initCmd() error {
	fmt.Printf("Creating project structure\n")
	dirs := []string{"bin", "log", "var", "tmp"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}
	// @TODO: create .gitignore file for /tmp /bin ?
	p := NewProject()
	return p.Save()
}
