package netdhcp

import (
	"context"
	"time"
)

// Global represents configuration that is shared by all instances of the DHCP
// service.
type Global struct {
	Config
	Attr
	Network string
	Types   map[string]Type
}

// NetworkID returns the network identifier specified by the configuration.
func (g *Global) NetworkID() string {
	return g.Network
}

// GlobalChan is a channel of global configuration updates.
type GlobalChan <-chan GlobalUpdate

// GlobalUpdate is a global configuration update.
type GlobalUpdate struct {
	Global    *Global
	Timestamp time.Time
	Err       error
}

// GlobalObserver observes changes in global configuration and provides cached
// access to the most recent data.
type GlobalObserver interface {
	// Ready returns true if the observer has retrieved a value.
	Ready() bool

	// Wait will block until the observer is ready or the context is cancelled.
	Wait(ctx context.Context)

	// Value returns the most recently retrieved global configuration without
	// blocking.
	//
	// If the observer is not ready the call will return ErrNotReady.
	Value() (Global, error)

	// Listen returns a channel on which configuration updates will be broadcast.
	// The channel will be closed when the observer is closed. If the observer has
	// been closed Listen will return a closed channel.
	//
	// The returned channel's buffer will be of the provided size.
	Listen(chanSize int) <-chan GlobalUpdate

	// Status returns the status of the provider that the observer is reliant
	// upon. It can be used to determine whether the underlying provider is in a
	// connected or disconnected state, and for how long.
	Status()

	// Close releases any resources consumed by the observer.
	Close()
}
