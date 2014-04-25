package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

const (
	phantomVersion = "1.9.7"
	phantomURL     = "https://bitbucket.org/ariya/phantomjs/downloads/phantomjs-" + phantomVersion + "-linux-i686.tar.bz2"
)

type Project struct {
	Config   Config
	Provider Provider
	Hosts    []*Host
}

type Provider struct {
	Id     string
	Secret string
}

func (p *Project) Save() error {
	if err := writeJSON("provider.json", p.Provider); err != nil {
		return err
	}
	if err := writeJSON("config.json", p.Config); err != nil {
		return err
	}
	if err := writeJSON("hosts.json", p.Hosts); err != nil {
		return err
	}
	return nil
}

type Config struct {
	DeployPath string
}

func DefaultConfig() Config {
	return Config{
		DeployPath: "browserflood",
	}
}

type Host struct {
	Id   string
	Addr string
	User string
}

func NewProject() *Project {
	return &Project{
		Config: Config{DeployPath: "browserflood"},
		Hosts:  []*Host{},
	}
}

func writeJSON(path string, data interface{}) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	e := json.NewEncoder(file)
	return e.Encode(data)
}

func readJSON(path string, data interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	d := json.NewDecoder(file)
	return d.Decode(data)
}

func LoadProject() (*Project, error) {
	p := &Project{}
	if err := readJSON("hosts.json", &p.Hosts); err != nil {
		return nil, err
	}
	if err := readJSON("config.json", &p.Config); err != nil {
		return nil, err
	}
	return p, nil
}

func download(url string, dst string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, res.Body); err != nil {
		return err
	}
	return nil
}
