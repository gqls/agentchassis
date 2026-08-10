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
- **SUBSTRATE EVIDENCE, measured 2026-08-09 (argues for the dedicated table):**
  `categories ? 'decision'` was ALREADY in use by three rows from other lanes,
  meaning "a note about a decision", not "an enforceable record". They are
  inert here (site_id NULL, no fences, and both readers require both), but the
  category this RFC declared as "the stable interface" is not exclusively
  ours. Mitigated in data today by adding a `decision-record` category to the
  four RFC_015 rows; **owed at the next roll: tighten both readers from
  `'decision'` to `'decision-record'`.** If the owner prefers, a dedicated
  `decisions` table removes the ambiguity permanently — this is the concrete
  cost of the doc_notes substrate, now paid once and visible.
- Approve the citation-gate semantics (refuse-unless-named) for
  decision-covered subjects, replacing locks for content protection.
- Rule on supersession authority: may any agent supersede any decision by
  naming it, or do some decisions (owner-stated ones, like this batch's)
  require `owner` source to supersede? (Recommended: a `authority` field on
  the decision — `session` supersedable by naming; `owner` requires an
  owner-sourced item — mirrors the existing owner_approval pattern on
  capability_gap items.)

## 5b. OWNER DECISION NEEDED — the second seam, gated twice by one seat while another approved

**Two council rounds on corr `cb547e0a` (2026-08-10) both returned REVISE, both
decided by the same `bug_historian` objection, at HIGH severity:** the
decision-protection mechanism guards ONE write seam (`apply_section_edit`) while
`save_page_sections` and the rerender paths — the bulk-rebuild mechanisms most
likely to discard protected content — stay generic and ungated. It cites 016b §9's
recurring shape: *one call site of a shared judgement gets the rigorous fix; the
sibling stays heuristic.*

**The objection is accepted, not answered.** And this RFC's own evidence is the
strongest case for it: on 2026-08-10 a `section_data_resolved` rerender resurrected
owner-removed content on the very site this RFC protects, and the mechanism caught
it **~7 hours later by async detection rather than preventing it at write time**.
Detective-only coverage with hours of latency IS a materially weaker guarantee
than the citation gate, exactly as the seat says.

**Why this goes to the owner rather than to a third round.** CLAUDE.md: *"A veto on
SCOPE is not answered by resubmitting with better measurements… Record it where the
change lives, route the seam to architecture review on its own merits, and let a
human break it — especially when seats disagree with each other."* They do disagree:
the `architecture` seat **approved** round 2 (its only objection was the unproven
positive path), while `bug_historian` gated on scope. Resubmitting a third time with
the same deferral would be re-litigating a judgement, not supplying evidence.

**The three options, costed:**

1. **Ship the second seam as a PRESERVE-AND-FILE gate (recommended).** At
   `save_page_sections`, a decision-covered slot arriving without a citation keeps
   its EXISTING row and the divergence is filed — it does not fail the rebuild.
   This is not a new mechanism: it is exactly what the lock gate already does in
   that function (`matchLockedRow`, line 802, keeps locked rows out of the DELETE,
   repositions them, discards the fresh copy, emits `lock_blocked_change`). So the
   cost is low and the pattern is proven in place. Per the owner ruling of
   2026-08-02 §2 it must ship as an **opt-in field with the unsafe default OFF**,
   not as a documented convention.
2. **Keep deferring until `page-content-writer` is steered.** The stated reason is
   real — the writer has no way to cite a decision today, so gating its rebuilds
   would block legitimate work — but the 08-10 regression shows what the gap costs
   while we wait.
3. **Accept detective-only coverage for the rebuild path permanently**, on the
   grounds that the guard plus the now-fixed `build_status` filters make
   resurrection much less likely. Weakest option; recorded because the seat's
   objection would then need answering as a decision rather than a debt.

**What has changed since the objection was written**, and it narrows the hole
without closing it: the specific resurrection route is fixed in BOTH readers
(`loadStoredSections` and `getPageSections`, commits `1c7c7c261` and `fba05b83a`),
and the `item_key` fix means a second page's regression under one decision can no
longer be swallowed by dedup. Neither is the seam gate.

## 5c. What the two rounds fixed, and the one piece of evidence still owed

Fixed in response to round 1 (commit `d644723b8`): the category filter tightened
from the ambiguous `'decision'` to `'decision-record'` in both readers (four seats
were right that deferring it bought nothing, since the code cannot run until a roll
either way); `item_key` now includes the page; and the `ItemVerifier` this lane owed
— which turned out to be a **three**-part obligation, the third part
(`claimed-item-timeout` exclusion, migration **374**) surfaced by the build rather
than by any seat.

**Two objections were checked and REJECTED with evidence:** `editquality`'s "no lock
gate exists in `section_editor_actions.go`" (`CheckComponentLock` is at line 305,
`emitLockBlockedChangeItem` at 309, the citation gate at 335 — re-verified after the
file moved by 9 lines; the seat had grepped `matchLockedRow`, a different spelling of
the same concept), and five seats' doubt about the migration ledger (354/355 applied
by `run-migrations.sh` at 14:50 on 08-09, every other migration in that window
`record-only`, so nothing else was swept in).

**Round 2's own failure was a SUBMISSION failure, not a code one:** three seats
(`editquality`, `tooling_provenance`, `prior_art_librarian`) objected that no edit
to the verifier registry was shown. The edits exist — `verifier_coverage_test.go`
and `220`'s declared list — but the submission's edit list omitted them, so a
reviewer reading only the plan correctly concluded the obligation was half-done.
**A reviewer can only object to what the submission shows.**

**RESOLVED 2026-08-10, and it was the last piece of evidence owed:** the
cited-edit control whose value DIFFERS from the stored one. Five seats named this
independently, and the `architecture` seat set the right condition — run it before
any second site is given decision rows.

`rfc015-gate-control-differing-20260810` set `brief-explanation.eyebrow` from
"How it works" to "How this works", citing D-001. Result, checked at each layer
rather than at the item status:

| layer | evidence |
|---|---|
| the gate | `success:true`, **no `skipped`**, **no `decision_gated`** — the citation was accepted |
| the store | `content_data->>'eyebrow'` = "How this works" |
| the artefact | vm-sites commit `bc1676204`, **1 addition / 1 deletion** — non-empty, unlike the 08-09 control — and the diff is exactly the eyebrow `<span>` |

So a cited write **does** reach the artefact, and the earlier ambiguity (a dropped
`field_updates` looks identical to a value-already-equal) is resolved in the right
direction: the gate passes the inputs through untouched. `rfc015-gate-control-restore-20260810`
then restored the original wording, also cited, exercising the path a second time.

**Both directions of the citation gate are now proven at the artefact:** an
uncited edit is refused with the decision named and terminates cleanly
(`rfc015-gate-proof-2`, `skipped:true`), and a cited edit lands bytes.

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
