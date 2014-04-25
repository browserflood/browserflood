package main

import (
	"fmt"

	c "github.com/gcloud/compute"
	_ "github.com/gcloud/compute/providers/digitalocean"
)

func init() {
	register("destroy", destroyCmd, "Destroy servers.")
}

func destroyCmd() error {
	s := c.Servers{account, "digitalocean"}
	n, err := s.Destroy("1533067")
	if err != nil {
		return err
	}
	fmt.Printf("%#v", n)
}
