package netdhcp

import (
	"errors"
	"net"
	"time"

	"golang.org/x/net/context"
)

/*
func example1(ctx context.Context) {
  p := netdhcpetcd.NewProvider()
	wen, err := netdhcp.Context(p).Network("wen.scj.io").Read(ctx)
	digium, err := gc.Type("phone.digium").Read(ctx)
	wendigium, err := netdhcp.Context(p).Network("wen.scj.io").Type("phone.digium").Read(ctx)
	ok, err := netdhcp.Context(p).Network("wen.scj.io").Lease(ip).Create(ctx, addr)
}

func example2(ctx context.Context, network string, ip net.IP, addr net.HardwareAddr) {
  p := netdhcpetcd.NewProvider()
	ok, err := netdhcp.Context(p).Network(network).Lease(ip).Create(ctx, addr)
}

func example3(ctx context.Context, instance string, addr net.HardwareAddr) {
  p := netdhcpetcd.NewProvider()
	gc := netdhcp.NewContext(p).Observe()
	ic := gc.Instance(instance).Observe()
	nc := gc.Network(iconfig.Network).Observe()
	defer gc.Close()
	defer ic.Close()
	defer nc.Close()
	global, err := gc.Read(ctx)
	inst, err := ic.Read(ctx)
	network, err := nc.Read(ctx)
	// TODO: Check errors
	cfg := MergeConfig(global, inst, network) // global + network + instance
	attr := MergeAttr(global, inst, network)
	nmac, err := nc.MAC(ctx, addr)
	if nmac.Type != "" {
		gtc := gc.Type(mac.IP)
		ntc := nc.Type(mac.IP)
		attr = MergeAttr(attr, gtc.Attr, ntc.Attr) // global + network + instance + type
	}
	if nmac.Device != "" {
		gdc := gc.Device(mac.Device)
		ndc := nc.Device(mac.Device)
		attr = MergeAttr(attr, gdc.Attr, ndc.Attr) // global + network + instance + type + device
	}
	gmac, err := gc.MAC(ctx, addr)
	attr = MergeAttr(attr, gmac.Attr, nmac.Attr) // global + network + instance + type + device + mac

	// cfg now contains the collapsed configuration
}
*/

// GlobalContext provides access to all netcore DHCP data. The data is retrieved
// from the provider that was supplied in the call to NewContext().
type GlobalContext struct {
	p Provider
}

// NewContext returns a context for the given provider.
func NewContext(p Provider) GlobalContext {
	return GlobalContext{p}
}

// Read returns a copy of the global configuration.
func (gc GlobalContext) Read(ctx context.Context) (*Global, error) {
	global, err := gc.p.Global(ctx)
	if err != nil {
		return nil, err
	}
	return &global, nil
}

/*
// Observe returns an observed view of the global configuration data that will
// will always have the most recent data.
//
// The returned observer consumes resources as long as it is active. It is
// the caller's responsibility to call Close() on the observer when
// finished with it.
func (gc GlobalContext) Observe(opt WatcherOptions) GlobalObserver {
	// FIXME: When do we check validity of gc?
	return gc.p.GlobalObserver(opt)
}
*/

// Instance returns a context for the given instance id.
func (gc GlobalContext) Instance(id string) InstanceContext {
	return InstanceContext{id: id, p: gc.p}
}

// Network returns a context for the given network id.
func (gc GlobalContext) Network(id string) NetworkContext {
	return NetworkContext{id: id, p: gc.p}
}

// Prefix returns a context for the given MAC prefix.
func (gc GlobalContext) Prefix(addr net.HardwareAddr) PrefixContext {
	return PrefixContext{addr: addr, p: gc.p}
}

// Type returns a context for the given type id.
func (gc GlobalContext) Type(id string) TypeContext {
	return TypeContext{id: id, p: gc.p}
}

// Device returns a context for the given device id.
func (gc GlobalContext) Device(device string) DeviceContext {
	return DeviceContext{id: device, p: gc.p}
}

// InstanceContext provides access to configuration of an individual instance.
type InstanceContext struct {
	id string
	p  Provider
}

// valid returns an error if the instance context is invalid.
func (ic InstanceContext) valid() error {
	if ic.id == "" {
		// TODO: Return error
	}
	return nil
}

// Read returns a copy of the instance configuration.
func (ic InstanceContext) Read(ctx context.Context) (*Instance, error) {
	if err := ic.valid(); err != nil {
		return nil, err
	}
	instance, err := ic.p.Instance(ctx, ic.id)
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

/*
// Observe returns an observed view of the global configuration data, meaning
// it will always have the most recent data.
//
// The returned observer consumes resources as long as it is active. It is
// the caller's responsibility to call Close() on the observer when
// finished with it.
func (ic InstanceContext) Observe() InstanceObserver {
	// FIXME: When do we check validity of ic?
	return ic.p.InstanceObserver(ic.id)
}
*/

/*
// Watcher returns an InstanceWatcher that provides notification of instance
// configuration changes.
func (ic InstanceContext) Watcher(opt WatcherOptions) InstanceWatcher {
}
*/

// NetworkContext provides access to configuration, type, host, and device
// attributes of a particular network. It also provides access to lease
// management.
type NetworkContext struct {
	id string
	p  Provider
}

// valid returns an error if the network context is invalid.
func (nc NetworkContext) valid() error {
	if nc.id == "" {
		// TODO: Return error
	}
	return nil
}

// Read returns a copy of the network configuration.
func (nc NetworkContext) Read(ctx context.Context) (*Network, error) {
	if err := nc.valid(); err != nil {
		return nil, err
	}
	network, err := nc.p.Network(ctx, nc.id)
	if err != nil {
		return nil, err
	}
	return &network, nil
}

/*
// Observe returns a NetworkObserver that provides notification of network
// configuration changes.
func (nc *NetworkContext) Observe() NetworkObserver {
	// FIXME: When do we check validity of nc?
	return nc.p.NetworkObserver(nc.id)
}
*/

// Prefix returns a context for the given MAC prefix.
func (nc NetworkContext) Prefix(addr net.HardwareAddr) PrefixContext {
	return PrefixContext{network: nc.id, addr: addr, p: nc.p}
}

// Type returns a context for the given type id.
func (nc NetworkContext) Type(id string) TypeContext {
	return TypeContext{network: nc.id, id: id, p: nc.p}
}

// Device returns a context for the given device id.
func (nc NetworkContext) Device(device string) DeviceContext {
	return DeviceContext{network: nc.id, id: device, p: nc.p}
}

// Lease returns a context for the given IP address.
func (nc NetworkContext) Lease(ip net.IP) LeaseContext {
	return LeaseContext{network: nc.id, ip: ip, p: nc.p}
}

// MAC returns a context for the given hardware address.
func (nc NetworkContext) MAC(addr net.HardwareAddr) MACContext {
	return MACContext{network: nc.id, addr: addr, p: nc.p}
}

// PrefixContext provides access to the attributes of a MAC address prefix.
type PrefixContext struct {
	network string
	addr    net.HardwareAddr
	p       Provider
}

// valid returns an error if the MAC prefix context is invalid.
func (pc PrefixContext) valid() error {
	if pc.addr == nil {
		// TODO: Return error
	}
	return nil
}

// Read returns a copy of the MAC prefix configuration.
func (pc PrefixContext) Read(ctx context.Context) (*Prefix, error) {
	if err := pc.valid(); err != nil {
		return nil, err
	}
	var (
		prefix Prefix
		err    error
	)
	if pc.network != "" {
		prefix, err = pc.p.NetworkPrefix(ctx, pc.network, pc.addr)
	} else {
		prefix, err = pc.p.Prefix(ctx, pc.addr)
	}
	if err != nil {
		return nil, err
	}
	return &prefix, nil
}

// TypeContext provides access to the attributes of a particular type.
type TypeContext struct {
	network string
	id      string
	p       Provider
}

// valid returns an error if the type context is invalid.
func (tc TypeContext) valid() error {
	if tc.id == "" {
		// TODO: Return error
	}
	return nil
}

// Read returns a copy of the type configuration.
func (tc TypeContext) Read(ctx context.Context) (*Type, error) {
	if err := tc.valid(); err != nil {
		return nil, err
	}
	var (
		typ Type
		err error
	)
	if tc.network != "" {
		typ, err = tc.p.NetworkType(ctx, tc.network, tc.id)
	} else {
		typ, err = tc.p.Type(ctx, tc.id)
	}
	if err != nil {
		return nil, err
	}
	return &typ, nil
}

// DeviceContext provides access to the attributes of a particular device.
type DeviceContext struct {
	network string
	id      string
	p       Provider
}

// valid returns an error if the device context is invalid.
func (dc DeviceContext) valid() error {
	if dc.id == "" {
		// TODO: Return error
	}
	return nil
}

// Read returns a copy of the device configuration.
func (dc DeviceContext) Read(ctx context.Context) (*Device, error) {
	if err := dc.valid(); err != nil {
		return nil, err
	}
	var (
		device Device
		err    error
	)
	if dc.network != "" {
		device, err = dc.p.NetworkDevice(ctx, dc.network, dc.id)
	} else {
		device, err = dc.p.Device(ctx, dc.id)
	}
	if err != nil {
		return nil, err
	}
	return &device, nil
}

/*
// Observe returns a copy of the type configuration.
func (tc *TypeContext) Observe(ctx context.Context) TypeObserver {
}
*/

// MACContext provides access to the attributes of a particular hardware
// address.
type MACContext struct {
	network string
	addr    net.HardwareAddr
	p       Provider
}

// valid returns an error if the MAC context is invalid.
func (mc MACContext) valid() error {
	if mc.addr == nil { // FIXME: Call a real hardware address validation function
		// TODO: Return error
	}
	return nil
}

// Read returns a copy of the MAC configuration.
func (mc MACContext) Read() (MAC, error) {
	return MAC{}, errors.New("Not implemented yet.")
}

// LeaseContext provides access to lease management functions for a particular
// IP address on a given network.
type LeaseContext struct {
	network string
	ip      net.IP
	p       Provider
}

// valid returns an error if the least context is invalid.
func (lc LeaseContext) valid() error {
	if lc.network == "" {
		// TODO: Return error
	}
	if lc.ip == nil { // FIXME: Call a real IP validation function
		// TODO: Return error
	}
	return nil
}

// Read returns a copy of the lease, if it exists.
func (lc LeaseContext) Read(ctx context.Context) (*Lease, error) {
	if err := lc.valid(); err != nil {
		return nil, err
	}
	return lc.p.NetworkLease(ctx, lc.network, lc.ip)
}

// Create will create a new lease.
func (lc LeaseContext) Create(ctx context.Context, addr net.HardwareAddr, expiration time.Time) (bool, error) {
	if err := lc.valid(); err != nil {
		return false, err
	}
	if addr == nil { // FIXME: Call a real hardware address validation function
		// TODO: Return error
	}
	return lc.p.NetworkLeaseCreate(ctx, lc.network, lc.ip, addr, expiration)
}

// Renew will renew a lease.
func (lc LeaseContext) Renew(ctx context.Context, addr net.HardwareAddr, expiration time.Time) (bool, error) {
	if err := lc.valid(); err != nil {
		return false, err
	}
	if addr == nil { // FIXME: Call a real hardware address validation function
		// TODO: Return error
	}
	return lc.p.NetworkLeaseRenew(ctx, lc.network, lc.ip, addr, expiration)
}

// Release will release a lease.
func (lc LeaseContext) Release(ctx context.Context, addr net.HardwareAddr) (bool, error) {
	if err := lc.valid(); err != nil {
		return false, err
	}
	if addr == nil { // FIXME: Call a real hardware address validation function
		// TODO: Return error
	}
	return lc.p.NetworkLeaseRelease(ctx, lc.network, lc.ip, addr)
}

// Hold will attempt to place a hold on the lease IP address without assigning
// it to a particular MAC address. This is typically used by the DHCP service to
// reserve addresses for offline use.
func (lc LeaseContext) Hold(ctx context.Context) (bool, error) {
	if err := lc.valid(); err != nil {
		return false, err
	}
	return lc.p.NetworkLeaseHold(ctx, lc.network, lc.ip)
}
