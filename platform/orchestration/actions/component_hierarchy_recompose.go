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
// WHY IT CANNOT RETURN AN ERROR. It runs inside apply_section_edit's
// transaction. If it could fail, a child edit that succeeds today would begin to
// fail whenever an ancestor's component or template would not resolve — a
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
// THE db/tx SPLIT IS FORCED, not stylistic. buildRenderContextFromDB and
// DeriveAndBindInstanceToken both take *sql.DB and cannot take a *sql.Tx. So
// edit-SENSITIVE reads (the just-written child's stored HTML, the ancestor row)
// go through tx and see the uncommitted edit; edit-INSENSITIVE reads (site data,
// page record, occurrence counting) go through db. Three rounds of plan review
// never surfaced this; compiling it did.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"
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
// surface (tx, so it sees the in-flight edit).
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
		       COALESCE(p.meta_desc,''), COALESCE(s.domain,'')
		  FROM pages p JOIN sites s ON s.id = p.site_id
		 WHERE p.id = $1
	`, pageID).Scan(&pi.Name, &pi.Title, &pi.Filename, &pi.MetaDesc, &pi.Domain)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

// recomposeAncestors re-renders every ancestor of a just-edited child, NEAREST
// FIRST so each level embeds an already-current child, and returns the slot
// names it could NOT refresh. It never returns an error — see the file header.
func recomposeAncestors(ctx context.Context, params ActionParams, db *sql.DB, tx *sql.Tx,
	childID, siteID uuid.UUID, logger *zap.Logger) (staleSlots []string) {

	chain, err := hierarchyAncestorChain(ctx, tx, childID)
	if err != nil {
		// A cycling or over-deep chain is malformed data, not an edit failure.
		logger.Warn("recomposeAncestors: ancestor chain unreadable, leaving ancestors untouched",
			zap.String("child", childID.String()), zap.Error(err))
		return nil
	}

	for _, ancestorID := range chain {
		row, err := loadHierarchyAncestorRow(ctx, tx, ancestorID)
		if err != nil {
			logger.Warn("recomposeAncestors: ancestor row unreadable",
				zap.String("ancestor", ancestorID.String()), zap.Error(err))
			staleSlots = append(staleSlots, ancestorID.String())
			continue
		}
		if !recomposeOneAncestor(ctx, params, db, tx, row, siteID, logger) {
			staleSlots = append(staleSlots, row.slotName)
		}
	}
	return staleSlots
}

// recomposeOneAncestor rebuilds a single ancestor from its own template plus its
// children's STORED html. Returns false if it could not, in which case the row
// keeps the bytes it already had.
func recomposeOneAncestor(ctx context.Context, params ActionParams, db *sql.DB, tx *sql.Tx,
	row *hierarchyAncestorRow, siteID uuid.UUID, logger *zap.Logger) bool {

	componentID, err := uuid.Parse(row.componentID)
	if err != nil {
		return false
	}
	comp, err := GetComponentByID(ctx, db, componentID, logger)
	if err != nil || comp == nil || comp.HTMLTemplate == "" {
		return false
	}

	kids, err := hierarchyChildrenOf(ctx, tx, row.id)
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
		// tx, deliberately: this must see the child write this transaction just made.
		if err := tx.QueryRowContext(ctx,
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

	// Digest in the SAME statement as the bytes (bugs_open/229) so the two can
	// never drift — the property that makes staleness checkable as a pure join.
	if _, err := tx.ExecContext(ctx, `
		UPDATE page_components
		   SET rendered_html = $2, rendered_html_digest = md5($2), updated_at = NOW()
		 WHERE id = $1
	`, row.id, rendered); err != nil {
		logger.Warn("recomposeAncestors: ancestor write failed",
			zap.String("ancestor", row.id.String()), zap.Error(err))
		return false
	}
	return true
}
