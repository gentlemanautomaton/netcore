package netdhcpetcd

/*
// GlobalObserver observes changes in global configuration and provides cached
// access to the most recent data.
type GlobalObserver struct {
	mutex sync.RWMutex
	ready sync.WaitGroup
	done  chan struct{} // Used by Close to signal completion
	//value GlobalResult
}

func newGlobalObserver(gc GlobalContext, w GlobalWatcher) *GlobalObserver {
	obs := &GlobalObserver{
		done: make(chan struct{}),
	}
	obs.ready.Add(1)
	go obs.run(w)
	return obs
}

func (obs *GlobalObserver) run(w GlobalWatcher) {
	// Step 1: Start the
	//obs.
}

// Ready returns true if the observer has already retrieved a value.
func (obs *GlobalObserver) Ready() bool {
	return false
}

// Wait will block until the observer is ready.
func (obs *GlobalObserver) Wait(ctx context.Context) {
	return
}

// Read returns the current value of the global configuration.
//
// If the observer has entered a ready state this call will return the most
// recent value and will not block. If the observer is not yet ready the call
// will block until the current configuration can be retrieved.
func (obs *GlobalObserver) Read(ctx context.Context) (Global, error) {
	return Global{}, errors.New("Not implemented yet.")
	// TODO: Block until ready, then return the value
}

// Close releases any resources consumed by the observer.
func (obs *GlobalObserver) Close() {
	obs.mutex.Lock()
	defer obs.mutex.Unlock()
	if obs.done == nil {
		return // Already closed
	}
	close(obs.done)
	obs.done = nil
}

// NetworkObserver observes changes in network configuration and provides
// cached access to the most recent data.
type NetworkObserver struct {
	NetworkContext
	mutex sync.RWMutex
	ready sync.WaitGroup
	done  chan struct{} // Used by Close to signal completion
	value NetworkResult
}

// InstanceObserver observes changes in instance configuration and provides
// cached access to the most recent data.
type InstanceObserver struct {
	InstanceContext
	mutex sync.RWMutex
	ready sync.WaitGroup
	done  chan struct{} // Used by Close to signal completion
	value NetworkResult
}

// Close releases any resources consumed by the cacher.
func (obs *InstanceObserver) Close() {
	obs.mutex.Lock()
	defer obs.mutex.Unlock()
	if obs.done == nil {
		return // Already closed
	}
	close(obs.done)
	obs.done = nil
}
*/
