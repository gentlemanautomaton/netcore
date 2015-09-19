package netdhcpetcd

import (
	"dustywilson/netcore/netdhcp"

	"github.com/coreos/go-etcd/etcd"
)

// Provider implements all etcd storage interfaces necessary for operation of
// the DHCP service.
type Provider struct {
	client   *etcd.Client
	defaults netdhcp.Config
}

// NewProvider creates a new etcd DNS provider with the given etcd client and
// default values.
func NewProvider(client etcd.Client, defaults netdhcp.Config) netdhcp.Provider {
	client.SetConsistency("WEAK_CONSISTENCY") // FIXME: Is this the right place for this?
	return &Provider{
		client:   client,
		defaults: defaults,
	}
}
