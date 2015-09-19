package netdnsetcd

import "strings"

// FIXME: Move this
func etcdKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Key not found")
}
