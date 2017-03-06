package main

import (
	"fmt"

	"github.com/boyvinall/aws-service-lookup/ec2tags"
	"github.com/urfave/cli"
)

var CmdHosts = cli.Command{
	Name:   "hosts",
	Usage:  "",
	Action: hosts,
	Flags:  []cli.Flag{},
}

func hosts(c *cli.Context) error {
	accesskey := c.GlobalString("aws-access-key")
	secretkey := c.GlobalString("aws-secret-key")

	vpc := c.GlobalStringSlice("vpc")
	v := make(map[string]struct{}, 0)
	for _, k := range vpc {
		if k == "local" {
			vpcs, err := ec2tags.GetLocalVPCs()
			if err != nil {
				continue
			}
			for _, j := range vpcs {
				v[j] = struct{}{}
			}
		} else {
			v[k] = struct{}{}
		}
	}

	hosts, err := ec2tags.Query(accesskey, secretkey, v, c.GlobalBool("running"))
	if err != nil {
		return err
	}

	r := hosts.Records()
	for name, ips := range r {
		for _, ip := range ips {
			fmt.Printf("%-20s%s\n", ip.String(), name)
		}
	}
	return nil
}
