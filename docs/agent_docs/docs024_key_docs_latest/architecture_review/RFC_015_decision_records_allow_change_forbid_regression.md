# RFC 015 — Decision records: allow change, forbid regression

**Filed 2026-08-06 by the idea_uk_vm_site lane ("ideauk sec" session), at the
owner's direction, after the owner removed all 35 component locks on idea.uk
("I haven't intentionally locked anything") and ruled: "the provenance records
should be what stops the planner from overwriting and regressing decisions —
we want the site to improve, so changes should be allowed, but not regress."**

Status: **APPROVED (owner, 2026-08-08: "please go ahead with this, all the way")** — implemented same day: steer LIVE on webdesign-agent (snapshot 9dc5f47a); guards + citation gate code committed (council corr c2940987), inert until the next chassis roll; migration 340 pending; save_sections seam deferred until page-content-writer is steered. Originally: steer + guards data/config may proceed
under the normal council gate; the citation gate (§4.3) is a shared write-seam
change and needs this RFC decided first (it changes what the seam GUARANTEES,
which is the 2026-07-29 ruling's RFC trigger).

## 1. The problem, from evidence

Three mechanisms currently claim to protect decisions; each fails differently:

- **Locks** (`page_components.locked_at`, lock_type='permanent') stop
  regression by stopping everything. They block improvement, they masquerade
  as human decisions when agents set them (all 35 on idea.uk were
  agent-placed), and they leak noise (11 `lock_blocked_change` items were the
  locks "working"). The owner has ruled the class away for this purpose.
- **Prose provenance** (travelling docs: `doc_plans`/`doc_notes`, convention
  037) stops nothing. Measured 2026-08-05: `visual-designer`,
  `visual-design-auditor`, `webdesign-agent` and `site-design-planner`
  reference neither doc_notes nor travelling docs in their live configs — and
  a note nobody reads is not a control (CLAUDE.md, owner ruling 2026-08-02:
  "a comment is not a control on a tree this many sessions share").
- **site_specs pins** steer generation (the palette pin genuinely works — the
  colour-churn landmine is disarmed by `design_intent.palette.reference_values`
  because it is rendered into the generator's prompt) but `pinned` is inert
  (`grep -c pinned site_spec_actions.go` → 0) and any run may supersede the row.

The platform has already solved this shape once: **bugfix 161** — the concept
register was made BOTH the writer's instruction set AND the gate's authority,
because "correcting the register does NOT arm detection; `banned_claims`
does." That dual-face pattern generalises, and this RFC is that
generalisation.

## 2. The design: one decision record, three faces

A **decision record** is one row (substrate: the existing travelling-docs
tables — `doc_notes` gains nothing new structurally; a `decision` category +
a machine `spec` block) with:

1. **WHY** — prose provenance, exactly what convention 037 already holds.
   Never enforced; for humans and future sessions.
2. **STEER** — a compact machine block injected into the GENERATION CONTEXT
   of the planner and design agents (per-site + per-subject pre_query on
   decision rows, a one-step agent-definition config change — measured
   2026-08-05). This is the palette-pin pattern made general: regression
   becomes unlikely because every run starts from the decision.
3. **GUARD** (where mechanically expressible) — an OUTCOME invariant over the
   served artifact, evaluated by the existing discovery/checker layer, filing
   a `decision_regression` work item naming the decision id on violation.
   `banned_claims` generalised, positive and negative. **The invariant pins
   the WHAT, never the HOW** — the planner stays free to restructure, reword
   and improve so long as the outcome holds. That is the whole
   allow-change/forbid-regression property.

### The citation gate, for decisions no assertion can check

Hand-authored copy quality ("do not regenerate the guides") has no selector
assertion. For these, enforcement moves to the WRITE SEAMS every writer
already passes through (`apply_section_edit`; the planner's persist/emit
step): if the target subject/slot is covered by an active decision, the write
must NAME it — `acknowledges_decision` / `supersedes_decision` fields on the
work item. Named → allowed, logged, auditable. Unnamed → refused with a
pointer to the decision it did not know about.

**You may change anything you can name; you may not change what you did not
know existed.** Accidental regression becomes impossible by construction;
deliberate change costs one visible field. This is the owner ruling of
2026-08-02 §2 applied to content: authority ships as an explicit opt-in
field with the unsafe default OFF, not as a documented convention.

## 3. Worked examples (the first typed decisions, from 2026-08-05's batch)

| decision | steer | guard | gate-only? |
|---|---|---|---|
| free check beside paid report on index, both labelled, either standalone | funnel block in planner context | index HAS link `/tools.html#audience-check` with label containing "free" AND link `/report.html` | no |
| no standalone tools-directory section on index | same block | index has NO nav-like section linking all four tools outside header/footer | no |
| logo reads IDEA, provider banana | `imagery_style_guide.provider` (already live) + brand block | visual-auditor seat: logo legibility (not text-greppable) | partially |
| guide copy hand-authored (p4_01–p4_17 provenance) | guide block: "improve structure freely; copy changes must supersede D-004" | none expressible | **yes — citation gate** |

## 4. Rollout, each stage independently useful

1. **STEER (data/config, now):** type the existing provenance rows as
   decisions; add the pre_query to the four design agents' definitions.
   Normal council gate covers it.
2. **GUARDS (data + check specs, next):** express guards for every decision
   that has one; run inert first (151's `check_content_duplication`
   precedent), then arm. Normal council gate.
3. **CITATION GATE (Go, last):** the seam check + the two work-item fields.
   Architecture-scope BY THIS RFC's own criteria — it changes the write
   seam's guarantee from "any triaged item may write" to "a decision-covered
   slot requires citation". Ships only after this RFC is decided; the code
   goes through the council gate as usual; the seam gets a concept-register
   entry in the same commit (2026-07-28 ruling, condition 2).

## 5. What this RFC asks the owner to decide

- Approve the three-face decision-record shape and the doc_notes substrate
  (vs a new table — doc_notes wins on "the tables the platform actually
  reads", but a `decisions` table is cleaner if the citation gate grows).
  **Substrate data point, hit while staging (2026-08-06):**
  `doc_notes_subject_type_check` constrains subject_type to
  `tool|pipeline|experience|action|experience-pattern|landmine|component` —
  no `decision`. The first four records are staged under `component` with the
  `decision` CATEGORY carrying the semantics (source `rfc015-staging`, keys
  D-001..D-004). Widening the constraint is a stage-1 migration — additive
  and inert until something writes the new type, so per the 2026-07-29 ruling
  it takes the normal council gate, not an RFC of its own.
- Approve the citation-gate semantics (refuse-unless-named) for
  decision-covered subjects, replacing locks for content protection.
- Rule on supersession authority: may any agent supersede any decision by
  naming it, or do some decisions (owner-stated ones, like this batch's)
  require `owner` source to supersede? (Recommended: a `authority` field on
  the decision — `session` supersedable by naming; `owner` requires an
  owner-sourced item — mirrors the existing owner_approval pattern on
  capability_gap items.)

## 6. Known limits, stated

- Guards cover only assertable outcomes; visual/qualitative decisions rely on
  the steer + auditor seats + citation gate.
- The steer is soft (LLM adherence); the guard is the backstop. Neither alone
  suffices — that is why the record has both faces.
- A decision record can itself go stale; supersession-by-citation keeps the
  trail, and the experience loop reads the same rows (category
  `experience-council`) so stale decisions surface in its sweeps.

**Related:** bugfix 161 (register ratifies the claim) · owner ruling
2026-08-02 §2 (opt-in authority fields) · RFC_005 §3.2 (docs that feed the
fleet) · the 2026-08-05 section-resurrection landmine (why "removed" needed
four data changes — a decision record would have carried the recipe) ·
idea_uk_vm_site/RUNNING_NOTES §X.43–X.46.
