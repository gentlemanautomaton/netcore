package main

import (
	"errors"
	"flag"
	"net"
	"sync"
	"time"
)

// Config is the host+zone config for this server
type Config struct {
	sync.Mutex
	db                 DB
	hostname           string
	zone               string
	domain             string
	subnet             *net.IPNet
	gateway            net.IP
	dhcpIP             net.IP
	dhcpNIC            string
	dhcpSubnet         *net.IPNet
	dhcpLeaseDuration  time.Duration
	dhcpTFTP           string
	dnsForwarders      []string
	dnsCacheMaxTTL     time.Duration
	dnsCacheMissingTTL time.Duration
}

type ConfigProvider interface {
	//Get(key string) string
	GetConfig() (*Config, error)
}

var setZone = flag.String("setZone", "", "Overwrite (permanently) the zone that this machine is in.")
var setDHCPIP = flag.String("setDHCPIP", "", "Overwrite (permanently) the DHCP hosting IP for this machine (or set it to empty to disable DHCP).")
var setDHCPNIC = flag.String("setDHCPNIC", "", "Overwrite (permanently) the DHCP hosting NIC name for this machine (or set it to empty to disable DHCP).")
var setDHCPSubnet = flag.String("setDHCPSubnet", "", "Overwrite (permanently) the DHCP subnet for this zone (requires setZone flag or it'll no-op).")
var setDHCPLeaseDuration = flag.String("setDHCPLeaseDuration", "", "Overwrite (permanently) the default DHCP lease duration for this zone (requires setZone flag or it'll no-op).")
var setDHCPTFTP = flag.String("setDHCPTFTP", "", "Overwrite (permanently) the DHCP TFTP Server Name for this machine (or set it to empty to disable DHCP).")

// ErrNoZone is an error returned during config init to indicate that the host has not been assigned to a zone in etcd keyed off of its hostname
var ErrNoZone = errors.New("This host has not been assigned to a zone.")

// ErrNoDHCPIP is an error returned during config init to indicate that the host has not been assigned to a zone in etcd keyed off of its hostname
var ErrNoDHCPIP = errors.New("This host has not been assigned a DHCP IP.")

// ErrNoGateway is an error returned during config init to indicate that the zone has not been assigned a gateway in etcd keyed off of the zone name
var ErrNoGateway = errors.New("This zone does not have an assigned gateway.")

// Hostname returns this machine's hostname
func (cfg *Config) Hostname() string {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.hostname
}

// Zone returns the zone name
func (cfg *Config) Zone() string {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.zone
}

// Domain returns the default domain for this zone
func (cfg *Config) Domain() string {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.domain
}

// Subnet returns the subnet for this zone
func (cfg *Config) Subnet() *net.IPNet {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.subnet
}

// Gateway returns the IP address for the network gateway within the subnet
func (cfg *Config) Gateway() net.IP {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.gateway
}

// DHCPIP returns the IP address for the DHCP process host
func (cfg *Config) DHCPIP() net.IP {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.dhcpIP
}

// DHCPNIC returns the NIC device name for the DHCP process host
func (cfg *Config) DHCPNIC() string {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.dhcpNIC
}

// DHCPSubnet returns the DHCP pool subnet for this zone
func (cfg *Config) DHCPSubnet() *net.IPNet {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.dhcpSubnet
}

// DHCPLeaseDuration returns the default DHCP lease duration for this zone
func (cfg *Config) DHCPLeaseDuration() time.Duration {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.dhcpLeaseDuration
}

// DHCPTFTP returns the TFTP Server Name for this zone
func (cfg *Config) DHCPTFTP() string {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.dhcpTFTP
}

// DNSForwarders returns the list of DNS resolvers we use for recursive lookups
func (cfg *Config) DNSForwarders() []string {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.dnsForwarders
}

// DNSCacheMaxTTL returns the maximum duration for which answers will be stored
// in the cache
func (cfg *Config) DNSCacheMaxTTL() time.Duration {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.dnsCacheMaxTTL
}

// DNSCacheMissingTTL returns the TTL for cached queries with no answer
func (cfg *Config) DNSCacheMissingTTL() time.Duration {
	cfg.Lock()
	defer cfg.Unlock()
	return cfg.dnsCacheMissingTTL
}
