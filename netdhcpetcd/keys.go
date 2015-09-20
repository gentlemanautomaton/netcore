package netdhcpetcd

import "net"

const (
	RootBucket     = "/netcore/dhcp"
	ConfigBucket   = RootBucket + "/config"
	ServerBucket   = RootBucket + "/server"
	NetworkBucket  = RootBucket + "/network"
	HostBucket     = RootBucket + "/host"
	HardwareBucket = RootBucket + "/mac"
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
