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
	if err := os.Mkdir("dist", 0777); err != nil {
		return err
	}
	p := NewProject()
	if err := p.Save(); err != nil {
		return err
	}
	fmt.Printf("Downloading phantomjs %s\n", phantomVersion)
	if err := download(phantomURL, "dist/phantomjs"); err != nil {
		return err
	}
	return nil
}
