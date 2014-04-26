package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	browserFloodPKG  = "github.com/browserflood/browserflood"
	phantomVersion   = "1.9.7"
	phantomDarwinURL = "https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-%s-macosx.zip"
)

func init() {
	register("run", runCmd, "Runs a load test.")
}

func runCmd() error {
	p, err := LoadProject()
	if err != nil {
		return err
	}

	targets := map[target]bool{}
	for _, host := range p.Hosts {
		targets[target{Arch: host.Arch, OS: host.OS}] = true

	}
	for target, _ := range targets {
		if err := crossCompileBrowserflood(target); err != nil {
			return err
		}
		if err := downloadPhantomJS(target, phantomVersion); err != nil {
			return err
		}
	}
	fmt.Printf("Syncing files to %d hosts\n", len(p.Hosts))
	results := make(chan error, len(p.Hosts))
	for _, host := range p.Hosts {
		go func() {
			results <- deploy(p, host)
		}()
	}
	for _ = range p.Hosts {
		if err := <-results; err != nil {
			return err
		}
	}
	return nil
}

func deploy(p *Project, host *Host) error {
	dst := fmt.Sprintf("%s@%s:%s", host.User, host.Host, p.Config.DeployPath)
	bin := fmt.Sprintf("bin/%s/%s/", host.OS, host.Arch)
	rsync := exec.Command("rsync", "-e", "ssh", "-rz", bin, "var/", dst)
	rsync.Stderr = os.Stderr
	return rsync.Run()
}

type target struct {
	Arch string
	OS   string
}

func crossCompileBrowserflood(t target) error {
	// @TODO Using browserflood should not require having go installed. But for
	// now this is ok / will allow us to iterate quickly.
	fmt.Printf("Building browserflood for %s/%s\n", t.OS, t.Arch)
	path := fmt.Sprintf("bin/%s/%s/browserflood", t.OS, t.Arch)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	build := exec.Command("go", "build", "-o", path, browserFloodPKG)
	build.Env = append(os.Environ(), "GOOS="+t.OS, "GOARCH="+t.Arch)
	build.Stderr = os.Stderr
	return build.Run()
}

func downloadPhantomJS(t target, version string) error {
	path := fmt.Sprintf("bin/%s/%s/phantomjs", t.OS, t.Arch)
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	fmt.Printf("Downloading phantomjs %s for %s/%s\n", version, t.OS, t.Arch)
	notSupported := fmt.Errorf(
		"Downloading phantomjs is not supported for %s/%s. Please download and place it in %s manually.",
		t.OS,
		t.Arch,
		path,
	)
	switch t.OS {
	case "darwin":
		if t.Arch != "amd64" {
			return notSupported
		}
		url := fmt.Sprintf(phantomDarwinURL, version)
		// We could also do this without the temporary file and directly pipe the
		// download into the zip reader, but for now this makes debugging the code
		// easier.
		dst := filepath.Join("tmp", filepath.Base(url))
		if err := download(url, dst); err != nil {
			return err
		}
		reader, err := zip.OpenReader(dst)
		if err != nil {
			return err
		}
		defer reader.Close()
		extracted := false
		for _, file := range reader.File {
			if strings.HasSuffix(file.Name, "bin/phantomjs") {
				data, err := file.Open()
				if err != nil {
					return err
				}
				defer data.Close()
				dstFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
				if err != nil {
					return err
				}
				defer dstFile.Close()
				if _, err := io.Copy(dstFile, data); err != nil {
					return err
				}
				extracted = true
				break
			}
		}
		if !extracted {
			return fmt.Errorf("Could not find phantomjs in zip file.")
		}
		return nil
	default:
		return notSupported
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
