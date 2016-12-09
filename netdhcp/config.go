package netdhcp

/*
var defaultConfig = Cfg{
	Enabled: true,
	NIC:     "eth0", // FIXME: Is it safe to have a default at all for this?
	IP:      nil,
	Network: "",
	Subnet:  nil,
	Attr: Attr{
		Gateway:       nil,
		Domain:        "",
		TFTP:          "",
		NTP:           nil,
		LeaseDuration: time.Minute * 60, // TODO: Look for guidance on what this should be
	},
	GuestPool: nil, // TODO: Consider specifying a default pool
	//Options:       nil,              // FIXME: Make sure a nil map doesn't screw anything up
}

const minimumLeaseDuration = 60 * time.Second // FIXME: put this in a config

// Config provides all of the necessary configuration context for the operation
// of a netcore DNS instance.
type ImmConfig interface {
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
	return &config{defaultConfig.Copy()}
}
*/

/*
// Config represents server configuration data.
type Config struct {
	//Enabled   bool
	NIC       string     // The network adaptor the server will listen on
	IP        net.IP     // The IP address of the network adaptor
	Subnet    *net.IPNet // The subnet the server will listen to
	GuestPool *net.IPNet
}
*/

/*

// ConfigOverlay composes a series of Config structures into a single
// overlay
type ConfigOverlay struct {
	source []*Config
}

// NewCfg creates a mutable Cfg instance from the given Config interface.
func NewCfg(c Config) Cfg {
	return Cfg{
		Instance:  c.Instance(),
		Enabled:   c.Enabled(),
		NIC:       c.NIC(),
		IP:        c.IP(),
		Network:   c.Network(),
		Subnet:    c.Subnet(),
		GuestPool: c.GuestPool(),
		Attr: Attr{
			Gateway:       c.Gateway(),
			Domain:        c.Domain(),
			TFTP:          c.TFTP(),
			NTP:           c.NTP(),
			LeaseDuration: c.LeaseDuration(),
		},
		//Options:       c.Options(),
	}
}

func (c *Config) Overlay(source *Config) {

}

// Copy will make a deep copy of the Cfg.
func (c *Config) Copy() Cfg {
	return Cfg{
		Instance:  c.Instance,
		Enabled:   c.Enabled,
		NIC:       c.NIC,
		IP:        c.IP,
		Network:   c.Network,
		Subnet:    c.Subnet,
		GuestPool: c.GuestPool,
		Attr: Attr{
			Gateway:       c.Gateway,
			Domain:        c.Domain,
			TFTP:          c.TFTP,
			NTP:           c.NTP,
			LeaseDuration: c.LeaseDuration,
		},
		//Options:       copyOptions(c.Options),
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

// MergeConfig will collapse the provided configurations into a single instance.
func Merge(c ...Config) Config {
	// TODO: Consider an alternative implementation that maintains references to
	//       all of the members and queries them in series for each method call.
	if len(c) == 0 {
		return nil
	}
	cfg := NewCfg(c[1])
	for _, s := range c[1:] {
		if enabled := s.Enabled(); enabled {
			cfg.Enabled = enabled
		}
		if nic := s.NIC(); nic != "" {
			cfg.NIC = nic
		}
		if ip := s.IP(); ip != nil {
			cfg.IP = ip
		}
		// FIXME: Finish writing this
	}
	return &config{cfg}
}

// config provides an immutable implementation of the Config interface.
type config struct {
	x Cfg
}

func (c *config) Instance() string {
	return c.x.Instance
}

func (c *config) Enabled() bool {
	return c.x.Enabled
}

func (c *config) NIC() string {
	return c.x.NIC
}

func (c *config) IP() net.IP {
	return c.x.IP
}

func (c *config) Network() string {
	return c.x.Network
}

func (c *config) Subnet() *net.IPNet {
	return c.x.Subnet
}

func (c *config) Gateway() net.IP {
	return c.x.IP
}

func (c *config) Domain() string {
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
	return copyOptions(c.x.Options)
}

func copyOptions(source dhcp4.Options) dhcp4.Options {
	var options dhcp4.Options
	if source != nil {
		options = make(dhcp4.Options, len(source))
		for k, v := range source {
			options[k] = v
		}
	}
	return options
}
*/

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
