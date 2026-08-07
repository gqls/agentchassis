# HANDOFF — mortgagecalculator.co.uk COMPLETE ADOPTION — cold start, read this first

**Written 2026-08-05 ~21:00 UTC; state RE-VERIFIED 2026-08-06 ~20:10 UTC** — chassis now
**v1.0.1261** (pods 19:54 UTC), site LOCKED, **0 armed**, homepage is the ORIGINAL
(11,125 B at the wire). §§1–8 below remain accurate; §7 is the work. Supersedes
`HANDOFF_2026-08-03_continue_here.md` as the entry point (its §3 lock correction, §12
chrome map and §13→`bugs_open/191` remain load-bearing — read them second).

**The milestone read-out for the owner:
`SUMMARY_2026-08-06_complete_adoption_tools_rebuilt_unverified.md`** — the lane's first;
read it for the plain-prose arc before diving into sections.

## 0. The owner's rulings, in order — these define the lane

1. **"We don't want to bring it down"** — the live original site is production. Every
   change verified file-by-file at the wire (`RUNBOOK` §10f — use `curl -sf`, and read
   §10f's trap first).
2. **2026-08-04: homepage rebuilt over live original → owner chose RESTORE.** Done,
   verified, components deleted so it is safe BY CONSTRUCTION (0 `page_components` ⇒
   assembles-to-nothing ⇒ skip). Restore point `gqls/sites 825a36994`; the rebuilt
   version is preserved at `fe6b81926` if ever wanted. **The homepage rebuild decision
   remains the owner's.**
3. **2026-08-05: "adopt it completely" + "create an arithmetic checker FIRST — but look
   hard for prior art."** The search found the checker already existed (TL-038). Do NOT
   build a new one.

## 1. State right now

| thing | state |
|---|---|
| Live original site | intact — 1 of N differing (`robots.txt`, Cloudflare, expected), re-verified 2026-08-05 ~20:50 |
| Homepage | ORIGINAL live (11,125 B, 28 links), 0 components, item deferred |
| Guides | 4 rebuilt + live with chrome (first-time-buyer, remortgaging, buy-to-let, negative-equity), new URLs, originals still serve |
| Tool recreations | **all 12 `complete`**; deploys in flight when this was written — 2 deployed (affordability, simple), 12 `page_rerender` items released ~20:55, backstop3 running |
| Site lock | UNLOCKED while rerenders drain — **re-lock when they finish** (§3 below) |
| Golden | `acceptance/GOLDEN_2026-08-05_original_tools.json` — 12/12 originals, 4 vectors |
| Fences | `acceptance/criteria/fact-finder.criteria.json` emitted; 11 pending id-complete rebuilds |
| bugs_open/191 | header CTA predicate — FIXED by another session, live v1.0.1251+, verified working |

## 2. The arithmetic-verification chain (owner's "checker first") — what exists and what's left

**Machinery (all pre-existing, verified live):** `computed_values` Tier-4 check type
(browser-runner; `INSTALL_GATE.sh` PASSED 2026-08-05) · `toolgolden.py` in
`../loancalculator_couk/` (capture / `--compare` / `--emit-criteria`) · enforcement =
fence in the tool's PLAN (`doc_plans`, `subject_type='tool'`, `subject_key`=pages.name)
→ `load_doc_context` → `tool-acceptance-agent`. Precedent row: `tool-loan-vs-savings`.

**This lane's additions (committed):**
- `toolgolden.py` gained a 4th **asym** vector — the uniform vectors falsely convict
  RATIO tools (investor.html: yield/LTV are scale-invariant). Non-regression proven
  11/11 vs loancalculator's `GOLDEN_2026-08-03b`. TL-038 addendum + their NOTES carry it.
- `acceptance/compare_rebuilt.py` — diffs each REBUILT tool against the original's
  golden; carries the old→new URL map (verified against `pages`). **Read its reading-trap
  docstring: renamed id ≠ wrong number.**
- CONTRIB filed in `staged_component_build/` — THEY own the PLAN/fence-authoring backlog
  (owner call 2026-08-05). Hand fences to them; do not write `doc_plans` rows into their
  tranche uncoordinated.

**The remaining sequence:**
1. Rerenders drain → all 12 tool pages `deployed` (new URLs; originals untouched).
2. `python3 acceptance/compare_rebuilt.py` — for each tool: MATCHES = arithmetic parity
   proven; number-diff under a shared id = BLOCKS that tool (fix, re-deploy); one-sided
   diffs = renamed ids (expected — the LLM had no id contract), fence must name new ids.
3. Re-emit fences from the id-complete rebuilds
   (`toolgolden.py --emit-criteria <dir> <new urls>`), hand to `staged_component_build`.
4. Re-lock. Report to owner in `README_where_we_are.md` (plain prose, append-only).

## 3. Site-control facts you must not re-learn the hard way

- **The lock holds the QUEUE, not the site** (08-03 handoff §3 correction; cost the
  homepage). Direct `orchestrate` publishes bypass it. Prove safety at the ARTEFACT:
  `git log -- <domain>/<page>` in `~/projects/sites`.
- **Backstops defer the follow-on items your own work needs.** The guides backstop
  (guide-names guard) outlived its purpose and deferred the TOOL rerenders filed before
  it died — that is why 10 of 12 recreations sat `complete` with pages `needs_rebuild`
  and nothing deployed. A backstop's keep-list must include the page names of whatever
  you are RUNNING, and its lifetime must not exceed your batch. RUNBOOK §10c has the
  pattern; this failure is its newest gotcha.
- **`complete` work item ≠ deployed artefact** (standing landmine; bit again here).
- Queue picks the SITE fleet-wide by oldest item FIFO; within a site it claims by
  PRIORITY. Both measured here.
- Deploys: rebase-never-merge on `~/projects/sites` or the domain drops from the run.

## 4. Corrections this lane has already paid for (do not repeat)

- `pages.rendered_header/footer/head` are vestigial fleet-wide; chrome =
  `site_components` (08-03 §12; LANDMINES).
- `include_statuses:["deployed","active"]` filters `pages.status` where `deployed`
  never occurs (LANDMINES).
- `deployed_at IS NULL` ≠ "does not serve" — curl before claiming blast radius (191's
  measurement correction).
- A whole-site integrity sweep with bare `curl -o` against a missing dir reports EVERY
  file differing (RUNBOOK §10f trap; caused a false catastrophe reading 08-04).
- investor.html golden covers the YIELD half only (one press per page). `[KNOWN GAP]`
- Goldens pin *parity with the original*, never *correctness* (TL-038 landmine 1).

## 5. Files of record

This dir: `NOTES` (technical log incl. every wrong turn) · `README_where_we_are`
(owner's log, append-only, plain prose) · `RUNBOOK` §10 (the whole rebuild chain) ·
`PLAN_2026-07-31` · `acceptance/{GOLDEN_2026-08-05_original_tools.json, criteria/,
compare_rebuilt.py}`. Sibling instrument: `../loancalculator_couk/toolgolden.py`.
Bug: `bugs_open/191` (fixed live; file not yet moved to bugs_closed — verify then move,
or leave for its fixing session). Related lanes: `staged_component_build` (fence
authoring), chrome lane (117/118/166/167/170 — owns render_site_components).

## 6. Immediate next actions for a fresh session

1. `SELECT count(*) FROM site_work_items WHERE site_id='62b5978e-4271-4589-8e00-4baebfc0447c' AND status IN ('triaged','claimed');`
   — 0 ⇒ rerenders drained; then **re-lock** (§3 SQL in the 08-03 handoff) and kill any
   backstop still running.
2. Tool deploy census (12 expected `deployed`): the §1 query. Any page still
   `needs_rebuild`/`planned` ⇒ its rerender skipped or failed — check components then
   the orchestration row.
3. `python3 acceptance/compare_rebuilt.py` — triage per §2 step 2.
4. Wire-check a couple of rebuilt tools by hand (chrome present, CSS 200, no literal
   `**`), and re-run the §10f whole-site sweep for the originals.
5. Owner README entry with the compare verdicts, then the fence re-emission + CONTRIB
   update for `staged_component_build`.

---

## 7. COMPARE VERDICTS (run 2026-08-05 ~21:25 UTC) — read before touching the tools

**Deploy state first:** 9 of 12 rebuilt tools are LIVE at their new URLs (wire-checked).
`pages.build_status` is STALE for 7 of them (`needs_rebuild` while the artefact serves —
the batch rerender path commits to git with `success:true` but does not flip the column;
`complete work item ≠ deployed artefact` cuts both ways). **Three are 404 and their
recreations saved ZERO components: `tool-overpayment`, `tool-portfolio`,
`game-fact-finder`** — items read `complete`, nothing was produced. Their
`spec.dispatch_correlation_id` is empty, so find their runs by summary text in
`orchestration_states` (>ret. window may apply). Unfinished, needs a fresh look.

**compare_rebuilt.py on the 9 live: ALL 9 "DIVERGED" — but the diff is dominated by
WHOLESALE ID RENAMING, not proven-wrong arithmetic.** The LLM rebuilds renamed nearly
every element (`amt`→`amount`, `sdltResult`→`totalDue`, …), so almost every diff line is
one-sided (None→value / value→None) — the comparator's unit of comparison was destroyed,
exactly the trap its docstring names. Two REAL behavioural findings survive the noise:

1. **The rebuilds compute on button-press only; several originals computed live on
   input** — golden `after_input` carries numbers where rebuilds show £0.00 until
   pressed. Not an arithmetic verdict, but it changes which phase to compare.
2. **Rebuilt stamp-duty reads £0 even `after_press` on driven inputs** — either its
   button wasn't the one pressed, its validation rejected the driven values, or its
   calculation is genuinely broken. **UNRESOLVED — do not adopt this tool as verified.**

**Arithmetic parity is therefore NOT yet proven for ANY rebuilt tool.** The path (pick
one, the first is the stated rewrite contract):
- **(a) Align the rebuilds to the original ids** — file per-tool fix items ("carry the
  original input/output ids verbatim; keep your button id") and re-run the comparator.
  This is what the fences need anyway.
- (b) Add a per-tool id-map to compare_rebuilt.py and compare after_press only — proves
  arithmetic without touching the pages, but leaves fences unauthorable from goldens.

## 8. Owner-visible summary of where "complete adoption" stands

Content: 4 guides + trial page rebuilt and live; homepage ORIGINAL by owner decision;
23 planned pages remain. Tools: 12 recreations ran; 9 live at new URLs (originals all
still serve untouched); 3 produced nothing (§7); **0 of 12 verified for arithmetic yet**
— the checker chain is built and live, the rebuilds don't satisfy its id contract yet.
Site LOCKED. Nothing armed. Every original file byte-verified after each step.
