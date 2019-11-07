package trie

import "github.com/alinz/baker"

type Node struct {
	children map[byte]*Node
	value    *baker.Service
}

var _ Trier = (*Node)(nil)

func (n *Node) Insert(key []byte, value *baker.Service) {
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

func (n *Node) Remove(key []byte) {
	parents := make([]*Node, 0)
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

func (n *Node) Search(key []byte) *baker.Service {
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
	Insert(key []byte, value *baker.Service)
	Remove(key []byte)
	Search(key []byte) *baker.Service
}

func New() *Node {
	return &Node{
		children: make(map[byte]*Node),
	}
}
