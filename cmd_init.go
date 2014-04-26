package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

const (
	phantomVersion = "1.9.7"
	phantomFile    = "phantomjs.tar.bz2"
	phantomDir     = "phantomjs-" + phantomVersion + "-linux-x86_64"
	phantomURL     = "https://bitbucket.org/ariya/phantomjs/downloads/" + phantomDir + ".tar.bz2"
)

func init() {
	register("init", initCmd, "Initializes a browserflood project in the current directory.")
}

func initCmd() error {
	fmt.Printf("Creating project structure\n")
	if err := os.MkdirAll("dist", 0777); err != nil {
		return err
	}
	// @TODO Using browserflood should not require having go installed. But for
	// now this is ok / will allow us to iterate quickly.
	fmt.Printf("Building browserflood for linux/amd64\n")
	build := exec.Command("go", "build", "-o", "dist/browserflood", "github.com/browserflood/browserflood")
	build.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return err
	}
	p := NewProject()
	if err := p.Save(); err != nil {
		return err
	}
	fmt.Printf("Downloading phantomjs %s\n", phantomVersion)
	if err := download(phantomURL, "dist/"+phantomFile); err != nil {
		return err
	}
	fmt.Printf("Extracting phantomjs\n")
	tar := exec.Command("tar", "-xzf", phantomFile)
	tar.Dir = "dist"
	tar.Stderr = os.Stderr
	if err := tar.Run(); err != nil {
		return err
	}
	extractDir := "dist/" + phantomDir
	if err := os.Rename(extractDir+"/bin/phantomjs", "dist/phantomjs"); err != nil {
		return err
	}
	if err := os.RemoveAll(extractDir); err != nil {
		return err
	}
	if err := os.RemoveAll("dist/" + phantomFile); err != nil {
		return err
	}
	return nil
}

func download(url string, dst string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, res.Body); err != nil {
		return err
	}
	return nil
}
