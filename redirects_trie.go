package main

import (
	"strings"
)

type TrieNode struct {
	Children    map[string]*TrieNode
	IsWildcard  bool
	RedirectURL string
}

type Trie struct {
	Root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{Root: &TrieNode{Children: make(map[string]*TrieNode)}}
}

// Insert inserts a redirect rule into the Trie
func (t *Trie) Insert(fromURL, toURL string) {
	node := t.Root
	segments := strings.Split(fromURL, "/")

	for _, segment := range segments {
		// Check if the segment is a wildcard
		isWildcard := strings.HasPrefix(segment, "(.*)")
		if isWildcard {
			segment = strings.TrimPrefix(segment, "(.*)")
		}

		if _, ok := node.Children[segment]; !ok {
			node.Children[segment] = &TrieNode{Children: make(map[string]*TrieNode)}
		}

		node = node.Children[segment]
		node.IsWildcard = isWildcard
	}

	node.RedirectURL = toURL
}

// Match matches a request URL against the redirect rules in the Trie
func (t *Trie) Match(requestURL string) (string, bool) {
	node := t.Root
	segments := strings.Split(requestURL, "/")
	var capturedArgs []string

	for _, segment := range segments {
		if child, ok := node.Children[segment]; ok {
			node = child
		} else if wildcardNode, wildcardExists := node.Children[""]; wildcardExists && wildcardNode.IsWildcard {
			node = wildcardNode
			// Capture the segment if it's a wildcard
			capturedArgs = append(capturedArgs, segment)
		} else {
			return "", false
		}
	}

	// Construct the final redirect URL
	redirectURL := node.RedirectURL
	for _, capturedArg := range capturedArgs {
		redirectURL = strings.Replace(redirectURL, "$1", capturedArg, 1)
	}

	return redirectURL, true
}

func (t *Trie) Clear() {
	t.Root = &TrieNode{Children: make(map[string]*TrieNode)}
}
