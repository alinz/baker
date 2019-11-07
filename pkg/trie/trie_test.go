package trie_test

import (
	"testing"

	"github.com/alinz/baker"
	"github.com/alinz/baker/pkg/trie"
)

func TestTrieInsertSearch(t *testing.T) {
	trie := trie.New()

	key1 := []byte("apple")
	key2 := []byte("app")
	key3 := []byte("appl")

	trie.Insert(key1, &baker.Service{})
	trie.Insert(key2, &baker.Service{})

	value := trie.Search(key1)
	if value == nil {
		t.Fatal("value should be there")
	}

	value = trie.Search(key2)
	if value == nil {
		t.Fatal("value should be there")
	}

	value = trie.Search(key3)
	if value != nil {
		t.Fatal("value should not be there")
	}
}

func TestTrieRemove(t *testing.T) {
	trie := trie.New()

	key1 := []byte("apple")
	key2 := []byte("app")

	trie.Insert(key1, &baker.Service{})
	trie.Insert(key2, &baker.Service{})

	trie.Remove(key2)

	value := trie.Search(key2)
	if value != nil {
		t.Fatal("value should not be there")
	}

	value = trie.Search(key1)
	if value == nil {
		t.Fatal("value should  be there")
	}
}

func BenchmarkSearch(b *testing.B) {
	trie := trie.New()

	key1 := []byte("apple")
	key2 := []byte("app")
	key3 := []byte("appl")

	trie.Insert(key1, &baker.Service{})
	trie.Insert(key2, &baker.Service{})

	for i := 0; i < b.N; i++ {
		trie.Search(key3)
	}
}

func BenchmarkTrie(b *testing.B) {
	trie := trie.New()

	key1 := []byte("apple")
	value1 := &baker.Service{}

	for i := 0; i < b.N; i++ {
		trie.Insert(key1, value1)
		trie.Search(key1)
		trie.Remove(key1)
	}
}
