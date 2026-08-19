# PLAN 2026-08-19 — page_component_history retention (STY-056 open-review (a), now due)

**Decision being taken:** the volume/pruning design the owner's candidate-1 ruling
deferred "until page-side measurements exist". Watch (a) tripped on 2026-08-19:
63MB (was 30MB on 08-10), ~3.5MB/day against the ~0.9MB/day worst-case projection.
Nine days of production measurements now exist; this plan spends them.

**090 substitution, stated per the owner ruling of 2026-07-31:** not through the
diagnosis loop — there is no defect and no symptom; a designed watch tripped its
threshold and the design's own instruction ("decide pruning on page-side
measurements") is being executed. Every figure below is first-hand from the live
DB (queries in NOTES 08-19 entry); the semantic claim the design leans on is the
platform's own council-reviewed contract, cited inline. The council gate reviews
the change itself.

## The measurements (2026-08-19, live DB)

- Trigger-arm rows since 08-10: **5,478** — delete/machine_made **4,085 (75%)**,
  delete/unstamped 1,075, overwrite/machine_made 189, overwrite/unstamped 101,
  delete/hand_patched 22, overwrite/hand_patched 6.
- Bytes: trigger-arm html ≈ 34MB over 9 days; snapshot-source rows ≈ 9MB
  (mostly content_data — their artefact arm is COALESCE-dropped by design).
- So the growth driver is, overwhelmingly, **machine_made payloads on the delete
  arm** — bytes the stamp itself classifies as reproducible, archived on every
  routine DELETE+INSERT save.

## The design

**Null the payload, keep everything else, machine_made-only, 30 days.**
A daily scheduled task (`page-component-history-retention`, the `database-cleanup`
/ pure-`pre_query` shape, `fire_message=false`) runs:

- `UPDATE page_component_history SET rendered_html=NULL WHERE
  source='artefact_archive_trigger' AND divergence='machine_made' AND
  created_at < now()-interval '30 days' AND rendered_html IS NOT NULL`
- writes ONE `doc_notes` row per run — **on zero too** (the WFA-013 precedent: a
  MISSING row means the job did not run and must not read as "nothing to do").
- a partial index (`source/divergence/rendered_html IS NOT NULL` on `created_at`)
  keeps the daily scan off the table's back.

**What is never touched — the preservation contract, in one place:**

| class | policy | why |
|---|---|---|
| `hand_patched` payloads | kept for ever | not reproducible; the whole reason the trigger exists |
| `unstamped` payloads | kept (revisit at a later watch) | provenance unknown = not provably reproducible; the class shrinks naturally as restamp-through-churn proceeds (measured falling 08-09→08-10) |
| `content_data`, all rows | kept for ever | the platform's restore recipes (migs 287/378/431) read it; it is the recovery source |
| the ledger row itself (op/divergence/digest/slot/position/created_at) | kept for ever | the audit trail and the sweep (open-review (e)) key on it |
| snapshot-source rows (`save_page_sections_overwrite`) | untouched | open-review (d)'s question, deliberately not answered here |
| chrome sibling (`site_component_history`) | untouched | 57-row table; no volume problem to solve, and symmetric machinery with no demand is the helper-with-no-callers shape |

**Why machine_made payloads are the droppable class — the platform's own
precedent, not this plan's invention:** the save-path snapshot has ALWAYS dropped
the artefact when `content_data` exists (COALESCE, measured 14,831/14,863 rows at
design time) — that is the accepted policy for machine-reproducible bytes, and
the trigger arm was built to catch the DIVERGENT class, not to reverse that
policy. `divergence='machine_made'` (outgoing md5 == same-statement stamp) is the
council-reviewed marker for "exactly what a stamped writer last wrote".
Known softness, stated: "reproducible from content_data" is the stamp's
semantics, and the LANDMINE "content_data often holds site-wide boilerplate"
means regeneration is via the render pipeline, not byte-exact for LLM-authored
html. The 30-day window and payload-null-not-row-delete are the mitigations: a
month of byte-exact recovery for machine content, indefinite recovery for
everything the platform cannot regenerate.

**The enablement runway is structural, not scheduled:** the trigger has only
existed since 08-09, so the oldest possible trigger-arm row reaches 30 days on
~2026-09-08. Enabling the task TODAY means ~20 daily no-op runs, each writing its
zero-count report, before the first byte is nulled — a free observation window in
which disabling the row (`UPDATE scheduled_tasks SET enabled=false WHERE
name='page-component-history-retention'`, live immediately) costs nothing.

**Steady state:** ~30 days × ~3.5MB/day ≈ 100–120MB of retained recent payloads
plus the permanent divergent/ledger classes — bounded, versus unbounded
(~1.3GB/yr) today. The parameter is a task-row literal the owner can change live.

## Reader census (done BEFORE this design; NOTES 08-19)

Runtime readers of `page_component_history`: save_page_sections_action,
rerender_page_sections_action, content_data_envelope_guard, section_visible_text,
save_sections_prune_floor (cohort calibration), page_component_divergence (ledger
read-back), save_component_history_action, core-manager page_admin_handlers; plus
one-off restore migrations 287/378/431. None requires a machine_made trigger-arm
**payload** older than 30 days: the restores read `content_data` (kept); the
divergence ledger read-back and the sweep read hand_patched rows (kept); admin
history browsing degrades to "payload pruned" on old machine rows (the row still
states op/divergence/digest).

## Rollback

`489_*_ROLLBACK.sql`: delete the task row, drop the index. **Nulled payloads are
not recoverable — stated plainly**; that is why the predicate takes only the
class the platform already declines to keep, and why the first prune is ~20 days
after review.

## Verification (the migration probes itself, 357-style DO/RAISE)

Probe inserts three rows on a real page FK (slot `mig489_probe`): an old
machine_made, an old hand_patched, a recent machine_made; runs the retention
UPDATE verbatim; asserts payload nulled ONLY on the first (row and content_data
surviving), then self-cleans. Post-apply: task row present + enabled; first
scheduled run's doc_notes row within 24h ("0 payloads" expected until ~09-08).
