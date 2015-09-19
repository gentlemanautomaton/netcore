package netdhcp

import (
	"errors"
	"net"
	"time"
)

var (
	// ErrNotFound indicates that the requested data does not exist
	ErrNotFound = errors.New("not found")
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
