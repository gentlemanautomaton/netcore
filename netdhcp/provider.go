package netdhcp

import (
	"net"
	"time"

	"golang.org/x/net/context"
)

// Provider implements all storage interfaces necessary for operation of
// the DHCP service.
type Provider struct {
	Global
	InstanceProvider
	NetworkProvider
}

// Global provides access to configuration, type, host, and MAC attributes
// shared across all networks.
type Global struct {
	ConfigProvider
	Type   TypeProvider
	Device DeviceProvider
	MAC    MACProvider
}

// InstanceProvider provides access to Instance data.
type InstanceProvider interface {
	Instance(instance string) Instance
}

// NetworkProvider provides access to Network data.
type NetworkProvider interface {
	Network(network string) Network
}

// Instance provides access to configuration of an individual instance.
type Instance struct {
	ID string
	ConfigProvider
}

// Network provides access to configuration, type, host, and device attributes
// of a particular network. It also provides an interface for lease management.
type Network struct {
	ID string
	ConfigProvider
	Type   TypeProvider
	Device DeviceProvider
	MAC    MACProvider
	Lease  LeaseProvider
}

// ConfigProvider provides DHCP configuration at global, network and instance
// levels.
type ConfigProvider interface {
	Init() error
	Config() (Config, error)
}

// TypeProvider provides access to type data.
type TypeProvider interface {
	Lookup(ctx context.Context, id string) (Type, bool, error)
}

// DeviceProvider provides access to device data.
type DeviceProvider interface {
	Lookup(ctx context.Context, id string) (Device, bool, error)
}

// MACProvider provides access to MAC data.
type MACProvider interface {
	Lookup(ctx context.Context, addr net.HardwareAddr) (MAC, bool, error)
	Assign(ctx context.Context, addr net.HardwareAddr, ip net.IP) (bool, error)
}

// LeaseProvider provides access to lease data.
type LeaseProvider interface {
	Lookup(ctx context.Context, ip net.IP) (Lease, bool, error)
	Create(ctx context.Context, ip net.IP, mac net.HardwareAddr, expiration time.Time) (bool, error)
	Renew(ctx context.Context, ip net.IP, mac net.HardwareAddr, expiration time.Time) (bool, error)
	Release(ctx context.Context, ip net.IP, mac net.HardwareAddr) (bool, error)
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
