// brief-negation-check — the SOURCE-side detector for bugs_open/305.
//
// WHAT 305 IS. Copy that says what a thing is NOT in order to say what it is:
// "The registry shows you what's possible, not what survives production." The
// owner read three live pages, quoted two such sentences, and ruled that this
// sort of copy must never leave the framework again.
//
// WHAT THIS CHECKS, AND WHY IT IS NOT THE SAME AS CHECKING THE PAGES. The
// writer-seam gate (rewrite_negations) repairs what the WRITER invents, and
// deliberately leaves alone anything the site's own brief handed it — because a
// site's voice specification outranks the fleet rules, and rewriting a brief's
// own words would be the platform overruling its owner. That leaves exactly one
// gap, and it is the one the owner's own complaint came through: a brief that
// SUPPLIES the construction. Measured 2026-08-19 on the complained-of site: its
// canonical tagline reaches the writer in 1,369 rendered prompts and comes back
// in 409 responses. The gate will count that and pass it; only a human editing
// the brief can fix it. This check is what puts it in front of a human.
//
// THE ONE RULE IT ENFORCES ON ITSELF: MEASURE THE TEXT THAT REACHES THE WRITER.
// Not the spec document. The owning lane published a fleet census over
// `data::text` on 2026-08-19 and withdrew it within hours — the writer reads a
// handful of named template fields, which on the worst site is ~23% of the
// document, so ~77% of that count was aimed at text with no consumer and the
// headline figure was roughly double. The visible surface here is DERIVED AT
// RUNTIME from the live agent config (the `{{.site_specs.specs.…}}` references
// in the writer's own prompt), never hardcoded, so a config change moves it and
// a copied list cannot go stale.
//
// SUPPLIED vs INSTRUCTIONAL, kept apart rather than added up. A brief that
// QUOTES a phrase, or lists one in a field the prompt injects verbatim, is
// handing the writer text: that class has a measured transfer chain. A brief
// that uses a contrast to give guidance ("use stack references naturally, not as
// buzzwords") is instructing, and whether that transfers into output is
// [UNMEASURED] — the question is open in two lanes and its test has to be
// designed to come out either way. Conflating them is exactly what made the
// withdrawn census wrong, so only the SUPPLIED class files a finding; the
// instructional count is reported and never acted on.
//
// EXIT CODES. 0 = ran, nothing supplied. 1 = findings (the Job shows failed,
// as its sibling checks do). 2 = could not run — a refusal, never a pass. An
// empty writer-visible surface, or a census that returns no sites, exits 2:
// this estate has been bitten by a check whose clean result meant it was blind.
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	_ "github.com/lib/pq"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

const (
	// writerAgent is the consumer whose prompt DEFINES the visible surface.
	writerAgent = "page-content-writer"

	// gapItemType is the item_type this check files. Classified in
	// discovery_checks/verifier_coverage_test.go in the same commit, per that
	// guard's standing obligation.
	gapItemType = "brief_supplies_negation"

	createdBy = "brief-negation-check"
)

// mandateRe marks a block that ORDERS a phrase onto pages rather than merely
// modelling it. On the complained-of site the emphasis block commands the
// tagline into "the homepage hero, services page hero, site footer, and meta
// descriptions" — that is the strongest evidence a brief can carry, so it is
// ranked highest and said plainly in the finding.
var mandateRe = regexp.MustCompile(`(?i)\b(must appear|must be used|always use|should appear|use this (?:exact|canonical)|canonical tagline|verbatim|every page)\b`)

// quotedRe finds spans a brief has put in quotes — the shape of a handover.
//
// ⚠ THE APOSTROPHE IS THE TRAP, and it produced junk on the owning lane's first
// fleet run: a naive ['"] class reads the ' in "the client's own voice" as an
// opening quote and returns everything up to the next apostrophe as a "supplied
// phrase". Single quotes are honoured only when the opening mark is not preceded
// by a letter and the closing mark is not followed by one — which is exactly
// what separates a quotation from a possessive. Go RE2 has no lookbehind, so the
// guard characters are captured and trimmed.
var quotedRe = regexp.MustCompile(`["“”]([^"“”]{12,220})["“”]|(^|[^\pL])['‘]([^'‘’]{12,220})['’]([^\pL]|$)`)

// blockSplitRe splits a `formatted` brief back into the labelled blocks
// FormatContentDirection wrote, which are joined by a blank line.
var blockSplitRe = regexp.MustCompile(`\n\s*\n`)

// templateRefRe pulls {{.site_specs.specs.<path>}} out of the live agent config.
var templateRefRe = regexp.MustCompile(`\.site_specs\.specs\.([A-Za-z0-9_.]+)`)

var psqlArgv = []string{"kubectl", "exec", "-n", "ai-persona-system", "postgres-clients-0", "--",
	"psql", "-U", "clients_user", "-d", "clients_db", "-tAc"}

func dbConn() (*sql.DB, error) {
	host := os.Getenv("PG_CLIENTS_HOST")
	if host == "" {
		return nil, nil // terminal route: kubectl exec, read-only
	}
	pw := os.Getenv("CLIENTS_DB_PASSWORD")
	if pw == "" {
		return nil, fmt.Errorf("PG_CLIENTS_HOST is set but CLIENTS_DB_PASSWORD is not — refusing to guess a connection")
	}
	port := os.Getenv("PG_CLIENTS_PORT")
	if port == "" {
		port = "5432"
	}
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=clients_user password=%s dbname=clients_db sslmode=disable", host, port, pw))
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func queryOne(db *sql.DB, q string) (string, error) {
	if db != nil {
		var raw sql.NullString
		if err := db.QueryRow(q).Scan(&raw); err != nil {
			return "", err
		}
		return raw.String, nil
	}
	out, err := exec.Command(psqlArgv[0], append(psqlArgv[1:], q)...).Output()
	if err != nil {
		return "", fmt.Errorf("via kubectl: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ---------------------------------------------------------------------------
// The visible surface — derived, never hardcoded.
// ---------------------------------------------------------------------------

const writerConfigSQL = `SELECT COALESCE(string_agg(default_config::text, ' '), '')
                           FROM agent_definitions
                          WHERE type = '` + writerAgent + `' AND is_active
                            AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL`

// visibleSurface returns the spec paths the writer's prompt actually reads, e.g.
// content_direction.formatted, identity.key_differentiators.
//
// An `{{if .site_specs.specs.content_direction}}` guard names the ASPECT rather
// than a field; it is dropped when a deeper path under the same aspect is also
// present, or the whole document would be counted through the back door — which
// is the very error this check exists not to repeat.
func visibleSurface(config string) []string {
	seen := map[string]bool{}
	for _, m := range templateRefRe.FindAllStringSubmatch(config, -1) {
		seen[m[1]] = true
	}
	deeper := map[string]bool{}
	for p := range seen {
		if i := strings.IndexByte(p, '.'); i > 0 {
			deeper[p[:i]] = true
		}
	}
	var out []string
	for p := range seen {
		if strings.Contains(p, ".") || !deeper[p] {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// ---------------------------------------------------------------------------
// The census.
// ---------------------------------------------------------------------------

type siteSpecs struct {
	Domain string                 `json:"domain"`
	SiteID string                 `json:"site_id"`
	Aspect string                 `json:"aspect"`
	Data   map[string]interface{} `json:"data"`
}

func censusSQL(aspects []string) string {
	quoted := make([]string, 0, len(aspects))
	for _, a := range aspects {
		quoted = append(quoted, "'"+strings.ReplaceAll(a, "'", "''")+"'")
	}
	return `SELECT COALESCE(jsonb_agg(jsonb_build_object(
	          'domain', s.domain, 'site_id', s.id::text, 'aspect', sp.aspect, 'data', sp.data)), '[]'::jsonb)
	          FROM site_specs sp JOIN sites s ON s.id = sp.site_id
	         WHERE sp.is_current AND sp.aspect IN (` + strings.Join(quoted, ",") + `)`
}

// ---------------------------------------------------------------------------
// The predicate — pure, and the part the tests hold.
// ---------------------------------------------------------------------------

// suppliedPhrase is one phrase a brief hands the writer that carries the
// construction.
type suppliedPhrase struct {
	Field    string `json:"field"`
	Route    string `json:"route"` // quoted | list_item
	Phrase   string `json:"phrase"`
	Shape    string `json:"shape"`
	Mandated bool   `json:"mandated"`
}

// siteAssessment is one site's answer.
type siteAssessment struct {
	Domain        string           `json:"domain"`
	SiteID        string           `json:"site_id"`
	VisibleChars  int              `json:"visible_chars"`
	DocChars      int              `json:"document_chars"`
	Supplied      []suppliedPhrase `json:"supplied"`
	Instructional int              `json:"instructional_only"`
	Regulatory    int              `json:"regulatory_left_alone"`
}

func (a siteAssessment) Finding() bool { return len(a.Supplied) > 0 }

func (a siteAssessment) Mandated() int {
	n := 0
	for _, s := range a.Supplied {
		if s.Mandated {
			n++
		}
	}
	return n
}

// quotedSpans returns every quoted span, possessives excluded.
func quotedSpans(text string) []string {
	var out []string
	for _, m := range quotedRe.FindAllStringSubmatch(text, -1) {
		if m[1] != "" {
			out = append(out, m[1])
		} else if m[3] != "" {
			out = append(out, m[3])
		}
	}
	return out
}

// assessValue walks one visible field and separates what the brief HANDS OVER
// from what it merely says.
//
// Two mechanically distinct handover routes:
//
//	quoted    — the author put quote marks round it inside prose. That is a
//	            handover in anybody's reading, and it is the shape of the one
//	            proven transfer chain (a canonical tagline, 1,369 prompts → 409
//	            outputs).
//	list_item — the field is a list the writer's template injects verbatim
//	            (identity.key_differentiators is rendered as its own text), so
//	            every element is supplied by construction and no quote marks are
//	            involved.
//
// regulatorySupplied reports whether a supplied phrase is a required disclosure
// rather than a house mannerism.
//
// It asks datahelpers.NegationExempt with NO brief text, which is exactly the
// "regulatory" arm of the rule the writer-seam gate applies to output. Using the
// same function on both sides is the point: a phrase the gate refuses to rewrite
// because it is a required disclosure must not be filed here as a defect to fix,
// or the two halves send a human in opposite directions — and the direction this
// one would send them is "edit a compliance statement".
func regulatorySupplied(phrase string) bool {
	for _, h := range datahelpers.ScanDefineByNegation(phrase) {
		if ok, why := datahelpers.NegationExempt(h, nil); ok && why == "regulatory" {
			return true
		}
	}
	return false
}

func assessValue(field string, v interface{}, mandated bool, out *[]suppliedPhrase, instructional *int, regulatory *int) {
	switch t := v.(type) {
	case string:
		// ⚠ MANDATE IS TESTED PER BLOCK, NOT PER FIELD, and the difference is not
		// cosmetic. `formatted` is one string of ~15,000 characters holding every
		// labelled block of the brief, joined by blank lines. Testing the whole
		// string means a single "must appear" anywhere in the document marks
		// EVERY phrase in it as mandated — measured on the first live run, where
		// a block describing sentence rhythm was reported as ordering its example
		// onto pages. MANDATED is the strongest thing this check says; it has to
		// mean "the block that hands this phrase over also orders it".
		for _, block := range blockSplitRe.Split(t, -1) {
			m := mandated || mandateRe.MatchString(block)
			for _, q := range quotedSpans(block) {
				hits := datahelpers.ScanDefineByNegation(q)
				if len(hits) == 0 {
					continue
				}
				if regulatorySupplied(q) {
					*regulatory++
					continue
				}
				*out = append(*out, suppliedPhrase{Field: field, Route: "quoted",
					Phrase: strings.TrimSpace(q), Shape: hits[0].Shape, Mandated: m})
			}
		}
		// Everything else in this string that carries the construction is the
		// brief instructing, not handing over. Counted, never filed.
		*instructional += len(datahelpers.ScanDefineByNegation(t)) - countInQuoted(t)
	case []interface{}:
		for _, e := range t {
			if s, ok := e.(string); ok {
				hits := datahelpers.ScanDefineByNegation(s)
				if len(hits) == 0 {
					continue
				}
				if regulatorySupplied(s) {
					*regulatory++
					continue
				}
				*out = append(*out, suppliedPhrase{Field: field, Route: "list_item",
					Phrase: strings.TrimSpace(s), Shape: hits[0].Shape,
					Mandated: mandated || mandateRe.MatchString(s)})
				continue
			}
			assessValue(field, e, mandated, out, instructional, regulatory)
		}
	case map[string]interface{}:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			assessValue(field+"."+k, t[k], mandated, out, instructional, regulatory)
		}
	}
}

func countInQuoted(text string) int {
	n := 0
	for _, q := range quotedSpans(text) {
		n += len(datahelpers.ScanDefineByNegation(q))
	}
	return n
}

func resolvePath(specs map[string]map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	aspect, ok := specs[parts[0]]
	if !ok {
		return nil
	}
	var cur interface{} = aspect
	for _, p := range parts[1:] {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return nil
		}
		cur, ok = m[p]
		if !ok {
			return nil
		}
	}
	return cur
}

func assess(rows []siteSpecs, surface []string) []siteAssessment {
	bySite := map[string]map[string]map[string]interface{}{}
	ids := map[string]string{}
	for _, r := range rows {
		if bySite[r.Domain] == nil {
			bySite[r.Domain] = map[string]map[string]interface{}{}
		}
		bySite[r.Domain][r.Aspect] = r.Data
		ids[r.Domain] = r.SiteID
	}
	domains := make([]string, 0, len(bySite))
	for d := range bySite {
		domains = append(domains, d)
	}
	sort.Strings(domains)

	var out []siteAssessment
	for _, d := range domains {
		a := siteAssessment{Domain: d, SiteID: ids[d]}
		for _, aspect := range bySite[d] {
			if b, err := json.Marshal(aspect); err == nil {
				a.DocChars += len(b)
			}
		}
		for _, path := range surface {
			v := resolvePath(bySite[d], path)
			if v == nil {
				continue
			}
			a.VisibleChars += len(flatten(v))
			assessValue(path, v, false, &a.Supplied, &a.Instructional, &a.Regulatory)
		}
		out = append(out, a)
	}
	return out
}

func flatten(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []interface{}:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, flatten(e))
		}
		return strings.Join(parts, " ")
	case map[string]interface{}:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(t))
		for _, k := range keys {
			parts = append(parts, flatten(t[k]))
		}
		return strings.Join(parts, " ")
	}
	return ""
}

// ---------------------------------------------------------------------------
// Writing.
// ---------------------------------------------------------------------------

func itemKey(siteID string) string { return "brief-negation:" + siteID }

// fileFinding writes one row per site. ON CONFLICT DO NOTHING against
// idx_swi_dedup: a daily re-run of an unfixed brief must not mint a second row.
func fileFinding(db *sql.DB, a siteAssessment) (bool, error) {
	spec, _ := json.Marshal(map[string]interface{}{
		"check":                 "brief_supplies_negation",
		"domain":                a.Domain,
		"supplied":              a.Supplied,
		"mandated_count":        a.Mandated(),
		"instructional_only":    a.Instructional,
		"regulatory_left_alone": a.Regulatory,
		"visible_chars":         a.VisibleChars,
		"document_chars":        a.DocChars,
		"fix": "Human decision, and it is the SITE OWNER's: edit the brief so it does not hand the writer " +
			"a phrase built on define-by-negation. The writer-seam gate deliberately exempts these, so no " +
			"code change will remove them. ⚠ Write the WHOLE content_direction object, not a patch " +
			"(bugs_open/327: a partial write shrinks the brief the writer reads), and verify by label " +
			"presence, never by diffing (formatted is regenerated in random key order).",
	})
	res, err := db.Exec(`
		INSERT INTO site_work_items
		  (site_id, source, pipeline, item_type, severity, summary, spec, priority,
		   handler_agent, status, max_attempts, created_by, item_key)
		VALUES ($1, 'discovery', 'content', $2, 'medium', $3, $4::jsonb, 45,
		        '', 'needs_human_review', 1, $5, $6)
		ON CONFLICT DO NOTHING`,
		a.SiteID, gapItemType,
		fmt.Sprintf("%s's brief hands the writer %d phrase(s) built on define-by-negation (%d mandated onto pages)",
			a.Domain, len(a.Supplied), a.Mandated()),
		string(spec), createdBy, itemKey(a.SiteID))
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// closeAnswered closes an open finding only on a POSITIVE re-observation: this
// run examined that site and found nothing supplied. An absence of evidence — a
// site missing from the census, a spec that failed to parse — never closes
// anything.
func closeAnswered(db *sql.DB, assessments []siteAssessment) ([]string, error) {
	rows, err := db.Query(`SELECT item_key FROM site_work_items
	                        WHERE item_type = $1 AND created_by = $2
	                          AND status NOT IN ('complete','cancelled','rejected')`,
		gapItemType, createdBy)
	if err != nil {
		return nil, err
	}
	open := map[string]bool{}
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err == nil {
			open[k] = true
		}
	}
	rows.Close()

	var closed []string
	for _, a := range assessments {
		if a.Finding() || !open[itemKey(a.SiteID)] {
			continue
		}
		if _, err := db.Exec(`UPDATE site_work_items
		                         SET status = 'complete', updated_at = now(),
		                             result = COALESCE(result,'{}'::jsonb) || $2::jsonb
		                       WHERE item_key = $1 AND item_type = $3 AND created_by = $4
		                         AND status NOT IN ('complete','cancelled','rejected')`,
			itemKey(a.SiteID), `{"closed_by":"brief-negation-check","reason":"re-examined: the writer-visible brief supplies no define-by-negation phrase"}`,
			gapItemType, createdBy); err != nil {
			return closed, err
		}
		closed = append(closed, a.Domain)
	}
	return closed, nil
}

// writeDocNote records EVERY run, clean or not: a check that only speaks when it
// fails is indistinguishable from one that has stopped running, so a MISSING row
// means the job did not run rather than "nothing is wrong".
func writeDocNote(db *sql.DB, body string) {
	if db == nil {
		return
	}
	if _, err := db.Exec(`INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
	                      VALUES ('pipeline', 'copy-quality', $1, '["copy-quality"]'::jsonb, $2)`,
		body, createdBy); err != nil {
		fmt.Fprintf(os.Stderr, "(could not write doc_notes: %v)\n", err)
	}
}

func render(assessments []siteAssessment, surface []string, filed, closed []string, dryRun, noWrites bool) string {
	var b strings.Builder
	b.WriteString("# brief-negation-check — briefs that HAND the writer a define-by-negation phrase (bugs_open/305)\n\n")
	b.WriteString("Writer-visible surface, derived from the live `" + writerAgent + "` config (never hardcoded):\n")
	for _, s := range surface {
		b.WriteString("  - " + s + "\n")
	}
	if dryRun {
		b.WriteString("\n> DRY RUN: nothing written.\n")
	}
	if noWrites {
		b.WriteString("\n> ⚠ READ-ONLY RUN: `PG_CLIENTS_HOST` unset, so this went through `kubectl exec`; no work item, no close-out and no doc_note were written.\n")
	}
	findings := 0
	b.WriteString("\n| site | supplied | mandated | regulatory (left alone) | instructional only | visible chars | document chars |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, a := range assessments {
		if a.Finding() {
			findings++
		}
		fmt.Fprintf(&b, "| %s | **%d** | %d | %d | %d | %d | %d |\n",
			a.Domain, len(a.Supplied), a.Mandated(), a.Regulatory, a.Instructional, a.VisibleChars, a.DocChars)
	}
	fmt.Fprintf(&b, "\n%d of %d sites hand the writer at least one such phrase.\n", findings, len(assessments))
	for _, a := range assessments {
		if !a.Finding() {
			continue
		}
		b.WriteString("\n## " + a.Domain + "\n")
		for _, s := range a.Supplied {
			tag := "supplied/" + s.Route
			if s.Mandated {
				tag = "**MANDATED**"
			}
			fmt.Fprintf(&b, "- %s [`%s`, %s] %q\n", tag, s.Field, s.Shape, datahelpers.TruncateString(s.Phrase, 160))
		}
	}
	b.WriteString("\n### What this does and does NOT establish\n")
	b.WriteString("- **Supplied** is the evidenced class: a phrase the brief hands over transfers verbatim " +
		"(measured: one site's canonical tagline, 1,369 rendered prompts → 409 responses).\n")
	b.WriteString("- **Instructional only** is counted and NOT acted on. Whether a brief's instructional " +
		"contrasts transfer into output is [UNMEASURED]; adding the two together is what made an earlier " +
		"fleet census wrong by roughly 2x.\n")
	b.WriteString("- **Regulatory** phrases are counted and LEFT ALONE, by the same rule the writer-seam gate " +
		"applies to output (`datahelpers.NegationExempt`): \"not financial advice\", \"we are not the FCA\", " +
		"\"not a quote or offer\". They are required disclosures, and a check that filed them would be sending " +
		"a human to edit a compliance statement.\n")
	b.WriteString("- Everything here is measured over the text the WRITER READS, never the spec document.\n")
	if len(filed) > 0 {
		b.WriteString("\nFiled: " + strings.Join(filed, ", ") + "\n")
	}
	if len(closed) > 0 {
		b.WriteString("Closed (re-examined and clean): " + strings.Join(closed, ", ") + "\n")
	}
	return b.String()
}

func main() {
	dryRun := flag.Bool("dry-run", false, "print the report; write nothing")
	emitJSON := flag.Bool("emit-json", false, "machine-readable assessments on stdout")
	flag.Parse()

	db, err := dbConn()
	if err != nil {
		fmt.Fprintf(os.Stderr, "REFUSING: %v\n", err)
		os.Exit(2)
	}
	if db != nil {
		defer db.Close()
	}

	cfg, err := queryOne(db, writerConfigSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "REFUSING: could not read the writer config: %v\n", err)
		os.Exit(2)
	}
	surface := visibleSurface(cfg)
	if len(surface) == 0 {
		fmt.Fprintln(os.Stderr, "REFUSING: the writer's prompt references no site_specs fields — "+
			"that is a broken read or a changed prompt, not a clean fleet. A zero here would report "+
			"every brief as clean while looking at nothing.")
		os.Exit(2)
	}
	aspects := map[string]bool{}
	var aspectList []string
	for _, p := range surface {
		a := strings.Split(p, ".")[0]
		if !aspects[a] {
			aspects[a] = true
			aspectList = append(aspectList, a)
		}
	}
	sort.Strings(aspectList)

	raw, err := queryOne(db, censusSQL(aspectList))
	if err != nil {
		fmt.Fprintf(os.Stderr, "REFUSING: census failed: %v\n", err)
		os.Exit(2)
	}
	var rows []siteSpecs
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		fmt.Fprintf(os.Stderr, "REFUSING: census would not decode: %v\n", err)
		os.Exit(2)
	}
	if len(rows) == 0 {
		fmt.Fprintln(os.Stderr, "REFUSING: no current site specs at all — the fleet is never empty, "+
			"so this is a broken read, not a clean result")
		os.Exit(2)
	}

	assessments := assess(rows, surface)
	var filed, closed []string
	if !*dryRun && db != nil {
		for _, a := range assessments {
			if !a.Finding() {
				continue
			}
			created, err := fileFinding(db, a)
			if err != nil {
				fmt.Fprintf(os.Stderr, "REFUSING: could not file %s: %v\n", a.Domain, err)
				os.Exit(2)
			}
			if created {
				filed = append(filed, a.Domain)
			}
		}
		closed, err = closeAnswered(db, assessments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "REFUSING: close-out failed: %v\n", err)
			os.Exit(2)
		}
	}

	report := render(assessments, surface, filed, closed, *dryRun, db == nil)
	fmt.Print(report)
	if *emitJSON {
		b, _ := json.MarshalIndent(assessments, "", "  ")
		fmt.Println(string(b))
	}
	if !*dryRun && db != nil {
		writeDocNote(db, report)
	}
	for _, a := range assessments {
		if a.Finding() {
			os.Exit(1)
		}
	}
}
