//go:build ignore

// Renders components/evidence-chart/template.html with sample_data.json using
// html/template exactly as the platform render path does, and asserts the
// behaviours that matter — including the failing branches (page filter
// excludes, dangling fact reference draws nothing, round millions do not
// become 1e+06).
//
// Usage: go run render_evidence_chart.go <template.html> <sample_data.json> [current_page]
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: render_evidence_chart <template> <data> [current_page]")
		os.Exit(2)
	}
	tb, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("read template:", err)
		os.Exit(1)
	}
	db, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Println("read data:", err)
		os.Exit(1)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(db, &data); err != nil {
		fmt.Println("unmarshal data:", err)
		os.Exit(1)
	}
	if len(os.Args) > 3 {
		data["current_page"] = os.Args[3]
	}

	t, err := template.New("evidence-chart").Parse(string(tb))
	if err != nil {
		fmt.Println("PARSE FAILED:", err)
		os.Exit(1)
	}
	var b strings.Builder
	if err := t.Execute(&b, data); err != nil {
		fmt.Println("EXECUTE FAILED:", err)
		os.Exit(1)
	}
	out := b.String()

	page, _ := data["current_page"].(string)
	fmt.Printf("==== rendered for current_page=%q (%d bytes) ====\n", page, len(out))
	fmt.Println(out)

	checks := []struct {
		name string
		ok   bool
	}{
		{"no ZgotmplZ (CSS filter did not reject a value)", !strings.Contains(out, "ZgotmplZ")},
		{"no 1e+06 exponent notation in CSS", !strings.Contains(out, "1e+06") && !strings.Contains(out, "1e&#43;06")},
		{"round million rendered as digits", strings.Contains(out, "--v:1000000.0000")},
		{"dangling fact_id drew no row", !strings.Contains(out, "Dangling reference")},
		{"unreferenced fact did not leak", !strings.Contains(out, "F1-live-sites")},
	}
	// Scoped to the pages that actually carry the relojistas chart — asserting
	// them everywhere would fail on `capabilities` for the right reason
	// (correctly excluded) and read as a template defect.
	if page == "index" || page == "" {
		checks = append(checks,
			struct {
				name string
				ok   bool
			}{"zero-value bar rendered", strings.Contains(out, "--v:0.0000")},
			struct {
				name string
				ok   bool
			}{"unit suffix applied", strings.Contains(out, "97%")},
			struct {
				name string
				ok   bool
			}{"verified date present", strings.Contains(out, "verified 2026-07-16")},
		)
	}
	if page == "index" {
		checks = append(checks,
			struct {
				name string
				ok   bool
			}{"index chart shown", strings.Contains(out, `data-chart="relojistas-feed-restoration"`)},
			struct {
				name string
				ok   bool
			}{"capabilities-only chart excluded", !strings.Contains(out, `data-chart="capabilities-only-chart"`)},
			struct {
				name string
				ok   bool
			}{"unpaged chart shown", strings.Contains(out, `data-chart="unpaged-chart"`)},
		)
	}
	if page == "capabilities" {
		checks = append(checks,
			struct {
				name string
				ok   bool
			}{"capabilities chart shown", strings.Contains(out, `data-chart="capabilities-only-chart"`)},
			struct {
				name string
				ok   bool
			}{"index-only chart excluded", !strings.Contains(out, `data-chart="relojistas-feed-restoration"`)},
		)
	}
	if page == "" {
		checks = append(checks, struct {
			name string
			ok   bool
		}{"empty current_page degrades to showing all charts", strings.Contains(out, `data-chart="capabilities-only-chart"`) && strings.Contains(out, `data-chart="relojistas-feed-restoration"`)})
	}

	fmt.Println("---- checks ----")
	failed := 0
	for _, c := range checks {
		status := "PASS"
		if !c.ok {
			status = "FAIL"
			failed++
		}
		fmt.Printf("%s  %s\n", status, c.name)
	}
	if failed > 0 {
		fmt.Printf("\n%d CHECK(S) FAILED\n", failed)
		os.Exit(1)
	}
	fmt.Println("\nall checks passed")
}
