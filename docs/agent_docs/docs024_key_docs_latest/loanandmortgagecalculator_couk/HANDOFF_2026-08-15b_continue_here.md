# HANDOFF — B2 COMPLETE (23/23 calculators, reuse proven). NEXT: the planner loop (§1). START HERE.

**Written 2026-08-15 evening, superseding `HANDOFF_2026-08-15_continue_here.md`** (that
file closed batch 2 and then accumulated the day's addenda; it remains the evidence trail
for the mixed-card five, the old-shape two, and the queue-latency measurements). Everything
below was measured against the live DB / live site / sites repo at the moment of writing.

## 0. The state in one paragraph

`loanandmortgagecalculator.co.uk` (site id `ed633ada-f8af-424b-b4d4-8af79160dbcd`):
**41 ACTIVE pages — 18 prose + 23 calculators, every calculator in the B2 shape** (machinery
in a per-page `content_components.html_template`; every clean copy span an unlocked schema
field; `rebuild_policy='owned'`; zero locked rows). **Reuse is DEMONSTRATED, not just
designed**: a second row on the repayment component with different copy served correct
arithmetic (12/12 + control) on a live page, then was retracted through `page-retraction`
(NOTES (g)). ⚠ **A bare `count(*) FROM pages` reads 42** — the demo's archived row is the
framework's terminal state (`page_component_history` FK; 098 precedent); filter
`status='active'`. Oracle baseline **PASS 170 / FAIL 0 / CONVENTION 6, N/A 0**, and — for
the first time in the lane's history — **all three mutation controls are green**
(parse OK · expectation 0/161/15 · crosstool 0/154/13) after three control blind-spot
fixes on 08-15 (NOTES (c)/(f) and commits `b40d7d982`/`f0eab34e0`/`a848812cd`).

## 1. THE NEXT WORK: site-spec seed + planner loop (owner D6 ruling, 2026-08-11)

**The ruling, constraints verbatim:** seed the spec, let the planner plan, reseed until
the plan is *reasonably close to today's site*. **The site must NOT shrink on rebuild; the
exact calculator/guide mix is NOT important; growth from the improvement loop is welcome.**
The ruling's full text: `HANDOFF_2026-08-11_after_track_a_decisions_pending.md` (D6).

Facts a fresh session should start from (verified 2026-08-15 evening):

- **`site_plans` has 0 rows for this site and 33 fleet-wide** — the 33 are your worked
  examples for the row shape; `\d site_plans` before writing anything.
- **The re-slot trap is settled at the CODE level, not the plan level:**
  `platform/orchestration/actions/save_sections_positional_tool_slot_test.go` proves
  positional slots match. **The live danger is a SEMANTIC plan that OMITS a tool slot** —
  a seeded plan must name every tool slot, and there are now **23** (list them from
  `pages.sections` where a `tool-*` slot appears, not from memory).
- The B2 shape is what makes a faithful plan *possible*: every calculator's machinery and
  copy are separable, so a plan can name the tool and let copy regenerate. The
  demonstrated reuse path (NOTES (g)) is the existence proof that one component can back
  a differently-worded page — which is what "the exact mix is not important" will lean on.
- **The site must not shrink** has a measurable meaning now: 41 active pages, 23 tool
  slots, 16 required homepage links (of which 6 deliberately missing — §4). Capture those
  as the plan's floor before the first reseed, in the plan file, with the queries.

## 2. Cheap re-verification before acting (run all four)

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
K="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

# 41 active / 42 total (archived demo row); 23 pages with a tool-* slot
$K -A -F$'\t' -c "SELECT status, count(*) FROM pages
  WHERE site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd' GROUP BY 1;"
$K -A -F$'\t' -c "SELECT count(*) FROM pages
  WHERE site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd' AND status='active'
    AND sections::text LIKE '%tool-%';"

# oracle baseline + at least the expectation control, SAME session
python3 $LANE/oracle.py            # expect PASS 170 / FAIL 0 / CONVENTION 6 / N/A 0
python3 $LANE/oracle.py --mutate expectation   # expect CONTROL OK (0 pass / 161 fail / 15 N/A)
```

## 3. Traps carried forward (each cost a real misstep; pointers, not restatements)

1. **`pages.name` is NOT the repo path** — build paths from `pages.url`;
   `git cat-file -e` before any `git log -- <path>` (LANDMINES; it inverted a
   staleness answer 0-for-5 on 08-15).
2. **The 08-05 backup table is a TIME MACHINE** — never restore B2/converted pages from
   it; rollback = re-seed from the sites repo at a clean pin, or the
   `*_bak_20260815_oldshape` tables for the two converted pages.
3. **A `complete` work item's `result` may hold only a spawn-handler record** (repayment's
   did) — prove a deploy at the repo's **unfiltered** `git log`; an identical-roundtrip
   deploy NEVER appears in a path-filtered one.
4. **`psql -A -t | tr -d '[:space:]'` glues the command tag onto RETURNING output** —
   extract UUIDs with `grep -Eo '[0-9a-f-]{36}'` and assert length (it corrupted a
   retraction dispatch on 08-15; the action refused conservatively).
5. **For identity-preserving conversions, `b2_verify` is a valid PRE-deploy gate**
   (DB-free: live HTTP + pin + seed). Run it BEFORE an unlock, not only after (NOTES (f)).
   For anything meant to change served bytes it tests the OLD page and says nothing.
6. **`kcat -P` exits 0 having sent nothing** — verify by the orchestration row.
7. **The oracle's mutation controls are green as of 08-15** — if one goes red after your
   change, read NOTES (c)/(f) before either believing it or dismissing it: both false-red
   (out-of-scope checks counted as passes) and true-red have happened here.

## 4. The rest of the queue (owner rulings standing, in order after §1)

2. **Bug 252 og: half** — AFTER verifying the 251 canonical fix is live (fix commit
   `61abbdbd0`, `Council-Submitted: 33fb41cb`; read the verdict before writing any
   `Council-Reviewed:` trailer).
3. **Complaint-deadline oracle** (loancash) — FOS six-month + limitation rules, verified
   at source, never from the page. FCA caps checked 08-12, CURRENT then.
4. **Track C** (loancash decomposition) — the mixed-card five + the old-shape two have
   now proven which assertions are site-general; `b2_convert_oldshape.py` is the worked
   example for in-place conversion, `b2_build`/`b2_load`/`b2_verify` for the from-source route.
5. **Stage 2's proof case stands untouched**: LMC homepage misses 6 of its 16 required
   links BY OWNER RULING ("leave it for stage 2 as proof"). `gate_page_links.py` exits 1
   on it deliberately. Do NOT "fix" it — the `copy_quality_two_stage` lane (another
   session's) must, via `section-editor`.

## 5. Read in this order if starting fresh

1. this file
2. `HANDOFF_2026-08-11_after_track_a_decisions_pending.md` — the D6 ruling you are executing
3. `NOTES_…md` entries **2026-08-15 (c)–(g)** — the day's missteps, each with its check
4. `HANDOFF_2026-08-15_continue_here.md` — the superseded evidence trail (batch 2 + old-shape two)
5. `SUMMARY_2026-08-15_every_calculator_is_now_parameterised_and_editable.md` — the milestone read-out
6. `bugs_open/263_…` — the whole Track B→B2 arc

**Council note:** the planner-loop work is site config + lane tooling (docs/ + DB rows) —
out of gate scope. If it grows a platform-code change (a new action, a schema change, a
shared-seam field), that changes: register it in the same commit and submit to the gate.

**Coordination:** three sessions worked this lane on 08-15 (batch-2/mixed-card;
this one — old-shape conversion, oracle control fixes, reuse demo; a third that ran the
owner's apply and found the (f) pre-deploy gate). All signed off; nothing in flight.
Check `MEMORY_workstreams.md` and message the owning session before routing work at an
open item.
