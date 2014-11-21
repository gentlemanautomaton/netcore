package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/coreos/go-etcd/etcd"
	"github.com/miekg/dns"
)

type dnsResult struct {
	rrType   string
	name     string
	ttl      int
	priority int
	weight   int
	target   string
}

func dnsSetup(cfg *Config, etc *etcd.Client) chan error {
	dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) { dnsQueryServe(cfg, etc, w, req) })
	etc.CreateDir("dns", 0)
	exit := make(chan error, 1)

	go func() {
		exit <- dns.ListenAndServe("0.0.0.0:53", "tcp", nil)
	}()

	go func() {
		exit <- dns.ListenAndServe("0.0.0.0:53", "udp", nil)
	}()

	return exit
}

func dnsQueryServe(cfg *Config, etc *etcd.Client, w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]

	if req.MsgHdr.Response == true { // supposed responses sent to us are bogus
		fmt.Printf("DNS Query IS BOGUS %s %s from %s.\n", q.Name, dns.Type(q.Qtype).String(), w.RemoteAddr())
		return
	}

	fmt.Printf("DNS Query %s %s from %s.\n", q.Name, dns.Type(q.Qtype).String(), w.RemoteAddr())

	// TODO: lookup in an in-memory cache (obeying TTLs!)

	pathParts := strings.Split(strings.TrimSuffix(q.Name, "."), ".") // breakup the queryed name
	queryPath := strings.Join(reverseSlice(pathParts), "/")          // reverse and join them with a slash delimiter
	key := "dns/" + queryPath + "/@" + dns.Type(q.Qtype).String()    // structure the lookup key
	fmt.Println(strings.ToLower(key))
	response, err := etc.Get(strings.ToLower(key), true, true) // do the lookup
	if err == nil && response != nil && response.Node != nil && len(response.Node.Nodes) > 0 {
		for _, node := range response.Node.Nodes {
			fmt.Printf("[%+v]\n", node)
		}
		os.Exit(1)
	}

	forwarders := cfg.DNSForwarders()
	if len(forwarders) == 0 {
		// we have no upstreams, so we'll just not use any
	} else if strings.TrimSpace(forwarders[0]) == "!" {
		// we've been told explicitly to not pass anything along to any upsteams
	} else {
		c := new(dns.Client)
		for _, server := range forwarders {
			c.Net = "udp"
			m, _, err := c.Exchange(req, strings.TrimSpace(server))

			if m != nil && m.MsgHdr.Truncated {
				c.Net = "tcp"
				m, _, err = c.Exchange(req, strings.TrimSpace(server))
			}

			if err != nil {
				fmt.Println(err)
			} else {
				w.WriteMsg(m)
				return // because we're done here
			}
		}
	}

	// if we got here, it means we didn't find what we were looking for.
	return // without writing a message, I think it just returns NXDOMAIN

	// FIXME: should we be sending explicit NXDOMAIN (or other failures), or is this good enough?
}
