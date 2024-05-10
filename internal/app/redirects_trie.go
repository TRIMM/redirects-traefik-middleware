package app

import (
	"fmt"
	"regexp"
	"strings"
)

type TrieNode struct {
	Children    map[string]*TrieNode
	IsWildcard  bool
	isStart     bool
	isEnd       bool
	Regex       string
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
		isWildcard := strings.HasPrefix(segment, "(.*)")
		isStart := strings.HasPrefix(segment, "^")
		isEnd := strings.HasSuffix(segment, "$")

		if isWildcard || isStart || isEnd {
			node.Regex = segment
		} else {
			if _, ok := node.Children[segment]; !ok {
				node.Children[segment] = &TrieNode{Children: make(map[string]*TrieNode)}
			}
			node = node.Children[segment]
		}

		node.IsWildcard = isWildcard
		node.isStart = isStart
		node.isEnd = isEnd
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
		} else {
			// Check if the node represents a regex pattern
			if node.IsWildcard || node.isStart || node.isEnd {
				// Compile the regex pattern
				regex, err := regexp.Compile(node.Regex)
				if err != nil {
					return "", false
				}

				if regex.MatchString(segment) {
					capturedArgs = append(capturedArgs, segment)
				} else {
					return "", false
				}
			} else {
				return "", false
			}
		}
	}

	// Construct the final redirect URL using capturedArgs
	redirectURL := node.RedirectURL
	for i, capturedArg := range capturedArgs {
		// Replace $1, $2, ... in the redirect URL with capturedArgs[i]
		placeholder := fmt.Sprintf("$%d", i+1)
		redirectURL = strings.Replace(redirectURL, placeholder, capturedArg, -1)
	}

	return redirectURL, true
}

func (t *Trie) Clear() {
	t.Root = &TrieNode{Children: make(map[string]*TrieNode)}
}
