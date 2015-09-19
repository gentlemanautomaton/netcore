package netdns

import "time"

const (
	ncDefaultTTL             = time.Hour * 3
	ncMinimumTTL             = time.Second * 60
	ncCacheRetention         = 0
	ncCacheRetentionNotFound = time.Second * 30
)

// Config provides all of the necessary configuration context for the operation
// of a netcore DNS instance.
type Config interface {
	DefaultTTL() time.Duration
	MinimumTTL() time.Duration
	CacheRetention() time.Duration
	//NegativeCacheRetention() time.Duration
	Forwarders() []string
}

// NewConfig creates an immutable instance of the Config interface.
func NewConfig(defaultTTL, minimumTTL, cacheRetention, cacheRetentionNotFound time.Duration, forwarders []string) Config {
	return &config{
		defaultTTL:     defaultTTL,
		minimumTTL:     minimumTTL,
		cacheRetention: cacheRetention,
		//cacheRetentionNotFound: cacheRetentionNotFound,
		forwarders: append([]string(nil), forwarders...), // Copy to avoid mutability
	}
}

// DefaultConfig returns a Config interface with the default values for netcore.
func DefaultConfig() Config {
	return NewConfig(ncDefaultTTL, ncMinimumTTL, ncCacheRetention, ncCacheRetentionNotFound, nil)
}

// config provides an immutable implementation of the Config interface.
type config struct {
	defaultTTL     time.Duration
	minimumTTL     time.Duration
	cacheRetention time.Duration
	//cacheRetentionNotFound time.Duration
	forwarders []string
}

// DefaultTTL is the default TTL for all positive answers.
func (c config) DefaultTTL() time.Duration {
	return c.defaultTTL
}

// MinimumTTL is the default value for the MINIMUM field in SOA records
// indicating how long to cache negative answers.
func (c config) MinimumTTL() time.Duration {
	return c.minimumTTL
}

// CacheRetention is the duration for which resource records are retained in the
// DNS cache.
func (c config) CacheRetention() time.Duration {
	return c.cacheRetention
}

// CacheRetentionNotFound is the duration for which records that aren't found
// are cached as missing. The value of CacheRetention is used instead if it is
// smaller.
/*
func (c config) CacheRetentionNotFound() time.Duration {
	return c.cacheRetentionNotFound
}
*/

// Forwarders returns a list of DNS forwarders to which DNS queries will be
// be forwarded. Only queries for which this server is not authoritative will
// be forwarded. If no forwarders are specified then the server will not forward
// DNS queries. Forwarders are necessary to resolve CNAME and DNAME queries
// for which this server is not authoritative.
func (c config) Forwarders() []string {
	return append([]string(nil), c.forwarders...) // Copy to avoid mutability
}

/*
// Config provides all of the necessary configuration context for the operation
// of a netcore DNS instance.
type Config struct {
	g GlobalConfig
	s ServerConfig
}

// GlobalConfig contains global DNS configuration.
type GlobalConfig struct {
	defaultTTL       time.Duration
	cacheNotFoundTTL time.Duration
	cacheRetention   time.Duration
	forwarders       []string
}

// ServerConfig contains configuration for a particular netcore server.
type ServerConfig struct {
	defaultTTL       time.Duration
	cacheRetention   time.Duration
	cacheNotFoundTTL time.Duration
	forwarders       []string
}
*/
