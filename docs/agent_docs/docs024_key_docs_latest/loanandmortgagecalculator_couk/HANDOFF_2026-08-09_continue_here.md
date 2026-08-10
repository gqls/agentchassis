# HANDOFF — bugfix 224 and everything it grew into. Start here in a new chat.

**Written 2026-08-09 by the bugfix-224 session.** Covers two sites and the
platform acceptance fences. Everything is done, live and verified — what is left
is three DECISIONS (§3), not tasks. Read §1 and §2; the rest is reference.

---

## 1. NOTHING IS OUTSTANDING — the last item landed before the session ended

**All six loancalculator.co.uk rerenders completed and are verified live**
(sweep 0 of 5 affected, `defect_vectors --live` 16/16, serving guard 26/26).
Both sites are fixed, live and proven; the 17 fences are installed and three are
proven in-cluster. The open items are the three DECISIONS in §3, not tasks.

The queue text below is kept because the pattern recurs: six items sat `triaged`
behind ~73 older ones for roughly an hour, which is waiting, not stuck.

### (historical) how the queued rerenders were tracked

```sql
SELECT spec->>'filename' AS file, status, created_at
FROM site_work_items WHERE source='bugfix224-zero-rate' ORDER BY 1;
```

- **`triaged` = waiting, not stuck.** `find_dispatchable_site` orders
  `created_at ASC`, so a new item joins the back of a fleet-wide queue that has
  run 300+ deep on this estate. Do not re-file them; a duplicate does not jump
  the queue.
- **The fix itself is already written into the database.** The six tool
  components are updated and their `page_components.rendered_html` rows are
  rewritten. The rerender only ASSEMBLES stored rows into the page file and
  pushes it — so nothing is at risk while they wait; the live pages simply still
  serve the old bytes until then.
- **When they complete, verify live** (all four, in one session):

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
python3 $LANE/rewrite/defect_vectors.py --live      # the 8 new 0% cases, on production
python3 $LANE/check_site_serving.sh                 # 26/26, guards the B2 NoSuchKey blob
python3 docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/zero_rate_sweep.py \
        https://loancalculator.co.uk/ index.html tools/compare-loans.html \
        tools/interest-rate-stress-test.html tools/overpayment-calculator.html \
        tools/settlement-calculator.html
# and consolidation BY HAND — the sweep is structurally blind to it (see §4)
```
  Expected: the sweep goes from **5 of 8 pages affected → 0**, and
  `defect_vectors --live` passes all 16.

**Then close it out**: append the live result to `bugs_open/224` (the shared
account), and tell the loancalculator lane in
`loancalculator_couk/CONTRIB_2026-08-09b_bugfix224_session_taking_the_zero_rate_fix.md`.

---

## 2. WHAT IS DONE, AND HOW IT WAS PROVEN

### loanandmortgagecalculator.co.uk — `bugs_open/224` FIXED + LIVE
Six calculators re-implemented the annuity formula privately and none handled a
0% rate: three printed `£NaN` (compare-loans also **inverted its verdict**),
three left a **stale** previous answer, consolidation quoted **£0.00/month**.
Fixed by deleting the private copies and calling the shared
`assets/js/calculators.js` engine, plus one new additive
`calculateBalloonAmortization` for the PCP balloon. Every submit now writes the
DOM — answer or cleared state — so the stale mode is dead as a class.
**Oracle 23 FAIL → 0; full estate PASS 170 FAIL 0; all four mutation controls
green.** Sites `ea72609d6`, rerender `5b55a1ca4`.

### loancalculator.co.uk — the same defect, SIX components, FIXED + LIVE
Found by sweeping the siblings after the owner asked for it fixed "in all the
calculators". Fixed through that lane's own pipeline (template → commit →
`update_component --apply` → `render_tool_row --apply` → assemble-only
rerender). Templates committed `767681e0d`. **Verified before shipping** by
driving the exact rewritten rows out of the database:
`rewrite/probe_zero_rate_rows.py`, 18/18 — then confirmed on the SERVED pages:
sweep **0 of 5 affected** (was 5 of 8), `defect_vectors --live` **16/16**,
serving guard 26/26. **8 new `defect_vectors.py` cases** cover the 0% rate on all
six — that lane had none for these tools, so the fix would otherwise have been
unguarded tomorrow — and all eight score **PROVEN** under `--both`.

### Platform acceptance — 17 fences INSTALLED and proven in-cluster
`--emit-criteria` first refused 10 of 17 tools because their action button had
no `id`. Ids added (markup only), re-emitted **17/17**, **72/72 pinned values
re-derived from `oracles.py`**, installed into `doc_plans` by new
`install_fences.py`, and three proven by real in-cluster runs (standard-calc
6/0, stamp-duty 6/0, consolidation 6/0, zero `not implemented` skips).

---

## 3. THE THREE OPEN DECISIONS

1. **These 17 fences only run when fired BY HAND.** `tool_acceptance_due`
   requires `component_level='tool'` or `page_type='tool'`
   (`discovery_checks/tool_eligibility.go:71-92`); this site has neither — 26
   `content`, 13 `guide`, and every component is `ported-page`/`ported-prose` at
   `section` level. Unattended coverage needs the site decomposed into per-tool
   components, which is what the loancalculator lane did. **Reading "17 fences
   installed" as "the calculators are watched" would be wrong.**
2. **A shared engine for loancalculator.co.uk.** Its six tools now each carry
   their own correct zero branch, following that lane's 08-03 precedent. The
   door-closing version — one shared implementation, as the sibling now has —
   was NOT done: the only shared JS plumbing reaching every page is generated
   from the **fleet-wide `js_snippets` table (no `site_id`)**, so adding to it
   changes a shared mechanism for every site. That is architecture scope and
   wants an RFC, not a bug patch.
3. **Whether to re-baseline the sibling site's golden.** `GOLDEN_2026-08-05_prechange.json`
   predates both fixes. It is still useful as a non-regression control at
   non-zero rates; it is not a correctness baseline and never was.

---

## 4. LANDMINES FOUND THIS SESSION — read before touching any of this

- **`gate_component_bytes.py --repair` would have destroyed a decomposed page.**
  It compared EVERY `page_components` row against the whole repo file, so
  consolidation's writable prose rows would each have been overwritten with the
  entire 12,865-byte document. Fixed: rows are only comparable when
  `deploy_mode='verbatim' AND components=1` — the same predicate
  `loadVerbatimPageHTML` uses — and assembled rows are skipped loudly. **ADO-038's
  "re-run --repair after any builder change" is only safe WITH this fix.**
- **The Tier-4 runner opens ONE page per (url, profile) and runs every check
  against it.** Emitted criteria assume a fresh page per vector. Consolidation
  failed 3 of 4 vectors on `#d-name-2` while the identical steps drove the live
  page perfectly, because each vector ends by removing a row. `install_fences.py`
  now prepends `{"action":"reload"}` to any check whose clicks come BEFORE its
  fills. **My first explanation was a parse race; driving the live page at
  `wait_until="commit"` refuted it before I changed anything.**
- **`render_tool_row.py`'s default `--control-ref` is stale** (`6e8098022`,
  pre-08-03), so its control cannot reproduce rows written since. It correctly
  REFUSED to write. The right control is the commit that produced the stored
  rows — here `767681e0d^`, at which it REPRODUCES on all seven rows. Pass
  `--control-ref` rather than assuming the renderer has drifted.
- **`/tools/standard-calc.html` on loancalculator.co.uk is a 404 with live DB
  rows** (retired by owner ruling; rows never cleaned up). It shares component
  `tool-loan-repayment` with the HOMEPAGE, so a row write touches both. Expect
  two rows, do not file a rerender for the dead page.
- **`zero_rate_sweep.py` is structurally blind to consolidation.** That tool's
  zero is DETERMINISTIC (`newMonthly` initialised to 0), so it is neither `NaN`
  nor history-dependent. A clean sweep for consolidation means *unmeasured*, not
  *passing*.
- **`toolgolden 11/11 exact` cannot refute this defect class.** Vectors scale
  each field's own default ×1/×2/×0.5, and no scaling of 7.9 is 0.

## 5. FOUR TIMES MY OWN CHECKER WAS THE THING THAT WAS WRONG

All four are in `WRONG_CALLS.md` / the lane NOTES, and the pattern is worth
carrying: **on this estate the prior that a red result is your harness is high.**
(1) I graded fractional terms by rounding to whole months where the pages do not
— 6 false "mismatches". (2) A test wrapper with no `<meta charset>` turned a
correct ✅ verdict into mojibake and a FAIL. (3) `verify_criteria.py` collected
only `fill` steps, dropped stamp-duty's `select`, graded a first-time buyer as a
standard one, and reported a **£5,000 mismatch — the very figure `bugs_open/225`
was about** — against a correct tool. (4) A "no NaN on the page" check searched
`page.content()`, which matched **my own fix comment explaining the NaN defect**.
Each was caught by printing the inputs or the raw value rather than trusting the
verdict.

## 6. WHERE THINGS LIVE

| what | where |
|---|---|
| the bug (shared account) | `bugs_open/224` — STATE block at the top |
| sibling bug, already closed by another session | `bugs_open/225` (SDLT, owner-approved) |
| this lane | `docs024_key_docs_latest/loanandmortgagecalculator_couk/` |
| the other lane, and my note to it | `docs024_key_docs_latest/loancalculator_couk/CONTRIB_2026-08-09b_*` |
| oracle + controls | `oracle.py`, `oracles.py`, `invariants.py` (this lane) |
| criteria emit → verify → install | `toolgolden.py --emit-criteria` (sibling lane) → `verify_criteria.py` → `install_fences.py` |
| fire one acceptance run | `docs/leopardessconsulting/scripts/tool_acceptance_run.sh <site_id> <domain> <subject_key>` |
| sites repo | `/home/ant/projects/sites` → GH Actions → B2 |

**Council**: nothing here was submittable — the gate's scope is
`platform/ internal/ pkg/` and every change was site content, DB rows or lane
tooling. No Go changed, so **nothing here rides a chassis build**.

---

## 6b. FIRST NIGHT OF UNATTENDED RUNNING — it found a defect, and the safety held

The sweep selected this site at **03:20 on 2026-08-10** and ran the 14 due tools
(16 eligible minus 2 in cooldown). **13 passed, 1 failed — a REAL defect**, not a
harness artefact: `mortgages/equity-release.html` left the previous answer on
screen for an ineligible age (`if(age < 55) { alert(); return; }`). Enter 65,
read £124,000; change the age to 32, still read £124,000 "for age 65".

**`no_auto_fix` was asked a real question and answered it**: 0 `improve_tool`
items, and the run's own note reads *"NOT auto-fixed — this fence declares
no_auto_fix"*. Yesterday it was an assertion; now it is behaviour.

I then grepped the SHAPE rather than waiting a week, and found **three more**
(bridging-loan's unviable-deal guard, investor's LTV and yield). **Ten
instances in total, and only six were rate-guarded** — the class is *a guard
that leaves a handler without writing the DOM*, which is now a LANDMINES entry.
All fixed and live.

Also fixed: the equity-release fence had pinned the page's **initial markup** as
an expected answer, because the tool wrote nothing for those vectors.
`--emit-criteria` refuses a wholly inert tool but has no gate for a tool inert on
ONE vector. Re-emitted; it now passes 4/0.

## 7. OWNER DECISIONS ON §3, taken 2026-08-09

1. **Unattended fences: YES — DONE, 16 of 17.** `page_type='tool'` on the 16
   calculators that have exactly one component; the eligibility predicate now
   returns exactly those 16, each matching an installed fence. No decomposition
   was needed. **`loans-consolidation` is NOT included** (2 active components
   fails the sole-component clause) and stays manual-only until decomposed.
   Cadence: 7-day cooldown per tool, and each run costs a Sonnet VISION call in
   the agent's `look` step — so budget ~16 vision calls a week, not zero.
   **Two auto-rewrite paths had to be closed first, and the second was one I
   created:** every fence carries `no_auto_fix: true` (binds Tier 4), and
   `page_status_ok` was REMOVED because it was the only check Tier 2 could fail
   — Tier 2 ignores `no_auto_fix` and its `improve_tool` aims at the
   `ported-page` shell shared by ~154 pages on three sites. See the LANDMINES
   entry and the NOTES entry of 2026-08-09 (late).
2. **A shared engine for loancalculator.co.uk: NO, and not worth an RFC.**
   Decision closed. Its six tools keep their own zero branch, following that
   lane's 08-03 precedent. **Do not reopen this as "tech debt"** — it was
   considered, costed (the only shared plumbing is the fleet-wide `js_snippets`
   table, so it is an architecture-scope change) and declined on the grounds
   that the RFC costs more than the duplication does. The eight
   `defect_vectors.py` cases are what stop the copies drifting.
3. **Re-baseline: DONE.** `GOLDEN_2026-08-09_postfix.json` (22 tools, 4 vectors)
   and `BASELINE_2026-08-09_stored_md5_at_b26fdc81b.txt`, both NEW files with
   the 08-05 pair kept. 17 of 41 pages moved and every one was accounted for
   before regenerating.
