package main

import (
	"encoding/json"
	"os"
	"runtime"
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

type Host struct {
	Id   string // provider specific
	Host string // e.g. localhost
	User string // e.g. root
	Arch string // e.g. amd64, x86, etc.
	OS   string // e.g. linux, darwin, etc.
}

func NewProject() *Project {
	return &Project{
		Config: Config{DeployPath: "browserflood"},
		Hosts: []*Host{
			{
				Host: "localhost",
				User: os.Getenv("USER"),
				Arch: runtime.GOARCH,
				OS:   runtime.GOOS,
			},
		},
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
