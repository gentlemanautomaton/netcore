package netdns

import (
	"errors"
	"time"
)

var (
	// ErrNotFound indicates that the requested resource record does not exist
	ErrNotFound = errors.New("not found")
)

// DNSEntry represents a resource record retrieved from the underlying provider.
type DNSEntry struct {
	TTL    uint32
	Values []DNSValue
	Meta   map[string]string
}

// DNSValue represents a resource record value retrieved from the underlying
// provider.
type DNSValue struct {
	Expiration *time.Time
	TTL        uint32
	Value      string
	Attr       map[string]string
}

type dnsEntryResult struct {
	Entry *DNSEntry
	Err   error
	RType uint16
}

const (
	dnsCacheBufferSize = 512
)
