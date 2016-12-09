package netdhcp

import "time"

// Type represents a kind of device.
type Type struct {
	Attr
}

// TypeChan is a channel of type configuration updates.
type TypeChan <-chan TypeUpdate

// TypeUpdate is a type configuration update.
type TypeUpdate struct {
	Type      Type
	Timestamp time.Time
	Err       error
}

/*
// TypeObserver observes changes in type configuration and provides
// cached access to the most recent data.
type TypeObserver struct {
	mutex  *sync.RWMutex
	source *typeObserver
}

// Value returns the current configuration value of the type.
func (obs *TypeObserver) Value() (t Type, err error) {
	obs.mutex.RLock()
	if obs.source == nil {
		obs.mutex.RUnlock()
		return Type{}, ErrClosed
	}
	t, err = obs.source.Value()
	obs.mutex.RUnlock()
}

// Close releases any resources consumed by the type observer.
func (obs *TypeObserver) Close() {
	obs.mutex.Lock()
	defer obs.mutex.Unlock()

	if obs.source == nil {
		return
	}
	obs.source = nil

	go source.Done()
}

func (obs *TypeObserver) Listen(chanSize int) <-chan TypeUpdate {
	return nil
}

// typeCache holds cached type observers.
type typeCache struct {
	origin TypeProvider
	mutex  sync.RWMutex
	lookup map[string]*typeObserver
}

func (cache *typeCache) Value(id string) (Type, error) {
	for {
		obs := cache.obs(id)
		t, err := obs.Value(id)
		if err != ErrClosed {
			return t, err
		}
	}
}

func (cache *typeCache) obs(id string) *typeObserver {
	cache.mutex.RLock()
	obs, ok := cache.lookup[id]
	cache.mutex.RUnlock()
	if ok {
		return obs
	}
	cache.mutex.Lock()
	obs, ok = cache.lookup[id]
	if ok {
		cache.mutex.Unlock()
		return obs
	}
	cache.mutex.Unlock()
}

// typeObserver is the internal struct that observes a particular type. It is
// shared by all open instances of TypeObserver.
type typeObserver struct {
	source Provider
	mutex  *sync.RWMutex
	count  uint64
	cache  *typeCache
	value  Type
}

func (obs *typeObserver) Value() (Type, error) {

}

func (obs *typeObserver) Listen() TypeChan {
	return nil
}

func (obs *typeObserver) Done() {
	// TODO: Consider using sync.atomic
	obs.mutex.Lock()
	obs.count--
	if obs.count == 0 {

	}
}
*/
