// FILE: platform/orchestration/actions/component_hierarchy_walk.go
//
// The composition walk for features_open/035 (component hierarchy), phase P1.
//
// A composed section is ONE entry in pages.sections (the parent) whose children
// are ordinary page_components rows carrying parent_instance_id (035 D1/D2).
// This file owns the traversal and nothing else: grouping, ordering, the depth
// cap, the cycle/completeness guard, slot assembly, required-slot enforcement,
// and — the subtle one — the ORDER in which per-instance tokens are consumed.
//
// It deliberately does NOT render. Rendering a node needs a render context that
// only the calling path knows how to build (base ⊕ stored content_data ⊕ fresh
// resolved_data, schema, markdown stripping, …), and those differ per path. So
// the caller supplies a hierarchyNodeRenderer and this file supplies the shape.
// That is what lets one walk serve several callers rather than each growing its
// own — 035 D4.6's "one walk implementation, two callers".
//
// NAMING. Nothing here may be called RenderTemplate* : render_seam_one_spelling_test.go
// asserts by AST that the exported RenderTemplate* symbols are exactly
// [RenderTemplate, RenderTemplateWithMap]. This walk calls RenderTemplate through
// the caller's closure and constructs no template itself, so it adds nothing to
// declaredTemplateExecutors either. "composition" is avoided as a name because on
// this estate that word already belongs to the CSS/theme family
// (resolve_composition_*.go); this feature's own name is component hierarchy.
//
// PORTED FROM A PROVEN PROTOTYPE. The traversal below is the P0 harness walk
// (docs024_key_docs_latest/editorial_design_uplift/harness/composewalk/main.go),
// which passed eight checks on 2026-08-22 against a byte-faithful replica of
// executeGoTemplate with the funcmap untouched — the falsifier for 035 D4 being
// "if this needs executor or funcmap changes, D4 is wrong". Two of those checks
// were proven able to fail by mutating the walk. The same behaviours are
// re-asserted here against the platform types in component_hierarchy_walk_test.go.
package actions

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// hierarchyMaxDepth is 035 D4.3's depth cap. Depth 1 is a top-level section, so
// this permits a parent, a child and a grandchild. It is a backstop against a
// pathological tree, NOT the cycle guard — see walkComponentHierarchy's
// completeness assertion for why a depth cap alone would be the wrong instrument.
const hierarchyMaxDepth = 3

// reHierarchyChildKey constrains the <child_key> half of a child's slot_name.
// Templates address a child as {{.slots.<key>}}, so the key has to be a Go
// template field identifier: a non-identifier key would not fail loudly, it
// would silently resolve to nothing under missingkey=zero and the child would
// vanish from the parent's output.
var reHierarchyChildKey = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// hierarchyNode is one page_components row as the walk needs to see it. It is
// deliberately narrow: the walk reads structure, never content.
type hierarchyNode struct {
	ID       string // page_components.id
	ParentID string // page_components.parent_instance_id; "" = top-level section
	Position int    // orders siblings WITHIN the parent (035 D1)
	SlotName string // child form is "<parent_slot>.<child_key>" (035 D2)
	Function string // content_components.function — what the token counter keys on
}

// hierarchySlotSpec is one entry of a parent schema's "slots" block (035 D3).
type hierarchySlotSpec struct {
	Key      string
	Required bool
}

// hierarchyRenderedNode is what a caller's renderer returns for ONE node.
//
// Stamp exists as a RETURN VALUE, not as something the walk reads off a shared
// context, and that is load-bearing rather than tidy — 035 §6.9. RenderTemplate
// does not return its provenance digest; it MUTATES it onto the RenderContext
// (component_library.go:1081), once per call. A walk that reused one context
// across nodes would therefore end holding only the LAST node's digest, and —
// worse — the empty-template branch returns early WITHOUT setting the field, so
// an empty-template child would inherit its predecessor's stamp: a false
// provenance claim of exactly the kind RFC_046 exists to prevent.
//
// Today's four live readers of that field are safe only because each renders
// exactly one template per context — a property held by convention across four
// files in four lanes, which nothing enforces. This walk is the first code on
// the estate that renders many templates per traversal. Requiring the stamp to
// come BACK from each node's renderer makes per-node capture structural instead
// of a rule in a comment: the owner's 2026-08-02 §2 ruling, that authority on a
// shared seam ships as a mechanism and not as a doc comment a later session
// need not read.
//
// The primitive-level fix (clear the field on entry to RenderTemplate, so
// "unknown" is guaranteed rather than coincidental) is architecture-scope and
// belongs to the bugfix_357/RFC_046 lane. It is deliberately NOT bolted in here.
type hierarchyRenderedNode struct {
	HTML  string
	Stamp string
}

// hierarchyNodeRenderer renders one node and returns its HTML and provenance
// stamp. slots holds each declared child key mapped to that child's ALREADY
// rendered HTML (035 D4: templates see pre-rendered slots; no {{template}} and
// no new funcmap entries are added for this feature). slots is nil for a leaf.
//
// token is the node's per-instance element-id token, already allocated in
// document order — see walkComponentHierarchy.
type hierarchyNodeRenderer func(node hierarchyNode, token string, slots map[string]interface{}) (hierarchyRenderedNode, error)

// hierarchyTokenSource allocates per-instance tokens. *InstanceCounter satisfies
// it (component_instance_scope.go); the interface exists so a caller that must
// NOT advance the page counter — a single-section path re-composing one parent —
// can supply its own derivation without this file knowing which case it is in.
type hierarchyTokenSource interface {
	Next(function string) string
}

// hierarchyWalkResult carries one entry per row, plus the order tokens were
// consumed in. Every row the walk was handed appears in Nodes exactly once, or
// the walk returned an error instead — see the completeness assertion.
type hierarchyWalkResult struct {
	// Nodes maps page_components.id to that row's own rendered output. A
	// parent's HTML has its children already embedded (which is what keeps every
	// downstream consumer of stored/served HTML seeing one section, one string —
	// 035 D4), while each child ALSO carries its own row here, which is what
	// makes per-child scanning, diffing and history work.
	Nodes map[string]hierarchyRenderedNode
	// TopLevel lists top-level row ids in render order — the sections the page
	// actually concatenates.
	TopLevel []string
	// TokenOrder is the document order tokens were consumed in, exposed so a
	// caller can assert it against a flat render of the same rows (035 §6.3).
	TokenOrder []string
}

// hierarchyChildKey extracts <child_key> from a child's "<parent_slot>.<child_key>"
// slot_name.
func hierarchyChildKey(slotName string) (string, error) {
	i := strings.LastIndex(slotName, ".")
	if i < 0 {
		return "", fmt.Errorf("child slot_name %q lacks the <parent_slot>.<child_key> form (035 D2)", slotName)
	}
	key := slotName[i+1:]
	if !reHierarchyChildKey.MatchString(key) {
		return "", fmt.Errorf("child key %q from slot_name %q is not identifier-safe (%s): "+
			"a template could not address it as {{.slots.%s}}, and under missingkey=zero it would "+
			"resolve to nothing rather than fail", key, slotName, reHierarchyChildKey, key)
	}
	return key, nil
}

type hierarchyWalker struct {
	children map[string][]hierarchyNode
	specs    map[string][]hierarchySlotSpec
	tokens   hierarchyTokenSource
	render   hierarchyNodeRenderer

	out        map[string]hierarchyRenderedNode
	tokenOrder []string
	visited    map[string]bool
}

func (w *hierarchyWalker) renderNode(n hierarchyNode, depth int, path map[string]bool) (string, error) {
	if depth > hierarchyMaxDepth {
		return "", fmt.Errorf("component hierarchy: depth cap exceeded (%d > %d) at row %s (slot %q)",
			depth, hierarchyMaxDepth, n.ID, n.SlotName)
	}
	if path[n.ID] {
		return "", fmt.Errorf("component hierarchy: cycle detected at row %s (slot %q)", n.ID, n.SlotName)
	}
	path[n.ID] = true
	defer delete(path, n.ID)
	w.visited[n.ID] = true

	// PRE-ORDER token allocation. Document order is parent-then-children, so the
	// parent must consume its token BEFORE any child does, even though the
	// parent's template executes last (it needs its children's HTML first).
	// Allocating on the way down and rendering on the way up is the whole point:
	// the P0 harness's mutation test bound tokens post-order and check 6 caught
	// it. Get this wrong and per-instance element ids drift between a composed
	// render and a flat one of the same rows (035 §6.3, bugs_open/283).
	token := w.tokens.Next(n.Function)
	w.tokenOrder = append(w.tokenOrder, token)

	var slots map[string]interface{}
	kids := w.children[n.ID]
	specs := w.specs[n.Function]
	if len(kids) > 0 || len(specs) > 0 {
		slots = make(map[string]interface{}, len(kids))
		for _, k := range kids {
			key, err := hierarchyChildKey(k.SlotName)
			if err != nil {
				return "", err
			}
			childHTML, err := w.renderNode(k, depth+1, path)
			if err != nil {
				return "", err
			}
			// A repeated key (a "fig-*" style wildcard slot) concatenates in
			// position order rather than the later child silently replacing the
			// earlier one.
			if prev, dup := slots[key]; dup {
				slots[key] = prev.(string) + childHTML
			} else {
				slots[key] = childHTML
			}
		}
		// A required slot with no child fails the PARENT's render (035 D4.4).
		// Fail-closed: the alternative is a parent that renders with a hole in
		// it, which is bugs_open/018's class — a page that looks fine and is not.
		for _, s := range specs {
			if !s.Required {
				continue
			}
			v, ok := slots[s.Key]
			if !ok || v == "" {
				return "", fmt.Errorf("component hierarchy: required slot %q has no child on row %s (%s)",
					s.Key, n.ID, n.Function)
			}
		}
	}

	rendered, err := w.render(n, token, slots)
	if err != nil {
		return "", fmt.Errorf("component hierarchy: row %s (slot %q) failed to render: %w", n.ID, n.SlotName, err)
	}
	w.out[n.ID] = rendered
	return rendered.HTML, nil
}

// walkComponentHierarchy renders every row exactly once, children into their
// parents' slots, and returns each row's own output.
//
// specs maps a PARENT's function to its declared slots (035 D3's "slots" block,
// read from content_components.input_schema). A function absent from specs simply
// declares no required slots; children still render.
//
// tokens must be the SAME counter the calling path uses for the rest of the page,
// or per-instance ids will not match a flat render of the same rows.
func walkComponentHierarchy(
	nodes []hierarchyNode,
	specs map[string][]hierarchySlotSpec,
	tokens hierarchyTokenSource,
	render hierarchyNodeRenderer,
) (*hierarchyWalkResult, error) {
	if render == nil {
		return nil, fmt.Errorf("component hierarchy: no node renderer supplied")
	}
	if tokens == nil {
		return nil, fmt.Errorf("component hierarchy: no token source supplied")
	}

	w := &hierarchyWalker{
		children: map[string][]hierarchyNode{},
		specs:    specs,
		tokens:   tokens,
		render:   render,
		out:      map[string]hierarchyRenderedNode{},
		visited:  map[string]bool{},
	}

	var tops []hierarchyNode
	byID := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		if byID[n.ID] {
			return nil, fmt.Errorf("component hierarchy: row %s appears twice in the input", n.ID)
		}
		byID[n.ID] = true
	}
	for _, n := range nodes {
		switch {
		case n.ParentID == "":
			tops = append(tops, n)
		case !byID[n.ParentID]:
			// A child whose parent is not in this page's row set. Fail rather
			// than render it as a top-level section: promoting an orphan would
			// put content on the page in a place nobody chose.
			return nil, fmt.Errorf("component hierarchy: row %s (slot %q) names parent %s, which is not on this page",
				n.ID, n.SlotName, n.ParentID)
		default:
			w.children[n.ParentID] = append(w.children[n.ParentID], n)
		}
	}

	// (position, id) — the same tie-break loadStoredSections uses, so the walk
	// and the flat readers cannot disagree about order on a tie.
	sortHierarchyNodes(tops)
	for p := range w.children {
		sortHierarchyNodes(w.children[p])
	}

	res := &hierarchyWalkResult{Nodes: w.out}
	for _, t := range tops {
		if _, err := w.renderNode(t, 1, map[string]bool{}); err != nil {
			return nil, err
		}
		res.TopLevel = append(res.TopLevel, t.ID)
	}

	// THE COMPLETENESS ASSERTION — the load-bearing cycle guard, and the reason
	// the path-set above is not sufficient on its own (035 D4.3, refined by P0).
	//
	// With one parent pointer per row a REACHABLE cycle cannot exist: a cycle's
	// members all have parents, so none of them is a top-level row, so the
	// traversal never enters it. The path-set therefore never fires in practice.
	// What a cycle actually produces is rows the walk never reaches — and a walk
	// guarded only by a path-set would return "successfully" having silently
	// DROPPED them, which is the exact silent omission 035 D4 forbids.
	//
	// Proven by mutation in P0: removing this assertion let a mutual cycle render
	// "successfully" minus its rows, and check 3 was the only one that failed.
	var unrendered []string
	for _, n := range nodes {
		if !w.visited[n.ID] {
			unrendered = append(unrendered, n.ID)
		}
	}
	if len(unrendered) > 0 {
		sort.Strings(unrendered)
		return nil, fmt.Errorf("component hierarchy: %d row(s) were never rendered (cycle, or a parent chain that "+
			"leaves the page): %s", len(unrendered), strings.Join(unrendered, ", "))
	}

	res.TokenOrder = w.tokenOrder
	return res, nil
}

func sortHierarchyNodes(ns []hierarchyNode) {
	sort.SliceStable(ns, func(i, j int) bool {
		if ns[i].Position != ns[j].Position {
			return ns[i].Position < ns[j].Position
		}
		return ns[i].ID < ns[j].ID
	})
}

// hierarchySlotsFromSchema reads a parent's declared slots out of a parsed
// input_schema (035 D3). Absent or malformed "slots" yields nil — a component
// with no slots block is simply not a composite, which is the safe default and
// the reason composition is opt-in (035 §7 / RFC_022).
func hierarchySlotsFromSchema(schema map[string]interface{}) []hierarchySlotSpec {
	raw, ok := schema["slots"].([]interface{})
	if !ok || len(raw) == 0 {
		return nil
	}
	var out []hierarchySlotSpec
	for _, e := range raw {
		m, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		key, _ := m["key"].(string)
		if key == "" {
			continue
		}
		req, _ := m["required"].(bool)
		out = append(out, hierarchySlotSpec{Key: key, Required: req})
	}
	return out
}

// ---------------------------------------------------------------------------
// Parent/child membership — the ONE spelling on the estate.
//
// Three council seats (reuse_agent, guardian, prior_art_librarian) pressed on
// where this lives, across two rounds. The answer is: here, beside the walk that
// already owns this structure, so a caller never hand-rolls a second test of
// "is this row a composition parent".
//
// NOTE, because round 2 of that review overstated it: there is no PRE-EXISTING
// spelling being retired. `pageComponentHasChildren`, named in round 1 as a
// third spelling, was round 1's own invention and exists nowhere in the tree
// (grep, 2026-08-26). These are the first and only ones.
// ---------------------------------------------------------------------------

// hierarchyDB is the read surface these helpers need. It is satisfied by both
// *sql.DB and *sql.Tx, so a caller that holds a transaction and one that does not
// share one implementation instead of two.
//
// ⚠ CORRECTED 2026-09-03: this comment used to say "apply_section_edit's
// transaction". THERE IS NO SUCH TRANSACTION — that action writes on the
// autocommit connection throughout (component_hierarchy_recompose.go's header
// carries the measurement). The interface is still right, and stays for the
// render paths; the justification was not. Both callers pass *sql.DB today.
type hierarchyDB interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// hierarchyChildrenOf returns a row's children in render order. The
// tombstone filter is datahelpers.NotRemovedSQL — THE shared predicate, not a
// copy of it: a tombstoned row is not on the page, so it must not occupy a slot
// either, and this population must exclude tombstones with the same clause the
// assembler uses or the two drift apart silently. Hand-spelling it here (even in
// the NULL-safe form, which this was) fails
// TestNoHandSpelledTombstonePredicate; the bare constant is right because the
// single-table FROM makes build_status unambiguous.
func hierarchyChildrenOf(ctx context.Context, db hierarchyDB, id uuid.UUID) ([]hierarchyNode, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id::text,
		       COALESCE(parent_instance_id::text, ''),
		       position,
		       COALESCE(slot_name, '')
		  FROM page_components
		 WHERE parent_instance_id = $1
		   AND `+datahelpers.NotRemovedSQL+`
		 ORDER BY position ASC, id ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []hierarchyNode
	for rows.Next() {
		var n hierarchyNode
		if err := rows.Scan(&n.ID, &n.ParentID, &n.Position, &n.SlotName); err != nil {
			// Propagate rather than skip. A child silently dropped here becomes a
			// slot the parent renders empty, which is the whole defect class this
			// feature exists to prevent — and it is the same shape bugs_open/410
			// found in loadStoredSections' own scan loop.
			return nil, fmt.Errorf("component hierarchy: child row scan failed for parent %s: %w", id, err)
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// hierarchyAncestorChain walks parent_instance_id upward from a row and returns
// the chain NEAREST FIRST, so a caller iterating it recomposes bottom-up and
// each level embeds an already-current child.
//
// Bounded by hierarchyMaxDepth and refuses a cycling chain rather than spinning:
// the FK cannot forbid a row pointing at itself or at a descendant, so this is
// the only thing standing between a malformed chain and an infinite loop.
func hierarchyAncestorChain(ctx context.Context, db hierarchyDB, id uuid.UUID) ([]uuid.UUID, error) {
	var chain []uuid.UUID
	cur := id
	seen := map[uuid.UUID]bool{id: true}

	for depth := 0; depth < hierarchyMaxDepth; depth++ {
		var parent sql.NullString
		err := db.QueryRowContext(ctx,
			`SELECT parent_instance_id::text FROM page_components WHERE id = $1`, cur).Scan(&parent)
		if err == sql.ErrNoRows {
			return chain, nil
		}
		if err != nil {
			return nil, err
		}
		if !parent.Valid || parent.String == "" {
			return chain, nil // reached a top-level row: this is the whole chain
		}
		p, perr := uuid.Parse(parent.String)
		if perr != nil {
			return nil, fmt.Errorf("component hierarchy: unparseable parent_instance_id on row %s: %w", cur, perr)
		}
		if seen[p] {
			return nil, fmt.Errorf("component hierarchy: parent chain cycles at %s", p)
		}
		seen[p] = true
		chain = append(chain, p)
		cur = p
	}
	return nil, fmt.Errorf("component hierarchy: parent chain from %s is deeper than the cap of %d",
		id, hierarchyMaxDepth)
}
