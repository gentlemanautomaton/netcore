package netdhcp

import (
	"net"
	"time"
)

// Device represents a single logical device on the network, which may have
// one or more MAC addresses associated with it.
type Device struct {
	Attr
	Name  string
	Alias []string
	Type  string
	Addr  []net.HardwareAddr
}

// DeviceChan is a channel of device configuration updates.
type DeviceChan <-chan Device

// DeviceUpdate is a device configuration update.
type DeviceUpdate struct {
	Device    Device
	Timestamp time.Time
	Err       error
}
