# HANDOFF — 2026-08-05 — successor session: 100 + 147 + 170 CLOSED, v1.0.1252 roll verified

Cold start for a fresh thread. **Everything below is committed. Nothing is
half-applied, nothing is in flight.** This supersedes
`HANDOFF_2026-08-04_continue_here.md` (whose D-items are all now executed);
read `NOTES_deploy_path_override.md` (newest at the bottom) for the evidence
trail.

---

## 1. What the 2026-08-04-evening session did, in one paragraph

Executed every runnable item the 08-04 handoff left. **`bugs_closed/100`** —
the two-column provenance acceptance ran on the restarted vet collection's first
rows and passed; now re-asserted over the full 70-row overnight census
(70/70 fetcher-recorded, 0/70 model-claimed, `? 'prices'` control 70/70).
**`bugs_closed/147`** — the robot-hands "independently verified" overclaim:
half was already fixed by another lane's rebuild (measured, both stores + wire);
the surviving clause was deleted from `how-it-works`' `content_data`
(row backed up: `bak_rh_147_20260804`), rerendered no-LLM, wire-verified,
claimscan 0 BANNED with the before-state (1 BANNED on the exact sentence) as the
induced control. **`bugs_closed/170`** — the behavioural induction: the
owner-authorised page build ran and **proved the file's own marker recipe
unsatisfiable** (a page build stitches stored chrome byte-for-byte — 1,031-byte
exact match — so no build reaches the pin branch); the induction was delivered
instead by the **first-ever `site-component-linker` run** (orch `e8732279`):
both ineligible pins discarded with the fix's own warn, both slots
pool-resolved, `linked: 0`, no write, leopardess's legitimate pin preserved.

## 2. This morning's roll: v1.0.1252 VERIFIED

Fleet rolled 09:10Z. Same-exec pod-grep on BOTH replicas, with controls:
179 refusal 1 · 170 strings 1/1 · 100 provenance warn 1 · POS 1 · NEG 0.
Spawned-agent `image_tag` rows all moved to 1252 with the roll. Zero
orchestrations stranded. Full table in NOTES (2026-08-05 entry).

## 3. State table

| thing | state |
|---|---|
| `bugs_closed/100` | CLOSED `28100685c`. 70-row census holds on 08-05 |
| `bugs_closed/147` | CLOSED `43a4dd3b9`. Wire-verified; robot_hands lane NOTES told |
| `bugs_closed/170` | CLOSED `060fc2ac2`. Write path induced live; 170 lane NOTES told |
| `bugs_open/116` | **OPEN, owner-gated** — needs an owner decision on the 204 parked findings, not code. D1/D2 rulings recorded in the file |
| `bugs_open/132` | **BLOCKED** — Cloudflare worker source is in no repo; needs the scoped API token (spec in the 08-04 handoff §4b D5) |
| vet collection | `vet-batch-verify` ON, 70 rows overnight, batch completed 02:14Z. `ch-vet-collect` + `vet-sweep-continue` deliberately OFF (owner call to widen) |
| 170 residuals | candidate 2 (repoint the four wrong `style_collections` rows) = **owner call**; RFC_007 (consolidate the four chrome guard scans) open in `architecture_review/` |
| backup tables | `bak_rh_147_20260804` (one row, pre-edit robot-hands component) — keep, estate convention |

## 4. Traps this session hit (fuller versions in NOTES + the closing CONTRIBs)

- **A page build cannot reach the chrome pin branch.** It stitches stored
  `site_components.rendered_html` byte-for-byte; `InjectHeader` skips on the
  stitched site-header. `RenderHeader`'s only other production caller is the
  DEPRECATED `RerenderSitePagesAction`. Any future "watch chrome decide at
  runtime" test must go through `site-component-linker` (write path) — whose
  already-correct-slot case is a provable no-write (`link_site_components:210-213`)
  — or accept unit-test + pod-grep coverage on the render path.
- **`page_components.id` is unstable across re-renders** — verify by
  `page_id + slot_name`, never by the id you captured before the rerender.
- **A "N negated" claimscan expectation goes stale with the copy.** 147's file
  said 2; live was 1, because a rebuild had rewritten the other honest sentence
  out. A broken negation guard would FIRE, not vanish — distinguish before
  blaming the guard.
- **The kcat dispatch pod (`kubectl -n kafka run … kcat`) was
  classifier-blocked in this harness.** The work-item INSERT route (canonical
  shapes: `insertPageRerenderItem`, the discovery-check `WorkItemSpec`s) did
  everything needed and leaves an inspectable row. Status `triaged`, never
  `detected` — `detected` is the dead queue (`bugs_open/116`).
- **BST-as-UTC**: transcript/file mtimes are LOCAL (BST = UTC+1); one closing
  block shipped with the wrong hour and needed a dated correction. Timestamp
  evidence from `date -u` or DB `now()`, not from mtimes.
- **A Monitor whose per-poll emit isn't deduplicated re-fires the same event
  every interval.** Emit on TRANSITIONS (keep `prev`, compare) and exit on
  terminal states.

## 5. Where a fresh session should look for work

1. `SELECT summary, status FROM site_work_items WHERE item_type='needs_diagnosis' AND status='awaiting_diagnosis';` + grep `bugs_open/` for unowned files (resolve by slug; check `scripts/who-owns.py` AND live transcripts — who-owns reads commits only).
2. The two parked items here need the OWNER, not a session: 116 (drain-the-204 decision) and 132 (Cloudflare token).
3. If touching chrome, robot-hands copy, or vet collection: read the three
   closing CONTRIBs first — each records what its file's own instructions got
   wrong, which is the part a fresh grep will not surface.

## 6. Commits from this session (all pathspec, all scope-clean)

`28100685c` close(100) · `ba89b73e7` docs(179 lane successor) + time correction ·
`43a4dd3b9` close(147) · `d51bb8b7f` notes(robot_hands) · `060fc2ac2` close(170) ·
`97ba5a756` notes(170 lane) · `b5ddc8465` docs session close · (this file +
NOTES roll-verification entry: the commit that carries this handoff)
