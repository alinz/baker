package gateway

type node struct {
	children map[byte]*node
	value    *Services
}

var _ Trier = (*node)(nil)

func (n *node) Insert(key []byte, value *Services) {
	ok := false
	curr := n
	next := curr

	for i := 0; i < len(key); i++ {
		next, ok = curr.children[key[i]]
		if !ok {
			next = New()
			curr.children[key[i]] = next
			curr = next
		}
		curr = next
	}

	curr.value = value
}

func (n *node) Remove(key []byte) {
	parents := make([]*node, 0)
	keys := make([]byte, 0)

	curr := n

	// create a path for both node and key
	for i := 0; i < len(key); i++ {
		k := key[i]
		m, ok := curr.children[k]
		if !ok {
			return
		}

		parents = append(parents, curr)
		keys = append(keys, k)
		curr = m
	}

	// not found a node
	// ignore deletion
	if curr.value == nil {
		return
	}

	// remove the reference to value in the current node
	curr.value = nil

	// need to traverse back and delete any tracing of node
	// if there
	for i := len(parents) - 1; i > 0; i-- {
		parent := parents[i]

		child := parent.children[keys[i]]
		if len(child.children) != 0 {
			return
		}

		delete(parent.children, keys[i])
	}
}

func (n *node) Search(key []byte) *Services {
	ok := false
	curr := n

	for i := 0; i < len(key); i++ {
		curr, ok = curr.children[key[i]]
		if !ok {
			return nil
		}
	}

	return curr.value
}

type Trier interface {
	Insert(key []byte, value *Services)
	Remove(key []byte)
	Search(key []byte) *Services
}

func New() *node {
	return &node{
		children: make(map[byte]*node),
	}
}
