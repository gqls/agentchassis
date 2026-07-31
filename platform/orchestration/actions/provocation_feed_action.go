// FILE: platform/orchestration/actions/provocation_feed_action.go
//
// Builds a site's daily provocation feed from the `provocations` table and
// commits it to the site's repository. Config-driven; nothing site-specific is
// hardcoded here and the domain is REQUIRED.
//
// WHY IT EXISTS
// The provocation on vonc.com had never rotated, because nothing deployable
// could choose today's. The selection rules were proven first in Python
// (provocation_pipeline/builder/, verified across 39 dates) but that script lives
// under docs/, which the cluster cannot execute. This is that logic in the
// platform, reading migration 282's pool, so a scheduled_tasks row finally has
// something to dispatch to.
//
// THE TWO CONTRACTS THIS ACTION SITS BETWEEN, which pull in opposite directions:
//
//  1. THE ENGINE. internal/tools-api/handlers/round.go FetchProvocation reads
//     `today` from the SERVED file server-side, and RoundHandler persists that
//     whole object as the round's provocation. So `today` must always carry
//     headline/body/slug/date. An earlier attempt at the seal below "sealed" the
//     feed by emptying those keys, on the premise that the Gauntlet page never
//     fetches the file. The page does not; the ENGINE does. It would have served
//     every round a blank question.
//
//  2. THE SEAL (owner ruling 2026-07-31). Today's provocation is readable in the
//     Gauntlet, after entry, and NOWHERE else. Home and the Arena show a past
//     provocation in full instead.
//
// Together those mean the seal can only ever be a DISPLAY-level invariant: the
// keys stay, and nothing outside `today` may name today's provocation. checkFeed
// enforces both directions and refuses to emit rather than publishing a feed that
// breaks either. That refusal is the point — it is cheaper to skip a day's
// rotation than to serve a broken round or leak the question.
//
// Actions:
//   - render_provocation_feed: select, build, verify, commit. Fails closed.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

type provocationFeedConfig struct {
	Domain       string
	RepoName     string
	DataPath     string
	Filename     string
	CommitPrefix string
	TaskName     string

	// AllowShrink permits publishing a feed with fewer archive entries than the
	// one currently served. Default false: a shrinking archive means rows were
	// deleted or un-approved, which is either a mistake or a deliberate retirement
	// that deserves to be stated. See checkAgainstServed.
	AllowShrink bool

	// ForceCommit publishes even when the only change would be the timestamp.
	// Default false — see the no-op skip in RenderProvocationFeedAction.
	ForceCommit bool
}

func parseProvocationFeedConfig(config map[string]interface{}) provocationFeedConfig {
	fc := provocationFeedConfig{
		RepoName:     "sites",
		DataPath:     "data",
		Filename:     "provocations.json",
		CommitPrefix: "Update daily provocation",
		TaskName:     "provocation-feed-refresh",
	}
	if v, ok := config["domain"].(string); ok {
		fc.Domain = v
	}
	if v, ok := config["repo_name"].(string); ok && v != "" {
		fc.RepoName = v
	}
	if v, ok := config["data_path"].(string); ok && v != "" {
		fc.DataPath = v
	}
	if v, ok := config["filename"].(string); ok && v != "" {
		fc.Filename = v
	}
	if v, ok := config["commit_message_prefix"].(string); ok && v != "" {
		fc.CommitPrefix = v
	}
	if v, ok := config["task_name"].(string); ok && v != "" {
		fc.TaskName = v
	}
	if v, ok := config["allow_shrink"].(bool); ok {
		fc.AllowShrink = v
	}
	if v, ok := config["force_commit"].(bool); ok {
		fc.ForceCommit = v
	}
	return fc
}

// ---------------------------------------------------------------------------
// The pool row
// ---------------------------------------------------------------------------

type provocation struct {
	Slug       string
	PublishOn  time.Time
	Title      string
	Teaser     string
	CardDesc   string
	DetailBody string
	Headline   string
	Body       string
}

// hasCase reports whether a full case is written. Entries without one render
// non-openable rather than offering a control that leads nowhere.
func (p provocation) hasCase() bool { return strings.TrimSpace(p.DetailBody) != "" }

// shortDate is the display format, "26 Jul". Deliberately not %e (which pads).
func shortDate(t time.Time) string {
	return fmt.Sprintf("%d %s", t.Day(), t.Format("Jan"))
}

// ---------------------------------------------------------------------------
// Selection
// ---------------------------------------------------------------------------

// loadProvocations returns every approved, dated provocation for a domain in
// ascending publish order. Only 'approved' rows are ever returned, so a draft or
// a gate-rejected provocation cannot reach the site by any path through here.
func loadProvocations(ctx context.Context, db *sql.DB, domain string) ([]provocation, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT slug, publish_on,
		       title, teaser,
		       COALESCE(card_desc, ''), COALESCE(detail_body, ''),
		       COALESCE(headline, ''), COALESCE(body, '')
		FROM provocations
		WHERE domain = $1
		  AND status = 'approved'
		  AND publish_on IS NOT NULL
		ORDER BY publish_on ASC`, domain)
	if err != nil {
		return nil, fmt.Errorf("query provocations: %w", err)
	}
	defer rows.Close()

	var out []provocation
	for rows.Next() {
		var p provocation
		if err := rows.Scan(&p.Slug, &p.PublishOn, &p.Title, &p.Teaser,
			&p.CardDesc, &p.DetailBody, &p.Headline, &p.Body); err != nil {
			return nil, fmt.Errorf("scan provocation: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate provocations: %w", err)
	}
	return out, nil
}

// selectForDate splits the schedule into today's provocation and the archive.
//
// today   = the latest entry whose publish date has arrived.
// archive = everything published strictly before it, newest first.
//
// This IS the owner's archive rule ("it can be archived when the new one is
// published") expressed as a property of the data: an entry is archived exactly
// when a later one takes over, so it is never in both places and never in the
// archive during its own day. Nothing has to remember to move it.
func selectForDate(schedule []provocation, on time.Time) (provocation, []provocation, bool) {
	var due []provocation
	for _, p := range schedule {
		if !p.PublishOn.After(on) {
			due = append(due, p)
		}
	}
	if len(due) == 0 {
		return provocation{}, nil, false
	}
	today := due[len(due)-1]

	archive := make([]provocation, 0, len(due)-1)
	for i := len(due) - 2; i >= 0; i-- {
		archive = append(archive, due[i])
	}
	return today, archive, true
}

// ---------------------------------------------------------------------------
// Shapes
// ---------------------------------------------------------------------------

var provocationStats = []map[string]interface{}{
	{"value": "20:00", "label": "On the Clock"},
	{"value": "3", "label": "Objectives"},
	{"value": "1", "label": "AI Verdict"},
}

var provocationIcons = map[string]string{
	"layers": `<path d="M12 2L2 7l10 5 10-5-10-5zm0 9L2 16l10 5 10-5-10-5z"/>`,
	"bolt":   `<path d="M13 2L3 14h7v8l10-12h-7l0-8z"/>`,
	"clock":  `<path d="M12 4a8 8 0 100 16 8 8 0 000-16zm0 3v5l4 2"/>`,
	"chart":  `<path d="M4 19h16M6 16l4-8 3 5 2-3 3 6"/>`,
	"star":   `<path d="M12 3l2.5 6H21l-5 4 2 7-6-4.5L6 20l2-7-5-4h6.5z"/>`,
	"pulse":  `<path d="M3 12h4l3-8 4 16 3-8h4"/>`,
}

var provocationIconCycle = []string{"bolt", "clock", "chart", "star", "pulse", "layers"}

// asToday builds the object the ENGINE consumes. Note the two fallbacks: the
// eight entries authored before rotation existed carry no long-form today-shape,
// so headline falls back to the title and body to the case (then the teaser).
// The fallback is deliberate and documented rather than hidden — anything added
// from now on should author both shapes.
func asToday(p provocation) map[string]interface{} {
	headline := p.Headline
	if headline == "" {
		headline = p.Title
	}
	body := p.Body
	if body == "" {
		body = p.DetailBody
	}
	if body == "" {
		body = p.Teaser
	}
	return map[string]interface{}{
		"eyebrow":     "Today's Provocation",
		"date":        shortDate(p.PublishOn),
		"slug":        p.Slug,
		"headline":    headline,
		"body":        body,
		"primary_cta": map[string]interface{}{"label": "File Your Position", "url": "/tools/gauntlet/index.html"},
		"secondary_cta": map[string]interface{}{
			"label": "See All Provocations", "url": "/provocations/index.html",
		},
		"stats": provocationStats,
	}
}

func asArchiveEntry(p provocation) map[string]interface{} {
	out := map[string]interface{}{
		"date":   shortDate(p.PublishOn),
		"slug":   p.Slug,
		"title":  p.Title,
		"teaser": p.Teaser,
	}
	if p.hasCase() {
		out["detail_body"] = p.DetailBody
		out["url"] = "/provocations/index.html?entry=" + p.Slug
	}
	return out
}

// buildSeal is display-only copy stating that today's provocation is sealed.
//
// A SIBLING of `today`, deliberately not a field inside it: round.go passes the
// whole `today` object through and RoundHandler persists it, so anything added
// there would end up inside every stored round for ever.
func buildSeal() map[string]interface{} {
	return map[string]interface{}{
		"headline": "Today's question is <em>sealed</em>.",
		"body": "You read it when the clock starts, and not before. That is the whole " +
			"point: you commit to arguing before you know what you are arguing " +
			"about, which is the one thing a chat window will never ask of you.",
		"cta": map[string]interface{}{
			"label": "Take On Today's Provocation", "url": "/tools/gauntlet/index.html",
		},
	}
}

// buildSample returns a PAST provocation shown in full, as the worked sample the
// display surfaces show in place of today's.
//
// Safe by construction: `archive` never contains today's entry, because the
// promotion rule keeps an entry out of the archive until a later one takes over.
// So whatever this returns has already been argued. Derived rather than
// configured, so it follows the schedule with no edit here.
//
// Returns nil when no archived entry has a case written; the renderers then fall
// back to the seal alone rather than showing an empty card.
func buildSample(archive []provocation) map[string]interface{} {
	for _, p := range archive {
		if !p.hasCase() {
			continue
		}
		// Opening paragraph only — the full case is one click away and the home
		// card has room for one idea.
		opening := p.DetailBody
		if i := strings.Index(opening, "\n\n"); i >= 0 {
			opening = opening[:i]
		}
		return map[string]interface{}{
			"eyebrow":   "A past provocation",
			"date":      shortDate(p.PublishOn),
			"slug":      p.Slug,
			"headline":  p.Title,
			"body":      opening,
			"cta_label": "Read the full case",
			"url":       "/provocations/index.html?entry=" + p.Slug,
		}
	}
	return nil
}

// buildArena assembles the lobby cards. Card 0 is the SEALED card.
//
// It used to be derived from today's title and blurb, which is right for rotation
// and wrong for the seal: both the home lobby grid and the Arena lobby render this
// card, so it was two of the three surfaces leaking the question the Gauntlet is
// built to hide. Its title and url must stay non-empty — both renderers drop a
// card missing either, which would silently remove the only lobby route into
// today's round.
func buildArena(archive []provocation) map[string]interface{} {
	cards := []map[string]interface{}{{
		"icon":  provocationIcons["layers"],
		"tag":   "Today",
		"title": "Sealed until you step in",
		"desc":  "Today's provocation is revealed when the clock starts, not before.",
		"stat":  "On the clock in the Gauntlet",
		"url":   "/tools/gauntlet/index.html",
	}}
	for i, p := range archive {
		if i >= 5 {
			break
		}
		cards = append(cards, map[string]interface{}{
			"icon":  provocationIcons[provocationIconCycle[i%len(provocationIconCycle)]],
			"tag":   "Archive · " + shortDate(p.PublishOn),
			"title": p.Title,
			"desc":  p.Teaser,
			"stat":  "Read the case",
			"url":   "/provocations/index.html?entry=" + p.Slug,
		})
	}
	return map[string]interface{}{
		"eyebrow": "The Arena",
		"title":   "Every provocation is <em>open</em> to argue.",
		"subtitle": "Pick one, read the case for it, then take a position into the Gauntlet and " +
			"defend it against an AI opponent on a twenty-minute clock.",
		"cta_label": "Not sure where to start? Today's provocation is the one on the clock.",
		"cta":       map[string]interface{}{"label": "See every provocation", "url": "/provocations/index.html"},
		"cards":     cards,
	}
}

// ---------------------------------------------------------------------------
// The checker — refuses to emit a feed that breaks either contract
// ---------------------------------------------------------------------------

// checkFeed enforces the engine contract and the seal together. They pull
// opposite ways, which is why this is one function rather than asserts scattered
// through the builders: satisfying one by itself is how both previous attempts
// went wrong.
//
// Ported from build_provocations.py check_seal(), whose invariants are verified
// across 39 dates by verify_rotation.py. Keep the two in step.
func checkFeed(feed map[string]interface{}, todayEntry provocation) []string {
	var problems []string

	today, _ := feed["today"].(map[string]interface{})
	if today == nil {
		return []string{"feed has no `today` object at all"}
	}

	for _, key := range []string{"headline", "body", "slug", "date"} {
		if s, _ := today[key].(string); strings.TrimSpace(s) == "" {
			problems = append(problems, fmt.Sprintf(
				"today.%s is missing or empty. round.go reads the whole `today` object "+
					"server-side as the round's provocation — this breaks the Gauntlet, "+
					"it does not seal it", key))
		}
	}

	// Everything the display surfaces read, checked against today's actual text.
	todayHeadline, _ := today["headline"].(string)
	todayBody, _ := today["body"].(string)
	todayStrings := []string{}
	for _, s := range []string{todayHeadline, todayBody, todayEntry.Title,
		todayEntry.Teaser, todayEntry.CardDesc} {
		if strings.TrimSpace(s) != "" {
			todayStrings = append(todayStrings, strings.TrimRight(strings.TrimSpace(s), "."))
		}
	}
	leaks := func(value string) bool {
		if value == "" {
			return false
		}
		for _, s := range todayStrings {
			if s != "" && strings.Contains(value, s) {
				return true
			}
		}
		return false
	}

	arena, _ := feed["arena"].(map[string]interface{})
	cards, _ := arena["cards"].([]map[string]interface{})
	if len(cards) == 0 {
		problems = append(problems, "arena has no cards — the lobby would have no route into the round")
	} else {
		card0 := cards[0]
		if tag, _ := card0["tag"].(string); tag != "Today" {
			problems = append(problems,
				"arena.cards[0] is no longer the Today card; this check is looking at the wrong card")
		}
		title, _ := card0["title"].(string)
		url, _ := card0["url"].(string)
		if title == "" || url == "" {
			problems = append(problems,
				"arena.cards[0] needs both a title and a url — either missing and both "+
					"renderers drop the card, removing the only lobby route into today's round")
		}
		desc, _ := card0["desc"].(string)
		if leaks(title) || leaks(desc) {
			problems = append(problems,
				"arena.cards[0] names today's provocation. It must state that today's is "+
					"sealed, not what it is")
		}
	}

	archive, _ := feed["archive"].(map[string]interface{})
	entries, _ := archive["entries"].([]map[string]interface{})
	todaySlug, _ := today["slug"].(string)
	for _, e := range entries {
		if slug, _ := e["slug"].(string); slug == todaySlug {
			problems = append(problems, fmt.Sprintf(
				"today's entry %q is ALSO in the archive — the archive page would publish "+
					"today's case in full", slug))
		}
	}

	if sample, ok := feed["sample"].(map[string]interface{}); ok && sample != nil {
		sampleSlug, _ := sample["slug"].(string)
		if sampleSlug == todaySlug {
			problems = append(problems, "sample is TODAY's provocation, not a past one")
		}
		inArchive := false
		for _, e := range entries {
			if slug, _ := e["slug"].(string); slug == sampleSlug {
				inArchive = true
				break
			}
		}
		if !inArchive {
			problems = append(problems, fmt.Sprintf(
				"sample %q is not in the archive, so it cannot be shown as already argued",
				sampleSlug))
		}
		sh, _ := sample["headline"].(string)
		sb, _ := sample["body"].(string)
		if leaks(sh) || leaks(sb) {
			problems = append(problems, "sample carries today's text")
		}
	}

	return problems
}

// ---------------------------------------------------------------------------
// Build
// ---------------------------------------------------------------------------

// buildProvocationFeed assembles the whole feed for a date and refuses to return
// one that fails checkFeed. Split out from the action so the invariants can be
// tested across many dates without a database or a Kafka producer — the same way
// verify_rotation.py tests them.
func buildProvocationFeed(schedule []provocation, on time.Time, generatedAt string) (map[string]interface{}, provocation, error) {
	today, archive, ok := selectForDate(schedule, on)
	if !ok {
		return nil, provocation{}, fmt.Errorf(
			"no provocation is published on or before %s — the pool holds none that early",
			on.Format("2006-01-02"))
	}

	entries := make([]map[string]interface{}, 0, len(archive))
	for _, p := range archive {
		entries = append(entries, asArchiveEntry(p))
	}

	feed := map[string]interface{}{
		"generated_at": generatedAt,
		"today":        asToday(today),
		// Display-only siblings of `today`. The engine reads `today` and ignores
		// these; the display surfaces read these and must not read `today`.
		"seal":    buildSeal(),
		"arena":   buildArena(archive),
		"archive": map[string]interface{}{"entries": entries},
	}
	if sample := buildSample(archive); sample != nil {
		feed["sample"] = sample
	}

	if problems := checkFeed(feed, today); len(problems) > 0 {
		return nil, today, fmt.Errorf("refusing to emit provocation feed: %s",
			strings.Join(problems, "; "))
	}
	return feed, today, nil
}

// ---------------------------------------------------------------------------
// Comparison against what is actually served
// ---------------------------------------------------------------------------

type servedFeed struct {
	TodaySlug    string
	ArchiveCount int
	Canonical    string // the feed with generated_at removed, for equality
}

// fetchServedFeed reads the file the site is currently serving — the artefact,
// not the repository and not the tag. Used for two things: skipping a no-op
// commit, and refusing a shrinking archive.
func fetchServedFeed(ctx context.Context, domain, dataPath, filename string) (*servedFeed, error) {
	url := fmt.Sprintf("https://%s/%s/%s", domain, strings.Trim(dataPath, "/"), filename)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse served feed: %w", err)
	}
	return summariseFeed(parsed)
}

// summariseFeed reduces a feed to the things worth comparing. `generated_at` is
// dropped before canonicalising, because it changes on every run by design and
// would otherwise make every feed look different from every other one.
func summariseFeed(parsed map[string]interface{}) (*servedFeed, error) {
	out := &servedFeed{}
	if today, ok := parsed["today"].(map[string]interface{}); ok {
		out.TodaySlug, _ = today["slug"].(string)
	}
	if archive, ok := parsed["archive"].(map[string]interface{}); ok {
		if entries, ok := archive["entries"].([]interface{}); ok {
			out.ArchiveCount = len(entries)
		} else if entries, ok := archive["entries"].([]map[string]interface{}); ok {
			out.ArchiveCount = len(entries)
		}
	}
	stripped := make(map[string]interface{}, len(parsed))
	for k, v := range parsed {
		if k == "generated_at" {
			continue
		}
		stripped[k] = v
	}
	canonical, err := json.Marshal(stripped)
	if err != nil {
		return nil, err
	}
	out.Canonical = string(canonical)
	return out, nil
}

// checkAgainstServed compares the feed about to be published with the one being
// served. It answers two different questions and they have opposite failure
// modes, so both are here rather than at the call site.
//
// SHRINK is an error. Fewer archive entries than are already published means rows
// were deleted or un-approved. That is either a mistake or a deliberate
// retirement, and a deliberate one can say so with allow_shrink. This is the
// same rule as the prune floor: a destructive publish must prove it saw the
// corpus.
//
// NO CHANGE is a skip, not an error. If only generated_at would move, committing
// would write a daily no-op into the sites repo AND — worse — advance the one
// timestamp people use to judge freshness while the site repeats itself. That is
// the original bug wearing the fix as a disguise. Skipping keeps the file's git
// history an honest record of rotation.
func checkAgainstServed(served *servedFeed, next *servedFeed, allowShrink bool) (skip bool, err error) {
	if served == nil {
		return false, nil
	}
	if next.ArchiveCount < served.ArchiveCount && !allowShrink {
		return false, fmt.Errorf(
			"refusing to publish: archive would shrink from %d entries to %d. "+
				"Provocations were deleted or un-approved; set allow_shrink to publish anyway",
			served.ArchiveCount, next.ArchiveCount)
	}
	if next.Canonical == served.Canonical {
		return true, nil
	}
	return false, nil
}

// ---------------------------------------------------------------------------
// Action
// ---------------------------------------------------------------------------

// RenderProvocationFeedAction selects today's provocation, builds the feed and
// commits it. It FAILS CLOSED at every step: an empty pool, a failed invariant or
// a shrinking archive all leave the live file exactly as it is. Serving
// yesterday's provocation for another day is a much smaller harm than serving a
// broken round or leaking today's question, and there is no human in this loop to
// catch either.
func RenderProvocationFeedAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RenderProvocationFeed: starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	config := params.StepConfig.Config
	if config == nil {
		config = make(map[string]interface{})
	}
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		for k, v := range inputData {
			config[k] = v
		}
	}

	fc := parseProvocationFeedConfig(config)
	if fc.Domain == "" {
		return nil, fmt.Errorf("provocation feed requires an explicit domain; refusing to publish without one")
	}

	schedule, err := loadProvocations(ctx, params.DB, fc.Domain)
	if err != nil {
		return nil, err
	}
	if len(schedule) == 0 {
		return nil, fmt.Errorf("no approved, dated provocations for %s — refusing to publish an empty feed", fc.Domain)
	}

	// UTC, to match the publish dates and the engine's own clock.
	now := time.Now().UTC()
	feed, today, err := buildProvocationFeed(schedule, now.Truncate(24*time.Hour), now.Format("2006-01-02T15:04:05Z"))
	if err != nil {
		return nil, err
	}

	next, err := summariseFeed(feed)
	if err != nil {
		return nil, fmt.Errorf("summarise built feed: %w", err)
	}

	// Read the artefact, never the repo. A failure here is logged and tolerated:
	// the feed's CONTENT is fully determined by the pool, so publishing without
	// the comparison is still correct — the only cost is a redundant commit. The
	// comparison is an optimisation and a guard, not a correctness input, so it
	// must not be able to block a legitimate publish.
	served, ferr := fetchServedFeed(ctx, fc.Domain, fc.DataPath, fc.Filename)
	if ferr != nil {
		params.Logger.Warn("RenderProvocationFeed: could not read the served feed; "+
			"publishing without the no-op and shrink checks", zap.Error(ferr))
	}

	skip, err := checkAgainstServed(served, next, fc.AllowShrink)
	if err != nil {
		return nil, err
	}
	if skip && !fc.ForceCommit {
		params.Logger.Info("RenderProvocationFeed: no change; skipping commit",
			zap.String("domain", fc.Domain), zap.String("today", today.Slug))
		_, _ = params.DB.ExecContext(ctx,
			`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, fc.TaskName)
		return map[string]interface{}{
			"status": "complete", "committed": false, "reason": "no change since the served feed",
			"domain": fc.Domain, "today": today.Slug,
			"archive_entries": next.ArchiveCount,
		}, nil
	}

	payload, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal feed: %w", err)
	}

	if params.Producer == nil {
		return nil, fmt.Errorf("kafka producer not available")
	}
	path := strings.Trim(fc.DataPath, "/") + "/" + fc.Filename
	files := map[string]interface{}{path: string(payload)}
	commitMsg := fmt.Sprintf("%s — %s (%s)", fc.CommitPrefix, today.Slug, shortDate(today.PublishOn))

	gitResult, err := sendExportFilesToGit(ctx, params, fc.RepoName, fc.Domain, commitMsg, files)
	if err != nil {
		return nil, fmt.Errorf("git commit: %w", err)
	}

	params.Logger.Info("RenderProvocationFeed: published",
		zap.String("domain", fc.Domain), zap.String("today", today.Slug),
		zap.Int("archive_entries", next.ArchiveCount))

	_, _ = params.DB.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1`, fc.TaskName)

	return map[string]interface{}{
		"status": "complete", "committed": true, "domain": fc.Domain,
		"today": today.Slug, "today_date": shortDate(today.PublishOn),
		"archive_entries": next.ArchiveCount, "git_result": gitResult,
	}, nil
}
