package main

import (
	"fmt"
)

func init() {
	register("deploy", deployCmd, "Deploys deps to all hosts.")
}

func deployCmd() {
	hosts, err := LoadHosts()
	if err != nil {
		fatal("%s", err)
	}
	for _, host := range hosts {
		deploy(host)
	}
}

func deploy(host *Host) {
	fmt.Printf("deploy to: %#v\n", host)
}
