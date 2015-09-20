package netdnsetcd

import (
	"crypto/sha1"
	"dustywilson/netcore/netdns"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
)

// RR returns the resource record for the given name and type.
func (p *Provider) RR(name string, rrType string) (*netdns.DNSEntry, error) {
	//log.Printf("[Lookup [%s] [%s]]\n", q.Name, qType)
	keys := client.NewKeysAPI(p.c)
	response, err := keys.Get(context.Background(), ResourceTypeKey(name, rrType), &client.GetOptions{Recursive: true, Sort: true}) // do the lookup
	if err != nil {
		return nil, err
	}

	if response != nil && response.Node != nil && len(response.Node.Nodes) > 0 {
		if rrType == "cname" {
			// FIXME: Check for infinite recursion?
		}
		return etcdNodeToDNSEntry(response.Node), nil
	}

	return nil, netdns.ErrNotFound
}

// HasRR returns true if a resource record exists with the given name and type.
func (p *Provider) HasRR(name string, rrType string) (bool, error) {
	keys := client.NewKeysAPI(p.c)
	response, err := keys.Get(context.Background(), ResourceTypeKey(name, rrType), nil) // do the lookup
	if err != nil {
		return false, err
	}
	if response != nil && response.Node != nil && response.Node.Dir {
		return true, nil
	}
	return false, nil
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

func etcdNodeToDNSEntry(root *client.Node) *netdns.DNSEntry {
	entry := &netdns.DNSEntry{}
	for _, node := range root.Nodes {
		key := strings.Replace(node.Key, root.Key+"/", "", 1)
		if node.Dir {
			if key == "val" {
				entry.Values = make([]netdns.DNSValue, len(node.Nodes))
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

func etcdNodeToDNSValue(node *client.Node, value *netdns.DNSValue) {
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
