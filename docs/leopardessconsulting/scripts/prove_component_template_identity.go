// Byte-identity proof for the opt-in carousel branch added to info-card-grid.
//
// Renders every live instance's content_data through BOTH the old and the new
// html_template with Go's text/template (the same engine
// RenderTemplateReportingMissing uses via executeGoTemplate) and asserts the
// output is byte-identical when `carousel` is absent.
//
// It also runs the two controls that make the pass mean something:
//   - MUTANT: the new template rendered WITH carousel:true must DIFFER, or the
//     comparison is measuring nothing.
//   - ERROR:  neither template may return a template execution error, because
//     the production renderer silently falls back to a regex renderer on error
//     and would mangle the markup rather than fail loudly.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"
)

type instance struct {
	Domain string                 `json:"domain"`
	Page   string                 `json:"page"`
	PCID   string                 `json:"pc_id"`
	CD     map[string]interface{} `json:"cd"`
}

func render(tplStr string, data map[string]interface{}) (string, error) {
	t, err := template.New("c").Parse(tplStr)
	if err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return "", fmt.Errorf("execute: %w", err)
	}
	return b.String(), nil
}

func main() {
	oldTpl := mustRead("../icg_template.html")
	newTpl := mustRead("../icg_template_new.html")

	raw := mustRead("../icg_instances.json")
	var instances []instance
	if err := json.Unmarshal([]byte(raw), &instances); err != nil {
		fatal("unmarshal instances: %v", err)
	}

	identical, differing, errs, mutantsDiffer, mutantsSame := 0, 0, 0, 0, 0

	for _, in := range instances {
		data := in.CD
		if data == nil {
			data = map[string]interface{}{}
		}

		got, err := render(oldTpl, data)
		if err != nil {
			errs++
			fmt.Printf("ERROR old  %s %s: %v\n", in.Domain, in.Page, err)
			continue
		}
		want, err := render(newTpl, data)
		if err != nil {
			errs++
			fmt.Printf("ERROR new  %s %s: %v\n", in.Domain, in.Page, err)
			continue
		}

		if got == want {
			identical++
		} else {
			differing++
			fmt.Printf("DIFF %s %s (old %d bytes, new %d bytes)\n",
				in.Domain, in.Page, len(got), len(want))
			showFirstDiff(got, want)
		}

		// MUTANT CONTROL: same data + the flag must produce different markup.
		mut := map[string]interface{}{}
		for k, v := range data {
			mut[k] = v
		}
		mut["carousel"] = true
		mutOut, err := render(newTpl, mut)
		if err != nil {
			errs++
			fmt.Printf("ERROR mutant %s %s: %v\n", in.Domain, in.Page, err)
			continue
		}
		if mutOut == want {
			mutantsSame++
			fmt.Printf("MUTANT-INERT %s %s — carousel:true changed NOTHING\n", in.Domain, in.Page)
		} else {
			mutantsDiffer++
		}
	}

	fmt.Printf("\n=== byte-identity (carousel ABSENT) ===\n")
	fmt.Printf("identical : %d\n", identical)
	fmt.Printf("differing : %d\n", differing)
	fmt.Printf("errors    : %d\n", errs)
	fmt.Printf("\n=== mutant control (carousel:true) ===\n")
	fmt.Printf("differs from default : %d  (must equal the instance count)\n", mutantsDiffer)
	fmt.Printf("inert                : %d  (must be 0)\n", mutantsSame)

	if differing != 0 || errs != 0 || mutantsDiffer != len(instances) {
		fmt.Printf("\nFAIL\n")
		os.Exit(1)
	}
	fmt.Printf("\nPASS — %d instances byte-identical, and the flag provably changes the output on all %d\n",
		identical, mutantsDiffer)
}

func showFirstDiff(a, b string) {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			lo := i - 120
			if lo < 0 {
				lo = 0
			}
			fmt.Printf("  first divergence at byte %d\n", i)
			fmt.Printf("  old: %q\n", a[lo:min(i+120, len(a))])
			fmt.Printf("  new: %q\n", b[lo:min(i+120, len(b))])
			return
		}
	}
	fmt.Printf("  common prefix identical; lengths differ (old %d, new %d)\n", len(a), len(b))
	if len(a) > n {
		fmt.Printf("  old tail: %q\n", a[n:min(n+200, len(a))])
	} else {
		fmt.Printf("  new tail: %q\n", b[n:min(n+200, len(b))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func mustRead(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		fatal("read %s: %v", p, err)
	}
	return string(b)
}

func fatal(f string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, f+"\n", a...)
	os.Exit(1)
}
