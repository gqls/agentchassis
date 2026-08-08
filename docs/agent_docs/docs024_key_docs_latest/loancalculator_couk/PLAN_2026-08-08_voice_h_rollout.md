# PLAN 2026-08-08 — roll voice H across loancalculator.co.uk's copy, through the framework

**Trigger:** the owner reviewed the H-voice canary and ruled *"yes voice H is much
better (still not perfect but is a huge improvement and we can look again later)
please go ahead with the rewrites."* That is approval to proceed, with an explicit
"not perfect" caveat — so this phase rolls out the voice **as it stands** and does
not tune the prompt mid-run. A later pass on the voice itself is expected and is
NOT in this scope.

Predecessor state: `bug_backlog_clearing/HANDOFF_2026-08-07_both_done_next_pickup.md`
(204 + 189 fixed and live — they were the whole blocker), and this lane's
`HANDOFF_2026-08-05_continue_here.md` §1/§3.

## 1. Scope

**Copy only. Calculators kept.** (Owner ruling 2026-08-05, unchanged.)

| | count | note |
|---|---|---|
| active pages on the site | 26 | |
| already rewritten | 1 | `guide-how-loans-are-calculated` — the 204 canary, the page the owner reviewed |
| **to rewrite this phase** | **25** | |
| archived page excluded | 1 | `tool-standard-calc` — retired, serves 404; rewriting it would resurrect nothing |
| prose rows in scope | 44 | of 51 site-wide: 5 belong to the archived page, 2 are the done canary |
| locked calculator rows | 12 | `lock_type='permanent'`, must end the phase byte-identical |

## 2. The route, and why there is no alternative

Every page goes through **`page-build-handler`** with a `content_rewrite` work item —
the same path the 204 canary proved. A CLI applier writing `page_components` directly
is precisely what the owner's 2026-08-04 ruling forbids, however good the output would
look: the point is that framework-written copy passes the framework's checks.

Script: `voiceh_rewrite.sh <page-name>` (this lane).

Three properties of that script are load-bearing and should not be "simplified" away:

1. **The prompt is copied by SQL from the reviewed canary item `2517bc4b`**, with only
   `page`/`page_id`/`page_name` substituted. It is never retyped, so what the owner
   approved is bit-for-bit what every page gets.
2. **The work item is filed `status='detected'`.** The dispatcher selects
   `('triaged','approved')`, so `detected` is never picked up — the direct Kafka
   publish is the only dispatch. Filing it dispatchable would earn every page a
   second, unrequested build hours later (049c's caveat).
3. **The item is completed only after grading.** The handler writes status only on its
   own no-op/failure paths, so a run that succeeds leaves the row `detected`. That is
   the honest state: "built, not yet graded".

## 3. What will happen to the calculator pages, predicted before firing

11 of the 25 pages carry a locked tool row. Traced rather than assumed:

- Every tool component has **0 fields with `source='llm'`** (measured: 12 components,
  0/26, 0/16, 0/20, 0/40, 0/18, 0/11, 0/15, 0/18, 0/14, 0/12, 0/8). An empty
  `llm_field_specs` serialises as *absent*, and `page-content-writer`'s
  `check_render_mode` sends an absent list down `render_from_template` — **no LLM is
  called for a tool section at all.** It re-renders from stored `content_data`.
- Underneath that, `save_page_sections` preloads actively-locked rows, keeps them out
  of the DELETE, and lets the locked copy stand (`:641`, `:769`). So the tool row keeps
  its **row id and its `updated_at`**.

So the arithmetic is protected twice over. **Expected side effect:** each preserved
lock emits a `lock_blocked_change` work item at `needs_human_review` (`:784`). Up to
**12** of them across the run. They are true statements about the mechanism and
spurious in substance — the "blocked" change was a byte-identical re-render. Cancel
them at the end with an explanatory note; do not leave them to look like real HITL work.

## 4. Grading, per page

The pass condition must distinguish the rewrite from **inaction**, not merely from
damage — 189's arc is the standing lesson here (a row count is satisfied identically
by "worked" and "did nothing").

| assertion | how |
|---|---|
| the save actually ran | prose rows have **NEW row ids** vs `page_components_bak_20260807_voiceh` (DELETE+INSERT) |
| the prose actually changed | `md5(content)` differs from `baseline_20260807.json`; length moved |
| the voice landed | the new opening is conditional/situational, not a cold assertion |
| **no fact was lost** | every figure, %, named institution, statute and internal link in the baseline text is still present |
| calculators untouched | locked row ids **and** `updated_at` unchanged |
| it reached the reader | served page HTTP 200, new opening present ≥1, **baseline opening returns 0** (negative control on the artefact, not just a positive match) |

Site-wide, at the end: 26/26 HTTP 200, `toolgolden.py --compare` against
`GOLDEN_2026-08-03b` → 11/11 calculators exact.

⚠ Guard every served-page fetch with `wc -c` + a `DOCTYPE` check first: a
deploy-window fetch returns a B2 error blob at HTTP 200 and every grep against it
reads clean.

## 5. Phasing

1. **Canary — one guide page** (`guide-can-i-overpay`, 3,012 b, statute + figures, no
   tool). Grade it in full, including fact preservation, before anything else fires.
   The 08-06 canary is **not** sufficient evidence for this run: it predates
   v1.0.1263 and today's `page-content-writer` row change.
2. **The 12 remaining guide pages** — prose only, no locks, lowest risk.
3. **The 11 calculator pages + `index` + `tool-credit-roadmap`** — locks in play;
   check the tool rows after each batch.
4. **`legal` last, and graded by hand.** See §6.
5. Site-wide verification, then cancel the `lock_blocked_change` items.

Batch size 4–5, not 25: a bad batch should cost one batch.

## 6. The one flagged risk: the `legal` page

`legal` is copy, so it is in the owner's scope as stated. It is also the one page where
a voice rewrite has a **compliance** failure mode rather than an aesthetic one — a
softened disclaimer reads better and is worse. The spec does say *preserve every
factual claim, figure, named institution and piece of legislation*, which is the right
instruction, but an instruction is not an enforcement mechanism.

**Decision:** it is rewritten like every other page, LAST, and its before/after is read
in full by hand rather than graded by md5. If any disclaimer, permission, statutory
reference or limitation-of-liability shifts in *meaning* — not merely in phrasing — that
page alone is reverted from `page_components_bak_20260807_voiceh` and the owner is told.
Recorded here so the judgement is visible rather than exercised silently.

## 7. Rollback

- `page_components_bak_20260807_voiceh` — 63 rows, taken 2026-08-07 before anything
  fired. Whole-site or per-page restore.
- `baseline_20260807.json` (scratchpad, 76 KB) — every prose row's full text, length,
  md5 and row id, for fact-by-fact comparison as well as restore.
- Per page, the revert is a restore of that page's rows plus an assemble-only rerender
  (spec with **no** `reason`).

## 8. Not in scope, deliberately

- **Tuning the H voice.** The owner said "not perfect… we can look again later".
- **The other finance sites** (mortgagecalculator, lendzy, loancash). Still held on the
  owner's 2026-08-05 condition, and `loanandmortgagecalculator` belongs to session
  `fffe0948`. Per-site seeding with that site's OWN exemplars when released.
- **The fleet-wide base-prompt change** (§6 of the 08-05 handoff) — architecture-scope,
  seven drifted prompts, needs its own council round.
