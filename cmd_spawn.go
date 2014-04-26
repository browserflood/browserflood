package main

import (
	"errors"
	"fmt"

	c "github.com/gcloud/compute"
	_ "github.com/gcloud/compute/providers/digitalocean"
	"github.com/gcloud/identity"
)

func init() {
	register("spawn", spawnCmd, "Launch [n] servers.")
}

func spawnCmd() error {
	p, err := LoadProject()
	if err != nil {
		return err
	}
	account := &identity.Account{Id: p.Provider.Id, Key: p.Provider.Secret}
	s := c.GetServers("digitalocean", account)
	n := 1
	p.Hosts = make([]*Host, 0)
	for i := 0; i < n; i++ {
		result, err := s.Create(s.New(c.Map{
			"name":        fmt.Sprintf("browserflood-%d", i),
			"image_id":    3101045,
			"size_id":     66,
			"region_id":   1,
			"ssh_key_ids": 18420,
		}))
		if err != nil {
			return errors.New(fmt.Sprintf("Provider %s", err))
		}
		server, err := s.Show(result)
		if err != nil {
			return errors.New(fmt.Sprintf("Get %s", err))
		}
		ips := server.Ips("public")
		p.Hosts = append(p.Hosts, &Host{
			Id: server.Id(), Host: ips[0], User: "root", Arch: "amd64", OS: "linux",
		})
		fmt.Printf("%s\n", server)
	}
	p.Save()
	return nil
}
