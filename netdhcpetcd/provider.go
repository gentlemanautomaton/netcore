package netdhcpetcd

import (
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

type Provider struct {
	client *etcd.Client
}

func NewProvider(client etcd.Client) Provider {
	client.SetConsistency("WEAK_CONSISTENCY")
	return Provider{client}
}

func etcdKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Key not found")
}
