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
	// ErrNoConfigNetwork indicates that no network was specified in the DHCP
	// configuration.
	ErrNoConfigNetwork = errors.New("Network not specified in configuration")
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

// Attr represents a set of common attributes for a MAC.
type Attr struct {
	TFTP string
	// TODO: Adds boatloads of DHCP options
}

// Type reprsents a kind of device
type Type struct {
	TFTP string
}

// Device represents a single logical device on the network, which may have
// one or more MAC addresses associated with it.
type Device struct {
	Name  string
	Alias []string
}

// MAC represents the data associated with a specific MAC.
type MAC struct {
	Attr
	Addr   net.HardwareAddr
	Device string // FIXME: What type are we using for device IDs?
	Type   string
	Mode   IPType
	Res    []IP
	Dyn    []IP
}

// HasMode returns true if the given IP type is enabled for this MAC.
func (m *MAC) HasMode(mode IPType) bool {
	return m.Mode&IPType != 0
}

// Prefix describes a MAC prefix and associates it with a type.
type Prefix struct {
	Attr
	Addr  net.HardwareAddr
	Label string
	Type  string
}

// IPType represents whether an IP address has been dynamically assigned to a
// MAC or has been manually reserved for it. When determining which lease to
// provide to a MAC, reservations always have first priority.
type IPType uint8

const (
	// Dynamic IP addresses are assigned automatically from an IP pool.
	Dynamic IPType = 1 << iota
	// Reserved IP addresses are manually assigned to a specific MAC.
	Reserved
)

func (t IPType) String() {
	switch t {
	case Dynamic:
		return "dyn"
	case Reservation:
		return "res"
	default:
		return ""
	}
}

// IP represents an IP address assigned to a MAC address.
type IP struct {
	Creation   time.Time
	Assignment time.Time
	Priority   int
	IP         net.IP
}
