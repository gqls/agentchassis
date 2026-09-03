// FILE: platform/orchestration/actions/evidence_citations.go
//
// The network half of citation evidence (V5 — SPEC_V5_researched_citations.md):
// fetching a cited source and verifying its verbatim quote, plus the
// acquisition action that turns researched candidate claims into registered,
// verified `citation` facts.
//
// The matching itself is pure and lives in datahelpers/citations.go. This file
// owns everything that touches the world, and classifies failures carefully,
// because they mean different things:
//
//   - quote found            → the citation stands; bump `accessed`.
//   - 200 but quote missing  → CITATION LOST — the source changed, moved or was
//     withdrawn. This is drift: the published claim now rests on nothing.
//   - fetch failed (network, 403, 5xx, unsupported content type) → UNKNOWN, not
//     drift. A paywall going up is not evidence the fact is wrong; it is
//     evidence we can no longer check. Reported as an error, never as loss.
//
// PDFs and other non-text content are refused rather than half-read: a fact
// whose source the platform cannot re-fetch as text should carry
// `"reverifiable": false` and a human attestation of having read it.

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

const (
	citationFetchTimeout = 25 * time.Second
	citationFetchMaxBody = 4 << 20 // 4 MiB of HTML is plenty for any article
	citationFetchUA      = "agentchassis-citation-verifier/1 (+evidence re-verification)"
)

// citationVerifyOutcome classifies one live verification.
type citationVerifyOutcome struct {
	Found      bool   `json:"found"`
	FailClass  string `json:"fail_class,omitempty"` // citation_lost | fetch_error | citation_invalid
	FailDetail string `json:"fail_detail,omitempty"`
}

// verifyCitationLive fetches the citation's URL and checks the quote against
// the document's visible text.
func verifyCitationLive(ctx context.Context, cit *datahelpers.Citation) citationVerifyOutcome {
	text, err := fetchCitationDocument(ctx, cit.URL)
	if err != nil {
		return citationVerifyOutcome{FailClass: "fetch_error", FailDetail: err.Error()}
	}
	if datahelpers.QuoteFoundInText(cit.Quote, text) {
		return citationVerifyOutcome{Found: true}
	}
	return citationVerifyOutcome{
		FailClass:  "citation_lost",
		FailDetail: "the source fetched successfully but the verbatim quote no longer appears in its visible text",
	}
}

// verifyCitationLiveForRule is verifyCitationLive plus RFC_060 §3d/Q6: on a
// page shaped like an FCA Handbook chapter (many rules, one page, no
// rule-level URL), "the quote is on this page" is not enough — the quote
// must be within the SPAN belonging to the rule ruleID names, or a citation
// can verify perfectly while pointing at the wrong rule (measured live:
// CONC 6.7.23's own wording verified against a citation labelled 6.7.17,
// same fetch, same bytes).
//
// A SEPARATE function, deliberately, rather than a parameter added to
// verifyCitationLive — that function has three other call sites
// (directory_claims.go x2, evidence_citations.go's own registration path)
// whose citations carry no rule-attribution concept at all; changing its
// signature would force every one of them to pass an empty string for
// something that means nothing there. ruleID == "" here behaves byte-
// identically to verifyCitationLive (CitationRuleSpan reports
// applicable=false and this falls through to the same whole-page check).
func verifyCitationLiveForRule(ctx context.Context, cit *datahelpers.Citation, ruleID string) citationVerifyOutcome {
	text, err := fetchCitationDocument(ctx, cit.URL)
	if err != nil {
		return citationVerifyOutcome{FailClass: "fetch_error", FailDetail: err.Error()}
	}
	if ruleID != "" {
		if found, applicable := datahelpers.CitationRuleSpan(text, ruleID, cit.Quote); applicable {
			if found {
				return citationVerifyOutcome{Found: true}
			}
			// Headings exist on this page but none of them is ruleID's own —
			// do NOT fall through to whole-page matching. That is precisely
			// the bug: the quote may verify perfectly against a NEIGHBOURING
			// rule's span, which is what made this file necessary.
			return citationVerifyOutcome{
				FailClass: "citation_lost",
				FailDetail: fmt.Sprintf(
					"the quote does not appear within %s's own span on the page — RFC_060 Q6: this page holds "+
						"multiple rules, and the quote may belong to a DIFFERENT one; the citation may name the wrong rule",
					ruleID),
			}
		}
	}
	// No rule-heading structure on this page (e.g. legislation.gov.uk, one
	// rule per page already) — whole-page matching is already correct here.
	if datahelpers.QuoteFoundInText(cit.Quote, text) {
		return citationVerifyOutcome{Found: true}
	}
	return citationVerifyOutcome{
		FailClass:  "citation_lost",
		FailDetail: "the source fetched successfully but the verbatim quote no longer appears in its visible text",
	}
}

// fetchCitationDocument GETs a cited URL and returns its visible text.
func fetchCitationDocument(ctx context.Context, url string) (string, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("refused: citation url must be http(s), got %q", url)
	}
	ctx, cancel := context.WithTimeout(ctx, citationFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", citationFetchUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain;q=0.9,*/*;q=0.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch: HTTP %d", resp.StatusCode)
	}
	ctype := strings.ToLower(resp.Header.Get("Content-Type"))
	isHTML := strings.Contains(ctype, "html") || strings.Contains(ctype, "xml")
	isText := strings.Contains(ctype, "text/plain")
	if ctype != "" && !isHTML && !isText {
		return "", fmt.Errorf("unsupported content type %q — mark the fact reverifiable:false and attest it by hand", ctype)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, citationFetchMaxBody))
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if isText {
		return string(body), nil
	}
	return datahelpers.VisibleTextFromHTML(string(body)), nil
}

// citationDateStale reports whether a citation has aged past its staleness
// policy. The age anchor is the source's own publication date when parseable
// (a 2024 statistic ages regardless of when we last looked), else the last
// verification date. No staleness_days → never stale by age.
func citationDateStale(published, verifiedAt string, stalenessDays float64, today time.Time) bool {
	if stalenessDays <= 0 {
		return false
	}
	anchor, ok := parseFlexibleDate(published)
	if !ok {
		if anchor, ok = parseFlexibleDate(verifiedAt); !ok {
			return false // nothing to age from — the register carries no usable date
		}
	}
	return today.Sub(anchor) > time.Duration(stalenessDays)*24*time.Hour
}

// parseFlexibleDate accepts the forms sources actually print: 2025-06-30,
// 2025-06, 2025.
func parseFlexibleDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"2006-01-02", "2006-01", "2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// citationFactID derives a stable id for a researched fact from its url+quote,
// so re-running acquisition cannot register the same citation twice.
func citationFactID(url, quote string) string {
	h := fnv.New64a()
	h.Write([]byte(url))
	h.Write([]byte{0})
	h.Write([]byte(datahelpers.NormalizeForQuoteMatch(quote)))
	return fmt.Sprintf("CIT-%x", h.Sum64())
}

// ============================================================================
// ACTION: verify_and_register_citations (V5 acquisition endpoint)
// ============================================================================
//
// Input: an array of candidate claims extracted by an LLM from scraped
// sources. Each candidate must carry a verbatim quote; each is verified LIVE
// against its URL before anything is written. Verified candidates become
// `citation` facts in the site's evidence_base (created if the site has none);
// failures become ONE needs_human_review work item and are NEVER registered.
//
// The order embodies the design: the model proposes, the string comparison
// disposes. An LLM cannot register a fact its cited source does not contain.
//
// Config:
//   "site_id":    path to the site uuid            (required)
//   "candidates": path to the extracted array      (required) — each element:
//       {claim, quote, url, publisher, title,
//        value?, unit?, kind?, tolerance?, published?, writer_line?,
//        staleness_days?, context_terms?,
//        event_date?, venue?, participants?, broadcaster?}
//   "dry_run":    verify and report, write nothing (optional, default false)

var VerifyAndRegisterCitationsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("verify_and_register_citations", VerifyAndRegisterCitationsInputSpec)
}

func VerifyAndRegisterCitationsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "verify_and_register_citations"))
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, VerifyAndRegisterCitationsInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}
	dryRun := configBoolOrDefault(params.StepConfig.Config, "dry_run", false)

	candidatesPath, _ := params.StepConfig.Config["candidates"].(string)
	if candidatesPath == "" {
		return nil, fmt.Errorf("config 'candidates' (dot-path to the extracted claims array) is required")
	}
	rawCandidates := datahelpers.ExtractNestedField(params.CollectedData, candidatesPath)
	candidateList, ok := rawCandidates.([]interface{})
	if !ok {
		return nil, fmt.Errorf("candidates at %q is not an array (got %T)", candidatesPath, rawCandidates)
	}

	// ── Load (or start) the register, remembering the row id for CAS ──
	var specRowID uuid.UUID
	var rawJSON []byte
	var pinned bool
	registerExists := true
	err = params.DB.QueryRowContext(ctx, `
		SELECT id, data, COALESCE(pinned, false) FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
	`, siteID).Scan(&specRowID, &rawJSON, &pinned)
	if err != nil {
		registerExists = false
		rawJSON = []byte(`{"facts": [], "banned_claims": []}`)
		pinned = true
	}
	var eb map[string]interface{}
	if err := json.Unmarshal(rawJSON, &eb); err != nil {
		return nil, fmt.Errorf("unmarshal evidence_base: %w", err)
	}
	facts, _ := eb["facts"].([]interface{})

	// Existing citation ids, for idempotent re-runs.
	existing := map[string]bool{}
	for _, fr := range facts {
		if fm, ok := fr.(map[string]interface{}); ok {
			existing[datahelpers.GetStringField(fm, "id", "")] = true
		}
	}

	today := currentDateString(ctx, params.DB)

	type rejected struct {
		Claim  string `json:"claim"`
		URL    string `json:"url"`
		Class  string `json:"class"`
		Detail string `json:"detail"`
	}
	var registered []string
	var skipped []string
	var failures []rejected
	changed := false

	for _, cr := range candidateList {
		cand, ok := cr.(map[string]interface{})
		if !ok {
			continue
		}
		claim := strings.TrimSpace(datahelpers.GetStringField(cand, "claim", ""))
		cit, cerr := datahelpers.ParseCitation(map[string]interface{}{"citation": map[string]interface{}{
			"publisher": cand["publisher"], "title": cand["title"], "url": cand["url"],
			"published": cand["published"], "quote": cand["quote"], "accessed": today,
		}})
		if cerr != nil || cit == nil || claim == "" {
			detail := "candidate missing claim text"
			if cerr != nil {
				detail = cerr.Error()
			}
			failures = append(failures, rejected{Claim: claim,
				URL: datahelpers.GetStringField(cand, "url", ""), Class: "citation_invalid", Detail: detail})
			continue
		}

		id := datahelpers.GetStringField(cand, "id", "")
		if id == "" {
			id = citationFactID(cit.URL, cit.Quote)
		}
		if existing[id] {
			skipped = append(skipped, id)
			continue
		}

		// The disposal step: the model proposed this; the source must confirm it.
		outcome := verifyCitationLive(ctx, cit)
		if !outcome.Found {
			failures = append(failures, rejected{Claim: claim, URL: cit.URL,
				Class: outcome.FailClass, Detail: outcome.FailDetail})
			continue
		}

		fact := map[string]interface{}{
			"id":    id,
			"claim": claim,
			"kind":  datahelpers.GetStringField(cand, "kind", "metric"),
			"source": map[string]interface{}{"citation": map[string]interface{}{
				"publisher": cit.Publisher, "title": cit.Title, "url": cit.URL,
				"published": cit.Published, "quote": cit.Quote, "accessed": today,
			}},
			"verified_at": today,
		}
		for _, k := range []string{
			"value", "unit", "tolerance", "writer_line", "staleness_days", "context_terms",
			// Event-shaped fields (bugs_open/427, news_feed_ingestion lane): a
			// confirmed dated real-world event extracted from a feed item,
			// registered with kind="entity" (datahelpers.FactKindEntity — "a
			// named thing that exists"; "event" is NOT in the canonical
			// vocabulary in claims.go and would silently demote to "metric"
			// via CanonicalKind() while also tripping UnrecognisedKinds()'s
			// per-build warning, bugs_open/105). Verification is unchanged —
			// the same live-quote check applies to these facts as to any
			// other kind; only the extra fields differ.
			"event_date", "venue", "participants", "broadcaster",
		} {
			if v, has := cand[k]; has && v != nil {
				fact[k] = v
			}
		}
		facts = append(facts, fact)
		existing[id] = true
		registered = append(registered, id)
		changed = true
	}

	if !dryRun && changed {
		eb["facts"] = facts
		if err := writeCitationRegister(ctx, params, siteID, specRowID, registerExists, eb, pinned,
			len(registered), logger); err != nil {
			return nil, err
		}
	}

	if !dryRun && len(failures) > 0 {
		if err := createCitationFailuresItem(ctx, params, siteID, failures, logger); err != nil {
			logger.Warn("verify_and_register_citations: failed to create review item", zap.Error(err))
		}
	}

	logger.Info("verify_and_register_citations: complete",
		zap.Int("candidates", len(candidateList)),
		zap.Int("registered", len(registered)),
		zap.Int("already_registered", len(skipped)),
		zap.Int("rejected", len(failures)),
		zap.Bool("register_created", !registerExists && changed && !dryRun),
		zap.Bool("dry_run", dryRun))

	return map[string]interface{}{
		"candidates":         len(candidateList),
		"registered":         registered,
		"already_registered": skipped,
		"rejected":           failures,
		"register_created":   !registerExists && changed && !dryRun,
		"dry_run":            dryRun,
	}, nil
}

// writeCitationRegister persists the grown register: CAS supersede when a row
// existed (a concurrent human edit must never be clobbered — same rule as the
// freshness pass), plain insert when the site had no register yet.
func writeCitationRegister(
	ctx context.Context, params ActionParams, siteID, specRowID uuid.UUID,
	registerExists bool, eb map[string]interface{}, pinned bool,
	added int, logger *zap.Logger,
) error {
	newJSON, err := json.Marshal(eb)
	if err != nil {
		return fmt.Errorf("marshal register: %w", err)
	}
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if registerExists {
		res, err := tx.ExecContext(ctx, `
			UPDATE site_specs SET is_current = false, superseded_at = now()
			WHERE id = $1 AND is_current = true
		`, specRowID)
		if err != nil {
			return fmt.Errorf("supersede register: %w", err)
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return fmt.Errorf("register changed under us (row %s no longer current) — nothing written; re-run to retry against the fresh copy", specRowID)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
		VALUES ($1, 'evidence_base', $2::jsonb, 'research',
		        $3, true, 'evidence-researcher', $4)
	`, siteID, string(newJSON),
		fmt.Sprintf("V5 acquisition: %d citation fact(s) registered after live quote verification", added),
		pinned,
	); err != nil {
		return fmt.Errorf("insert register: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit register: %w", err)
	}
	logger.Info("verify_and_register_citations: register written",
		zap.String("site_id", siteID.String()), zap.Int("facts_added", added))
	return nil
}

// createCitationFailuresItem raises the rejects for a human. HITL-terminal,
// like every other judgement in this layer: a citation the machine could not
// verify is a decision for a person, never a silent drop and never a fact.
func createCitationFailuresItem(
	ctx context.Context, params ActionParams, siteID uuid.UUID,
	failures interface{}, logger *zap.Logger,
) error {
	specJSON, err := json.Marshal(map[string]interface{}{
		"check":    "citation_verification",
		"rejected": failures,
		"fix": "Each rejected candidate either cited a source that does not contain its quote " +
			"(citation_lost / possible hallucination — discard or re-research), could not be fetched " +
			"(fetch_error — consider reverifiable:false with a human attestation), or was structurally " +
			"incomplete (citation_invalid). None was registered.",
	})
	if err != nil {
		return err
	}
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// refreshOnConflict (bugs_open/184, the bugs_closed/091 class). The key is per
	// SITE; the finding is a LIST of rejected candidates that differs every run. Under
	// the default policy a second research pass on a site whose earlier item is still
	// open wrote nothing, and the open row went on naming the FIRST batch of rejects —
	// the live row here has described oufe.com's single 2026-07-26 reject ever since,
	// through every pass after it.
	//
	// The judgement 091's council asked to see confirmed rather than assumed — that a
	// refresh can rewrite an item under a human mid-read — is WEAKER here, and the
	// evidence is NARROWER than I first wrote it. I claimed `needs_human_review` "has no
	// working surface at all"; that is FALSE as a general statement and the council's
	// prior_art_librarian seat was right to refuse it as an absence sourced from another
	// bug file. Measured 2026-08-03: of 428 rows at that status fleet-wide, **86 have
	// been claimed and 55 handled**, the newest claim on 2026-08-02.
	//
	// What IS true, and is the actual warrant: of the four item types that opt into
	// refreshOnConflict (stale_evidence, citation_unverified,
	// directory_citation_unverified, stale_directory_claim) — 8 rows all told — **not
	// one has ever been claimed or handled.** So there is no reader of THESE to disturb,
	// while what the old behaviour protected is a description that is already wrong.
	//
	// The distinction is not pedantry: it means a HITL surface already works some of this
	// queue, so `needs_human_review` becoming a held status (BATCH-005's open question) is
	// nearer than "there is no surface" implied.
	w, err := writeWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "research",
		pipeline:     "content",
		itemType:     "citation_unverified",
		severity:     "medium",
		summary:      "Citation acquisition: candidate claims failed live verification and need a human ruling",
		spec:         string(specJSON),
		priority:     35,
		handlerAgent: "human-review",
		status:       "needs_human_review",
		createdBy:    params.AgentType,
		itemKey:      "citation_unverified:" + siteID.String(),
	}, refreshOnConflict, logger)
	if err != nil {
		return err
	}
	// A DELIBERATE STRING LITERAL, because a Go COMMENT cannot be pod-grepped and
	// this change adds no other literal — the council's debug_historian seat asked
	// how the new wiring would be confirmed in a running binary and the honest answer
	// was "it cannot". It is also the only record this emitter makes of its own
	// outcome: the row is otherwise the sole evidence it ran.
	logger.Info("HITL citation-reject item written (bugs_open/184)",
		zap.String("item_type", "citation_unverified"),
		zap.Bool("inserted", w.Inserted),
		zap.Bool("refreshed", w.Refreshed))
	return tx.Commit()
}
