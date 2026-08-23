package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

type suite struct {
	Patterns  []string `json:"patterns"`
	Forbidden []string `json:"forbidden"`
	Permitted []string `json:"permitted"`
}

// Re-runs the apis.uk ban suite under Go/RE2 — the engine production uses.
// Python's `re` and RE2 differ (inline flags, backtracking), so a suite proven
// in Python is proven in the wrong place.
func main() {
	var s suite
	f, _ := os.Open(os.Args[1])
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		fmt.Println("decode:", err)
		os.Exit(1)
	}
	res := make([]*regexp.Regexp, 0, len(s.Patterns))
	for _, p := range s.Patterns {
		re, err := regexp.Compile("(?i)" + p)
		if err != nil {
			fmt.Printf("  COMPILE FAIL %q: %v\n", p, err)
			os.Exit(1)
		}
		res = append(res, re)
	}
	bad := 0
	for _, t := range s.Forbidden {
		hit := false
		for _, re := range res {
			if re.MatchString(t) {
				hit = true
				break
			}
		}
		if !hit {
			bad++
			fmt.Printf("  GAP under RE2: %q\n", t)
		}
	}
	for _, t := range s.Permitted {
		for _, re := range res {
			if re.MatchString(t) {
				bad++
				fmt.Printf("  FALSE POSITIVE under RE2: %q <- %s\n", t, re.String())
			}
		}
	}
	fmt.Printf("\nRE2: patterns=%d forbidden=%d permitted=%d  problems=%d\n",
		len(res), len(s.Forbidden), len(s.Permitted), bad)
	if bad == 0 {
		fmt.Println("RESULT: the suite holds in the PRODUCTION engine, not just in Python.")
	}
}
