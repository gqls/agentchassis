// Command backfill-tool-crosslinks repairs the damage recorded in bugs_open/353.
//
// WHAT WENT WRONG (the short version; the file has the full chain): between
// 2026-08-03 and the fix, every genuinely NEW tool's related-page cross-links
// were withheld at birth and nothing ever re-emitted them. The forward fix stops
// new losses; it does not repair the ones already taken, because a tool's birth
// is a one-shot and nothing revisits it.
//
// WHY A BACKFILL IS POSSIBLE AT ALL: the guard that withheld them wrote a
// COUNTABLE row each time (agent_error_log, skip_reason
// 'tool_page_will_not_go_live') — and that row carries `related_pages` verbatim,
// along with the site, the tool function, its display name and the resolved tool
// page URL. The design decision that made this bug findable is the same one that
// makes it repairable. Nothing here is reconstructed or guessed: every input is
// read back out of the record the guard itself wrote.
//
// WHY IT CALLS THE EMITTER RATHER THAN INSERTING ROWS: the cross-link item's
// shape (its spec JSON, its item_key namespace, its dedup clause, its
// depends_on) lives in ONE place — emitToolCrossLinkItems, which inserts through
// the central work-item helper so the ON CONFLICT clause and
// workItemTerminalStatuses stay in lockstep with idx_swi_dedup. A SQL backfill
// would be a second copy of all of that, drifting from the first the moment
// either changes. So this command reuses the real emitter and inherits its
// guards — including the one that refuses a page that is not live.
//
// SAFETY: dry-run by default. --apply writes. --only <function> is the canary.
// Re-running is harmless: the item_key is the dedup unit, so a tool that already
// has its cross-links produces no new rows.
//
//	go run ./cmd/backfill-tool-crosslinks                      # dry run, all
//	go run ./cmd/backfill-tool-crosslinks --only tool-x --apply # canary
//	go run ./cmd/backfill-tool-crosslinks --apply               # the rest
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/actions"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// withheld is one tool's loss, read back from the guard's own record.
type withheld struct {
	siteID       uuid.UUID
	domain       string
	toolFunction string
	toolName     string
	toolDesc     string
	toolPageID   uuid.UUID
	toolPageURL  string
	relatedPages []string
	buildStatus  string
	existing     int
	lastWithheld time.Time
}

// candidateQuery reads the most recent withholding per (site, tool) and joins
// the CURRENT page state plus any cross-link items that already exist.
//
// DISTINCT ON keeps the latest row per tool: a tool withheld more than once
// (a rebuild after a first failure) must be repaired once, not twice, and the
// latest row carries the most recent related_pages the platform actually saw.
const candidateQuery = `
SELECT DISTINCT ON (e.context->>'site_id', e.context->>'tool_function')
       (e.context->>'site_id')::uuid              AS site_id,
       COALESCE(s.domain, '')                     AS domain,
       e.context->>'tool_function'                AS tool_function,
       COALESCE(e.context->>'tool_display_name','') AS tool_name,
       COALESCE(e.context->>'tool_description','')  AS tool_desc,
       (e.context->>'tool_page_id')::uuid         AS tool_page_id,
       COALESCE(e.context->>'tool_page_url','')   AS tool_page_url,
       COALESCE(e.context->'related_pages', '[]'::jsonb) AS related_pages,
       COALESCE(p.build_status, '')               AS build_status,
       (SELECT count(*) FROM site_work_items w
         WHERE w.item_key LIKE 'tool_crosslink:' || (e.context->>'tool_function') || ':%') AS existing,
       e.occurred_at                              AS last_withheld
  FROM agent_error_log e
  LEFT JOIN sites s ON s.id = (e.context->>'site_id')::uuid
  LEFT JOIN pages p ON p.id = (e.context->>'tool_page_id')::uuid
 WHERE e.error_code = 'tool_crosslink_not_emitted:tool_page_will_not_go_live'
   AND e.context ? 'site_id' AND e.context ? 'tool_function' AND e.context ? 'tool_page_id'
 ORDER BY e.context->>'site_id', e.context->>'tool_function', e.occurred_at DESC`

func main() {
	apply := flag.Bool("apply", false, "write the cross-link items (default: dry run)")
	only := flag.String("only", "", "restrict to one tool function — the canary")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	db, err := dbConn()
	if err != nil {
		fmt.Fprintf(os.Stderr, "backfill-tool-crosslinks: %v\n", err)
		os.Exit(1)
	}
	if db == nil {
		fmt.Fprintln(os.Stderr, "backfill-tool-crosslinks: PG_CLIENTS_HOST is not set — refusing to guess a connection")
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	cands, err := loadCandidates(ctx, db, *only)
	if err != nil {
		fmt.Fprintf(os.Stderr, "backfill-tool-crosslinks: %v\n", err)
		os.Exit(1)
	}
	if len(cands) == 0 {
		// A backfill that finds nothing must say so loudly rather than exit 0
		// looking like success: it means the query is wrong, or the damage has
		// already been repaired, and those are different facts.
		fmt.Println("no withheld tools found — either the damage is already repaired or the census query is wrong. NOT reporting success.")
		os.Exit(2)
	}

	var eligible, skippedNotLive, skippedNoPages, skippedHasItems int
	for _, c := range cands {
		switch {
		case len(c.relatedPages) == 0:
			skippedNoPages++
		case c.buildStatus != "deployed" && c.buildStatus != "needs_rebuild":
			skippedNotLive++
		case c.existing > 0:
			skippedHasItems++
		default:
			eligible++
		}
	}

	fmt.Printf("bugs_open/353 backfill — %d withheld tool(s) found (as of %s)\n",
		len(cands), time.Now().UTC().Format("2006-01-02 15:04Z"))
	fmt.Printf("  eligible now:              %d\n", eligible)
	fmt.Printf("  skipped, page not live:    %d\n", skippedNotLive)
	fmt.Printf("  skipped, no related_pages: %d\n", skippedNoPages)
	fmt.Printf("  skipped, items already:    %d\n", skippedHasItems)
	fmt.Println()

	created := 0
	for _, c := range cands {
		tag := fmt.Sprintf("%-46s %-28s", c.toolFunction, c.domain)
		if len(c.relatedPages) == 0 {
			fmt.Printf("  SKIP %s no related_pages in the record\n", tag)
			continue
		}
		if c.buildStatus != "deployed" && c.buildStatus != "needs_rebuild" {
			fmt.Printf("  SKIP %s page build_status=%q — not live, the emitter would refuse anyway\n", tag, c.buildStatus)
			continue
		}
		if c.existing > 0 {
			fmt.Printf("  SKIP %s already has %d cross-link item(s)\n", tag, c.existing)
			continue
		}
		if !*apply {
			fmt.Printf("  WOULD %s -> %d page(s): %v\n", tag, len(c.relatedPages), c.relatedPages)
			continue
		}

		n := actions.EmitToolCrossLinksForBackfill(ctx, backfillParams(ctx, db, logger), logger, actions.BackfillCrossLinkRequest{
			SiteID:       c.siteID,
			ToolFunction: c.toolFunction,
			ToolName:     c.toolName,
			ToolDesc:     c.toolDesc,
			ToolPageID:   c.toolPageID,
			ToolPageURL:  c.toolPageURL,
			RelatedPages: c.relatedPages,
		})
		created += n
		fmt.Printf("  DONE %s created %d item(s) of %d named page(s)\n", tag, n, len(c.relatedPages))
	}

	fmt.Println()
	if *apply {
		fmt.Printf("backfill complete: %d cross-link item(s) created.\n", created)
		fmt.Println("VERIFY AT THE ARTEFACT, not here: the items are content_rewrite requests, so a created row")
		fmt.Println("is a request that the reference will be woven, NOT a page that now carries the link.")
	} else {
		fmt.Println("DRY RUN — nothing written. Re-run with --apply (and --only <function> first).")
	}
}

func loadCandidates(ctx context.Context, db *sql.DB, only string) ([]withheld, error) {
	rows, err := db.QueryContext(ctx, candidateQuery)
	if err != nil {
		return nil, fmt.Errorf("candidate query failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []withheld
	for rows.Next() {
		var c withheld
		var pagesRaw []byte
		if err := rows.Scan(&c.siteID, &c.domain, &c.toolFunction, &c.toolName, &c.toolDesc,
			&c.toolPageID, &c.toolPageURL, &pagesRaw, &c.buildStatus, &c.existing, &c.lastWithheld); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		if only != "" && c.toolFunction != only {
			continue
		}
		var pages []string
		if err := json.Unmarshal(pagesRaw, &pages); err != nil {
			// A record we cannot read is reported, never silently dropped.
			fmt.Fprintf(os.Stderr, "  WARN %s: related_pages unreadable (%v) — skipping\n", c.toolFunction, err)
			continue
		}
		c.relatedPages = pages
		// The display name is not always in the record; fall back to the
		// function rather than emitting an empty tool name into the prose.
		if c.toolName == "" {
			c.toolName = c.toolFunction
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// backfillParams builds the minimum ActionParams the emitter and its telemetry
// need. ExecutionContext must be non-nil: recordCrossLinkSkip reads
// Sender.AgentType and StepName from it, and a nil there is a panic, not a
// no-op.
func backfillParams(ctx context.Context, db *sql.DB, logger *zap.Logger) actions.ActionParams {
	return actions.ActionParams{
		Context: ctx,
		DB:      db,
		Logger:  logger,
		ExecutionContext: &orchtypes.ExecutionContext{
			StepName: "backfill_tool_crosslinks",
			Sender:   orchtypes.AgentIdentity{AgentType: "backfill-tool-crosslinks"},
		},
		AgentType:   "backfill-tool-crosslinks",
		CurrentStep: "backfill_tool_crosslinks",
	}
}
