package main

import (
	"flag"
	"log"
	"os"
)

func init() {
	flag.Parse()
}

func main() {
	log.Println("INIT PRECONFIG")
	etcdclient := etcdClient()
	dhcpProvider := netdhcpetcd.NewProvider(etcdclient)
	dhcpConfig := dhcpProvider.Config()
	dnsProvider := netdnsetcd.NewProvider(etcdclient)
	cfg, err := db.GetConfig()
	log.Println("INIT POSTCONFIG")

	if err != nil {
		log.Printf("Configuration failed: %s\n", err)
		os.Exit(1)
	}

	var dhcpExit chan error
	if cfg.DHCPIP() == nil {
		log.Println("DHCP service is disabled; this machine does not have a DHCP IP assigned.")
	} else if cfg.DHCPSubnet() == nil {
		log.Println("DHCP service is disabled; this machine's zone does not have a DHCP subnet assigned.")
	} else if cfg.DHCPNIC() == "" {
		log.Println("DHCP service is disabled; this machine does not have a DHCP NIC assigned.")
	} else {
		dhcpExit = dhcpSetup(cfg)
	}

	dnsExit := dnsSetup(cfg)

	log.Println("NETCORE Started.")

	select {
	case err := <-dhcpExit:
		log.Printf("DHCP Exited: %s\n", err)
		os.Exit(1)
	case err := <-dnsExit:
		log.Printf("DNS Exited: %s\n", err)
		os.Exit(1)
	}
}
