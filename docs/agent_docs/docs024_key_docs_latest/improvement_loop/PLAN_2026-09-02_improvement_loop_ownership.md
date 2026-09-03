# PLAN 2026-09-02 — taking ownership of the improvement loop

Lane opened 2026-09-02 on the owner's instruction ("please take on responsibility for
the improvement loop in this thread"). No lane existed for the loop itself: the only
`improvement`-named directories under `docs024_key_docs_latest/` were single-bug lanes
(`bugfix_150_improvement_loop_false_clean`, `bugfix_193_loop_bool_config`,
`bugfix_323_cta_improvement_refusal`). The pipeline that runs ~50 times a day had no
owner and no standing account.

Ownership checked before opening: `scripts/who-owns.py` on the two nearest bugs (083,
284) names no active owner of the loop; `git log --since=2026-08-20` on
`triage_detect_items_action.go` and `check_site_structural_validity.go` returns nothing.
`bugs_open/405` (the promoter's known-good door) IS owned, by the loanzy lane — that is
a neighbour, not this lane, and I will not touch it.

---

## 1. What the improvement loop is

Post-build quality cycle. A scheduled task (`improvement-sweep`, `scheduled_tasks`)
picks one site every 15 minutes, runs algorithmic discovery checks and — when the site
has changed — LLM auditors, writes what it finds into `site_work_items`, promotes what
it finds into the dispatch queue, and closes with a rerender.

`docs/agent_docs/docs024_key_docs_latest/004_improvement_loop.md` is the design doc.
**It is STALE in two ways** and I have not yet corrected it (see §5):

- It documents a **3-pass audit cap**. Migration `291` replaced that with a
  **convergence gate** on 2026-08-02 — a site fingerprint (rendered components +
  palette + chrome) plus a 14-day cooldown. The live workflow's steps are
  `check_audit_due` / `check_not_converging` / `load_audit_state`, none of which
  appear in 004.
- It says nothing about the **routability guard** (migration-era `bugs_closed/284`),
  which is now the single most consequential thing the loop does.

## 2. State as measured, 2026-09-02

**The loop runs, and it is mechanically healthy.** All figures `[MEASURED 2026-09-02]`
from `orchestration_states` (retention is ~2 days, so this window is all there is):

| fact | value |
|---|---|
| `improvement-sweep` enabled | **true**, every 900s, `max_concurrent` 2 in group `dispatch` |
| last triggered / completed | 2026-09-02 11:59:27Z / 12:00:16Z |
| runs, last 2 days | 98 (80 `complete_clean`, 15 `complete`, 2 failed) |
| distinct domains covered | 32, fairly rotated (max 5 runs, min 1) |
| audit chain actually ran | 24 of 98 (`audit_due=true`) — the gate is discriminating, not stuck |
| `not_converging` fired | 0 of 98 |
| items promoted | 136, all in the 15 `complete` runs |

> **CORRECTION to the fleet's standing belief.** Several documents and one auto-memory
> entry still carry the **owner ruling of 2026-07-29** that "the improvement loop is
> stopped DELIBERATELY … do not re-enable it", on the evidence that `improvement-sweep`
> had `enabled=f` since 2026-05-02. **That is superseded.** Migration
> `389_park_contrast_failures_and_reenable_improvement_sweep.sql` re-enabled it, and it
> has been running since. Anything reasoning from "the sweep is off" — including
> `bugfix_136`'s stated reason for not seeking a live witness — needs re-checking
> against the live row, not the ruling.

## 3. The one real gap: findings nothing can act on and nobody can see

**What a flag-only finding is.** Most findings are jobs: something is wrong and an
agent can fix it, so the item names that agent. Some findings are deliberately *not*
jobs — nobody can automatically repaint a brand, restart a customer's VM, or repoint a
dead third-party image. For those the check leaves `handler_agent` empty on purpose.

**The rule.** Since `bugs_closed/284`, the loop's triage step will not promote a row
that names no handler, because the dispatcher would only stamp it `blocked` — a correct
observation filed as a machinery failure. The guard is right and I am not proposing to
remove it.

**Where it goes wrong.** The guard holds the row at `status='detected'`, which means
"newly found, awaiting triage" — and triage is the step that just refused it. Nothing
else moves it: `detected-item-promoter`'s own live `pre_query` says in as many words
*"Flag-only rows (no handler_agent) are NOT here"*, and a grep for any reader of
`handler_agent IS NULL OR handler_agent = ''` outside migrations and one-off repair SQL
returns nothing. The same class of finding filed by other producers goes to
`needs_human_review`, which IS a visible parking state — **912** such rows today. So one
class of finding lands in two states and only one of them is looked at.

Meanwhile the loop's terminal step reports **`complete_clean`** over the top of it.

`[MEASURED 2026-09-02]` standing backlog, `status='detected'` with no handler:

| item_type | rows | sites | oldest |
|---|---|---|---|
| head_essentials_missing | 978 | 31 | 2026-08-16 |
| heading_promise_unmet | 139 | 23 | 2026-08-25 |
| image_url_404 | 71 | 28 | 2026-07-26 |
| prerequisite_missing | 63 | 31 | 2026-08-25 |
| canonical_mismatch | 48 | 6 | 2026-08-17 |
| structure_floor_unmet | 25 | 25 | 2026-08-25 |
| dead_internal_link_live | 22 | 15 | 2026-08-17 |
| page_content_divergence | 19 | 6 | 2026-08-23 |
| sitemap_entry_dead_live | 8 | 2 | 2026-08-21 |
| archived_page_still_serving | 8 | 5 | 2026-08-26 |
| asset_reference_404 | 4 | 4 | 2026-08-26 |
| **total** | **1,385** | **31** | |

**It is growing.** The `bugfix_284` lane's own README recorded **722** of this class on
2026-08-19. It is 1,385 today — near enough doubled in a fortnight.

**This lane is entitled to pick it up.** On closing, `bugfix_284` listed four things
that would justify reopening it. One was *"the improvement sweep being switched back on
… it is worth one look"*. It has been. This is that look.

## 4. What the backlog actually contains — and it is not what the counts suggest

I probed rather than trusted the rows. `[MEASURED 2026-09-02]`

**(a) 867 of the 978 `head_essentials_missing` rows are one missing skip link.** Not
978 broken pages: one fleet-wide chrome/template omission, filed 867 times as per-page
findings, at a granularity at which nothing can act on it. A single template fix would
retire the lot.

**(b) 36 rows on farmerinsurance.uk claim `["title","skip_link","footer"]` — and two
thirds of that claim is FALSE.** Curled `/about.html` and `/blog/crop-insurance.html`
with an invented-URL 404 control on the same domain (the control returned 9 bytes and
no title, so the probe discriminates): both live pages return 200 with a real `<title>`
and a `<footer>`. Only the skip link is genuinely absent.

  **Mechanism, confirmed in code, not inferred.** `insertWorkItem` writes with
  `dropOnConflict` (`load_work_item_actions.go:1787`), so a re-run drops the new row and
  the ORIGINAL `spec.missing` is never refreshed. The check only retracts a finding when
  `len(missing) == 0` (`check_site_structural_validity.go:1116`). Skip link is never
  present, so the row can never be retracted, and it holds its first-ever missing-list
  for ever. **Any consumer reading `spec.missing` is reading a claim of unknown age.**

**(c) 20 rows on boxingonline.com are true, but they are not about pages.** Every path
on that domain returns 200 with a 114-byte stub —
`<script>window.onload=function(){window.location.href="/lander"}</script>` — including
`/`. The domain is **parked**; it is not serving our site at all. "Missing title,
skip-link and footer" is a true statement about a parked domain and a misleading one
about our page. (This is the known `a parked domain 200s EVERY path` landmine, arriving
from the other direction: here the parked domain generated 20 *true-but-mis-framed*
findings rather than hiding real ones.)

## 5. Work, in the order I intend to do it

Nothing below is started; this section is the plan, not a status.

1. **Correct `004_improvement_loop.md`** — the pass cap and the missing routability
   guard. Cheap, and it is the doc every new session reads.
2. **Retire the false and mis-framed rows** before anyone builds a consumer over them:
   the farmerinsurance staleness (b) and the boxingonline parked-domain framing (c).
   A consumer built over a corpus that is one-third wrong will be distrusted once and
   then ignored for ever.
3. **The skip link (a)** — establish whether the chrome genuinely omits it fleet-wide,
   and if so fix it at the template, which retires 867 rows at once. Verify at the
   served page, never at the row count.
4. **Then, and only then, the structural question**: a flag-only finding filed at
   `detected` has no exit. Either the producers should file at `needs_human_review`
   like their peers, or the guard should demote rather than hold. That is a shared-seam
   change (26 producers write these rows) and therefore architecture-scope — it goes
   through the council gate, and the durable root-cause claim goes through `090` first,
   per the diagnosis-before-debugging norm. **Not before 1–3**: the honest size of the
   problem is not 1,385, and I do not yet know what it is.

## 6. Decisions I expect to need from the owner

- **(D1) Is the skip link wanted?** It is an accessibility affordance. If the answer is
  "we don't want one", the fix is to retire the check, not the pages — and 867 rows go
  with it. If it is wanted, it is a chrome change affecting every page on the estate.
- **(D2) boxingonline.com is parked.** Is that deliberate (a domain we hold but do not
  serve), or has it come unpointed? The answer decides whether the 20 rows are damage
  or noise.

---

## ADDITION 2026-09-03 — item 4 has its evidence, and it is two changes

Census of the producers (NOTES §(xx)). §5 item 4 said *"either the producers should file at
`needs_human_review` … or the guard should demote rather than hold"* — the code answers that
the producers are **two populations that were never one**:

- **5 types (~813 rows) are DEFERRED HANDLERS** — the code says *"THIS PASS"*, *"v1"*, *"gated
  on 251"*. These are not human decisions; they are handlers nobody built. The right consumer
  is a **count with a date** (the daily check family), not a review queue — and a ruling per
  type on whether the handler is coming.
- **7 types (~270 rows) are HUMAN JUDGEMENTS** by the code's own words. These belong at
  `needs_human_review` beside their 912 peers, carrying the brief three of them already write.
- **A retraction contract is missing**: `image_url_404` never emits a `ResolvedFinding` (87
  rows, oldest 07-20, 0 ever closed by the check), nor does `asset_reference_404` for
  `empty_src`. A finding that cannot clear when it stops being true is a defect independent of
  who reads it — fix it first, it is local and self-evidencing per file.

Order, revised: **(4a)** the retraction gaps — two files, own tests, mutation-proven, council;
**(4b)** the seam change (`discovery_checks.go:249` → `writeWorkItem`) routing by producer
class — `090` first for the durable claim, then its own council round, with the 11 producers
named as consumers and told. Item 3's residual (the 458 non-skip-link rows, read at row
level) stays ahead of both: a producer-level census is not a row-level one.

Item 1 (the held-longer-than-N report) — BUILT 2026-09-03, council pending; see the handoff.
