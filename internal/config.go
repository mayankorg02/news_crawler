// Package internal package contains the core logic for scraping
package internal

import (
	"os"
	"time"
)

type urlConfig struct {
	url   string
	depth int
}

type browserConfig struct {
	executablePath string
	outputDir      string
	userAgent      string
	locale         string
	timeout        time.Duration
	renderWait     time.Duration
	hostDelay      time.Duration
}

func getBrowserConfig() browserConfig {
	return browserConfig{
		executablePath: os.Getenv("NEWS_CRAWLER_CHROME_PATH"),
		outputDir:      envOrDefault("NEWS_CRAWLER_OUTPUT_DIR", "data"),
		userAgent:      "news-crawler/1.0",
		locale:         "en-US",
		timeout:        30 * time.Second,
		renderWait:     time.Second,
		hostDelay:      time.Second,
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getUrls() []urlConfig {
	return []urlConfig{
		{
			url:   "https://www.moneycontrol.com/news/business/markets/",
			depth: 10,
		},
	}
}
