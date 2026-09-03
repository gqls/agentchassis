// FILE: platform/orchestration/actions/render_rss_feed_action.go
//
// RenderRSSFeedAction produces an OUTBOUND RSS 2.0 feed (feed.xml) from a
// site's curated content_feed_items — the publishing counterpart of
// RenderNewsSectionAction's JSON. Each <item> links OUT to the original
// source article (title + summary + attribution only, never full text).
//
// Per-site gate: sites.deploy_config->'rss_feed' must be {"enabled": true, ...}.
// Sites without the flag return {rendered:false, item_count:0} so the shared
// content-feed-orchestrator workflow can carry the step fleet-wide and a
// conditional skips the commit — only opted-in sites (relojistas.com) publish.
//
// deploy_config.rss_feed keys (all optional except enabled):
//   enabled              bool   — the gate
//   channel_title        string — <channel><title>; default "<domain> — News"
//   channel_link         string — <channel><link>;  default https://<domain>/
//   channel_description  string — <channel><description>
//   language             string — <channel><language>, e.g. "es"
//   self_url             string — atom:link rel=self href. For relojistas this
//                                 is the LEGACY vBulletin feed URL
//                                 (/external.php?type=RSS2) so surviving
//                                 subscribers keep their original address.
//
// Output: {files: {"feed.xml": "<xml>"}, domain, item_count, rendered}
//
// Workflow config (content-feed-orchestrator, after commit_news):
//   "render_rss_xml": {
//       "action": "render_rss_feed",
//       "config": {"site_id": "input_data.site_id"},
//       "output_field": "rss_render_result",
//       "next_step": "check_has_rss"
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RenderRSSFeedInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"max_items", "max_age_hours"},
}

func init() {
	datahelpers.RegisterActionInputSpec("render_rss_feed", RenderRSSFeedInputSpec)
}

type rssFeedXML struct {
	XMLName xml.Name      `xml:"rss"`
	Version string        `xml:"version,attr"`
	AtomNS  string        `xml:"xmlns:atom,attr"`
	Channel rssChannelXML `xml:"channel"`
}

type rssChannelXML struct {
	Title         string       `xml:"title"`
	Link          string       `xml:"link"`
	AtomLink      *atomLinkXML `xml:"atom:link,omitempty"`
	Description   string       `xml:"description"`
	Language      string       `xml:"language,omitempty"`
	LastBuildDate string       `xml:"lastBuildDate"`
	Items         []rssItemXML `xml:"item"`
}

type atomLinkXML struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItemXML struct {
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Description string     `xml:"description"`
	PubDate     string     `xml:"pubDate,omitempty"`
	GUID        rssGUIDXML `xml:"guid"`
}

type rssGUIDXML struct {
	Value       string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr"`
}

// rssFeedConfig mirrors sites.deploy_config->'rss_feed'.
type rssFeedConfig struct {
	Enabled            bool   `json:"enabled"`
	ChannelTitle       string `json:"channel_title"`
	ChannelLink        string `json:"channel_link"`
	ChannelDescription string `json:"channel_description"`
	Language           string `json:"language"`
	SelfURL            string `json:"self_url"`
}

func RenderRSSFeedAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "render_rss_feed"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		RenderRSSFeedInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	maxItems := inputs.GetInt("max_items", 30)
	maxAgeHours := inputs.GetInt("max_age_hours", 336) // 14 days — feeds carry more history than the homepage card

	// -----------------------------------------------------------------------
	// 1. Load domain + the per-site rss_feed gate
	// -----------------------------------------------------------------------
	var domain string
	var rssConfigJSON sql.NullString
	err = params.DB.QueryRowContext(ctx, `
		SELECT domain, deploy_config->'rss_feed'
		FROM sites WHERE id = $1
	`, siteID).Scan(&domain, &rssConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("query site: %w", err)
	}

	cfg := rssFeedConfig{}
	if rssConfigJSON.Valid {
		if err := json.Unmarshal([]byte(rssConfigJSON.String), &cfg); err != nil {
			logger.Warn("RenderRSSFeedAction: bad rss_feed config, treating as disabled",
				zap.Error(err))
		}
	}
	if !cfg.Enabled {
		logger.Info("RenderRSSFeedAction: rss_feed not enabled for site, skipping",
			zap.String("domain", domain))
		return map[string]interface{}{
			"rendered":   false,
			"item_count": 0,
			"domain":     domain,
			"reason":     "rss_feed not enabled in deploy_config",
		}, nil
	}

	if cfg.ChannelTitle == "" {
		cfg.ChannelTitle = domain + " — News"
	}
	if cfg.ChannelLink == "" {
		cfg.ChannelLink = "https://" + domain + "/"
	}

	// -----------------------------------------------------------------------
	// 2. Load items — chronological (feed semantics), deduped by source URL
	// -----------------------------------------------------------------------
	items, err := loadRSSItems(ctx, params.DB, siteID, maxAgeHours, maxItems, logger)
	if err != nil {
		return nil, fmt.Errorf("query feed items: %w", err)
	}

	// -----------------------------------------------------------------------
	// 3. Build the XML
	// -----------------------------------------------------------------------
	channel := rssChannelXML{
		Title:         cfg.ChannelTitle,
		Link:          cfg.ChannelLink,
		Description:   cfg.ChannelDescription,
		Language:      cfg.Language,
		LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
		Items:         items,
	}
	if cfg.SelfURL != "" {
		channel.AtomLink = &atomLinkXML{Href: cfg.SelfURL, Rel: "self", Type: "application/rss+xml"}
	}

	feed := rssFeedXML{
		Version: "2.0",
		AtomNS:  "http://www.w3.org/2005/Atom",
		Channel: channel,
	}

	xmlBody, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal rss xml: %w", err)
	}
	feedXML := xml.Header + string(xmlBody) + "\n"

	logger.Info("RenderRSSFeedAction: feed produced",
		zap.String("domain", domain),
		zap.Int("items", len(items)))

	return map[string]interface{}{
		"files":      map[string]interface{}{"feed.xml": feedXML},
		"domain":     domain,
		"item_count": len(items),
		"file_path":  "feed.xml",
		"rendered":   true,
	}, nil
}

// loadRSSItems returns feed items newest-first with raw RFC1123Z dates,
// deduplicated by source URL (the same article can arrive via an RSS source
// and an api_news search). Items without a source URL are skipped — every
// entry must link out to a real article.
func loadRSSItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxAgeHours, maxItems int, logger *zap.Logger) ([]rssItemXML, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			cfi.source_title,
			cfi.source_summary,
			cfi.source_url,
			cfi.source_published_at,
			cfi.created_at,
			COALESCE(cs.name, '') as source_name
		FROM content_feed_items cfi
		LEFT JOIN content_sources cs ON cs.id = cfi.source_id
		WHERE cfi.site_id = $1
		  AND cfi.status IN ('relevant', 'ingested')
		  AND cfi.created_at > NOW() - make_interval(hours => $2)
		  AND (cfi.source_published_at IS NULL
		       OR cfi.source_published_at <= NOW() + INTERVAL '1 day')
		ORDER BY
			COALESCE(cfi.source_published_at, cfi.created_at) DESC,
			CASE WHEN cfi.status = 'relevant' THEN 0 ELSE 1 END,
			cfi.relevance_score DESC NULLS LAST
		LIMIT $3
	`, siteID, maxAgeHours, maxItems*2) // over-fetch: dedupe below may drop rows
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []rssItemXML{}
	seen := map[string]bool{}
	for rows.Next() {
		var title, summary, url, sourceName sql.NullString
		var publishedAt sql.NullTime
		var createdAt time.Time

		if err := rows.Scan(&title, &summary, &url, &publishedAt, &createdAt, &sourceName); err != nil {
			logger.Warn("loadRSSItems: scan error", zap.Error(err))
			continue
		}
		if url.String == "" || title.String == "" || seen[url.String] {
			continue
		}
		seen[url.String] = true

		// STRIPPED 2026-09-03 (bugs_open/332) — this is the surface the bug was
		// originally FILED on. XML escaping keeps the feed well-formed and passes
		// the markdown MARKER characters through as text, so a reader shows
		// "# Heading" and "[text](url)" literally.
		//
		// ORDER IS LOAD-BEARING: strip and cut FIRST, attribution AFTER. The
		// "(Fuente: X)" suffix must never be inside what gets truncated, or the
		// source credit is what the cut eats.
		//
		// Still gated to ONE site (relojistas.com, re-verified 2026-09-03), whose
		// own feed rows carry ZERO markdown — so a clean feed.xml after this is a
		// NO-REGRESSION CONTROL, not evidence the strip works. The signal to watch
		// is the other direction: an <item> count below 25, or an empty
		// <description>, would mean the strip emptied a live feed.
		desc := queryresolve.FeedDisplaySummary(summary.String, 500)
		if sourceName.String != "" {
			if desc != "" {
				desc += " "
			}
			desc += "(Fuente: " + sourceName.String + ")"
		}

		pubTime := createdAt
		if publishedAt.Valid {
			pubTime = publishedAt.Time
		}

		items = append(items, rssItemXML{
			// Titles were emitted verbatim; 2 of 834 rss-sourced rows carried
			// markdown in their title [MEASURED 2026-09-03].
			Title:       queryresolve.FeedDisplayTitle(title.String),
			Link:        url.String,
			Description: strings.TrimSpace(desc),
			PubDate:     pubTime.UTC().Format(time.RFC1123Z),
			GUID:        rssGUIDXML{Value: url.String, IsPermaLink: "true"},
		})
		if len(items) >= maxItems {
			break
		}
	}

	return items, nil
}
