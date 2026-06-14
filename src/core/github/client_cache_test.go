package github

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type fakeCache struct {
	etag   string
	body   []byte
	status int
	found  bool
}

func (c *fakeCache) GetHTTPCache(key string) (string, []byte, int, bool, error) {
	return c.etag, c.body, c.status, c.found, nil
}

func (c *fakeCache) SetHTTPCache(key, etag string, body []byte, status int) error {
	c.etag = etag
	c.body = body
	c.status = status
	c.found = true
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestETagTransportReturnsCachedBodyOnNotModified(t *testing.T) {
	cache := &fakeCache{}
	seenIfNoneMatch := ""
	transport := &etagTransport{
		cache: cache,
		base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			seenIfNoneMatch = req.Header.Get("If-None-Match")
			if seenIfNoneMatch != "" {
				return &http.Response{
					StatusCode: http.StatusNotModified,
					Status:     "304 Not Modified",
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
					Request:    req,
				}, nil
			}
			header := make(http.Header)
			header.Set("ETag", `"v1"`)
			header.Set("Content-Type", "application/json; charset=utf-8")
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     header,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Request:    req,
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/owner/repo", nil)
	if err != nil {
		t.Fatal(err)
	}
	first, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = first.Body.Close()

	second, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Body.Close()

	if seenIfNoneMatch != `"v1"` {
		t.Fatalf("expected If-None-Match to use cached etag, got %q", seenIfNoneMatch)
	}
	if second.StatusCode != http.StatusOK {
		t.Fatalf("expected cached 304 response to become 200, got %d", second.StatusCode)
	}
	body, err := io.ReadAll(second.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("expected cached body, got %s", body)
	}
}
