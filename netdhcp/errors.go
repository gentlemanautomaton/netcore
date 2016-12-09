package netdhcp

import "errors"

var (
	// ErrNotReady indicates that the configuration has not yet been retrieved.
	ErrNotReady = errors.New("Configuration is not ready.")
	// ErrNoConfig indicates that no configuration was provided to the DHCP
	// service.
	ErrNoConfig = errors.New("Configuration not provided")
	// ErrNoConfigNetwork indicates that no network was specified in the DHCP
	// configuration.
	ErrNoConfigNetwork = errors.New("Network not specified in configuration")
	// ErrNoConfigIP indicates that no IP address was provided in the DHCP
	// configuration.
	ErrNoConfigIP = errors.New("IP not specified in configuration")
	// ErrNoConfigSubnet indicates that no subnet was provided in the DHCP
	// configuration.
	ErrNoConfigSubnet = errors.New("Subnet not specified in configuration")
	// ErrNoConfigNIC indicates that no network interface was provided in the
	// DHCP configuration.
	ErrNoConfigNIC = errors.New("NIC not specified in configuration")
	// ErrNotFound indicates that the requested data does not exist
	ErrNotFound = errors.New("Not found")

	// ErrNoZone is an error returned during config init to indicate that the host has not been assigned to a zone in etcd keyed off of its hostname
	ErrNoZone = errors.New("This host has not been assigned to a zone.")

	// ErrNoDHCPIP is an error returned during config init to indicate that the host has not been assigned to a zone in etcd keyed off of its hostname
	ErrNoDHCPIP = errors.New("This host has not been assigned a DHCP IP.")

	// ErrNoGateway is an error returned during config init to indicate that the zone has not been assigned a gateway in etcd keyed off of the zone name
	ErrNoGateway = errors.New("This zone does not have an assigned gateway.")
)

var (
	// ErrClosed is returned when an action cannot be performed because a
	// necessary resource has already been disposed of.
	ErrClosed = errors.New("The resource is closing or closed.")

	// ErrNoLeaseGateway is returned when a lease cannot be provided because its
	// configuration does not specify a gateway.
	ErrNoLeaseGateway = errors.New("A gateway has not been specified for the lease.")

	// ErrNoLeaseSubnet is returned when a lease cannot be provided because its
	// configuration does not specify a subnet.
	ErrNoLeaseSubnet = errors.New("A subnet has not been specified for the lease.")
)
