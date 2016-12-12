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

// BindingConfig is a source of binding configuration data.
type BindingConfig interface {
	// TODO: Consider adding a "present" or "specified" boolean as a return value
	BindingSubnet() *net.IPNet
	BindingGateway() net.IP
	BindingDomain() string
	BindingTFTP() string
	BindingNTP() net.IP
	BindingPools() PoolSet
	BindingAssignments() AssignmentSet
	BindingLeaseDuration() time.Duration
}

// MergeBindingConfig will overlay the provided sets of DHCP binding
// configuration data and return the result. The data is overlayed in order, so
// later values ovewrite earlier values.
func MergeBindingConfig(c ...BindingConfig) (attr Attr) {
	if len(c) == 0 {
		return
	}

	for _, s := range c[:] {
		if subnet := s.BindingSubnet(); subnet != nil {
			attr.Subnet = subnet
		}
		if gateway := s.BindingGateway(); gateway != nil {
			attr.Gateway = gateway
		}
		if domain := s.BindingDomain(); domain != "" {
			attr.Domain = domain
		}
		if tftp := s.BindingTFTP(); tftp != "" {
			attr.TFTP = tftp
		}
	}
	return
}

// ValidateBindingConfig returns an error if the binding configuration is
// invalid, otherwise it returns nil.
func ValidateBindingConfig(c BindingConfig) error {
	if c == nil {
		return ErrNoConfig
	}
	if c.BindingGateway() == nil {
		return ErrNoLeaseGateway
	}
	if c.BindingSubnet() == nil {
		return ErrNoLeaseSubnet
	}
	// FIXME: Check IP address assignment
	return nil
}

// Attr represents a set of attributes for a DHCP binding.
type Attr struct {
	Subnet   *net.IPNet
	Gateway  net.IP
	Domain   string
	TFTP     string
	NTP      net.IP
	Pools    PoolSet // The pools of addresses the lease will draw from for dynamic assignments
	Duration time.Duration
	Options  dhcp4.Options // TODO: Get rid of this and add a property for each option?
	// TODO: Adds boatloads of DHCP options
}

// BindingSubnet returns the subnet that the DHCP server will provide to clients
// when issuing leases.
func (a *Attr) BindingSubnet() *net.IPNet {
	return a.Subnet
}

// BindingGateway returns the gateway that the DHCP server will provide to
// clients when issuing leases.
func (a *Attr) BindingGateway() net.IP {
	return a.Gateway
}

// BindingDomain returns the domain that the DHCP server will provide to clients
// when issuing leases.
func (a *Attr) BindingDomain() string {
	return a.Domain
}

// BindingTFTP returns the TFTP address that the DHCP server provide issue to
// clients when issuing leases.
func (a *Attr) BindingTFTP() string {
	return a.TFTP
}

// BindingNTP returns the NTP address that the DHCP server will provide to
// clients when issuing leases.
func (a *Attr) BindingNTP() net.IP {
	return a.NTP
}

// BindingPools returns the IP address pool that the DHCP server will issue
// addresses from when granting leases.
func (a *Attr) BindingPools() PoolSet {
	return a.Pools
}

// TODO: Consider adding this function and have it return nothing here, then
//       override it on the MAC type. Doing this would allow the
//       server.Binding() function to return a single BindingConfig that
//       includes all of the needed information.

// BindingAssignments returns the set of IP assignments for the binding,
// including both reserved and previous dynamic IP assignments.
func (a *Attr) BindingAssignments() AssignmentSet {
	return nil
}

// BindingLeaseDuration returns the lease duration that the DHCP server will use
// when issuing leases.
func (a *Attr) BindingLeaseDuration() time.Duration {
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

/*
// MACEntry represents a MAC address record retrieved from the underlying
// provider.
type MACEntry struct {
	Network     string
	MAC         net.HardwareAddr
	Assignments AssignmentSet
	Duration    time.Duration
	Attr        map[string]string
}
*/

// Lease represents a DHCP lease.
type Lease struct {
	MAC        net.HardwareAddr
	Expiration time.Time
}

/*
type LeaseAttr struct {
}
*/

// MAC represents the binding configuration for a specific MAC address.
type MAC struct {
	Attr
	Addr        net.HardwareAddr
	Device      string // FIXME: What type are we using for device IDs?
	Type        string
	Restriction Mode // TODO: Decide whether this is inclusive or exclusive
	Assignments AssignmentSet
}

// BindingAssignments returns the set of IP assignments for the binding,
// including both reserved and previous dynamic IP assignments.
func (m *MAC) BindingAssignments() AssignmentSet {
	return m.Assignments
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
