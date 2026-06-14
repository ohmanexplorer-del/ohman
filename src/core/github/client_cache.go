package github

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
)

type etagTransport struct {
	base  http.RoundTripper
	cache HTTPCache
}

func (t *etagTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.cacheable(req) {
		return t.base.RoundTrip(req)
	}

	key := "GET " + req.URL.String()
	etag, body, _, found, err := t.cache.GetHTTPCache(key)
	if err != nil {
		log.Printf("GitHub cache read failed for %s: %v", req.URL.Path, err)
	}
	if found && etag != "" {
		req = req.Clone(req.Context())
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if resp == nil {
		return resp, nil
	}

	if resp.StatusCode == http.StatusNotModified && found && len(body) > 0 {
		resp.StatusCode = http.StatusOK
		resp.Status = "200 OK"
		resp.Body = io.NopCloser(bytes.NewReader(body))
		resp.ContentLength = int64(len(body))
		resp.Header.Set("Content-Length", stringInt(len(body)))
		resp.Header.Set("X-Ohman-Cache", "etag-hit")
		return resp, nil
	}

	responseETag := resp.Header.Get("ETag")
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if responseETag == "" || resp.StatusCode < 200 || resp.StatusCode >= 300 || !strings.Contains(contentType, "json") || resp.Body == nil {
		return resp, nil
	}

	body, readErr := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil && readErr == nil {
		readErr = closeErr
	}
	if readErr != nil {
		return nil, readErr
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))
	resp.ContentLength = int64(len(body))
	if err := t.cache.SetHTTPCache(key, responseETag, body, resp.StatusCode); err != nil {
		log.Printf("GitHub cache write failed for %s: %v", req.URL.Path, err)
	}
	return resp, nil
}

func (t *etagTransport) cacheable(req *http.Request) bool {
	return t != nil &&
		t.base != nil &&
		t.cache != nil &&
		req != nil &&
		req.Method == http.MethodGet &&
		req.URL != nil &&
		req.URL.Host == "api.github.com"
}

func stringInt(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
