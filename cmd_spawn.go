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
		fmt.Println("Spawning server.")
		result, err := s.Create(s.New(c.Map{
			"name":        fmt.Sprintf("browserflood-%d", i),
			"image_id":    p.Provider.Image_id,
			"size_id":     p.Provider.Size_id,
			"region_id":   p.Provider.Region_id,
			"ssh_key_ids": p.Provider.Ssh_key_ids,
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
	fmt.Println("hosts.json saved.")
	return nil
}
