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
	case err := <-dhcpExit:
		fmt.Printf("DHCP Exited: %s\n", err)
		os.Exit(1)
	case err := <-dnsExit:
		fmt.Printf("DNS Exited: %s\n", err)
		os.Exit(1)
	}
}
