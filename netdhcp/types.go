package netdhcp

import (
	"net"
	"time"
)

// IPEntry represents an IP address allocation retrieved from the underlying
// provider.
type IPEntry struct {
	MAC net.HardwareAddr
}

// MACEntry represents a MAC address record retrieved from the underlying
// provider.
type MACEntry struct {
	MAC      net.HardwareAddr
	IP       net.IP
	Duration time.Duration
	Attr     map[string]string
}
