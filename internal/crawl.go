package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

type crawlJob struct {
	url      string
	host     string
	depth    int
	maxDepth int
	seed     int
}

type crawlResult struct {
	job    crawlJob
	status int
	links  []string
	err    error
}

func runCrawler(ctx context.Context, seeds []urlConfig, client *http.Client, workers int, output io.Writer) error {
	pipeline, err := NewPipeline(workers, func(ctx context.Context, job crawlJob) crawlResult {
		return fetchPage(ctx, client, job)
	})
	if err != nil {
		return err
	}

	seen := make([]map[string]struct{}, len(seeds))
	queue := make([]crawlJob, 0, len(seeds))
	var runErrors []error
	for i, seed := range seeds {
		seen[i] = make(map[string]struct{})
		seedURL, err := parseSeed(seed.url)
		if err == nil && seed.depth < 0 {
			err = errors.New("depth cannot be negative")
		}
		if err != nil {
			seedErr := fmt.Errorf("invalid seed %q: %w", seed.url, err)
			if _, writeErr := fmt.Fprintf(output, "error=%q seed=%s\n", seedErr, seed.url); writeErr != nil {
				return errors.Join(errors.Join(runErrors...), seedErr, writeErr)
			}
			runErrors = append(runErrors, seedErr)
			continue
		}

		seedURL.Fragment = ""
		normalized := seedURL.String()
		seen[i][normalized] = struct{}{}
		queue = append(queue, crawlJob{
			url:      normalized,
			host:     seedURL.Hostname(),
			depth:    0,
			maxDepth: seed.depth,
			seed:     i,
		})
	}

	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			runErrors = append(runErrors, err)
			break
		}

		results := pipeline.Run(ctx, queue)
		sort.Slice(results, func(i, j int) bool {
			return results[i].job.url < results[j].job.url
		})

		next := make([]crawlJob, 0)
		for _, result := range results {
			if result.err != nil {
				if _, err := fmt.Fprintf(output, "depth=%d error=%q url=%s\n", result.job.depth, result.err, result.job.url); err != nil {
					return errors.Join(errors.Join(runErrors...), err)
				}
				runErrors = append(runErrors, fmt.Errorf("fetch %s: %w", result.job.url, result.err))
				continue
			}

			if _, err := fmt.Fprintf(output, "depth=%d status=%d links=%d url=%s\n", result.job.depth, result.status, len(result.links), result.job.url); err != nil {
				return errors.Join(errors.Join(runErrors...), err)
			}

			if result.status < http.StatusOK || result.status >= http.StatusMultipleChoices || result.job.depth >= result.job.maxDepth {
				continue
			}
			for _, link := range result.links {
				if _, ok := seen[result.job.seed][link]; ok {
					continue
				}
				seen[result.job.seed][link] = struct{}{}
				next = append(next, crawlJob{
					url:      link,
					host:     result.job.host,
					depth:    result.job.depth + 1,
					maxDepth: result.job.maxDepth,
					seed:     result.job.seed,
				})
			}
		}
		queue = next
	}

	return errors.Join(runErrors...)
}

func parseSeed(raw string) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return nil, err
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return nil, errors.New("seed must be an absolute HTTP or HTTPS URL")
	}
	return parsed, nil
}

func fetchPage(ctx context.Context, client *http.Client, job crawlJob) crawlResult {
	result := crawlResult{job: job}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, job.url, nil)
	if err != nil {
		result.err = err
		return result
	}
	req.Header.Set("User-Agent", "news-crawler/1.0")

	response, err := client.Do(req)
	if err != nil {
		result.err = err
		return result
	}
	defer response.Body.Close()
	result.status = response.StatusCode
	if result.status < http.StatusOK || result.status >= http.StatusMultipleChoices {
		return result
	}

	document, err := html.Parse(io.LimitReader(response.Body, 10<<20))
	if err != nil {
		result.err = err
		return result
	}
	result.links = extractLinks(document, response.Request.URL, job.host)
	return result
}

func extractLinks(document *html.Node, base *url.URL, host string) []string {
	var links []string
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "a") {
			for _, attr := range node.Attr {
				if !strings.EqualFold(attr.Key, "href") {
					continue
				}
				reference, err := url.Parse(strings.TrimSpace(attr.Val))
				if err != nil {
					continue
				}
				target := base.ResolveReference(reference)
				if (target.Scheme != "http" && target.Scheme != "https") ||
					!strings.EqualFold(target.Hostname(), host) {
					continue
				}
				target.Fragment = ""
				links = append(links, target.String())
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(document)
	return links
}
