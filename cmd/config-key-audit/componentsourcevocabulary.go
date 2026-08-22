// FILE: cmd/config-key-audit/componentsourcevocabulary.go
//
// bugs_open/309 §5 candidate 2 — the AT-REST half of the source-vocabulary rule.
//
// CLC-018 shipped the BIRTH gate: a generated component whose input_schema
// declares a source the platform cannot resolve (a `site_specs.<aspect>` no site
// has ever carried, an unregistered `query.*` name, a prefix outside the
// vocabulary) is refused at `store_generated_component`. That closed one door.
// This is the room: nothing has ever asked the same question of the components
// ALREADY in the database, where — measured 2026-08-22 — 69 fields across 17
// active components declare a source that resolves nowhere, six of them live on
// 46 page instances.
//
// WHY THE BIRTH GATE CANNOT EVER COVER THIS. A component is routinely inserted
// or altered by a hand-written migration or by hand SQL, which never passes
// through the Go action at all (a standing LANDMINE entry, and precisely how the
// motivating component `blog-listing_pre_037` got there). Only a clock against
// live `content_components` sees that population. Same settled reasoning as
// single-owner-carriers-check and the capped-schedule-ordering sibling.
//
// WHY IT CALLS THE GUARD'S OWN FUNCTION rather than re-implementing the rule.
// The concept register asked for exactly this in as many words — CLC-018:
// "sourceVocabularyIssues (pure, reusable — a future daily audit of EXISTING
// config should call IT)" and "build it ON sourceVocabularyIssues, not on a
// second predicate, or they drift". A mirror can only ever DETECT drift; calling
// the function makes drift unrepresentable. That is why this mode ships in a
// pre-built Go image and not as a Python script beside the older checks.
//
// EXIT CODES: 0 = every live finding is grandfathered by the frozen baseline;
// 1 = at least one red (see componentSourceReds); 2 = the check did not run,
// which must never read as a pass.
//
// TWO VACUITY REFUSALS, AND THEY FAIL IN OPPOSITE DIRECTIONS — both are exit 2:
//   - zero active components decoded: the classic silence, where "0 findings"
//     reads clean;
//   - zero site_specs aspects: the opposite, a FLOOD, because an empty aspect
//     set marks every `site_specs.*` source in the estate as phantom. A read
//     that returns no aspects never means the estate has none.
//
// ⚠ THE FAIL-OPEN ASYMMETRY WITH THE BIRTH GATE IS DELIBERATE. The gate skips
// the aspect half when the aspect set is unreadable, because it must not block
// all component generation. This audit refuses instead: an audit has nothing to
// block, and a daily report that silently skipped a third of its rule is the
// blind-pass landmine one rung higher. Do not "make them consistent".
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/actions"
)

// componentSourceBaselineClosedOn is the date the grandfather baseline was
// frozen. Every entry must carry it. An entry with any other date means someone
// APPENDED to a file that exists only to shrink — see the closure test.
const componentSourceBaselineClosedOn = "2026-08-22"

const defaultComponentSourceBaselinePath = "docs/agent_docs/docs024_key_docs_latest/" +
	"bugfix_309_unclickable_index_cards/component_source_baseline.json"

// componentRow is one active component as the audit reads it: enough to run the
// rule and to say how much live page surface a finding is sitting on.
type componentRow struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	InputSchema   string `json:"input_schema"`
	LiveInstances int    `json:"live_instances"`
}

// componentSourceFinding is one live offence, plus the two facts that decide
// whether it is red: whether the frozen baseline holds its exact tuple, and
// whether a component that was DORMANT when baselined has since been deployed.
type componentSourceFinding struct {
	ComponentID   string `json:"component_id"`
	Component     string `json:"component"`
	Field         string `json:"field"`
	Source        string `json:"source"`
	Class         string `json:"class"`
	LiveInstances int    `json:"live_instances"`
	Message       string `json:"message"`

	Grandfathered bool `json:"grandfathered"`
	// WokeUp is true for a finding whose baseline entry recorded ZERO live
	// instances and which now has some. Grandfathering the dormant components
	// is conditional on their STAYING dormant: deploying one is a new page
	// acquiring a known silent field-drop, which is new damage.
	WokeUp bool `json:"woke_up"`
}

type componentSourceBaselineEntry struct {
	ComponentID             string `json:"component_id"`
	Component               string `json:"component"`
	Field                   string `json:"field"`
	Source                  string `json:"source"`
	Class                   string `json:"class"`
	LiveInstancesAtBaseline int    `json:"live_instances_at_baseline"`
	Baselined               string `json:"baselined"`
	Route                   string `json:"route"`
}

type componentSourceBaselineFile struct {
	Doc     string                         `json:"_doc"`
	Entries []componentSourceBaselineEntry `json:"entries"`
}

// componentSourceStdin is the offline payload (--json), used by the controls and
// the unit tests: it supplies both halves of the live state so every branch is
// reachable without a cluster.
type componentSourceStdin struct {
	Aspects    []string       `json:"aspects"`
	Components []componentRow `json:"components"`
}

// baselineKey is the NARROWEST key that identifies a finding, and the narrowness
// is the whole safety property. Keyed on the component alone, or on the class, a
// baseline entry would swallow future offences on the same component — which is
// the allow-list-silences-your-own-detector trap this estate has already been
// bitten by (scripts/pattern-check.py's COMPONENT_WRITE_ALLOWED note). Keyed on
// the exact tuple, an entry can only ever excuse the one finding it was written
// for; a changed source string on the same field is a NEW finding.
func baselineKey(componentID, field, source, class string) string {
	return componentID + "\x00" + field + "\x00" + source + "\x00" + class
}

func loadComponentSourceBaseline(path string) (map[string]componentSourceBaselineEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file componentSourceBaselineFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("baseline file is not the expected object: %w", err)
	}
	out := make(map[string]componentSourceBaselineEntry, len(file.Entries))
	for i, e := range file.Entries {
		if e.ComponentID == "" || e.Field == "" || e.Class == "" {
			return nil, fmt.Errorf("baseline entry %d is missing component_id/field/class", i)
		}
		if e.Baselined != componentSourceBaselineClosedOn {
			// Refuse rather than warn. A warning on stderr in a CronJob is a
			// warning nobody reads, and the whole value of the baseline is that
			// it cannot grow.
			return nil, fmt.Errorf(
				"baseline entry %d (%s.%s) is dated %q, not %q — the baseline is CLOSED: "+
					"a new finding is fixed or routed, never baselined. If a genuine "+
					"re-baseline has been licensed, that is an owner decision and this "+
					"constant moves with it",
				i, e.Component, e.Field, e.Baselined, componentSourceBaselineClosedOn)
		}
		out[baselineKey(e.ComponentID, e.Field, e.Source, e.Class)] = e
	}
	return out, nil
}

// componentSourceFindings runs the BIRTH GATE'S OWN RULE over every active
// component and marks each offence against the baseline.
func componentSourceFindings(components []componentRow, aspects map[string]bool,
	baseline map[string]componentSourceBaselineEntry) []componentSourceFinding {

	var findings []componentSourceFinding
	for _, c := range components {
		for _, iss := range actions.SourceVocabularyFindings(c.InputSchema, aspects) {
			key := baselineKey(c.ID, iss.Field, iss.Source, iss.Class)
			entry, known := baseline[key]
			findings = append(findings, componentSourceFinding{
				ComponentID:   c.ID,
				Component:     c.Name,
				Field:         iss.Field,
				Source:        iss.Source,
				Class:         iss.Class,
				LiveInstances: c.LiveInstances,
				Message:       iss.Message,
				Grandfathered: known,
				WokeUp:        known && entry.LiveInstancesAtBaseline == 0 && c.LiveInstances > 0,
			})
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].LiveInstances != findings[j].LiveInstances {
			return findings[i].LiveInstances > findings[j].LiveInstances // repair-owed first
		}
		if findings[i].Component != findings[j].Component {
			return findings[i].Component < findings[j].Component
		}
		return findings[i].Field < findings[j].Field
	})
	return findings
}

// staleBaselineEntries are entries matching nothing live — repaired, or the
// component deactivated. THIS IS THE RATCHET'S PAWL: it is what makes the file
// shrink mechanically as repairs land, instead of accumulating dead entries that
// could later mask a re-offence on the same tuple.
func staleBaselineEntries(baseline map[string]componentSourceBaselineEntry,
	findings []componentSourceFinding) []componentSourceBaselineEntry {

	seen := make(map[string]bool, len(findings))
	for _, f := range findings {
		seen[baselineKey(f.ComponentID, f.Field, f.Source, f.Class)] = true
	}
	var stale []componentSourceBaselineEntry
	for key, e := range baseline {
		if !seen[key] {
			stale = append(stale, e)
		}
	}
	sort.Slice(stale, func(i, j int) bool {
		if stale[i].Component != stale[j].Component {
			return stale[i].Component < stale[j].Component
		}
		return stale[i].Field < stale[j].Field
	})
	return stale
}

// componentSourceReds counts the conditions that make the job fail. A red is
// always EITHER a real new finding OR a one-line baseline trim — never a
// standing backlog, which is the thing that trains people to ignore a check.
func componentSourceReds(findings []componentSourceFinding,
	stale []componentSourceBaselineEntry) (ungrandfathered, wokeUp, staleCount int) {

	for _, f := range findings {
		if !f.Grandfathered {
			ungrandfathered++
		}
		if f.WokeUp {
			wokeUp++
		}
	}
	return ungrandfathered, wokeUp, len(stale)
}

// componentSourceRunSummary is the doc_notes body and the --report stdout.
//
// It reports the WHOLE grandfathered population every run, not just the reds.
// The baseline governs the EXIT CODE and never the REPORT: a detector that still
// names all 69 offences every morning cannot quietly go blind, which is the
// failure mode an allow-list normally introduces.
func componentSourceRunSummary(components int, aspects int, baselinePath string,
	findings []componentSourceFinding, stale []componentSourceBaselineEntry) string {

	ungrandfathered, wokeUp, staleCount := componentSourceReds(findings, stale)

	var b strings.Builder
	fmt.Fprintf(&b, "component source-vocabulary audit (bugs_open/309, CLC-018's at-rest half)\n")
	fmt.Fprintf(&b, "scope: %d active components with an input_schema.fields object, "+
		"against %d live site_specs aspects and the registered query vocabulary.\n",
		components, aspects)
	fmt.Fprintf(&b, "baseline: %s\n\n", baselinePath)

	byClass := map[string]int{}
	liveSurface := map[string]int{}
	for _, f := range findings {
		byClass[f.Class]++
		if f.LiveInstances > 0 {
			liveSurface[f.Component] = f.LiveInstances
		}
	}
	instances := 0
	for _, n := range liveSurface {
		instances += n
	}

	fmt.Fprintf(&b, "%d fields declare a source that resolves nowhere "+
		"(phantom_aspect %d, unregistered_query %d, prefix_outside_vocabulary %d), "+
		"across %d components; %d of those components are live on %d page instances.\n",
		len(findings),
		byClass[actions.SourceIssuePhantomAspect],
		byClass[actions.SourceIssueUnregisteredQuery],
		byClass[actions.SourceIssuePrefixOutsideVocabulary],
		countComponents(findings), len(liveSurface), instances)

	if len(findings) == 0 {
		b.WriteString("\nNo field in the active library declares an unresolvable source. " +
			"The baseline is empty or fully repaired.\n")
	}

	if ungrandfathered > 0 {
		b.WriteString("\nRED — NOT in the frozen baseline (a component born or altered by a " +
			"route the birth gate cannot see, or a field that has changed source):\n")
		for _, f := range findings {
			if !f.Grandfathered {
				fmt.Fprintf(&b, "  * %s.%s [%s] source=%q, %d live instances\n    %s\n",
					f.Component, f.Field, f.Class, f.Source, f.LiveInstances, f.Message)
			}
		}
	}
	if wokeUp > 0 {
		b.WriteString("\nRED — baselined while DORMANT and now DEPLOYED (grandfathering was " +
			"conditional on staying dormant; a new page has acquired a known silent " +
			"field-drop):\n")
		for _, f := range findings {
			if f.WokeUp {
				fmt.Fprintf(&b, "  * %s.%s [%s] now on %d live instances\n",
					f.Component, f.Field, f.Class, f.LiveInstances)
			}
		}
	}
	if staleCount > 0 {
		b.WriteString("\nRED — STALE baseline entries, matching nothing live. This is the " +
			"ratchet working: the finding is gone, so delete these lines from the " +
			"baseline file (the file may only ever shrink):\n")
		for _, e := range stale {
			fmt.Fprintf(&b, "  * %s.%s [%s] source=%q\n", e.Component, e.Field, e.Class, e.Source)
		}
	}

	if ungrandfathered == 0 && wokeUp == 0 && staleCount == 0 {
		fmt.Fprintf(&b, "\nAll %d are grandfathered by the frozen baseline and unchanged. "+
			"CLEAN: nothing new, nothing woken, nothing stale.\n", len(findings))
	}

	b.WriteString("\nGrandfathered population — repair is owed, not excused. Live first:\n")
	shown := 0
	for _, f := range findings {
		if !f.Grandfathered {
			continue
		}
		if f.LiveInstances == 0 && shown >= 25 {
			continue // dormant tail: the count above is the record
		}
		fmt.Fprintf(&b, "  - %s.%s [%s] source=%q, %d live instances\n",
			f.Component, f.Field, f.Class, f.Source, f.LiveInstances)
		shown++
	}
	fmt.Fprintf(&b, "\n(%d grandfathered entries in total; the dormant tail is truncated in "+
		"this body, never in the exit code.)\n", len(findings)-ungrandfathered)

	return b.String()
}

func countComponents(findings []componentSourceFinding) int {
	seen := map[string]bool{}
	for _, f := range findings {
		seen[f.ComponentID] = true
	}
	return len(seen)
}

const activeComponentsQuery = `
SELECT cc.id::text,
       cc.name,
       COALESCE(cc.input_schema::text, ''),
       (SELECT count(*) FROM page_components pc
          JOIN pages p ON p.id = pc.page_id
         WHERE pc.component_id = cc.id
           AND p.status IN ('active','deployed'))
FROM content_components cc
WHERE cc.is_active
  AND jsonb_typeof(cc.input_schema->'fields') = 'object';`

// loadActiveComponents reads the live library.
//
// SCOPE, and it matters: `content_components` ONLY. information_schema lists 56
// tables carrying an `input_schema` column, and every one but this and
// `component_versions` is a dated backup snapshot (`bak_*`, `*_backup_*`);
// `component_versions` is history. A census that globs the column name reports
// the same dead field dozens of times and cannot be reconciled with anything.
func loadActiveComponents(db *sql.DB) ([]componentRow, error) {
	rows, err := db.Query(activeComponentsQuery)
	if err != nil {
		return nil, fmt.Errorf("active components query failed: %w", err)
	}
	defer rows.Close()

	var out []componentRow
	for rows.Next() {
		var c componentRow
		if err := rows.Scan(&c.ID, &c.Name, &c.InputSchema, &c.LiveInstances); err != nil {
			return nil, fmt.Errorf("scanning component: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func loadSpecAspects(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT DISTINCT aspect FROM site_specs`)
	if err != nil {
		return nil, fmt.Errorf("site_specs aspect query failed: %w", err)
	}
	defer rows.Close()

	aspects := map[string]bool{}
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, fmt.Errorf("scanning aspect: %w", err)
		}
		aspects[a] = true
	}
	return aspects, rows.Err()
}

// emitComponentSourceVocabulary: [--baseline <file>] [--report] [--json].
func emitComponentSourceVocabulary(args []string) {
	baselinePath := defaultComponentSourceBaselinePath
	report := false
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--baseline":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr,
					"config-key-audit --component-source-vocabulary: --baseline needs a file path")
				os.Exit(2)
			}
			baselinePath = args[i+1]
			i++
		case args[i] == "--report":
			report = true
		case args[i] == "--json":
			// read state from stdin instead of the cluster; default already
		default:
			fmt.Fprintf(os.Stderr,
				"config-key-audit --component-source-vocabulary: unrecognised argument %q "+
					"(want: [--baseline <file>] [--report] [--json])\n", args[i])
			os.Exit(2)
		}
	}

	baseline, err := loadComponentSourceBaseline(baselinePath)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --component-source-vocabulary: baseline %q: %v — refusing to "+
				"run without it. Without the baseline every one of the grandfathered "+
				"findings reads as new, which is a flood, not a report.\n", baselinePath, err)
		os.Exit(2)
	}

	var (
		components []componentRow
		aspects    map[string]bool
	)
	if report {
		db, err := dbConn()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --component-source-vocabulary: %v\n", err)
			os.Exit(2)
		}
		if db == nil {
			fmt.Fprintln(os.Stderr,
				"config-key-audit --component-source-vocabulary --report: PG_CLIENTS_HOST is "+
					"not set, so there is no library to read. In the CronJob this comes from "+
					"the pod env; by hand, pipe a --json payload instead.")
			os.Exit(2)
		}
		defer db.Close()
		if components, err = loadActiveComponents(db); err == nil {
			aspects, err = loadSpecAspects(db)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --component-source-vocabulary: %v\n", err)
			os.Exit(2)
		}
	} else {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"config-key-audit --component-source-vocabulary: reading stdin: %v\n", err)
			os.Exit(2)
		}
		var in componentSourceStdin
		if err := json.Unmarshal(raw, &in); err != nil {
			fmt.Fprintln(os.Stderr,
				"config-key-audit --component-source-vocabulary: stdin must be one object "+
					`{"aspects": ["identity", ...], "components": [{"id","name","input_schema","live_instances"}]} `+
					"— both halves are needed, because the aspect vocabulary is half the rule.")
			os.Exit(2)
		}
		components = in.Components
		aspects = map[string]bool{}
		for _, a := range in.Aspects {
			aspects[a] = true
		}
	}

	// The two vacuity refusals. They fail in OPPOSITE directions and both must
	// exit 2 — see this file's header.
	if len(components) == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit --component-source-vocabulary: 0 active components with an "+
				"input_schema.fields object — refusing to print a clean report over an empty "+
				"library. The estate has hundreds; zero means the read broke.")
		os.Exit(2)
	}
	if len(aspects) == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit --component-source-vocabulary: 0 site_specs aspects — refusing "+
				"to run. This is the OPPOSITE failure to the one above: an empty aspect set "+
				"marks every site_specs.* source in the estate as phantom, so the run would "+
				"be a flood rather than a silence. A read returning no aspects never means "+
				"the estate carries none.")
		os.Exit(2)
	}

	findings := componentSourceFindings(components, aspects, baseline)
	stale := staleBaselineEntries(baseline, findings)
	ungrandfathered, wokeUp, staleCount := componentSourceReds(findings, stale)
	summary := componentSourceRunSummary(len(components), len(aspects), baselinePath, findings, stale)

	if report {
		fmt.Print(summary)
		// ONE row per run, clean or not — a check that only speaks when it fails
		// is indistinguishable from one that has stopped running.
		//
		// Deliberately NOT agent_error_log: bugs_open/358 measures that channel
		// as write-only with a 30-day retention delete, and STRUCTURAL_KEY_CARRY_MISS
		// — the runtime detector of this very silence — is one of its unread codes.
		writeDocNote("component-library", summary,
			"component-integrity", "component-source-vocabulary-check")
	} else {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(findings); err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --component-source-vocabulary: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprint(os.Stderr, summary)
	}

	if ungrandfathered > 0 || wokeUp > 0 || staleCount > 0 {
		os.Exit(1)
	}
}
