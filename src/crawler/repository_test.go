package crawler_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sjain93/web-crawler-go/config"
	"github.com/sjain93/web-crawler-go/src/crawler"
	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	inMemDB := config.GetInMemoryStore()
	preLoad(inMemDB, crawler.Metadata{ID: "d28becaf-afb8-422a-a88f-00759050a965"})

	crawlerRepo, err := crawler.NewCrawlerRepository(inMemDB)
	assert.NoError(t, err)

	testCases := map[string]struct {
		metadata    crawler.Metadata
		expectedErr error
	}{
		"Happy Path - saves record": {
			metadata: crawler.Metadata{
				ID:         uuid.NewString(),
				InitialURL: "https://spacy.io",
				Host:       "spacy.io",
				CrawlResultSet: []string{
					"https://spacy.io/",
					"https://spacy.io/usage/v3-6#pipelines",
				},
			},
			expectedErr: nil,
		},
		"Error - conflict on ID": {
			metadata: crawler.Metadata{
				ID: "d28becaf-afb8-422a-a88f-00759050a965",
			},
			expectedErr: crawler.ErrUniqueKeyViolated,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := crawlerRepo.Save(&tc.metadata)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	inMemDB := config.GetInMemoryStore()
	preLoad(inMemDB, crawler.Metadata{ID: "085eeb21-4737-4b21-a501-680c8dc23e95"})

	crawlerRepo, err := crawler.NewCrawlerRepository(inMemDB)
	assert.NoError(t, err)

	testCases := map[string]struct {
		metadata    crawler.Metadata
		expectedErr error
	}{
		"Success": {
			metadata:    crawler.Metadata{ID: "085eeb21-4737-4b21-a501-680c8dc23e95"},
			expectedErr: nil,
		},
		"Failure": {
			metadata:    crawler.Metadata{ID: uuid.NewString()},
			expectedErr: crawler.ErrRecordNotFound,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := crawlerRepo.GetCrawlByID(&tc.metadata)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestGetByHost(t *testing.T) {
	inMemDB := config.GetInMemoryStore()
	setupMockData(inMemDB)

	crawlerRepo, err := crawler.NewCrawlerRepository(inMemDB)
	assert.NoError(t, err)

	testCases := map[string]struct {
		metadata          crawler.Metadata
		expectedResultLen int
	}{
		"Success - ordered slice": {
			metadata:          crawler.Metadata{Host: "monzo.com"},
			expectedResultLen: 2,
		},
		"Empty Slice": {
			metadata:          crawler.Metadata{Host: "google.com"},
			expectedResultLen: 0,
		},
		"Success Single element": {
			metadata:          crawler.Metadata{Host: "www.koho.ca"},
			expectedResultLen: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			crawls, err := crawlerRepo.GetCrawlsByHost(&tc.metadata)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResultLen, len(crawls))

			if tc.expectedResultLen == 2 {
				assert.True(t, crawls[0].CreatedAt.After(crawls[1].CreatedAt))
			}
		})
	}
}

func setupMockData(db config.MemoryStore) {
	now := time.Now().UTC()
	earlierRun := now.Add(time.Duration(-30) * time.Minute)

	records := []crawler.Metadata{
		{
			ID:         "5eb020a4-54cc-4b57-b19f-cbd33a2df881",
			InitialURL: "https://monzo.com/",
			Host:       "monzo.com",
			CrawlResultSet: []string{
				"http://monzo.com/isa/",
				"https://monzo.com/blog/2023/09/12/the-skadnetwork-puzzle-using-data-to-solve-for-effective-performance/",
			},
			CreatedAt: time.Now(),
		},
		{
			ID:         "d28becaf-afb8-422a-a88f-00759050a965",
			InitialURL: "https://www.koho.ca/",
			Host:       "www.koho.ca",
			CrawlResultSet: []string{
				"https://www.koho.ca/learn/dry-january-mocktail-recipes/",
				"https://www.koho.ca",
			},
		},
		{
			ID:         "085eeb21-4737-4b21-a501-680c8dc23e95",
			InitialURL: "https://monzo.com/",
			Host:       "monzo.com",
			CrawlResultSet: []string{
				"http://monzo.com/isa/",
				"https://monzo.com/blog/2023/09/12/the-skadnetwork-puzzle-using-data-to-solve-for-effective-performance/",
			},
			CreatedAt: earlierRun,
		},
	}
	preLoad(db, records...)
}

func preLoad(db config.MemoryStore, data ...crawler.Metadata) {
	for _, d := range data {
		if d.CreatedAt.IsZero() {
			d.CreatedAt = time.Now().UTC()
		}
		db[d.ID] = d
	}
}
