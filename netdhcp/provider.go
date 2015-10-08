package netdhcp

import (
	"net"
	"time"

	"golang.org/x/net/context"
)

// Provider carries all storage interfaces necessary for operation of
// the DHCP service.
/*
type Provider struct {
	ConfigProvider
	DeviceProvider
	LeaseProvider
	InstanceProvider
	NetworkProvider
	TypeProvider
}
*/

// Provider carries all storage interfaces necessary for operation of
// the DHCP service.
type Provider struct {
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

// Provider implements all storage interfaces necessary for operation of
// the DHCP service.
/*
type Provider struct {
	Global
	InstanceProvider
	NetworkProvider
}
*/

// Global provides access to configuration, type, host, and MAC attributes
// shared across all networks.
/*
type Global struct {
	ConfigProvider
	Type   TypeProvider
	Device DeviceProvider
	MAC    MACProvider
}
*/

// ConfigProvider provides DHCP configuration at global, network and instance
// levels.
/*
type ConfigProvider interface {
	Config() (Config, error) // TODO: Add ctx?
}
*/

// GlobalProvider provides access to global configuration data.
type GlobalProvider interface {
	Global(ctx context.Context) (Global, error)
}

// InstanceProvider provides access to instance configuration data.
type InstanceProvider interface {
	Instance(ctx context.Context, id string) (Instance, error)
}

// NetworkProvider provides access to Network data.
type NetworkProvider interface {
	Network(ctx context.Context, id string) (Network, error)
}

// TypeProvider provides access to type data.
type TypeProvider interface {
	Type(ctx context.Context, id string) (Type, error)
	TypeList(ctx context.Context) ([]Type, error)
}

// NetworkTypeProvider provides access to type data for a particular network.
type NetworkTypeProvider interface {
	NetworkType(ctx context.Context, network string, id string) (Type, error)
	NetworkTypeList(ctx context.Context, network string) ([]Type, error)
}

// DeviceProvider provides access to global device data.
type DeviceProvider interface {
	Device(ctx context.Context, device string) (Device, error)
	DeviceList(ctx context.Context) ([]Device, error)
}

// NetworkDeviceProvider provides access to device data for a particular
// network.
type NetworkDeviceProvider interface {
	NetworkDevice(ctx context.Context, network string, device string) (Device, error)
	NetworkDeviceList(ctx context.Context, network string) ([]Device, error)
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
func hostname() string {
	return "server.oly.scj.io"
}

// merge combines configuration data in sensible way
func merge(c1 Config, c2 Config) Config {
	return NewConfig(&Cfg{}) // TODO: Write me
}

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

// Old design:
/*
// Provider implements all storage interfaces necessary for operation of
// the DHCP service.
type Provider struct {
	ConfigProvider
	DataProvider
}

// ConfigProvider provides DHCP configuration to the DHCP service.
type ConfigProvider interface {
	Init() error
	Config() (Config, error)
	//ConfigStream() chan<- Config
}

// DataProvider provides DHCP data to the DHCP service.
type DataProvider interface {
	IP(net.IP) (IPEntry, error)
	HasIP(net.IP) bool
	MAC(mac net.HardwareAddr, cascade bool) (entry *MACEntry, found bool, err error)
	RenewLease(lease *MACEntry) error
	CreateLease(lease *MACEntry) error
	WriteLease(lease *MACEntry) error
	RegisterA(fqdn string, ip net.IP, exclusive bool, ttl uint32, expiration time.Duration) error
}
*/
