package main

import (
	"errors"
	"fmt"

	c "github.com/gcloud/compute"
	_ "github.com/gcloud/compute/providers/digitalocean"
	"github.com/gcloud/identity"
)

func init() {
	register("destroy", destroyCmd, "Destroy servers.")
}

func destroyCmd() error {
	p, err := LoadProject()
	if err != nil {
		return err
	}
	account := &identity.Account{Id: p.Provider.Id, Key: p.Provider.Secret}
	s := c.GetServers("digitalocean", account)
	n, err := s.Destroy(s.New(c.Map{"id": "1533067"}))
	if err != nil {
		return errors.New(fmt.Sprintf("Provider %s", err))
	}
	fmt.Printf("%#v", n)
	return nil
}
