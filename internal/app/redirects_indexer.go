package app

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type Rule struct {
	pattern    *regexp.Regexp
	target     string
	fromDomain *regexp.Regexp
	isDomain   bool
}

type IndexedRedirects struct {
	LengthMap   map[int]map[string][]*Rule
	DomainRules []*Rule
	mu          sync.RWMutex
}

func NewIndexedRedirects() *IndexedRedirects {
	return &IndexedRedirects{
		LengthMap:   make(map[int]map[string][]*Rule),
		DomainRules: []*Rule{},
	}
}

func (idx *IndexedRedirects) IndexRule(pattern, fromDomain, target string) {
	rule := &Rule{
		pattern:    regexp.MustCompile(pattern),
		target:     target,
		fromDomain: regexp.MustCompile(fromDomain),
		isDomain:   fromDomain != "",
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	if rule.isDomain {
		idx.DomainRules = append(idx.DomainRules, rule)
	} else {
		length := len(strings.Split(pattern, "/"))
		if _, ok := idx.LengthMap[length]; !ok {
			idx.LengthMap[length] = make(map[string][]*Rule)
		}
		prefix := getPrefix(pattern)
		idx.LengthMap[length][prefix] = append(idx.LengthMap[length][prefix], rule)
	}
}

// Match matches the incoming requests against the redirect rules
func (idx *IndexedRedirects) Match(url string) (string, bool) {
	isFullURL := strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if isFullURL {
		for _, rule := range idx.DomainRules {
			if matches := rule.fromDomain.FindStringSubmatch(url); matches != nil {
				redirectURL := rule.target
				for i := 1; i < len(matches); i++ {
					placeholder := fmt.Sprintf("$%d", i)
					redirectURL = strings.ReplaceAll(redirectURL, placeholder, matches[i])
				}
				return redirectURL, true
			}
		}
	}

	urlParts := strings.Split(url, "/")
	length := len(urlParts)
	prefix := urlParts[1]

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

func (idx *IndexedRedirects) Update(pattern, fromDomain, target string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if fromDomain != "" {
		for _, rule := range idx.DomainRules {
			if rule.fromDomain.String() == fromDomain {
				rule.target = target
				break
			}
		}
	} else {
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
}

func (idx *IndexedRedirects) Delete(pattern, fromDomain string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if fromDomain != "" {
		index := -1
		for i, rule := range idx.DomainRules {
			if rule.fromDomain.String() == fromDomain {
				index = i
				break
			}
		}
		if index != -1 {
			idx.DomainRules = append(idx.DomainRules[:index], idx.DomainRules[index+1:]...)
		}
	} else {
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

		if index != -1 {
			idx.LengthMap[length][prefix] = append(rulesSlice[:index], rulesSlice[index+1:]...)
		}
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
