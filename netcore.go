package main

import (
	"fmt"
	"os"
)

func main() {
	etc := etcdSetup()
	dhcpExit := dhcpSetup(etc)
	dnsExit := dnsSetup(etc)

	fmt.Println("NETCORE Started.")

	select {
	case <-dhcpExit:
		fmt.Println("DHCP Exited")
		os.Exit(1)
	case <-dnsExit:
		fmt.Println("DNS Exited")
		os.Exit(1)
	}
}
