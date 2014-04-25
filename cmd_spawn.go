package main

import (
	"fmt"

	c "github.com/gcloud/compute"
	p "github.com/gcloud/compute/providers"
	_ "github.com/gcloud/compute/providers/digitalocean"
	"github.com/gcloud/identity"
)

func init() {
	register("spawn", spawnCmd, "Launch [n] servers.")
}

func spawnCmd() error {
	s := c.Servers{account, "digitalocean"}
	n := 1
	for i := 0; i < n; i++ {
		result, err := s.Create(p.Map{
			"name":      fmt.Sprintf("browserflood-%d", i),
			"image_id":  3101045,
			"size_id":   66,
			"region_id": 1,
			"ssh_keys":  18420,
		})
		if err != nil {
			return err
		}
		fmt.Printf("%s", result)
	}
	return nil
}
