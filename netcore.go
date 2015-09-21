package main

import (
	"flag"
	"log"
	"os"

	"github.com/dustywilson/netcore/netdhcp"
	"github.com/dustywilson/netcore/netdhcpetcd"
	"github.com/dustywilson/netcore/netdns"
	"github.com/dustywilson/netcore/netdnsetcd"
)

func init() {
	flag.Parse()
}

func main() {
	log.Println("NETCORE INITIALIZING")

	inst, err := instance()
	if err != nil {
		log.Printf("FAILURE: Unable to determine instance: %s\n", err)
		os.Exit(1)
	}

	etcdclient, err := etcdClient()
	if err != nil {
		log.Printf("FAILURE: Unable to create etcd client: %s\n", err)
		os.Exit(1)
	}

	dhcpService := netdhcp.NewService(netdhcpetcd.NewProvider(etcdclient, netdhcp.DefaultConfig()), inst)
	dnsService := netdns.NewService(netdnsetcd.NewProvider(etcdclient, netdns.DefaultConfig()), inst)

	logAfterSuccess(dhcpService.Started(), "NETCORE DHCP STARTED")
	logAfterSuccess(dnsService.Started(), "NETCORE DNS STARTED")

	// FIXME: This will exit immediately if one of the services is disabled.
	select {
	case d := <-dhcpService.Done():
		if d.Initialized {
			log.Printf("NETCORE DHCP STOPPED: %s\n", d.Err)
			os.Exit(1)
		}
		log.Printf("NETCORE DHCP DID NOT START: %s\n", d.Err)
	case d := <-dnsService.Done():
		if d.Initialized {
			log.Printf("NETCORE DHCP STOPPED: %s\n", d.Err)
			os.Exit(1)
		}
		log.Printf("NETCORE DNS DID NOT START: %s\n", d.Err)
	}
}
