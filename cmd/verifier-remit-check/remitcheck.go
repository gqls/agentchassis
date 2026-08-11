// verifier-remit-check — the CLASS detector for bugs_open/213.
//
// WHAT 213 WAS. `site_work_items.item_type` is the join between WHO FILED an item
// and WHAT PREDICATE re-grades it before it closes. The registry keys on that name
// alone (`verifiers[itemType]`), which silently assumes one item_type means one
// predicate. Two producers filed `hardcoded_section_colors`; only one wrote the
// verifier. The other producer's items were graded against a predicate describing a
// different defect — the verifier answered ITS OWN question correctly, returned
// Resolved:true, and the item closed `complete` with the defect untouched. Nothing
// errored. 11 of 11 second-producer items closed clean and not one ever failed to
// close, while every item that ever did fail belonged to the producer who wrote the
// verifier. The fix (WII-013) added `VerifierPolicy.Grades`, a remit a verifier
// declares so the gate can refuse an item it does not speak for.
//
// WHY THIS EXISTS ANYWAY. Grades is OPT-IN — deliberately, per the owner's
// 2026-08-02 shared-seam ruling, so it cannot change a type that has not asked for
// it. The council's `architecture` seat named the cost in the same round that
// approved it: "VerifierPolicy.Grades is opt-in — the NEXT converging producer on
// any of the other 10 verified item_types reproduces this exact bug unless a human
// remembers to write a Grades function, which is precisely the discipline that
// already failed once here." Owner ruling D3 (2026-08-11): build the detector.
//
// THE QUESTION IT ASKS, once a day, of live data:
//
//	does any item_type with a REGISTERED VERIFIER carry rows from more than one
//	PRODUCER SHAPE, while its verifier declares NO REMIT?
//
// IT KEYS ON THE ROW, NEVER ON A PRODUCER LIST — the same design decision as
// Grades itself, and for the same measured reason: any agent definition can file
// any item_type from DB config with no code change (bugs_open/213 §5.3, measured by
// the bugs_open/071 lane), so a code-side producer list is authoritative-looking and
// permanently behind live config. `created_by` cannot substitute either: it bottoms
// out at the literal `generic`, and [MEASURED 2026-08-11] it reads 2 or 3 distinct
// values on empty_section, literal_markdown and hardcoded_section_colors, so keying
// on it would fire on single-producer types. The `source` COLUMN is worse: it reads
// 2 on page_canonical_collision, which has exactly one producer.
//
// A PRODUCER SHAPE is the pair (spec.audit_source, the spec's top-level key set),
// clustered — see producerFamilies for why the clustering is the load-bearing part
// and what the threshold was measured against.
//
// WHAT IT CANNOT SEE, stated because a detector's blind spots belong next to its
// claims: a second producer that copies the first's spec shape exactly and writes no
// audit_source is invisible to any row-shaped test, and a convergence that has not
// yet filed a row is invisible to a data-driven one. This narrows the class; it does
// not close it.
//
// EXIT CODES. 0 = ran, no findings. 1 = findings (the Job shows failed, as
// single-owner-carriers-check does). 2 = could not run — a refusal, never a pass;
// an empty registry exits 2 as well, so a linking accident cannot read as a clean
// bill of health.
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	_ "github.com/lib/pq"
)

const (
	// gapItemType is the item_type this check files. Classified in
	// discovery_checks/verifier_coverage_test.go (itemTypesWithoutVerifiers AND
	// liveItemTypes) in the same commit that shipped this file, per that guard's
	// standing obligation.
	gapItemType = "verifier_remit_gap"

	// systemSiteID is the system.internal pseudo-site. site_id is NOT NULL and
	// this finding is about the fleet, not a site — diagnose_triage_action.go:39
	// established the anchor ("every needs_diagnosis anchors here; the real site
	// under diagnosis travels in the spec") and this reuses it rather than
	// inventing a second convention.
	systemSiteID = "eac60db8-b032-432b-b36d-76f37632045d"

	// createdBy identifies this check's own rows, so the close-out below can never
	// touch a row somebody else filed under the same type.
	createdBy = "verifier-remit-check"

	// familyOverlapThreshold is what makes two spec shapes ONE producer.
	//
	// It is not tuned; it sits in an empty band. [MEASURED 2026-08-11, live
	// clients_db, all 12 verified item_types] every same-producer key-set pair
	// overlaps 0.667 or more — the narrowest is page_canonical_collision, whose
	// probe rows share 2 of their 3 keys with the full-shape rows — and the one
	// genuine cross-producer pair overlaps 0.000: hardcoded_section_colors'
	// producer A ({check, components_found, …}) and producer B ({audit_source,
	// acceptance_test, …}) share NO key at all. Anything from just above 0 to
	// 0.667 gives today's answer; 0.5 is the middle of that gap.
	familyOverlapThreshold = 0.5
)

// censusRow is one (item_type, audit_source label, spec key-set) group, with the
// counts that make a finding readable. Times stay STRINGS: they are reported to a
// human and never computed with, and a parsed timestamp is one more format trap
// between two fetch routes for no gain.
type censusRow struct {
	ItemType string `json:"item_type"`
	Label    string `json:"label"`
	Keyset   string `json:"keyset"`
	Rows     int    `json:"row_count"`
	First    string `json:"first_seen"`
	Last     string `json:"last_seen"`
	SampleID string `json:"sample_id"`
}

// family is one producer's rows: the shapes that cluster together under one
// audit_source label.
type family struct {
	Label   string   `json:"label"`
	Shapes  []string `json:"shapes"`
	Rows    int      `json:"rows"`
	First   string   `json:"first_seen"`
	Last    string   `json:"last_seen"`
	Samples []string `json:"sample_item_ids"`
}

// assessment is one verified item_type's verdict.
type assessment struct {
	ItemType      string   `json:"item_type"`
	Rows          int      `json:"rows"`
	Shapeless     int      `json:"shapeless_rows"`
	Families      []family `json:"families"`
	DeclaresRemit bool     `json:"declares_remit"`
}

// Finding is true when BOTH halves hold: more than one producer family AND no
// declared remit. Either half alone is fine — a type with one producer needs no
// remit, and a type that declares one has answered the question.
func (a assessment) Finding() bool { return len(a.Families) > 1 && !a.DeclaresRemit }

// Suppressed is the POSITIVE CONTROL, and it is why the report names it. A run
// that says only "0 findings" cannot be told apart from one that looked at nothing
// (016b §9: a gate's 0 findings has two causes with opposite fixes). A run that
// says "hardcoded_section_colors has 2 producer families and is answered by a
// declared remit; 0 findings" has shown its working.
func (a assessment) Suppressed() bool { return len(a.Families) > 1 && a.DeclaresRemit }

// ---------------------------------------------------------------------------
// The census — one query, two fetch routes, ONE decoder.
// ---------------------------------------------------------------------------

// censusSQL is the one SQL text both fetch routes run, so they cannot drift in
// what they ask — only in how they connect (the fleetdb.go lesson: two hand-written
// paths to the same data go blind in the same direction and then agree with each
// other, bugs_open/144).
//
// IT TAKES NO PARAMETERS AND CONCATENATES NOTHING, which is the point. The first
// version filtered `item_type IN (<the registry>)` by interpolating a
// regex-validated list, because the kubectl route cannot bind `$1`. The council
// (constitution seat, corr fc082c4a) called that what it is: regex-validating a
// literal before interpolation is the classic workaround for parameterisation, not
// parameterisation. Censusing EVERY item_type and filtering in Go removes the
// question — assess() already iterates the registry, so the SQL-side filter was
// redundant defence — and it costs nothing measurable: one GROUP BY over ~6.5k rows
// returning ~150, against 17 before.
//
// The key-set is computed in SQL because it collapses the table to one row per
// (type, label, shape), so the binary never holds the work-item corpus in memory.
const censusSQL = `SELECT COALESCE(jsonb_agg(t), '[]'::jsonb)::text FROM (
  SELECT item_type,
         COALESCE(spec->>'audit_source','') AS label,
         COALESCE((SELECT string_agg(k, ',' ORDER BY k) FROM jsonb_object_keys(spec) k), '') AS keyset,
         count(*) AS row_count,
         min(created_at)::text AS first_seen,
         max(created_at)::text AS last_seen,
         min(id::text) AS sample_id
  FROM site_work_items
  GROUP BY 1,2,3
) t;`

func decodeCensus(raw []byte) ([]censusRow, error) {
	var out []censusRow
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("census JSON: %w", err)
	}
	return out, nil
}

// psqlArgv is the terminal route: a session at a keyboard reaches the database
// through kubectl exec. The CronJob CANNOT — the ai-persona-app service account has
// no pods/exec RBAC in this namespace, the constraint component-render-check,
// single-owner-carriers-check and bugs-open-staleness-sweep all hit — so it sets
// PG_CLIENTS_HOST and takes the direct route instead.
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
		"host=%s port=%s user=clients_user password=%s dbname=clients_db sslmode=disable", host, port, pw))
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func fetchCensus(db *sql.DB, query string) ([]censusRow, error) {
	if db != nil {
		var raw string
		if err := db.QueryRow(query).Scan(&raw); err != nil {
			return nil, fmt.Errorf("census query: %w", err)
		}
		return decodeCensus([]byte(raw))
	}
	out, err := exec.Command(psqlArgv[0], append(psqlArgv[1:], query)...).Output()
	if err != nil {
		return nil, fmt.Errorf("census via kubectl: %w", err)
	}
	return decodeCensus([]byte(strings.TrimSpace(string(out))))
}

// ---------------------------------------------------------------------------
// The predicate — pure, and the part the tests hold.
// ---------------------------------------------------------------------------

// overlap is the overlap coefficient |a ∩ b| / min(|a|,|b|), NOT Jaccard.
//
// Jaccard was measured and rejected: page_canonical_collision's two real shapes
// are 11 keys and 3 keys sharing 2, which is J=0.167 — a THIRD producer under any
// usable Jaccard threshold, and it has exactly one producer. The asymmetry is the
// point: a small shape that is almost entirely contained in a large one is a
// variant of it, not a rival to it.
func overlap(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	in := map[string]bool{}
	for _, k := range a {
		in[k] = true
	}
	shared := 0
	for _, k := range b {
		if in[k] {
			shared++
		}
	}
	min := len(a)
	if len(b) < min {
		min = len(b)
	}
	return float64(shared) / float64(min)
}

// producerFamilies clusters one item_type's census rows into producers.
//
// TWO AXES, and each is there because the other one alone gets a real case wrong:
//
//   - audit_source SPLITS FIRST. It is the only column that named the two producers
//     in the motivating bug, and rows carrying different labels are different
//     producers however similar their shapes.
//   - within a label, KEY-SETS CLUSTER. A raw distinct-key-set count cannot be used:
//     [MEASURED 2026-08-11] it reads 5 for hardcoded_section_colors but also 2 for
//     empty_section, 2 for literal_markdown, 2 for page_canonical_collision and 2
//     for truncated_component — four single-producer types — because producers add
//     and drop optional keys over their life (original_pipeline, out_of_remit,
//     intact_version_number). Clustering is what makes the count mean "producers"
//     rather than "spec revisions".
//
// A THIRD AXIS WAS MEASURED AND REJECTED (council round 1, editquality seat, corr
// fc082c4a): splitting on the VALUE of spec.check. It looks like a producer name
// and it is not — [MEASURED 2026-08-11] count(DISTINCT spec->>'check') reads 2 on
// page_canonical_collision, whose probe rows omit the key entirely, and that type
// has exactly one producer. Adding it would trade a real false negative for a
// certain false positive. The PRESENCE of `check` is already part of the key-set,
// so it discriminates where it legitimately can.
//
// THE RESIDUAL, stated where the code is rather than only in a doc: two producers
// that share an audit_source label AND overlap ≥50% of their keys merge into one
// family, and this check will not see them. No row-shaped test can, which is why
// the shape axis is a narrowing of the class and not a proof about it. What IS
// covered — and what the same-label test in remitcheck_test.go pins — is the case
// the live data cannot exercise: two disjoint shapes under one label, i.e. a second
// producer that never writes an audit_source at all.
//
// Rows whose spec has NO keys carry no shape evidence, so they are excluded from
// the families and returned separately — never silently dropped, because a count
// that quietly loses rows is how a census stops being one.
func producerFamilies(rows []censusRow) (families []family, shapeless int) {
	byLabel := map[string][]censusRow{}
	labels := []string{}
	for _, r := range rows {
		if strings.TrimSpace(r.Keyset) == "" {
			shapeless += r.Rows
			continue
		}
		if _, seen := byLabel[r.Label]; !seen {
			labels = append(labels, r.Label)
		}
		byLabel[r.Label] = append(byLabel[r.Label], r)
	}
	sort.Strings(labels)

	for _, label := range labels {
		shapes := byLabel[label]
		// Stable order in, stable clusters out — the report and the tests both
		// depend on this being deterministic.
		sort.Slice(shapes, func(i, j int) bool { return shapes[i].Keyset < shapes[j].Keyset })

		parent := make([]int, len(shapes))
		for i := range parent {
			parent[i] = i
		}
		var find func(int) int
		find = func(i int) int {
			for parent[i] != i {
				parent[i] = parent[parent[i]]
				i = parent[i]
			}
			return i
		}
		keys := make([][]string, len(shapes))
		for i, s := range shapes {
			keys[i] = strings.Split(s.Keyset, ",")
		}
		for i := range shapes {
			for j := i + 1; j < len(shapes); j++ {
				if overlap(keys[i], keys[j]) >= familyOverlapThreshold {
					if a, b := find(i), find(j); a != b {
						parent[a] = b
					}
				}
			}
		}

		clusters := map[int]*family{}
		order := []int{}
		for i, s := range shapes {
			root := find(i)
			f, seen := clusters[root]
			if !seen {
				f = &family{Label: label, First: s.First, Last: s.Last}
				clusters[root] = f
				order = append(order, root)
			}
			f.Shapes = append(f.Shapes, s.Keyset)
			f.Rows += s.Rows
			f.Samples = append(f.Samples, s.SampleID)
			if s.First < f.First {
				f.First = s.First
			}
			if s.Last > f.Last {
				f.Last = s.Last
			}
		}
		for _, root := range order {
			families = append(families, *clusters[root])
		}
	}
	return families, shapeless
}

// assess turns the census into one verdict per REGISTERED type. It iterates the
// registry, not the census, so a verified type with no rows at all is still
// reported as evaluated — "we looked and there was nothing to look at" is a
// different statement from "we did not look".
func assess(census []censusRow, registered []string, declaresRemit func(string) bool) []assessment {
	byType := map[string][]censusRow{}
	for _, r := range census {
		byType[r.ItemType] = append(byType[r.ItemType], r)
	}
	sorted := append([]string(nil), registered...)
	sort.Strings(sorted)

	out := make([]assessment, 0, len(sorted))
	for _, t := range sorted {
		rows := byType[t]
		families, shapeless := producerFamilies(rows)
		total := shapeless
		for _, f := range families {
			total += f.Rows
		}
		out = append(out, assessment{
			ItemType:      t,
			Rows:          total,
			Shapeless:     shapeless,
			Families:      families,
			DeclaresRemit: declaresRemit(t),
		})
	}
	return out
}

// ---------------------------------------------------------------------------
// Reporting and writing.
// ---------------------------------------------------------------------------

func itemKeyFor(itemType string) string { return "verifier-remit:" + itemType }

func summaryFor(a assessment) string {
	labels := make([]string, 0, len(a.Families))
	for _, f := range a.Families {
		l := f.Label
		if l == "" {
			l = "<no audit_source>"
		}
		labels = append(labels, fmt.Sprintf("%s (%d rows)", l, f.Rows))
	}
	return fmt.Sprintf(
		"verified item_type %q carries %d producer shapes [%s] but its verifier declares no remit — "+
			"one producer's items are being graded by a predicate written for another (bugs_open/213). "+
			"Fix: register a Grades on that verifier (VerifierPolicy.Grades, WII-013), or give the second producer its own item_type.",
		a.ItemType, len(a.Families), strings.Join(labels, "; "))
}

func specFor(a assessment) (string, error) {
	b, err := json.Marshal(map[string]interface{}{
		"check":           "verifier_remit_gap",
		"subject_type":    a.ItemType,
		"families":        a.Families,
		"rows":            a.Rows,
		"shapeless_rows":  a.Shapeless,
		"declares_remit":  a.DeclaresRemit,
		"detector":        createdBy,
		"builder_needed":  "verifier-remit:" + a.ItemType,
		"bug":             "bugs_open/213",
		"register_entry":  "WII-015",
		"remedy":          "RegisterVerifierWithPolicy(" + a.ItemType + ", v, VerifierPolicy{Grades: …}) — a POSITIVE shape match on what the predicate does grade, never a producer list",
		"acceptance_test": "discovery_checks.VerifierDeclaresRemit(\"" + a.ItemType + "\") == true, or the type resolves to one producer family",
	})
	return string(b), err
}

// fileFinding writes ONE undispatchable work item per subject type.
//
// UNDISPATCHABLE BY CONSTRUCTION, using remit.go's existing double lock: status
// 'deferred' and an EMPTY handler_agent. Neither half is new to this table, which
// was checked rather than argued from analogy (council round 1, guardian seat):
// [MEASURED 2026-08-11] `deferred` carries 316 live rows across 14 item_types, and
// 15 of those also carry an empty handler_agent — the capability_gap shape this
// copies. The dispatch loop selects `status IN ('triaged','approved')` and the
// promoter selects `status='detected'`, so neither can reach it; the one loose
// filter over this table anywhere in Go (`wi.status != 'complete'`, an admin
// listing) makes the row VISIBLE, which is the point of filing it.
//
// The remedy is a code change by a human or
// a session, so there is nothing to claim; a row that could be claimed would be
// claimed, attempted, and failed three times by a handler that cannot possibly fix
// it. 'deferred' also puts it on the roadmap view the fixloop digest already reads
// (diagnose_triage_action.go groups deferred rows by spec.builder_needed), so the
// finding surfaces somewhere a human already looks instead of somewhere new.
//
// ON CONFLICT DO NOTHING against idx_swi_dedup: re-running the check every day
// re-states the same finding under the same item_key, and must not churn the queue.
func fileFinding(db *sql.DB, a assessment) (bool, error) {
	spec, err := specFor(a)
	if err != nil {
		return false, err
	}
	res, err := db.Exec(`
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, created_by, item_key, max_attempts
		) VALUES (
			$1, $2, 'maintenance', $3, 'high', $4, $5::jsonb, 50, '', 'deferred', $2, $6, 1
		)
		ON CONFLICT DO NOTHING`,
		systemSiteID, createdBy, gapItemType, summaryFor(a), spec, itemKeyFor(a.ItemType))
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// closeAnswered retracts findings that no longer hold.
//
// THE SAFETY PROPERTY, and it is why this takes the assessments rather than the
// findings: a row is closed ONLY when THIS run positively evaluated that subject
// type and found it answered. A type that has vanished from the registry — a
// verifier deleted, or a census that failed — is left open and named in the report.
// Nothing anywhere derives "resolved" from an absence, because an errored or
// blinded run returns exactly that (WII-009 / RFC_010: retraction fires only on a
// positive observation).
func closeAnswered(db *sql.DB, assessments []assessment) (closedKeys, leftOpen []string, err error) {
	answered := map[string]assessment{}
	for _, a := range assessments {
		if !a.Finding() {
			answered[itemKeyFor(a.ItemType)] = a
		}
	}
	rows, err := db.Query(`
		SELECT item_key FROM site_work_items
		WHERE item_type = $1 AND created_by = $2
		  AND status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled')`,
		gapItemType, createdBy)
	if err != nil {
		return nil, nil, err
	}
	var open []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			rows.Close()
			return nil, nil, err
		}
		open = append(open, k)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var closed []string
	for _, k := range open {
		a, ok := answered[k]
		if !ok {
			// Either the finding still holds, or this run never evaluated that
			// subject type (its verifier was removed, so the registry no longer
			// names it). The second case is reported rather than closed: an
			// absence is not an observation.
			if !stillAFinding(assessments, k) {
				leftOpen = append(leftOpen, k)
			}
			continue
		}
		reason := fmt.Sprintf(
			"closed by verifier-remit-check: item_type %q now resolves to %d producer family/families with declares_remit=%v — the finding no longer holds",
			a.ItemType, len(a.Families), a.DeclaresRemit)
		res, err := db.Exec(`
			UPDATE site_work_items
			SET status = 'complete', completed_at = now(), updated_at = now(),
			    result = COALESCE(result,'{}'::jsonb) || jsonb_build_object('closed_by',$1,'reason',$2::text),
			    error = $2
			WHERE item_type = $3 AND created_by = $1 AND item_key = $4
			  AND status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled')`,
			createdBy, reason, gapItemType, k)
		if err != nil {
			return closed, leftOpen, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			closed = append(closed, k)
		}
	}
	return closed, leftOpen, nil
}

// stillAFinding reports whether an open item_key belongs to a subject type this
// run assessed AS a finding — i.e. the item is correctly still open. Anything
// else that is not in the answered set was not evaluated at all.
func stillAFinding(assessments []assessment, itemKey string) bool {
	for _, a := range assessments {
		if itemKeyFor(a.ItemType) == itemKey {
			return a.Finding()
		}
	}
	return false
}

// writeDocNote records the run on EVERY run, clean or not — the convention
// check_placeholder_fallbacks.py established and single-owner-carriers-check
// restated: a check that only speaks when it fails is indistinguishable from one
// that has stopped running. Best-effort: failing to RECORD a run must never become
// failing to REPORT it.
func writeDocNote(db *sql.DB, body string) {
	if db == nil {
		return
	}
	if _, err := db.Exec(`INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
	                      VALUES ('pipeline', 'work-item-integrity', $1, '["work-item-integrity"]'::jsonb, $2)`,
		body, createdBy); err != nil {
		fmt.Fprintf(os.Stderr, "(could not write doc_notes: %v)\n", err)
	}
}

type runOutcome struct {
	Filed     []string
	Closed    []string
	LeftOpen  []string // open items whose subject type this run did not evaluate
	DryRun    bool
	NoWrites  bool // no direct write route (PG_CLIENTS_HOST unset — the terminal route)
	IgnoreRem bool
}

func render(assessments []assessment, r runOutcome) string {
	filed, closed, dryRun, ignoreRemit := r.Filed, r.Closed, r.DryRun, r.IgnoreRem
	var b strings.Builder
	findings, suppressed := 0, 0
	for _, a := range assessments {
		if a.Finding() {
			findings++
		}
		if a.Suppressed() {
			suppressed++
		}
	}
	b.WriteString("# verifier-remit-check\n\n")
	b.WriteString(fmt.Sprintf("Verified item_types evaluated: **%d**. Findings: **%d**. "+
		"Multi-producer types answered by a declared remit: **%d**.\n\n",
		len(assessments), findings, suppressed))
	if ignoreRemit {
		b.WriteString("> ⚠ `--ignore-remit`: declared remits were IGNORED for this run (diagnostic mode, writes refused). " +
			"This is the disconfirmability check — it shows what the census says with the suppression turned off.\n\n")
	}
	if dryRun {
		b.WriteString("> `--dry-run`: nothing was written.\n\n")
	}
	if r.NoWrites && !dryRun {
		b.WriteString("> ⚠ READ-ONLY RUN: `PG_CLIENTS_HOST` is unset, so the census came through `kubectl exec` and " +
			"**nothing was written** — no work item, no close-out, no doc_note. This is a session at a terminal, " +
			"not the CronJob. Stated because a silent no-write reads exactly like a clean run.\n\n")
	}

	if findings == 0 {
		b.WriteString("No verified item_type has more than one producer shape without a declared remit.\n\n")
	}
	for _, a := range assessments {
		if !a.Finding() {
			continue
		}
		b.WriteString(fmt.Sprintf("## FINDING — `%s`\n\n%s\n\n", a.ItemType, summaryFor(a)))
		for _, f := range a.Families {
			label := f.Label
			if label == "" {
				label = "<no audit_source>"
			}
			b.WriteString(fmt.Sprintf("- **%s** — %d rows, %s → %s, shapes: `%s` (sample %s)\n",
				label, f.Rows, f.First, f.Last, strings.Join(f.Shapes, "` / `"), strings.Join(f.Samples, ", ")))
		}
		b.WriteString("\n")
	}

	// The positive control, always printed: what the census FOUND and the remit
	// answered. A zero with this section populated is a zero that looked.
	b.WriteString("## Multi-producer types answered by a declared remit (the positive control)\n\n")
	if suppressed == 0 {
		b.WriteString("_None today — no verified item_type with 2+ producer families declares a remit._\n\n")
	}
	for _, a := range assessments {
		if a.Suppressed() {
			b.WriteString(fmt.Sprintf("- `%s` — %d producer families, remit declared, so it is answered.\n",
				a.ItemType, len(a.Families)))
		}
	}
	b.WriteString("\n## Every verified item_type\n\n| item_type | rows | producer families | remit declared |\n|---|---|---|---|\n")
	for _, a := range assessments {
		note := ""
		if a.Shapeless > 0 {
			note = fmt.Sprintf(" (+%d rows with an empty spec, excluded from the shape census)", a.Shapeless)
		}
		b.WriteString(fmt.Sprintf("| `%s` | %d%s | %d | %v |\n", a.ItemType, a.Rows, note, len(a.Families), a.DeclaresRemit))
	}
	if len(filed) > 0 {
		b.WriteString(fmt.Sprintf("\nFiled: %s\n", strings.Join(filed, ", ")))
	}
	if len(closed) > 0 {
		b.WriteString(fmt.Sprintf("\nClosed (finding no longer holds): %s\n", strings.Join(closed, ", ")))
	}
	if len(r.LeftOpen) > 0 {
		b.WriteString(fmt.Sprintf("\n⚠ Left OPEN because this run did not evaluate their subject type (a verifier that "+
			"no longer exists, most likely — an absence is not an observation, so nothing was closed on it): %s\n",
			strings.Join(r.LeftOpen, ", ")))
	}
	return b.String()
}

func main() {
	dryRun := flag.Bool("dry-run", false, "print the report; write nothing (no work items, no closes, no doc_note)")
	emitJSON := flag.Bool("emit-json", false, "machine-readable assessments on stdout")
	ignoreRemit := flag.Bool("ignore-remit", false,
		"DIAGNOSTIC: ignore declared remits, so the census alone decides. Implies --dry-run — it exists to prove the check CAN fire, not to file on a type that has answered")
	flag.Parse()

	// The writing path is the DEFAULT and --dry-run is what suppresses it. The
	// inverse (a --report flag the CronJob must remember to pass) is how the
	// sibling detector ended up inert-by-omission — the exact objection that
	// produced its CronJob.
	if *ignoreRemit {
		*dryRun = true
	}

	registered := discovery_checks.RegisteredVerifierItemTypes()
	if len(registered) == 0 {
		fmt.Fprintln(os.Stderr, "REFUSING: the verifier registry is empty — that is a linking accident, not a clean bill of health")
		os.Exit(2)
	}

	db, err := dbConn()
	if err != nil {
		fmt.Fprintf(os.Stderr, "REFUSING: %v\n", err)
		os.Exit(2)
	}
	if db != nil {
		defer db.Close()
	}
	census, err := fetchCensus(db, censusSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "REFUSING: %v\n", err)
		os.Exit(2)
	}
	if len(census) == 0 {
		fmt.Fprintln(os.Stderr, "REFUSING: the census returned no rows at all — site_work_items is never empty, so this is a broken read, not a clean fleet")
		os.Exit(2)
	}

	declares := discovery_checks.VerifierDeclaresRemit
	if *ignoreRemit {
		declares = func(string) bool { return false }
	}
	assessments := assess(census, registered, declares)

	outcome := runOutcome{DryRun: *dryRun, NoWrites: db == nil, IgnoreRem: *ignoreRemit}
	if !*dryRun && db != nil {
		for _, a := range assessments {
			if !a.Finding() {
				continue
			}
			created, err := fileFinding(db, a)
			if err != nil {
				fmt.Fprintf(os.Stderr, "REFUSING: could not file %s: %v\n", a.ItemType, err)
				os.Exit(2)
			}
			if created {
				outcome.Filed = append(outcome.Filed, a.ItemType)
			}
		}
		outcome.Closed, outcome.LeftOpen, err = closeAnswered(db, assessments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "REFUSING: close-out failed: %v\n", err)
			os.Exit(2)
		}
	}

	report := render(assessments, outcome)
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
