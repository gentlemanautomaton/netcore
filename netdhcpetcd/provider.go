package netdhcpetcd

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/gentlemanautomaton/netcore/netdhcp"
)

// Provider implements all etcd storage interfaces necessary for operation of
// the DHCP service.
type Provider struct {
	c        clientv3.Client
	defaults netdhcp.Config
}

// NewProvider creates a new etcd DHCP provider with the given etcd client and
// default values.
func NewProvider(c clientv3.Client, defaults netdhcp.Config) netdhcp.Provider {
	return &Provider{
		c:        c,
		defaults: defaults,
	}
}
