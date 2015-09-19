package netdhcpetcd

import (
	"dustywilson/netcore/netdhcp"
	"net"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

// IP returns an IPEntry for the given IP address if it exists, otherwise it
// returns netdhcp.ErrNotFound
func (p *Provider) IP(ip net.IP) (netdhcp.IPEntry, error) {
	key := etcdKeyFromIP(ip)
	response, err := p.client.Get(key, false, false)
	if response == nil || response.Node == nil {
		return IPEntry{}, netdhcp.ErrNotFound
	}
	mac, err := net.ParseMAC(response.Node.Value)
	if err != nil {
		return IPEntry{}, err
	}
	return IPEntry{MAC: mac}, nil
}

// HasIP returns true if the IP address has been allocated.
func (p *Provider) HasIP(ip net.IP) bool {
	key := etcdKeyFromIP(ip)
	response, _ := p.client.Get(key, false, false)
	if response != nil && response.Node != nil {
		return true
	}
	return false
}

// MAC returns a MACEntry for the given hardware address if it exists, otherwise
// it returns netdhcp.ErrNotFound
func (p *Provider) MAC(mac net.HardwareAddr, cascade bool) (*netdhcp.MACEntry, bool, error) {
	// TODO: First attempt to retrieve the entry from a cache of some kind (that can be dirtied)
	// NOTE: The cache should always return a deep copy of the cached value
	entry := MACEntry{MAC: mac}

	// Copy cascaded attributes by making recursive calls to this function
	if cascade && len(mac) > 1 {
		parent, _, _ := p.MAC(mac[0:len(mac)-1], cascade) // Chop off the last byte for each recursive call
		if parent != nil {
			entry.Attr = parent.Attr // Only safe if we receive a deep copy of the cached value
		}
	}

	// Fetch attributes and lease data for this MAC
	key := etcdKeyFromMAC(mac)
	response, err := p.client.Get(key, true, true) // do the lookup
	if err != nil {
		// FIXME: Return the etcd error for everything except missing keys
		//return nil, false, err
		return &entry, false, nil
	}

	if response.Node == nil || !response.Node.Dir {
		// Not found
		// NOTE: Retuning the entry is necessary for recursive calls
		return &entry, false, nil
	}

	etcdNodeToMACEntry(response.Node, &entry)

	return &entry, true, nil
}

// RenewLease will attempt to update the duration of the given lease.
func (p *Provider) RenewLease(lease *netdhcp.MACEntry) error {
	// FIXME: Validate lease
	duration := uint64(lease.Duration.Seconds() + 0.5) // Half second jitter to hide network delay
	_, err := p.client.CompareAndSwap("dhcp/"+lease.IP.String(), lease.MAC.String(), duration, lease.MAC.String(), 0)
	if err == nil {
		return db.WriteLease(lease)
	}
	return err
}

// CreateLease will attempt to create a new lease.
func (p *Provider) CreateLease(lease *netdhcp.MACEntry) error {
	// FIXME: Validate lease
	duration := uint64(lease.Duration.Seconds() + 0.5)
	_, err := p.client.Create("dhcp/"+lease.IP.String(), lease.MAC.String(), duration)
	if err == nil {
		return p.WriteLease(lease)
	}
	return err
}

// WriteLease will attempt to write the lease data.
func (p *Provider) WriteLease(lease *netdhcp.MACEntry) error {
	// FIXME: Validate lease
	// NOTE: This does not save attributes. That should probably happen in a different function.
	duration := uint64(lease.Duration.Seconds() + 0.5) // Half second jitter to hide network delay
	// FIXME: Decide what to do if either of these calls returns an error
	p.client.CreateDir("dhcp/"+lease.MAC.String(), 0)
	p.client.Set("dhcp/"+lease.MAC.String()+"/ip", lease.IP.String(), duration)
	return nil
}

// TODO: Write function for saving attributes to etcd?

func etcdNodeToMACEntry(root *etcd.Node, entry *netdhcp.MACEntry) {
	for _, node := range root.Nodes {
		if node.Dir {
			continue // Ignore subdirectories
		}
		key := strings.Replace(node.Key, root.Key+"/", "", 1)
		switch key {
		case "ip":
			entry.IP = net.ParseIP(node.Value)
			entry.Duration = time.Duration(node.TTL)
		default:
			if entry.Attr == nil {
				entry.Attr = make(map[string]string)
			}
			entry.Attr[key] = node.Value
		}
	}
}

func etcdKeyFromIP(ip net.IP) string {
	return "/dhcp/" + ip.String()
}

func etcdKeyFromMAC(mac net.HardwareAddr) string {
	return "/dhcp/" + mac.String()
}
