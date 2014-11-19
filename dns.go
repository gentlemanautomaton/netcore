package main

import (
	"fmt"

	"github.com/coreos/go-etcd/etcd"
	"github.com/miekg/dns"
)

func dnsSetup(etc *etcd.Client) chan error {
	dns.HandleFunc(".", proxyServe)
	etc.CreateDir("dhcp", 0)
	exit := make(chan error, 1)

	go func() {
		exit <- dns.ListenAndServe("0.0.0.0:53", "tcp", nil)
	}()

	go func() {
		exit <- dns.ListenAndServe("0.0.0.0:53", "udp", nil)
	}()

	return exit
}

func proxyServe(w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]

	if req.MsgHdr.Response == true { // supposed responses sent to us are bogus
		fmt.Printf("DNS Query IS BOGUS %s %s from %s.\n", q.Name, dns.Type(q.Qtype).String(), w.RemoteAddr())
		return
	}

	fmt.Printf("DNS Query %s %s from %s.\n", q.Name, dns.Type(q.Qtype).String(), w.RemoteAddr())

	c := new(dns.Client)
	m, _, err := c.Exchange(req, "8.8.8.8:53")

	if err != nil {
		fmt.Println(err)
	} else {
		w.WriteMsg(m)
	}
}
