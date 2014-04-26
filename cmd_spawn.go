package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	c "github.com/gcloud/compute"
	_ "github.com/gcloud/compute/providers/digitalocean"
	"github.com/gcloud/identity"
)

func init() {
	register("spawn", spawnCmd, "Launch [n] servers.")
}

func spawnCmd() error {
	if len(os.Args[2:]) < 1 {
		return errors.New("Specify the number of servers to spawn.")
	}
	p, err := LoadProject()
	if err != nil {
		return err
	}
	n, _ := strconv.ParseInt(os.Args[2:][0], 0, 64)
	account := &identity.Account{Id: p.Provider.Id, Key: p.Provider.Secret}
	s := c.GetServers("digitalocean", account)
	p.Hosts = make([]*Host, 0)
	allerrors := make([]error, 0)
	results := make(chan *Host, n)
	errs := make(chan error, n)
	fmt.Printf("Spawning %d servers.\n", n)
	for i := 0; i < int(n); i++ {
		go func() {
			r, e := spawn(s, i, p)
			results <- r
			errs <- e
		}()
	}
	for i := 0; i < int(n); i++ {
		if err := <-errs; err != nil {
			allerrors = append(allerrors, err)
		}
		if host := <-results; host != nil {
			p.Hosts = append(p.Hosts, host)
		}
	}
	if len(p.Hosts) > 0 {
		p.Save()
		fmt.Println("hosts.json saved.")
	}
	if len(allerrors) > 0 {
		return errors.New(fmt.Sprintf("%#v", allerrors))
	}
	return nil
}

func spawn(s c.Servers, i int, p *Project) (*Host, error) {
	result, err := s.Create(s.New(c.Map{
		"name":        fmt.Sprintf("browserflood-%d", i),
		"image_id":    p.Provider.ImageId,
		"size_id":     p.Provider.SizeId,
		"region_id":   p.Provider.RegionId,
		"ssh_key_ids": p.Provider.SshKeyIds,
	}))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Provider %s", err))
	}
	server, err := s.Show(result)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Get %s", err))
	}
	ips := server.Ips("public")
	h := &Host{
		Id: server.Id(), Host: ips[0], User: "root", Arch: "amd64", OS: "linux",
	}
	return h, nil
}
