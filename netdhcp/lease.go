package netdhcp

import (
	"net"
	"time"
)

// Lease represents a DHCP lease.
type Lease struct {
	MAC        net.HardwareAddr
	Expiration time.Time
}

/*
type LeaseAttr struct {
}
*/
