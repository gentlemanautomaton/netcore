package netdhcp

import (
	"net"
	"time"
)

// Prefix describes a MAC prefix and associates it with a type.
type Prefix struct {
	Attr
	Addr  net.HardwareAddr
	Label string
	Type  string
}

// PrefixChan is a channel of prefix configuration updates.
type PrefixChan <-chan PrefixUpdate

// PrefixUpdate is a prefix configuration update.
type PrefixUpdate struct {
	Type      Type
	Timestamp time.Time
	Err       error
}
