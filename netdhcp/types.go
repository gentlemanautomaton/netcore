package netdhcp

import (
	"errors"
	"net"
	"time"
)

var (
	// ErrNoConfig indicates that no configuration was provided to the DHCP
	// service.
	ErrNoConfig = errors.New("Configuration not provided")
	// ErrNoConfigIP indicates that no IP address was provided in the DHCP
	// configuration.
	ErrNoConfigIP = errors.New("IP not specified in configuration")
	// ErrNoConfigSubnet indicates that no subnet was provided in the DHCP
	// configuration.
	ErrNoConfigSubnet = errors.New("Subnet not specified in configuration")
	// ErrNoConfigNIC indicates that no network interface was provided in the
	// DHCP configuration.
	ErrNoConfigNIC = errors.New("NIC not specified in configuration")
	// ErrNotFound indicates that the requested data does not exist
	ErrNotFound = errors.New("Not found")

	// ErrNoZone is an error returned during config init to indicate that the host has not been assigned to a zone in etcd keyed off of its hostname
	ErrNoZone = errors.New("This host has not been assigned to a zone.")

	// ErrNoDHCPIP is an error returned during config init to indicate that the host has not been assigned to a zone in etcd keyed off of its hostname
	ErrNoDHCPIP = errors.New("This host has not been assigned a DHCP IP.")

	// ErrNoGateway is an error returned during config init to indicate that the zone has not been assigned a gateway in etcd keyed off of the zone name
	ErrNoGateway = errors.New("This zone does not have an assigned gateway.")
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

// MACEntry represents a MAC address record retrieved from the underlying
// provider.
type MACEntry struct {
	Network  string
	MAC      net.HardwareAddr
	IP       net.IP
	Duration time.Duration
	Attr     map[string]string
}

// Lease represents a DHCP lease.
type Lease struct {
	MAC        net.HardwareAddr
	Expiration time.Time
}

type Type struct {
}

type Device struct {
	Name  string
	Alias []string
}

type MAC struct {
}
