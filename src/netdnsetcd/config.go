package netdnsetcd

import (
	"fmt"
	"strings"

	"github.com/coreos/etcd/client"
	"netdns"
	"golang.org/x/net/context"
)

// Init creates the initial etcd buckets for DNS data.
func (p *Provider) Init() error {
	buckets := []string{RootBucket, ConfigBucket, ServerBucket}
	keys := client.NewKeysAPI(p.c)
	for _, b := range buckets {
		_, err := keys.Set(context.Background(), b, "", &client.SetOptions{Dir: true})
		if err != nil {
			return err
		}
	}
	return nil
}

// Config returns a point-in-time view of the configuration for the instance.
func (p *Provider) Config(instance string) (netdns.Config, error) {
	fmt.Println("DNS ETCD CONFIG FETCH")

	keys := client.NewKeysAPI(p.c)

	cfg := netdns.NewCfg(p.defaults)
	cfg.Instance = instance
	cfg.Enabled = true // Can be overridden at any level

	_, configNodes, ok, err := responseNodes(keys.Get(context.Background(), ConfigBucket, &client.GetOptions{Recursive: true}))
	if err != nil && !etcdKeyNotFound(err) {
		return nil, err
	}
	if ok {
		nodesToConfig(configNodes, &cfg)
	}

	_, serverNodes, ok, err := responseNodes(keys.Get(context.Background(), ServerKey(instance), &client.GetOptions{Recursive: true}))
	if err != nil && !etcdKeyNotFound(err) {
		// FIXME: Return nil config when server isn't defined
		return nil, err
	}
	if ok {
		nodesToConfig(serverNodes, &cfg)
	}

	fmt.Printf("DHCP ETCD CONFIG: [%+v]\n", &cfg)

	return netdns.NewConfig(&cfg), nil
}

func nodesToConfig(nodes client.Nodes, cfg *netdns.Cfg) error {
	for _, n := range nodes {
		switch nodeKey(n) {
		case EnabledField:
			if n.Value != "1" {
				cfg.Enabled = false
			}
		case DefaultTTLField:
			value, err := Atod(n.Value)
			if err != nil {
				return err
			}
			cfg.DefaultTTL = value
		case MinimumTTLField:
			value, err := Atod(n.Value)
			if err != nil {
				return err
			}
			cfg.MinimumTTL = value
		case CacheRetentionField:
			value, err := Atod(n.Value)
			if err != nil {
				return err
			}
			cfg.CacheRetention = value
		case ForwardersField:
			cfg.Forwarders = strings.Split(",", n.Value)
		}
	}
	return nil
}
