package netdhcp

import (
	"context"
	"time"
)

// Instance represents an instance of the DHCP service and includes its
// configuration.
type Instance struct {
	Config
	Attr
	ID      string
	Network string
}

// NetworkID returns the network identifier specified by the configuration.
func (i *Instance) NetworkID() string {
	return i.Network
}

// InstanceChan is a channel of instance configuration updates.
type InstanceChan <-chan InstanceUpdate

// InstanceUpdate is an instance configuration update.
type InstanceUpdate struct {
	Instance  *Instance
	Timestamp time.Time
	Err       error
}

// InstanceObserver observes changes in instance configuration and provides
// cached access to the most recent data.
type InstanceObserver interface {
	// Ready returns true if the observer has retrieved a value.
	Ready() bool

	// Wait will block until the observer is ready or the context is cancelled.
	Wait(ctx context.Context)

	// Value returns the most recently retrieved instance configuration without
	// blocking.
	//
	// If the observer is not ready the call will return ErrNotReady.
	Value() (Instance, error)

	// Listen returns a channel on which configuration updates will be broadcast.
	// The channel will be closed when the observer is closed. If the observer has
	// been closed Listen will return a closed channel.
	//
	// The returned channel's buffer will be of the provided size.
	Listen(chanSize int) <-chan InstanceUpdate

	// Status returns the status of the provider that the observer is reliant
	// upon. It can be used to determine whether the underlying provider is in a
	// connected or disconnected state, and for how long.
	Status()

	// Close releases any resources consumed by the observer.
	Close()
}
