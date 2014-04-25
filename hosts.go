package main

import (
	"encoding/json"
	"os"
)

type Host struct {
	HostAddr string
	SSHUser  string
	SSHPort  string
}

func LoadHosts() ([]*Host, error) {
	file, err := os.Open("hosts.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	d := json.NewDecoder(file)
	results := []*Host{}
	err = d.Decode(&results)
	return results, err
}
