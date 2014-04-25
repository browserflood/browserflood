package main

import (
	"fmt"
	"os"
	"os/exec"
)

func init() {
	register("deploy", deployCmd, "Deploys deps to all hosts.")
}

func deployCmd() error {
	project, err := LoadProject()
	if err != nil {
		return err
	}
	results := make(chan error, len(project.Hosts))
	for _, host := range project.Hosts {
		go func() {
			results <- deploy(project.Config, host)
		}()
	}
	for _ = range project.Hosts {
		if err := <-results; err != nil {
			return err
		}
	}
	return nil
}

func deploy(config Config, host *Host) error {
	dst := fmt.Sprintf("%s@%s:%s", host.User, host.Addr, config.DeployPath)
	rsync := exec.Command("rsync", "-e", "ssh", "-rz", "dist/", dst)
	rsync.Stderr = os.Stderr
	return rsync.Run()
}
