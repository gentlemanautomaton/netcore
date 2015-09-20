package netdhcpetcd

import (
	"net"
	"strings"
)

const (
	RootBucket     = "/netcore/dhcp"
	ConfigBucket   = RootBucket + "/config"
	ServerBucket   = RootBucket + "/server"
	NetworkBucket  = RootBucket + "/network"
	HostBucket     = RootBucket + "/host"
	HardwareBucket = RootBucket + "/mac"
	ResourceBucket = RootBucket + "/resource"
	ArpaBucket     = ResourceBucket + "/arpa/in-addr/"
)

const (
	NICField           = "nic"
	IPField            = "ip"
	EnabledField       = "enabled"
	NetworkField       = "network"
	SubnetField        = "subnet"
	GatewayField       = "gw"
	DomainField        = "domain"
	TFTPField          = "tftp"
	NTPField           = "ntp"
	LeaseDurationField = "leaseduration"
	GuestPoolField     = "pool"
	TTLField           = "ttl"
)

// ServerKey returns the etcd key of the given server instance
func ServerKey(instance string) string {
	return ServerBucket + "/" + instance
}

// NetworkKey returns the etcd key of the given network id
func NetworkKey(network string) string {
	return NetworkBucket + "/" + network
}

// HostKey returns the etcd key of the given host ID
func HostKey(host string) string {
	return HostBucket + "/" + host
}

// HardwareKey returns the etcd key of the given mac address
func HardwareKey(mac net.HardwareAddr) string {
	return HardwareBucket + "/" + mac.String()
}

// IPKey returns the etcd key of the given ip address reservation in the
// the provided network.
func IPKey(network string, ip net.IP) string {
	return NetworkBucket + "/" + network + "/" + IPField + "/" + ip.String()
}

// ResourceKey returns the etcd key of DNS resource record entry for the given
// FQDN.
func ResourceKey(fqdn string) string {
	parts := strings.Split(cleanFQDN(fqdn), ".")   // breakup the queryed name
	path := strings.Join(reverseSlice(parts), "/") // reverse and join them with a slash delimiter
	return ResourceBucket + "/" + path
}

// ResourceTypeKey returns the etcd key of DNS resource record entry for the given
// FQDN and resource record type.
func ResourceTypeKey(fqdn string, rrType string) string {
	rrType = strings.ToLower(rrType)
	return ResourceKey(fqdn) + "/@" + rrType
}

// ArpaKey returns the etcd key of DNS arpa entry for the given IP address
func ArpaKey(ip net.IP) string {
	// FIXME: Support IPv6 addresses
	slashedIP := strings.Replace(ip.To4().String(), ".", "/", -1)
	return ArpaBucket + slashedIP
}
