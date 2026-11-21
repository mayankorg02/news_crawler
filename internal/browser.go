package internal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
)

const maxPageBytes = 10 << 20

type browserFetcher struct {
	browserCtx    context.Context
	cancelBrowser context.CancelFunc
	cancelAlloc   context.CancelFunc
	config        browserConfig
	delayMu       sync.Mutex
	nextHostFetch map[string]time.Time
}

type renderedPage struct {
	URL    string `json:"url"`
	HTML   string `json:"html"`
	Status int    `json:"status"`
}

func newBrowserFetcher(ctx context.Context, config browserConfig) (*browserFetcher, error) {
	options := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	options = append(options,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("lang", config.locale),
		chromedp.Flag("user-agent", config.userAgent),
	)
	if config.executablePath != "" {
		options = append(options, chromedp.ExecPath(config.executablePath))
	}

	allocatorCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, options...)
	browserCtx, cancelBrowser := chromedp.NewContext(allocatorCtx)
	if _, err := chromedp.Run(browserCtx, chromedp.Navigate("about:blank")); err != nil {
		cancelBrowser()
		cancelAlloc()
		return nil, fmt.Errorf("start headless browser: %w", err)
	}

	return &browserFetcher{
		browserCtx:    browserCtx,
		cancelBrowser: cancelBrowser,
		cancelAlloc:   cancelAlloc,
		config:        config,
		nextHostFetch: make(map[string]time.Time),
	}, nil
}

func (f *browserFetcher) Close() {
	f.cancelBrowser()
	f.cancelAlloc()
}

func (f *browserFetcher) Fetch(ctx context.Context, job crawlJob) crawlResult {
	result := crawlResult{job: job}
	if err := f.waitForHost(ctx, job.host); err != nil {
		result.err = err
		return result
	}

	pageCtx, cancel := chromedp.NewContext(f.browserCtx)
	defer cancel()
	pageCtx, cancelTimeout := context.WithTimeout(pageCtx, f.config.timeout)
	defer cancelTimeout()

	err := chromedp.Do(pageCtx,
		chromedp.EmulateViewport(1365, 768),
		chromedp.Navigate(job.url),
		chromedp.Sleep(f.config.renderWait),
	)
	if err != nil {
		result.err = fmt.Errorf("render page: %w", err)
		return result
	}
	page, err := chromedp.Run(pageCtx, chromedp.Evaluate[renderedPage](`(() => {
			const navigation = performance.getEntriesByType("navigation")[0];
			return {
				url: location.href,
				html: document.documentElement ? document.documentElement.outerHTML : "",
				status: navigation ? navigation.responseStatus : 0
			};
		})()`))
	if err != nil {
		result.err = fmt.Errorf("render page: %w", err)
		return result
	}
	if page.Status < 1 {
		result.err = errors.New("browser did not provide the page response status")
		return result
	}
	if len(page.HTML) > maxPageBytes {
		result.err = fmt.Errorf("rendered page exceeds %d bytes", maxPageBytes)
		return result
	}

	pageURL, err := url.Parse(page.URL)
	if err != nil {
		result.err = fmt.Errorf("parse rendered page URL: %w", err)
		return result
	}
	result.status = page.Status
	result.file, err = f.savePage(job.url, page.HTML)
	if err != nil {
		result.err = fmt.Errorf("save rendered page: %w", err)
		return result
	}
	if result.status < 200 || result.status >= 300 {
		return result
	}

	document, err := html.Parse(strings.NewReader(page.HTML))
	if err != nil {
		result.err = fmt.Errorf("parse rendered HTML: %w", err)
		return result
	}
	result.links = extractLinks(document, pageURL, job.host)
	return result
}

func (f *browserFetcher) waitForHost(ctx context.Context, host string) error {
	f.delayMu.Lock()
	now := time.Now()
	start := now
	if next := f.nextHostFetch[host]; next.After(start) {
		start = next
	}
	f.nextHostFetch[host] = start.Add(f.config.hostDelay)
	f.delayMu.Unlock()

	delay := time.Until(start)
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (f *browserFetcher) savePage(pageURL, content string) (string, error) {
	hash := sha256.Sum256([]byte(pageURL))
	name := hex.EncodeToString(hash[:]) + ".html"
	if err := os.MkdirAll(f.config.outputDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(f.config.outputDir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
