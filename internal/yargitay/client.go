// SPDX-License-Identifier: AGPL-3.0-only
// This file is the sole boundary for HTTP requests to the official decision source.
package yargitay

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type flight struct {
	done chan struct{}
	data []byte
	err  error
}
type Client struct {
	origin  string
	config  Config
	http    *http.Client
	limiter *limiter
	cache   *memoryCache
	mu      sync.Mutex
	pending int
	flights map[string]*flight
}

func NewClient(c Config) (*Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.Proxy = nil
	t.MaxConnsPerHost = 1
	t.MaxIdleConnsPerHost = 1
	return &Client{config: c, http: &http.Client{Transport: t, Timeout: c.Timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, limiter: sharedLimiter(c), cache: newCache(c.CacheItems, c.CacheBytes), flights: map[string]*flight{}}, nil
}
func (c *Client) Close()         { c.http.CloseIdleConnections() }
func sourceURL(id string) string { return BaseURL + "/getDokuman?" + url.Values{"id": {id}}.Encode() }

func (c *Client) cached(ctx context.Context, namespace string, key any, ttl time.Duration, fetch func(context.Context) ([]byte, error)) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Deadline)
	defer cancel()
	encoded, _ := json.Marshal([]any{namespace, key})
	sum := sha256.Sum256(encoded)
	k := hex.EncodeToString(sum[:])
	if c.config.CacheEnabled {
		if hit := c.cache.get(k); hit != nil {
			return hit, nil
		}
	}
	c.mu.Lock()
	if c.pending >= c.config.MaxQueue {
		c.mu.Unlock()
		return nil, ErrCapacity
	}
	c.pending++
	defer func() { c.mu.Lock(); c.pending--; c.mu.Unlock() }()
	if f := c.flights[k]; f != nil {
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ErrUnavailable
		case <-f.done:
			return f.data, f.err
		}
	}
	f := &flight{done: make(chan struct{})}
	c.flights[k] = f
	c.mu.Unlock()
	// Recheck after obtaining ownership: a previous flight may have populated the cache.
	if c.config.CacheEnabled {
		f.data = c.cache.get(k)
	}
	if f.data == nil {
		f.data, f.err = fetch(ctx)
		if f.err == nil && c.config.CacheEnabled {
			c.cache.set(k, f.data, ttl)
		}
	}
	c.mu.Lock()
	delete(c.flights, k)
	close(f.done)
	c.mu.Unlock()
	return f.data, f.err
}

func searchPayload(r SearchRequest) map[string]any {
	// Public queries are literal phrases, not the upstream's advanced syntax.
	phrase := ""
	if r.Query != nil {
		phrase = `"` + *r.Query + `"`
	}
	d := map[string]any{"arananKelime": phrase, "birimYrgKurulDaire": "", "birimYrgHukukDaire": "", "birimYrgCezaDaire": "", "pageNumber": *r.Page, "pageSize": *r.PageSize, "siralama": map[string]string{"case_number": "1", "decision_number": "2", "decision_date": "3"}[*r.SortBy], "siralamaDirection": *r.SortDirection}
	for k, v := range map[string]*int{"esasYil": r.CaseYear, "esasIlkSiraNo": r.CaseFrom, "esasSonSiraNo": r.CaseTo, "kararYil": r.DecisionYear, "kararIlkSiraNo": r.DecisionFrom, "kararSonSiraNo": r.DecisionTo} {
		d[k] = ""
		if v != nil {
			d[k] = strconv.Itoa(*v)
		}
	}
	for k, v := range map[string]*string{"baslangicTarihi": r.DateFrom, "bitisTarihi": r.DateTo} {
		d[k] = ""
		if v != nil {
			t, _ := time.Parse("2006-01-02", *v)
			d[k] = t.Format("02.01.2006")
		}
	}
	if r.Unit != nil {
		u, _ := FindUnit(*r.Unit)
		d[u.UpstreamField] = u.UpstreamValue
	}
	return map[string]any{"data": d}
}
func (c *Client) Search(ctx context.Context, r SearchRequest) (SearchResponse, error) {
	if err := r.Normalize(); err != nil {
		return SearchResponse{}, err
	}
	if *r.PageSize > c.config.MaxPageSize {
		return SearchResponse{}, ErrInvalid
	}
	b, e := c.cached(ctx, "search", r, c.config.SearchTTL, func(ctx context.Context) ([]byte, error) {
		payload, _ := json.Marshal(searchPayload(r))
		raw, e := c.request(ctx, http.MethodPost, "/aramadetaylist", payload, nil)
		if e != nil {
			return nil, e
		}
		result, e := parseSearch(raw, r)
		if e != nil {
			return nil, e
		}
		return json.Marshal(result)
	})
	if e != nil {
		return SearchResponse{}, e
	}
	var result SearchResponse
	if json.Unmarshal(b, &result) != nil {
		return result, ErrSchema
	}
	return result, nil
}
func (c *Client) GetDecision(ctx context.Context, id string) (Document, error) {
	if !idPattern.MatchString(id) {
		return Document{}, ErrInvalid
	}
	b, e := c.cached(ctx, "document", id, c.config.DocumentTTL, func(ctx context.Context) ([]byte, error) {
		raw, e := c.request(ctx, http.MethodGet, "/getDokuman", nil, url.Values{"id": {id}})
		if e != nil {
			return nil, e
		}
		root, e := envelope(raw)
		if e != nil {
			return nil, e
		}
		var body string
		if json.Unmarshal(root["data"], &body) != nil {
			return nil, ErrSchema
		}
		text, md, e := CleanHTML(body)
		if e != nil {
			return nil, e
		}
		return json.Marshal(Document{ID: id, Text: text, Markdown: md, SourceURL: sourceURL(id), FetchedAt: time.Now().UTC()})
	})
	if e != nil {
		return Document{}, e
	}
	var d Document
	if json.Unmarshal(b, &d) != nil {
		return d, ErrSchema
	}
	return d, nil
}

func (c *Client) request(ctx context.Context, method, path string, body []byte, params url.Values) ([]byte, error) {
	if c.origin != "" {
		return nil, ErrInvalid
	}
	if path != "/aramadetaylist" && path != "/getDokuman" {
		return nil, ErrInvalid
	}
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		release, e := c.limiter.acquire(ctx)
		if e != nil {
			return nil, e
		}
		status, headers, data, e := c.attempt(ctx, method, path, body, params)
		retry := status == 429 || status == 500 || status == 502 || status == 503 || status == 504
		if e == nil && retry {
			d, ok := retryAfter(headers.Get("Retry-After"), time.Now())
			if !ok {
				d = time.Duration(1<<attempt)*time.Second + time.Duration(rand.IntN(1000))*time.Millisecond
			}
			c.limiter.deferFor(d)
		}
		release()
		if e != nil {
			return nil, e
		}
		if retry {
			if attempt == c.config.MaxRetries {
				if status == 429 {
					return nil, ErrRate
				}
				return nil, ErrUnavailable
			}
			continue
		}
		if status == 404 && path == "/getDokuman" {
			return nil, ErrNotFound
		}
		if status == 401 || status == 403 || (status >= 300 && status < 400) {
			return nil, ErrAccess
		}
		if status != 200 {
			return nil, ErrUnavailable
		}
		media, _, err := mime.ParseMediaType(headers.Get("Content-Type"))
		if err != nil || media != "application/json" {
			if bytes.Contains(bytes.ToLower(data), []byte("captcha")) {
				return nil, ErrAccess
			}
			return nil, ErrSchema
		}
		return data, nil
	}
	return nil, ErrUnavailable
}
func (c *Client) attempt(ctx context.Context, method, path string, body []byte, params url.Values) (int, http.Header, []byte, error) {
	origin := c.origin
	if origin == "" {
		origin = BaseURL
	}
	u := origin + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, e := http.NewRequestWithContext(ctx, method, u, bytes.NewReader(body))
	if e != nil {
		return 0, nil, nil, ErrInvalid
	}
	req.Header.Set("User-Agent", "yargitay-mcp/0.3.1 (+https://github.com/sakiphan/yargitay-mcp)")
	req.Header.Set("Accept", "application/json, text/html;q=0.9")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, e := c.http.Do(req)
	if e != nil {
		return 0, nil, nil, ErrUnavailable
	}
	defer resp.Body.Close()
	data, e := io.ReadAll(io.LimitReader(resp.Body, c.config.MaxResponseBytes+1))
	if e != nil {
		return resp.StatusCode, resp.Header, nil, ErrUnavailable
	}
	if int64(len(data)) > c.config.MaxResponseBytes {
		return resp.StatusCode, resp.Header, nil, ErrSchema
	}
	return resp.StatusCode, resp.Header, data, nil
}
func retryAfter(s string, now time.Time) (time.Duration, bool) {
	if s == "" {
		return 0, false
	}
	if n, e := strconv.ParseUint(strings.TrimSpace(s), 10, 64); e == nil {
		n = min(n, 86400)
		return time.Duration(n) * time.Second, true
	}
	if t, e := http.ParseTime(s); e == nil {
		return min(max(t.Sub(now), 0), 24*time.Hour), true
	}
	return 0, false
}
func envelope(raw []byte) (map[string]json.RawMessage, error) {
	var root map[string]json.RawMessage
	if json.Unmarshal(raw, &root) != nil || root == nil {
		return nil, ErrSchema
	}
	var meta map[string]any
	_ = json.Unmarshal(root["metadata"], &meta)
	if meta["FMTY"] == "ERROR" {
		return nil, ErrAccess
	}
	var status string
	_ = json.Unmarshal(root["status"], &status)
	if strings.Contains(status, "EXCEPTION") {
		return nil, ErrAccess
	}
	return root, nil
}
func nullable(row map[string]json.RawMessage, k string) (*string, error) {
	v, ok := row[k]
	if !ok || string(v) == "null" {
		return nil, nil
	}
	var s string
	if json.Unmarshal(v, &s) != nil {
		return nil, ErrSchema
	}
	if s == "" {
		return nil, nil
	}
	return &s, nil
}
func parseSearch(raw []byte, r SearchRequest) (SearchResponse, error) {
	root, e := envelope(raw)
	if e != nil {
		return SearchResponse{}, e
	}
	var data map[string]json.RawMessage
	if json.Unmarshal(root["data"], &data) != nil || data == nil {
		return SearchResponse{}, ErrSchema
	}
	var rows []map[string]json.RawMessage
	if json.Unmarshal(data["data"], &rows) != nil || rows == nil || len(rows) > *r.PageSize {
		return SearchResponse{}, ErrSchema
	}
	out := SearchResponse{Page: *r.Page, PageSize: *r.PageSize, Results: []Summary{}, Warnings: []Warning{}, Source: BaseURL, FetchedAt: time.Now().UTC()}
	out.Review = PendingRelevanceReview(false)
	out.Filters = r
	out.QueryMode = "metadata_only"
	if r.Query != nil {
		out.QueryMode = "phrase"
	}
	if t, ok := data["recordsTotal"]; ok && string(t) != "null" {
		var n int
		if json.Unmarshal(t, &n) != nil || n < 0 {
			return out, ErrSchema
		}
		out.Total = &n
	}
	for _, row := range rows {
		var id string
		if json.Unmarshal(row["id"], &id) != nil {
			id = string(row["id"])
		}
		if !idPattern.MatchString(id) {
			return out, ErrSchema
		}
		s := Summary{ID: id, SourceURL: sourceURL(id)}
		for k, p := range map[string]**string{"daire": &s.Unit, "esasNo": &s.CaseNumber, "kararNo": &s.DecisionNumber, "kararTarihi": &s.DecisionDate} {
			*p, e = nullable(row, k)
			if e != nil {
				return out, e
			}
		}
		if s.DecisionDate != nil {
			d, e := time.Parse("02.01.2006", *s.DecisionDate)
			if e != nil {
				return out, ErrSchema
			}
			s.DecisionDate = ptr(d.Format("2006-01-02"))
		}
		out.Results = append(out.Results, s)
	}
	return out, nil
}

// Probe runs only when the CLI explicitly opts in: one small search + at most one document, no retries.
func Probe(ctx context.Context) (map[string]any, error) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 0
	cfg.CacheEnabled = false
	c, e := NewClient(cfg)
	if e != nil {
		return nil, e
	}
	defer c.Close()
	return probeClient(ctx, c)
}

func probeClient(ctx context.Context, c *Client) (map[string]any, error) {
	report := map[string]any{"search_verified": false, "document_verified": false, "phrase_found_in_first_document": false}
	r, e := c.Search(ctx, SearchRequest{Query: ptr("fazla çalışma"), PageSize: ptr(1)})
	if e != nil {
		return report, e
	}
	report["search_response_received"] = true
	report["returned_rows"] = len(r.Results)
	if len(r.Results) > 0 {
		d, e := c.GetDecision(ctx, r.Results[0].ID)
		if e != nil {
			return report, e
		}
		report["document_verified"] = true
		report["document_chars"] = len([]rune(d.Text))
		// Literal text presence is not a finding of legal relevance.
		found := containsPhrase(d.Text, "fazla çalışma")
		report["phrase_found_in_first_document"] = found
		report["search_verified"] = found
		if !found {
			return report, errors.New("İlk belgedeki ifade eşleşmesi doğrulanamadı; sonuçları ilgili emsal olarak sunmayın.")
		}
	}
	return report, nil
}

func (c *Client) String() string {
	return fmt.Sprintf("YargitayClient(interval=%s)", c.config.Interval)
}

// OfficialRequest shares the process-wide queue and bounded cache. No automatic
// retries, CAPTCHA claims, external URLs, or pagination are accepted here.
func (c *Client) OfficialRequest(ctx context.Context, operation string, params url.Values, body []byte, validate ...func([]byte) error) ([]byte, time.Time, error) {
	path, method, media := "", http.MethodGet, "application/json"
	switch c.origin {
	case AYMOrigin:
		if operation == "search" || operation == "document" {
			path, method = "/api/core/public/search", http.MethodPost
		}
	case DanistayOrigin:
		if operation == "search" {
			path, method = "/aramalist", http.MethodPost
		}
		if operation == "document" {
			path = "/getDokuman"
		}
	case AIHMOrigin:
		if operation == "search" {
			path = "/app/query/results"
		}
		if operation == "document" {
			path, media = "/app/conversion/docx/html/body", "text/html"
		}
	}
	if path == "" {
		return nil, time.Time{}, ErrInvalid
	}
	ttl := c.config.SearchTTL
	if operation == "document" {
		ttl = c.config.DocumentTTL
	}
	type response struct {
		Body      []byte
		FetchedAt time.Time
	}
	b, err := c.cached(ctx, operation, []any{params, string(body)}, ttl, func(ctx context.Context) ([]byte, error) {
		release, err := c.limiter.acquire(ctx)
		if err != nil {
			return nil, err
		}
		defer release()
		status, headers, data, err := c.attempt(ctx, method, path, body, params)
		if err != nil {
			return nil, err
		}
		if status == 429 || status == 503 {
			d, ok := retryAfter(headers.Get("Retry-After"), time.Now())
			if !ok {
				d = 30 * time.Second
			}
			c.limiter.deferFor(d)
		}
		if status == 429 {
			return nil, ErrRate
		}
		if status >= 300 && status < 400 {
			return nil, ErrRedirect
		}
		if status == 401 || status == 403 {
			return nil, ErrAccess
		}
		if status == 404 {
			return nil, ErrNotFound
		}
		if status != 200 {
			return nil, ErrUnavailable
		}
		got, _, err := mime.ParseMediaType(headers.Get("Content-Type"))
		// Danıştay's HTML viewer includes unrelated CAPTCHA scripts. A keyword
		// or script name alone is not evidence of an access challenge.
		danistayHTML := c.origin == DanistayOrigin && operation == "document" && got == "text/html"
		if err != nil || got != media && !danistayHTML {
			return nil, ErrSchema
		}
		for _, check := range validate {
			if err := check(data); err != nil {
				return nil, err
			}
		}
		return json.Marshal(response{data, time.Now().UTC()})
	})
	if err != nil {
		return nil, time.Time{}, err
	}
	var r response
	if json.Unmarshal(b, &r) != nil {
		return nil, time.Time{}, ErrSchema
	}
	return r.Body, r.FetchedAt, nil
}
