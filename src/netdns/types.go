package netdns

import (
	"errors"
	"time"
)

var (
	// ErrNoConfig indicates that no configuration was provided to the DHCP
	// service.
	ErrNoConfig = errors.New("Configuration not provided")

	// ErrNotFound indicates that the requested resource record does not exist
	ErrNotFound = errors.New("not found")
)

// Completion is returned via the Service.Done() channel when the service exits.
type Completion struct {
	// Initialized indiciates whether the services finished initializing before exiting.
	Initialized bool
	// Err indictes the error that caused the service to exit in the case of
	// failure.
	Err error
}

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
