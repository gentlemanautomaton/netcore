package netdnsetcd

import (
	"dustywilson/netcore/netdns"

	"github.com/coreos/go-etcd/etcd"
)

// Provider implements all etcd storage interfaces necessary for operation of
// the DNS service.
type Provider struct {
	client   *etcd.Client
	defaults netdns.Config
}

// NewProvider creates a new etcd DNS provider with the given etcd client and
// default values.
func NewProvider(client *etcd.Client, defaults netdns.Config) netdns.Provider {
	client.SetConsistency("WEAK_CONSISTENCY") // FIXME: Is this the right place for this?
	return &Provider{
		client:   client,
		defaults: defaults,
	}
}
