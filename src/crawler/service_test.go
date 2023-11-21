package crawler_test

import (
	"testing"

	"github.com/sjain93/web-crawler-go/config"
	"github.com/sjain93/web-crawler-go/src/crawler"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCrawlSiteWithCache(t *testing.T) {
	logger, err := zap.NewProduction()
	assert.NoError(t, err)

	inMemDB := config.GetInMemoryStore()
	setupMockData(inMemDB)

	crawlerRepo, err := crawler.NewCrawlerRepository(inMemDB)
	assert.NoError(t, err)

	crawlerSvc := crawler.NewCrawlerService(crawlerRepo, logger)

	testCases := map[string]struct {
		initialURL        string
		expectedID        string
		expectedResultLen int
	}{
		"Crawl Site succesfully checks and returns a valid past crawl": {
			initialURL:        "https://monzo.com/",
			expectedID:        "5eb020a4-54cc-4b57-b19f-cbd33a2df881",
			expectedResultLen: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			output, err := crawlerSvc.CrawlSite(crawler.Metadata{InitialURL: tc.initialURL})
			assert.NoError(t, err)
			assert.Len(t, output, tc.expectedResultLen)
			assert.Equal(t, tc.expectedID, output[0].ID)
		})
	}
}
