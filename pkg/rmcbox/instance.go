package rmcbox

import (
	"bytes"
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

func (ins *instance) SendData(w io.Writer) error {
	ins.Lock()
	defer ins.Unlock()
	err := encodeUint32(w, ins.rawLen)
	if err != nil {
		return err
	}
	_, err = w.Write(ins.signedData)
	return err
}

func (ins *instance) SendProof(w io.Writer) error {
	ins.Lock()
	defer ins.Unlock()
	if ins.stat != Finished {
		return errors.New("no proof to send")
	}
	_, err := w.Write(ins.proof.Marshal())
	return err
}

func (ins *instance) SendFinished(w io.Writer) error {
	err := ins.SendData(w)
	if err != nil {
		return err
	}
	return ins.SendProof(w)
}

func (ins *instance) AcceptSignature(pid uint16, r io.Reader) (bool, error) {
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

func (ins *instance) SendSignature(w io.Writer) error {
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

func (ins *instance) AcceptProof(r io.Reader) error {
	ins.Lock()
	defer ins.Unlock()
	if ins.stat == Unknown {
		return errors.New("cannot accept proof of unknown data")
	}
	nProc := uint16(ins.keys.Length())
	proof := multi.NewSignature(crypto.MinimalQuorum(nProc), ins.signedData)
	data := make([]byte, proof.MarshaledLength())
	_, err := io.ReadFull(r, data)
	if err != nil {
		return err
	}
	_, err = proof.Unmarshal(data)
	if err != nil {
		return err
	}
	if !ins.keys.MultiVerify(proof) {
		return errors.New("wrong multisignature")
	}
	if ins.stat != Finished {
		ins.proof = proof
		ins.stat = Finished
	}
	return nil
}

func (ins *instance) Data() []byte {
	if int(ins.rawLen) == len(ins.signedData) {
		return ins.signedData
	}
	return ins.signedData[8 : 8+ins.rawLen]
}

func (ins *instance) Status() Status {
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

func (in *incoming) AcceptData(r io.Reader) ([]byte, error) {
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
	if in.stat == Unknown {
		in.stat = Data
	} else {
		thisData := signedData[8 : 8+rawLen]
		if !bytes.Equal(thisData, in.Data()) {
			return nil, errors.New("different data already accepted")
		}
		return in.Data(), nil
	}
	in.signedData = signedData
	in.rawLen = rawLen
	in.proof = proof
	return in.Data(), nil
}

func (in *incoming) AcceptFinished(r io.Reader) ([]byte, error) {
	result, err := in.AcceptData(r)
	if err != nil {
		return nil, err
	}
	return result, in.AcceptProof(r)
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
