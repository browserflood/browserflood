package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func init() {
	register("init", initCmd, "Initializes a browserflood project in the current directory.")
}

const (
	phantomVersion = "1.9.7"
	phantom32URL   = "https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-" + phantomVersion + "-linux-x86_64.tar.bz2"
	phantom64URL   = "https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-" + phantomVersion + "-linux-i686.tar.bz2"
)

func initCmd() {
	fmt.Printf("Creating project structure\n")
	if err := os.Mkdir("deps", 0777); err != nil {
		fatal("%s", err)
	}
	if err := os.Mkdir("deps/32bit", 0777); err != nil {
		fatal("%s", err)
	}
	if err := os.Mkdir("deps/64bit", 0777); err != nil {
		fatal("%s", err)
	}
	if _, err := os.OpenFile("config.json", os.O_CREATE, 0x666); err != nil {
		fatal("%s", err)
	}
	if _, err := os.OpenFile("hosts.json", os.O_CREATE, 0x666); err != nil {
		fatal("%s", err)
	}
	if _, err := os.OpenFile("deploy.bash", os.O_CREATE, 0x666); err != nil {
		fatal("%s", err)
	}
	fmt.Printf("Downloading phantomjs %s (32bit)\n", phantomVersion)
	if err := download(phantom32URL, "deps/32bit/phantomjs"); err != nil {

		fatal("%s", err)
	}
	fmt.Printf("Downloading phantomjs %s (64bit)\n", phantomVersion)
	if err := download(phantom64URL, "deps/64bit/phantomjs"); err != nil {
		fatal("%s", err)
	}
	fmt.Printf("Done\n")
}

func download(url string, dst string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0x666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, res.Body); err != nil {
		return err
	}
	return nil
}
