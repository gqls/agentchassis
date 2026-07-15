package main

// accessdigest.go — parse this box's nginx combined access log into a compact,
// key-gated digest: status mix, top referers, top request paths, top 404 paths,
// user-agent buckets, and top client IPs (real client IP, since nginx real_ip
// rewrites $remote_addr from CF-Connecting-IP behind Cloudflare). This is the
// passive half of P4: referer / 404-intent-paths / bot classification that the
// structured /events stream never sees on a static page load.
//
// Per-domain log: setup.sh writes /var/log/nginx/<domain>.access.log, so the
// digest reads one domain's file. The engine user must be able to read it
// (setup.sh adds it to the adm group).

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// combined: $remote_addr - $remote_user [$time_local] "$request" $status $bytes "$referer" "$ua"
var combinedRe = regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "([^"]*)" (\d{3}) \S+ "([^"]*)" "([^"]*)"`)

const nginxTimeLayout = "02/Jan/2006:15:04:05 -0700"

type countPair struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type AccessDigest struct {
	Host         string         `json:"host"`
	Since        string         `json:"since,omitempty"`
	Lines        int            `json:"lines_parsed"`
	Requests     int            `json:"requests"`
	StatusCounts map[string]int `json:"status_counts"`
	NotFound     int            `json:"not_found_total"`
	TopReferers  []countPair    `json:"top_referers"`
	TopPaths     []countPair    `json:"top_paths"`
	Top404Paths  []countPair    `json:"top_404_paths"`
	UABuckets    map[string]int `json:"ua_buckets"`
	TopIPs       []countPair    `json:"top_ips"`
	ServerTime   string         `json:"server_time"`
}

// safeHost guards the log filename against path traversal — only a bare domain.
var safeHostRe = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

func safeHost(h string) bool {
	return h != "" && len(h) <= 255 && safeHostRe.MatchString(h) && !strings.Contains(h, "..")
}

// buildAccessDigest reads {logDir}/{host}.access.log and aggregates. since==zero
// means all lines. topN caps each ranked list.
func buildAccessDigest(logDir, host string, since time.Time, topN int) (*AccessDigest, error) {
	d := &AccessDigest{
		Host:         host,
		StatusCounts: map[string]int{},
		UABuckets:    map[string]int{},
		ServerTime:   time.Now().UTC().Format(time.RFC3339),
	}
	if !since.IsZero() {
		d.Since = since.UTC().Format(time.RFC3339)
	}
	f, err := os.Open(filepath.Join(logDir, host+".access.log"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ref := map[string]int{}
	paths := map[string]int{}
	notFound := map[string]int{}
	ips := map[string]int{}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		m := combinedRe.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		d.Lines++
		if !since.IsZero() {
			if t, err := time.Parse(nginxTimeLayout, m[2]); err == nil && t.Before(since) {
				continue
			}
		}
		d.Requests++
		ip, request, status, referer, ua := m[1], m[3], m[4], m[5], m[6]

		d.StatusCounts[status]++
		ips[ip]++

		// request = "METHOD /path HTTP/1.1"
		path := request
		if parts := strings.Fields(request); len(parts) >= 2 {
			path = parts[1]
		}
		if i := strings.IndexByte(path, '?'); i >= 0 {
			path = path[:i] // drop query string for grouping
		}
		paths[path]++
		if status == "404" {
			d.NotFound++
			notFound[path]++
		}

		if rh := refererHost(referer); rh != "" && rh != host {
			ref[rh]++
		}
		d.UABuckets[classifyUA(ua)]++
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	d.TopReferers = topPairs(ref, topN)
	d.TopPaths = topPairs(paths, topN)
	d.Top404Paths = topPairs(notFound, topN)
	d.TopIPs = topPairs(ips, topN)
	return d, nil
}

func refererHost(referer string) string {
	if referer == "" || referer == "-" {
		return ""
	}
	s := referer
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '@'); i >= 0 {
		s = s[i+1:]
	}
	// canonicalHost lowercases, strips port AND "www." — matches how /events
	// reduces ref_host, so the digest and intent_events.ref_host agree.
	return canonicalHost(s)
}

// classifyUA buckets a user-agent. Heuristic, deliberately coarse — the goal is
// "how much of this is automated", not precise identification.
func classifyUA(ua string) string {
	l := strings.ToLower(ua)
	if ua == "" || ua == "-" {
		return "empty"
	}
	switch {
	case strings.Contains(l, "googlebot"), strings.Contains(l, "bingbot"),
		strings.Contains(l, "duckduckbot"), strings.Contains(l, "applebot"),
		strings.Contains(l, "claude-searchbot"), strings.Contains(l, "gptbot"),
		strings.Contains(l, "oai-searchbot"), strings.Contains(l, "perplexitybot"):
		return "known_search_bot"
	case strings.Contains(l, "semrush"), strings.Contains(l, "ahrefs"),
		strings.Contains(l, "mj12bot"), strings.Contains(l, "dotbot"),
		strings.Contains(l, "dataforseo"), strings.Contains(l, "petalbot"),
		strings.Contains(l, "bytespider"), strings.Contains(l, "yandexbot"):
		return "seo_or_scraper_bot"
	case strings.Contains(l, "bot"), strings.Contains(l, "spider"),
		strings.Contains(l, "crawl"), strings.Contains(l, "http"),
		strings.Contains(l, "python"), strings.Contains(l, "curl"),
		strings.Contains(l, "wget"), strings.Contains(l, "go-http"):
		return "other_bot"
	case strings.Contains(l, "mozilla"):
		return "browser_like"
	default:
		return "other"
	}
}

func topPairs(m map[string]int, n int) []countPair {
	out := make([]countPair, 0, len(m))
	for k, v := range m {
		out = append(out, countPair{Key: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Key < out[j].Key
	})
	if n > 0 && len(out) > n {
		out = out[:n]
	}
	return out
}
