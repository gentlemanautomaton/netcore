package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

func (db EtcdDB) GetConfig() (*Config, error) {
	fmt.Println("Getting CONFIG")

	etc := db.client

	fmt.Println("precreate")
	etc.CreateDir("config", 0)
	fmt.Println("postcreate")

	cfg := new(Config)

	// Hostname
	{
		var hostname string
		if len(os.Getenv("ETCD_NAME")) > 0 {
			re := regexp.MustCompile(`^/([^/]+)/`)
			hostnameParts := re.FindStringSubmatch(os.Getenv("ETCD_NAME"))
			if len(hostnameParts) > 1 && len(hostnameParts[1]) > 0 {
				hostname = hostnameParts[1]
			}
		} else {
			var err error
			hostname, err = getHostname()
			if err != nil {
				return nil, err
			}
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
			cfg.dhcpLeaseDuration = time.Duration(value) * time.Minute
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

	// dnsCacheMaxTTL
	{
		cfg.dnsCacheMaxTTL = 0 // default to no caching
		response, err := etc.Get("config/"+cfg.zone+"/dnscachemaxttl", false, false)
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			value, err := strconv.Atoi(response.Node.Value)
			if err != nil {
				return nil, err
			}
			cfg.dnsCacheMaxTTL = time.Duration(value) * time.Second
		}
	}

	// dnsCacheMissingTTL
	{
		cfg.dnsCacheMissingTTL = 30 * time.Second // default setting is 30 seconds
		response, err := etc.Get("config/"+cfg.zone+"/dnscachemissingttl", false, false)
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			value, err := strconv.Atoi(response.Node.Value)
			if err != nil {
				return nil, err
			}
			cfg.dnsCacheMissingTTL = time.Duration(value) * time.Second
		}
	}

	fmt.Printf("CONFIG: [%+v]\n", cfg)

	return cfg, nil
}
