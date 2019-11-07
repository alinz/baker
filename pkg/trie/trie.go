package trie

type Err string

func (e Err) Error() string {
	return string(e)
}

const (
	ErrNotFound = Err("not found")
)

type Store interface {
	Insert(k []byte, val interface{})
	Remove(k []byte)
	Search(k []byte) (interface{}, error)
}

const (
	wild byte = '*'
)

type Node struct {
	children map[byte]*Node
	value    interface{}
	hasValue bool
	wild     bool
}

var _ Store = (*Node)(nil)

func (n *Node) Insert(key []byte, val interface{}) {
	ok := false
	curr := n
	next := curr

	for i := 0; i < len(key); i++ {
		b := key[i]

		if b == wild {
			curr.wild = true
			break
		}

		next, ok = curr.children[b]
		if !ok {
			next = New()
			curr.children[b] = next
			curr = next
		}

		curr = next
	}

	curr.hasValue = true
	curr.value = val
}

func (n *Node) Remove(key []byte) {
	parents := make([]*Node, 0)
	keys := make([]byte, 0)

	curr := n

	// create a path for both node and key
	for i := 0; i < len(key); i++ {
		if curr.wild {
			break
		}

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
	if !curr.hasValue {
		return
	}

	// remove the reference to value in the current node
	curr.value = nil
	curr.hasValue = false

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

func (n *Node) Search(key []byte) (interface{}, error) {
	ok := false
	curr := n

	for i := 0; i < len(key); i++ {
		b := key[i]

		if curr.wild {
			break
		}

		curr, ok = curr.children[b]
		if !ok {
			return nil, ErrNotFound
		}
	}

	if !curr.hasValue {
		return nil, ErrNotFound
	}

	return curr.value, nil
}

func New() *Node {
	return &Node{
		children: make(map[byte]*Node),
	}
}
