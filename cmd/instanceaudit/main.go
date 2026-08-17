// instanceaudit — classify every live component template by what actually
// happens if it is placed on a page TWICE.
//
// WHY THIS EXISTS (bugs_open/283, RFC_034 §3a). Sizing the conversion by regex
// gave the wrong answer in the direction that made the job look easy: three
// patterns (`window.onload`, inline `on*=`, a `function` keyword near the top of
// a script) flagged 24 of 91 templates, and the real classifier says 88. The 64
// it missed declare globals in spellings it did not search for.
//
// So this runs the PRODUCTION detector — actions.DetectInstanceCollisions, the
// same function that gates a render — over the live templates, rather than a
// second implementation of its judgement. That matters twice over: the numbers
// are reproducible, and they are the numbers the acceptance gate will use, so a
// component this tool calls unscoped is one the gate will reject.
//
// USAGE. Export the live templates, then classify them:
//
//	kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
//	  psql -U clients_user -d clients_db -t -A -c "
//	    SELECT json_agg(json_build_object('id', x.id, 'function', x.function, 'tpl', x.html_template))
//	    FROM (SELECT DISTINCT c.id::text AS id, c.function, c.html_template
//	          FROM content_components c JOIN page_components pc ON pc.component_id=c.id
//	               JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
//	          WHERE c.is_active AND c.html_template ~ 'getElementById') x;" > /tmp/templates.json
//	go run ./cmd/instanceaudit /tmp/templates.json [--list]
//
// The export is deliberately the caller's job: this reads a file and touches no
// database, so it can be run against a snapshot, a single candidate template, or
// a proposed conversion's OUTPUT — which is how you check a converter's work
// rather than only its input.
//
// READING THE OUTPUT. "unscoped" counts inline script bodies the detector does
// not recognise as wrapped; it errs toward reporting, so treat it as a floor on
// "must satisfy the gate" rather than proof that a given script is unsafe.
// "dupIfDoubled" is the honest headline: the number of element ids that collide
// when the component appears twice on one page. Baseline 2026-08-17, 91 live
// templates: 3 already scoped, 88 declaring into global scope, 8 assigning
// window.onload, and 91 of 91 producing duplicate ids (1,345 in total) — which
// independently corroborates the SQL census of 1,346 literal id attributes.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/actions"
)

type row struct {
	ID       string `json:"id"`
	Function string `json:"function"`
	Tpl      string `json:"tpl"`
}

type detail struct {
	fn           string
	unscopedN    int
	onloadN      int
	dupIfDoubled int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: instanceaudit <templates.json> [--list]")
		os.Exit(2)
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", os.Args[1], err)
		os.Exit(2)
	}
	var rows []row
	if err := json.Unmarshal(b, &rows); err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse %s: %v\n", os.Args[1], err)
		os.Exit(2)
	}
	// A clean report over an empty export would be indistinguishable from a
	// clean report over a healthy fleet. Refuse rather than reassure.
	if len(rows) == 0 {
		fmt.Fprintln(os.Stderr, "REFUSING: 0 templates in the export — the query "+
			"returned nothing, and a clean report over an empty set is not evidence.")
		os.Exit(2)
	}

	var scoped, unscoped, onload int
	buckets := map[string][]string{}
	details := make([]detail, 0, len(rows))

	for _, r := range rows {
		one := actions.DetectInstanceCollisions(r.Tpl)         // does it declare globally?
		two := actions.DetectInstanceCollisions(r.Tpl + r.Tpl) // what collides if doubled?

		if one.UnscopedInlineScripts == 0 {
			scoped++
		} else {
			unscoped++
		}
		if one.WindowOnloadAssignments > 0 {
			onload++
		}

		var k string
		switch {
		case one.UnscopedInlineScripts == 0 && one.WindowOnloadAssignments == 0:
			k = "A. script already SAFE (ids only)"
		case one.WindowOnloadAssignments > 0:
			k = "C. window.onload (single slot, last wins)"
		default:
			k = "B. declares into GLOBAL scope"
		}
		buckets[k] = append(buckets[k], r.Function)
		details = append(details, detail{r.Function, one.UnscopedInlineScripts,
			one.WindowOnloadAssignments, len(two.DuplicateElementIDs)})
	}

	fmt.Printf("templates analysed: %d\n\n", len(rows))
	fmt.Printf("  script bodies already scoped (0 unscoped):   %d\n", scoped)
	fmt.Printf("  declare into global scope (>=1 unscoped):    %d\n", unscoped)
	fmt.Printf("  assign window.onload:                        %d\n\n", onload)

	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%-45s %3d\n", k, len(buckets[k]))
	}

	sort.Slice(details, func(i, j int) bool { return details[i].dupIfDoubled > details[j].dupIfDoubled })
	var totalDup, zeroDup int
	for _, d := range details {
		totalDup += d.dupIfDoubled
		if d.dupIfDoubled == 0 {
			zeroDup++
		}
	}
	fmt.Printf("\nIf each component appeared TWICE on one page:\n")
	fmt.Printf("  total duplicate ids across all %d:           %d\n", len(rows), totalDup)
	fmt.Printf("  components with ZERO id collisions:          %d\n", zeroDup)
	fmt.Printf("  worst 8: ")
	for i := 0; i < 8 && i < len(details); i++ {
		fmt.Printf("%s(%d) ", details[i].fn, details[i].dupIfDoubled)
	}
	fmt.Println()

	if len(os.Args) > 2 && os.Args[2] == "--list" {
		fmt.Println("\n--- per component ---")
		for _, d := range details {
			fmt.Printf("%-45s unscoped=%d onload=%d dupIfDoubled=%d\n",
				strings.TrimSpace(d.fn), d.unscopedN, d.onloadN, d.dupIfDoubled)
		}
	}
}
