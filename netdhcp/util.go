package netdhcp

import "net"

func macPrefixes(mac net.HardwareAddr) []net.HardwareAddr {
	// Copy cascaded attributes by making recursive calls to this function
	p := make([]net.HardwareAddr, 0, len(mac))
	for len(mac) > 0 {
		p = append(p, mac)
		mac = mac[0 : len(mac)-1]
	}
	return p
}
