package rmc

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"gitlab.com/alephledger/core-go/pkg/crypto"
	"gitlab.com/alephledger/core-go/pkg/crypto/multi"
)

type instance struct {
	sync.Mutex
	id         uint64
	keys       *multi.Keychain
	rawLen     uint32
	signedData []byte
	proof      *multi.Signature
	stat       Status
}

func (ins *instance) sendData(w io.Writer) error {
	ins.Lock()
	defer ins.Unlock()
	err := encodeUint32(w, ins.rawLen)
	if err != nil {
		return err
	}
	_, err = w.Write(ins.signedData)
	return err
}

func (ins *instance) sendProof(w io.Writer) error {
	ins.Lock()
	defer ins.Unlock()
	if ins.stat != Finished {
		return errors.New("no proof to send")
	}
	_, err := w.Write(ins.proof.Marshal())
	return err
}

func (ins *instance) sendFinished(w io.Writer) error {
	err := ins.sendData(w)
	if err != nil {
		return err
	}
	return ins.sendProof(w)
}

func (ins *instance) acceptSignature(pid uint16, r io.Reader) (bool, error) {
	signature := make([]byte, multi.SignatureLength)
	_, err := io.ReadFull(r, signature)
	ins.Lock()
	defer ins.Unlock()
	if err != nil {
		return false, err
	}
	if !ins.keys.Verify(pid, append(ins.signedData, signature...)) {
		return false, errors.New("wrong signature")
	}
	if ins.stat != Finished {
		done, err := ins.proof.Aggregate(pid, signature)
		if done {
			ins.stat = Finished
			return true, err
		}
		return false, err
	}
	return false, nil
}

func (ins *instance) sendSignature(w io.Writer) error {
	ins.Lock()
	defer ins.Unlock()
	if ins.stat == Unknown {
		return errors.New("cannot sign unknown data")
	}
	signature := ins.keys.Sign(ins.signedData)
	_, err := w.Write(signature)
	if err != nil {
		return err
	}
	if ins.stat == Data {
		ins.stat = Signed
	}
	return nil
}

func (ins *instance) acceptProof(r io.Reader) error {
	ins.Lock()
	defer ins.Unlock()
	if ins.stat == Unknown {
		return errors.New("cannot accept proof of unknown data")
	}
	data := make([]byte, ins.proof.MarshaledLength())
	_, err := io.ReadFull(r, data)
	if err != nil {
		return err
	}
	_, err = ins.proof.Unmarshal(data)
	if err != nil {
		return err
	}
	if !ins.keys.MultiVerify(ins.proof) {
		return errors.New("wrong multisignature")
	}
	ins.stat = Finished
	return nil
}

func (ins *instance) data() []byte {
	if int(ins.rawLen) == len(ins.signedData) {
		return ins.signedData
	}
	return ins.signedData[8 : 8+ins.rawLen]
}

func (ins *instance) status() Status {
	ins.Lock()
	defer ins.Unlock()
	return ins.stat
}

type incoming struct {
	instance
	pid uint16
}

func newIncoming(id uint64, pid uint16, keys *multi.Keychain) *incoming {
	return &incoming{
		instance{
			id:   id,
			keys: keys,
		},
		pid,
	}
}

func (in *incoming) acceptData(r io.Reader) ([]byte, error) {
	rawLen, err := decodeUint32(r)
	if err != nil {
		return nil, err
	}
	signedData := make([]byte, 8+rawLen+multi.SignatureLength)
	_, err = io.ReadFull(r, signedData)
	if err != nil {
		return nil, err
	}
	id := binary.LittleEndian.Uint64(signedData[:8])
	if id != in.id {
		return nil, errors.New("incoming id mismatch")
	}
	if !in.keys.Verify(in.pid, signedData) {
		return nil, errors.New("wrong data signature")
	}
	nProc := uint16(in.keys.Length())
	proof := multi.NewSignature(crypto.MinimalQuorum(nProc), signedData)
	in.Lock()
	defer in.Unlock()
	in.signedData = signedData
	in.rawLen = rawLen
	in.proof = proof
	if in.stat == Unknown {
		in.stat = Data
	}
	return in.data(), nil
}

func (in *incoming) acceptFinished(r io.Reader) ([]byte, error) {
	result, err := in.acceptData(r)
	if err != nil {
		return nil, err
	}
	return result, in.acceptProof(r)
}

func newOutgoing(id uint64, data []byte, keys *multi.Keychain) *instance {
	rawLen := uint32(len(data))
	buf := make([]byte, 8, 8+rawLen)
	binary.LittleEndian.PutUint64(buf, id)
	buf = append(buf[:8], data...)
	signedData := append(buf, keys.Sign(buf)...)
	nProc := uint16(keys.Length())
	proof := multi.NewSignature(crypto.MinimalQuorum(nProc), signedData)
	proof.Aggregate(keys.Pid(), keys.Sign(signedData))
	return &instance{
		id:         id,
		keys:       keys,
		rawLen:     rawLen,
		signedData: signedData,
		proof:      proof,
		stat:       Data,
	}
}

func newRaw(id uint64, data []byte, keys *multi.Keychain) *instance {
	rawLen := uint32(len(data))
	nProc := uint16(keys.Length())
	proof := multi.NewSignature(crypto.MinimalQuorum(nProc), data)
	proof.Aggregate(keys.Pid(), keys.Sign(data))
	return &instance{
		id:         id,
		keys:       keys,
		rawLen:     rawLen,
		signedData: data,
		proof:      proof,
		stat:       Data,
	}
}

func encodeUint32(w io.Writer, i uint32) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, i)
	_, err := w.Write(buf)
	return err
}

func decodeUint32(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf), nil
}
