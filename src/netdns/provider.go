package netdns

// Provider implements all storage interfaces necessary for operation of
// the DNS service.
type Provider interface {
	ConfigProvider
	DataProvider
}

// ConfigProvider provides DHCP configuration to the DHCP service.
type ConfigProvider interface {
	Init() error
	Config(instance string) (Config, error)
	//WatchConfig() chan<- Config
	//GlobalConfig() (*GlobalConfig, error)
	//ServerConfig(server string) (ServerConfig, error)
}

// DataProvider provides DHCP data to the DHCP service.
type DataProvider interface {
	RR(name string, rtype string) (*DNSEntry, error)
	HasRR(name string, rtype string) (bool, error)
}
