// SPDX-License-Identifier: AGPL-3.0-or-later

package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/temoto/robotstxt"
)

// RobotsCache fetches and caches robots.txt files in memory for the lifetime
// of a crawl. It is safe for concurrent use.
type RobotsCache struct {
	mu        sync.Mutex
	cache     map[string]*robotstxt.RobotsData // keyed by scheme+host
	userAgent string
	client    *http.Client
}

// NewRobotsCache creates a RobotsCache that will identify itself using the
// given userAgent when fetching robots.txt files.
func NewRobotsCache(userAgent string) *RobotsCache {
	cache, _ := NewRobotsCacheWithProxy(userAgent, "")
	return cache
}

// NewRobotsCacheWithProxy creates a RobotsCache that uses proxy for robots.txt
// requests. The proxy accepts the same URL formats as crawler backends.
func NewRobotsCacheWithProxy(userAgent, proxy string) (*RobotsCache, error) {
	proxyURL, err := parseProxyURL(proxy)
	if err != nil {
		return nil, err
	}
	return &RobotsCache{
		cache:     make(map[string]*robotstxt.RobotsData),
		userAgent: userAgent,
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transportWithProxy(proxyURL),
		},
	}, nil
}

// Allowed reports whether the given URL is allowed to be fetched according
// to the site's robots.txt. If the robots.txt cannot be fetched it is treated
// as permissive (all URLs allowed) to avoid blocking legitimate crawls due to
// transient network errors.
func (r *RobotsCache) Allowed(ctx context.Context, rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return true
	}
	key := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)

	r.mu.Lock()
	data, ok := r.cache[key]
	if !ok {
		data = r.fetch(ctx, key)
		r.cache[key] = data
	}
	r.mu.Unlock()

	if data == nil {
		return true
	}
	agent := r.userAgent
	if agent == "" {
		agent = "*"
	}
	return data.TestAgent(parsed.RequestURI(), agent)
}

// fetch retrieves and parses robots.txt for the given origin (scheme://host).
// Returns nil when the file is unavailable or unparseable (allow-all).
func (r *RobotsCache) fetch(ctx context.Context, origin string) *robotstxt.RobotsData {
	robotsURL := origin + "/robots.txt"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL, nil)
	if err != nil {
		return nil
	}
	if r.userAgent != "" {
		req.Header.Set("User-Agent", r.userAgent)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.Debug().Err(err).Str("url", robotsURL).Msg("robots: failed to fetch")
		return nil
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warn().Err(err).Msg("robots: failed to close response body")
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		// No robots.txt — allow everything.
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Debug().Err(err).Str("url", robotsURL).Msg("robots: failed to read body")
		return nil
	}

	data, err := robotstxt.FromBytes(body)
	if err != nil {
		log.Debug().Err(err).Str("url", robotsURL).Msg("robots: failed to parse")
		return nil
	}

	log.Debug().Str("origin", origin).Msg("robots: cached robots.txt")
	return data
}
