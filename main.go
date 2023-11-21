package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/manifoldco/promptui"
	"github.com/sjain93/web-crawler-go/config"
	"github.com/sjain93/web-crawler-go/src/crawler"
	"github.com/sjain93/web-crawler-go/src/util"
	"go.uber.org/zap"
)

const (
	NewCrawlOption  = "New Crawl"
	LoadCrawlOption = "Load Crawl"
	AllCrawlOption  = "All Crawls"
	ExitOption      = "Exit"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err.Error())
	}

	inMemDB := config.GetInMemoryStore()
	crawlerRepo, err := crawler.NewCrawlerRepository(inMemDB)
	if err != nil {
		logger.Sugar().Fatalf("Error initializing in memory datastore: %v", err.Error())
	}

	crawlerSvc := crawler.NewCrawlerService(crawlerRepo, logger)

	for {
		prompt := promptui.Select{
			Label: "Select Option",
			Items: []string{
				NewCrawlOption,
				LoadCrawlOption,
				AllCrawlOption,
				ExitOption,
			},
		}

		_, result, err := prompt.Run()
		if err != nil {
			logger.Sugar().Fatalf("Prompt failed %v\n", err)
		}

		var report []crawler.Metadata
		switch result {
		case NewCrawlOption:
			inPrompt := promptui.Prompt{
				Label: "Enter a website to crawl",
				Validate: func(inURL string) error {
					_, err := util.GetHost(inURL)
					return err
				},
			}

			initURL, err := inPrompt.Run()
			if err != nil {
				logger.Sugar().Fatalf("Prompt failed %v\n", err)
			}
			report, err = crawlerSvc.CrawlSite(crawler.Metadata{InitialURL: initURL})
			if err != nil {
				logger.Sugar().Errorf("Error running crawler: %v", err.Error())
				continue
			}
		case LoadCrawlOption:
			inPrompt := promptui.Prompt{
				Label: "Enter a previous crawl result ID",
				Validate: func(crawlID string) error {
					_, err := uuid.Parse(crawlID)
					return err
				},
			}

			crawlID, err := inPrompt.Run()
			if err != nil {
				logger.Sugar().Fatalf("Prompt failed %v\n", err)
			}
			report, err = crawlerSvc.GetCrawl(crawlID)
			if err != nil {
				logger.Sugar().Errorf("Error running crawler: %v", err.Error())
				continue
			}
		case AllCrawlOption:
			report, err = crawlerSvc.GetCrawlHistory()
			if err != nil {
				logger.Sugar().Errorf("Error running crawler: %v", err.Error())
				continue
			}
		case ExitOption:
			os.Exit(0)
		}

		if err = writeReportFile(report); err != nil {
			logger.Sugar().Warnf("Error generating crawl report: %v", err.Error())
			continue
		}

		continue
	}
}

func writeReportFile(report []crawler.Metadata) error {
	file, err := json.MarshalIndent(report, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile("report.json", file, 0o644)
}
