package utils

import (
	"sync"
)

// ObserverManager allows one to remove its callback from its underlying Observable instance.
type ObserverManager interface {
	// RemoveObserver removes underlying callback from its Observable value.
	RemoveObserver()
}

// Observable allows one to attach an callback which will be called when someone calls the Notify method.
// It might be for example a change of some property's value.
type Observable interface {
	// AddObserver adds a new callback to an observable value.
	AddObserver(observer func(data interface{})) ObserverManager
	// Notify executes all callbacks with provided data.
	Notify(data interface{})
}

type observerPair struct {
	observer func(data interface{})
	ix       *int
}

type observable struct {
	removed   int
	observers []observerPair
}

func newObservable() *observable {
	return &observable{removed: 0, observers: nil}
}

// NewObservable creates an default instance of the Observable type.
func NewObservable() Observable {
	return newObservable()
}

type observerMemo struct {
	ix      *int
	manager *observable
}

func (om *observerMemo) RemoveObserver() {
	om.manager.removeObserver(om)
}

func (om *observable) fitToSize() {
	if om.removed == 0 {
		return
	}
	lastEmptyIx := 0
	for _, obs := range om.observers {
		if obs.ix != nil {
			om.observers[lastEmptyIx] = obs
			*obs.ix = lastEmptyIx
			lastEmptyIx++
		}
	}
	om.observers = om.observers[:lastEmptyIx]
	om.removed = 0
}

func (om *observable) AddObserver(observer func(data interface{})) ObserverManager {
	if cap(om.observers) == len(om.observers) {
		om.reallocate()
	}

	pair := observerPair{ix: new(int), observer: observer}
	*pair.ix = len(om.observers)
	om.observers = append(om.observers, pair)
	return &observerMemo{ix: pair.ix, manager: om}
}

func (om *observable) removeObserver(memo *observerMemo) {
	ix := *memo.ix
	om.observers[ix].observer = nil
	om.observers[ix].ix = nil
	om.removed++
	if om.size() < cap(om.observers)/4 {
		om.reallocate()
	} else if om.removed > cap(om.observers)/2 {
		om.fitToSize()
	}
}

func (om *observable) reallocate() {
	newObservers := make([]observerPair, 0, 2*om.size())
	newObservers = append(newObservers, om.observers...)
	om.observers = newObservers
	om.fitToSize()
}

func (om *observable) size() int {
	return len(om.observers) - om.removed
}

func (om *observable) notify(data interface{}) {
	for _, obs := range om.observers {
		if obs.ix != nil {
			obs.observer(data)
		}
	}
}

func (om *observable) Notify(data interface{}) {
	om.fitToSize()
	om.notify(data)
}

type safeObservable struct {
	manager *observable
	mx      sync.RWMutex
}

// NewThreadSafeObservable creates an instance of the Observable type that can be used safely by multiple go-routines.
func NewThreadSafeObservable() Observable {
	return &safeObservable{manager: newObservable(), mx: sync.RWMutex{}}
}

type safeObserverManager struct {
	manager ObserverManager
	mx      *sync.RWMutex
}

func (som *safeObserverManager) RemoveObserver() {
	som.mx.Lock()
	defer som.mx.Unlock()
	som.manager.RemoveObserver()
}

func (som *safeObservable) AddObserver(observer func(data interface{})) ObserverManager {
	som.mx.Lock()
	defer som.mx.Unlock()
	obsMgr := som.manager.AddObserver(observer)
	return &safeObserverManager{manager: obsMgr, mx: &som.mx}
}

func (som *safeObservable) Notify(data interface{}) {
	som.mx.RLock()
	refit := som.manager.removed > 0
	som.mx.RUnlock()
	if refit {
		som.mx.Lock()
		som.manager.fitToSize()
		som.mx.Unlock()
	}

	som.mx.RLock()
	defer som.mx.RUnlock()
	som.manager.notify(data)
}
