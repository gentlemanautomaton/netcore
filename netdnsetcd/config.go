package netdnsetcd

/*
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

// Init creates the initial etcd structure for DNS data.
func (p *Provider) Init() {
	p.client.CreateDir("dns", 0)
}

// Config returns a point-in-time view of the configuration for the instance.
func (p *Provider) Config(instance string) {
	return p.d
}
