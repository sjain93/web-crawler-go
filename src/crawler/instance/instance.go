package instance

import (
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"github.com/sjain93/web-crawler-go/src/util"
	"golang.org/x/net/html"
)

// Interface bound to the crawler object, exposes public functions
type CrawlerIManager interface {
	Process()
	GetLinks() []string
	GetErrors() []error
	// ✋🏻 For Testing Only!
	ExtractTestCall(resp *http.Response, url string)
}

// crawlerInstance is an internal implementation of a web crawler instance
// each crawl site request spawns a new instance of the same
type crawlerInstance struct {
	initialURL string
	// using sync map allows multiple threads to interact
	// with this data structure in a thread safe manner
	linkMap sync.Map
	// collect any errors from threads, goal is to return them all at the end
	errMap sync.Map
	// using a channel to allow concurrent crawling
	linkChan chan string
	wg       *sync.WaitGroup
	done     chan struct{}
	// using a semaphore in place of a mutex, to limit concurrency
	// with multithreaded operation
	sem chan struct{}
	// HTTP request client, dereferenced for each instance of the crawler
	client http.Client
}

type Config struct {
	WokerSetting *util.ConcurrencyConfig
	HttpClient   *http.Client
}

// Setup crawler config based on default values defined in utl
func NewDefaultConfig() *Config {
	return &Config{
		WokerSetting: util.SetupDefaultConcurrency(),
		HttpClient:   util.NewDefaultHTTPClient(),
	}
}

func NewCrawler(
	initlUrl string,
	config Config,
) (CrawlerIManager, error) {
	if config.WokerSetting == nil || config.HttpClient == nil {
		return &crawlerInstance{}, errors.New("crawler has invalid or missing config")
	}

	c := &crawlerInstance{
		initialURL: initlUrl,
		linkChan:   make(chan string, config.WokerSetting.TotalWorkers),
		wg:         new(sync.WaitGroup),
		done:       make(chan struct{}),
		// goal is to construct a buffered channel to keep threads in check
		sem:    make(chan struct{}, config.WokerSetting.TotalWorkers),
		client: *config.HttpClient,
	}
	return c, nil
}

// The main orchestrator of the crawler
func (c *crawlerInstance) Process() {
	// Set up a loop that switches on if there are any links in the channel
	// to process
	go func() {
		for {
			select {
			case link := <-c.linkChan:
				go c.crawl(link)
			case <-c.done:
				return
			}
		}
	}()

	// initial call to the function that kicks off the parsing
	c.beginLinkProcessing(c.initialURL)

	// Global call to wait for any potential wait processes in progress
	c.wg.Wait()
	close(c.done)
}

// Public function to get links stored in the sync map
func (c *crawlerInstance) GetLinks() []string {
	links := []string{}
	c.linkMap.Range(func(key, _ interface{}) bool {
		links = append(links, key.(string))
		return true
	})
	return links
}

// Public function to get errors stored in the sync map
// decision to store errors like this was done so that they could
// be inspected when all threads were complete
func (c *crawlerInstance) GetErrors() []error {
	errors := []error{}
	c.errMap.Range(func(key, _ interface{}) bool {
		errors = append(errors, key.(error))
		return true
	})
	return errors
}

// Adding a semaphore for multithreaded locking against a buffered channel
func (c *crawlerInstance) start() {
	c.sem <- struct{}{}
}

// Decrement the waigroup and remove semaphore "lock"
func (c *crawlerInstance) end() {
	c.wg.Done()
	<-c.sem
}

// Start a thread that gets a page's content via HTTP
func (c *crawlerInstance) crawl(urlStr string) {
	c.start()
	defer c.end()

	// fetch the page
	res, err := c.client.Get(urlStr)
	if err != nil {
		c.errMap.Store(
			errors.Wrapf(err, "error fetching page: %s", urlStr),
			struct{}{},
		)
		return
	}
	// scan the page
	c.extract(res, urlStr)
}

// Pull out links from the HTTP response and dispatch them to be validated
// and potentially added to the processing channel
func (c *crawlerInstance) extract(res *http.Response, urlStr string) {
	const (
		htmlATag    = "a"
		htmlHrefTag = "href"
	)

	body := res.Body
	defer body.Close()
	tokenizer := html.NewTokenizer(body)

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == htmlATag {
				for _, attr := range token.Attr {
					if attr.Key == htmlHrefTag {
						link := attr.Val
						c.validateAndDispatch(link, urlStr)
					}
				}
			}
		}
	}
}

// Reference resolution, scheme verification and domain validation
// before preparing it to be pulled and repeat the process as with its parent
func (c *crawlerInstance) validateAndDispatch(link string, baseURL string) {
	// if the link is departing the given base URL's domain, no need to process
	if !util.IsSameDomain(link, baseURL) {
		return
	}

	absUrl, err := util.GetAbsoluteURL(link, baseURL)
	if err != nil {
		c.errMap.Store(
			errors.Wrapf(err, "error getting abs url: %s baseURL: %s", link, baseURL),
			struct{}{},
		)
		return
	}

	// We don't need schemes like mailto
	if !util.IsHTTPScheme(absUrl) {
		return
	}

	if absUrl != "" && util.IsSameDomain(absUrl, c.initialURL) {
		c.beginLinkProcessing(absUrl)
	}
}

// Ensuring the link is new, adding to the waitgroup increment
// to indicate that there is a new thread that the process will need
// to wait for.
func (c *crawlerInstance) beginLinkProcessing(absURL string) {
	// check to see if link has already been stored in the linkmap
	_, visited := c.linkMap.Load(absURL)
	if visited {
		return
	}

	// otherwise add to new link map
	c.linkMap.Store(absURL, struct{}{})

	// continue the iteration and feed the link back into the channel for
	// processing
	c.wg.Add(1)
	c.linkChan <- absURL
}

/*
⚠️ NOTE THE FUNCTION BELOW IS ONLY USED FOR TESTING
*/
func (c *crawlerInstance) ExtractTestCall(resp *http.Response, url string) {
	c.extract(resp, url)
}
