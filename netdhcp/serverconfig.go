package netdhcp

import "net"

// ServerConfig is a source of server configuration data.
//
// TODO: Add DHCP relay configuration.
type ServerConfig interface {
	ServerEnabled() bool
	ServerNIC() string
	ServerIP() net.IP
	ServerSubnet() *net.IPNet // TODO: Make this a PoolSet?
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
