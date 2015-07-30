package main

import (
	"errors"
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

	return nil, errors.New("Not Found") // FIXME: Return a more proper error type
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

func etcdDNSKeyFromFQDN(fqdn string) string {
	parts := strings.Split(strings.TrimSuffix(fqdn, "."), ".") // breakup the queryed name
	path := strings.Join(reverseSlice(parts), "/")             // reverse and join them with a slash delimiter
	return strings.ToLower("/dns/" + path)
}
