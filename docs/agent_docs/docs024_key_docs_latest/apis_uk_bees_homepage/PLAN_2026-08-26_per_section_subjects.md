# PLAN 2026-08-26 — per-section subjects (`pages.sections` entries gain a `subject`)

**Owner go-ahead 2026-08-25 ("go ahead with per section subjects"), design approved 08-24
(handoff §5). Council-scope: platform Go + one migration + three seeds.**

## The defect (measured, not re-argued)

Every slot with the same component name gets an identical brief. `site_plan_sections` for
apis.uk today: 6 × `generic-text-block`, nothing distinguishing them; one `content_rewrite`
rewrote all six sections about the waggle dance. Bug family: `bugs_open/151` (writer isolation),
this lane's NOTES (four measurements).

## The contract this extends — NOT a new seam

RFC_016 §5.1, **RATIFIED by the owner 2026-08-08**: the section-entry object form is a
validate_plan-internal transient; every downstream carrier keeps its historical shape plus a
**named, aligned, per-page sibling key**; the NEXT structured per-section field must extend this
instead of inventing a rival carrier. `subject` is that next field. Precedent thread, live end to
end for `facts` (PBP-037, Slice B applied — the v4 prompt's `facts_scoped` branch is in the live
`page-content-writer` config, checked 2026-08-25):

```
planner LLM {"name","facts"}                            → + "subject"
  validate_plan normalise → sections + section_facts    → + section_subjects
    write_site_plan → site_plan_sections.assigned_fact_ids → + .subject (migration 638)
      load_page_sections_from_spec → section_facts aligned  → + section_subjects
        plan_sections (wired 328) → sectionPlanItem.FactsScoped/… → + .Subject
          writer loop current_section → v4 prompt block          → + v5 subject block (seed, HOLD)
```

## Edits

| # | file | change |
|---|---|---|
| 1 | `docs/agent_docs/sql_for_agents/638_site_plan_sections_subject.sql` | `ALTER TABLE site_plan_sections ADD COLUMN subject text` (nullable, additive). **Apply immediately on commit** — the Go loader/writer name the column, so the column must exist before any roll |
| 2 | `write_site_plan_action.go` | `sectionEntry.Subject`; `extractSectionEntries` reads object `subject` + page-level `section_subjects` sibling (raw-index, same rule as facts); INSERT adds `subject` (empty ⇒ NULL) |
| 3 | `v3_site_actions.go` | normalise pass emits `section_subjects` aligned; `carrySectionFactsOntoRealised` carries subject too (subject-only object entries enqueue; "facts" key never fabricated) |
| 4 | `load_page_sections_from_spec_action.go` | SELECT + scan `subject`; aligned `section_subjects` out (authoritative tier only); locked-merge nil-insertion mirrored |
| 5 | `plan_sections_action.go` | Optional input `section_subjects`; triple-aligned parse + site-level filter; `sectionPlanItem.Subject` attached |
| 6 | tests | alignment + attach + carry + loader subject alignment; mutation-style where cheap |
| 7 | seed `639_…_HOLD` | page-build-handler `plan_sections` config `+ "section_subjects": "spec_sections.section_subjects"` (mirror of 328) |
| 8 | seeds `640_…_HOLD` (planner rule 17 amendment: `subject` allowed always, REQUIRED when a component repeats on a page; object-without-facts allowed when no verified facts) · `641_…_HOLD` (writer prompt v5 subject block — **⚠ NEEDS A FRESH OWNER READ**: the v4 approval "attaches to the committed text … any later edit voids the approval", RFC_016 §5.2) |

**Inertness/order:** 638 applied now (safe under every live binary — no `SELECT *` scans of the
table in Go, checked); Go inert until an image rolls; 639–641 are `_HOLD` (the `SIDECAR_RE`
convention) and apply **after** the roll, image-first per 328's own note; 641 additionally after
the owner reads the v5 text.

**Falsifier:** a plan with no subjects builds byte-identically (all-nil sibling, `omitempty`
fields absent); a plan whose entries carry subjects yields `sections_ready` items whose
`subject` differs per slot; the v5 prompt renders the block only when `current_section.subject`
is non-empty.

**Consumers to tell (owner ruling 2026-07-29 §3):** `brochure_component_library` (RFC_016/151 —
their seam extended per its own §5.1), `bugfix_285` lane (loader alignment code touched, LOCK-008),
`copy_quality_two_stage` (they asked to be told when the plan starts assigning subjects — their
form-vs-phrase experiment preconditions on it).

**Un-defer condition for the two apis.uk `content_rewrite` items:** the whole chain live
(image rolled + 639/640/641 applied) — then a replan or plan-row `subject` backfill on apis.uk
gives the two new sections distinct subjects. Phase 2, separate decision (the locks refuse
page re-renders by design).
