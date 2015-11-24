package netdnsetcd

import (
	"github.com/coreos/etcd/client"
	"netdns"
)

// Provider implements all etcd storage interfaces necessary for operation of
// the DNS service.
type Provider struct {
	c        client.Client
	defaults netdns.Config
}

// NewProvider creates a new etcd DNS provider with the given etcd client and
// default values.
func NewProvider(c client.Client, defaults netdns.Config) netdns.Provider {
	return &Provider{
		c:        c,
		defaults: defaults,
	}
}
