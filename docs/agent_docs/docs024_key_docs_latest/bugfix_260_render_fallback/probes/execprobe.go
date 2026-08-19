// execprobe — how many STORED sections would fail to render on a rerender today?
//
// Faithful because contextToInterfaceMap merges ContentData at the TOP LEVEL of
// the data map (component_library.go:1266-1268), and missingkey=zero makes every
// absent site-level field safe (bugs_open/260 §2's own table). So a failure here
// is caused by content_data and nothing else. It is CONSERVATIVE, not inflated:
// it cannot manufacture a failure from a missing site field.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"
)

type sec struct {
	PcID  string                 `json:"pc_id"`
	CName string                 `json:"cname"`
	CFunc string                 `json:"cfunc"`
	Tmpl  string                 `json:"tmpl"`
	CD    map[string]interface{} `json:"cd"`
}

func main() {
	raw, _ := os.ReadFile(os.Args[1])
	var secs []sec
	if err := json.Unmarshal(raw, &secs); err != nil {
		panic(err)
	}
	fm := template.FuncMap{
		"default": func(a, b interface{}) interface{} { return a },
		"eq":      func(a, b interface{}) bool { return false },
		"ne":      func(a, b interface{}) bool { return false },
		"lower":   strings.ToLower,
		"upper":   strings.ToUpper,
		"isset":   func(v interface{}) bool { return false },
		"safe":    func(v interface{}) string { return "" },
	}
	run := func(t string, d map[string]interface{}) error {
		tmpl, err := template.New("component").Option("missingkey=zero").Funcs(fm).Parse(t)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		var b bytes.Buffer
		return tmpl.Execute(&b, d)
	}
	// CONTROLS, both directions — the 260 §2 A/B pair in miniature.
	if run(`{{range $s := .steps}}{{$s.title}}{{end}}`, map[string]interface{}{"steps": "prose"}) == nil {
		panic("control failed: a string where an array is ranged did NOT error")
	}
	if err := run(`{{range $s := .steps}}{{$s.title}}{{end}}`,
		map[string]interface{}{"steps": []interface{}{map[string]interface{}{"title": "x"}}}); err != nil {
		panic("control failed: the correctly-shaped value errored: " + err.Error())
	}

	byComp := map[string]int{}
	var fails []string
	for _, s := range secs {
		if err := run(s.Tmpl, s.CD); err != nil {
			msg := err.Error()
			if i := strings.Index(msg, "executing"); i > 0 {
				msg = msg[i:]
			}
			if len(msg) > 120 {
				msg = msg[:120]
			}
			byComp[s.CName]++
			fails = append(fails, fmt.Sprintf("%-30s %-40s %s", s.CName, s.PcID, msg))
		}
	}
	fmt.Printf("stored sections executed: %d\n", len(secs))
	fmt.Printf("EXECUTE FAILURES: %d (%.2f%%)\n\n", len(fails), 100*float64(len(fails))/float64(len(secs)))
	type kv struct {
		k string
		v int
	}
	var agg []kv
	for k, v := range byComp {
		agg = append(agg, kv{k, v})
	}
	sort.Slice(agg, func(i, j int) bool { return agg[i].v > agg[j].v })
	for _, a := range agg {
		fmt.Printf("  %-34s %d\n", a.k, a.v)
	}
	fmt.Println()
	sort.Strings(fails)
	for i, f := range fails {
		if i >= 25 {
			fmt.Printf("  … and %d more\n", len(fails)-25)
			break
		}
		fmt.Println("  ", f)
	}
}
