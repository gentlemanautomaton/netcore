package netdhcp

import (
	"net"
	"time"
)

// Assignment represents an IP address assigned to a MAC address.
type Assignment struct {
	Mode     Mode
	Priority int
	Created  time.Time
	Assigned time.Time
	Address  net.IP
}

// AssignmentSet represents a set of assigned IP addreses that can be sorted
// according to the address selection rules.
type AssignmentSet []*Assignment

func (slice AssignmentSet) Len() int {
	return len(slice)
}

func (slice AssignmentSet) Less(i, j int) bool {
	a, b := slice[i], slice[j]
	if a.Mode < b.Mode {
		return true
	}
	if a.Mode > b.Mode {
		return false
	}
	if a.Priority < b.Priority {
		return true
	}
	if a.Priority > b.Priority {
		return false
	}
	if a.Assigned.Before(b.Assigned) {
		return true
	}
	if b.Assigned.Before(a.Assigned) {
		return false
	}
	return false
}

func (slice AssignmentSet) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
