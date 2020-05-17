package tests

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"strconv"

	"gitlab.com/alephledger/core-go/pkg/core"
)

// NopPreblockConsumer drains the provided source doing nothing with incoming preblocks.
func NopPreblockConsumer(ps core.PreblockSource) {
	for range ps {
	}
}

// CountingPreblockConsumer drains the provided source and counts incoming units.
// When done, writes the result to provided Writer.
func CountingPreblockConsumer(ps core.PreblockSource, w io.Writer) {
	n := 0
	for range ps {
		n++
	}
	msg := "Preblocks consumed: " + strconv.Itoa(n) + "\n"
	w.Write([]byte(msg))
}

// ControlSumPreblockConsumer drains the provided source and calculates the hash of each preblock,
// which depends on Preblock data and previous preblock's hash).
// When done, writes the result to provided Writer.
func ControlSumPreblockConsumer(ps core.PreblockSource, w io.Writer) {
	n := 0
	last := []byte{}
	hash := sha256.New()
	for pb := range ps {
		n++
		hash.Reset()
		hash.Write(last)
		for _, data := range pb.Data {
			hash.Write(data)
		}
		hash.Write(pb.RandomBytes)
		last = hash.Sum(nil)
	}
	msg := "Preblocks consumed: " + strconv.Itoa(n) + "\n"
	msg += "Control sum: " + base64.StdEncoding.EncodeToString(last) + "\n"
	w.Write([]byte(msg))
}

// DataExtractingPreblockConsumer reads preblocks from the source and writes all data contained in them to the provided writer.
func DataExtractingPreblockConsumer(ps core.PreblockSource, w io.Writer) {
	n := 0
	for pb := range ps {
		msg := []byte("Preblock " + strconv.Itoa(n) + "\n")
		w.Write(msg)
		n++
		for _, d := range pb.Data {
			if len(d) > 0 {
				w.Write(d)
				w.Write([]byte("\n"))
			}
		}
	}
}
