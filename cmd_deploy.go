package main

import (
	"fmt"
)

func init() {
	register("deploy", deployCmd, "Deploys deps to all hosts.")
}

func deployCmd() error {
	project, err := LoadProject()
	if err != nil {
		return err
	}
	for _, host := range project.Hosts {
		deploy(host)
	}
	return nil
}

func deploy(host *Host) {
	fmt.Printf("deploy to: %#v\n", host)
}
