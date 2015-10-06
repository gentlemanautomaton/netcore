package netdhcp

import (
	"net"
	"time"

	"golang.org/x/net/context"
)

type GlobalContext struct {
	p Provider
}

// Context returns a context for the given provider.
func Context(p Provider) *GlobalContext {
	return &GlobalContext{p}
}

// Instance returns a context for the given instance id.
func (gc *GlobalContext) Instance(id string) InstanceContext {
	return Instance{id: id, p: &gc.p}
}

// Network returns a context for the given network id.
func (gc *GlobalContext) Network(id string) NetworkContext {
	return Network{id: id, p: &gc.p}
}

// Type returns a context for the given type id.
func (gc *GlobalContext) Type(id string) TypeContext {
	return TypeContext{id: id, p: &gc.p}
}

/*
func example(ctx context.Context) {
  p := netdhcpetcd.NewProvider()
	cfg, err := netdhcp.Context(p).Network("wen.scj.io").Type("phone.digium").Config(ctx)
	ok, err := netdhcp.Context(p).Network("wen.scj.io").Lease(ip).Create(addr)
}
*/

// InstanceContext provides access to configuration of an individual instance.
type InstanceContext struct {
	id string
	p  *Provider
}

func (ic *InstanceContext) Read(ctx context.Context) (Instance, error) {
	return ic.p.Instance(ctx, ic.id)
}

// NetworkContext provides access to configuration, type, host, and device
// attributes of a particular network. It also provides access to lease
// management.
type NetworkContext struct {
	id string
	p  *Provider
}

func (nc *NetworkContext) Config(ctx context.Context) (Config, error) {
	return
}

// Type returns a context for the given type id.
func (nc *NetworkContext) Type(id string) TypeContext {
	return TypeContext{network: nc.id, id: id, p: nc.p}
}

// Device returns a context for the given device id.
func (nc *NetworkContext) Device(device string) DeviceContext {
	return DeviceContext{network: nc.id, id: device, p: nc.p}
}

// Lease returns a context for the given IP address.
func (nc *NetworkContext) Lease(ip net.IP) LeaseContext {
	return LeaseContext{network: nc.id, id: id, p: nc.p}
}

// MAC returns a context for the given hardware address.
func (nc *NetworkContext) MAC(addr net.HardwareAddr) MACContext {
	return MACContext{network: nc.id, id: id, p: nc.p}
}

// DeviceContext provides access to the attributes of a particular device.
type DeviceContext struct {
	network string
	id      string
	p       *Provider
}

func (dc *DeviceContext) Fetch(ctx context.Context) (Device, error) {

}

// TypeContext provides access to the attributes of a particular type.
type TypeContext struct {
	network string
	id      string
	p       *Provider
}

func (tc *TypeContext) Read(ctx context.Context) (Type, error) {
	if tc.network != nil {
		return tc.p.NetworkType(ctx, tc.network, tc.id)
	}
	return tc.p.Type(ctx, tc.id)
}

// MACContext provides access to the attributes of a particular hardware
// address.
type MACContext struct {
	network string
	addr    net.HardwareAddr
	p       *Provider
}

// LeaseContext provides access to lease management functions for a particular
// IP address on a given network.
type LeaseContext struct {
	network string
	ip      net.IP
	p       *Provider
}

func (lc *LeaseContext) Create(ctx context.Context, mac net.HardwareAddr, expiration time.Time) (bool, error) {
}

func (lc *LeaseContext) Renew(ctx context.Context, mac net.HardwareAddr, expiration time.Time) (bool, error) {
}

func (lc *LeaseContext) Release(ctx context.Context, mac net.HardwareAddr) (bool, error) {
}

// Hold will attempt to place a hold on the lease IP address without assigning
// it to a particular MAC address. This is typically used by the DHCP service to
// reserve addresses for offline use.
func (lc *LeaseContext) Hold(ctx context.Context) {

}
