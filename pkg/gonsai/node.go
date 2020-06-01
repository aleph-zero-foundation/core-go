package gonsai

type node struct {
	name     []byte
	hash     hash
	data     []byte
	children map[byte]*node
	// stage zone
	stName     []byte
	stHash     hash // stage hash is not updated after every change, needs calling calcStHash() to be calculated
	stData     []byte
	stChildren map[byte]*node // stores only changes wrt children
	stage      bool           // indicates if there are any nodes with non-empty stage in the whole subtree
	// Invariants:
	// inv0: data != nil <=> node is a leaf
	// inv1: in children and stChildren name of the node stored under X starts with X
	// inv2: every non-leaf non-root node has at least 2 children (path compression)
	// inv3: stHash != nil <=> no changes in stage of the whole subtree since last calcStHash()
}

func newStagedNode(name []byte) *node {
	return &node{
		stName:     name,
		children:   make(map[byte]*node),
		stChildren: make(map[byte]*node),
		stage:      true,
	}
}

func newStagedLeaf(name, data []byte) *node {
	return &node{
		stName: name,
		data:   data, // inv0
		stData: data,
		stage:  true,
	}
}

func (nd *node) getChild(char byte) *node {
	if ch, ok := nd.stChildren[char]; ok {
		return ch
	}
	return nd.children[char]
}

func (nd *node) getName() []byte {
	if nd.stName != nil {
		return nd.stName
	}
	return nd.name
}

func (nd *node) getData() []byte {
	if nd.stData != nil {
		return nd.stData
	}
	return nd.data
}

// mark flags the node as having some recent changes in stage (in the whole subtree)
func (nd *node) mark() {
	nd.stage = true
	nd.stHash = nil
}

// Find ignores any staged values
func (nd *node) Find(key []byte) ([]byte, *proof) {
	if ch, ok := nd.children[key[0]]; ok {
		cp := commonPrefix(key, ch.name) // inv1 => cp>0
		if cp == len(key) {
			return ch.data, proofOfSuccess(nd, ch.name, ch.hash)
		}
		if cp == len(ch.name) {
			res, pr := ch.Find(key[cp:])
			return res, extendProof(pr, nd)
		}
		return nil, proofOfFailure(nd, key)
	}
	return nil, proofOfFailure(nd, key)
}

func (nd *node) StageUpdate(key []byte, upd updater) bool {
	if ch := nd.getChild(key[0]); ch != nil {
		name := ch.getName()
		cp := commonPrefix(key, name) // inv1 => cp>0
		if cp == len(key) {
			// we reached desired leaf, update data
			data := ch.getData()
			newStData := upd(data)
			if sameData(data, newStData) {
				return false
			}
			ch.stData = newStData
			ch.mark()
			nd.mark()
			return true
		}
		if cp == len(name) {
			// name is a prefix of key, go into subtree
			changed := ch.StageUpdate(key[cp:], upd)
			if changed {
				nd.mark()
			}
			return changed
		}
		return false // key not present
	}
	return false // key not present
}

func (nd *node) StageInsert(key, value []byte) bool {
	if ch := nd.getChild(key[0]); ch != nil {
		name := ch.getName()
		cp := commonPrefix(key, name) // inv1 => cp>0
		if cp == len(key) {
			// we reached a leaf, turns out key is already present
			if sameData(value, ch.getData()) {
				return false
			}
			ch.stData = value
			ch.mark()
			nd.mark()
			return true
		}
		if cp == len(name) {
			// name is a prefix of key, go into subtree
			changed := ch.StageInsert(key[cp:], value)
			if changed {
				nd.mark()
			}
			return changed
		}
		// we need to create a new branching point
		fresh := newStagedNode(key[:cp])
		fresh.stChildren[key[cp]] = newStagedLeaf(key[cp:], value)
		fresh.stChildren[name[cp]] = ch
		ch.stName = name[cp:]
		nd.stChildren[key[0]] = fresh
		nd.mark()
		return true
	}
	nd.stChildren[key[0]] = newStagedLeaf(key, value)
	nd.mark()
	return true
}

// If nd has one future child after this operation, that child is returned as second value, otherwise nil
func (nd *node) StageDelete(key []byte) (bool, *node) {
	if ch := nd.getChild(key[0]); ch != nil {
		name := ch.getName()
		cp := commonPrefix(key, name) // inv1 => cp>0
		if cp == len(key) {
			// we reached the leaf we want to delete
			nd.stChildren[key[0]] = nil
			ch.mark()
			nd.mark()
			return true, nd.checkChildren()
		}
		if cp == len(name) {
			// name is a prefix of key, go into subtree
			changed, onlyChild := ch.StageDelete(key[cp:])
			if onlyChild != nil { // keep inv2
				nd.stChildren[key[0]] = onlyChild
				nd.mark()
				onlyChild.stName = append(append([]byte{}, name...), onlyChild.getName()...)
			}
			if changed {
				nd.mark()
			}
			return changed, nil // onlyChild != nil can happen only once in the recursion stack, if we're here it's already taken care of
		}
		return false, nil // key not present
	}
	return false, nil // key not present
}

// checkChildren checks if the staged version of this node will have only one child. If yes, returns that child.
func (nd *node) checkChildren() *node {
	count := len(nd.children)
	for k, v := range nd.stChildren {
		_, ok := nd.children[k]
		if v == nil && ok { // deletion of existing child
			count--
		}
		if v != nil && !ok { // insertion of a new child
			count++
		}
	}
	if count == 1 {
		// nd has only one future child left, find it
		for _, v := range nd.stChildren {
			if v == nil {
				continue
			}
			return v
		}
		for k, v := range nd.children {
			if _, ok := nd.stChildren[k]; !ok {
				return v
			}
		}
	}
	return nil
}

func (nd *node) Reset() {
	if nd.stage {
		for _, v := range nd.children {
			v.Reset()
		}
		nd.stChildren = make(map[byte]*node)
		nd.stName, nd.stHash, nd.stData = nil, nil, nil
		nd.stage = false
	}
}

func (nd *node) Commit() {
	if nd.stage && nd.stHash != nil {
		nd.hash = nd.stHash
		nd.stHash = nil
		for k, v := range nd.stChildren {
			if v != nil {
				nd.children[k] = v
			} else {
				delete(nd.children, k)
			}
		}
		for _, ch := range nd.children {
			if ch.stName != nil {
				ch.name = ch.stName
				ch.stName = nil
			}
			ch.Commit()
		}
		if nd.stData != nil {
			nd.data = nd.stData
			nd.stData = nil
		}
		nd.stChildren = make(map[byte]*node)
		nd.stage = false
	}
}

func (nd *node) calcStHash() hash {
	if !nd.stage { // nothing staged in the whole subtree, hash will not change
		return nd.hash
	}

	if nd.stHash != nil { // no changes since last calcStHash()
		return nd.stHash
	}

	if nd.data != nil { // leaf
		nd.stHash = hashData(nd.stData)
		return nd.stHash
	}

	infos := make([]*info, 0)
	for i := 0; i < 256; i++ {
		if ch, ok := nd.stChildren[byte(i)]; ok {
			if ch == nil {
				continue
			}
			infos = append(infos, &info{ch.getName(), ch.calcStHash()})
			continue
		}
		if ch, ok := nd.children[byte(i)]; ok {
			infos = append(infos, &info{ch.getName(), ch.calcStHash()})
		}
	}
	nd.stHash = hashNodes(infos)
	return nd.stHash
}
