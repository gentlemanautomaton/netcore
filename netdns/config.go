package netdns

import "time"

const (
	netEnabled        = true
	netDefaultTTL     = time.Hour * 3
	netMinimumTTL     = time.Second * 60
	netCacheRetention = 0
)

var (
	netForwarders = []string{"8.8.8.8:53", "8.8.4.4:53"} // default uses Google's Public DNS servers
)

// Config provides all of the necessary configuration context for the operation
// of a netcore DNS instance.
type Config interface {
	Instance() string
	Enabled() bool
	DefaultTTL() time.Duration
	MinimumTTL() time.Duration
	CacheRetention() time.Duration
	Forwarders() []string
}

// NewConfig creates an immutable instance of the Config interface.
func NewConfig(c *Cfg) Config {
	return &config{c.Copy()}
}

// DefaultConfig returns a Config interface with the default values for netcore.
func DefaultConfig() Config {
	return config{Cfg{
		Enabled:        netEnabled,
		DefaultTTL:     netDefaultTTL,
		MinimumTTL:     netMinimumTTL,
		CacheRetention: netCacheRetention,
	}}
}

// Cfg provides a mutable implementation of the Config interface. It can be made
// into an immutable Config instance via the NewConfig function.
type Cfg struct {
	Instance       string
	Enabled        bool
	DefaultTTL     time.Duration
	MinimumTTL     time.Duration
	CacheRetention time.Duration
	Forwarders     []string
}

// Copy will make a deep copy of the Cfg.
func (c Cfg) Copy() Cfg {
	return Cfg{
		Enabled:        c.Enabled,
		DefaultTTL:     c.DefaultTTL,
		MinimumTTL:     c.MinimumTTL,
		CacheRetention: c.CacheRetention,
		Forwarders:     append([]string(nil), c.Forwarders...), // Copy to avoid mutability
	}
}

// NewCfg creates a mutable Cfg instance from the given Config interface.
func NewCfg(c Config) Cfg {
	return Cfg{
		Instance:       c.Instance(),
		Enabled:        c.Enabled(),
		DefaultTTL:     c.DefaultTTL(),
		MinimumTTL:     c.MinimumTTL(),
		CacheRetention: c.CacheRetention(),
		Forwarders:     c.Forwarders(),
	}
}

// Validate returns an error if the config is invalid, otherwise it returns nil.
func Validate(c Config) error {
	if c == nil {
		return ErrNoConfig
	}
	return nil
}

// config provides an immutable implementation of the Config interface.
type config struct {
	x Cfg
}

func (c config) Instance() string {
	return c.x.Instance
}

func (c config) Enabled() bool {
	return c.x.Enabled
}

// DefaultTTL is the default TTL for all positive answers.
func (c config) DefaultTTL() time.Duration {
	return c.x.DefaultTTL
}

// MinimumTTL is the default value for the MINIMUM field in SOA records
// indicating how long to cache negative answers.
func (c config) MinimumTTL() time.Duration {
	return c.x.MinimumTTL
}

// CacheRetention is the duration for which resource records are retained in the
// DNS cache.
func (c config) CacheRetention() time.Duration {
	return c.x.CacheRetention
}

// Forwarders returns a list of DNS forwarders to which DNS queries will be
// be forwarded. Only queries for which this server is not authoritative will
// be forwarded. If no forwarders are specified then the server will not forward
// DNS queries. Forwarders are necessary to resolve CNAME and DNAME queries
// for which this server is not authoritative.
func (c config) Forwarders() []string {
	return append([]string(nil), c.x.Forwarders...) // Copy to avoid mutability
}

// Experimental:
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
