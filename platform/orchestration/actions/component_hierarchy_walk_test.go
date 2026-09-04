// Tests for the 035 P1 component-hierarchy walk.
//
// These re-assert against the PLATFORM types and the REAL RenderTemplate what
// the P0 harness proved against a replica on 2026-08-22
// (docs024_key_docs_latest/editorial_design_uplift/harness/composewalk). The
// harness answered "does the walk need executor or funcmap changes?" (no). These
// answer the question a harness cannot: "does it still hold against the seam
// this estate actually renders through?"
//
// Every fail-closed behaviour here INDUCES its failure rather than asserting a
// happy path, per the mutate-to-prove rule — a guard proven only by a quiet test
// is not proven. TestHierarchyWalkEmptyTemplateChildStamp goes further and runs
// BOTH arms of 035 §6.9's falsifier, which is how it discovered that one of the
// two remedies §6.9 offers does not work.
package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// renderViaPlatformSeam builds a hierarchyNodeRenderer that renders each node
// through the real RenderTemplate with a FRESH RenderContext per node — 035 §6.9.
func renderViaPlatformSeam(t *testing.T, templates map[string]string, data map[string]map[string]interface{}) hierarchyNodeRenderer {
	t.Helper()
	logger := zap.NewNop()
	return func(n hierarchyNode, token string, slots map[string]interface{}) (hierarchyRenderedNode, error) {
		rc := &RenderContext{Domain: "example.test", Year: "2026"}
		rc.ContentData = map[string]interface{}{}
		for k, v := range data[n.ID] {
			rc.ContentData[k] = v
		}
		if slots != nil {
			rc.ContentData["slots"] = slots
		}
		BindInstanceToken(rc, token)
		html, _, _, err := RenderTemplate(templates[n.ID], rc, logger)
		if err != nil {
			return hierarchyRenderedNode{}, err
		}
		return hierarchyRenderedNode{HTML: html, Stamp: rc.RenderedTemplateSHA}, nil
	}
}

func TestHierarchyWalkComposesChildrenIntoParentSlots(t *testing.T) {
	nodes := []hierarchyNode{
		{ID: "p", ParentID: "", Position: 1, SlotName: "insight-article", Function: "insight-article"},
		{ID: "c2", ParentID: "p", Position: 2, SlotName: "insight-article.quote", Function: "editorial-pullquote"},
		{ID: "c1", ParentID: "p", Position: 1, SlotName: "insight-article.lead", Function: "prose-block"},
	}
	templates := map[string]string{
		"p":  `<article>{{.slots.lead}}{{.slots.quote}}</article>`,
		"c1": `<p>lead</p>`,
		"c2": `<blockquote>q</blockquote>`,
	}
	specs := map[string][]hierarchySlotSpec{
		"insight-article": {{Key: "lead", Required: true}, {Key: "quote", Required: false}},
	}

	res, err := walkComponentHierarchy(nodes, specs, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	want := `<article><p>lead</p><blockquote>q</blockquote></article>`
	if got := res.Nodes["p"].HTML; got != want {
		t.Errorf("parent HTML = %q, want %q", got, want)
	}
	// Children keep their OWN rendered_html too — that is what makes per-child
	// scanning, diffing and history work (035 D4).
	if got := res.Nodes["c1"].HTML; got != `<p>lead</p>` {
		t.Errorf("child c1 HTML = %q", got)
	}
	if len(res.Nodes) != 3 {
		t.Errorf("rendered %d rows, want 3 (every row exactly once)", len(res.Nodes))
	}
	if len(res.TopLevel) != 1 || res.TopLevel[0] != "p" {
		t.Errorf("TopLevel = %v, want [p]", res.TopLevel)
	}
}

// The instance counter must advance in DOCUMENT order — parent, then its
// children — even though the parent's template executes last. P0's mutation
// bound tokens post-order and this is the check that caught it (035 §6.3,
// bugs_open/283).
func TestHierarchyWalkAllocatesTokensInPreOrder(t *testing.T) {
	nodes := []hierarchyNode{
		{ID: "p", Position: 1, SlotName: "insight-article", Function: "insight-article"},
		{ID: "c1", ParentID: "p", Position: 1, SlotName: "insight-article.lead", Function: "prose-block"},
		{ID: "c2", ParentID: "p", Position: 2, SlotName: "insight-article.quote", Function: "prose-block"},
		{ID: "t2", Position: 2, SlotName: "footer-note", Function: "prose-block"},
	}
	templates := map[string]string{"p": `<article>{{.slots.lead}}{{.slots.quote}}</article>`, "c1": `a`, "c2": `b`, "t2": `c`}

	res, err := walkComponentHierarchy(nodes, nil, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	// Document order: parent(insight-article), lead(prose-block #1),
	// quote(prose-block #2), then the second top-level prose-block #3.
	want := []string{"c-insight-article", "c-prose-block", "c-prose-block-2", "c-prose-block-3"}
	if strings.Join(res.TokenOrder, ",") != strings.Join(want, ",") {
		t.Errorf("token order = %v, want %v", res.TokenOrder, want)
	}
}

// A page with no parent_instance_id anywhere must walk exactly as the flat
// readers do — the opt-out claim behind 035 §7 / RFC_022. If this drifts,
// composition is no longer opt-in and every uncomposed page on the estate is in
// its blast radius.
func TestHierarchyWalkLeavesAFlatPageUnchanged(t *testing.T) {
	nodes := []hierarchyNode{
		{ID: "a", Position: 2, SlotName: "b", Function: "prose-block"},
		{ID: "b", Position: 1, SlotName: "a", Function: "hero"},
	}
	templates := map[string]string{"a": `<p data-i="{{.InstanceID}}">A</p>`, "b": `<h1 data-i="{{.InstanceID}}">B</h1>`}

	res, err := walkComponentHierarchy(nodes, nil, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if res.Nodes["b"].HTML != `<h1 data-i="c-hero">B</h1>` {
		t.Errorf("hero = %q", res.Nodes["b"].HTML)
	}
	if res.Nodes["a"].HTML != `<p data-i="c-prose-block">A</p>` {
		t.Errorf("prose = %q", res.Nodes["a"].HTML)
	}
	// position ordering, not input order
	if strings.Join(res.TopLevel, ",") != "b,a" {
		t.Errorf("TopLevel = %v, want [b a] (position order)", res.TopLevel)
	}
}

// THE COMPLETENESS ASSERTION. With one parent pointer per row a cycle is
// UNREACHABLE, so the path-set never fires — the cycle shows up as rows the walk
// never visits. A walk guarded only by a path-set returns "successfully" minus
// those rows, which is the silent omission 035 D4 forbids.
func TestHierarchyWalkRefusesAnUnreachableCycle(t *testing.T) {
	nodes := []hierarchyNode{
		{ID: "top", Position: 1, SlotName: "hero", Function: "hero"},
		{ID: "x", ParentID: "y", Position: 1, SlotName: "hero.x", Function: "prose-block"},
		{ID: "y", ParentID: "x", Position: 1, SlotName: "hero.y", Function: "prose-block"},
	}
	templates := map[string]string{"top": `<h1>t</h1>`, "x": `x`, "y": `y`}

	_, err := walkComponentHierarchy(nodes, nil, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
	if err == nil {
		t.Fatal("walk succeeded on a mutual cycle — the completeness assertion is not firing. " +
			"Note the top-level row renders fine, so a walk without this assertion looks entirely healthy.")
	}
	// It must NAME the rows, not merely refuse: an unnamed refusal on a page of
	// 30 sections tells an operator nothing.
	if !strings.Contains(err.Error(), "x") || !strings.Contains(err.Error(), "y") {
		t.Errorf("error must name the unrendered rows, got: %v", err)
	}
}

func TestHierarchyWalkRefusesAnOrphanParentReference(t *testing.T) {
	nodes := []hierarchyNode{
		{ID: "top", Position: 1, SlotName: "hero", Function: "hero"},
		{ID: "kid", ParentID: "not-on-this-page", Position: 1, SlotName: "hero.kid", Function: "prose-block"},
	}
	templates := map[string]string{"top": `<h1>t</h1>`, "kid": `k`}

	_, err := walkComponentHierarchy(nodes, nil, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
	if err == nil {
		t.Fatal("walk accepted a child whose parent is not on the page; promoting it to top level would " +
			"put content on the page in a place nobody chose")
	}
}

func TestHierarchyWalkEnforcesRequiredSlotsAndAllowsOptionalOnes(t *testing.T) {
	base := []hierarchyNode{
		{ID: "p", Position: 1, SlotName: "insight-article", Function: "insight-article"},
	}
	templates := map[string]string{"p": `<article>{{.slots.lead}}{{.slots.quote}}</article>`}

	t.Run("required slot with no child fails the parent", func(t *testing.T) {
		specs := map[string][]hierarchySlotSpec{"insight-article": {{Key: "lead", Required: true}}}
		_, err := walkComponentHierarchy(base, specs, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
		if err == nil {
			t.Fatal("a required slot with no child must fail the parent render, not render a hole")
		}
		if !strings.Contains(err.Error(), "lead") {
			t.Errorf("error should name the missing slot, got: %v", err)
		}
	})

	t.Run("optional slot absent renders empty", func(t *testing.T) {
		specs := map[string][]hierarchySlotSpec{"insight-article": {{Key: "quote", Required: false}}}
		res, err := walkComponentHierarchy(base, specs, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
		if err != nil {
			t.Fatalf("optional absent slot must not fail: %v", err)
		}
		if got := res.Nodes["p"].HTML; got != `<article></article>` {
			t.Errorf("parent HTML = %q, want empty slots to render as empty string", got)
		}
	})
}

func TestHierarchyWalkRefusesANonIdentifierChildKey(t *testing.T) {
	nodes := []hierarchyNode{
		{ID: "p", Position: 1, SlotName: "insight-article", Function: "insight-article"},
		// "fig-1" is not addressable as {{.slots.fig-1}} — under missingkey=zero
		// it would resolve to nothing and the child would silently vanish.
		{ID: "c", ParentID: "p", Position: 1, SlotName: "insight-article.fig-1", Function: "evidence-figure"},
	}
	templates := map[string]string{"p": `<article>{{.slots.lead}}</article>`, "c": `<figure/>`}

	_, err := walkComponentHierarchy(nodes, nil, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
	if err == nil {
		t.Fatal("a non-identifier child key must be refused at the walk, not lost at render time")
	}
}

func TestHierarchyWalkEnforcesTheDepthCap(t *testing.T) {
	nodes := []hierarchyNode{
		{ID: "l1", Position: 1, SlotName: "wrap", Function: "wrap"},
		{ID: "l2", ParentID: "l1", Position: 1, SlotName: "wrap.inner", Function: "wrap"},
		{ID: "l3", ParentID: "l2", Position: 1, SlotName: "wrap.inner", Function: "wrap"},
		{ID: "l4", ParentID: "l3", Position: 1, SlotName: "wrap.inner", Function: "wrap"},
	}
	templates := map[string]string{"l1": `<div>{{.slots.inner}}</div>`, "l2": `<div>{{.slots.inner}}</div>`,
		"l3": `<div>{{.slots.inner}}</div>`, "l4": `<div/>`}

	_, err := walkComponentHierarchy(nodes, nil, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
	if err == nil {
		t.Fatalf("depth %d must be refused by the cap of %d", 4, hierarchyMaxDepth)
	}
	if !strings.Contains(err.Error(), "depth cap") {
		t.Errorf("expected a depth-cap error, got: %v", err)
	}
}

// 035 §6.9's FALSIFIER, both arms.
//
// RenderTemplate does not RETURN its provenance digest; it MUTATES it onto the
// RenderContext — and its empty-template branch returns early WITHOUT setting the
// field, on the stated RFC_046 reasoning that "empty means unknown" and a token
// pointing at a template that renders nothing is worse than no token. That
// reasoning silently assumes a FRESH context.
//
// A test that renders only NON-empty children passes whether or not this is
// handled, which is why the fixture below gives one child an empty template.
func TestHierarchyWalkEmptyTemplateChildStamp(t *testing.T) {
	nodes := []hierarchyNode{
		{ID: "p", Position: 1, SlotName: "insight-article", Function: "insight-article"},
		{ID: "c1", ParentID: "p", Position: 1, SlotName: "insight-article.lead", Function: "prose-block"},
		{ID: "c2", ParentID: "p", Position: 2, SlotName: "insight-article.quote", Function: "editorial-pullquote"},
	}
	templates := map[string]string{
		"p":  `<article>{{.slots.lead}}{{.slots.quote}}</article>`,
		"c1": `<p>lead</p>`,
		"c2": ``, // the empty-template child
	}

	t.Run("fresh context per node: the empty-template child carries NO stamp", func(t *testing.T) {
		res, err := walkComponentHierarchy(nodes, nil, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
		if err != nil {
			t.Fatalf("walk failed: %v", err)
		}
		if res.Nodes["c2"].Stamp != "" {
			t.Errorf("empty-template child stamp = %q, want empty. A non-empty value here is a FALSE "+
				"provenance claim — RFC_046's whole subject.", res.Nodes["c2"].Stamp)
		}
		if res.Nodes["c1"].Stamp == "" {
			t.Error("non-empty child must carry its own stamp — otherwise this test would pass vacuously")
		}
		if res.Nodes["p"].Stamp == res.Nodes["c1"].Stamp {
			t.Error("parent and child must not share a stamp; they render different templates")
		}
	})

	// THE DISCRIMINATING ARM, and the reason this test is worth its length.
	// 035 §6.9 offers two remedies: "capture the digest immediately after each
	// node's RenderTemplate call, OR give each node its own context." This arm
	// runs the FIRST one faithfully — a shared context, read immediately after
	// every call — and shows it does NOT work: because the empty-template branch
	// returns before touching the field, an immediate read still observes the
	// PREVIOUS node's digest. Only a fresh context (or clearing the field) is
	// sufficient. Recorded as a correction to §6.9.
	t.Run("shared context with immediate capture: the documented first remedy FAILS", func(t *testing.T) {
		logger := zap.NewNop()
		shared := &RenderContext{Domain: "example.test", Year: "2026"}
		shared.ContentData = map[string]interface{}{}
		render := func(n hierarchyNode, token string, slots map[string]interface{}) (hierarchyRenderedNode, error) {
			if slots != nil {
				shared.ContentData["slots"] = slots
			} else {
				delete(shared.ContentData, "slots")
			}
			BindInstanceToken(shared, token)
			html, _, _, err := RenderTemplate(templates[n.ID], shared, logger)
			if err != nil {
				return hierarchyRenderedNode{}, err
			}
			// Immediately after the call, exactly as §6.9's first remedy says.
			return hierarchyRenderedNode{HTML: html, Stamp: shared.RenderedTemplateSHA}, nil
		}

		res, err := walkComponentHierarchy(nodes, nil, NewInstanceCounter(), render)
		if err != nil {
			t.Fatalf("walk failed: %v", err)
		}
		if res.Nodes["c2"].Stamp == "" {
			t.Skip("empty-template child came back unstamped on a shared context — RenderTemplate now clears " +
				"the field on entry (the 357/RFC_046 primitive fix). Good: §6.9's first remedy is sound again " +
				"and this arm has served its purpose.")
		}
		if res.Nodes["c2"].Stamp != res.Nodes["c1"].Stamp {
			t.Fatalf("expected the empty-template child to inherit its SIBLING's digest on a shared context; "+
				"got c2=%q c1=%q. If this changed, re-derive §6.9 before trusting either remedy.",
				res.Nodes["c2"].Stamp, res.Nodes["c1"].Stamp)
		}
		t.Logf("CONFIRMED 035 §6.9 correction: with a shared context, immediate capture still gives the "+
			"empty-template child its sibling's digest (%s). Only a fresh context per node is sufficient.",
			res.Nodes["c2"].Stamp[:12])
	})
}

func TestHierarchySlotsFromSchema(t *testing.T) {
	got := hierarchySlotsFromSchema(map[string]interface{}{
		"slots": []interface{}{
			map[string]interface{}{"key": "lead", "required": true},
			map[string]interface{}{"key": "quote"},
			map[string]interface{}{"required": true}, // no key — skipped
		},
	})
	if len(got) != 2 || got[0].Key != "lead" || !got[0].Required || got[1].Key != "quote" || got[1].Required {
		t.Errorf("parsed slots = %+v", got)
	}
	// A component with no slots block is NOT a composite — the opt-in default.
	if s := hierarchySlotsFromSchema(map[string]interface{}{"fields": map[string]interface{}{}}); s != nil {
		t.Errorf("a schema with no slots block must yield nil, got %+v", s)
	}
}

// REPEATED SLOT KEYS CONCATENATE IN POSITION ORDER.
//
// This is the migration primitive, and it was written without a test until
// 2026-09-03. Decomposing an existing `article-body` — one llm-owned `content`
// blob, 360 of them on the estate — means partitioning it at heading boundaries
// into N children that all carry the SAME slot key, because the section count
// varies per page (measured: avg 4.8 h2 per row, max 16 h3) and named slots
// cannot express a variable count.
//
// So byte-equivalence after decomposition depends entirely on this: the walk must
// reassemble the children in POSITION ORDER, adding nothing and dropping nothing.
// If the later child ever replaced the earlier one instead, a decomposed article
// would silently lose every section but its last — and the page would still
// render, which is what makes it worth a test rather than a comment.
func TestHierarchyWalkConcatenatesRepeatedSlotKeysInPositionOrder(t *testing.T) {
	nodes := []hierarchyNode{
		{ID: "p", Position: 1, SlotName: "article-body", Function: "article-body"},
		// Deliberately out of input order, to prove POSITION decides and not
		// arrival order — the partition is only lossless if order is preserved.
		{ID: "c3", ParentID: "p", Position: 3, SlotName: "article-body.section", Function: "prose-block"},
		{ID: "c1", ParentID: "p", Position: 1, SlotName: "article-body.section", Function: "prose-block"},
		{ID: "c2", ParentID: "p", Position: 2, SlotName: "article-body.section", Function: "prose-block"},
	}
	templates := map[string]string{
		"p":  `<article>{{.slots.section}}</article>`,
		"c1": `<h2>One</h2><p>first</p>`,
		"c2": `<h2>Two</h2><p>second</p>`,
		"c3": `<h2>Three</h2><p>third</p>`,
	}

	res, err := walkComponentHierarchy(nodes, nil, NewInstanceCounter(), renderViaPlatformSeam(t, templates, nil))
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}

	// The reassembly property the migration rests on: the parent's HTML is the
	// children's bytes in position order, wrapped by the parent's own template.
	want := `<article><h2>One</h2><p>first</p><h2>Two</h2><p>second</p><h2>Three</h2><p>third</p></article>`
	if got := res.Nodes["p"].HTML; got != want {
		t.Errorf("repeated-key reassembly is not a lossless in-order concatenation.\n got=%q\nwant=%q", got, want)
	}
	if len(res.Nodes) != 4 {
		t.Errorf("every child must still render its own row: got %d of 4", len(res.Nodes))
	}
}
