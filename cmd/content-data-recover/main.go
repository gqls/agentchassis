// FILE: cmd/content-data-recover/main.go
//
// Recover a page component's content_data from its OWN rendered_html, by
// inverting the component template — and refuse to emit anything that cannot be
// PROVEN correct by re-rendering.
//
// WHY THIS EXISTS. 32 template-backed components fleet-wide hold rendered_html
// and NO content_data (measured 2026-08-22; 27 of them carry a parked
// `required_fields_missing` finding, bugs_open/277's remaining half). They serve
// correctly and cannot be rebuilt from their own stored data.
//
// ⚠ THE TRAP THAT DICTATES THIS TOOL'S SHAPE — READ BEFORE CHANGING IT.
// datahelpers.ContentDataCanFillTemplate returns true when content_data holds
// ANY ONE of the template's top-level fields. So writing a SINGLE recovered
// field flips a component from "cannot regenerate" to "can regenerate", and the
// next regeneration renders the template with that one field and blanks every
// other under missingkey=zero — the 004/007 blanking family. A partial backfill
// converts a page that is SAFE BECAUSE IT IS UNFILLABLE into one that rebuilds
// nearly empty. Therefore:
//
//	THE ONLY OUTPUT THIS TOOL MAY EMIT IS A content_data THAT RE-RENDERS TO THE
//	STORED rendered_html BYTE FOR BYTE.
//
// That single gate is the whole safety argument. It is not a heuristic about
// extraction quality: if the round-trip holds, then regeneration reproduces what
// is being served today, which is exactly the property the write must preserve.
//
// HOW IT VERIFIES. The re-render uses text/template with `missingkey=zero` and
// the same FuncMap as actions.executeGoTemplate (call_agent.go:1170) — the
// function that renders every component on this estate. To keep that comparison
// honest the tool REFUSES any component whose template reads a key the
// RenderContext supplies (domain, colors, nav, year, …): those values do not
// live in content_data, so a match would prove nothing about a real render.
// Only templates that read content fields alone are attempted.
//
// INPUT is JSON on stdin (the estate's kubectl-exec-psql export convention, as
// in cmd/config-key-audit) so this needs no credentials of its own. OUTPUT is a
// report on stderr and, with -sql, guarded UPDATE statements on stdout.
//
// WHAT IT DELIBERATELY CANNOT DO: invent a value, fill a field it did not find
// in the HTML, or write anything for a component whose stored HTML is not its
// template's output at all (9 of the 27 are a whole tool page stored in a `hero`
// slot — a different defect, filed separately, and backfilling those would swap
// a working 16KB tool for a 2KB hero band).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"
	"text/template/parse"
)

// Row is one candidate, exported by the SQL in RUNBOOK "Export the recovery
// candidates".
type Row struct {
	WorkItemID      string `json:"work_item_id"`
	PageComponentID string `json:"page_component_id"`
	Component       string `json:"component"`
	Domain          string `json:"domain"`
	Page            string `json:"page"`
	Template        string `json:"html_template"`
	RenderedHTML    string `json:"rendered_html"`
	MissingFields   string `json:"missing_fields"`
}

// contextKeys are supplied by actions.RenderContext, not by content_data
// (component_library.go, contextToInterfaceMap). A template reading any of them
// cannot be verified by this tool, because the value would not come from the
// content_data we are about to write. Erring toward refusal is the safe
// direction: a refused row stays parked, which is today's behaviour.
var contextKeys = map[string]bool{
	"domain": true, "logo_text": true, "logo_url": true, "company_name": true,
	"tagline": true, "nav_items": true, "footer_nav_items": true, "current_page": true,
	"primary_color": true, "secondary_color": true, "accent_color": true,
	"text_color": true, "background_color": true, "year": true, "contact_email": true,
	"site_id": true, "InstanceID": true,
}

// funcMap mirrors actions.executeGoTemplate's FuncMap exactly. If that map
// gains a function, add it here or a template using it will fail to parse and
// be reported as unsupported (which is safe, but a silent loss of coverage).
var funcMap = template.FuncMap{
	"default": func(defaultVal, val interface{}) interface{} {
		if val == nil || val == "" {
			return defaultVal
		}
		return val
	},
	"eq":    func(a, b interface{}) bool { return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b) },
	"ne":    func(a, b interface{}) bool { return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b) },
	"lower": strings.ToLower,
	"upper": strings.ToUpper,
	"isset": func(val interface{}) bool {
		if val == nil {
			return false
		}
		if s, ok := val.(string); ok {
			return s != ""
		}
		return true
	},
	"safe": func(val interface{}) string {
		if val == nil {
			return ""
		}
		return fmt.Sprintf("%v", val)
	},
}

func main() {
	emitSQL := flag.Bool("sql", false, "emit guarded UPDATE statements on stdout for rows that round-trip")
	flag.Parse()

	var rows []Row
	if err := json.NewDecoder(os.Stdin).Decode(&rows); err != nil {
		fmt.Fprintf(os.Stderr, "content-data-recover: cannot read candidate JSON on stdin: %v\n", err)
		os.Exit(2)
	}

	var passed, failed, unsupported int
	var sqlOut []string
	byReason := map[string]int{}

	for _, r := range rows {
		data, reason := recoverRow(r)
		switch {
		case reason == "":
			passed++
			stmt, err := updateStatement(r, data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  !! %s %s/%s: cannot encode recovered data: %v\n",
					r.PageComponentID, r.Domain, r.Page, err)
				failed++
				passed--
				continue
			}
			sqlOut = append(sqlOut, stmt)
			fmt.Fprintf(os.Stderr, "  OK   %-24s %-28s %s  (%d field(s) recovered, round-trip byte-identical)\n",
				r.Component, r.Page, r.Domain, len(data))
		case strings.HasPrefix(reason, "unsupported:"):
			unsupported++
			byReason[reason]++
			fmt.Fprintf(os.Stderr, "  skip %-24s %-28s %s  %s\n", r.Component, r.Page, r.Domain, reason)
		default:
			failed++
			byReason[reason]++
			fmt.Fprintf(os.Stderr, "  FAIL %-24s %-28s %s  %s\n", r.Component, r.Page, r.Domain, reason)
		}
	}

	fmt.Fprintf(os.Stderr, "\ncandidates %d | recovered+proven %d | unsupported %d | failed %d\n",
		len(rows), passed, unsupported, failed)
	reasons := make([]string, 0, len(byReason))
	for k := range byReason {
		reasons = append(reasons, k)
	}
	sort.Strings(reasons)
	for _, k := range reasons {
		fmt.Fprintf(os.Stderr, "  %3d  %s\n", byReason[k], k)
	}

	if *emitSQL {
		for _, s := range sqlOut {
			fmt.Println(s)
		}
	}
}

// recoverRow returns the recovered content_data, or a reason it could not be
// proven. An empty reason means the round-trip held byte for byte.
func recoverRow(r Row) (map[string]interface{}, string) {
	if r.Template == "" || r.RenderedHTML == "" {
		return nil, "unsupported: empty template or rendered_html"
	}

	tmpl, err := template.New("c").Funcs(funcMap).Option("missingkey=zero").Parse(r.Template)
	if err != nil {
		return nil, "unsupported: template does not parse: " + err.Error()
	}
	for _, f := range topLevelFields(tmpl.Tree.Root) {
		if contextKeys[f] {
			return nil, "unsupported: template reads RenderContext key ." + f + " (not content_data)"
		}
	}

	m := &matcher{in: r.RenderedHTML}
	b := binding{}
	end, ok := m.matchList(tmpl.Tree.Root.Nodes, 0, b)
	if !ok || end != len(r.RenderedHTML) {
		return nil, "no template inversion matches the stored HTML"
	}

	data := b.toData()

	// THE GATE. Re-render with the recovered data and demand the exact bytes.
	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return nil, "round-trip execute failed: " + err.Error()
	}
	got := strings.ReplaceAll(sb.String(), "<no value>", "") // as RenderTemplate does
	if got != r.RenderedHTML {
		return nil, fmt.Sprintf("round-trip DIFFERS (%d bytes out vs %d stored)", len(got), len(r.RenderedHTML))
	}
	if len(data) == 0 {
		return nil, "unsupported: template is static — nothing to recover"
	}
	return data, ""
}

func topLevelFields(n parse.Node) []string {
	var out []string
	var walk func(parse.Node)
	walk = func(n parse.Node) {
		switch t := n.(type) {
		case *parse.ListNode:
			if t == nil {
				return
			}
			for _, c := range t.Nodes {
				walk(c)
			}
		case *parse.ActionNode:
			out = append(out, pipeFields(t.Pipe)...)
		case *parse.IfNode:
			out = append(out, pipeFields(t.Pipe)...)
			walk(t.List)
			walk(t.ElseList)
		case *parse.RangeNode:
			out = append(out, pipeFields(t.Pipe)...)
			walk(t.List)
			walk(t.ElseList)
		case *parse.WithNode:
			out = append(out, pipeFields(t.Pipe)...)
			walk(t.List)
			walk(t.ElseList)
		}
	}
	walk(n)
	return out
}

func pipeFields(p *parse.PipeNode) []string {
	var out []string
	if p == nil {
		return out
	}
	for _, cmd := range p.Cmds {
		for _, a := range cmd.Args {
			if f, ok := a.(*parse.FieldNode); ok && len(f.Ident) > 0 {
				out = append(out, f.Ident[0])
			}
		}
	}
	return out
}
