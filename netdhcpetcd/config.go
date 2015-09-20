package netdhcpetcd

import (
	"dustywilson/netcore/netdhcp"
	"fmt"
	"net"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
)

// Init creates the initial etcd buckets for DHCP data.
func (p *Provider) Init() error {
	buckets := []string{RootBucket, ConfigBucket, ServerBucket, NetworkBucket, HostBucket, HardwareBucket}
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
func (p *Provider) Config(instance string) (netdhcp.Config, error) {
	fmt.Println("Getting CONFIG")

	keys := client.NewKeysAPI(p.c)

	cfg := netdhcp.NewCfg(p.defaults)
	cfg.Instance = instance
	cfg.Enabled = true // Can be overridden at any level

	_, configNodes, ok, err := responseNodes(keys.Get(context.Background(), ConfigBucket, &client.GetOptions{Recursive: true}))
	if err != nil && !etcdKeyNotFound(err) {
		return nil, err
	}
	if ok {
		nodesToConfig(configNodes, &cfg)
	}

	_, server, ok, err := responseNodes(keys.Get(context.Background(), ServerKey(instance), &client.GetOptions{Recursive: true}))
	if err != nil && !etcdKeyNotFound(err) {
		// FIXME: Return nil config when server isn't defined
		return nil, err
	}
	if ok {
		nodesToConfig(configNodes, &cfg)
	}

	if cfg.Network != "" {
		_, server, ok, err := responseNodes(keys.Get(context.Background(), NetworkKey(cfg.Network), &client.GetOptions{Recursive: true}))
		if err != nil && !etcdKeyNotFound(err) {
			// FIXME: Return nil config when server isn't defined
			return nil, err
		}
		if ok {
			nodesToConfig(configNodes, &cfg)
		}
	}

	fmt.Printf("DHCP ETCD CONFIG: [%+v]\n", &cfg)

	return netdhcp.NewConfig(&cfg), nil
}

func nodesToConfig(nodes client.Nodes, cfg *netdhcp.Cfg) error {
	for _, n := range nodes {
		switch nodeKey(n) {
		case NICField:
			cfg.NIC = n.Value
		case IPField:
			cfg.IP = net.ParseIP(n.Value).To4()
		case EnabledField:
			if n.Value != "1" {
				cfg.Enabled = false
			}
		case NetworkField:
			cfg.Network = n.Value
		case SubnetField:
			_, value, err := net.ParseCIDR(n.Value)
			if err != nil {
				return err
			}
			cfg.Subnet = value
		case GatewayField:
			cfg.Gateway = net.ParseIP(n.Value).To4()
		case DomainField:
			cfg.Domain = n.Value
		case TFTPField:
			cfg.TFTP = n.Value
		case NTPField:
			cfg.NTP = net.ParseIP(n.Value).To4()
		case LeaseDurationField:
			value, err := Atod(n.Value)
			if err != nil {
				return err
			}
			cfg.LeaseDuration = value
		case GuestPoolField:
			_, value, err := net.ParseCIDR(n.Value)
			if err != nil {
				return err
			}
			cfg.GuestPool = value
		}
	}
	return nil
}
