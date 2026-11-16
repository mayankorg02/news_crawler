package internal

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
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

	var output bytes.Buffer
	err := runCrawler(context.Background(), []urlConfig{{url: server.URL, depth: 1}}, server.Client(), 2, &output)
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
	err := runCrawler(context.Background(), []urlConfig{{url: "file:///tmp/news.html"}}, http.DefaultClient, 1, &output)
	if err == nil {
		t.Fatal("runCrawler() error = nil, want invalid seed error")
	}
	if !strings.Contains(output.String(), "invalid seed") {
		t.Fatalf("output does not report invalid seed: %q", output.String())
	}
}
