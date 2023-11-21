package instance_test

import (
	"testing"

	"github.com/sjain93/web-crawler-go/src/crawler/instance"
	"github.com/sjain93/web-crawler-go/src/util"
	"github.com/stretchr/testify/assert"
)

// This is a long test and takes roughly 90 seconds locally
func TestScraper(t *testing.T) {
	testCases := map[string]struct {
		config     instance.Config
		initialURL string
		len        int
	}{
		"Successful crawl (LONG TEST)": {
			config:     *instance.NewDefaultConfig(),
			initialURL: "https://www.monzo.com",
			len:        0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			c, err := instance.NewCrawler(tc.initialURL, tc.config)
			assert.NoError(t, err)

			// start the crawl
			c.Process()

			// retrieve visited links from object
			visitedLinks := c.GetLinks()
			assert.Greater(t, len(visitedLinks), tc.len)

			mainHost, err := util.GetHost(tc.initialURL)
			assert.NoError(t, err)

			for _, link := range visitedLinks {
				linkHost, err := util.GetHost(link)
				assert.NoError(t, err)

				assert.True(t, mainHost == linkHost)
			}
		})
	}
}

func TestExtractAndDispatch(t *testing.T) {
	testcases := map[string]struct {
		url             string
		expectedResults []string
	}{
		"Happy Path - http": {
			url: "https://www.york.ac.uk/teaching/cws/wws/webpage3.html",
			expectedResults: []string{
				"https://www.york.ac.uk/teaching/cws/wws/webpage1.html",
				"https://www.york.ac.uk/teaching/cws/wws/webpage4.html",
				"https://www.york.ac.uk/teaching/cws/wws/webpage2.html",
				"https://www.york.ac.uk/teaching/cws/wws/col3.html",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg := instance.NewDefaultConfig()
			c, err := instance.NewCrawler(tc.url, *cfg)
			assert.NoError(t, err)

			res, err := cfg.HttpClient.Get(tc.url)
			assert.NoError(t, err)

			c.ExtractTestCall(res, tc.url)
			errlist := c.GetErrors()
			assert.Zero(t, len(errlist))
			links := c.GetLinks()
			assert.Equal(t, len(links), len(tc.expectedResults))
			assert.True(t, unorderedEqual(links, tc.expectedResults))
		})
	}
}

func unorderedEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	exists := make(map[string]bool)
	for _, value := range first {
		exists[value] = true
	}
	for _, value := range second {
		if !exists[value] {
			return false
		}
	}
	return true
}
