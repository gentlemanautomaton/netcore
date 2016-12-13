package netdhcp

import (
	"net"
	"time"

	"github.com/krolaw/dhcp4"
)

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
