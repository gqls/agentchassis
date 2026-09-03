// FILE: platform/orchestration/actions/component_hierarchy_recompose.go
//
// Direction 2 of features_open/035 P1: after a CHILD row is edited, the parent
// that embeds it is stale, and the parent is what the page serves.
//
// WHY THE WHOLE CHAIN, NOT THE IMMEDIATE PARENT. The walk permits depth 3
// (hierarchyMaxDepth), so with nesting the row the page actually serves is the
// TOPMOST ancestor. Recomposing one level up leaves every higher ancestor
// embedding stale bytes — which is bugs_open/384's exact shape one level
// removed, inside the fix for that shape. Caught by the council's bug_historian
// seat on round 1 of correlation 53d71504 and it was the best catch of the round.
//
// WHY IT CANNOT RETURN AN ERROR. It runs at the tail of apply_section_edit,
// after the child write. If it could fail, a child edit that succeeds today would
// begin to fail whenever an ancestor's component or template would not resolve — a
// fleet-wide change to the failure semantics of the live edit path (council
// guardian seat, round 1). Instead an ancestor it cannot refresh keeps its stored
// bytes and its slot name is RETURNED, so the caller can file durable work rather
// than log into a stream that rotates within minutes (council bug_historian, round 2).
//
// WHY IT DOES NOT CALL walkComponentHierarchy. Round 2 of that review claimed it
// would; that claim is retracted and would have been wrong. The walk RE-RENDERS
// children, and an edit path must embed each child's ALREADY STORED bytes —
// regenerating siblings nobody asked to change is its own defect. What is shared
// is the structure vocabulary (hierarchyChildrenOf, hierarchyAncestorChain,
// hierarchyChildKey), so there is still exactly one spelling of membership.
//
// ~~THE db/tx SPLIT IS FORCED, not stylistic.~~
// **CORRECTED 2026-09-03 by wiring it — THERE IS NO TRANSACTION, and the
// parameter that assumed one is why this file had no caller.** The paragraph
// that stood here reasoned about which reads must see "the uncommitted edit"
// inside apply_section_edit's transaction. `[MEASURED 2026-09-03]`
// `grep -nE 'BeginTx|\.Begin\(|Commit\(\)|Rollback\(\)' section_editor_actions.go`
// returns NOTHING: that action persists through
// `updatePageComponentAfterEdit(ctx, params.DB, …)` on the autocommit
// connection, like every other write on the path. So the `tx *sql.Tx` parameter
// could not be supplied by the one caller it was designed for — the function
// compiled, shipped and was eliminated by the linker for want of a caller
// (`editorial_design_uplift/HANDOFF_2026-09-02` §9 recorded the absence and
// correctly read it as "not reachable", not "did not ship").
//
// Every read now goes through *sql.DB, and the guarantee the tx was there to
// provide arrives for free: by the time this runs the child's UPDATE has
// COMMITTED, so an ordinary read sees it. The "cannot return an error" contract
// below is strengthened rather than weakened by that — the child edit is already
// durable, so a refused ancestor costs staleness alone.
//
// The trap worth inheriting: **a comment can assert a transaction the caller
// does not hold, and nothing type-checks a comment.** What caught this was
// attempting the wiring; three council rounds on the design did not.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// hierarchyAncestorRow is one ancestor as recompose needs to see it.
type hierarchyAncestorRow struct {
	id          uuid.UUID
	pageID      string
	componentID string
	slotName    string
	position    int
	contentData map[string]interface{}
	// renderedHTML is the ancestor's CURRENT bytes — the baseline the shrink
	// floor compares against. Without it a recompose is an unguarded rewrite of
	// page_components.rendered_html, which page_component_writer_coverage_test.go
	// refuses (bugs_open/253: both floors guarded 1 of 9 writers until 2026-08-13).
	renderedHTML string
}

// loadHierarchyAncestorRow reads one page_components row through the caller's
// surface. It stays on the hierarchyDB interface rather than *sql.DB so the
// render paths can reuse it from inside a transaction they open themselves —
// apply_section_edit, its first caller, holds none (see the file header).
func loadHierarchyAncestorRow(ctx context.Context, db hierarchyDB, id uuid.UUID) (*hierarchyAncestorRow, error) {
	r := &hierarchyAncestorRow{id: id}
	var cd []byte
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(page_id::text, ''),
		       COALESCE(component_id::text, ''),
		       COALESCE(slot_name, ''),
		       position,
		       content_data,
		       COALESCE(rendered_html, '')
		  FROM page_components
		 WHERE id = $1
	`, id).Scan(&r.pageID, &r.componentID, &r.slotName, &r.position, &cd, &r.renderedHTML)
	if err != nil {
		return nil, err
	}
	if len(cd) > 0 {
		_ = json.Unmarshal(cd, &r.contentData)
	}
	if r.contentData == nil {
		r.contentData = map[string]interface{}{}
	}
	return r, nil
}

// loadHierarchyPageInfo reads the page fields buildRenderContextFromDB needs.
func loadHierarchyPageInfo(ctx context.Context, db *sql.DB, pageID uuid.UUID, siteID uuid.UUID) (*PageInfo, error) {
	pi := &PageInfo{ID: pageID, SiteID: siteID}
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(p.name,''), COALESCE(p.title,''), COALESCE(p.filename,''),
		       COALESCE(p.meta_desc,''), COALESCE(p.url,''), COALESCE(s.domain,'')
		  FROM pages p JOIN sites s ON s.id = p.site_id
		 WHERE p.id = $1
	`, pageID).Scan(&pi.Name, &pi.Title, &pi.Filename, &pi.MetaDesc, &pi.URL, &pi.Domain)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

// recomposeAncestors re-renders every ancestor of a just-edited child, NEAREST
// FIRST so each level embeds an already-current child, and returns the slot
// names it could NOT refresh. It never returns an error — see the file header.
func recomposeAncestors(ctx context.Context, params ActionParams, db *sql.DB,
	childID, siteID uuid.UUID, logger *zap.Logger) (staleSlots []string) {

	chain, err := hierarchyAncestorChain(ctx, db, childID)
	if err != nil {
		// A cycling or over-deep chain is malformed data, not an edit failure.
		logger.Warn("recomposeAncestors: ancestor chain unreadable, leaving ancestors untouched",
			zap.String("child", childID.String()), zap.Error(err))
		return nil
	}

	for _, ancestorID := range chain {
		row, err := loadHierarchyAncestorRow(ctx, db, ancestorID)
		if err != nil {
			logger.Warn("recomposeAncestors: ancestor row unreadable",
				zap.String("ancestor", ancestorID.String()), zap.Error(err))
			staleSlots = append(staleSlots, ancestorID.String())
			continue
		}
		if !recomposeOneAncestor(ctx, params, db, row, siteID, logger) {
			staleSlots = append(staleSlots, row.slotName)
		}
	}
	return staleSlots
}

// recomposeOneAncestor rebuilds a single ancestor from its own template plus its
// children's STORED html. Returns false if it could not, in which case the row
// keeps the bytes it already had.
func recomposeOneAncestor(ctx context.Context, params ActionParams, db *sql.DB,
	row *hierarchyAncestorRow, siteID uuid.UUID, logger *zap.Logger) bool {

	componentID, err := uuid.Parse(row.componentID)
	if err != nil {
		return false
	}
	comp, err := GetComponentByID(ctx, db, componentID, logger)
	if err != nil || comp == nil || comp.HTMLTemplate == "" {
		return false
	}

	kids, err := hierarchyChildrenOf(ctx, db, row.id)
	if err != nil {
		// hierarchyChildrenOf propagates a scan failure rather than returning a
		// short list, so this is a real read failure and not a missing child.
		// Refusing here keeps the parent's good stored bytes.
		return false
	}

	slots := make(map[string]interface{}, len(kids))
	for _, k := range kids {
		key, keyErr := hierarchyChildKey(k.SlotName)
		if keyErr != nil {
			return false
		}
		kidID, idErr := uuid.Parse(k.ID)
		if idErr != nil {
			return false
		}
		var storedHTML string
		// An ORDINARY read, and it is enough: the child's UPDATE has already
		// committed on the autocommit connection before this runs (file header),
		// so there is no in-flight write for it to miss. This is the read the
		// removed tx parameter existed to serve.
		if err := db.QueryRowContext(ctx,
			`SELECT COALESCE(rendered_html,'') FROM page_components WHERE id = $1`, kidID).Scan(&storedHTML); err != nil {
			return false
		}
		if prev, dup := slots[key]; dup {
			slots[key] = prev.(string) + storedHTML // repeated key concatenates in position order
		} else {
			slots[key] = storedHTML
		}
	}

	pageID, err := uuid.Parse(row.pageID)
	if err != nil {
		return false
	}
	pageInfo, err := loadHierarchyPageInfo(ctx, db, pageID, siteID)
	if err != nil {
		return false
	}

	// FRESH context for this ancestor — never shared across nodes. RenderTemplate
	// mutates the provenance digest onto the context and its empty-template branch
	// returns WITHOUT setting it, so a reused context hands one node another's
	// stamp (035 §6.9, RFC_046).
	rc, err := buildRenderContextFromDB(ctx, db, siteID, pageInfo, row.contentData, logger)
	if err != nil {
		return false
	}
	if rc.ContentData == nil {
		rc.ContentData = map[string]interface{}{}
	}
	rc.ContentData["slots"] = slots
	rc.InputSchema = comp.InputSchema

	DeriveAndBindInstanceToken(ctx, db, rc, comp.Function, PlacementFromStoredRow(map[string]interface{}{
		"page_id":  row.pageID,
		"position": row.position,
		"id":       row.id.String(),
	}), logger)

	rendered, _, _, err := RenderTemplate(comp.HTMLTemplate, rc, logger)
	if err != nil || rendered == "" {
		// Never replace good stored bytes with a failed or empty render.
		return false
	}

	// DEAD-INTERNAL-LINK REPAIR, before the floors measure anything. 079's repair
	// guards the full-page section save; a sibling writer that persists
	// rendered_html on its own bypasses it, and an invented /pricing ships as a 404
	// with a green status (bugs_open/136). It is NOT exempt here: this writer
	// re-renders the ANCESTOR's own template against its content_data, so it can
	// mint links of its own — the children's HTML arrives already repaired from
	// whichever writer stored it, but the parent's chrome around them does not.
	// Fail-open by construction, and placed BEFORE the floors so what is measured
	// is what will actually be persisted.
	rendered = repairComponentHTMLBeforePersist(ctx, params, siteID,
		pageInfo.Domain, pageInfo.Name, pageInfo.URL, "recompose_ancestors", rendered, logger)

	// THE SHRINK/COMPONENT FLOORS. This is a writer of page_components.rendered_html
	// and the estate refuses an unguarded one (page_component_writer_coverage_test.go,
	// bugs_open/253). It is deliberately NOT declared exempt: the failure this
	// recompose can produce is a parent that lost its children, which is precisely
	// what a shrink floor is for.
	//
	// A BREACH REFUSES THIS ANCESTOR, it does not error. That matters because
	// enforceSingleSlotFloors' own header warns that a SECOND caller "inherits the
	// visible axis with no calibration of its own" — and I cannot calibrate against
	// composed parents, because there are none yet (0 of 2,249 rows carry a
	// parent_instance_id, 2026-08-31). Refusing rather than failing bounds that
	// uncalibrated risk: a mis-set floor costs a stale ancestor plus the work item
	// the caller files, never a broken child edit. When composed pages exist, the
	// calibration harness (shrink_axis_calibration_test.go) should be re-run over
	// them and this comment revisited.
	if err := enforceSingleSlotFloors(ctx, params, siteID, pageID, pageInfo.Name,
		row.slotName, row.renderedHTML, rendered); err != nil {
		logger.Warn("recomposeAncestors: recomposed ancestor breached a content floor, leaving stored bytes",
			zap.String("ancestor", row.id.String()), zap.String("slot", row.slotName), zap.Error(err))
		return false
	}

	// The guarded, stamped write. `written == false` is a REFUSAL by one of its
	// two predicates (locked or tombstoned ancestor), not a failure — see
	// writeRecomposedAncestor. Either way this ancestor is reported STALE, exactly
	// as a floor breach is.
	written, err := writeRecomposedAncestor(ctx, db, row.id, rendered)
	if err != nil {
		logger.Warn("recomposeAncestors: ancestor write failed",
			zap.String("ancestor", row.id.String()), zap.Error(err))
		return false
	}
	if !written {
		logger.Warn("recomposeAncestors: ancestor is locked or tombstoned — leaving its stored bytes, which now embed a stale child",
			zap.String("ancestor", row.id.String()),
			zap.String("slot", row.slotName))
		return false
	}
	return true
}

// writeRecomposedAncestor persists one recomposed ancestor's bytes and reports
// whether the row actually took them. It is a FUNCTION rather than an inline
// statement for the same reason updatePageComponentAfterEdit and
// updatePageComponentSwap are: that is the seam the section editor's guard tests
// drive directly (section_editor_tombstone_guard_test.go), and a write with no
// such seam can only be tested by mocking the whole render chain in front of it —
// which is how a guard ends up asserted in prose instead of in a test.
//
// FALSE, NIL means the row REFUSED the write, not that nothing happened. The two
// predicates below are the refusers, and neither is decoration: a recompose is a
// MECHANICAL consequence of a child edit, so it arrives at rows nobody chose.
// ApplySectionEditAction checks the lock (bugs_open/058) and the tombstone
// (bugs_open/360) on the row it was ASKED to edit and checks neither on that
// row's ANCESTORS — so without these clauses the one code path in the estate that
// can rewrite a human-locked section is the one path no human pointed at it.
// (Nothing existing would have caught that: TestNoHandSpelledTombstonePredicate
// catches a WRONG spelling of the predicate, never a MISSING one.)
//
// AgentWritableSQLFor / NotRemovedSQL are the shared predicates, not copies —
// hand-spelling the tombstone one fails that test, and hand-spelling the lock one
// is how a lock stops meaning the same thing in two places.
//
// ZERO ROWS AFFECTED MUST BE READ AS A REFUSAL. The statement succeeds either
// way, so an unchecked result would report a locked ancestor as recomposed: a
// green result over stale bytes, which is this feature's own defect class arriving
// inside its fix.
//
// STAMPED (bugs_open/355 A1): the 357 trigger archives whatever this replaces, and
// application_name is what lets the archive row name this writer instead of the
// connection's socket. A new writer of an archived column that does not stamp
// re-opens the attribution hole A1 closed.
//
// The digest is set in the SAME statement as the bytes (bugs_open/229) so the two
// cannot drift — the property that makes staleness checkable as a pure join.
func writeRecomposedAncestor(ctx context.Context, db *sql.DB, id uuid.UUID, html string) (bool, error) {
	res, err := stampedExecContext(ctx, db, contentWriterRecomposeAncestors, `
		UPDATE page_components
		   SET rendered_html = $2, rendered_html_digest = md5($2), updated_at = NOW()
		 WHERE id = $1
		   AND `+datahelpers.NotRemovedSQL+`
		   AND `+pageComponentAgentWritableSQL("")+`
	`, id, html)
	if err != nil {
		return false, err
	}
	n, rowsErr := res.RowsAffected()
	if rowsErr != nil {
		// FAIL CLOSED. ~~Treat it as written~~ — **corrected 2026-09-03 by the
		// council's debug_historian seat (corr cab931b1, medium), and the seat was
		// right.** The first version returned true here, reasoning that the
		// statement had succeeded and a false "stale" would send the caller filing
		// work about a healthy row. But this function's entire purpose is that a
		// write which may not have landed must not read as one; an "assume written"
		// fallback re-opens the exact door the predicates and the row count were
		// added to close, on the one path where the driver has told us it cannot
		// say. The asymmetry decides it: a false STALE costs a log line and a key
		// nobody reads; a false WRITTEN costs a page serving a parent that embeds
		// the pre-edit child, indefinitely and silently.
		return false, nil
	}
	return n > 0, nil
}
