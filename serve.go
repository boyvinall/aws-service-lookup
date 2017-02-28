// Package coremain contains the functions for starting CoreDNS.
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/mholt/caddy"
	"github.com/urfave/cli"
	"github.com/boyvinall/aws-service-lookup/ec2tags"

	// plug in the standard directives (sorted)
	// _ "github.com/coredns/coredns/middleware/auto"
	_ "github.com/coredns/coredns/middleware/bind"
	_ "github.com/coredns/coredns/middleware/cache"
	// _ "github.com/coredns/coredns/middleware/chaos"
	// _ "github.com/coredns/coredns/middleware/dnssec"
	// _ "github.com/coredns/coredns/middleware/erratic"
	_ "github.com/coredns/coredns/middleware/errors"
	// _ "github.com/coredns/coredns/middleware/etcd"
	// _ "github.com/coredns/coredns/middleware/file"
	// _ "github.com/coredns/coredns/middleware/health"
	// _ "github.com/coredns/coredns/middleware/kubernetes"
	// _ "github.com/coredns/coredns/middleware/loadbalance"
	_ "github.com/coredns/coredns/middleware/log"
	// _ "github.com/coredns/coredns/middleware/metrics"
	// _ "github.com/coredns/coredns/middleware/pprof"
	_ "github.com/coredns/coredns/middleware/proxy"
	// _ "github.com/coredns/coredns/middleware/reverse"
	// _ "github.com/coredns/coredns/middleware/rewrite"
	// _ "github.com/coredns/coredns/middleware/root"
	// _ "github.com/coredns/coredns/middleware/secondary"
	// _ "github.com/coredns/coredns/middleware/trace"
	// _ "github.com/coredns/coredns/middleware/whoami"
)

var CmdServe = cli.Command{
	Name:   "serve",
	Usage:  "Run a CoreDNS server",
	Action: serve,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "conf",
			Usage:       "Corefile to load",
			Value:       "",
			Destination: &conf,
		},
		cli.StringFlag{
			Name:        "cpu",
			Value:       `100%`,
			Usage:       "CPU cap",
			Destination: &cpu,
		},
		cli.BoolFlag{
			Name:        "plugins",
			Usage:       "List installed plugins",
			Destination: &plugins,
		},
		cli.StringFlag{
			Name:        "log",
			Value:       "",
			Usage:       "Process log file",
			Destination: &logfile,
		},
		cli.StringFlag{
			Name:        "pidfile",
			Value:       "",
			Usage:       "Path to write pid file",
			Destination: &caddy.PidFile,
		},
		cli.BoolFlag{
			Name:        "quiet",
			Usage:       "Quiet mode (no initialization output)",
			Destination: &dnsserver.Quiet,
		},
	},
}

// mustLogFatal wraps log.Fatal() in a way that ensures the
// output is always printed to stderr so the user can see it
// if the user is still there, even if the process log was not
// enabled. If this process is an upgrade, however, and the user
// might not be there anymore, this just logs to the process
// log and exits.
func mustLogFatal(args ...interface{}) {
	if !caddy.IsUpgrade() {
		log.SetOutput(os.Stderr)
	}
	log.Fatal(args...)
}

// confLoader loads the Caddyfile using the -conf flag.
func confLoader(serverType string) (caddy.Input, error) {
	if conf == "" {
		return nil, nil
	}

	if conf == "stdin" {
		return caddy.CaddyfileFromPipe(os.Stdin, "dns")
	}

	contents, err := ioutil.ReadFile(conf)
	if err != nil {
		return nil, err
	}
	return caddy.CaddyfileInput{
		Contents:       contents,
		Filepath:       conf,
		ServerTypeName: serverType,
	}, nil
}

// defaultLoader loads the Corefile from the current working directory.
func defaultLoader(serverType string) (caddy.Input, error) {
	contents, err := ioutil.ReadFile(caddy.DefaultConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return caddy.CaddyfileInput{
		Contents:       contents,
		Filepath:       caddy.DefaultConfigFile,
		ServerTypeName: serverType,
	}, nil
}

// setCPU parses string cpu and sets GOMAXPROCS
// according to its value. It accepts either
// a number (e.g. 3) or a percent (e.g. 50%).
func setCPU(cpu string) error {
	var numCPU int

	availCPU := runtime.NumCPU()

	if strings.HasSuffix(cpu, `%`) {
		// Percent
		var percent float32
		pctStr := cpu[:len(cpu)-1]
		pctInt, err := strconv.Atoi(pctStr)
		if err != nil || pctInt < 1 || pctInt > 100 {
			return errors.New("invalid CPU value: percentage must be between 1-100")
		}
		percent = float32(pctInt) / 100
		numCPU = int(float32(availCPU) * percent)
	} else {
		// Number
		num, err := strconv.Atoi(cpu)
		if err != nil || num < 1 {
			return errors.New("invalid CPU value: provide a number or percent greater than 0")
		}
		numCPU = num
	}

	if numCPU > availCPU {
		numCPU = availCPU
	}

	runtime.GOMAXPROCS(numCPU)
	return nil
}

// Flags that control program flow or startup
var (
	conf    string
	cpu     string
	logfile string
	version bool
	plugins bool
)

const (
	coreName    = "CoreDNS"
	coreVersion = "006"
	serverType  = "dns"
)

func serve(c *cli.Context) error {
	ec2tags.AccessKey = c.GlobalString("aws-access-key")
	ec2tags.SecretKey = c.GlobalString("aws-secret-key")
	ec2tags.VPC = c.GlobalStringSlice("vpc")

	caddy.TrapSignals()
	caddy.DefaultConfigFile = ""
	caddy.Quiet = false //true // don't show init stuff from caddy

	caddy.RegisterCaddyfileLoader("flag", caddy.LoaderFunc(confLoader))
	caddy.SetDefaultCaddyfileLoader("default", caddy.LoaderFunc(defaultLoader))

	caddy.AppName = coreName
	caddy.AppVersion = coreVersion

	// Set up process log before anything bad happens
	switch logfile {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	default:
		log.SetOutput(os.Stdout)
	}
	log.SetFlags(log.LstdFlags)

	if plugins {
		fmt.Println(caddy.DescribePlugins())
		os.Exit(0)
	}

	// Set CPU cap
	if err := setCPU(cpu); err != nil {
		mustLogFatal(err)
	}

	// Get Corefile input
	corefile, err := caddy.LoadCaddyfile(serverType)
	if err != nil {
		mustLogFatal(err)
	}

	// Start your engines
	instance, err := caddy.Start(corefile)
	if err != nil {
		mustLogFatal(err)
	}

	log.Printf("[INFO] %s-%s\n", caddy.AppName, caddy.AppVersion)

	// Twiddle your thumbs
	instance.Wait()
	return nil
}
