package gonsai

import (
	"bytes"
	"golang.org/x/crypto/sha3"
)

const hashLen = 32 // for longer hashes ShakeSum128 needs to be replaced with sth stronger
var zeroHash = []byte{0}

type hash []byte
type updater func([]byte) []byte

// DataStorage is an interface for storing slices of bytes using fixed length keys.
type DataStorage interface {
	// Hash returns control hash of the current state
	Hash() hash
	// Find looks for the value for the given key. Returns proof for that value, or proof that key not present.
	// Multiple Finds can happen at the same time. They ignore the stage-zone.
	Find([]byte) ([]byte, *proof)
	// StageUpdate stages an update of the given key with supplied updater function. Returns true if any change was applied to the stage-zon.
	StageUpdate([]byte, updater) bool
	// StageInsert stages an insert of the given key-value pair. If the key is already present, overwrites the value. Returns true if any change was applied to the stage-zone.
	StageInsert([]byte, []byte) bool
	// StageDelete stages deletion of the given key, if present. Returns true if any change was applied to the stage-zone.
	StageDelete([]byte) bool
	// StageHash() calculates control hash of the database current state together with all currently staged changes.
	StageHash() hash
	// Commit Updates the current state of the database with the contents of the stage-zone. Find calls are stopped during that.
	Commit()
	// Reset discards all the changes present in the stage-zone.
	Reset()
	// Key length
	KeyLen() int
}

// info about a node needed for calculating hash
type info struct {
	name []byte
	hash hash
}

func hashNodes(nodes []*info) hash {
	if len(nodes) == 0 {
		return zeroHash
	}
	buf := bytes.NewBuffer(nil)
	for _, nd := range nodes {
		buf.Write(nd.name)
		buf.Write(nd.hash)
	}
	result := make([]byte, hashLen)
	sha3.ShakeSum128(result, buf.Bytes())
	return result
}

func hashData(data []byte) hash {
	result := make([]byte, hashLen)
	sha3.ShakeSum128(result, data)
	return result
}

func sameData(one, two []byte) bool {
	if len(one) != len(two) {
		return false
	}
	for i := range one {
		if one[i] != two[i] {
			return false
		}
	}
	return true
}

// commonPrefix is the length of common prefix of two keys
func commonPrefix(one, two []byte) int {
	i := 0
	stop := len(one)
	if len(two) < stop {
		stop = len(two)
	}
	for i < stop {
		if one[i] != two[i] {
			break
		}
		i++
	}
	return i
}
