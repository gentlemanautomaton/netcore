package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

// Config is the host+zone config for this server
type Config struct {
	sync.Mutex
	etcdClient        *etcd.Client
	hostname          string
	zone              string
	domain            string
	subnet            *net.IPNet
	gateway           net.IP
	dhcpIP            net.IP
	dhcpNIC           string
	dhcpSubnet        *net.IPNet
	dhcpLeaseDuration time.Duration
	dhcpTFTP          string
	dnsForwarders     []string
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

func getConfig(etc *etcd.Client) (*Config, error) {
	fmt.Println("Getting CONFIG")

	fmt.Println("precreate")
	etc.CreateDir("config", 0)
	fmt.Println("postcreate")

	cfg := new(Config)

	// Hostname
	{
		hostname, err := getHostname()
		if err != nil {
			return nil, err
		}
		cfg.hostname = hostname
	}

	// Zone
	{
		var response *etcd.Response
		var err error
		if setZone != nil && *setZone != "" {
			response, err = etc.Set("config/"+cfg.hostname+"/zone", *setZone, 0)
		} else {
			response, err = etc.Get("config/"+cfg.hostname+"/zone", false, false)
		}
		if err != nil {
			return nil, err
		}
		if response == nil || response.Node == nil || response.Node.Value == "" {
			return nil, ErrNoZone
		}
		cfg.zone = response.Node.Value
	}

	// Domain
	{
		response, err := etc.Get("config/"+cfg.zone+"/domain", false, false)
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			cfg.domain = response.Node.Value
		}
	}

	// Subnet
	{
		response, err := etc.Get("config/"+cfg.zone+"/subnet", false, false)
		if err != nil {
			return nil, err
		}
		if response == nil || response.Node == nil || response.Node.Value == "" {
			return nil, ErrNoZone
		}
		_, subnet, err := net.ParseCIDR(response.Node.Value)
		if err != nil {
			return nil, err
		}
		cfg.subnet = subnet
	}

	// Gateway
	{
		response, err := etc.Get("config/"+cfg.zone+"/gateway", false, false)
		if err != nil {
			return nil, err
		}
		if response == nil || response.Node == nil || response.Node.Value == "" {
			return nil, ErrNoGateway
		}
		gateway := net.ParseIP(response.Node.Value).To4()
		cfg.gateway = gateway
	}

	// DHCPIP
	{
		var response *etcd.Response
		var err error
		if setDHCPIP != nil && *setDHCPIP != "" {
			response, err = etc.Set("config/"+cfg.hostname+"/dhcpip", *setDHCPIP, 0)
		} else {
			response, err = etc.Get("config/"+cfg.hostname+"/dhcpip", false, false)
		}
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			dhcpIP := net.ParseIP(response.Node.Value).To4()
			cfg.dhcpIP = dhcpIP
		}
	}

	// DHCPNIC
	{
		var response *etcd.Response
		var err error
		if setDHCPNIC != nil && *setDHCPNIC != "" {
			response, err = etc.Set("config/"+cfg.hostname+"/dhcpnic", *setDHCPNIC, 0)
		} else {
			response, err = etc.Get("config/"+cfg.hostname+"/dhcpnic", false, false)
		}
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			cfg.dhcpNIC = response.Node.Value
		}
	}

	// DHCPSubnet
	{
		var response *etcd.Response
		var err error
		if setZone != nil && *setZone != "" && setDHCPSubnet != nil && *setDHCPSubnet != "" {
			response, err = etc.Set("config/"+cfg.zone+"/dhcpsubnet", *setDHCPSubnet, 0)
		} else {
			response, err = etc.Get("config/"+cfg.zone+"/dhcpsubnet", false, false)
		}
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			_, dhcpSubnet, err := net.ParseCIDR(response.Node.Value)
			if err != nil {
				return nil, err
			}
			cfg.dhcpSubnet = dhcpSubnet
		}
	}

	// DHCPLeaseDuration
	{
		cfg.dhcpLeaseDuration = 12 * time.Hour // default setting is 12 hours
		var response *etcd.Response
		var err error
		if setZone != nil && *setZone != "" && setDHCPLeaseDuration != nil && *setDHCPLeaseDuration != "" {
			response, err = etc.Set("config/"+cfg.zone+"/dhcpleaseduration", *setDHCPLeaseDuration, 0)
		} else {
			response, err = etc.Get("config/"+cfg.zone+"/dhcpleaseduration", false, false)
		}
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			value, err := strconv.Atoi(response.Node.Value)
			if err != nil {
				return nil, err
			}
			dhcpLeaseDuration := time.Duration(value) * time.Minute
			if err != nil {
				return nil, err
			}
			cfg.dhcpLeaseDuration = dhcpLeaseDuration
		}
	}

	// DHCPTFTP
	{
		var response *etcd.Response
		var err error
		if setDHCPTFTP != nil && *setDHCPTFTP != "" {
			response, err = etc.Set("config/"+cfg.hostname+"/dhcptftp", *setDHCPTFTP, 0)
		} else {
			response, err = etc.Get("config/"+cfg.hostname+"/dhcptftp", false, false)
		}
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			cfg.dhcpTFTP = response.Node.Value
		}
	}

	// DNSForwarders
	{
		cfg.dnsForwarders = []string{"8.8.8.8:53", "8.8.4.4:53"} // default uses Google's Public DNS servers
		response, err := etc.Get("config/"+cfg.zone+"/dnsforwarders", false, false)
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			cfg.dnsForwarders = strings.Split(",", response.Node.Value)
		}
	}

	fmt.Printf("CONFIG: [%+v]\n", cfg)

	return cfg, nil
}

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
