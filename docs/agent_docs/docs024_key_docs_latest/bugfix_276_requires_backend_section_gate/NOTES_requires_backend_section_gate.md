# NOTES — requires-backend section gate (bugs_open/276), append-only, newest at bottom

## 2026-08-15 — picked up

Checked `bugs_open/277` first (session was renamed "bugfix 277" by the user's `/rename`), but
`who-owns.py 277` plus a dirty, 6-minutes-old `RUNBOOK_required_fields_repair.md` in the
owning workstream showed a live session actively mid-task on it right now — not just
historically owned. Declined to compete (per CLAUDE.md's who-owns norm) and surveyed
`bugs_open/` for the next genuinely free bug instead. `bugs_open/275` and `bugs_open/276` were
both explicitly flagged "for pickup (either lane, or a fresh thread)" in the owning
workstream's own 2026-08-15 handoff. Picked **276** — the higher-priority structural gate
(prevents a real broken-page outcome; raised by the council itself), over 275 (a simpler
alphabetical-cap fix).

Read `bugs_open/276` in full, the VMB-010 concept register entry, and the two worked-example
migrations (406, 407). VMB-010 and the bug file both name only `build-site-planner`'s
`load_components` step as the gate target.

**Enumerated every `agent_definitions` step fleet-wide whose `query` references
`content_components`** (14 hits) rather than trusting that single named call site. Classified
all 14: 3 are section-candidate "menus" (the placement-risk shape), the rest are schema-hint
dumps, already-placed-component reads, or single-tool lookups by id — none of those can place
a NEW requires-backend section, so none needed the gate.

The 3 real menus: `content-gap-planner.load_available_components` (NOT named anywhere in the
bug or the register — 131 dispatches/30d, most recent today), `build-site-planner.load_components`
(2 dispatches/30d — the minority path, despite being the one everyone named), `site-planner.
load_available_components` (0 dispatches ever — dead/legacy agent, confirmed with an unbounded
query, not just a 30-day window).

Checked `intent-probe`'s current placements: exactly one, `relojistas.com`/`index`, which
already carries `deploy_config->capabilities ? 'backend'`. **No live damage exists** — this is
a forward-looking gate.

Traced `content-gap-planner`'s workflow graph (`ensure_site_record` → `load_specs` →
`load_existing_pages` → `load_available_components` → `plan_gaps` → `apply_plan`) and
confirmed two sibling steps upstream of the target already bind `site_record.site_id` —
verified live: 131/131 recent orchestration rows carry a non-null
`collected_data#>>'{site_record,site_id}'`. Safe to add `params: ["site_record.site_id"]`.

Traced `site-planner`'s workflow (`load_available_components` → `load_style_collections` →
`plan_site` → `validate_plan` → `complete`) and found **no** `ensure_site_record`-equivalent
step anywhere — no proven site-id binding exists. Decided: gate this step too but
*unconditionally* (strip the tag with no capability check), rather than guess an unproven
params path on dead code — 407's own header warns a nil-resolving param path hard-fails the
step, an outage-class risk for zero present benefit on an agent that has never run.

Ran a second, independent planning pass via a Fable-model agent (per the owner's/user's
standing instruction to use Fable for bug-fix planning), giving it the same evidence. It
converged on the same three-call-site finding and the same scope decisions independently, and
added one concrete improvement I adopted: the council submission should use `FORCE=1` with a
stated reason rather than naming a Go registry file as a pseudo-edit anchor — a same-shaped
comment-only-sketch submission was refused server-side on 2026-08-14 (round 2, corr
c78ed496), and migration 406 (this bug's sibling) already used `FORCE=1` successfully with
exactly this reasoning. It also caught that migration numbers had moved (413/414 got
reclaimed and renumbered by another session mid-planning) — re-checked live before writing
anything, confirmed 410–417 are now ALL taken (including a same-number collision at 415
between two other concurrent threads), claimed 418–420.

Wrote the workstream's standing docs before touching any live config, per house norm.

Re-checked migration numbering immediately before writing: 410–417 were ALL taken by
concurrent sessions in the time between planning and writing (including a same-number
collision at 415 between two other threads). Claimed **418, 419, 420** and re-verified
free right before each individual file write, catching nothing further.

Wrote all three migrations + `_ROLLBACK.sql` sidecars, each id-scoped with a byte-exact
pre-state gate (per 406's addendum guidance for future migrations) rather than 406/407's own
looser type-scoped `WHERE`.

**Disagreeing-pair proof, run against the literal query text BEFORE any apply** (real site
ids: relojistas.com `ecf15e75-a966-4900-bcb0-1c85f689dbfd` = backend-capable,
gamesdesign.co.uk `e33263f4-74f8-494f-b191-546845dbbddf` = static control):
- 418 (content-gap-planner): relojistas old=new=141 rows, zero-row diff. gamesdesign
  old=141, new=140, diff={intent-probe}. Exactly as predicted.
- 419 (build-site-planner): relojistas old=new=141, zero diff. gamesdesign old=141, new=140,
  diff={intent-probe}. Exactly as predicted.
- 420 (site-planner): unconditional old-minus-new = {intent-probe} exactly, no site id
  needed. Exactly as predicted.

**Dry-run**: each migration file run with its trailing `COMMIT;` swapped for `ROLLBACK;` —
all three completed both `DO` blocks with no `RAISE EXCEPTION` (syntax valid, pre-state
matched, post-update shape correct), then rolled back — live DB confirmed untouched.

**Mutation test** (proves the pre-state gate can actually fail, not just pass): ran an
isolated `DO` block against the live `content-gap-planner` row with a deliberately wrong
`expected_pre` literal — it raised `MUTATION TEST OK: pre-state gate correctly detects
mismatch` as expected. The check has teeth.

Ready to apply. Council submission next (FORCE=1, per the reasoning in PLAN — this is
genuinely config-only, and a Go-file pseudo-edit anchor risks the "comment-only sketch"
server-side refusal that hit corr c78ed496 round 2 on 2026-08-14).

**Submitted to council** via `097_TRIGGER_council_review_v1.sh` with `FORCE=1` (5 edits: the
three migrations + the register update + the bugs_open/276 update).
`SUBMISSION_CORR=d1ae766d-b1e4-4dec-a8e4-efd9ef064137`, run orchestration
`302030db-66f1-44e1-baaf-57ad1b9e3afc`. Dispatch queue was clear (LAG 0) at submit time.
Per house norm, applying now rather than waiting on the verdict (~30 min typical); will
commit with `Council-Submitted:` and let 098 resolve the trailer at report time.
