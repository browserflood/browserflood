package main

import (
	"errors"
	"fmt"
	"strconv"

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
	for _, host := range p.Hosts {
		id, _ := strconv.ParseInt(host.Id, 0, 64)
		server := s.New(c.Map{"id": id})
		n, err := s.Destroy(server)
		if err != nil {
			return errors.New(fmt.Sprintf("Provider %s", err))
		}
		if n {
			fmt.Printf("Host %s destroyed.\n", server.Id())
		}
	}
	return nil
}
