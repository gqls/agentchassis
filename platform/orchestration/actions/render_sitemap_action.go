// FILE: platform/orchestration/actions/render_sitemap_action.go
//
// RenderSitemapAction produces a site's `sitemap.xml` from the `pages` table.
//
// WHY IT EXISTS (owner ruling 2026-08-20: *"All future sites should have
// sitemaps."*). A generator already existed — `scripts/site-discovery-files.py`,
// register SEO-002 — and it is good, but it is a script someone has to remember
// to run. That entry's own `verify-later` asked whether it should become an
// action; this is the answer. The cost of it staying manual, measured 2026-08-20
// by fetching every live site: **only 8 of 25 serve a sitemap of ours**, and the
// missing 17 include `remortgagecalculator.uk`, built four days earlier with every
// current guard applied. A manual step is not a mechanism.
//
// ⚠ AND ONE OF THOSE "SITEMAPS" WAS NOT OURS. `adversecreditmortgage.co.uk`
// returned 200 on `/sitemap.xml` carrying a single `<loc>` for `/lander` — the
// parking provider's file, still served from the old infrastructure. There is no
// `pages` row named `lander`. **A 200 on /sitemap.xml is not evidence that the
// site has YOUR sitemap**, which is why the count above says "of ours" and why
// anyone re-measuring should read the body.
//
// THE GATE IS INVERTED RELATIVE TO render_rss_feed, deliberately. RSS is opt-IN
// (`deploy_config.rss_feed.enabled`) because most sites should not publish a
// feed. A sitemap is the opposite: every site should have one, so this is ON by
// default and a site opts OUT with `deploy_config.sitemap.enabled = false`. If
// this were opt-in it would reproduce exactly the situation the ruling exists to
// end — a mechanism that works and is switched on almost nowhere.
//
// TWO RULES CARRIED OVER FROM SEO-002 rather than rediscovered:
//
//  1. **Probe before listing.** A sitemap advertising a 404 is worse than no
//     sitemap. Every URL is fetched and only 200s are emitted. But the probe is
//     POINT-IN-TIME and that cuts both ways — SEO-002 records it dropping a URL
//     that was 404 at that moment and deployed 1.5 hours later. So a dropped URL
//     means "not fetchable now", never "broken": the count of drops is returned
//     for the caller to see, and a probe failure never fails the action.
//
//  2. **Read every column that decides whether a page should be FOUND.** SEO-002's
//     generalisable lesson: such a reader is a consumer of every visibility
//     column, and when a new one appears it will not fail — it will silently keep
//     answering the old question. `noindex` arrived after the script and
//     contradicted it for weeks. Enumerated here 2026-08-20, from
//     `information_schema`, with what each one is doing today:
//       status='active'                — 747 rows
//       noindex IS NOT TRUE            — 1 row currently excluded
//       deployed_at IS NOT NULL        — 60 rows currently excluded
//       expires_at NULL or in future   — **0 rows today: LATENT.** Honoured anyway,
//                                        because that is the whole lesson above.
//     `build_status` is deliberately NOT filtered: `deployed_at IS NOT NULL`
//     already excludes `planned`, and `needs_rebuild` pages are live-but-stale,
//     which is a reason to list them, not to hide them.
//
// Output: {files: {"sitemap.xml": "<xml>"}, domain, url_count, probe_dropped,
//          rendered}. The `files` shape matches render_rss_feed so the same
//          commit step can publish it.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RenderSitemapInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	// probe defaults TRUE: rule 1 above is the expensive lesson, so skipping the
	// probe has to be an explicit choice by the caller, not the default.
	Optional:   []string{"probe", "probe_timeout_seconds", "max_urls"},
	Defaults:   map[string]interface{}{"probe": true, "probe_timeout_seconds": 10, "max_urls": 5000},
	Deprecated: map[string]string{},
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

func RenderSitemapAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "render_sitemap"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, RenderSitemapInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	var domain string
	var sitemapCfgJSON sql.NullString
	if err := params.DB.QueryRowContext(ctx, `
		SELECT domain, deploy_config->'sitemap' FROM sites WHERE id = $1
	`, siteID).Scan(&domain, &sitemapCfgJSON); err != nil {
		return nil, fmt.Errorf("query site: %w", err)
	}

	// Opt-OUT, not opt-in. Absent config means ON.
	if sitemapCfgJSON.Valid && sitemapCfgJSON.String != "" {
		var cfg map[string]interface{}
		if err := json.Unmarshal([]byte(sitemapCfgJSON.String), &cfg); err == nil {
			if en, ok := cfg["enabled"].(bool); ok && !en {
				logger.Info("sitemap disabled for this site by deploy_config.sitemap.enabled=false",
					zap.String("domain", domain))
				return map[string]interface{}{
					"rendered": false, "domain": domain, "url_count": 0,
					"reason": "deploy_config.sitemap.enabled is false",
				}, nil
			}
		}
	}

	// See rule 2 in the header for why each clause is here and what it excludes today.
	rows, err := params.DB.QueryContext(ctx, `
		SELECT url,
		       to_char(GREATEST(updated_at, COALESCE(last_built_at, updated_at)), 'YYYY-MM-DD')
		FROM pages
		WHERE site_id = $1
		  AND status = 'active'
		  AND noindex IS NOT TRUE
		  AND deployed_at IS NOT NULL
		  AND (expires_at IS NULL OR expires_at > now())
		ORDER BY url
		LIMIT $2
	`, siteID, inputs.GetInt("max_urls", 5000))
	if err != nil {
		return nil, fmt.Errorf("query pages: %w", err)
	}
	defer rows.Close()

	type candidate struct{ path, lastmod string }
	var candidates []candidate
	for rows.Next() {
		var u, lm sql.NullString
		if err := rows.Scan(&u, &lm); err != nil {
			logger.Warn("sitemap: row scan failed, skipping", zap.Error(err))
			continue
		}
		if !u.Valid || strings.TrimSpace(u.String) == "" {
			continue
		}
		candidates = append(candidates, candidate{path: u.String, lastmod: lm.String})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pages: %w", err)
	}

	probe := inputs.GetBool("probe", true)
	client := &http.Client{Timeout: time.Duration(inputs.GetInt("probe_timeout_seconds", 10)) * time.Second}

	var urls []sitemapURL
	dropped := 0
	var droppedPaths []string
	for _, c := range candidates {
		loc := absoluteURL(domain, c.path)
		if probe {
			ok, status := probeOK(ctx, client, loc)
			if !ok {
				dropped++
				if len(droppedPaths) < 25 {
					droppedPaths = append(droppedPaths, fmt.Sprintf("%s (%d)", c.path, status))
				}
				// Info, not Warn: a drop is expected on a site mid-build and is
				// not by itself a fault. See rule 1.
				logger.Info("sitemap: URL not fetchable now, not listed",
					zap.String("url", loc), zap.Int("status", status))
				continue
			}
		}
		urls = append(urls, sitemapURL{Loc: loc, LastMod: c.lastmod})
	}

	set := sitemapURLSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls}
	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal sitemap: %w", err)
	}
	doc := xml.Header + string(body) + "\n"

	logger.Info("sitemap rendered",
		zap.String("domain", domain),
		zap.Int("candidates", len(candidates)),
		zap.Int("listed", len(urls)),
		zap.Int("dropped_by_probe", dropped),
		zap.Bool("probed", probe))

	// An EMPTY sitemap is worse than none — it tells a crawler the site has no
	// pages. Report rendered=false so a conditional can skip the commit rather
	// than publishing a file that actively misinforms.
	if len(urls) == 0 {
		return map[string]interface{}{
			"rendered": false, "domain": domain, "url_count": 0,
			"candidate_count": len(candidates), "probe_dropped": dropped,
			"dropped_sample": droppedPaths,
			"reason":         "no listable URLs — refusing to publish an empty sitemap",
		}, nil
	}

	return map[string]interface{}{
		"rendered":        true,
		"domain":          domain,
		"url_count":       len(urls),
		"candidate_count": len(candidates),
		"probe_dropped":   dropped,
		"dropped_sample":  droppedPaths,
		"files":           map[string]interface{}{"sitemap.xml": doc},
	}, nil
}

// absoluteURL turns a stored `pages.url` into an absolute https URL. `pages.url`
// is the authoritative path (git_deployer_actions.go:494 says so of nav, sitemap
// and link checks alike), but it is stored inconsistently across the estate —
// some rows carry a leading slash, some do not, some are already absolute.
func absoluteURL(domain, path string) string {
	p := strings.TrimSpace(path)
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return p
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "https://" + domain + p
}

// probeOK fetches a URL and reports whether it may be listed. A GET rather than a
// HEAD: the worker fronting these sites maps `<hostname><path>` onto an object
// store, and HEAD support on that path is not something to assume.
//
// Only a 2xx qualifies. A redirect is deliberately NOT listed — a sitemap should
// carry the canonical URL, and listing a 301 invites a crawler to discover the
// real one by accident instead of being told it.
func probeOK(ctx context.Context, client *http.Client, url string) (bool, int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, 0
	}
	req.Header.Set("User-Agent", "agentchassis-sitemap/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return false, 0
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300, resp.StatusCode
}
