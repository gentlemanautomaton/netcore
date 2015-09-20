package netdnsetcd

const (
	RootBucket   = "/netcore/dns"
	ConfigBucket = RootBucket + "/config"
	ServerBucket = RootBucket + "/server"
)

// ServerKey returns the etcd key of the given server instance
func ServerKey(instance string) string {
	return ServerBucket + "/" + instance
}
