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
	ix       int
}

type observable struct {
	observers []*observerPair
}

func newObservable() *observable {
	return &observable{observers: nil}
}

// NewObservable creates an default instance of the Observable type.
func NewObservable() Observable {
	return newObservable()
}

type observerMemo struct {
	id      *observerPair
	manager *observable
}

func (om *observerMemo) RemoveObserver() {
	om.manager.removeObserver(om.id)
}

func (om *observable) AddObserver(observer func(data interface{})) ObserverManager {
	pair := &observerPair{ix: len(om.observers), observer: observer}
	om.observers = append(om.observers, pair)
	return &observerMemo{id: pair, manager: om}
}

func (om *observable) removeObserver(memo *observerPair) {
	ix := memo.ix
	new := make([]*observerPair, 0, len(om.observers)-1)
	new = append(new, om.observers[:ix]...)
	new = append(new, om.observers[ix+1:]...)
	for ix, obs := range new {
		obs.ix = ix
	}
	om.observers = new
}

func (om *observable) Notify(data interface{}) {
	for _, obs := range om.observers {
		obs.observer(data)
	}
}

type safeObservable struct {
	manager *observable
	mx      sync.RWMutex
}

// NewThreadSafeObservable creates an instance of the Observable type that can be used safely by multiple go-routines.
func NewThreadSafeObservable() Observable {
	return &safeObservable{manager: newObservable(), mx: sync.RWMutex{}}
}

func (som *safeObservable) AddObserver(observer func(data interface{})) ObserverManager {
	som.mx.Lock()
	defer som.mx.Unlock()
	obsMgr := som.manager.AddObserver(observer)
	return &safeObserverManager{manager: obsMgr, mx: &som.mx}
}

func (som *safeObservable) Notify(data interface{}) {
	som.mx.RLock()
	defer som.mx.RUnlock()
	som.manager.Notify(data)
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
