package actions

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// Duplicate-section collapse at the save choke point (bugs_open/156).
//
// On 2026-07-28 vonc.com/about was saved with 12 page_components rows that were
// 6 byte-identical ADJACENT pairs — identical on rendered_html, on content_data
// and on component_id. The whole page read twice over, live, for two days, and
// was found only because a human ran a census by hand. Every persisted source of
// truth said 6: site_plan_sections said 6, pages.sections said 6, and there was
// one pages row. The doubling existed only in page_components, and the producer
// is [UNRECOVERABLE] — the 12-entry list lived in the run's collected_data,
// which database-cleanup prunes at ~24h, and the orchestration rows had aged out
// before anyone looked.
//
// SavePageSectionsAction has seven guards and record-writers ahead of its
// INSERT. Every one of them compares the incoming set against EXISTING rows, or
// against a floor. None of them compares the incoming set against ITSELF, so a
// list where every planned section appears twice passes all of them — and worse,
// makes four of them measure a number that is not true (see the placement note
// at the call site).
//
// WHAT THIS IS NOT: A UNIQUE INDEX ON (page_id, slot_name)
// ---------------------------------------------------------------------------
// That is the obvious fix and it is wrong. Fleet census 2026-08-04: 12 duplicate
// (page_id, slot_name) groups exist and 11 of them are LEGITIMATE — a page
// repeating generic-text-block 2–3× with different content (ai-agent-
// orchestration, leopardess, gaswholesalers, finetuning, idea.uk), info-card-grid
// twice on webdesign.co.uk/index. Forbidding a repeated slot name breaks all of
// them. THE DISCRIMINATOR IS CONTENT IDENTITY, NOT SLOT REPETITION.
//
// AND IT IS NOT THE BUG FILE'S OWN PROPOSED KEY EITHER
// ---------------------------------------------------------------------------
// bugs_open/156 candidate 1 proposes (slot_name, md5(content_data)). Its own
// census footnote, one paragraph away, records that finetuning.uk/
// our-position-on-ai has two rows whose content_data is NULL on BOTH. Under that
// key every NULL-content pair is "identical" whatever their HTML says, so the
// literal recommendation would delete a live section on a shape the same file
// flags as a trap. Same-content/different-markup fails the same way, and
// rendered_html is what actually serves.
//
// THE RULE: INDISTINGUISHABLE UNDER THIS ACTION'S OWN INSERT
// ---------------------------------------------------------------------------
// An entry is collapsed into an earlier kept entry only when every value the
// INSERT at save_page_sections_action.go would bind is equal, position excluded:
// slot_name, rendered_html, component_id and content_data — each carrying the
// insert loop's OWN normalisations (nil and empty content_data both bind SQL
// NULL; an unparseable component_id binds NULL). content_brief is a pure
// function of page purpose and slot name, so slot equality subsumes it;
// build_status is the constant 'deployed'; position is renumbered by the loop
// from the kept order regardless.
//
// So the collapsed row would have been indistinguishable from its survivor in
// the database: no query, render, edit, lock or history snapshot could have told
// them apart. Nothing representable is lost — which is what makes a guard that
// DELETES BEFORE PERSIST safe enough to run unattended on every save in the
// fleet. It still catches the whole recorded incident: vonc's six pairs matched
// on all four values.
//
// WHY THIS DOES NOT CALL datahelpers.SectionIdentityKey, AND WHY THAT IS NOT A FORK
// ---------------------------------------------------------------------------
// SectionIdentityKey is the shared definition used by the post-hoc detector
// (check_content_duplication) and its repair (remove_duplicate_page_sections).
// Its contract requires content_data::text READ FROM THE JSONB COLUMN — Postgres
// renders jsonb canonically, which is what makes a byte comparison a true
// equality test — and it says in terms that a Go-remarshalled blob does not meet
// that precondition. This guard runs PRE-persist, where content_data is still a
// Go map. Calling it here would violate the stated precondition and quietly
// widen a contract whose scope boundary an architecture seat wrote into that
// comment in council round 2.
//
// The two keys are related by a proof rather than by a convention, which is what
// stops them drifting: encoding/json sorts map keys, so equal Marshal output
// implies identical documents, which implies identical jsonb text once persisted.
// EVERYTHING THIS GUARD COLLAPSES, THE DETECTOR WOULD ALSO HAVE FLAGGED. This
// guard can only under-collapse relative to the detector; it can never
// over-collapse. They answer different questions — the detector asks "is this the
// same section rendered twice?" (minimal sufficient, post-hoc, on stored rows);
// this asks "would this row be indistinguishable from one already being written?"
// (maximally conservative, anchored to one INSERT's bind list), which is why it
// lives beside that INSERT and not in datahelpers.
//
// One deliberate divergence: the detector and repair apply an 80-char normalised-
// prose floor, because their key is (slot, blob) alone and a tiny boilerplate blob
// could byte-match while the markup differs. This key contains rendered_html, so
// the floor is unnecessary — a fully byte-identical row is indistinguishable at
// any size.
//
// PLAN PARITY, AND THE OPPOSITE FAILURE DIRECTION
// ---------------------------------------------------------------------------
// The repair refuses to delete a repetition the EFFECTIVE plan source specifies
// (council trail da3f2d9b, bug_historian seat; owner decision 2026-07-31). A
// save-time collapse that ignored the plan would make the two halves disagree
// about the same question, on the same table — the drift class this council
// reviews for. So this consults the same datahelpers.PlanSpecifiedSectionCounts
// with the same per-slot accounting, and never takes a slot below its planned
// count. It is called LAZILY, only once a duplicate group has been found, so the
// normal path costs no query.
//
// The failure direction is inverted on purpose. The repair FAILS CLOSED — an
// unreadable plan store aborts it, because it is about to DELETE stored rows.
// The conservative direction for a collapse guard is NOT COLLAPSING: on any plan
// read error it returns the incoming set untouched. Both mean "do nothing
// destructive"; they differ because the destructive act differs. Refusing the
// whole save on a plan read error would add a new failure mode to the fleet's
// busiest save path, which is bugs_closed/073's defect, not a fix for it.
//
// IT RECORDS, IT NEVER REFUSES
// ---------------------------------------------------------------------------
// The collapse is lossless by construction, so there is nothing for a human to
// adjudicate before the page ships, and the producer is still unknown, so a
// refusal would break builds for a cause nobody can act on yet. What it does
// instead is leave evidence that outlives collected_data's ~24h retention: which
// extraction path built the list, the metadata field and its origin, the step,
// the driving work item, and the ADJACENCY SIGNATURE. That last one is the
// forensic clue bugs_open/156 turned on — 1,1,2,2,3,3 (each iteration emitted
// its section twice) versus 1,2,3,1,2,3 (the whole loop ran twice) is what ruled
// out the concurrent-save race, and it is the first thing whoever hunts the
// producer will want.
//
// SCOPE: save_page_sections only. Seven Go call sites INSERT into
// page_components; the other six write a single row each and cannot manufacture
// a doubled LIST. This is not fleet-wide coverage of the table and does not
// claim to be.

const (
	// sectionDedupErrorCode is this guard's own agent_error_log code. One code
	// per question, as the other writers in this family do — a shared code
	// makes "what happened on this page" an unanswerable query.
	sectionDedupErrorCode = "CONTENT_DUPLICATE_SECTIONS_COLLAPSED"

	// dedupIdentitySep joins the parts of the identity key. NUL cannot appear
	// in any of them, so the join is unambiguous — the same discipline, and for
	// the same reason, as datahelpers.SectionIdentityKey.
	dedupIdentitySep = "\x00"

	// dedupNullSentinel stands for a value the INSERT would bind as SQL NULL.
	// It is distinguishable from the empty string so that "no content_data" and
	// "content_data that marshals to nothing" cannot collide.
	dedupNullSentinel = "\x00<null>"
)

// collapsedDupGroup is one set of incoming entries that would have written
// indistinguishable rows. Positions are 1-based positions in the ARRIVING list,
// not database positions — the whole point of the record is to describe the list
// as the producer emitted it.
type collapsedDupGroup struct {
	Slot                    string
	KeptArrivalPosition     int
	RemovedArrivalPositions []int
	RenderedHTMLMD5         string
	ContentDataMD5          string
	ComponentID             string
}

// sectionPersistIdentity returns the identity of the row SavePageSectionsAction
// would INSERT for s, position excluded, with that INSERT's own normalisations
// applied. Two entries with equal identities would write rows nothing can tell
// apart.
//
// SectionData.ComponentName IS the slot_name — it is bound to that column
// directly (`save_page_sections_action.go`, the INSERT: `slot_name` is $4 and
// the argument is `section.ComponentName`). Stated because the two names differ
// and a reader checking this key against the stated rule "(slot_name,
// rendered_html, component_id, content_data)" would otherwise have to go and
// confirm it; council round 1 (edit-quality seat, correlation 1a3f4f27) asked
// exactly that question.
//
// idx is used only to make an unmarshallable content_data unique: if we cannot
// compute the blob we cannot claim two entries are identical, so the guard
// abstains for that entry rather than guessing. (Mirroring the INSERT exactly
// would bind NULL for both and call them identical; abstaining is the safe
// direction and costs only a duplicate row that the post-hoc detector still
// sees.)
func sectionPersistIdentity(s SectionData, idx int) string {
	// content_data: nil and empty both bind SQL NULL in the insert loop.
	contentPart := dedupNullSentinel
	if len(s.ContentData) > 0 {
		blob, err := json.Marshal(s.ContentData)
		if err != nil {
			return fmt.Sprintf("\x00<unmarshallable:%d>", idx)
		}
		contentPart = string(blob)
	}

	// component_id: the insert loop parses it and binds NULL on failure, so an
	// empty string and a malformed one are the same stored value.
	componentPart := dedupNullSentinel
	if s.ComponentID != "" {
		if parsed, err := uuid.Parse(s.ComponentID); err == nil {
			componentPart = parsed.String()
		}
	}

	return strings.Join([]string{
		s.ComponentName,
		s.HTML,
		componentPart,
		contentPart,
	}, dedupIdentitySep)
}

// dedupAdjacencySignature preserves the distinction bugs_open/156's forensics
// turned on. "adjacent" means every removed entry sat immediately after the one
// it duplicates (each loop iteration emitted its section twice); "non_adjacent"
// means none did (the whole loop ran more than once); "mixed" is both.
func dedupAdjacencySignature(groups []collapsedDupGroup) string {
	if len(groups) == 0 {
		return ""
	}
	sawAdjacent, sawGap := false, false
	for _, g := range groups {
		prev := g.KeptArrivalPosition
		for _, r := range g.RemovedArrivalPositions {
			if r == prev+1 {
				sawAdjacent = true
			} else {
				sawGap = true
			}
			prev = r
		}
	}
	switch {
	case sawAdjacent && sawGap:
		return "mixed"
	case sawAdjacent:
		return "adjacent"
	default:
		return "non_adjacent"
	}
}

// collapseDuplicateSections is the pure seam — no database, no logging — where
// the behaviour is tested, following scanSectionClaims and evaluateSectionShrink.
//
// planned is PlanSpecifiedSectionCounts' map (component name -> instances the
// effective plan specifies). A slot is never collapsed below its planned count;
// pass nil to impose no constraint.
//
// Returns the kept set in first-occurrence order, the groups collapsed, and the
// slots left alone because the plan specifies their repetition.
func collapseDuplicateSections(sections []SectionData, planned map[string]int) (
	kept []SectionData, groups []collapsedDupGroup, planProtected []map[string]interface{},
) {
	if len(sections) < 2 {
		return sections, nil, nil
	}

	// Cheap prefilter: a slot appearing once can never be a duplicate, and the
	// overwhelmingly common case is that no slot repeats at all.
	slotCount := map[string]int{}
	for _, s := range sections {
		slotCount[s.ComponentName]++
	}
	repeats := false
	for _, n := range slotCount {
		if n > 1 {
			repeats = true
			break
		}
	}
	if !repeats {
		return sections, nil, nil
	}

	// Group by identity, keeping arrival order.
	type identityGroup struct {
		order   int
		indices []int
	}
	byIdentity := map[string]*identityGroup{}
	var identityOrder []string
	for i, s := range sections {
		k := sectionPersistIdentity(s, i)
		g, seen := byIdentity[k]
		if !seen {
			g = &identityGroup{order: len(identityOrder)}
			byIdentity[k] = g
			identityOrder = append(identityOrder, k)
		}
		g.indices = append(g.indices, i)
	}

	// Per-SLOT accounting, never positional — the same shape the repair uses,
	// and for the same reason: page_components.position does not track plan
	// ordering, so "may this slot drop below N" is the only answerable question.
	removalsInSlot := map[string]int{}
	for _, k := range identityOrder {
		g := byIdentity[k]
		if len(g.indices) > 1 {
			removalsInSlot[sections[g.indices[0]].ComponentName] += len(g.indices) - 1
		}
	}
	if len(removalsInSlot) == 0 {
		return sections, nil, nil
	}

	protectedSlots := map[string]bool{}
	for slot, removals := range removalsInSlot {
		want := planned[slot]
		if slotCount[slot]-removals < want {
			protectedSlots[slot] = true
			planProtected = append(planProtected, map[string]interface{}{
				"slot_name":     slot,
				"planned_count": want,
				"rows_now":      slotCount[slot],
				"would_remove":  removals,
			})
		}
	}
	sort.Slice(planProtected, func(i, j int) bool {
		return planProtected[i]["slot_name"].(string) < planProtected[j]["slot_name"].(string)
	})

	drop := map[int]bool{}
	for _, k := range identityOrder {
		g := byIdentity[k]
		if len(g.indices) < 2 {
			continue
		}
		slot := sections[g.indices[0]].ComponentName
		if protectedSlots[slot] {
			continue
		}
		grp := collapsedDupGroup{
			Slot:                slot,
			KeptArrivalPosition: g.indices[0] + 1,
			RenderedHTMLMD5:     shortMD5(sections[g.indices[0]].HTML),
			ComponentID:         sections[g.indices[0]].ComponentID,
		}
		if len(sections[g.indices[0]].ContentData) > 0 {
			if blob, err := json.Marshal(sections[g.indices[0]].ContentData); err == nil {
				grp.ContentDataMD5 = shortMD5(string(blob))
			}
		}
		for _, idx := range g.indices[1:] {
			drop[idx] = true
			grp.RemovedArrivalPositions = append(grp.RemovedArrivalPositions, idx+1)
		}
		groups = append(groups, grp)
	}

	if len(drop) == 0 {
		return sections, nil, planProtected
	}

	kept = make([]SectionData, 0, len(sections)-len(drop))
	for i, s := range sections {
		if drop[i] {
			continue
		}
		kept = append(kept, s)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].KeptArrivalPosition < groups[j].KeptArrivalPosition
	})
	return kept, groups, planProtected
}

func shortMD5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])[:8]
}

// dedupSectionsBeforePersist is the DB-facing seam: it consults the plan store
// only when a duplicate has actually been found, collapses, logs loudly and
// writes the durable record. It NEVER refuses a save and never returns an
// error — on any measurement failure the incoming set is returned untouched.
//
// Returns the kept set and how many entries were collapsed.
func dedupSectionsBeforePersist(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	domain, pageName, pageURL string,
	sectionsSource, metaField, metaFieldOrigin string,
	sections []SectionData,
) ([]SectionData, int) {
	logger := params.Logger

	// First pass with no plan constraint, purely to learn whether there is
	// anything to protect. Nothing is applied from this pass.
	_, provisional, _ := collapseDuplicateSections(sections, nil)
	if len(provisional) == 0 {
		return sections, 0
	}

	// A duplicate exists, so the plan read is now worth its cost.
	var planned map[string]int
	var planSource string
	if params.DB != nil {
		p, src, err := datahelpers.PlanSpecifiedSectionCounts(ctx, params.DB, siteID, pageName)
		if err != nil {
			// Conservative direction for a COLLAPSE guard is not collapsing.
			logger.Warn("SavePageSectionsAction: duplicate sections found but the plan store is unreadable — not collapsing (bugs_open/156)",
				zap.String("page_name", pageName),
				zap.Int("duplicate_groups", len(provisional)),
				zap.Error(err))
			return sections, 0
		}
		planned, planSource = p, src
	}

	kept, groups, planProtected := collapseDuplicateSections(sections, planned)
	collapsed := len(sections) - len(kept)

	if len(groups) == 0 && len(planProtected) == 0 {
		return sections, 0
	}

	signature := dedupAdjacencySignature(groups)
	slots := make([]string, 0, len(groups))
	for _, g := range groups {
		slots = append(slots, g.Slot)
	}

	if collapsed > 0 {
		// A distinctive compiled string: this is the grep marker for the
		// post-roll pod check, and the phrase a log search will be run on.
		logger.Warn("SavePageSectionsAction: DUPLICATE SECTIONS COLLAPSED — the incoming list carried byte-identical entries (bugs_open/156)",
			zap.String("page_name", pageName),
			zap.Int("arrival_count", len(sections)),
			zap.Int("kept_count", len(kept)),
			zap.Int("collapsed_count", collapsed),
			zap.String("adjacency_signature", signature),
			zap.Strings("slots", slots),
			zap.String("sections_source", sectionsSource),
			zap.String("metadata_field_origin", metaFieldOrigin),
			zap.String("plan_source", planSource))
	}
	for _, p := range planProtected {
		logger.Warn("SavePageSectionsAction: byte-identical sections left in place — the plan specifies the repetition (bugs_open/156)",
			zap.String("page_name", pageName),
			zap.String("slot_name", p["slot_name"].(string)),
			zap.Int("planned_count", p["planned_count"].(int)),
			zap.String("plan_source", planSource))
	}

	writeSectionDedupLog(ctx, params, siteID, domain, pageName, pageURL,
		len(sections), len(kept), collapsed, signature, groups, planProtected,
		sectionsSource, metaField, metaFieldOrigin, planSource, logger)

	return kept, collapsed
}

// writeSectionDedupLog persists the finding beyond collected_data's ~24h
// retention. Best-effort: a logging failure must never disturb a save whose
// content is now correct.
//
// The payload is chosen against a specific past failure. When vonc's doubling
// was investigated on 2026-07-30 the producer could not be identified at all,
// because the only trace of the 12-entry list had been pruned. Every field here
// is one the next investigator will have and that one did not.
func writeSectionDedupLog(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	domain, pageName, pageURL string,
	arrival, keptCount, collapsed int,
	signature string,
	groups []collapsedDupGroup,
	planProtected []map[string]interface{},
	sectionsSource, metaField, metaFieldOrigin, planSource string,
	logger *zap.Logger,
) {
	if params.DB == nil {
		return
	}

	// The outcome is stored, not inferred from the counts — a later reader must
	// not have to re-derive "was anything actually removed?" from two integers
	// whose relationship could change.
	outcome := "collapsed"
	if collapsed == 0 {
		outcome = "plan_protected"
	}

	groupMaps := make([]map[string]interface{}, 0, len(groups))
	for _, g := range groups {
		groupMaps = append(groupMaps, map[string]interface{}{
			"slot_name":                 g.Slot,
			"kept_arrival_position":     g.KeptArrivalPosition,
			"removed_arrival_positions": g.RemovedArrivalPositions,
			"occurrences":               len(g.RemovedArrivalPositions) + 1,
			"rendered_html_md5":         g.RenderedHTMLMD5,
			"content_data_md5":          g.ContentDataMD5,
			"component_id":              g.ComponentID,
		})
	}

	// The work item driving this save, when the step declares one — the same
	// three lines the action uses for history attribution. Without it a finding
	// names a page but not the request that rebuilt it.
	workItemID := ""
	if f, ok := params.StepConfig.Config["work_item_id_field"].(string); ok && f != "" {
		workItemID = datahelpers.ExtractNestedFieldString(params.CollectedData, f)
	}

	// Severity encodes the SAVE'S FATE in this family — the claims guard writes
	// 'error' only when it refuses, and the content_data record writes 'warning'
	// because it allows. This guard allows, so: warning. The loudness lives in
	// the pod line above and in the error_code being queryable, not here.
	LogActionEntry(ctx, params, agenterrors.Entry{
		SiteID:    siteID.String(),
		Domain:    domain,
		AgentType: componentRepairAgentType(params),
		StepName:  saveSectionsStepName(params),
		Action:    "save_page_sections",
		ErrorMessage: fmt.Sprintf("Duplicate sections %s on page %s: %d arrived, %d were byte-identical duplicates of earlier entries (%s), %d saved",
			outcome, pageName, arrival, collapsed, signature, keptCount),
		ErrorCode: sectionDedupErrorCode,
		Severity:  "warning",
		Context: map[string]interface{}{
			"outcome":               outcome,
			"page_name":             pageName,
			"page_url":              pageURL,
			"arrival_count":         arrival,
			"kept_count":            keptCount,
			"collapsed_count":       collapsed,
			"adjacency_signature":   signature,
			"collapsed_groups":      groupMaps,
			"plan_protected_slots":  planProtected,
			"plan_source":           planSource,
			"sections_source":       sectionsSource,
			"metadata_field":        metaField,
			"metadata_field_origin": metaFieldOrigin,
			"work_item_id":          workItemID,
			"identity_rule":         "slot_name+rendered_html+component_id+content_data as bound by the INSERT; position excluded",
			"bug":                   "bugs_open/156",
		},
	}, logger)
}
