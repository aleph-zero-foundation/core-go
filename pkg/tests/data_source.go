package tests

import (
    "bufio"
    "crypto/rand"
    "os"

    "gitlab.com/alephledger/core-go/pkg/core"
)

// EmptyDataSource returns a test data source producing empty data.
func EmptyDataSource() core.DataSource {
    return &emptyDS{}
}

// RandomDataSource returns a test data source producing random slices of data of a given size.
func RandomDataSource(size int) core.DataSource {
    return &randomDS{size}
}

// StdinDataSource returns a test data source producing slices of data taken from the standard input.
func StdinDataSource() core.DataSource {
    return &stdinDS{bufio.NewScanner(os.Stdin)}
}

type emptyDS struct{}

func (*emptyDS) GetData() core.Data { return []byte{} }

type randomDS struct {
    size int
}

func (ds *randomDS) GetData() core.Data {
    data := make([]byte, ds.size)
    rand.Read(data)
    return data
}

type stdinDS struct {
    sc *bufio.Scanner
}

func (ds *stdinDS) GetData() core.Data {
    if ds.sc.Scan() {
        return ds.sc.Bytes()
    }
    return []byte{}
}
