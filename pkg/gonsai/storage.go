// Package gonsai implements the most bad-ass merkle tree that has ever existed.
package gonsai

import (
	"sync"
)

type storage struct {
	keyLen   int
	root     *node
	nilStage bool
	mx       sync.RWMutex
	stageMx  sync.Mutex
}

// NewStorage initializes empty storage with the given key length
func NewStorage(keyLen int) DataStorage {
	return &storage{
		keyLen: keyLen,
		root: &node{
			name:       nil,
			hash:       zeroHash,
			children:   make(map[byte]*node),
			stChildren: make(map[byte]*node),
		},
	}
}

func (ds *storage) KeyLen() int { return ds.keyLen }

func (ds *storage) Hash() hash {
	ds.mx.RLock()
	defer ds.mx.RUnlock()
	return ds.root.hash
}

func (ds *storage) Find(key []byte) ([]byte, *proof) {
	ds.checkKey(key)
	ds.mx.RLock()
	defer ds.mx.RUnlock()
	return ds.root.Find(key)
}

func (ds *storage) StageUpdate(key []byte, upd updater) bool {
	ds.checkKey(key)
	ds.stageMx.Lock()
	defer ds.stageMx.Unlock()
	return ds.root.StageUpdate(key, upd)
}

func (ds *storage) StageInsert(key, value []byte) bool {
	ds.checkKey(key)
	ds.stageMx.Lock()
	defer ds.stageMx.Unlock()
	return ds.root.StageInsert(key, value)
}

func (ds *storage) StageDelete(key []byte) bool {
	ds.checkKey(key)
	ds.stageMx.Lock()
	defer ds.stageMx.Unlock()
	ret, _ := ds.root.StageDelete(key)
	return ret
}
func (ds *storage) Reset() {
	ds.stageMx.Lock()
	defer ds.stageMx.Unlock()
	ds.root.Reset()
}

func (ds *storage) StageHash() hash {
	ds.stageMx.Lock()
	defer ds.stageMx.Unlock()
	return ds.root.calcStHash()
}

func (ds *storage) Commit() {
	ds.stageMx.Lock()
	defer ds.stageMx.Unlock()
	ds.mx.Lock()
	defer ds.mx.Unlock()
	ds.root.calcStHash()
	ds.root.Commit()
}

func (ds *storage) checkKey(key []byte) {
	if len(key) != ds.keyLen {
		panic("wrong key length")
	}
}
