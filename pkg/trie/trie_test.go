package trie_test

import (
	"testing"

	"github.com/alinz/baker/pkg/trie"
)

func TestTrieInsertSearch(t *testing.T) {
	m := trie.New()

	key1 := []byte("apple")
	key2 := []byte("app")
	key3 := []byte("ap")
	key4 := []byte("apple*")

	m.Insert(key1, 1)
	m.Insert(key2, 2)
	m.Insert(key4, 3)

	_, err := m.Search(key1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Search(key2)
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Search(key3)
	if err != trie.ErrNotFound {
		t.Fatalf("should be not found but got %s", err)
	}

	val1, err := m.Search([]byte("apples"))
	if err != nil {
		t.Fatal("should found the wild search")
	}

	val2, err := m.Search([]byte("apples2222"))
	if err != nil {
		t.Fatal("should found the wild search")
	}

	if val1 != val2 {
		t.Fatalf("values should be the same for same wild keys")
	}
}

func TestTrie(t *testing.T) {
	m := trie.New()

	key1 := []byte("/session/*")
	key2 := []byte("/users/*")

	m.Insert(key1, "session path")
	m.Insert(key2, "users path")

	session, err := m.Search([]byte("/session/1"))
	if err != nil {
		t.Fatal("failed to grab /session/1")
	}

	if session != "session path" {
		t.Fatal("not same session object")
	}

	m.Remove(key1)

	_, err = m.Search([]byte("/session/1"))
	if err != trie.ErrNotFound {
		t.Fatal("session should not be presented")
	}
}
