package netdhcpetcd

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
)

func etcdKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Key not found")
}

// IP returns an IPEntry for the given IP address if it exists, otherwise it
// returns netdhcp.ErrNotFound
func (p *Provider) ip(key string) (net.IP, error) {
}

func responseNodes(r *client.Response, err error) (*client.Node, client.Nodes, bool, error) {
	if err != nil {
		return nil, nil, false, err
	}
	if r == nil || r.Node == nil || len(r) == 0 {
		return nil, nil, false, nil
	}
	return r.Node, r.Node.Nodes, true, nil
}

// nodeKey extracts the portion of the node's key after the last slash
func nodeKey(node *client.Node) string {
	split := strings.LastIndex(node.Key, "/")
	if split == -1 {
		return node.Key
	}
	return node.Key[split:]
}

// Atod converts a string containing a number of seconds into a time.Duration.
func Atod(value string) (time.Duration, error) {
	i, err := strconv.Atoi(value)
	if err != nil {
		return time.Duration{}, err
	}
	return time.Seconds * i, nil
}
