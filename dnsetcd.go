package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

func (db EtcdDB) InitDNS() {
	db.client.CreateDir("dns", 0)
}

func (db EtcdDB) GetDNS(name string, rrType string) (*DNSEntry, error) {
	//log.Printf("[Lookup [%s] [%s]]\n", q.Name, qType)
	rrType = strings.ToLower(rrType)
	key := etcdDNSKeyFromFQDN(name) + "/@" + rrType // structure the lookup key

	response, err := db.client.Get(key, true, true) // do the lookup
	if err != nil {
		return nil, err
	}

	if response != nil && response.Node != nil && len(response.Node.Nodes) > 0 {
		if rrType == "cname" {
			// FIXME: Check for infinite recursion?
		}
		return etcdNodeToDNSEntry(response.Node), nil
	}

	return nil, ErrNotFound
}

func (db EtcdDB) HasDNS(name string, rrType string) (bool, error) {
	rrType = strings.ToLower(rrType)
	key := etcdDNSKeyFromFQDN(name) + "/@" + rrType // structure the lookup key

	response, err := db.client.Get(key, false, false) // do the lookup
	if err != nil {
		return false, err
	}

	if response != nil && response.Node != nil && response.Node.Dir {
		return true, nil
	}

	return false, nil
}

func (db EtcdDB) RegisterA(fqdn string, ip net.IP, exclusive bool, ttl uint32, expiration uint64) error {
	fqdn = cleanFQDN(fqdn)
	ipString := ip.String()
	ttlString := fmt.Sprintf("%d", ttl)
	ipHash := fmt.Sprintf("%x", sha1.Sum([]byte(ipString))) // hash the IP address so we can have a unique key name (no other reason for this, honestly)
	fqdnHash := fmt.Sprintf("%x", sha1.Sum([]byte(fqdn)))   // hash the hostname so we can have a unique key name (no other reason for this, honestly)

	// Register the A record
	aKey := etcdDNSKeyFromFQDN(fqdn) + "/@a"
	log.Printf("[REGISTER] [%s %d] %s. %d IN A %s\n", aKey, expiration, fqdn, ttl, ipString)
	_, err := db.client.Set(aKey+"/val/"+ipHash, ipString, expiration)
	if err != nil {
		return err
	}
	if ttl != 0 {
		_, err := db.client.Set(aKey+"/ttl", ttlString, expiration)
		if err != nil {
			return err
		}
	}

	// Register the PTR record
	ptrKey := etcdDNSArpaKeyFromIP(ip) + "/@ptr"
	log.Printf("[REGISTER] [%s %d] %s. %d IN A %s\n", ptrKey, expiration, fqdn, ttl, ipString)
	_, err = db.client.Set(ptrKey+"/val/"+fqdnHash, fqdn, expiration)
	if err != nil {
		return err
	}
	if ttl != 0 {
		_, err := db.client.Set(aKey+"/ttl", ttlString, expiration)
		if err != nil {
			return err
		}
	}

	return err
}

func etcdNodeToDNSEntry(root *etcd.Node) *DNSEntry {
	entry := &DNSEntry{}
	for _, node := range root.Nodes {
		key := strings.Replace(node.Key, root.Key+"/", "", 1)
		if node.Dir {
			if key == "val" {
				entry.Values = make([]DNSValue, len(node.Nodes))
				for i, child := range node.Nodes {
					etcdNodeToDNSValue(child, &entry.Values[i])
				}
			}
		} else {
			switch key {
			case "ttl":
				ttl, _ := strconv.Atoi(node.Value)
				if ttl > 0 {
					entry.TTL = uint32(ttl)
				}
			default:
				if entry.Meta == nil {
					entry.Meta = make(map[string]string)
				}
				entry.Meta[key] = node.Value // NOTE: the keys are case-sensitive
			}
		}
	}
	return entry
}

func etcdNodeToDNSValue(node *etcd.Node, value *DNSValue) {
	value.Expiration = node.Expiration

	if node.TTL > 0 {
		value.TTL = uint32(node.TTL)
	}

	value.Value = node.Value

	if node.Nodes != nil && len(node.Nodes) > 0 {
		value.Attr = make(map[string]string)
		for _, attrNode := range node.Nodes {
			key := strings.Replace(attrNode.Key, node.Key+"/", "", 1)
			value.Attr[key] = attrNode.Value
		}
	}
}

func cleanFQDN(fqdn string) string {
	return strings.ToLower(strings.TrimSuffix(fqdn, "."))
}

func etcdDNSKeyFromFQDN(fqdn string) string {
	parts := strings.Split(cleanFQDN(fqdn), ".")   // breakup the queryed name
	path := strings.Join(reverseSlice(parts), "/") // reverse and join them with a slash delimiter
	return "/dns/" + path
}

func etcdDNSArpaKeyFromIP(ip net.IP) string {
	// FIXME: Support IPv6 addresses
	slashedIP := strings.Replace(ip.To4().String(), ".", "/", -1)
	return "dns/arpa/in-addr/" + slashedIP
}
