package crawler

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sjain93/web-crawler-go/src/crawler/instance"
	"github.com/sjain93/web-crawler-go/src/util"
	"go.uber.org/zap"
)

var (
	once sync.Once
	svc  *crawlerService
)

// service errors
var (
	ErrSvcRecordExists   = errors.New("target record id already exists")
	ErrSvcRecordNotFound = errors.New("target was not found")
	ErrSvcHostNotFound   = errors.New("provide resource is missing domain")
	ErrSvcProcessError   = errors.New("there was an error during the crawl process")
)

type CrawlerServiceManager interface {
	CrawlSite(crawlRec Metadata) ([]Metadata, error)
	GetCrawlHistory() ([]Metadata, error)
	GetCrawl(id string) ([]Metadata, error)
}

type crawlerService struct {
	logger      *zap.Logger
	crawlerRepo CrawlerRepoManager
}

func NewCrawlerService(r CrawlerRepoManager, l *zap.Logger) CrawlerServiceManager {
	once.Do(func() {
		svc = &crawlerService{
			logger:      l,
			crawlerRepo: r,
		}
	})
	return svc
}

func (s *crawlerService) CrawlSite(crawlRec Metadata) ([]Metadata, error) {
	host, err := util.GetHost(crawlRec.InitialURL)
	if err != nil {
		return []Metadata{}, err
	}
	crawlRec.Host = host
	s.logger.Sugar().Info("valid host")

	// opportunity to early exit if crawler results exist already
	prevCrawls, err := s.crawlerRepo.GetCrawlsByHost(&crawlRec)
	if err != nil {
		return []Metadata{}, err
	}
	if len(prevCrawls) > 0 && inTimeSpan(prevCrawls[0].CreatedAt) {
		s.logger.Sugar().Infof(
			"previous results exist for host - %v",
			crawlRec.Host,
		)
		return []Metadata{prevCrawls[0]}, nil
	}

	// Populate metadata with a new ID for this crawl
	crawlRec.ID = uuid.NewString()

	crawler, err := instance.NewCrawler(
		crawlRec.InitialURL,
		instance.Config{
			WokerSetting: util.SetupDefaultConcurrency(),
			HttpClient:   util.NewDefaultHTTPClient(),
		},
	)
	if err != nil {
		return []Metadata{}, err
	}

	s.logger.Sugar().Info("beginning new web crawl, this may take some time")
	// execute the crawl
	crawler.Process()

	errList := crawler.GetErrors()
	crawlRec.ErrList = errList
	if len(errList) > 0 {
		s.logger.Sugar().Warnf("detected %+v error(s) while web crawling", len(errList))
	}

	validLinks := crawler.GetLinks()
	crawlRec.CrawlResultSet = validLinks

	s.logger.Sugar().Info("crawl complete, caching results")
	err = s.crawlerRepo.Save(&crawlRec)
	if err != nil && errors.Is(err, ErrUniqueKeyViolated) {
		return []Metadata{crawlRec}, ErrSvcRecordExists
	} else if err != nil {
		return []Metadata{crawlRec}, err
	}

	return []Metadata{crawlRec}, nil
}

func (s *crawlerService) GetCrawl(id string) ([]Metadata, error) {
	c := Metadata{ID: id}

	crawlRec, err := s.crawlerRepo.GetCrawlByID(&c)
	if err != nil && errors.Is(err, ErrRecordNotFound) {
		return []Metadata{crawlRec}, ErrSvcRecordNotFound
	} else if err != nil {
		return []Metadata{crawlRec}, err
	}
	s.logger.Sugar().Info("crawl found")
	return []Metadata{crawlRec}, nil
}

func (s *crawlerService) GetCrawlHistory() ([]Metadata, error) {
	return s.crawlerRepo.GetCrawlHistory()
}

// HELPERS ----------------------------------------------------------------
func inTimeSpan(check time.Time) bool {
	end := time.Now().UTC()
	start := end.Add(time.Duration(-24) * time.Hour)

	// start is always 24 hours prior to end
	return !check.Before(start) && !check.After(end)
}
