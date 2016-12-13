package netdhcp

import "net"

// MAC represents the binding configuration for a specific MAC address.
type MAC struct {
	Attr
	Addr        net.HardwareAddr
	Device      string // FIXME: What type are we using for device IDs?
	Type        string
	Restriction Mode // TODO: Decide whether this is inclusive or exclusive
	Assignments AssignmentSet
}

// BindingAssignments returns the set of IP assignments for the binding,
// including both reserved and previous dynamic IP assignments.
func (m *MAC) BindingAssignments() AssignmentSet {
	return m.Assignments
}

/*
type MACAttr struct {
}
*/

/*
// MACEntry represents a MAC address record retrieved from the underlying
// provider.
type MACEntry struct {
	Network     string
	MAC         net.HardwareAddr
	Assignments AssignmentSet
	Duration    time.Duration
	Attr        map[string]string
}
*/

// HasMode returns true if the given IP type is enabled for this MAC.
/*
func (m *MAC) HasMode(mode IPType) bool {
	return m.Mode&IPType != 0
}
*/
