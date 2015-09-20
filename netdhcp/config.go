package netdhcp

import (
	"net"
	"time"

	"github.com/krolaw/dhcp4"
)

const minimumLeaseDuration = 60 * time.Second // FIXME: put this in a config

// Config provides all of the necessary configuration context for the operation
// of a netcore DNS instance.
type Config interface {
	Instance() string
	Enabled() bool
	NIC() string
	IP() net.IP
	Network() string
	Subnet() *net.IPNet
	Gateway() net.IP
	Domain() string
	TFTP() string
	NTP() net.IP
	LeaseDuration() time.Duration
	GuestPool() *net.IPNet
	Options() dhcp4.Options
}

// NewConfig creates an immutable instance of the Config interface from the
// given mutable data.
func NewConfig(c *Cfg) Config {
	return &config{c.Copy()}
}

// DefaultConfig returns a Config interface with the default values for netcore.
func DefaultConfig() Config {
	return &config{Cfg{
	// TODO: Create these!
	}}
}

// Cfg provides a mutable implementation of the Config interface. It can be made
// into an immutable Config instance via the NewConfig function.
type Cfg struct {
	Instance      string
	Enabled       bool
	NIC           string
	IP            net.IP
	Network       string
	Subnet        *net.IPNet
	Gateway       net.IP
	Domain        string
	TFTP          string
	NTP           net.IP
	LeaseDuration time.Duration
	GuestPool     *net.IPNet
	Options       dhcp4.Options
}

// NewCfg creates a mutable Cfg instance from the given Config interface.
func NewCfg(c Config) Cfg {
	return Cfg{
		Instance:      c.Instance(),
		Enabled:       c.Enabled(),
		NIC:           c.NIC(),
		IP:            c.IP(),
		Network:       c.Network(),
		Subnet:        c.Subnet(),
		Gateway:       c.Gateway(),
		Domain:        c.Domain(),
		TFTP:          c.TFTP(),
		NTP:           c.NTP(),
		LeaseDuration: c.LeaseDuration(),
		GuestPool:     c.GuestPool(),
		Options:       c.Options(),
	}
}

// Copy will make a deep copy of the Cfg.
func (c *Cfg) Copy() Cfg {
	return Cfg{
		Instance:      c.Instance,
		Enabled:       c.Enabled,
		NIC:           c.NIC,
		IP:            c.IP,
		Network:       c.Network,
		Subnet:        c.Subnet,
		Gateway:       c.Gateway,
		Domain:        c.Domain,
		TFTP:          c.TFTP,
		NTP:           c.NTP,
		LeaseDuration: c.LeaseDuration,
		GuestPool:     c.GuestPool,
		Options:       c.Options,
	}
}

// Validate returns an error if the config is invalid, otherwise it returns nil.
func Validate(c Config) error {
	if c == nil {
		return ErrNoConfig
	}
	if c.IP() == nil {
		return ErrNoConfigIP
	}
	if c.Subnet() == nil {
		return ErrNoConfigSubnet
	}
	if c.NIC() == "" {
		return ErrNoConfigNIC
	}
	return nil
}

// config provides an immutable implementation of the Config interface.
type config struct {
	x Cfg
}

func (c config) Instance() string {
	return c.x.Instance
}

func (c config) Enabled() bool {
	return c.x.Enabled
}

func (c config) NIC() string {
	return c.x.NIC
}

func (c config) IP() net.IP {
	return c.x.IP
}

func (c config) Network() string {
	return c.x.Network
}

func (c config) Subnet() *net.IPNet {
	return c.x.Subnet
}

func (c config) Gateway() net.IP {
	return c.x.IP
}

func (c config) Domain() string {
	return c.x.Domain
}

func (c config) TFTP() string {
	return c.x.TFTP
}

func (c config) NTP() net.IP {
	return c.x.NTP
}

func (c config) LeaseDuration() time.Duration {
	return c.x.LeaseDuration
}

func (c config) GuestPool() *net.IPNet {
	return c.x.GuestPool
}

func (c config) Options() dhcp4.Options {
	return c.x.Options
}

// Reference:
/*

// ServiceOld is the DHCP server instance
type ServiceOld struct {
	ip             net.IP
	domain         string
	subnet         *net.IPNet
	guestPool      *net.IPNet
	leaseDuration  time.Duration
	defaultOptions dhcp4.Options // FIXME: make different options per pool?
}

d := &DHCPService{
	ip:            cfg.DHCPIP(),
	leaseDuration: cfg.DHCPLeaseDuration(),
	db:            cfg.db,
	subnet:        cfg.Subnet(),
	guestPool:     cfg.DHCPSubnet(),
	domain:        cfg.Domain(),
	defaultOptions: dhcp4.Options{
		dhcp4.OptionSubnetMask:       net.IP(cfg.Subnet().Mask),
		dhcp4.OptionRouter:           cfg.Gateway(),
		dhcp4.OptionDomainNameServer: cfg.DHCPIP(),
	},
}
dhcpTFTP := cfg.DHCPTFTP()
if dhcpTFTP != "" {
	d.defaultOptions[dhcp4.OptionTFTPServerName] = []byte(dhcpTFTP)
}
*/

// Experimental:
/*
// Config provides all of the necessary configuration context for the operation
// of a netcore DHCP instance.
type Config struct {
	Global  GlobalConfig
	Network NetworkConfig
	Server  ServerConfig
}

// GlobalConfig contains global DHCP configuration.
type GlobalConfig struct {
	Netname       string
	NIC           string
	LeaseDuration time.Duration
}

// NetworkConfig contains configuration for a DHCP network.
type NetworkConfig struct {
	Netname       string
	Subnet        *net.IPNet
	Gateway       net.IP
	Domain        string
	TFTP          string
	LeaseDuration time.Duration
	Pools         []PoolConfig
}

// PoolConfig contains configuration for a DHCP pool.
type PoolConfig struct {
	Subnet *net.IPNet
}

// ServerConfig contains configuration for a DHCP server.
type ServerConfig struct {
	Hostname string
	Netname  string
	NIC      string
	IP       net.IP
	Enabled  bool
}

// Hostname returns the hostname of the instance for which this configuration
// data is intended.
func (c *Config) Hostname() string {
	return c.Server.Hostname
}

// Netname returns the name of the network that the config applies to.
func (c *Config) Netname() string {
	switch {
	case c.Server.Netname != "":
		return c.Server.Netname
	case c.Global.Netname != "":
		return c.Global.Netname
	default:
		return ""
	}
}
*/
