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

	dhcpDone := dhcpService.Done()
	dnsDone := dnsService.Done()

	for running := 2; running > 0; running-- {
		select {
		case d := <-dhcpDone:
			if d.Initialized {
				log.Printf("NETCORE DHCP STOPPED: %s\n", d.Err)
				os.Exit(1) // FIXME: Attempt graceful shutdown first?
			}
			dhcpDone = nil // Read from each channel once
			log.Printf("NETCORE DHCP DID NOT START: %s\n", d.Err)
			// FIXME: Evaluate the reason why the service couldn't start and take
			//        appropriate action.
		case d := <-dnsDone:
			if d.Initialized {
				log.Printf("NETCORE DNS STOPPED: %s\n", d.Err)
				os.Exit(1) // FIXME: Attempt graceful shutdown first?
			}
			dnsDone = nil // Read from each channel once
			log.Printf("NETCORE DNS DID NOT START: %s\n", d.Err)
			// FIXME: Evaluate the reason why the service couldn't start and take
			//        appropriate action.
		}
	}
}
