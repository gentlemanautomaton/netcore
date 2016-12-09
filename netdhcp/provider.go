package netdhcp

import (
	"context"
	"net"
	"time"
)

// Provider is the interface that must be fulfilled by providers of DHCP
// configuration data.
type Provider interface {
	GlobalProvider
	InstanceProvider
	NetworkProvider
	PrefixProvider
	TypeProvider
	DeviceProvider
	MACProvider
	NetworkPrefixProvider
	NetworkTypeProvider
	NetworkDeviceProvider
	NetworkMACProvider
	NetworkLeaseProvider
	// TODO: Add interface for Provider state change notification, including Disconnected, and LocalOnly states
}

/*
// MixedProvider carries all storage interfaces necessary for operation of
// the DHCP service.
type MixedProvider struct {
	GlobalProvider
	InstanceProvider
	NetworkProvider
	TypeProvider
	DeviceProvider
	MACProvider
	NetworkTypeProvider
	NetworkDeviceProvider
	NetworkMACProvider
	NetworkLeaseProvider
}
*/

// GlobalProvider provides access to global configuration data.
type GlobalProvider interface {
	Global(ctx context.Context) (Global, error)
}

// GlobalChanProvider provides a channel of global configuration changes.
type GlobalChanProvider interface {
	GlobalChan() GlobalChan // TODO: Consider moving this into its own interface that is optionally supported
}

// InstanceProvider provides access to instance configuration data.
type InstanceProvider interface {
	Instance(ctx context.Context, id string) (Instance, error)
}

// InstanceChanProvider provides a channel of instance configuration changes.
type InstanceChanProvider interface {
	InstanceChan(ctx context.Context, id string) InstanceChan
}

// NetworkProvider provides access to network configuration.
type NetworkProvider interface {
	Network(ctx context.Context, id string) (Network, error)
}

// NetworkChanProvider provides a channel of network configuration changes.
type NetworkChanProvider interface {
	NetworkChan(ctx context.Context, id string) NetworkChan
}

// PrefixProvider provides access to global MAC prefix configuration.
type PrefixProvider interface {
	Prefix(ctx context.Context, addr net.HardwareAddr) (Prefix, error)
	PrefixList(ctx context.Context) ([]Prefix, error)
}

// PrefixChanProvider provides a channel of global type configuration changes.
type PrefixChanProvider interface {
	PrefixChan(ctx context.Context, addr net.HardwareAddr) PrefixChan
}

// NetworkPrefixProvider provides access to type data for a particular network.
type NetworkPrefixProvider interface {
	NetworkPrefix(ctx context.Context, network string, addr net.HardwareAddr) (Prefix, error)
	NetworkPrefixList(ctx context.Context, network string) ([]Prefix, error)
}

// NetworkPrefixChanProvider provides a channel of network type configuration
// changes.
type NetworkPrefixChanProvider interface {
	NetworkPrefixChan(ctx context.Context, network string, addr net.HardwareAddr) PrefixChan
}

// TypeProvider provides access to global type configuration.
type TypeProvider interface {
	Type(ctx context.Context, id string) (Type, error)
	TypeList(ctx context.Context) ([]Type, error)
}

// TypeChanProvider provides a channel of global type configuration changes.
type TypeChanProvider interface {
	TypeChan(ctx context.Context, id string) TypeChan
}

// NetworkTypeProvider provides access to type data for a particular network.
type NetworkTypeProvider interface {
	NetworkType(ctx context.Context, network string, id string) (Type, error)
	NetworkTypeList(ctx context.Context, network string) ([]Type, error)
}

// NetworkTypeChanProvider provides a channel of network type configuration
// changes.
type NetworkTypeChanProvider interface {
	NetworkTypeChan(ctx context.Context, network string, id string) TypeChan
}

// DeviceProvider provides access to global device data.
type DeviceProvider interface {
	Device(ctx context.Context, device string) (Device, error)
	DeviceList(ctx context.Context) ([]Device, error)
}

// DeviceChanProvider provides a channel of global device configuration
// changes.
type DeviceChanProvider interface {
	DeviceChan(ctx context.Context, id string) DeviceChan
}

// NetworkDeviceProvider provides access to device data for a particular
// network.
type NetworkDeviceProvider interface {
	NetworkDevice(ctx context.Context, network string, device string) (Device, error)
	NetworkDeviceList(ctx context.Context, network string) ([]Device, error)
}

// NetworkDeviceChanProvider provides a channel of network device configuration
// changes.
type NetworkDeviceChanProvider interface {
	NetworkDeviceChan(ctx context.Context, network string, id string) DeviceChan
}

// MACProvider provides access to MAC data.
type MACProvider interface {
	MAC(ctx context.Context, addr net.HardwareAddr) (MAC, error)
	MACList(ctx context.Context) ([]MAC, error)
}

// NetworkMACProvider provides access to MAC data for a particular network.
type NetworkMACProvider interface {
	NetworkMAC(ctx context.Context, network string, addr net.HardwareAddr) (MAC, error)
	NetworkMACList(ctx context.Context, network string) ([]MAC, error)
	NetworkMACAssign(ctx context.Context, network string, addr net.HardwareAddr, mode Mode, ip net.IP, priority int) (bool, error)
}

// NetworkLeaseProvider provides access to lease data for a particular network.
type NetworkLeaseProvider interface {
	NetworkLease(ctx context.Context, network string, ip net.IP) (*Lease, error)
	NetworkLeaseList(ctx context.Context, network string) ([]*Lease, error)
	NetworkLeaseCreate(ctx context.Context, network string, ip net.IP, addr net.HardwareAddr, expiration time.Time) (bool, error)
	NetworkLeaseRenew(ctx context.Context, network string, ip net.IP, addr net.HardwareAddr, expiration time.Time) (bool, error)
	NetworkLeaseRelease(ctx context.Context, network string, ip net.IP, addr net.HardwareAddr) (bool, error)
	NetworkLeaseHold(ctx context.Context, network string, ip net.IP) (bool, error)
}

/*
// example demonstrates how to use a provider to create a new lease
func exampleCreate(p Provider, ip net.IP, mac net.HardwareAddr) {
	gc, _ := p.Config()
	ic, _ := p.Instance(hostname()).Config()
	cfg := merge(gc, ic)
	network := p.Network(cfg.Network())
	nc, _ := network.Config()
	cfg = merge(cfg, nc)
	duration := time.Hour * 6
	expiration := time.Now().Add(duration)
	if ok, _ := network.Lease.Create(context.Background(), ip, mac, expiration); ok {
		network.MAC.Assign(context.Background(), mac, ip)
	}
}
*/
