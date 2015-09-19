package main

import (
	"flag"
	"log"
	"os"

	"dustywilson/netcore/netdhcp"
	"dustywilson/netcore/netdhcpetcd"
	"dustywilson/netcore/netdns"
	"dustywilson/netcore/netdnsetcd"
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

	dhcpService := netdhcp.NewService(netdhcpetcd.NewProvider(etcdclient, netdhcp.DefaultConfig()), instance)
	dnsService := netdns.NewService(netdnsetcd.NewProvider(etcdclient, netdns.DefaultConfig()), instance)

	// TODO: Print NETCORE [SERVICE] STARTED for each service when they become
	//       ready.
	log.Println("NETCORE STARTED")

	// TODO: Make sure this exits properly if neither service is enabled.
	select {
	case d, ok := <-dhcpService.Done():
		if ok {
			if d.Initialized {
				log.Printf("NETCORE DHCP EXITED: %s\n", d.Err)
				os.Exit(1)
			}
			log.Printf("NETCORE DHCP NOT STARTED: %s\n", d.Err)
		}
	case d, ok := <-dnsService.Done():
		if ok {
			if d.Initialized {
				log.Printf("NETCORE DHCP EXITED: %s\n", d.Err)
				os.Exit(1)
			}
			log.Printf("NETCORE DNS NOT STARTED: %s\n", d.Err)
		}
	}
}
