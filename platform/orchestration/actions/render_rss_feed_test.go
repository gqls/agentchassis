package actions

import (
	"context"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// The feed must be valid RSS 2.0 with the legacy URL as atom:self, items
// linking OUT to sources, XML-escaped, deduped by URL — and per-site gated
// so non-opted-in sites skip without committing anything.
func TestLoadRSSItems(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	siteID := uuid.New()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	created := time.Date(2026, 7, 16, 8, 0, 0, 0, time.UTC)

	cols := []string{"source_title", "source_summary", "source_url", "source_published_at", "created_at", "source_name"}
	mock.ExpectQuery("FROM content_feed_items").
		WillReturnRows(sqlmock.NewRows(cols).
			// normal item, published date, needs XML escaping
			AddRow("Rolex & Tudor: <novedades>", "Resumen breve", "https://example.com/a", pub, created, "Tiempo de Relojes").
			// duplicate URL — must be dropped
			AddRow("Duplicado", "Otro resumen", "https://example.com/a", pub, created, "TR Magazine").
			// no published date — falls back to created_at
			AddRow("Sin fecha", "Resumen", "https://example.com/b", nil, created, "Debajo del Reloj").
			// no URL — must be skipped entirely
			AddRow("Sin enlace", "Resumen", "", pub, created, "X"))

	items, err := loadRSSItems(ctx, db, siteID, 336, 30, logger)
	if err != nil {
		t.Fatalf("loadRSSItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items after dedupe+skip, got %d", len(items))
	}
	if items[0].Link != "https://example.com/a" || items[1].Link != "https://example.com/b" {
		t.Fatalf("unexpected links: %+v", items)
	}
	if items[0].PubDate != pub.Format(time.RFC1123Z) {
		t.Fatalf("pubDate: got %q", items[0].PubDate)
	}
	if items[1].PubDate != created.Format(time.RFC1123Z) {
		t.Fatalf("created_at fallback: got %q", items[1].PubDate)
	}
	if !strings.Contains(items[0].Description, "(Fuente: Tiempo de Relojes)") {
		t.Fatalf("source attribution missing: %q", items[0].Description)
	}
	if items[0].GUID.IsPermaLink != "true" || items[0].GUID.Value != items[0].Link {
		t.Fatalf("guid should be the permalink URL: %+v", items[0].GUID)
	}

	// Marshal the full feed and confirm well-formed XML with escaping intact.
	feed := rssFeedXML{
		Version: "2.0",
		AtomNS:  "http://www.w3.org/2005/Atom",
		Channel: rssChannelXML{
			Title:         "Relojistas — Noticias de relojería",
			Link:          "https://relojistas.com/",
			AtomLink:      &atomLinkXML{Href: "https://relojistas.com/external.php?type=RSS2", Rel: "self", Type: "application/rss+xml"},
			Description:   "Noticias en español",
			Language:      "es",
			LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
			Items:         items,
		},
	}
	out, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := xml.Header + string(out)
	if !strings.Contains(s, "Rolex &amp; Tudor: &lt;novedades&gt;") {
		t.Fatalf("XML escaping missing:\n%s", s)
	}
	if !strings.Contains(s, `atom:link href="https://relojistas.com/external.php?type=RSS2" rel="self"`) {
		t.Fatalf("atom:self missing:\n%s", s)
	}
	// Round-trip: the output must parse back as XML.
	var reparsed rssFeedXML
	if err := xml.Unmarshal(out, &reparsed); err != nil {
		t.Fatalf("output is not well-formed XML: %v", err)
	}
	if len(reparsed.Channel.Items) != 2 {
		t.Fatalf("round-trip lost items: %d", len(reparsed.Channel.Items))
	}
}

func TestRenderRSSFeedGate(t *testing.T) {
	// Sites without deploy_config.rss_feed.enabled must skip (rendered=false,
	// item_count=0) so the shared workflow's conditional bypasses the commit.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery("SELECT domain, deploy_config->'rss_feed'").
		WillReturnRows(sqlmock.NewRows([]string{"domain", "rss_feed"}).AddRow("gaswholesalers.com", nil))

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    map[string]interface{}{"input_data": map[string]interface{}{"site_id": siteID.String()}},
		StepConfig:       models.Step{Config: map[string]interface{}{"site_id": "input_data.site_id"}},
	}

	result, err := RenderRSSFeedAction(context.Background(), params)
	if err != nil {
		t.Fatalf("gate should skip, not error: %v", err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if m["rendered"] != false {
		t.Fatalf("expected rendered=false for non-opted-in site, got %+v", m)
	}
	if m["item_count"] != 0 {
		t.Fatalf("expected item_count=0, got %+v", m)
	}
	if _, hasFiles := m["files"]; hasFiles {
		t.Fatalf("skip result must not carry files: %+v", m)
	}
}

// bugs_open/332: the RSS surface is the one this bug was FILED on, and it was
// the one place with no test and no detector — the file's own fix candidate
// asks for "a unit test on loadRSSItems' projection mirroring
// TestProjectNewsItemsStripsLiteralMarkdown". This is it.
//
// IT IS ALSO THE ONLY REAL EVIDENCE FOR THIS ARM. relojistas.com is still the
// only site with rss_feed enabled and its own feed rows carry ZERO markdown
// [MEASURED 2026-09-03], so a clean live feed.xml after this ships proves
// nothing at all — it read clean before. The fixtures below are what actually
// exercise the strip, and each is a shape taken from the live corpus.
func TestLoadRSSItemsStripsLiteralMarkdown(t *testing.T) {
	t.Setenv("DISABLE_NEWS_MARKDOWN_STRIP", "")
	logger := zap.NewNop()
	ctx := context.Background()
	siteID := uuid.New()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	created := time.Date(2026, 9, 3, 8, 0, 0, 0, time.UTC)
	cols := []string{"source_title", "source_summary", "source_url", "source_published_at", "created_at", "source_name"}
	mock.ExpectQuery("FROM content_feed_items").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("## [Novedades](https://example.com/n) de Rolex",
				"# Titular\nTexto con **negrita** y [un enlace](https://example.com/x).",
				"https://example.com/1", nil, created, "Tiempo de Relojes").
			// the live shape: a link severed mid-URL by the 197-byte snippet cut
			AddRow("Sin markdown",
				"Itauma punched himself out and [lost in the ninth round](https://sports.yahoo.com/boxing/live/moses-...",
				"https://example.com/2", nil, created, "Yahoo"))

	items, err := loadRSSItems(ctx, db, siteID, 336, 30, logger)
	if err != nil {
		t.Fatalf("loadRSSItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	for _, it := range items {
		for _, marker := range []string{"](http", "**", "# ", "!["} {
			if strings.Contains(it.Title, marker) {
				t.Errorf("marker %q survived in title: %q", marker, it.Title)
			}
			if strings.Contains(it.Description, marker) {
				t.Errorf("marker %q survived in description: %q", marker, it.Description)
			}
		}
	}

	// The visible text must SURVIVE — a strip that emptied a live feed would
	// pass a "no markers" assertion perfectly.
	if !strings.Contains(items[0].Title, "Novedades de Rolex") {
		t.Errorf("title lost its visible text: %q", items[0].Title)
	}
	if !strings.Contains(items[0].Description, "negrita") || !strings.Contains(items[0].Description, "un enlace") {
		t.Errorf("description lost its visible text: %q", items[0].Description)
	}

	// The truncation marker must survive the severed link, or the feed asserts a
	// complete sentence the source never wrote.
	if !strings.HasSuffix(strings.TrimSuffix(items[1].Description, " (Fuente: Yahoo)"), "...") {
		t.Errorf("truncation marker lost: %q", items[1].Description)
	}

	// ATTRIBUTION ORDER: the suffix is appended AFTER the cut, so it can never be
	// what the truncation eats. Assert it survives on both items.
	for i, it := range items {
		if !strings.Contains(it.Description, "(Fuente: ") {
			t.Errorf("item %d lost its source attribution: %q", i, it.Description)
		}
	}
}

// The kill switch must reach THIS reader too — that is the whole point of moving
// it into the shared projection. If this test passes while the strip test above
// also passes, one lever genuinely disarms this producer.
func TestLoadRSSItemsHonoursTheKillSwitch(t *testing.T) {
	t.Setenv("DISABLE_NEWS_MARKDOWN_STRIP", "1")
	logger := zap.NewNop()
	ctx := context.Background()
	siteID := uuid.New()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	created := time.Date(2026, 9, 3, 8, 0, 0, 0, time.UTC)
	cols := []string{"source_title", "source_summary", "source_url", "source_published_at", "created_at", "source_name"}
	mock.ExpectQuery("FROM content_feed_items").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("Sin markdown", "Texto con **negrita**", "https://example.com/1", nil, created, "TR"))

	items, err := loadRSSItems(ctx, db, siteID, 336, 30, logger)
	if err != nil {
		t.Fatalf("loadRSSItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if !strings.Contains(items[0].Description, "**negrita**") {
		t.Errorf("kill switch did not reach the RSS producer — raw text should have survived, got %q",
			items[0].Description)
	}
}
