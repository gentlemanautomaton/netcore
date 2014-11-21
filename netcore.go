package main

import (
	"flag"
	"fmt"
	"os"
)

var etcdServers = flag.String("etcd", "", "Comma-separated list of etcd servers.")

func init() {
	flag.Parse()
}

func main() {
	etc := etcdSetup(*etcdServers)

	cfg, err := getConfig(etc)

	if err != nil {
		fmt.Printf("Configuration failed: %s\n", err)
		os.Exit(1)
	}

	var dhcpExit chan error
	if cfg.DHCPIP() == nil {
		fmt.Println("DHCP service is disabled; this machine does not have a DHCP IP assigned.")
	} else if cfg.DHCPSubnet() == nil {
		fmt.Println("DHCP service is disabled; this machine's zone does not have a DHCP subnet assigned.")
	} else if cfg.DHCPNIC() == "" {
		fmt.Println("DHCP service is disabled; this machine does not have a DHCP NIC assigned.")
	} else {
		dhcpExit = dhcpSetup(cfg, etc)
	}

	dnsExit := dnsSetup(cfg, etc)

	fmt.Println("NETCORE Started.")

	select {
	case err := <-dhcpExit:
		fmt.Printf("DHCP Exited: %s\n", err)
		os.Exit(1)
	case err := <-dnsExit:
		fmt.Printf("DNS Exited: %s\n", err)
		os.Exit(1)
	}
}
