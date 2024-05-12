package app

import (
	"fmt"
	"regexp"
	"strings"
)

type TrieNode struct {
	Children    map[string]*TrieNode
	Regex       *regexp.Regexp
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

	// Split the fromURL into segments
	segments := strings.Split(fromURL, "/")

	// Skip the first empty segment if it exists
	if len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}

	for _, segment := range segments {
		isWildcard := strings.HasPrefix(segment, "(.*)")
		isStart := strings.HasPrefix(segment, "^")
		isEnd := strings.HasSuffix(segment, "$")

		if isWildcard || isStart || isEnd {
			// Compile the regex pattern
			regexPattern := segment

			// Adjust regex pattern for beginning and end anchors
			if isStart && isEnd {
				// If both ^ and $ are present, treat it as an exact match
				regexPattern = strings.TrimSuffix(strings.TrimPrefix(segment, "^"), "$")
			} else if isStart {
				// If ^ is present, treat it as beginning of line
				regexPattern = fmt.Sprintf("^%s", strings.TrimPrefix(segment, "^"))
			} else if isEnd {
				// If $ is present, treat it as end of line
				regexPattern = fmt.Sprintf("%s$", strings.TrimSuffix(segment, "$"))
			}

			regex, err := regexp.Compile(regexPattern)
			if err != nil {
				// Handle regex compilation error
				panic(err)
			}
			node.Regex = regex
		} else {
			if _, ok := node.Children[segment]; !ok {
				node.Children[segment] = &TrieNode{Children: make(map[string]*TrieNode)}
			}
			node = node.Children[segment]
		}
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
		} else if node.Regex != nil {
			// If there's a regex pattern but no match, return false
			if !node.Regex.MatchString(segment) {
				return "", false
			}
			// Capture regex groups
			matches := node.Regex.FindStringSubmatch(segment)
			capturedArgs = append(capturedArgs, matches[1:]...) // Exclude the full match
		} else {
			// If there's no match and no regex pattern, return false
			if node.RedirectURL != "" {
				// If the parent node has a redirect URL, use it
				return node.RedirectURL, true
			}
			return "", false
		}
	}

	// Check if the last node contains a redirect URL
	if node.RedirectURL != "" {
		redirectURL := node.RedirectURL
		for i, capturedArg := range capturedArgs {
			// Replace $1, $2, ... in the redirect URL with capturedArgs[i]
			placeholder := fmt.Sprintf("$%d", i+1)
			redirectURL = strings.Replace(redirectURL, placeholder, capturedArg, -1)
		}
		return redirectURL, true
	}

	return "", false
}

func (t *Trie) Clear() {
	t.Root = &TrieNode{Children: make(map[string]*TrieNode)}
}
