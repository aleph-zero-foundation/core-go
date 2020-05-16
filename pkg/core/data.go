package core

import (
	"bytes"
	"encoding/binary"

	"gitlab.com/alephledger/core-go/pkg/crypto/multi"
	"golang.org/x/crypto/sha3"
)

// Data is a packet of binary data that is embedded in a single unit.
type Data []byte

// DataSource is a provider of Data.
type DataSource interface {
	GetData() Data
}

// Preblock is a set of Data from units contained in one timing round (block)
// together with some RandomBytes produced by common randomness.
type Preblock struct {
	Data        []Data
	RandomBytes []byte
}

// NewPreblock constructs a preblock from given data and randomBytes.
func NewPreblock(data []Data, randomBytes []byte) *Preblock {
	return &Preblock{data, randomBytes}
}

// PreblockSink is a channel on which preblocks are written.
type PreblockSink chan<- *Preblock

// PreblockSource is a channel providing new preblocks.
type PreblockSource <-chan *Preblock

// Block is a preblock that has been processed and signed by committee members.
// It is the final building block of the blockchain produced by the protocol.
type Block struct {
	Preblock
	ID             uint64
	AdditionalData []Data
	Signature      *multi.Signature
}

// BlockHash computes the hash of a given block.
// For obvious reasons this does not include the signature.
func BlockHash(b *Block) []byte {
	var data bytes.Buffer
	idBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(idBytes, b.ID)
	data.Write(idBytes)
	for _, d := range b.Preblock.Data {
		data.Write(d)
	}
	data.Write(b.Preblock.RandomBytes)
	for _, d := range b.AdditionalData {
		data.Write(d)
	}
	result := make([]byte, 32)
	sha3.ShakeSum128(result, data.Bytes())
	return result
}

// BlockSource is a source of blocks.
type BlockSource <-chan *Block

// BlockSink is an output channel for the blockchain produced.
type BlockSink chan<- *Block

// ToBlock creates a block from a given preblock and additional data.
func ToBlock(pb *Preblock, id uint64, additionalData []Data) *Block {
	return &Block{
		Preblock:       *pb,
		ID:             id,
		AdditionalData: additionalData,
	}
}
