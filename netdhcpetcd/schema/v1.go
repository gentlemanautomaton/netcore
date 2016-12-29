package schema

var v1 = schema{
	root:   "/netcore/dhcp",
	schema: "",
}

// V1 returns the netcore etcd schema version 1.
func V1() Schema {
	return &v1
}
