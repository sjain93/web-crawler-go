package crawler

import (
	"sort"
	"time"

	"github.com/pkg/errors"

	"github.com/sjain93/web-crawler-go/config"
)

var (
	ErrNoDatastore       = errors.New("no datastore provided")
	ErrRecordNotFound    = errors.New("record not found")
	ErrInvalidDataType   = errors.New("invalid user data type")
	ErrUniqueKeyViolated = errors.New("duplicated key not allowed")
)

// Shared model for Service and Repository layer
type Metadata struct {
	ID             string
	InitialURL     string
	Host           string
	CrawlResultSet []string
	ErrList        []error
	CreatedAt      time.Time
}

// Publiv interface for the repository layer, if the datastore is changed
// the new implementation simply needs to satisfy this interface
type CrawlerRepoManager interface {
	Save(crawlRec *Metadata) error
	GetCrawlHistory() ([]Metadata, error)
	GetCrawlByID(crawlRec *Metadata) (Metadata, error)
	GetCrawlsByHost(crawlRec *Metadata) ([]Metadata, error)
}

type CrawlerRepository struct {
	memstore config.MemoryStore
}

// Create a new repository instance
func NewCrawlerRepository(inMemStore config.MemoryStore) (CrawlerRepoManager, error) {
	if inMemStore != nil {
		return &CrawlerRepository{
			memstore: inMemStore,
		}, nil
	}

	return &CrawlerRepository{}, ErrNoDatastore
}

// Records a crawl request
func (r *CrawlerRepository) Save(crawlRec *Metadata) error {
	_, ok := r.memstore[crawlRec.ID]
	if ok {
		return ErrUniqueKeyViolated
	}
	crawlRec.CreatedAt = time.Now().UTC()
	r.memstore[crawlRec.ID] = *crawlRec

	return nil
}

// Returns a crawl request provided the request ID
func (r *CrawlerRepository) GetCrawlByID(crawlRec *Metadata) (Metadata, error) {
	val, ok := r.memstore[crawlRec.ID]
	if !ok {
		return *crawlRec, ErrRecordNotFound
	}

	cR, ok := val.(Metadata)
	if !ok {
		return *crawlRec, ErrInvalidDataType
	}
	crawlRec = &cR

	return *crawlRec, nil
}

// Returns all crawl requests from memory store
func (r *CrawlerRepository) GetCrawlHistory() ([]Metadata, error) {
	var crawls []Metadata

	for _, cR := range r.memstore {
		crawlRecord, ok := cR.(Metadata)
		if !ok {
			return crawls, ErrInvalidDataType
		}
		crawls = append(crawls, crawlRecord)
	}

	return crawls, nil
}

// Returns all crawl requests from memory that match the provided Host
// (in sorted order from most recent to last)
func (r *CrawlerRepository) GetCrawlsByHost(crawlRec *Metadata) ([]Metadata, error) {
	var crawls []Metadata

	for _, cR := range r.memstore {
		storedRecord, ok := cR.(Metadata)
		if !ok {
			return crawls, ErrInvalidDataType
		}

		if storedRecord.Host == crawlRec.Host {
			crawls = append(crawls, storedRecord)
		}
	}

	// time ordered sort
	if len(crawls) > 1 {
		sort.Slice(crawls, func(i, j int) bool {
			return crawls[i].CreatedAt.After(crawls[j].CreatedAt)
		})
	}

	return crawls, nil
}
