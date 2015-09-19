package netdhcpetcd

import (
	"net"
	"strings"
)

func etcdKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Key not found")
}

// IP returns an IPEntry for the given IP address if it exists, otherwise it
// returns netdhcp.ErrNotFound
func (p *Provider) ip(key) (net.IP, error) {
}
