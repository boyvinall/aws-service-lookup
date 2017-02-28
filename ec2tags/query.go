package ec2tags

import (
	"log"
	"net"
	"regexp"

	"strings"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/pkg/errors"
)

type Host struct {
	PrivateIPAddress string
	PublicIPAddress  string
	PrivateDNSName   string
	Tags             map[string]string
}

type Hosts []Host

func Query(accesskey, secretkey string, vpc map[string]struct{}, running bool) (Hosts, error) {
	if accesskey == "" {
		// TODO: try to get key from EC2 metadata service
		return nil, errors.New("accesskey is empty")
	} else if secretkey == "" {
		return nil, errors.New("secretkey is empty")
	}

	auth := aws.Auth{AccessKey: accesskey, SecretKey: secretkey}
	e := ec2.New(auth, aws.EUWest)
	resp, err := e.Instances(nil, nil)
	if err != nil {
		return nil, err
	}

	hosts := make([]Host, 0)

	// log.Printf("%d Reservations", len(resp.Reservations))
	for _, res := range resp.Reservations {
		// log.Printf("%d Instances", len(res.Instances))
		for _, inst := range res.Instances {
			if _, ok := vpc[inst.VpcId]; len(vpc) > 0 && !ok {
				continue
			}

			// b, err := json.MarshalIndent(inst, "", "  ")
			// if err == nil {
			// 	log.Printf("%+v", string(b))
			// }

			h := Host{
				PrivateIPAddress: inst.PrivateIpAddress,
				PublicIPAddress:  inst.PublicIpAddress,
				PrivateDNSName:   inst.PrivateDNSName,
				Tags:             make(map[string]string, 0),
			}

			for _, t := range inst.Tags {
				h.Tags[t.Key] = t.Value
			}

			hosts = append(hosts, h)
		}
	}

	return hosts, nil
}

func (hosts Hosts) Records() map[string][]net.IP {
	re, err := regexp.Compile("ip-[^.]*")
	if err != nil {
		log.Printf("unable to compile regex: %s", err.Error())
		return nil
	}

	r := make(map[string][]net.IP, 0)
	for _, h := range hosts {
		name := re.ReplaceAllString(h.PrivateDNSName, h.Tags["Name"]) + "."
		ip := net.ParseIP(h.PrivateIPAddress)
		if ip == nil {
			continue
		}
		r[name] = append(r[name], ip)

		services := h.Tags["Services"]
		if services == "" {
			continue
		}

		for _, service := range strings.Split(services, " ") {
			service = service + ".service.local."
			r[service] = append(r[service], ip)
		}
	}

	return r
}
