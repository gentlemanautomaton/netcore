package netdnsetcd

import "strings"

// FIXME: Move this
func etcdKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Key not found")
}

func reverseSlice(in []string) []string {
	out := make([]string, len(in))
	for i := range in {
		out[len(in)-i-1] = in[i]
	}
	return out
}
