package app

import "testing"

func TestIndexedRedirects_Index_And_Match(t *testing.T) {
	idx := NewIndexedRedirects()
	idx.IndexRule("/school/assignments", "", "/school/items")
	idx.IndexRule("", "old-domain.com$", "https://new-domain.com/welcome")
	idx.IndexRule("/home/company/careers/(.*)", "", "/careers/$1")
	idx.IndexRule("", "example.com/(.*)", "https://new-example.com/$1")

	testCases := []struct {
		name             string
		request          string
		expectedRedirect string
	}{
		{
			name:             "Exact domain match redirect",
			request:          "https://old-domain.com",
			expectedRedirect: "https://new-domain.com/welcome",
		},
		{
			name:             "Exact relative path match redirect",
			request:          "/school/assignments",
			expectedRedirect: "/school/items",
		},
		{
			name:             "Captured group relative path redirect",
			request:          "/home/company/careers/software-engineer-hengelo",
			expectedRedirect: "/careers/software-engineer-hengelo",
		},
		{
			name:             "Captured group domain redirect",
			request:          "https://example.com/hello",
			expectedRedirect: "https://new-example.com/hello",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			answer, isMatch := idx.Match(testCase.request)
			if !isMatch {
				t.Errorf("answer is not match: got %v want %v", answer, testCase.expectedRedirect)
			}
		})
	}

}
