package main

import (
	"fmt"

	"github.com/mholt/caddy"
	"github.com/urfave/cli"
)

var CmdListPlugins = cli.Command{
	Name:     "list-plugins",
	Usage:    "",
	Action:   listPlugins,
	Flags:    []cli.Flag{},
	Category: "Misc",
}

func listPlugins(c *cli.Context) error {
	fmt.Println(caddy.DescribePlugins())
	return nil
}
