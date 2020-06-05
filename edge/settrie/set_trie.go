package settrie

import (
	"github.com/beevik/prefixtree"
	"sort"
	"strings"
)

type SetTrie struct {
	trie *prefixtree.Tree
}

func New() *SetTrie {
	return &SetTrie{prefixtree.New()}
}

func (t *SetTrie) Add(s []string, data interface{}) {
	sort.Strings(s)
	t.trie.Add(strings.Join(s, ""), data)
}

func (t *SetTrie) Find(s []string) (interface{}, error){
	sort.Strings(s)
	return t.trie.Find(strings.Join(s, ""))
}

func (t *SetTrie) Output() {
	t.trie.Output()
}

func (t *SetTrie) Delete(s []string) {
	// TODO implement delete functionality
}
