package util_test

import (
	"testing"

	"github.com/sjain93/web-crawler-go/src/util"
	"github.com/stretchr/testify/assert"
)

func TestURLScheme(t *testing.T) {
	testCases := map[string]struct {
		url            string
		expectedResult bool
	}{
		"Happy Path - http": {
			url:            "http://youtube.com/",
			expectedResult: true,
		},
		"Happy Path - https": {
			url:            "https://monzo.com/",
			expectedResult: true,
		},
		"mailto scheme - failure": {
			url:            "mailto://help@monzo.com",
			expectedResult: false,
		},
		"empty string - failure": {
			url:            "",
			expectedResult: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			res := util.IsHTTPScheme(tc.url)
			assert.Equal(t, tc.expectedResult, res)
		})
	}
}

func TestGetHost(t *testing.T) {
	testCases := map[string]struct {
		rawURL    string
		wantError bool
	}{
		"Invalid Host": {
			rawURL:    "/s",
			wantError: true,
		},
		"Invalid Host - url typo": {
			rawURL:    "ww.monzo.com",
			wantError: true,
		},
		"Valid Host": {
			rawURL:    "https://www.monzo.com",
			wantError: false,
		},
	}

	for _, tc := range testCases {

		res, err := util.GetHost(tc.rawURL)
		if tc.wantError {
			assert.Error(t, err)
			assert.Equal(t, "", res)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestGetAbsoluteURL(t *testing.T) {

	testCases := map[string]struct {
		current  string
		root     string
		expected string
	}{
		"Should resolve - reference 1": {
			current:  "/monzo-premium/",
			root:     "https://monzo.com/#mainContent",
			expected: "https://monzo.com/monzo-premium/",
		},
		"Should resolve - reference 2": {
			current:  "/features/16-plus/",
			root:     "https://monzo.com/blog/",
			expected: "https://monzo.com/features/16-plus/",
		},
		"Should not resolve - already absolute": {
			current:  "https://www.gov.uk/set-up-business",
			root:     "https://monzo.com/",
			expected: "https://www.gov.uk/set-up-business",
		},
		"Should not resolve - invalid scheme": {
			current:  ":xyz",
			root:     "https://monzo.com/us/",
			expected: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			res, err := util.GetAbsoluteURL(tc.current, tc.root)
			if tc.expected == "" {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.expected, res)
			}
		})
	}

}

func TestIsHTTPScheme(t *testing.T) {

	testCases := map[string]struct {
		link     string
		expected bool
	}{
		"Pass - http": {
			link:     "http://www.umbc.edu/cwit/",
			expected: true,
		},
		"Pass - https": {
			link:     "https://www.gov.uk/set-up-",
			expected: true,
		},
		"Fail - referntial url": {
			link:     "webpage1.html",
			expected: false,
		},
		"Fail - mailto scheme": {
			link:     "mailto:eakn1@york.ac.uk",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := util.IsHTTPScheme(tc.link)
			assert.Equal(t, tc.expected, result)
		})
	}

}

func TestIsSameDomain(t *testing.T) {
	testCases := map[string]struct {
		current  string
		root     string
		expected bool
	}{
		"Assumption that reference links are derived from previous page": {
			current:  "/monzo-premium/",
			root:     "https://monzo.com/#mainContent",
			expected: true,
		},
		"Different domains": {
			current:  "https://www.gov.uk/set-up-business",
			root:     "https://monzo.com/",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := util.IsSameDomain(tc.current, tc.root)
			assert.Equal(t, tc.expected, result)

		})
	}
}
