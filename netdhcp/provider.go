package netdhcp

import (
	"net"
	"time"
)

// Provider implements all storage interfaces necessary for operation of
// the DHCP service.
type Provider interface {
	ConfigProvider
	DataProvider
}

// ConfigProvider provides DHCP configuration to the DHCP service.
type ConfigProvider interface {
	Init() error
	Config(instance string) (Config, error)
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
