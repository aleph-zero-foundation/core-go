package gonsai

import (
	"sort"
)

type proof struct {
	name     []byte
	siblings []*info
	next     *proof
	dataHash hash
}

func proofOfSuccess(nd *node, name []byte, hash hash) *proof {
	return &proof{
		name:     name,
		siblings: sortedInfo(nd.children, name),
		dataHash: hash,
	}
}

func proofOfFailure(nd *node, name []byte) *proof {
	return &proof{
		name:     name,
		siblings: sortedInfo(nd.children, nil),
	}
}

func extendProof(pr *proof, nd *node) *proof {
	return &proof{
		name:     nd.name,
		siblings: sortedInfo(nd.children, pr.name),
		next:     pr,
	}
}

func sortedInfo(nodes map[byte]*node, exclude []byte) []*info {
	infos := make([]*info, 0, len(nodes))
	for k, v := range nodes {
		if exclude != nil && k == exclude[0] {
			continue
		}
		infos = append(infos, &info{v.name, v.hash})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].name[0] < infos[j].name[0] })
	return infos
}

// parse(proof) == (dataHash, key, root) means:
// 1) if dataHash == nil: root is hash of a tree in which key is not present
// 2) if dataHash != nil: root is hash of a tree in which key is present and stores a value with dataHash
func parse(pr *proof) (hash, []byte, hash) {
	var dh, root hash
	var name []byte
	if pr.next == nil { // tail
		if pr.dataHash == nil { // proof of failure
			return nil, pr.name, hashNodes(pr.siblings)
		}
		dh = pr.dataHash
		root = pr.dataHash
	} else {
		dh, name, root = parse(pr.next)
	}
	n := len(pr.siblings)
	infos := make([]*info, n+1)
	copy(infos, pr.siblings)
	for n > 0 {
		if infos[n-1].name[0] > name[0] {
			infos[n] = infos[n-1]
			n--
		} else {
			infos[n] = &info{name, root}
			break
		}
	}
	newRoot := hashNodes(infos)
	newName := append(append([]byte{}, pr.name...), name...)
	return dh, newName, newRoot
}
