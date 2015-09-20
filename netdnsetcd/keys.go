package netdnsetcd

const (
	RootBucket   = "/netcore/dns"
	ConfigBucket = RootBucket + "/config"
	ServerBucket = RootBucket + "/server"
)

const (
	EnabledField        = "enabled"
	DefaultTTLField     = "defaultttl"
	MinimumTTLField     = "minimumttl"
	CacheRetentionField = "cacheretention"
	ForwardersField     = "forwarders"
)

// ServerKey returns the etcd key of the given server instance
func ServerKey(instance string) string {
	return ServerBucket + "/" + instance
}
