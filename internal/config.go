// Package internal package contains the core logic for scraping
package internal

type urlConfig struct {
	url   string
	depth int
}

func getUrls() []urlConfig {
	return []urlConfig{
		{
			url:   "https://www.moneycontrol.com/news/business/markets/",
			depth: 10,
		},
	}
}
