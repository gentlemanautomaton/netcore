package netdhcp

import (
	"net"
	"time"
)

// Completion is returned via the Service.Done() channel when the service exits.
type Completion struct {
	// Initialized indiciates whether the services finished initializing before exiting.
	Initialized bool
	// Err indictes the error that caused the service to exit in the case of
	// failure.
	Err error
}

// IPEntry represents an IP address allocation retrieved from the underlying
// provider.
type IPEntry struct {
	MAC net.HardwareAddr
}

// Mode represents whether an IP address has been dynamically assigned to a
// MAC or has been manually reserved for it. When determining which lease to
// provide to a MAC, reservations always have first priority.
type Mode uint8

const (
	// Dynamic IP addresses are assigned automatically from an IP pool.
	Dynamic Mode = iota + 1
	// Reserved IP addresses are manually assigned to a specific MAC.
	Reserved
)

func (mode Mode) String() string {
	switch mode {
	case Dynamic:
		return "dyn"
	case Reserved:
		return "res"
	default:
		return ""
	}
}

// Pool represents an IP address pool from which IP addresses can be dynamically
// assigned.
type Pool struct {
	Name     string
	Mode     Mode
	Priority int
	Created  time.Time
	Range    *net.IPNet
}

// PoolSet represents a set of IP address pools that can be sorted
// according to the Pool selection rules.
type PoolSet []*Pool

func (slice PoolSet) Len() int {
	return len(slice)
}

func (slice PoolSet) Less(i, j int) bool {
	a, b := slice[i], slice[j]
	if a.Priority < b.Priority {
		return true
	}
	if a.Priority > b.Priority {
		return false
	}
	if a.Created.Before(b.Created) {
		return true
	}
	if b.Created.Before(a.Created) {
		return false
	}
	return false
}

func (slice PoolSet) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
