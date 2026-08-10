// FILE: platform/orchestration/actions/provocation_generator_action.go
//
// The GENERATIVE half of the provocation pipeline (PLAN_2026-07-31 Phase 3), and
// the scheduler that decides when an approved provocation actually runs.
//
// WHY IT EXISTS
// vonc.com's home page promises a daily provocation five times in served HTML and
// served one dated 26 July for ten days (measured 2026-08-05). The publisher was
// never the problem — `provocation-feed-refresh` fired that morning and completed
// in 1.1s through its correct "nothing new, skip" path. The pool was empty: nine
// approved rows, all dated 2026-07-26 or earlier, every one written by hand, and
// zero rows dated ahead. Phases 1 and 2 of the 2026-06-25 plan shipped; 3 and 4
// never did, so nothing has ever refilled the pool.
//
// TWO ACTIONS, DELIBERATELY SEPARATE FROM THE GATE
//
//	generate_provocations  — write candidates into the pool as DRAFTS
//	schedule_provocations  — give approved, undated provocations a publish date
//
// Neither can publish. Generation writes `status='draft'` and NOTHING ELSE — it
// is structurally incapable of producing a publishable row, because the feed
// action selects `status='approved' AND publish_on IS NOT NULL` and this file
// never writes either. Approval is `gate_provocation`'s job and dating is the
// scheduler's, so a runaway generator produces a pile of drafts and no site
// change. That separation is the containment: three steps, each of which can
// only fail in the direction of publishing less.
//
// §10.5 IS ENFORCED HERE, AND IT IS THE ONE PROPERTY WORTH READING THE CODE FOR
// "At most one rotation per day means a broken generator can produce at most one
// bad day before anyone notices, rather than a flood. Do not add a catch-up mode
// that publishes several at once to fill a gap."
//
// So scheduleProvocations assigns AT MOST ONE DATE PER CALENDAR DAY, always
// strictly in the future, starting the day after the latest date already in the
// pool — never filling the ten-day hole behind us. Filling that hole is exactly
// the catch-up the ruling forbids: it would dump ten provocations into the
// archive at once, none of which was ever the provocation "of" its day, and it
// would spend the buffer that makes a bad day survivable.
//
// CURRENCY IS OPTIONAL, AND THE PLAN'S ASSUMPTION ABOUT IT IS WRONG FOR THIS SITE
// PLAN Phase 3 says to "reuse feed-ingester + content_sources for currency".
// Measured 2026-08-05: vonc.com has **0 content_sources and 0 content_feed_items**
// — that pipeline has never run for this site. So currency is read WHEN PRESENT
// and generation proceeds without it when absent, rather than hard-depending on a
// pipeline nobody has configured. That is also defensible on the plan's own
// terms: §10.7 rules criterion (c) "current" the weakest part of the gate and
// keeps it out of the publish decision, so it cannot be a prerequisite for
// producing a candidate either.
//
// Actions:
//   - generate_provocations: LLM writes candidates; inserted as drafts only.
//   - schedule_provocations: date approved provocations, one per day, forward only.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Generation
// ---------------------------------------------------------------------------

// generatedProvocation is one candidate as the model returns it. Field names
// match the pool's columns so the mapping is obvious at the call site.
type generatedProvocation struct {
	Slug   string `json:"slug"`
	Title  string `json:"title"`
	Teaser string `json:"teaser"`
	Body   string `json:"body"`
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// buildGeneratorPrompt asks for candidates in the corpus's shape.
//
// The examples are REAL entries from the pool, not invented ones. The corpus is
// the specification (PLAN §4) and a model shown paraphrases would learn my idea
// of a provocation rather than the owner's — the same reason the calibration
// tests quote the nine verbatim.
func buildGeneratorPrompt(n int, recentTitles []string, currency []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, `Write %d candidate "provocations" for a daily debate site.

A provocation is a deliberately contestable claim, stated flatly as fact, that an
ordinary person can disagree with from their own experience. It is published for
people to argue against.

REQUIRED SHAPE, taken from entries the site has already published:

  title:  "The four-day week is a productivity myth"
  teaser: "The pilots that prove it were self-selected true believers."
  body:   makes the case in a short paragraph, THEN genuinely puts the
          counter-case ("The counter is that...", "Against that:...").

  title:  "Group chats replaced friendship maintenance"
  teaser: "Presence without effort. The bar has never been lower."

RULES, all of which are enforced by a gate that will reject you:
  - State the claim FLATLY. No hedging ("might", "perhaps", "arguably").
  - The title must NOT be a question.
  - The body MUST put the counter-case. A one-sided piece is rejected.
  - NO party politics and no culture-war topics. Not a single named politician,
    party, election, war, or identity-politics subject. This is a hard rule.
  - Arguable from ordinary life. No specialist knowledge needed to disagree.
  - Do NOT invent statistics, studies, named sources or quantities. The thesis
    is opinion and is allowed to be contestable; the supporting prose is
    fact-checked and an invented figure gets the whole candidate rejected.
  - Body between 250 and 900 characters.
  - slug: lowercase words separated by single hyphens, derived from the title.
`, n)

	if len(recentTitles) > 0 {
		b.WriteString("\nDo NOT repeat or closely rephrase any of these, which the site has already used:\n")
		for _, t := range recentTitles {
			b.WriteString("  - " + t + "\n")
		}
	}
	if len(currency) > 0 {
		b.WriteString("\nFor topicality you MAY draw on what is currently being discussed:\n")
		for _, c := range currency {
			b.WriteString("  - " + c + "\n")
		}
		b.WriteString("Do not report the news. Use it only to pick a subject people are arguing about.\n")
	}

	b.WriteString(`
Reply with ONLY a JSON array, no prose, no code fence:
[{"slug":"...","title":"...","teaser":"...","body":"..."}]
`)
	return b.String()
}

// parseGenerated decodes the model's reply.
//
// Strict, for the same reason parseJudgement is: a truncated completion that a
// lenient parser "recovers" is how half a provocation reaches the pool.
// `output_tokens == max_tokens` means the completion was CUT, not finished.
func parseGenerated(raw string) ([]generatedProvocation, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, fmt.Errorf("generator returned an empty reply")
	}
	if i := strings.Index(s, "["); i > 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "]"); j >= 0 && j < len(s)-1 {
		s = s[:j+1]
	}
	var out []generatedProvocation
	dec := json.NewDecoder(strings.NewReader(s))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("generator reply is not the agreed shape: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("generator returned no candidates")
	}
	return out, nil
}

// validateGenerated drops candidates that are structurally unusable.
//
// This is NOT the gate and must not become one. It rejects only what would make
// the row invalid or unjudgeable — a missing field, a malformed slug. Everything
// evaluative belongs to gate_provocation, so that there is exactly one place
// where a provocation is judged. Two judges that can disagree is the drift class
// this estate keeps rediscovering.
func validateGenerated(in []generatedProvocation) (ok []generatedProvocation, dropped []string) {
	seen := map[string]bool{}
	for _, g := range in {
		g.Slug = strings.TrimSpace(strings.ToLower(g.Slug))
		g.Title = strings.TrimSpace(g.Title)
		g.Teaser = strings.TrimSpace(g.Teaser)
		g.Body = strings.TrimSpace(g.Body)

		switch {
		case g.Slug == "" || !slugRe.MatchString(g.Slug):
			dropped = append(dropped, fmt.Sprintf("%q: unusable slug", g.Title))
		case g.Title == "" || g.Teaser == "" || g.Body == "":
			dropped = append(dropped, fmt.Sprintf("%q: missing a required field", g.Slug))
		case seen[g.Slug]:
			dropped = append(dropped, fmt.Sprintf("%q: duplicate slug within the same batch", g.Slug))
		default:
			seen[g.Slug] = true
			ok = append(ok, g)
		}
	}
	return ok, dropped
}

// insertDrafts writes candidates into the pool.
//
// EVERY ROW IS status='draft' WITH publish_on NULL, AND THAT IS NOT NEGOTIABLE.
// The feed action publishes `status='approved' AND publish_on IS NOT NULL`; this
// function writes neither, so no path through the generator can put text on the
// site. If you are adding a config flag to "skip the gate for trusted models",
// stop: PLAN §10 removed the human approver on the explicit understanding that
// the gate is the only remaining control.
//
// ON CONFLICT DO NOTHING on (domain, slug) makes a re-run idempotent rather than
// erroring — a model asked twice for provocations about the same subject often
// produces the same slug, and that is not a failure worth aborting a batch for.
// generatorInsertSQL is the ONLY statement that creates a generated provocation.
//
// It is a named constant rather than an inline string so that the containment
// property — drafts only, never dated, never approved — can be ASSERTED by
// TestGeneratorInsertsDraftsOnly instead of depending on a reviewer noticing.
// A doc comment is not an enforcement mechanism; a test that reads the statement
// is.
const generatorInsertSQL = `
	INSERT INTO provocations
	      (domain, slug, title, teaser, body, detail_body, status, source, source_ref)
	VALUES ($1, $2, $3, $4, $5, $5, 'draft', 'llm', $6)
	ON CONFLICT (domain, slug) DO NOTHING`

func insertDrafts(ctx context.Context, db *sql.DB, domain string, gs []generatedProvocation, sourceRef string) (int, error) {
	inserted := 0
	for _, g := range gs {
		res, err := db.ExecContext(ctx, generatorInsertSQL,
			domain, g.Slug, g.Title, g.Teaser, g.Body, sourceRef)
		if err != nil {
			return inserted, fmt.Errorf("insert draft %q: %w", g.Slug, err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			inserted++
		}
	}
	return inserted, nil
}

// loadRecentTitles gives the generator the titles already used, so it does not
// propose them again. Includes drafts and rejects: proposing a slug that was
// already rejected wastes a gate call on a candidate we have judged.
func loadRecentTitles(ctx context.Context, db *sql.DB, domain string, limit int) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT title FROM provocations
		 WHERE domain = $1
		 ORDER BY created_at DESC
		 LIMIT $2`, domain, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// loadCurrency returns recent ingested headlines for a site, or nothing.
//
// Returning nothing is a normal, expected outcome — see the file header:
// vonc.com has no content_sources at all, so this is empty today. The caller
// must treat an empty result as "no currency signal", never as an error.
func loadCurrency(ctx context.Context, db *sql.DB, siteID string, limit int) ([]string, error) {
	if strings.TrimSpace(siteID) == "" {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `
		SELECT source_title FROM content_feed_items
		 WHERE site_id = $1::uuid
		   AND source_title IS NOT NULL AND source_title <> ''
		   AND created_at > now() - interval '7 days'
		 ORDER BY created_at DESC
		 LIMIT $2`, siteID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GenerateProvocationsAction is the registered entry point for generation.
func GenerateProvocationsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("GenerateProvocations: starting")

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

	domain, _ := config["domain"].(string)
	if strings.TrimSpace(domain) == "" {
		return nil, fmt.Errorf("generate_provocations requires an explicit domain")
	}
	siteID, _ := config["site_id"].(string)
	count := 5
	if n, ok := config["count"].(float64); ok && n > 0 && n <= 20 {
		count = int(n)
	}

	aiCfg := getAIServiceConfig(params)
	if aiCfg == nil {
		return nil, fmt.Errorf("generate_provocations requires an ai_service configuration")
	}
	client, err := createAIClient(ctx, aiCfg)
	if err != nil {
		return nil, fmt.Errorf("create AI client: %w", err)
	}
	model, _ := aiCfg["model"].(string)
	provider, _ := aiCfg["provider"].(string)

	recent, err := loadRecentTitles(ctx, params.DB, domain, 40)
	if err != nil {
		return nil, fmt.Errorf("load recent titles: %w", err)
	}
	currency, cerr := loadCurrency(ctx, params.DB, siteID, 15)
	if cerr != nil {
		// Currency is optional; losing it must not lose the batch.
		params.Logger.Warn("GenerateProvocations: currency lookup failed; generating without it",
			zap.Error(cerr))
		currency = nil
	}
	if len(currency) == 0 {
		params.Logger.Info("GenerateProvocations: no currency signal for this site " +
			"(no content_feed_items); generating from the corpus's own thematic space")
	}

	raw, err := client.GenerateText(ctx, buildGeneratorPrompt(count, recent, currency), map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("generator call failed: %w", err)
	}
	cands, err := parseGenerated(raw)
	if err != nil {
		return nil, err
	}
	ok, dropped := validateGenerated(cands)
	for _, d := range dropped {
		params.Logger.Warn("GenerateProvocations: dropped a candidate", zap.String("why", d))
	}
	if len(ok) == 0 {
		return map[string]interface{}{
			"status": "complete", "generated": len(cands), "inserted": 0,
			"dropped": len(dropped),
		}, nil
	}

	sourceRef := fmt.Sprintf("%s/%s", provider, model)
	if len(currency) > 0 {
		sourceRef += fmt.Sprintf(" (+%d feed items)", len(currency))
	}
	n, err := insertDrafts(ctx, params.DB, domain, ok, sourceRef)
	if err != nil {
		return nil, err
	}

	params.Logger.Info("GenerateProvocations: done",
		zap.String("domain", domain), zap.Int("inserted", n), zap.Int("dropped", len(dropped)))
	return map[string]interface{}{
		"status": "complete", "generated": len(cands),
		"inserted": n, "dropped": len(dropped),
		"note": "all rows are drafts; gate_provocation must approve before anything can be scheduled",
	}, nil
}

// ---------------------------------------------------------------------------
// Scheduling — where §10.5 lives
// ---------------------------------------------------------------------------

// nextPublishDates returns the dates to assign, one per calendar day.
//
// FORWARD ONLY, AND NEVER MORE THAN ONE PER DAY. `from` is the day after the
// latest date already in the pool, or tomorrow, whichever is later. The gap
// behind us is deliberately NOT filled: back-dating would publish several
// provocations at once, each landing straight in the archive without ever having
// been the provocation "of" its day, and it would consume in one run the buffer
// that makes a single bad day survivable (§10.5).
//
// Pure and total, so the rule is testable without a database — which matters,
// because this is the property that bounds the blast radius of a bad generator.
func nextPublishDates(latestInPool *time.Time, today time.Time, n int) []time.Time {
	if n <= 0 {
		return nil
	}
	day := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}
	start := day(today).AddDate(0, 0, 1) // tomorrow at the earliest
	if latestInPool != nil {
		if after := day(*latestInPool).AddDate(0, 0, 1); after.After(start) {
			start = after
		}
	}
	out := make([]time.Time, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, start.AddDate(0, 0, i))
	}
	return out
}

// pendingProvocation is one approved-but-undated row awaiting a date.
type pendingProvocation struct{ id, slug, category string }

// groupPendingByCategory splits the batch by category, PRESERVING two orders that
// the scheduler's correctness depends on:
//
//   - the returned category order is FIRST-APPEARANCE order, which (because the
//     query is `ORDER BY gated_at ASC NULLS LAST, created_at ASC`) means the
//     category holding the longest-waiting row is scheduled first;
//   - within a category, the original oldest-first order is untouched.
//
// Both matter because the query applies `LIMIT max_assign` BEFORE this grouping.
// A map-iteration order here would make which category gets dated non-
// deterministic between runs on an over-subscribed pool — the same batch would
// schedule different rows each time, which is indistinguishable from a scheduler
// that is simply losing work.
//
// Pure and total so the ordering property can be tested without a database; the
// per-category high-water mark, which is the other half of the fix, is a SQL
// predicate and is exercised live.
func groupPendingByCategory(pending []pendingProvocation) ([]string, map[string][]pendingProvocation) {
	order := make([]string, 0, 4)
	byCategory := make(map[string][]pendingProvocation, 4)
	for _, r := range pending {
		if _, seen := byCategory[r.category]; !seen {
			order = append(order, r.category)
		}
		byCategory[r.category] = append(byCategory[r.category], r)
	}
	return order, byCategory
}

// ScheduleProvocationsAction dates approved-but-undated provocations.
func ScheduleProvocationsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ScheduleProvocations: starting")

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
	domain, _ := config["domain"].(string)
	if strings.TrimSpace(domain) == "" {
		return nil, fmt.Errorf("schedule_provocations requires an explicit domain")
	}
	maxAssign := 7
	if n, ok := config["max_assign"].(float64); ok && n > 0 && n <= 30 {
		maxAssign = int(n)
	}

	// Approved, undated, oldest first — so a provocation that has waited longest
	// runs first and the queue cannot starve.
	// `human_approved_at IS NOT NULL` added 2026-08-09 (migration 320), and it was
	// added because I had already written the opposite into migration 321's comment
	// — "only rows that are already approved AND human_approved_at IS NOT NULL" —
	// which the query did not do. A doc comment is not an enforcement mechanism;
	// the predicate is.
	//
	// Defence in depth rather than duplication: `loadProvocations` also requires the
	// stamp, so an unstamped row could not have PUBLISHED either way. But dating one
	// would have put a row a human never approved into the schedule, where it reads
	// exactly like an approved one — and the next person to relax the feed's
	// predicate (or to read the schedule as a to-do list) inherits that as a live
	// defect rather than a latent one.
	// CATEGORY-AWARE SINCE 2026-08-09 (owner ruling, PLAN §13 ruling 9), and this
	// is the half of RFC_013 that was still missing.
	//
	// The index half already shipped: `idx_provocations_one_per_category_day` is
	// UNIQUE on (domain, category, publish_on), so one provocation PER CATEGORY per
	// day is the representable state. This scheduler was still whole-domain: it took
	// `max(publish_on)` across every category and handed a mixed batch consecutive
	// dates. Two consequences, both silent:
	//
	//  1. A NEW CATEGORY IS NEVER SCHEDULED NEAR-TERM. It inherits the busiest
	//     category's high-water mark, so its first provocation is dated after
	//     everything already queued — months out, with nothing reporting it. This is
	//     the failure the handoff carried as "a category is silently never
	//     scheduled", and it is a scheduling bug, not a feed bug.
	//  2. The per-category day slots are wasted: two categories that could each
	//     publish tomorrow are instead spread across two days.
	//
	// Fixed now, deliberately, WHILE EVERY LIVE ROW IS STILL `general` — so the
	// change is a no-op on today's data (one category in, one group out, identical
	// dates) and cannot bite the day a second category is introduced. Doing it later
	// means doing it under a live defect.
	rows, err := params.DB.QueryContext(ctx, `
		SELECT id::text, slug, category FROM provocations
		 WHERE domain = $1 AND status = 'approved' AND publish_on IS NULL
		   AND human_approved_at IS NOT NULL
		 ORDER BY gated_at ASC NULLS LAST, created_at ASC
		 LIMIT $2`, domain, maxAssign)
	if err != nil {
		return nil, fmt.Errorf("load undated approved provocations: %w", err)
	}
	var pending []pendingProvocation
	for rows.Next() {
		var r pendingProvocation
		if err := rows.Scan(&r.id, &r.slug, &r.category); err != nil {
			rows.Close()
			return nil, err
		}
		pending = append(pending, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(pending) == 0 {
		params.Logger.Info("ScheduleProvocations: nothing approved and undated", zap.String("domain", domain))
		return map[string]interface{}{"status": "complete", "scheduled": 0}, nil
	}

	order, byCategory := groupPendingByCategory(pending)

	scheduled := 0
	firstDate := ""
	perCategory := make(map[string]int, len(order))
	now := time.Now().UTC()

	for _, category := range order {
		group := byCategory[category]

		// The high-water mark is read WITHIN the category. This single predicate is
		// the fix: a category with no approved rows yet correctly starts tomorrow.
		var latest *time.Time
		var lt sql.NullTime
		if err := params.DB.QueryRowContext(ctx,
			`SELECT max(publish_on) FROM provocations
			  WHERE domain = $1 AND category = $2 AND status = 'approved'`,
			domain, category).Scan(&lt); err != nil {
			return nil, fmt.Errorf("read latest publish_on for category %q: %w", category, err)
		}
		if lt.Valid {
			latest = &lt.Time
		}

		dates := nextPublishDates(latest, now, len(group))
		for i, r := range group {
			// The partial unique index on (domain, category, publish_on) makes two
			// approved provocations in one category on one date unrepresentable. A
			// conflict here means another session scheduled concurrently; skip that
			// date rather than failing the batch, and the next run picks the row up.
			res, err := params.DB.ExecContext(ctx, `
				UPDATE provocations SET publish_on = $2
				 WHERE id = $1::uuid AND publish_on IS NULL`, r.id, dates[i])
			if err != nil {
				params.Logger.Warn("ScheduleProvocations: could not date a provocation",
					zap.String("slug", r.slug), zap.String("category", category), zap.Error(err))
				continue
			}
			if n, _ := res.RowsAffected(); n > 0 {
				scheduled++
				perCategory[category]++
				if firstDate == "" || dates[i].Format("2006-01-02") < firstDate {
					firstDate = dates[i].Format("2006-01-02")
				}
				params.Logger.Info("ScheduleProvocations: dated",
					zap.String("slug", r.slug), zap.String("category", category),
					zap.String("publish_on", dates[i].Format("2006-01-02")))
			}
		}
	}

	return map[string]interface{}{
		"status": "complete", "scheduled": scheduled,
		"first":        firstDate,
		"per_category": perCategory,
		"categories":   order,
		"note":         "one per day PER CATEGORY, forward only; the gap behind today is deliberately not backfilled (§10.5)",
	}, nil
}
