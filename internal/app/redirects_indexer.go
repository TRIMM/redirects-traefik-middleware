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

func (idx *IndexedRedirects) IndexRule(pattern, target string) {
	rule := &Rule{
		pattern: regexp.MustCompile(pattern),
		target:  target,
	}

	// Extract the prefix from the pattern
	prefix := ""
	if pattern != "^/$" {
		patternParts := strings.Split(pattern, "/")
		if len(patternParts) > 1 {
			prefix = patternParts[1]
		}
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	length := len(strings.Split(pattern, "/"))
	if _, ok := idx.LengthMap[length]; !ok {
		idx.LengthMap[length] = make(map[string][]*Rule)
	}
	idx.LengthMap[length][prefix] = append(idx.LengthMap[length][prefix], rule)
}

func (idx *IndexedRedirects) Match(url string) (string, bool) {
	urlParts := strings.Split(url, "/")
	length := len(urlParts)
	prefix := urlParts[1]

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if prefixes, ok := idx.LengthMap[length]; ok {
		if rules, ok := prefixes[prefix]; ok {
			for _, rule := range rules {
				if matches := rule.pattern.FindStringSubmatch(url); matches != nil {
					redirectURL := rule.target
					for i := 1; i < len(matches); i++ {
						placeholder := fmt.Sprintf("$%d", i)
						redirectURL = strings.ReplaceAll(redirectURL, placeholder, matches[i])
					}
					return redirectURL, true
				}
			}
		}
	}

	return "", false
}
