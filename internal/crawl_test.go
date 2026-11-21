package internal

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"golang.org/x/net/html"
)

func TestRunCrawlerFollowsSameHostLinksUpToDepth(t *testing.T) {
	var mu sync.Mutex
	requested := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requested[r.URL.Path]++
		mu.Unlock()

		switch r.URL.Path {
		case "/":
			_, _ = w.Write([]byte(`<a href="/article#section">article</a><a href="/article">duplicate</a><a href="https://example.com/outside">outside</a>`))
		case "/article":
			_, _ = w.Write([]byte(`<a href="/too-deep">too deep</a>`))
		case "/too-deep":
			_, _ = w.Write([]byte(`<p>should not be fetched</p>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	fetch := func(ctx context.Context, job crawlJob) crawlResult {
		result := crawlResult{job: job}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, job.url, nil)
		if err != nil {
			result.err = err
			return result
		}
		response, err := server.Client().Do(request)
		if err != nil {
			result.err = err
			return result
		}
		defer response.Body.Close()
		result.status = response.StatusCode
		document, err := html.Parse(response.Body)
		if err != nil {
			result.err = err
			return result
		}
		base, err := url.Parse(job.url)
		if err != nil {
			result.err = err
			return result
		}
		result.links = extractLinks(document, base, job.host)
		return result
	}

	var output bytes.Buffer
	err := runCrawler(context.Background(), []urlConfig{{url: server.URL, depth: 1}}, fetch, 2, &output)
	if err != nil {
		t.Fatalf("runCrawler() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if requested["/"] != 1 || requested["/article"] != 1 {
		t.Fatalf("unexpected requests: %#v", requested)
	}
	if requested["/too-deep"] != 0 {
		t.Fatalf("fetched beyond configured depth: %#v", requested)
	}
	if got := output.String(); !strings.Contains(got, "depth=0 status=200") ||
		!strings.Contains(got, "depth=1 status=200") {
		t.Fatalf("output does not report seed and discovered page results: %q", got)
	}
}

func TestRunCrawlerRejectsInvalidSeed(t *testing.T) {
	var output bytes.Buffer
	err := runCrawler(context.Background(), []urlConfig{{url: "file:///tmp/news.html"}}, func(context.Context, crawlJob) crawlResult {
		t.Fatal("fetch should not be called for an invalid seed")
		return crawlResult{}
	}, 1, &output)
	if err == nil {
		t.Fatal("runCrawler() error = nil, want invalid seed error")
	}
	if !strings.Contains(output.String(), "invalid seed") {
		t.Fatalf("output does not report invalid seed: %q", output.String())
	}
}
