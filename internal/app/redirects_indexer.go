package app

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type Rule struct {
	pattern *regexp.Regexp
	target  string
}

type IndexedRedirects struct {
	LengthMap map[int]map[string][]*Rule
	mu        sync.RWMutex
}

func NewIndexedRedirects() *IndexedRedirects {
	return &IndexedRedirects{
		LengthMap: make(map[int]map[string][]*Rule),
	}
}

/*
IndexRule creates an index of segments length
It narrows the indexing to pattern prefix for a quick lookup
*/
func (idx *IndexedRedirects) IndexRule(pattern, target string) {
	rule := &Rule{
		pattern: regexp.MustCompile(pattern),
		target:  target,
	}
	prefix := getPrefix(pattern)

	idx.mu.Lock()
	defer idx.mu.Unlock()

	length := len(strings.Split(pattern, "/"))
	if _, ok := idx.LengthMap[length]; !ok {
		idx.LengthMap[length] = make(map[string][]*Rule)
	}
	idx.LengthMap[length][prefix] = append(idx.LengthMap[length][prefix], rule)
}

// Match matches the incoming requests against the redirect rules
func (idx *IndexedRedirects) Match(url string) (string, bool) {
	urlParts := strings.Split(url, "/")
	length := len(urlParts)
	prefix := urlParts[1]

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Checks for the length index and provides prefixes
	if prefixes, ok := idx.LengthMap[length]; ok {
		// Checks for the rules under the prefixes that match the url prefix
		if rules, ok := prefixes[prefix]; ok {
			for _, rule := range rules {
				// Matches the url against the pattern
				if matches := rule.pattern.FindStringSubmatch(url); matches != nil {
					redirectURL := rule.target
					for i := 1; i < len(matches); i++ {
						placeholder := fmt.Sprintf("$%d", i)
						// Construct the redirectURL by filling the captured groups
						redirectURL = strings.ReplaceAll(redirectURL, placeholder, matches[i])
					}
					return redirectURL, true
				}
			}
		}
	}

	return "", false
}

func (idx *IndexedRedirects) Update(pattern, target string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	prefix := getPrefix(pattern)
	length := len(strings.Split(pattern, "/"))
	rulesSlice := idx.LengthMap[length][prefix]

	for _, rule := range rulesSlice {
		if rule.pattern.String() == pattern {
			rule.target = target
			break
		}
	}
}

func (idx *IndexedRedirects) Delete(pattern string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	prefix := getPrefix(pattern)
	length := len(strings.Split(pattern, "/"))
	rulesSlice := idx.LengthMap[length][prefix]

	index := -1
	for i, rule := range rulesSlice {
		if rule.pattern.String() == pattern {
			index = i
			break
		}
	}

	// If the rule is found, remove it from the slice
	if index != -1 {
		idx.LengthMap[length][prefix] = append(rulesSlice[:index], rulesSlice[index+1:]...)
	}
}

// Extract the prefix from the pattern
func getPrefix(pattern string) string {
	prefix := ""
	// Check if it's not the root URL pattern
	if pattern != "^/$" {
		patternParts := strings.Split(pattern, "/")
		if len(patternParts) > 1 {
			prefix = patternParts[1]
		}
	}
	return prefix
}
