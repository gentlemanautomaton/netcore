package netdhcp

import (
	"net"
	"time"

	"github.com/krolaw/dhcp4"
)

// ServerConfig is a source of server configuration data.
//
// TODO: Add DHCP relay configuration.
type ServerConfig interface {
	ServerEnabled() bool
	ServerNIC() string
	ServerIP() net.IP
	ServerSubnet() *net.IPNet
}

/*
type ServerConn interface {

}
*/

// MergeServerConfig will overlay the provided sets of server configuration data
// and return the result. The data is overlayed in order, so later values
// ovewrite earlier values.
func MergeServerConfig(c ...ServerConfig) (config Config) {
	if len(c) == 0 {
		return
	}

	for _, s := range c[:] {
		if enabled := s.ServerEnabled(); enabled {
			config.Enabled = enabled
		}
		if nic := s.ServerNIC(); nic != "" {
			config.NIC = nic
		}
		if ip := s.ServerIP(); ip != nil {
			config.IP = ip
		}
		if subnet := s.ServerSubnet(); subnet != nil {
			config.Subnet = subnet
		}
	}
	return
}

// Config represents server configuration data.
type Config struct {
	Enabled bool
	NIC     string     // The network adaptor the server will listen on
	IP      net.IP     // The IP address of the network adaptor
	Subnet  *net.IPNet // The subnet the server will listen to
}

// ServerEnabled returns true if the server should be enabled.
func (c *Config) ServerEnabled() bool {
	return c.Enabled
}

// ServerNIC returns the network adaptor that the DHCP server should listen on.
func (c *Config) ServerNIC() string {
	return c.NIC
}

// ServerIP returns the IP address that the DHCP server should listen on.
func (c *Config) ServerIP() net.IP {
	return c.IP
}

// ServerSubnet returns the subnet that the DHCP server should listen on.
func (c *Config) ServerSubnet() *net.IPNet {
	return c.Subnet
}

// ValidateServerConfig returns an error if the server configuration is invalid,
// otherwise it returns nil.
func ValidateServerConfig(c ServerConfig) error {
	if c == nil {
		return ErrNoConfig
	}
	if c.ServerIP() == nil {
		return ErrNoConfigIP
	}
	if c.ServerSubnet() == nil {
		return ErrNoConfigSubnet
	}
	if c.ServerNIC() == "" {
		return ErrNoConfigNIC
	}
	return nil
}

// NetworkIDConfig represents any source of network identifier configuration.
type NetworkIDConfig interface {
	NetworkID() string
}

// MergeNetworkID will overlay the provided sets of network IDs and return
// the selected network. The data is overlayed in order, so later values
// ovewrite earlier values. Empty network IDs are ignored.
func MergeNetworkID(ns ...NetworkIDConfig) (network string) {
	if len(ns) == 0 {
		return ""
	}

	for _, selector := range ns[:] {
		if id := selector.NetworkID(); id != "" {
			network = id
		}
	}
	return
}

// LeaseConfig is a source of lease configuration data.
type LeaseConfig interface {
	LeaseSubnet() *net.IPNet
	LeaseGateway() net.IP
	LeaseDomain() string
	LeaseTFTP() string
	LeaseNTP() net.IP
	LeasePool() *net.IPNet
	LeaseDuration() time.Duration
}

// MergeLeaseConfig will overlay the provided sets of DHCP lease configuration
// data and return the result. The data is overlayed in order, so later values
// ovewrite earlier values.
func MergeLeaseConfig(c ...LeaseConfig) (attr Attr) {
	if len(c) == 0 {
		return
	}

	for _, s := range c[:] {
		if subnet := s.LeaseSubnet(); subnet != nil {
			attr.Subnet = subnet
		}
		if gateway := s.LeaseGateway(); gateway != nil {
			attr.Gateway = gateway
		}
		if domain := s.LeaseDomain(); domain != "" {
			attr.Domain = domain
		}
		if tftp := s.LeaseTFTP(); tftp != "" {
			attr.TFTP = tftp
		}
	}
	return
}

// ValidateLeaseConfig returns an error if the lease configuration is invalid,
// otherwise it returns nil.
func ValidateLeaseConfig(c LeaseConfig) error {
	if c == nil {
		return ErrNoConfig
	}
	if c.LeaseGateway() == nil {
		return ErrNoLeaseGateway
	}
	if c.LeaseSubnet() == nil {
		return ErrNoLeaseSubnet
	}
	// FIXME: Check IP address assignment
	return nil
}

// Attr represents a set of attributes for a DHCP lease.
type Attr struct {
	Subnet      *net.IPNet
	Gateway     net.IP
	Domain      string
	TFTP        string
	NTP         net.IP
	Pool        *net.IPNet    // The pool of addresses the lease will draw from for dynamic assignments
	Assignments AssignmentSet // IP address assignments
	Duration    time.Duration
	Options     dhcp4.Options // TODO: Get rid of this and add a property for each option?
	// TODO: Adds boatloads of DHCP options
}

// LeaseSubnet returns the subnet that the DHCP server will provide to clients
// when issuing leases.
func (a *Attr) LeaseSubnet() *net.IPNet {
	return a.Subnet
}

// LeaseGateway returns the gateway that the DHCP server will provide to clients
// when issuing leases.
func (a *Attr) LeaseGateway() net.IP {
	return a.Gateway
}

// LeaseDomain returns the domain that the DHCP server will provide to clients
// when issuing leases.
func (a *Attr) LeaseDomain() string {
	return a.Domain
}

// LeaseTFTP returns the TFTP address that the DHCP server provide issue to
// clients when issuing leases.
func (a *Attr) LeaseTFTP() string {
	return a.TFTP
}

// LeaseNTP returns the NTP address that the DHCP server will provide to
// clients when issuing leases.
func (a *Attr) LeaseNTP() net.IP {
	return a.NTP
}

// LeasePool returns the IP address pool that the DHCP server will issue
// addresses from when granting leases.
//
// TODO: Consider making this a slice of possible pools.
func (a *Attr) LeasePool() *net.IPNet {
	return a.Pool
}

// LeaseDuration returns the lease duration that the DHCP server will use
// when issuing leases.
func (a *Attr) LeaseDuration() time.Duration {
	return a.Duration
}

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

/*
type LeaseAttr struct {
}
*/

// MAC represents the data associated with a specific MAC.
type MAC struct {
	Attr
	Addr        net.HardwareAddr
	Device      string // FIXME: What type are we using for device IDs?
	Type        string
	Restriction Mode // TODO: Decide whether this is inclusive or exclusive
	IP          AssignmentSet
}

/*
type MACAttr struct {
}
*/

// HasMode returns true if the given IP type is enabled for this MAC.
/*
func (m *MAC) HasMode(mode IPType) bool {
	return m.Mode&IPType != 0
}
*/

// Mode represents whether an IP address has been dynamically assigned to a
// MAC or has been manually reserved for it. When determining which lease to
// provide to a MAC, reservations always have first priority.
type Mode uint8

const (
	// Dynamic IP addresses are assigned automatically from an IP pool.
	Dynamic Mode = 1 << iota
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

// Assignment represents an IP address assigned to a MAC address.
type Assignment struct {
	Mode     Mode
	Priority int
	Created  time.Time
	Assigned time.Time
	Address  net.IP
}

// AssignmentSet represents a set of assigned IP addreses that can be sorted
// according to the address selection rules.
type AssignmentSet []*Assignment

func (slice AssignmentSet) Len() int {
	return len(slice)
}

func (slice AssignmentSet) Less(i, j int) bool {
	a, b := slice[i], slice[j]
	if a.Mode < b.Mode {
		return true
	}
	if a.Mode > b.Mode {
		return false
	}
	if a.Priority < b.Priority {
		return true
	}
	if a.Priority > b.Priority {
		return false
	}
	if a.Assigned.Before(b.Assigned) {
		return true
	}
	if b.Assigned.Before(a.Assigned) {
		return false
	}
	return false
}

func (slice AssignmentSet) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
