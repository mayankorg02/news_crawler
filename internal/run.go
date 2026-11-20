// Package internal package contains the core logic for scraping
package internal

import (
	"context"
	"errors"
	"io"
)

const defaultWorkers = 2

// Runner configures the whole scraping process.
type Runner struct {
	Urls []urlConfig
}

// InitRunner creates a runner with the configured seed URLs.
func InitRunner() *Runner {
	return &Runner{
		Urls: getUrls(),
	}
}

// Run fetches the configured seeds and follows same-host links up to each
// seed's configured depth.
func (rn *Runner) Run(ctx context.Context, output io.Writer) error {
	if output == nil {
		return errors.New("output writer cannot be nil")
	}

	fetcher, err := newBrowserFetcher(ctx, getBrowserConfig())
	if err != nil {
		return err
	}
	defer fetcher.Close()

	return runCrawler(ctx, rn.Urls, fetcher.Fetch, defaultWorkers, output)
}
