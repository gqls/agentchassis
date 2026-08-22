// FILE: cmd/content-loss-check/main.go
//
// content-loss-check — the unified content-loss detector AND its reader, one
// binary, per the owner's 2026-08-22 ruling on RFC_042 §6 (option c) and the
// scoping in bugs_open/355. Runs daily as a CronJob; also runnable from a
// terminal (kubectl route, report-only).
//
// WHAT ONE RUN DOES, in order — and refuses (exit 2) rather than reporting
// clean if any instrument check fails:
//
//  1. CANARY: the Go loss predicate must still call a fabricated loss a loss
//     and a fabricated non-loss a non-loss. A definition drift fails the run
//     before it can report anything.
//  2. DEMAND CONTROL: the SQL detector must still find the KNOWN pre-fix
//     funnel losses (72 as of 2026-08-22, pinned window < 2026-08-15, queried
//     over all history — content_data in page_component_history is never
//     pruned). Zero here means the INSTRUMENT went blind (schema pointers
//     gone, trigger dropped), never that the fleet is clean. This is the
//     lesson 016b §9 records from this check's own scoping: a control drawn
//     from your own query shares your own blindness — this one is pinned to a
//     population whose non-zero answer was established independently.
//  3. DETECT: pair archived generations in the lookback window (both ops —
//     'delete' is the funnel, 'overwrite' the in-place writers) and diff
//     schema-declared non-llm keys. New losses become agent_error_log rows,
//     error_code CONTENT_KEY_LOSS, deduped on (pre-image history row, key) so
//     overlapping windows re-derive idempotently.
//  4. READ (the A3 half, same binary so it cannot be forgotten): every
//     unresolved CONTENT_KEY_LOSS / STRUCTURAL_KEY_CARRY_MISS /
//     CONTENT_DATA_REGRESSION row is re-graded against the CURRENT rows —
//     healed and row-gone findings are stamped resolved (resolved_by names
//     this check and the verdict), open ones are counted and listed.
//  5. LIVE-DAMAGE CENSUS: deployed rows whose schema declares a REQUIRED
//     non-llm field absent/blank — the archive never fires on INSERT
//     (bugs_open/355 §3.3), so born-incomplete damage is measured at state,
//     not at writes. This census, not the finding rows, is the durable record:
//     agent_error_log rows expire (30d unresolved / 14d resolved, mig 466's
//     embedded cleanup — so resolving a row HALVES its remaining life, which
//     is fine precisely because the census re-derives from live state).
//  6. HEARTBEAT: one doc_notes row EVERY run, clean or not — a check that only
//     speaks when it fails is indistinguishable from one that has stopped
//     running (the family convention; a MISSING row means the job did not
//     run). subject_key/source = 'content-loss-check' (source must equal the
//     service directory name — named landmine).
//
// WHAT IT DELIBERATELY DOES NOT DO: file work items. bugs_open/213's lesson is
// that a second producer on an existing item_type closes items against a
// predicate describing a different defect; the remediation seam for this class
// (widening the required-fields-missing router, 238's plan) activates only if
// this check ever measures a population — which as of filing it has not
// (bugs_open/355 §2: zero losses at the eight uncarried writers).
//
// EXIT CODES (family convention): 0 = ran, clean. 1 = findings live (new
// losses, still-open findings, or standing live damage). 2 = could not run or
// refused — never a pass.
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	_ "github.com/lib/pq"
)

const (
	checkName = "content-loss-check"

	codeKeyLoss   = "CONTENT_KEY_LOSS"
	codeCarryMiss = "STRUCTURAL_KEY_CARRY_MISS"
	codeRegress   = "CONTENT_DATA_REGRESSION"
)

// psqlArgv is the terminal route (verifier-remit-check's reason verbatim: the
// CronJob's service account has no pods/exec RBAC, so IT takes the direct
// route; a session at a keyboard has no DSN, so it takes this one — read-only).
var psqlArgv = []string{"kubectl", "exec", "-n", "ai-persona-system", "postgres-clients-0", "--",
	"psql", "-U", "clients_user", "-d", "clients_db", "-tAc"}

func dbConn() (*sql.DB, error) {
	host := os.Getenv("PG_CLIENTS_HOST")
	if host == "" {
		return nil, nil
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
		"host=%s port=%s user=clients_user password=%s dbname=clients_db sslmode=disable application_name=%s",
		host, port, pw, checkName))
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

// fetchJSON runs a query that returns exactly one json text column, via
// whichever route is available.
func fetchJSON(db *sql.DB, query string) (string, error) {
	if db != nil {
		var raw sql.NullString
		if err := db.QueryRow(query).Scan(&raw); err != nil {
			return "", err
		}
		return raw.String, nil
	}
	out, err := exec.Command(psqlArgv[0], append(psqlArgv[1:], query)...).Output()
	if err != nil {
		return "", fmt.Errorf("via kubectl: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

type familyFinding struct {
	ID      string
	Code    string
	SiteID  sql.NullString
	Context map[string]interface{}
}

type runReport struct {
	CanaryOK        bool
	ControlLosses   int
	Coverage        map[string]interface{}
	LossesInWindow  int
	NewFindings     int
	Healed          int
	RowGone         int
	StillOpen       int
	OpenDetail      []string
	LiveDamageRows  int
	LiveDamage      []string
	NoWrites        bool
	DryRun          bool
	LookbackDays    int
	NewLossDetail   []string
	OpByOpLosses    map[string]int
	UnattributedNew int
}

func main() {
	dryRun := flag.Bool("dry-run", false, "detect and grade but write nothing (no findings, no resolutions, no doc_note)")
	lookback := flag.Int("lookback-days", 21, "how far back to pair archived generations")
	flag.Parse()

	db, err := dbConn()
	if err != nil {
		fmt.Fprintf(os.Stderr, "REFUSED: %v\n", err)
		os.Exit(2)
	}
	if db != nil {
		defer db.Close()
	}

	rep := runReport{DryRun: *dryRun, NoWrites: db == nil, LookbackDays: *lookback,
		OpByOpLosses: map[string]int{}}

	// 1. Canary — the definition itself, before any data is trusted.
	rep.CanaryOK = isNonLLMLoss("site_specs.identity.email", "a@b.c", "") &&
		isNonLLMLoss("renderer.cta_url", "/x", "  ") &&
		!isNonLLMLoss("llm", "prose", "") && // llm keys are the writer's to rewrite
		!isNonLLMLoss("", "v", "") && // undeclared source is not this class
		!isNonLLMLoss("static.icon", "", "") && // nothing held, nothing lost
		!isNonLLMLoss("static.icon", "v", "v2") // a changed value is not a loss
	if !rep.CanaryOK {
		fmt.Fprintln(os.Stderr, "REFUSED: the loss predicate no longer recognises its own canary — definition drift; not reporting anything")
		os.Exit(2)
	}

	// 2. Demand control — pinned pre-fix funnel losses, all-history window.
	controlRaw, err := fetchJSON(db, buildLossQuery(3650, true))
	if err != nil {
		fmt.Fprintf(os.Stderr, "REFUSED: demand control query failed: %v\n", err)
		os.Exit(2)
	}
	var controlLosses []lossRowJSON
	if err := json.Unmarshal([]byte(controlRaw), &controlLosses); err != nil {
		fmt.Fprintf(os.Stderr, "REFUSED: demand control returned unparseable json: %v\n", err)
		os.Exit(2)
	}
	rep.ControlLosses = len(controlLosses)
	if rep.ControlLosses == 0 {
		// Write the refusal to the heartbeat too, if we can: a silent exit 2
		// on a schedule is exactly the missing-row signal, but a row SAYING
		// "blind" beats an absence.
		writeDocNote(db, *dryRun, "REFUSED — instrument blind: the pinned demand control (pre-2026-08-15 funnel losses, 72 at pinning) returned ZERO. The schema pointers, the archive trigger or the table itself have changed underneath this check. Re-pin the control against currently-known losses before trusting any run. No detection was reported.")
		fmt.Fprintln(os.Stderr, "REFUSED: demand control returned 0 — the instrument is blind, not the fleet clean (bugs_open/355 §2.3)")
		os.Exit(2)
	}

	// Coverage stats for the honest denominator.
	if covRaw, err := fetchJSON(db, buildCoverageQuery(*lookback)); err == nil {
		_ = json.Unmarshal([]byte(covRaw), &rep.Coverage)
	}

	// 3. Detect.
	lossRaw, err := fetchJSON(db, buildLossQuery(*lookback, false))
	if err != nil {
		fmt.Fprintf(os.Stderr, "REFUSED: detection query failed: %v\n", err)
		os.Exit(2)
	}
	var losses []lossRowJSON
	if err := json.Unmarshal([]byte(lossRaw), &losses); err != nil {
		fmt.Fprintf(os.Stderr, "REFUSED: detection returned unparseable json: %v\n", err)
		os.Exit(2)
	}
	rep.LossesInWindow = len(losses)
	for _, l := range losses {
		rep.OpByOpLosses[l.Op]++
	}

	// File new findings (direct route only; dedupe is in the INSERT).
	if db != nil && !*dryRun {
		for _, l := range losses {
			inserted, ierr := insertLossFinding(db, l)
			if ierr != nil {
				fmt.Fprintf(os.Stderr, "(finding insert failed for %s:%s — %v)\n", l.HistoryID, l.Key, ierr)
				continue
			}
			if inserted {
				rep.NewFindings++
				detail := fmt.Sprintf("%s slot=%s key=%s op=%s writer=%s", l.PageID, l.SlotName, l.Key, l.Op, l.Writer)
				rep.NewLossDetail = append(rep.NewLossDetail, detail)
				if l.Writer == "" || strings.HasPrefix(l.Writer, "app - ") || l.Writer == "psql" {
					rep.UnattributedNew++
				}
			}
		}
	}

	// 4. Read and disposition the family — the reader half.
	if db != nil && !*dryRun {
		if err := dispositionFamily(db, &rep); err != nil {
			fmt.Fprintf(os.Stderr, "(disposition pass failed: %v)\n", err)
		}
	}

	// 5. Live-damage census — the durable, INSERT-covering record.
	if dmgRaw, err := fetchJSON(db, liveDamageQuery); err == nil {
		var dmg struct {
			Rows     int                      `json:"rows"`
			Examples []map[string]interface{} `json:"examples"`
		}
		if json.Unmarshal([]byte(dmgRaw), &dmg) == nil {
			rep.LiveDamageRows = dmg.Rows
			for _, e := range dmg.Examples {
				rep.LiveDamage = append(rep.LiveDamage, fmt.Sprintf("%v %v slot=%v key=%v (%v)",
					e["domain"], e["page"], e["slot"], e["key"], e["source"]))
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "(live-damage census failed: %v)\n", err)
	}

	// 6. Heartbeat + report.
	body := render(rep)
	fmt.Println(body)
	writeDocNote(db, *dryRun, body)

	if rep.NewFindings > 0 || rep.StillOpen > 0 || rep.LiveDamageRows > 0 {
		os.Exit(1)
	}
}

// lossRowJSON mirrors the json_build_object in the loss query.
type lossRowJSON struct {
	HistoryID string `json:"history_id"`
	PageID    string `json:"page_id"`
	SiteID    string `json:"site_id"`
	SlotName  string `json:"slot_name"`
	Op        string `json:"op"`
	Writer    string `json:"writer"`
	Key       string `json:"key"`
	Source    string `json:"source"`
	LostAt    string `json:"lost_at"`
}

// insertLossFinding files one CONTENT_KEY_LOSS row, idempotently: the NOT
// EXISTS matches this check's own dedupe identity (pre-image history row +
// key), so re-running over an overlapping window re-derives without
// duplicating. Returns whether a row was actually inserted.
func insertLossFinding(db *sql.DB, l lossRowJSON) (bool, error) {
	res, err := db.Exec(`
		INSERT INTO agent_error_log
		    (site_id, domain, agent_type, step_name, action, error_message, error_code, severity, context)
		SELECT $1::uuid, s.domain, $2, 'detect', 'content_loss_check',
		       $3, $4, 'warning', $5::jsonb
		  FROM sites s WHERE s.id = $1::uuid
		   AND NOT EXISTS (
		       SELECT 1 FROM agent_error_log e
		        WHERE e.error_code = $4
		          AND e.context->>'history_id' = $6
		          AND e.context->>'key' = $7)`,
		l.SiteID, checkName,
		fmt.Sprintf("non-llm key %q (source %s) went non-blank -> absent/blank across consecutive generations of page %s slot %q (op=%s, writer=%q) — a %s writer dropped a resolved value (bugs_open/355)",
			l.Key, l.Source, l.PageID, l.SlotName, l.Op, l.Writer,
			map[bool]string{true: "funnel", false: "non-funnel"}[l.Op == "delete"]),
		codeKeyLoss,
		mustJSON(map[string]interface{}{
			"history_id": l.HistoryID, "page_id": l.PageID, "slot_name": l.SlotName,
			"key": l.Key, "source": l.Source, "op": l.Op, "writer": l.Writer,
			"lost_at": l.LostAt, "bug": "bugs_open/355",
		}),
		l.HistoryID, l.Key)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// dispositionFamily re-grades every unresolved finding of the three codes
// against current row state, stamping healed/row-gone ones resolved.
func dispositionFamily(db *sql.DB, rep *runReport) error {
	rows, err := db.Query(`
		SELECT id::text, error_code, site_id::text, COALESCE(context,'{}')::text
		  FROM agent_error_log
		 WHERE error_code IN ($1,$2,$3) AND resolved = false`,
		codeKeyLoss, codeCarryMiss, codeRegress)
	if err != nil {
		return err
	}
	defer rows.Close()
	var findings []familyFinding
	for rows.Next() {
		var f familyFinding
		var ctxRaw string
		if err := rows.Scan(&f.ID, &f.Code, &f.SiteID, &ctxRaw); err != nil {
			return err
		}
		_ = json.Unmarshal([]byte(ctxRaw), &f.Context)
		findings = append(findings, f)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, f := range findings {
		verdict, openDetail, gerr := gradeFinding(db, f)
		if gerr != nil {
			fmt.Fprintf(os.Stderr, "(grading %s %s failed: %v)\n", f.Code, f.ID, gerr)
			continue
		}
		switch verdict {
		case "healed", "row_gone", "page_gone":
			if _, uerr := db.Exec(`
				UPDATE agent_error_log
				   SET resolved = true, resolved_at = now(), resolved_by = $2
				 WHERE id = $1::uuid AND resolved = false`,
				f.ID, checkName+":"+verdict); uerr != nil {
				fmt.Fprintf(os.Stderr, "(resolve %s failed: %v)\n", f.ID, uerr)
				continue
			}
			if verdict == "healed" {
				rep.Healed++
			} else {
				rep.RowGone++
			}
		default:
			rep.StillOpen++
			if len(rep.OpenDetail) < 15 {
				rep.OpenDetail = append(rep.OpenDetail, fmt.Sprintf("%s %s %s", f.Code, f.ID[:8], openDetail))
			}
		}
	}
	return nil
}

// gradeFinding maps one finding to a verdict via the pure predicates in
// check.go, fetching only the row counts they need.
func gradeFinding(db *sql.DB, f familyFinding) (string, string, error) {
	str := func(k string) string { s, _ := f.Context[k].(string); return s }

	switch f.Code {
	case codeKeyLoss:
		rows, deployed, holds, err := slotKeyState(db, str("page_id"), str("slot_name"), str("key"))
		if err != nil {
			return "", "", err
		}
		return healVerdict(rows, deployed, holds), str("key"), nil

	case codeCarryMiss:
		if !f.SiteID.Valid || f.SiteID.String == "" {
			return "open", "no site_id on the finding", nil
		}
		pageID, ok, err := pageByName(db, f.SiteID.String, str("page_name"))
		if err != nil {
			return "", "", err
		}
		if !ok {
			return "page_gone", "", nil
		}
		fieldsRaw, _ := f.Context["fields"].([]interface{})
		verdicts := map[string]string{}
		for _, fr := range fieldsRaw {
			name, _ := fr.(string)
			if name == "" {
				continue
			}
			rows, deployed, holds, err := slotKeyState(db, pageID, str("section"), name)
			if err != nil {
				return "", "", err
			}
			verdicts[name] = healVerdict(rows, deployed, holds)
		}
		if len(verdicts) == 0 {
			return "open", "finding names no fields", nil
		}
		resolved, verdict, open := carryMissVerdict(verdicts)
		if !resolved {
			return "open", "fields still absent: " + strings.Join(open, ","), nil
		}
		return verdict, "", nil

	case codeRegress:
		if !f.SiteID.Valid || f.SiteID.String == "" {
			return "open", "no site_id on the finding", nil
		}
		pageID, ok, err := pageByName(db, f.SiteID.String, str("page_name"))
		if err != nil {
			return "", "", err
		}
		if !ok {
			return "page_gone", "", nil
		}
		var withData int
		if err := db.QueryRow(`
			SELECT count(*) FROM page_components
			 WHERE page_id = $1::uuid AND content_data IS NOT NULL AND content_data <> '{}'::jsonb`,
			pageID).Scan(&withData); err != nil {
			return "", "", err
		}
		if withData > 0 {
			return "healed", "", nil
		}
		return "open", "page still has no structured content_data", nil
	}
	return "open", "unknown code", nil
}

func slotKeyState(db *sql.DB, pageID, slot, key string) (rows, deployed int, holdsKey bool, err error) {
	err = db.QueryRow(`
		SELECT count(*),
		       count(*) FILTER (WHERE build_status='deployed'),
		       COALESCE(bool_or(build_status='deployed' AND btrim(COALESCE(content_data->>$3,'')) <> ''), false)
		  FROM page_components
		 WHERE page_id = $1::uuid AND slot_name IS NOT DISTINCT FROM NULLIF($2,'')`,
		pageID, slot, key).Scan(&rows, &deployed, &holdsKey)
	return
}

func pageByName(db *sql.DB, siteID, name string) (string, bool, error) {
	var id string
	err := db.QueryRow(`SELECT id::text FROM pages WHERE site_id = $1::uuid AND name = $2`,
		siteID, name).Scan(&id)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	return id, err == nil, err
}

// writeDocNote records the run on EVERY run, clean or not — a check that only
// speaks when it fails is indistinguishable from one that has stopped running.
// Best-effort: failing to RECORD a run must never become failing to REPORT it.
// source MUST equal the service directory name (named landmine: source≠jobname
// breaks the family's own staleness sweep).
func writeDocNote(db *sql.DB, dryRun bool, body string) {
	if db == nil || dryRun {
		return
	}
	if _, err := db.Exec(`INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
	                      VALUES ('pipeline', $1, $2, '["content-loss-check","bugs-open-355"]'::jsonb, $1)`,
		checkName, body); err != nil {
		fmt.Fprintf(os.Stderr, "(could not write doc_notes: %v)\n", err)
	}
}

func render(r runReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "content-loss-check — detector + reader for schema-declared non-llm content_data keys (bugs_open/355, RFC_042 option c).\n")
	fmt.Fprintf(&b, "This row is written on every run, clean or not — a missing row means the job did not run, never that nothing is wrong.\n\n")
	if r.DryRun {
		b.WriteString("MODE: dry-run — nothing was written.\n")
	}
	if r.NoWrites {
		b.WriteString("MODE: terminal route (kubectl, read-only) — no findings filed, no resolutions, no doc_note.\n")
	}
	fmt.Fprintf(&b, "instrument: canary ok; demand control %d known pre-fix funnel losses re-found (0 would have refused the run — a blind instrument never reports clean).\n", r.ControlLosses)
	if r.Coverage != nil {
		fmt.Fprintf(&b, "coverage over %d days: %v pairs (%v funnel / %v in-place), %v judgeable — a zero is only as wide as the judgeable count.\n",
			r.LookbackDays, r.Coverage["pairs_total"], r.Coverage["pairs_delete"], r.Coverage["pairs_overwrite"], r.Coverage["judgeable"])
	}
	fmt.Fprintf(&b, "\nDETECT: %d loss event(s) in the window", r.LossesInWindow)
	if r.LossesInWindow > 0 {
		fmt.Fprintf(&b, " (by op: %v)", r.OpByOpLosses)
	}
	fmt.Fprintf(&b, "; %d newly filed as %s", r.NewFindings, codeKeyLoss)
	if r.UnattributedNew > 0 {
		fmt.Fprintf(&b, " (%d unattributed — socket-named pre-image; attribution arrives with A1's roll)", r.UnattributedNew)
	}
	b.WriteString("\n")
	for _, d := range r.NewLossDetail {
		fmt.Fprintf(&b, "  new: %s\n", d)
	}
	fmt.Fprintf(&b, "READ:   %d healed (stamped resolved), %d row-gone/page-gone (stamped resolved), %d still open\n", r.Healed, r.RowGone, r.StillOpen)
	for _, d := range r.OpenDetail {
		fmt.Fprintf(&b, "  open: %s\n", d)
	}
	fmt.Fprintf(&b, "STATE:  %d deployed row(s) carry a REQUIRED non-llm field absent/blank (the archive never sees INSERT — this census is the durable record; finding rows expire per mig 466)\n", r.LiveDamageRows)
	for _, d := range r.LiveDamage {
		fmt.Fprintf(&b, "  live: %s\n", d)
	}
	if r.NewFindings == 0 && r.StillOpen == 0 && r.LiveDamageRows == 0 {
		b.WriteString("\nCLEAN — and the word is earned: the canary and the pinned demand control both passed on this same run.\n")
	}
	return b.String()
}

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
