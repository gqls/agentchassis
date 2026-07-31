// Census of the content-duplication in-remit set, computed with the SHIPPED
// functions rather than a restatement of them. See dedup_census_shipped.md.
//
// Usage: go run . [-legacy] < pc_dump.jsonl
//   default   the current rule: same page + same slot + byte-identical blob
//   -legacy   the pre-43492ec94 rule: same page + identical normalised text,
//             which found exactly one group fleet-wide and it was a false
//             positive that would have deleted a live section
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

type rec struct {
	PCID   string `json:"pc_id"`
	PageID string `json:"page_id"`
	Pos    int    `json:"position"`
	Slot   string `json:"slot"`
	Comp   string `json:"comp"`
	Domain string `json:"domain"`
	URL    string `json:"url"`
	Raw    string `json:"raw"`
}

func main() {
	legacy := flag.Bool("legacy", false, "use the pre-fix prose-identity rule")
	flag.Parse()

	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)

	groups := map[string][]rec{}
	var total, eligible int
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var r rec
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			fmt.Fprintln(os.Stderr, "skip unparseable line:", err)
			continue
		}
		total++
		text := datahelpers.NormaliseSectionText(r.Raw)
		// The floor is a PROSE question ("is this a meaningful section at all")
		// and is identical in both halves of the shipped code:
		// check_content_duplication.go findIdenticalSamePage and
		// remove_duplicate_page_sections_action.go.
		if len(text) < 80 {
			continue
		}
		eligible++
		var k string
		if *legacy {
			k = r.PageID + "\x01" + text
		} else {
			k = r.PageID + "\x01" + datahelpers.SectionIdentityKey(r.Slot, r.Raw)
		}
		groups[k] = append(groups[k], r)
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		os.Exit(1)
	}

	var keys []string
	for k, rs := range groups {
		if len(rs) >= 2 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	deleted := 0
	for _, k := range keys {
		rs := groups[k]
		sort.Slice(rs, func(i, j int) bool { return rs[i].Pos < rs[j].Pos })
		deleted += len(rs) - 1
		distinct := map[string]bool{}
		for _, r := range rs {
			distinct[r.Raw] = true
		}
		fmt.Printf("  %s%s  rows=%d  WOULD DELETE %d (keep pos %d)\n",
			rs[0].Domain, rs[0].URL, len(rs), len(rs)-1, rs[0].Pos)
		for _, r := range rs {
			fmt.Printf("      slot=%-20s pos=%-3d comp=%.8s\n", r.Slot, r.Pos, r.Comp)
		}
		if len(distinct) > 1 {
			fmt.Printf("      *** %d DISTINCT raw blobs — the rule collapsed a real difference ***\n", len(distinct))
		} else {
			fmt.Printf("      raw blobs byte-identical\n")
		}
	}

	rule := "SHIPPED RULE (page + slot + byte-identical blob)"
	if *legacy {
		rule = "LEGACY RULE (page + identical normalised text)"
	}
	fmt.Printf("sections=%d  eligible(>=80 chars prose)=%d\n", total, eligible)
	fmt.Printf("%s: groups=%d  rows_deleted=%d\n", rule, len(keys), deleted)
}
