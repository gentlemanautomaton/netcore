package main

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

func (db EtcdDB) InitDHCP() {
	db.client.CreateDir("dhcp", 0)
}

func (db EtcdDB) GetIP(ip net.IP) (IPEntry, error) {
	key := etcdKeyFromIP(ip)
	response, err := db.client.Get(key, false, false)
	if response == nil || response.Node == nil {
		return IPEntry{}, errors.New("Not Found")
	}
	mac, err := net.ParseMAC(response.Node.Value)
	if err != nil {
		return IPEntry{}, err
	}
	return IPEntry{MAC: mac}, nil
}

func (db EtcdDB) HasIP(ip net.IP) bool {
	key := etcdKeyFromIP(ip)
	response, _ := db.client.Get(key, false, false)
	if response != nil && response.Node != nil {
		return true
	}
	return false
}

func (db EtcdDB) GetMAC(mac net.HardwareAddr, cascade bool) (*MACEntry, bool, error) {
	// TODO: First attempt to retrieve the entry from a cache of some kind (that can be dirtied)
	// NOTE: The cache should always return a deep copy of the cached value
	entry := MACEntry{MAC: mac}

	// Copy cascaded attributes by making recursive calls to this function
	if cascade && len(mac) > 1 {
		parent, _, _ := db.GetMAC(mac[0:len(mac)-1], cascade) // Chop off the last byte for each recursive call
		if parent != nil {
			entry.Attr = parent.Attr // Only safe if we receive a deep copy of the cached value
		}
	}

	// Fetch attributes and lease data for this MAC
	key := etcdKeyFromMAC(mac)
	response, err := db.client.Get(key, true, true) // do the lookup
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

func (db EtcdDB) RenewLease(lease *MACEntry) error {
	// FIXME: Validate lease
	duration := uint64(lease.Duration.Seconds() + 0.5) // Half second jitter to hide network delay
	_, err := db.client.CompareAndSwap("dhcp/"+lease.IP.String(), lease.MAC.String(), duration, lease.MAC.String(), 0)
	if err == nil {
		return db.WriteLease(lease)
	}
	return err
}

func (db EtcdDB) CreateLease(lease *MACEntry) error {
	// FIXME: Validate lease
	duration := uint64(lease.Duration.Seconds() + 0.5)
	_, err := db.client.Create("dhcp/"+lease.IP.String(), lease.MAC.String(), duration)
	if err == nil {
		return db.WriteLease(lease)
	}
	return err
}

func (db EtcdDB) WriteLease(lease *MACEntry) error {
	// FIXME: Validate lease
	// NOTE: This does not save attributes. That should probably happen in a different function.
	duration := uint64(lease.Duration.Seconds() + 0.5) // Half second jitter to hide network delay
	// FIXME: Decide what to do if either of these calls returns an error
	db.client.CreateDir("dhcp/"+lease.MAC.String(), 0)
	db.client.Set("dhcp/"+lease.MAC.String()+"/ip", lease.IP.String(), duration)
	return nil
}

// TODO: Write function for saving attributes to etcd?

func etcdNodeToMACEntry(root *etcd.Node, entry *MACEntry) {
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
