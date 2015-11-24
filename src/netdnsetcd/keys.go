package netdnsetcd

import (
	"net"
	"strings"
)

const (
	RootBucket     = "/netcore/dns"
	ConfigBucket   = RootBucket + "/config"
	ServerBucket   = RootBucket + "/server"
	ResourceBucket = RootBucket + "/resource"
	ArpaBucket     = ResourceBucket + "/arpa/in-addr/"
)

const (
	EnabledField        = "enabled"
	DefaultTTLField     = "defaultttl"
	MinimumTTLField     = "minimumttl"
	CacheRetentionField = "cacheretention"
	ForwardersField     = "forwarders"
	TTLField            = "ttl"
)

// ServerKey returns the etcd key of the given server instance
func ServerKey(instance string) string {
	return ServerBucket + "/" + instance
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
