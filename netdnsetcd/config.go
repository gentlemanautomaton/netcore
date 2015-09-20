package netdnsetcd

import (
	"dustywilson/netcore/netdns"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/coreos/go-etcd/etcd"
	"golang.org/x/net/context"
)

/*
// GlobalConfig contains global DNS configuration.
type GlobalConfig struct {
	defaultTTL       time.Duration
	cacheNotFoundTTL time.Duration
	cacheRetention   time.Duration
	forwarders       []string
}

// ServerConfig contains configuration for a particular netcore server.
type ServerConfig struct {
	defaultTTL       time.Duration
	cacheRetention   time.Duration
	cacheNotFoundTTL time.Duration
	forwarders       []string
}
*/

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
	return netdns.NewConfig(netdns.Cfg{
	// FIXME: Write this
	}), nil
}

func old() {
	// Network
	{
		var response *etcd.Response
		var err error
		if setZone != nil && *setZone != "" {
			response, err = etc.Set("config/"+cfg.hostname+"/zone", *setZone, 0)
		} else {
			response, err = etc.Get("config/"+cfg.hostname+"/zone", false, false)
		}
		if err != nil {
			return nil, err
		}
		if response == nil || response.Node == nil || response.Node.Value == "" {
			return nil, ErrNoZone
		}
		cfg.zone = response.Node.Value
	}

	// DNSForwarders
	{
		cfg.dnsForwarders = []string{"8.8.8.8:53", "8.8.4.4:53"} // default uses Google's Public DNS servers
		response, err := etc.Get("config/"+cfg.zone+"/dnsforwarders", false, false)
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			cfg.dnsForwarders = strings.Split(",", response.Node.Value)
		}
	}

	// dnsCacheMaxTTL
	{
		cfg.dnsCacheMaxTTL = 0 // default to no caching
		response, err := etc.Get("config/"+cfg.zone+"/dnscachemaxttl", false, false)
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			value, err := strconv.Atoi(response.Node.Value)
			if err != nil {
				return nil, err
			}
			cfg.dnsCacheMaxTTL = time.Duration(value) * time.Second
		}
	}

	// dnsCacheMissingTTL
	{
		cfg.dnsCacheMissingTTL = 30 * time.Second // default setting is 30 seconds
		response, err := etc.Get("config/"+cfg.zone+"/dnscachemissingttl", false, false)
		if err != nil && !etcdKeyNotFound(err) {
			return nil, err
		}
		if response != nil && response.Node != nil && response.Node.Value != "" {
			value, err := strconv.Atoi(response.Node.Value)
			if err != nil {
				return nil, err
			}
			cfg.dnsCacheMissingTTL = time.Duration(value) * time.Second
		}
	}

	fmt.Printf("CONFIG: [%+v]\n", cfg)

	return cfg, nil
}
