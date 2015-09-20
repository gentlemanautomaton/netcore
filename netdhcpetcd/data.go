package netdhcpetcd

import (
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/dustywilson/netcore/netdhcp"
	"golang.org/x/net/context"
)

// IP returns an IPEntry for the given IP address if it exists, otherwise it
// returns netdhcp.ErrNotFound
func (p *Provider) IP(ip net.IP) (netdhcp.IPEntry, error) {
	key := etcdKeyFromIP(ip)
	keys := client.NewKeysAPI(p.c)
	response, err := keys.Get(context.Background(), key, nil)
	if err != nil {
		if etcdKeyNotFound(err) {
			err = netdhcp.ErrNotFound
		}
		return netdhcp.IPEntry{}, err
	}
	if response == nil || response.Node == nil {
		return netdhcp.IPEntry{}, netdhcp.ErrNotFound
	}
	mac, err := net.ParseMAC(response.Node.Value)
	if err != nil {
		return netdhcp.IPEntry{}, err
	}
	return netdhcp.IPEntry{MAC: mac}, nil
}

// HasIP returns true if the IP address has been allocated.
func (p *Provider) HasIP(ip net.IP) bool {
	key := etcdKeyFromIP(ip)
	keys := client.NewKeysAPI(p.c)
	response, _ := keys.Get(context.Background(), key, nil)
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
	entry := netdhcp.MACEntry{MAC: mac}

	// Copy cascaded attributes by making recursive calls to this function
	if cascade && len(mac) > 1 {
		parent, _, _ := p.MAC(mac[0:len(mac)-1], cascade) // Chop off the last byte for each recursive call
		if parent != nil {
			entry.Attr = parent.Attr // Only safe if we receive a deep copy of the cached value
		}
	}

	// Fetch attributes and lease data for this MAC
	key := etcdKeyFromMAC(mac)
	keys := client.NewKeysAPI(p.c)
	options := &client.GetOptions{Recursive: true, Sort: true}
	response, err := keys.Get(context.Background(), key, options) // do the lookup
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
	duration := lease.Duration + (time.Second / 2) // Half second jitter to hide network delay
	keys := client.NewKeysAPI(p.c)
	// FIXME: Verify that this conversion to the new etcd client API is right
	_, err := keys.Set(context.Background(), "dhcp/"+lease.IP.String(), lease.MAC.String(), &client.SetOptions{
		PrevValue: lease.MAC.String(),
		TTL:       duration,
	})
	if err == nil {
		return p.WriteLease(lease)
	}
	return err
}

// CreateLease will attempt to create a new lease.
func (p *Provider) CreateLease(lease *netdhcp.MACEntry) error {
	// FIXME: Validate lease
	keys := client.NewKeysAPI(p.c)
	cfg, err := p.Config(`WHAT-GOES-HERE?`)
	if err != nil {
		return err
	}
	network := cfg.Network() // FIXME: Is this how I get the network name?
	_, err = keys.Set(context.Background(), IPKey(network, lease.IP), lease.MAC.String(), &client.SetOptions{
		TTL: lease.Duration,
	})
	if err == nil {
		return p.WriteLease(lease)
	}
	return err
}

// WriteLease will attempt to write the lease data.
func (p *Provider) WriteLease(lease *netdhcp.MACEntry) error {
	// FIXME: Validate lease
	// NOTE: This does not save attributes. That should probably happen in a different function.
	// FIXME: Decide what to do if this call returns an error
	keys := client.NewKeysAPI(p.c)
	options := &client.SetOptions{TTL: lease.Duration} // FIXME: Add half second jitter to hide network delay?
	keys.Set(context.Background(), HardwareKey(lease.MAC)+"/ip", lease.IP.String(), options)
	return nil
}

// TODO: Write function for saving attributes to etcd?

func etcdNodeToMACEntry(root *client.Node, entry *netdhcp.MACEntry) {
	for _, node := range root.Nodes {
		if node.Dir {
			continue // Ignore subdirectories
		}
		key := strings.Replace(node.Key, root.Key+"/", "", 1)
		switch key {
		case "ip":
			entry.IP = net.ParseIP(node.Value)
			entry.Duration = time.Second * time.Duration(node.TTL) // FIXME: is this the best overall way to turn node.TTL into time.Duration?
		default:
			if entry.Attr == nil {
				entry.Attr = make(map[string]string)
			}
			entry.Attr[key] = node.Value
		}
	}
}

// RegisterA creates an A record for the given fully qualified domain name.
func (p *Provider) RegisterA(fqdn string, ip net.IP, exclusive bool, ttl uint32, expiration time.Duration) error {
	fqdn = cleanFQDN(fqdn)
	ipString := ip.String()
	ttlString := fmt.Sprintf("%d", ttl)
	ipHash := fmt.Sprintf("%x", sha1.Sum([]byte(ipString))) // hash the IP address so we can have a unique key name (no other reason for this, honestly)
	fqdnHash := fmt.Sprintf("%x", sha1.Sum([]byte(fqdn)))   // hash the hostname so we can have a unique key name (no other reason for this, honestly)

	keys := client.NewKeysAPI(p.c)

	options := &client.SetOptions{TTL: expiration}

	// Register the A record
	aKey := ResourceTypeKey(fqdn, "A")
	log.Printf("[REGISTER] [%s %d] %s. %d IN A %s\n", aKey, expiration, fqdn, ttl, ipString)
	_, err := keys.Set(context.Background(), aKey+"/val/"+ipHash, ipString, options)
	if err != nil {
		return err
	}
	if ttl != 0 {
		_, err := keys.Set(context.Background(), aKey+"/"+TTLField, ttlString, options)
		if err != nil {
			return err
		}
	}

	// Register the PTR record
	ptrKey := ArpaKey(ip) + "/@ptr"
	log.Printf("[REGISTER] [%s %d] %s. %d IN A %s\n", ptrKey, expiration, fqdn, ttl, ipString)
	_, err = keys.Set(context.Background(), ptrKey+"/val/"+fqdnHash, fqdn, options)
	if err != nil {
		return err
	}
	if ttl != 0 {
		_, err := keys.Set(context.Background(), aKey+"/ttl", ttlString, options)
		if err != nil {
			return err
		}
	}

	return err
}

func etcdKeyFromIP(ip net.IP) string {
	return "/dhcp/" + ip.String()
}

func etcdKeyFromMAC(mac net.HardwareAddr) string {
	return "/dhcp/" + mac.String()
}
