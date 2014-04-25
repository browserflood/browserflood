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
	if err := os.Mkdir("bin", 0777); err != nil {
		return err
	}
	if err := os.Mkdir("var", 0777); err != nil {
		return err
	}
	if err := os.Mkdir("bin/32bit", 0777); err != nil {
		return err
	}
	if err := os.Mkdir("bin/64bit", 0777); err != nil {
		return err
	}
	p := NewProject()
	if err := p.Save(); err != nil {
		return err
	}
	fmt.Printf("Downloading phantomjs %s (32bit)\n", phantomVersion)
	if err := download(phantom32URL, "bin/32bit/phantomjs"); err != nil {
		return err
	}
	fmt.Printf("Downloading phantomjs %s (64bit)\n", phantomVersion)
	if err := download(phantom64URL, "bin/64bit/phantomjs"); err != nil {
		return err
	}
	fmt.Printf("Done\n")
	return nil
}
