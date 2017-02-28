package ec2tags

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/middleware"
	"github.com/coredns/coredns/request"
	"github.com/mholt/caddy"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

const PluginName = "ec2tags"
const ConfigDirective = "ec2tags"

func init() {
	caddy.RegisterPlugin(PluginName, caddy.Plugin{
		ServerType: "dns",
		Action: func(c *caddy.Controller) error {
			return CreatePlugin(c)
		},
	})

	// This is a bad thing to do, but is the only thing possible
	// until/unless the plugin is upstreamed.
	dnsserver.RegisterDevDirective(ConfigDirective, "")
}

var (
	AccessKey = ""
	SecretKey = ""
	VPC       = []string{}
)

type Plugin struct {
	Next   middleware.Handler
	Domain string
	VPC    map[string]struct{}
	TTL    int
}

func CreatePlugin(c *caddy.Controller) error {
	p := &Plugin{
		Domain: ".local",
		TTL:    120,
		VPC:    make(map[string]struct{}, 0),
	}

	for _, k := range VPC {
		p.VPC[k] = struct{}{}
	}

	for c.Next() {
		args := c.RemainingArgs()
		fmt.Printf("aws config %+v\n", args)
		for c.NextBlock() {
			switch c.Val() {
			case "domain":
				a := c.RemainingArgs()
				if len(a) != 1 {
					return middleware.Error(PluginName, errors.New("domain takes exactly one arg"))
				}
				p.Domain = a[0]
			case "ttl":
				a := c.RemainingArgs()
				if len(a) != 1 {
					return middleware.Error(PluginName, errors.New("ttl takes exactly one arg"))
				}
				var err error
				p.TTL, err = strconv.Atoi(a[0])
				if err != nil {
					return err
				}
			}
		}
	}

	dnsserver.GetConfig(c).AddMiddleware(func(next middleware.Handler) middleware.Handler {
		p.Next = next
		return p
	})

	return nil
}

// ServeDNS implements the middleware.Handler interface.
func (p *Plugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	if state.QType() != dns.TypeA {
		return middleware.NextOrFailure(p.Name(), p.Next, ctx, w, r)
	}

	if !strings.HasSuffix(state.Name(), p.Domain) {
		// return p.Next.ServeDNS(ctx, w, r)
		return middleware.NextOrFailure(p.Name(), p.Next, ctx, w, r)
	}

	m := new(dns.Msg)
	m.SetReply(r)

	h, err := Query(AccessKey, SecretKey, p.VPC, false)
	if err != nil {
		return 0, err
	}

	// log.Printf("h %+v", h)

	hdr := dns.RR_Header{
		Name:   state.QName(),
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    uint32(p.TTL),
	}

	name := state.Name()
	// qname := state.QName()
	for n, ips := range h.Records() {
		// log.Printf("[%s] [%s] [%s] %+v", n, name, qname, ips)
		if n == name {
			for _, ip := range ips {
				m.Answer = append(m.Answer, &dns.A{Hdr: hdr, A: ip})
			}
		}
	}

	state.SizeAndDo(m)
	err = w.WriteMsg(m)
	return 0, err
}

// Name implements the Handler interface.
func (p *Plugin) Name() string {
	return PluginName
}
