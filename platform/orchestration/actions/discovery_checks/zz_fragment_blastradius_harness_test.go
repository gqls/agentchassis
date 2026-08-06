//go:build fleetharness

// Blast-radius harness for the dead_fragment_link arm. NOT a unit test — it is
// the SHIPPING functions run over a dump of live fleet HTML, so the number it
// prints is what the check would file on its first real run.
//
// Run: go test -tags fleetharness ./platform/orchestration/actions/discovery_checks/ \
//        -run TestFleetFragmentBlastRadius -v -fleet=<path to fleet.json>
//
// The dump query is in RUNBOOK_fragment_blindspot.md. Reimplementing the
// predicate in SQL was rejected deliberately (LANDMINES: the SQL has to
// hand-copy NormalizePagePath and ClassifyLinkScope, and the two answers
// differed the first time they were compared).

package discovery_checks

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

var fleetPath = flag.String("fleet", "", "path to fleet.json dump")

type fleetSite struct {
	SiteID string `json:"site_id"`
	Domain string `json:"domain"`
	Pages  []struct {
		PageID        string `json:"page_id"`
		Name          string `json:"name"`
		URL           string `json:"url"`
		NeverDeployed bool   `json:"never_deployed"`
	} `json:"pages"`
	PageComponents []struct {
		PageID string `json:"page_id"`
		Slot   string `json:"slot"`
		HTML   string `json:"html"`
	} `json:"page_components"`
	SiteComponents []struct {
		Slot string `json:"slot"`
		HTML string `json:"html"`
	} `json:"site_components"`
}

func TestFleetFragmentBlastRadius(t *testing.T) {
	if *fleetPath == "" {
		t.Skip("no -fleet dump supplied")
	}
	raw, err := os.ReadFile(*fleetPath)
	if err != nil {
		t.Fatalf("read dump: %v", err)
	}
	var sites []fleetSite
	if err := json.Unmarshal(raw, &sites); err != nil {
		t.Fatalf("parse dump: %v", err)
	}

	totalFrag, totalFindings := 0, 0
	for _, s := range sites {
		// Rebuild exactly what findPhantomInternalLinks builds.
		var urls []string
		unbuilt := map[string]string{}
		built := map[string]bool{}
		pageName := map[string]string{}
		for _, p := range s.Pages {
			urls = append(urls, p.URL)
			pageName[p.PageID] = p.Name
			n := datahelpers.NormalizePagePath(p.URL)
			if p.NeverDeployed {
				if _, seen := unbuilt[n]; !seen {
					unbuilt[n] = p.PageID
				}
			} else {
				built[n] = true
			}
		}
		for n := range built {
			delete(unbuilt, n)
		}
		urls = append(urls, "/", "/index.html")
		delete(unbuilt, datahelpers.NormalizePagePath("/"))
		targets := sitePageTargets{valid: datahelpers.NewPageURLSet(urls), unbuilt: unbuilt}

		pageHTML := map[string]string{}
		pathByID := map[string]string{}
		for _, pc := range s.PageComponents {
			pageHTML[pc.PageID] += "\n" + pc.HTML
		}
		for _, p := range s.Pages {
			if _, ok := pageHTML[p.PageID]; ok && p.URL != "" {
				pathByID[datahelpers.NormalizePagePath(p.URL)] = p.PageID
			}
		}
		chrome := ""
		for _, sc := range s.SiteComponents {
			chrome += "\n" + sc.HTML
		}
		idx := newFragmentIndex(pageHTML, chrome, pathByID)

		counts := map[plKey]int{}
		for _, pc := range s.PageComponents {
			for _, h := range datahelpers.HrefOffsets(pc.HTML) {
				if _, f := datahelpers.SplitFragment(h.Value); f != "" {
					totalFrag++
				}
			}
			accumulateFragmentIssues(counts, "page_component", pageName[pc.PageID], pc.PageID, pc.Slot, pc.HTML, targets, idx)
		}
		for _, sc := range s.SiteComponents {
			for _, h := range datahelpers.HrefOffsets(sc.HTML) {
				if _, f := datahelpers.SplitFragment(h.Value); f != "" {
					totalFrag++
				}
			}
			accumulateFragmentIssues(counts, "site_component", "", "", sc.Slot, sc.HTML, targets, idx)
		}

		var lines []string
		for k, n := range counts {
			if k.issue != "dead_fragment_link" {
				continue
			}
			lines = append(lines, fmt.Sprintf("    %-14s %-28s %-18s %s  x%d", k.surface, k.pageName, k.slotName, k.href, n))
		}
		sort.Strings(lines)
		totalFindings += len(lines)
		if len(lines) > 0 {
			fmt.Printf("  %s: %d dead_fragment_link finding(s)\n", s.Domain, len(lines))
			for _, l := range lines {
				fmt.Println(l)
			}
		}
	}
	fmt.Printf("\nFLEET TOTAL: %d fragment-bearing hrefs scanned, %d dead_fragment_link findings\n", totalFrag, totalFindings)
}
