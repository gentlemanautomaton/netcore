package schema

import "net"

// InstanceKey returns the etcd key of the given server instance.
func InstanceKey(instance string) string {
	return Instance + "/" + clean(instance)
}

// NetworkKey returns the etcd key of the given network id.
func NetworkKey(network string) string {
	return Network + "/" + clean(network)
}

// DeviceKey returns the etcd key of the given device id.
func DeviceKey(device string) string {
	return Device + "/" + clean(device)
}

// TypeKey returns the etcd key of the given type id.
func TypeKey(typ string) string {
	return Type + "/" + clean(typ)
}

// HardwareKey returns the etcd key of the given hardware address.
func HardwareKey(haddr net.HardwareAddr) string {
	return Hardware + "/" + clean(haddr.String())
}

// HostKey returns the etcd key of the given host ID
/*
func HostKey(host string) string {
	return HostBucket + "/" + host
}
*/

// NetworkLeaseKey returns the etcd key for the lease of a specific IP address
// in the the provided network.
func NetworkLeaseKey(network string, ip net.IP) string {
	return Network + "/" + clean(network) + "/" + LeaseField + "/" + clean(ip.String())
}

// NetworkDeviceKey returns the etcd key for the given device on the
// specified network.
func NetworkDeviceKey(network string, device string) string {
	return Network + "/" + clean(network) + "/" + DeviceTag + "/" + clean(device)
}

// NetworkHardwareKey returns the etcd key for the given hardware address on the
// specified network.
func NetworkHardwareKey(network string, haddr net.HardwareAddr) string {
	return Network + "/" + clean(network) + "/" + HardwareTag + "/" + clean(haddr.String())
}

// ResourceKey returns the etcd key of DNS resource record entry for the given
// FQDN.
/*
func ResourceKey(fqdn string) string {
	parts := strings.Split(cleanFQDN(fqdn), ".")   // breakup the queryed name
	path := strings.Join(reverseSlice(parts), "/") // reverse and join them with a slash delimiter
	return ResourceBucket + "/" + path
}
*/

// ResourceTypeKey returns the etcd key of DNS resource record entry for the given
// FQDN and resource record type.
/*
func ResourceTypeKey(fqdn string, rrType string) string {
	rrType = strings.ToLower(rrType)
	return ResourceKey(fqdn) + "/@" + rrType
}
*/

// ArpaKey returns the etcd key of DNS arpa entry for the given IP address
/*
func ArpaKey(ip net.IP) string {
	// FIXME: Support IPv6 addresses
	slashedIP := strings.Replace(ip.To4().String(), ".", "/", -1)
	return ArpaBucket + slashedIP
}
*/
