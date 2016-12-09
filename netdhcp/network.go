package netdhcp

import (
	"context"
	"time"
)

// Network represents a DHCP network with common configuration.
type Network struct {
	Config
	Attr
	ID    string
	Types map[string]Type
}

// NetworkChan is a channel of network configuration updates.
type NetworkChan <-chan NetworkUpdate

// NetworkUpdate is a network configuration update.
type NetworkUpdate struct {
	Network   *Network
	Timestamp time.Time
	Err       error
}

// NetworkObserver observes changes in network configuration and provides
// cached access to the most recent data.
type NetworkObserver interface {
	// Ready returns true if the observer has already retrieved a value.
	Ready() bool

	// Wait will block until the observer is ready or the context is cancelled.
	Wait(ctx context.Context)

	// Value returns the most recently retrieved network configuration without
	// blocking.
	//
	// If the observer is not ready the call will return ErrNotReady.
	Value() (Network, error)

	// Listen returns a channel on which configuration updates will be broadcast.
	// The channel will be closed when the observer is closed. If the observer has
	// been closed Listen will return a closed channel.
	//
	// The returned channel's buffer will be of the provided size.
	Listen(chanSize int) <-chan NetworkUpdate

	// Status returns the status of the provider that the observer is reliant
	// upon. It can be used to determine whether the underlying provider is in a
	// connected or disconnected state, and for how long.
	Status()

	// Close releases any resources consumed by the observer.
	Close()
}
