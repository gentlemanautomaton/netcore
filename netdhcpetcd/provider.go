package netdhcpetcd

import (
	"dustywilson/netcore/netdhcp"

	"github.com/coreos/etcd/client"
)

// Provider implements all etcd storage interfaces necessary for operation of
// the DHCP service.
type Provider struct {
	c        client.Client
	defaults netdhcp.Config
}

// NewProvider creates a new etcd DNS provider with the given etcd client and
// default values.
func NewProvider(c client.Client, defaults netdhcp.Config) netdhcp.Provider {
	return &Provider{
		c:        c,
		defaults: defaults,
	}
}
