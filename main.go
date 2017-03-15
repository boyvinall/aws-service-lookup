package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/urfave/cli"
)

const Name string = "aws-service-lookup"
const Version string = "0.1.0"

var (
	vpc       map[string]struct{}
	AccessKey string
	SecretKey string
)

func main() {
	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Author = ""
	app.Email = ""
	app.Usage = ""
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "aws-access-key",
			Value:  "",
			EnvVar: "AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "aws-secret-key",
			Value:  "",
			EnvVar: "AWS_SECRET_ACCESS_KEY",
		},
		cli.StringSliceFlag{
			Name: "vpc",
		},
		cli.BoolFlag{
			Name: "verbose, V",
		},
		cli.BoolFlag{
			Name:  "running",
			Usage: "only reports running instances",
		},
	}
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("verbose") {
			log.SetOutput(os.Stderr)
		} else {
			log.SetOutput(ioutil.Discard)
		}

		return nil
	}
	app.Commands = []cli.Command{
		CmdHosts,
		CmdServe,
		CmdListPlugins,
		CmdBashCompletion,
	}

	app.Run(os.Args)
}
