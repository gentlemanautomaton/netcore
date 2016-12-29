package schema

// etcd tags used in construction of various keys
const (
	SchemaTag   = "schema"
	ProviderTag = "provider"
	GlobalTag   = "global"
	InstanceTag = "instance"
	NetworkTag  = "network"
	DeviceTag   = "device"
	TypeTag     = "type"
	HardwareTag = "mac"
	ConfigTag   = "config"
)

// Schema provides all of the ectd layout information necessary to access
// netcore DCHP configuration data.
type Schema interface {
}

type schema struct {
	root   string
	schema string
}

func (s *schema) Root() string {
	return s.root
}

func (s *schema) Schema() string {
	return s.schema
}

// etcd bucket keys
const (
	Root = "/netcore/dhcp"

	//Schema   = Root + "/" + SchemaTag
	Global   = Root + "/" + GlobalTag
	Instance = Root + "/" + InstanceTag
	Network  = Root + "/" + NetworkTag
	Device   = Root + "/" + DeviceTag
	Type     = Root + "/" + TypeTag
	Hardware = Root + "/" + HardwareTag

	GlobalType   = Global + "/" + TypeTag
	GlobalConfig = Global + "/" + ConfigTag

	//HostBucket       = RootBucket + "/host"
	//HardwareBucket   = RootBucket + "/mac"
	//ResourceBucket   = RootBucket + "/resource"
	//ArpaBucket       = ResourceBucket + "/arpa/in-addr/"
)

// etcd field keys
const (
	NICField           = "nic"
	LeaseField         = "lease"
	EnabledField       = "enabled"
	NetworkField       = "network"
	SubnetField        = "subnet"
	GatewayField       = "gw"
	DomainField        = "domain"
	TFTPField          = "tftp"
	NTPField           = "ntp"
	LeaseDurationField = "leaseduration"
	GuestPoolField     = "pool"
	TTLField           = "ttl"
)
