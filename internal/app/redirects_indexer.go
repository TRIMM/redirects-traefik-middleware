package app

import (
	"fmt"
	"regexp"
	"strings"
)

type Rule struct {
	pattern *regexp.Regexp
	target  string
}

type IndexedRedirects struct {
	LengthMap map[int][]*Rule
	PrefixMap map[string][]*Rule
}

func NewIndexedRedirects() *IndexedRedirects {
	return &IndexedRedirects{
		LengthMap: make(map[int][]*Rule),
		PrefixMap: make(map[string][]*Rule),
	}
}

func (idx *IndexedRedirects) IndexRule(pattern, target string) {
	rule := &Rule{
		pattern: regexp.MustCompile(pattern),
		target:  target,
	}
	urlLen := len(strings.Split(rule.pattern.String(), "/"))
	idx.LengthMap[urlLen] = append(idx.LengthMap[urlLen], rule)

	prefix := getPrefix(rule)
	idx.PrefixMap[prefix] = append(idx.PrefixMap[prefix], rule)
}

func getPrefix(rule *Rule) string {
	// Implement logic to extract prefix
	parts := strings.Split(rule.pattern.String(), "/")
	if len(parts) > 1 {
		return parts[1] // returns the first segment as prefix
	}
	return ""
}

func (idx *IndexedRedirects) Match(url string) (string, bool) {
	urlParts := strings.Split(url, "/")
	urlLen := len(urlParts)
	possibleRules := idx.LengthMap[urlLen]

	for _, rule := range possibleRules {
		if matches := rule.pattern.FindStringSubmatch(url); matches != nil {
			redirectURL := rule.target
			for i := 1; i < len(matches); i++ {
				placeholder := fmt.Sprintf("$%d", i)
				redirectURL = strings.ReplaceAll(redirectURL, placeholder, matches[i])
			}
			return redirectURL, true
		}
	}

	return "", false
}
