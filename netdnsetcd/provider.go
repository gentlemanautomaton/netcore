package netdnsetcd

import (
	"dustywilson/netcore/netdns"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

// Provider implements all etcd storage interfaces necessary for operation of
// the DNS service.
type Provider struct {
	client   *etcd.Client
	defaults Config
}

// NewProvider creates a new etcd DNS provider with the given etcd client.
func NewProvider(client etcd.Client, defaults netdns.Config) Provider {
	client.SetConsistency("WEAK_CONSISTENCY")
	return Provider{client}
}

func etcdKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Key not found")
}
