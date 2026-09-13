// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Timeout          time.Duration
	Deadline         time.Duration
	Interval         time.Duration
	MaxQueue         int
	MaxRetries       int
	MaxResponseBytes int64
	CacheEnabled     bool
	SearchTTL        time.Duration
	DocumentTTL      time.Duration
	CacheItems       int
	CacheBytes       int
	MaxPageSize      int
	Viewer           bool
}

func DefaultConfig() Config {
	return Config{45 * time.Second, 60 * time.Second, 3 * time.Second, 20, 3, 2_000_000, true, 300 * time.Second, 86400 * time.Second, 256, 32_000_000, 20, false}
}
func (c Config) Validate() error {
	if c.Timeout <= 0 || c.Timeout > time.Minute || c.Deadline <= 0 || c.Deadline > 5*time.Minute || c.Interval < 3*time.Second || c.Interval > 5*time.Minute || c.MaxQueue < 1 || c.MaxQueue > 100 || c.MaxRetries < 0 || c.MaxRetries > 5 || c.MaxResponseBytes < 1024 || c.MaxResponseBytes > 10_000_000 || c.CacheItems < 1 || c.CacheItems > 10000 || c.CacheBytes < 1024 || c.CacheBytes > 256_000_000 || c.SearchTTL < 0 || c.SearchTTL > 24*time.Hour || c.DocumentTTL < 0 || c.DocumentTTL > 7*24*time.Hour || c.MaxPageSize < 1 || c.MaxPageSize > 20 {
		return fmt.Errorf("geçersiz yapılandırma sınırı")
	}
	return nil
}
func ConfigFromEnv() (Config, error) {
	c := DefaultConfig()
	for name, dest := range map[string]*time.Duration{"TIMEOUT_SECONDS": &c.Timeout, "DEADLINE_SECONDS": &c.Deadline, "MIN_INTERVAL_SECONDS": &c.Interval, "SEARCH_CACHE_TTL_SECONDS": &c.SearchTTL, "DOCUMENT_CACHE_TTL_SECONDS": &c.DocumentTTL} {
		if raw, ok := os.LookupEnv("YARGITAY_" + name); ok {
			n, e := strconv.ParseFloat(raw, 64)
			if e != nil || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 || n > 604800 {
				return c, fmt.Errorf("geçersiz YARGITAY_%s", name)
			}
			*dest = time.Duration(n * float64(time.Second))
		}
	}
	for name, dest := range map[string]*int{"MAX_QUEUE": &c.MaxQueue, "MAX_RETRIES": &c.MaxRetries, "CACHE_MAX_ITEMS": &c.CacheItems, "CACHE_MAX_BYTES": &c.CacheBytes, "MAX_PAGE_SIZE": &c.MaxPageSize} {
		if raw, ok := os.LookupEnv("YARGITAY_" + name); ok {
			n, e := strconv.Atoi(raw)
			if e != nil {
				return c, fmt.Errorf("geçersiz YARGITAY_%s", name)
			}
			*dest = n
		}
	}
	if raw, ok := os.LookupEnv("YARGITAY_MAX_RESPONSE_BYTES"); ok {
		n, e := strconv.ParseInt(raw, 10, 64)
		if e != nil {
			return c, fmt.Errorf("geçersiz yanıt sınırı")
		}
		c.MaxResponseBytes = n
	}
	for name, dest := range map[string]*bool{"CACHE_ENABLED": &c.CacheEnabled, "VIEWER_ENABLED": &c.Viewer} {
		if raw, ok := os.LookupEnv("YARGITAY_" + name); ok {
			b, e := strconv.ParseBool(raw)
			if e != nil {
				return c, fmt.Errorf("geçersiz YARGITAY_%s", name)
			}
			*dest = b
		}
	}
	if raw, ok := os.LookupEnv("YARGITAY_BASE_URL"); ok && raw != BaseURL {
		return c, fmt.Errorf("yalnızca sabit resmî HTTPS kaynağı kullanılabilir")
	}
	return c, c.Validate()
}
