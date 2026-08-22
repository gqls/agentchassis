// P0 harness for features_open/035_FEATURE_component_hierarchy.md — the local
// render-walk proof, no cluster writes.
//
// The claim under test (035 §5 P0): the D4 composition walk needs NO executor or
// funcmap changes. So the executor below is a byte-faithful replica of
// executeGoTemplate (platform/orchestration/actions/call_agent.go:1170-1221,
// read 2026-08-22) minus only the *zap.Logger parameter, plus the "<no value>"
// strip the production render paths apply after execution; the instance-token
// rule is replicated from component_instance_scope.go:102-160 (read
// 2026-08-22). ALL composition logic lives OUTSIDE those replicas — if this
// file had to touch either to pass its checks, D4 would be refuted.
//
// Checks 0-7 print one line each; any FAIL exits 1. The fail-closed checks
// (2,3,4,7) INDUCE their failure — they are mutation tests by construction, per
// the mutate-to-prove rule. Nested go.mod keeps this out of the platform build
// (precedent: the provocation lane's paired-mode prototype).
package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// Replicas — do not edit these two blocks except to track the platform source.
// ---------------------------------------------------------------------------

// replica of component_instance_scope.go:90,102-160
var reNotSelectorSafe = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

func InstanceToken(function string, occurrence int) string {
	s := strings.ToLower(strings.TrimSpace(function))
	s = strings.Trim(reNotSelectorSafe.ReplaceAllString(s, "-"), "-")
	if s == "" {
		s = fmt.Sprintf("anon-%d", occurrence)
	}
	if occurrence <= 0 {
		return "c-" + s
	}
	return fmt.Sprintf("c-%s-%d", s, occurrence+1)
}

type InstanceCounter struct{ seen map[string]int }

func NewInstanceCounter() *InstanceCounter { return &InstanceCounter{seen: make(map[string]int)} }

func (c *InstanceCounter) Next(function string) string {
	key := strings.ToLower(strings.TrimSpace(function))
	tok := InstanceToken(function, c.seen[key])
	c.seen[key]++
	return tok
}

func InstanceTokensForPage(functions []string) []string {
	c := NewInstanceCounter()
	out := make([]string, len(functions))
	for i, fn := range functions {
		out[i] = c.Next(fn)
	}
	return out
}

// ---------------------------------------------------------------------------
// The walk under test (035 D4). This is the NEW code P0 exists to prove.
// ---------------------------------------------------------------------------

// Row mirrors the page_components columns the walk reads.
type Row struct {
	ID       string
	Parent   string // parent_instance_id; "" = top level
	Position int
	Slot     string // slot_name; child form is "<parent_slot>.<child_key>"
	Function string
	Template string
	Data     map[string]interface{}
}

// SlotSpec mirrors one entry of the parent schema's "slots" block (035 D3).
type SlotSpec struct {
	Key      string
	Required bool
}

const maxDepth = 3

var reChildKey = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func childKey(slotName string) (string, error) {
	i := strings.LastIndex(slotName, ".")
	if i < 0 {
		return "", fmt.Errorf("child slot_name %q lacks the <parent_slot>.<child_key> form", slotName)
	}
	key := slotName[i+1:]
	if !reChildKey.MatchString(key) {
		return "", fmt.Errorf("child key %q is not identifier-safe (%s); templates could not address it as .slots.<key>", key, reChildKey)
	}
	return key, nil
}

type walker struct {
	children map[string][]Row
	specs    map[string][]SlotSpec // parent function -> declared slots
	counter  *InstanceCounter
	tokens   []string // pre-order (document-order) record
	visited  map[string]bool
}

func (w *walker) renderNode(r Row, depth int, path map[string]bool) (string, error) {
	if depth > maxDepth {
		return "", fmt.Errorf("depth cap exceeded (%d > %d) at row %s", depth, maxDepth, r.ID)
	}
	if path[r.ID] {
		return "", fmt.Errorf("cycle detected at row %s", r.ID)
	}
	path[r.ID] = true
	defer delete(path, r.ID)
	w.visited[r.ID] = true

	// Pre-order token binding: document order is parent-then-children, so the
	// parent consumes its token before any child does (035 §6.3).
	tok := w.counter.Next(r.Function)
	w.tokens = append(w.tokens, tok)

	data := make(map[string]interface{}, len(r.Data)+2)
	for k, v := range r.Data {
		data[k] = v
	}
	data["InstanceID"] = tok // InstanceContentKey

	kids := w.children[r.ID]
	if len(kids) > 0 || len(w.specs[r.Function]) > 0 {
		slots := map[string]interface{}{}
		for _, k := range kids {
			key, err := childKey(k.Slot)
			if err != nil {
				return "", err
			}
			html, err := w.renderNode(k, depth+1, path)
			if err != nil {
				return "", err
			}
			if prev, dup := slots[key]; dup {
				slots[key] = prev.(string) + html // same key repeats: concatenate in position order
			} else {
				slots[key] = html
			}
		}
		for _, s := range w.specs[r.Function] {
			if !s.Required {
				continue
			}
			if v, ok := slots[s.Key]; !ok || v == "" {
				return "", fmt.Errorf("required slot %q empty on row %s (%s)", s.Key, r.ID, r.Function)
			}
		}
		data["slots"] = slots
	}
	return executeGoTemplate(r.Template, data)
}

// renderPage walks the rows exactly as assembleComponents walks a flat page,
// with composition below the section grain. Returns concatenated HTML and the
// pre-order token sequence.
func renderPage(rows []Row, specs map[string][]SlotSpec) (string, []string, error) {
	w := &walker{
		children: map[string][]Row{},
		specs:    specs,
		counter:  NewInstanceCounter(),
		visited:  map[string]bool{},
	}
	var tops []Row
	for _, r := range rows {
		if r.Parent == "" {
			tops = append(tops, r)
		} else {
			w.children[r.Parent] = append(w.children[r.Parent], r)
		}
	}
	sort.SliceStable(tops, func(i, j int) bool { return tops[i].Position < tops[j].Position })
	for p := range w.children {
		kids := w.children[p]
		sort.SliceStable(kids, func(i, j int) bool { return kids[i].Position < kids[j].Position })
		w.children[p] = kids
	}

	var out strings.Builder
	for _, t := range tops {
		html, err := w.renderNode(t, 1, map[string]bool{})
		if err != nil {
			return "", nil, err
		}
		out.WriteString(html)
	}

	// Completeness assertion — the load-bearing cycle/orphan guard. With one
	// parent pointer per row, a REACHABLE cycle cannot exist (a top row has no
	// parent), so every cycle manifests as rows the walk never reaches. A walk
	// that only kept a path-set would DROP those rows silently — the exact
	// silent omission D4 forbids.
	var unrendered []string
	for _, r := range rows {
		if !w.visited[r.ID] {
			unrendered = append(unrendered, r.ID)
		}
	}
	if len(unrendered) > 0 {
		sort.Strings(unrendered)
		return "", nil, fmt.Errorf("unrendered rows (cycle or orphaned parent ref): %s", strings.Join(unrendered, ", "))
	}
	return out.String(), w.tokens, nil
}

// ---------------------------------------------------------------------------
// Fixtures — realistic editorial-family templates, six-function funcmap only.
// ---------------------------------------------------------------------------

const (
	tplHero    = `<section class="hero" data-instance="{{.InstanceID}}"><h1>{{.title}}</h1></section>`
	tplProse   = `<div class="prose" data-instance="{{.InstanceID}}">{{.content}}</div>`
	tplFigure  = `<figure data-instance="{{.InstanceID}}"><img src="{{.src}}" alt="{{safe .alt}}">{{if isset .caption}}<figcaption>{{.caption}}</figcaption>{{end}}</figure>`
	tplQuote   = `<blockquote data-instance="{{.InstanceID}}">{{.quote}}</blockquote>`
	tplArticle = `<article class="insight" data-instance="{{.InstanceID}}">{{if isset .standfirst}}<p class="standfirst">{{.standfirst}}</p>{{end}}{{.slots.lead}}{{.slots.fig_installations}}{{.slots.quote}}</article>`
	// nesting fixture: a parent whose child is itself a parent
	tplWrap = `<div class="wrap" data-instance="{{.InstanceID}}">{{.slots.inner}}</div>`
)

var articleSpecs = map[string][]SlotSpec{
	"insight-article": {
		{Key: "lead", Required: true},
		{Key: "fig_installations", Required: false},
		{Key: "quote", Required: false},
	},
}

func flatRow(id string, pos int, fn, tpl string, data map[string]interface{}) Row {
	return Row{ID: id, Position: pos, Slot: fn, Function: fn, Template: tpl, Data: data}
}

// ---------------------------------------------------------------------------
// Checks
// ---------------------------------------------------------------------------

var failures int

func report(name string, err error) {
	if err != nil {
		failures++
		fmt.Printf("FAIL  %-34s %v\n", name, err)
		return
	}
	fmt.Printf("PASS  %-34s\n", name)
}

func main() {
	// 0 — degenerate case: a page with no composition renders byte-identical
	// to today's flat token-bind-then-concatenate loop (the opt-out proof).
	report("0 flat page byte-identity", func() error {
		rows := []Row{
			flatRow("h", 0, "hero", tplHero, map[string]interface{}{"title": "Robot demand"}),
			flatRow("f1", 1, "evidence-figure", tplFigure, map[string]interface{}{"src": "/a.jpg", "caption": "IFR 2024"}),
			flatRow("f2", 2, "evidence-figure", tplFigure, map[string]interface{}{"src": "/b.jpg"}),
		}
		got, _, err := renderPage(rows, nil)
		if err != nil {
			return err
		}
		var want strings.Builder // today's assembleComponents shape
		c := NewInstanceCounter()
		for _, r := range rows {
			d := map[string]interface{}{"InstanceID": c.Next(r.Function)}
			for k, v := range r.Data {
				d[k] = v
			}
			h, err := executeGoTemplate(r.Template, d)
			if err != nil {
				return err
			}
			want.WriteString(h)
		}
		if got != want.String() {
			return fmt.Errorf("composed walk diverges on a flat page:\n got %q\nwant %q", got, want.String())
		}
		return nil
	}())

	// 1 — depth-1 composition: children render into the parent's slots in
	// declared order, each equal to its standalone render under its token.
	report("1 depth-1 compose", func() error {
		rows := []Row{
			flatRow("h", 0, "hero", tplHero, map[string]interface{}{"title": "T"}),
			{ID: "a", Position: 1, Slot: "insight-article", Function: "insight-article", Template: tplArticle,
				Data: map[string]interface{}{"standfirst": "Why it matters"}},
			{ID: "c1", Parent: "a", Position: 0, Slot: "insight-article.lead", Function: "prose-block", Template: tplProse,
				Data: map[string]interface{}{"content": "Lead prose."}},
			{ID: "c2", Parent: "a", Position: 1, Slot: "insight-article.fig_installations", Function: "evidence-figure", Template: tplFigure,
				Data: map[string]interface{}{"src": "/ifr.jpg", "caption": "Installations"}},
			{ID: "c3", Parent: "a", Position: 2, Slot: "insight-article.quote", Function: "editorial-pullquote", Template: tplQuote,
				Data: map[string]interface{}{"quote": "A step change."}},
		}
		got, _, err := renderPage(rows, articleSpecs)
		if err != nil {
			return err
		}
		wantLead := `<div class="prose" data-instance="c-prose-block">Lead prose.</div>`
		wantFig := `<figure data-instance="c-evidence-figure"><img src="/ifr.jpg" alt="">` + `<figcaption>Installations</figcaption></figure>`
		wantQuote := `<blockquote data-instance="c-editorial-pullquote">A step change.</blockquote>`
		for _, frag := range []string{wantLead, wantFig, wantQuote, `<p class="standfirst">Why it matters</p>`} {
			if !strings.Contains(got, frag) {
				return fmt.Errorf("missing fragment %q in %q", frag, got)
			}
		}
		if !(strings.Index(got, wantLead) < strings.Index(got, wantFig) && strings.Index(got, wantFig) < strings.Index(got, wantQuote)) {
			return fmt.Errorf("slot order wrong in %q", got)
		}
		if !strings.HasPrefix(got, `<section class="hero"`) {
			return fmt.Errorf("hero not first in %q", got)
		}
		return nil
	}())

	// 2 — depth: three levels render (and the grandchild's HTML is present);
	// four levels are refused by the cap.
	report("2 depth cap (3 passes, 4 refused)", func() error {
		three := []Row{
			{ID: "w1", Position: 0, Slot: "wrap", Function: "wrap-outer", Template: tplWrap},
			{ID: "w2", Parent: "w1", Position: 0, Slot: "wrap.inner", Function: "wrap-inner", Template: tplWrap},
			{ID: "p", Parent: "w2", Position: 0, Slot: "wrap.inner", Function: "prose-block", Template: tplProse,
				Data: map[string]interface{}{"content": "deep"}},
		}
		got, _, err := renderPage(three, nil)
		if err != nil {
			return fmt.Errorf("3-level render should pass: %v", err)
		}
		if !strings.Contains(got, ">deep</div>") {
			return fmt.Errorf("grandchild content missing from %q", got)
		}
		four := append(append([]Row{}, three...), Row{ID: "p2", Parent: "p", Position: 0, Slot: "prose.x", Function: "prose-block", Template: tplProse})
		if _, _, err := renderPage(four, nil); err == nil || !strings.Contains(err.Error(), "depth cap") {
			return fmt.Errorf("4-level render should refuse on depth, got err=%v", err)
		}
		return nil
	}())

	// 3 — cycles are unreachable by construction and must FAIL the render via
	// the completeness assertion, never be silently dropped.
	report("3 cycle/orphan refusal", func() error {
		mutual := []Row{
			flatRow("h", 0, "hero", tplHero, map[string]interface{}{"title": "T"}),
			{ID: "b", Parent: "c", Position: 0, Slot: "x.inner", Function: "wrap-inner", Template: tplWrap},
			{ID: "c", Parent: "b", Position: 0, Slot: "x.inner", Function: "wrap-inner", Template: tplWrap},
		}
		if _, _, err := renderPage(mutual, nil); err == nil || !strings.Contains(err.Error(), "unrendered rows") {
			return fmt.Errorf("mutual cycle should refuse via completeness, got err=%v", err)
		}
		self := []Row{
			flatRow("h", 0, "hero", tplHero, map[string]interface{}{"title": "T"}),
			{ID: "s", Parent: "s", Position: 0, Slot: "x.inner", Function: "wrap-inner", Template: tplWrap},
		}
		if _, _, err := renderPage(self, nil); err == nil || !strings.Contains(err.Error(), "unrendered rows") {
			return fmt.Errorf("self-cycle should refuse via completeness, got err=%v", err)
		}
		return nil
	}())

	// 4 — a required slot with no child fails the parent render, naming the slot.
	report("4 required slot fails closed", func() error {
		rows := []Row{
			{ID: "a", Position: 0, Slot: "insight-article", Function: "insight-article", Template: tplArticle,
				Data: map[string]interface{}{"standfirst": "S"}},
			{ID: "c2", Parent: "a", Position: 1, Slot: "insight-article.fig_installations", Function: "evidence-figure", Template: tplFigure,
				Data: map[string]interface{}{"src": "/x.jpg"}},
		}
		_, _, err := renderPage(rows, articleSpecs)
		if err == nil || !strings.Contains(err.Error(), `required slot "lead"`) {
			return fmt.Errorf("missing required lead should refuse, got err=%v", err)
		}
		return nil
	}())

	// 5 — an optional absent slot renders as nothing: no error, no "<no value>".
	report("5 optional slot absent = empty", func() error {
		rows := []Row{
			{ID: "a", Position: 0, Slot: "insight-article", Function: "insight-article", Template: tplArticle},
			{ID: "c1", Parent: "a", Position: 0, Slot: "insight-article.lead", Function: "prose-block", Template: tplProse,
				Data: map[string]interface{}{"content": "Lead."}},
		}
		got, _, err := renderPage(rows, articleSpecs)
		if err != nil {
			return err
		}
		if strings.Contains(got, "no value") {
			return fmt.Errorf("unstripped missing-slot marker in %q", got)
		}
		if strings.Contains(got, "standfirst") {
			return fmt.Errorf("isset guard failed on absent field in %q", got)
		}
		if !strings.Contains(got, "Lead.") {
			return fmt.Errorf("present child lost in %q", got)
		}
		return nil
	}())

	// 6 — instance tokens: the composed walk's pre-order sequence equals the
	// canonical InstanceTokensForPage over the hand-flattened function list.
	report("6 instance-token threading", func() error {
		rows := []Row{
			flatRow("f0", 0, "evidence-figure", tplFigure, map[string]interface{}{"src": "/0.jpg"}),
			{ID: "a", Position: 1, Slot: "insight-article", Function: "insight-article", Template: tplArticle},
			{ID: "c1", Parent: "a", Position: 0, Slot: "insight-article.lead", Function: "prose-block", Template: tplProse,
				Data: map[string]interface{}{"content": "L"}},
			{ID: "c2", Parent: "a", Position: 1, Slot: "insight-article.fig_installations", Function: "evidence-figure", Template: tplFigure,
				Data: map[string]interface{}{"src": "/1.jpg"}},
			{ID: "c3", Parent: "a", Position: 2, Slot: "insight-article.fig_installations", Function: "evidence-figure", Template: tplFigure,
				Data: map[string]interface{}{"src": "/2.jpg"}},
			flatRow("f3", 2, "evidence-figure", tplFigure, map[string]interface{}{"src": "/3.jpg"}),
		}
		got, tokens, err := renderPage(rows, articleSpecs)
		if err != nil {
			return err
		}
		want := InstanceTokensForPage([]string{
			"evidence-figure", "insight-article", "prose-block", "evidence-figure", "evidence-figure", "evidence-figure",
		})
		if strings.Join(tokens, " ") != strings.Join(want, " ") {
			return fmt.Errorf("token sequence\n got %v\nwant %v", tokens, want)
		}
		// and the tokens really reached the markup, including deep ones
		for _, tok := range []string{`data-instance="c-evidence-figure-2"`, `data-instance="c-evidence-figure-4"`} {
			if !strings.Contains(got, tok) {
				return fmt.Errorf("token %s missing from markup %q", tok, got)
			}
		}
		return nil
	}())

	// 7 — a child key templates could not address is refused, nothing rendered.
	report("7 non-identifier child key refused", func() error {
		rows := []Row{
			{ID: "a", Position: 0, Slot: "insight-article", Function: "insight-article", Template: tplArticle},
			{ID: "c1", Parent: "a", Position: 0, Slot: "insight-article.Fig-A", Function: "prose-block", Template: tplProse,
				Data: map[string]interface{}{"content": "L"}},
		}
		_, _, err := renderPage(rows, articleSpecs)
		if err == nil || !strings.Contains(err.Error(), "identifier-safe") {
			return fmt.Errorf("bad child key should refuse, got err=%v", err)
		}
		return nil
	}())

	if failures > 0 {
		fmt.Printf("\n%d check(s) FAILED\n", failures)
		os.Exit(1)
	}
	fmt.Println("\nALL CHECKS PASSED — D4 walk holds with the executor and funcmap untouched")
}
