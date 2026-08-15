# HANDOFF — B2 batch 2 CLOSED (all 21 calculators parameterised, oracle 170/0/6). START HERE.

> **UPDATED 2026-08-15, later the same day — §1 and §2 below are now HISTORY, not
> instructions.** The five items drained at **09:53:58–09:56:01Z** and the full
> verification chain has been RUN. Results, all in one session, all with controls:
>
> | check | result |
> |---|---|
> | live md5 vs the pre-deploy baseline | **all five MOVED**; `class="card"` present on all five |
> | `ported-prose` | 1/1/1/**0/0** — the two zeros predicted, see the ⚠ in §2 |
> | `b2_verify.py` | **5 of 5** — verbatim render, script bodies, classes, ids |
> | per-tool oracle | **PASS 33 / FAIL 0 / CONVENTION 0**, == the pre-deploy baseline |
> | per-tool mutation control | fired correctly (**33 FAIL, 0 PASS**) |
> | **full sweep** | **PASS 170 / FAIL 0 / CONVENTION 6** — matches the 08-11 baseline |
> | parse control | OK |
> | full-sweep expectation control | OK **after fixing the control itself** — see §10 |
>
> **All 21 calculator pages are now in the B2 shape, live and verified.** The next work
> is §6.1 (the last two old-shape pages). Keep §1/§2 for their queue-latency evidence and
> the re-runnable command block; ignore their "still queued" framing.
>
> The chassis rolled to `0115f2b4528b0063fd01e7af275ccefe9c5a991d` at 10:14:35Z, AFTER
> this batch deployed. It bears on nothing here — every change in this batch is Python
> and DB config. (A sibling lane independently confirmed the pods have not restarted
> since 10:14Z and that this is the same build already verified this morning.)

**Written 2026-08-15 by the mixed-card session.** Supersedes
`HANDOFF_2026-08-14_continue_here.md` as the entry point. That file's §4 (the mixed-card
five) is **done up to the deploy**; its §5 wider queue is carried forward here unchanged.
**Three of its stated facts did not survive re-measurement — see §3.**

---

## 0. The state in one paragraph

41 pages on `loanandmortgagecalculator.co.uk` (site id `ed633ada-f8af-424b-b4d4-8af79160dbcd`).
**All 21 calculator pages are now in the B2 shape, LIVE and VERIFIED** — the mixed-card five
were seeded, deployed and proven this session, joining the 16 from batch 1. Every clean copy
span is an unlocked schema field; machinery lives in a per-page `html_template` no content
writer can touch; rows unlocked by design; pages `rebuild_policy='owned'`. Full oracle
**PASS 170 / FAIL 0 / CONVENTION 6**, unchanged from the 08-11 baseline, with parse and
mutation controls fired in-session. Two pages remain on the old Track-B shape
(`loans-consolidation`, `mortgages-repayment`) and they are **two different shapes, not one**
(§3) — that is the next work.

## 1. ⚠ THE ONE THING IN FLIGHT

Five `page_rerender` items, `created_by='claude-session-b2-mixedcard-20260815'`:

```bash
K="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
$K -A -F$'\t' -c "SELECT p.name, w.status, round(EXTRACT(epoch FROM (now()-w.created_at)))::int AS age
 FROM site_work_items w JOIN pages p ON p.id=w.page_id
 WHERE w.created_by='claude-session-b2-mixedcard-20260815' ORDER BY 1;"
```

They were filed at 08:44Z and were still `triaged` at 09:08Z (~24 min). **This is queue
DEPTH, not a stall, and it was measured, not assumed**: the dispatcher's `pre_query` selects
sites with claimable build work and services them in **`site_id` order, one site per
invocation**; this site ranked **15 of 16, then 14 of 15** as the queue drained, with items
ahead falling 88 → 48 over the same period. 27 rerenders completed fleet-wide in one
10-minute window. The items are correctly shaped — identical in every dispatch-relevant
column to `260f03e9`, the item that completed in 3 minutes at 07:58Z (`item_type`, `status`,
`handler_agent='page-rerender'`, `source`, `created_by`, `spec={page_id}` with no
`spec.reason`, `pipeline='build'`, `attempt_count=0 < max_attempts=3`, site not locked).

**DO NOT re-diagnose the items — that path is closed, with proof.** At 09:20Z (36 min
queued) I ran the dispatcher's **own selection query, verbatim from
`load_work_item_actions.go:641-684`**, against this site, and it returns **all five**:

```sql
SELECT left(wi.id::text,8), wi.item_type, wi.priority, wi.created_at::time(0)
FROM site_work_items wi
WHERE wi.site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd'
  AND wi.status IN ('triaged','approved')
  AND wi.attempt_count < wi.max_attempts
  AND (COALESCE(wi.approval_mode,'auto')='auto' OR wi.status='approved')
  AND (wi.depends_on IS NULL OR NOT EXISTS (
       SELECT 1 FROM unnest(wi.depends_on) dep_id
       WHERE dep_id NOT IN (SELECT id FROM site_work_items
                            WHERE site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd'
                              AND status IN ('complete','verified'))))
  AND wi.pipeline='build' AND wi.handler_agent='page-rerender'
ORDER BY wi.priority ASC, wi.created_at ASC LIMIT 50;   -- returns 5 rows
```

So `approval_mode` (`auto`), `depends_on` (NULL), `attempt_count` (0 < 3), `priority`
(100 — same as the item that completed in 3 minutes this morning) are all fine. **The
bottleneck is which SITE the loop picks per invocation, upstream of this query.** Item
ordering within a site is `priority ASC, created_at ASC`, so **a lower `priority` number
wins** if you ever need to jump our own queue — but that reorders only within this site
and will not move us up the site rotation.

⚠ **The site rotation is NOT simply `site_id` ascending, whatever RUNBOOK §10 says.**
Measured over 36 minutes: this site moved 15th → 14th → **10th** while
`loancash.co.uk`, which ranked *behind* it, completed two rerenders at 09:15Z. Sites
after us in `site_id` order are being serviced before us. I did not establish the real
rule (the obvious candidate — oldest-pending-item first, which would put our
same-day items dead last fleet-wide — is **unverified**, and the one LMC item that
completed at 08:01Z was also same-day, which argues against it). **Treat the ordering as
unknown rather than inheriting either story.**

**See where you are in the rotation:**
```bash
$K -A -F$'\t' -c "SELECT row_number() OVER (ORDER BY s.id) AS pos, s.domain, count(*) AS items
 FROM sites s JOIN site_work_items wi ON wi.site_id=s.id
 WHERE s.locked_at IS NULL AND wi.status='triaged' AND wi.pipeline='build'
   AND wi.attempt_count<wi.max_attempts GROUP BY s.id, s.domain ORDER BY s.id;"
```

**If it has genuinely stopped** (nothing fleet-wide completing for ~30 min), the documented
bypass is to fire `page-rerender` directly — RUNBOOK §10, the `kcat` block. **`kcat -P` exits
0 having sent nothing**, so verify by the orchestration row, never the exit code. Prefer
waiting: a direct fire duplicates a queued item (harmless — rerenders are byte-identical and
idempotent — but it is more moving parts than the situation needs).

## 2. THE VERIFICATION CHAIN — run it the moment the five go `complete`

Non-negotiable batch protocol, in this order. **Baselines were captured before the deploy so
each step has a real before/after**, which is the thing batch 1 lacked:

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
P="loans-damage-checker mortgages-bridging-loan mortgages-equity-release mortgages-fee-analyser mortgages-rate-forecaster"

# 1. identity-based post-deploy verify (verbatim tool render + script BODIES + calculators.js + classes + ids)
python3 $LANE/b2_verify.py $P

# 2. per-tool oracle. PRE-DEPLOY BASELINE, measured this session: PASS 33 / FAIL 0 / CONVENTION 0
python3 $LANE/oracle.py --tools rate-forecaster,bridging-loan,equity-release,fee-analyser
python3 $LANE/oracle.py --tools rate-forecaster,bridging-loan,equity-release,fee-analyser --mutate expectation  # CONTROL: must FAIL

# 3. THE FULL SWEEP + controls, before calling the batch closed. Baseline 170/0/6.
python3 $LANE/oracle.py
python3 $LANE/oracle.py --selftest-parse
python3 $LANE/oracle.py --mutate expectation
```

**Pre-deploy live baselines (so you can prove the deploy actually happened):** all five served
bytes **md5-identical to the pinned source** at `7e6b993ef` — `damage-checker 0f7abeec4b2c`,
`bridging-loan b75f3a56ceab`, `equity-release 7351d185fafa`, `fee-analyser 949f48086f63`,
`rate-forecaster c5c5f579637f` — and `ported-prose` absent from all five. A page still
matching those md5s **has not deployed yet.**

> ⚠ **`ported-prose` is NOT a valid deploy tell for two of the five.** `mortgages-fee-analyser`
> and `mortgages-rate-forecaster` decompose to a **single tool block and no prose rows at all**
> (their whole `#content` is the card), so `grep ported-prose` returns 0 on them before AND
> after a successful deploy. Use the md5 change, or `b2_verify`.

## 3. CORRECTIONS to the 2026-08-14 handoff (all re-measured against live)

1. **"16 B2 components" reads 15 — the handoff's own §1 query undercounts by one BY
   CONSTRUCTION.** It matches `description LIKE 'Parameterised calculator component (Track B2%'`,
   and `mortgages-simple` — the page the design was proven on, by hand, before `b2_load`
   existed — carries the older `"Parameterised calculator component: panel,…"`. 15 + 1 = 16.
   The prose was right; the instrument is what misleads. (Now 20 + `mortgages-simple` = 21.)
2. **"19 prose pages" reads 18.** The 08-14 breakdown sums to 42 against a 41-page site.
   Live: 18 prose + 16 B2 + 2 old-shape + 5 (now-seeded) = 41.
3. **The two old-shape pages are TWO SHAPES.** §5 there describes them as one.
   - `loans-consolidation` — locked row, **component-backed** (`function='loans-consolidation'`,
     7,681-char template, zero fields). Carries the fence subject; keep the function name.
   - `mortgages-repayment` — locked row with a **NULL `component_id`**. There is no component:
     3,943 chars of `rendered_html` live directly on the `page_components` instance (card +
     inline script + `calculators.js` all present and intact). An inner join to
     `content_components` **silently drops this row**, which is how it can look like a page
     with no tool at all. It needs a different conversion route from consolidation's.
4. **§7's "the 08-05 backup may be stale for some of the mixed 5 — check per page"** is
   resolved and the answer is mostly STALE: `loans/damage-checker` **0 commits** since 08-05
   (backup valid); `bridging-loan` **2**, `equity-release` **2**, `fee-analyser` **1**,
   `rate-forecaster` **1** (backup STALE — they carry the `bugs_open/224` stale-answer guards
   and the btn-id fix). **Read §4 before re-running that check yourself.**

## 4. ⚠ NEW TRAP, and it fired on me: `pages.name` is NOT the repo path

`pages.name` is the repo path with the **FIRST slash hyphenated** —
`loans-damage-checker` ⇄ `loans/damage-checker.html`. Paste the DB name into a path and you
get `loanandmortgagecalculator.co.uk/loans-damage-checker.html`, which **has never existed**;
`git log` on it prints nothing and **exits 0**, byte-for-byte identical to "this file has not
changed". I ran §7's staleness check that way, got 0-for-5, and wrote down "the backup is safe
for all five" — the **inverse** of the truth (§3.4), on a reading that feeds a **restore**
decision, i.e. straight into the time-machine class this lane spent two days unpicking.

**The check, one line before the one you care about:**
```bash
git cat-file -e "$REF:$SITE/$REL" || echo "PATH DOES NOT EXIST — the git log below is meaningless"
```
Better: `pages.url` already holds the real relative path (`/loans/damage-checker.html`) — use
it instead of reconstructing one from `name`. Full entry in `LANDMINES.md`; incident in
`WRONG_CALLS.md`.

**Also, smaller, same session:** I reported the queue "stalled for an hour" from a DB
`updated_at` against local `date`. **The database is UTC and the box is BST.** Ask the
database for the age — `EXTRACT(epoch FROM (now()-created_at))` — so both ends of the
subtraction come from one clock.

## 5. What changed in the tooling this session (5 commits, all narrow)

| commit | what |
|---|---|
| `86517ba07` | `split_ordered(whole_wrapper_classes=())` — **opt-in, default OFF**, so the sibling lane is byte-for-byte unchanged; `decompose_lmc.WHOLE_CARD_PAGES` names the five. The 1500-char mixed-visible ceiling is waived for them and **prints what it waived** rather than skipping silently (it did not fire — every card is under it anyway) |
| `110e178bc` | **`b2_verify.py` was pinned to the POISONED ref `0a0e89326`** — the one `decompose_lmc` had already abandoned because it holds pre-224 arithmetic. Now `from decompose_lmc import PINNED_REF`. Harmless for these five (byte-identical at both pins), **REAL for `loans/standard-calc`**, which differs between them. Also adds `B2_SEEDS` env override |
| `fff8dc06f` | the five seeds **committed** (batch 1's were scratch and are gone — that is why the 08-14 §4 warns about `b2_seeds/` entries that no longer exist) + NOTES |
| `95964e0eb` | landmine: `pages.name` is not the repo path |
| `3efac4820` | WRONG_CALLS: both false readings above |

**Why "take the card whole" is right now and was refuted on 2026-08-05:** back then *whole*
meant **frozen in a LOCKED verbatim row** — a page's h1 and intro became uneditable. Under B2
the block is a parameterised template whose every clean copy span is an **unlocked** schema
field, so the same copy becomes *editable*, which is the owner's 2026-08-13 direction.

**Evidence the batch is sound (all in-session, all with controls):** `gate_wrapper_parity`
passes all five **and its induced-shortfall control still fails all five** (a gate you have
just made pass is exactly when to re-run its control); `b2_build` render == block via **Go's
own `text/template`**; **59 fields** (8/12/14/11/14), no span left literal; python
substitution == the Go render, checked explicitly — **`b2_load`'s docstring claims that
invariant and the code never asserted it**; all five seeded `md5(rendered_html)` equal the
Go-proven bytes; **0 locked**; `rebuild_policy='owned'` on all five.

## 6. THEN: the remaining lane work (owner rulings standing, in order)

1. **`loans-consolidation` + `mortgages-repayment`** — the last two, and **two routes**, see
   §3.3. Unlock decision + parameterise, same pipeline.
2. **Site-spec seed + planner loop** (owner D6, 08-11): seed the spec, let the planner plan,
   reseed until the plan is *reasonably close to today's site*. **Verbatim constraints: the
   site must NOT shrink on rebuild; the exact calculator/guide mix is NOT important; growth
   from the improvement loop is welcome.** `site_plans` still 0 rows. A seeded plan must name
   every tool slot (`save_sections_positional_tool_slot_test.go`: positional slots match; the
   danger is a semantic plan that OMITS a tool slot).
3. **Bug 252 og: half** — AFTER verifying the 251 canonical fix is live (fix commit
   `61abbdbd0`, `Council-Submitted: 33fb41cb`; read the verdict before writing any
   `Council-Reviewed:` trailer).
4. **Complaint-deadline oracle** (loancash) — FOS six-month + limitation rules, verified at
   source, never from the page. FCA caps checked 08-12, CURRENT.
5. **Track C** (loancash decomposition) after the mixed-5 prove which assertions are site-general.
6. **Reuse is still not demonstrated** — a second `page_components` row on the same component
   with different `content_data`. It is the cheapest remaining proof of the owner's "reuse them
   with their own slightly different copy" and nobody has done it.
7. **Stage 2's proof case stands untouched**: the LMC homepage is missing 6 of its 16 required
   links **BY OWNER RULING** ("leave it for stage 2 as proof"). `gate_page_links.py` exits 1 on
   it deliberately — **re-confirmed still 6 of 16 this session. Do NOT "fix" it.**

## 7. Traps carried forward from 08-14 (still live, still true)

1. **`load_lmc.py --restore` is a TIME MACHINE** — its backup table is dated 08-05, before the
   224 fix (08-08) and the btn-id fix (08-09). **Do not restore B2 pages from it at all**; the
   correct rollback for a B2 page is re-seeding from the sites repo at a clean pin. Check per
   page first — **and read §4 before you write that command.**
2. **The manifest's `scripts` key is load-bearing** — body-level inline scripts + the
   `calculators.js` tag ride in `b["scripts"]`, appended at write time. Any NEW consumer must
   do the same or it ships dead calculators.
3. **Counts can be masked; identities cannot.** Use `b2_verify.py`, not ad-hoc greps. The full
   oracle sweep is the mandatory closing gate of every batch — it is the only instrument that
   measures *behaviour* rather than consistency-with-a-chosen-source.

## 8. Rollback for the five, if the deploy goes wrong

Re-seed from the sites repo at `7e6b993ef` — **verified clean this session**: all three
224-class fixes (`19543d40f`, `9bf26db81`, `79d6158d0`) are ancestors of it, and **nothing has
changed on `origin/master` for these five pages since the pin**. The committed seeds
(`b2_seeds/*.json`) hold the exact bytes loaded. The **08-05 backup table is NOT a valid source
for four of the five** (§3.4).

## 9. Read in this order if starting fresh

1. this file
2. `bugs_open/263_…` — the whole Track B→B2 arc
3. `NOTES_…md`, the **2026-08-15 (b)** entry — this session's missteps with their checks
4. `LANDMINES.md` §"`pages.name` is NOT the repo path" and §time-machine; `WRONG_CALLS.md` 08-15
5. `HANDOFF_2026-08-14_continue_here.md` — the batch-1 arc this file carries forward
6. `copy_quality_two_stage/` — the stage-2 lane (another session's; feed it, don't fork it)

**Council note:** everything this session touched is lane tooling under `docs/`, site content
and DB config — out of gate scope (docs are refused client-side and never spend credits).

---

## 10. ⚠ THE CLOSING GATE WAS BLIND IN TWO MORE PLACES — fixed, but read this before you trust it

The full-sweep expectation control reported **"CONTROL DID NOT FAIL — 8 checks PASSED
under a mutation that makes every answer wrong. The checker is inert."** It was **crying
wolf**, and `oracle.py`'s own docstring records this happening once before (2026-08-12,
the determinism checks) with the note that *a control that cries wolf gets ignored*.
**That fix covered ONE immune class. There were three.**

None of the 8 was on a page this batch touched (4 `compare-loans`, 1 `investor`,
3 `overpayment`), so the first move was to **list them** rather than assume innocence:

1. **`raw_contains` checks assert TEXT** — "Option A is Cheaper", "2 Years 3 Months" —
   and that branch returns **before** the numeric mutation is applied, so nothing the
   control does can move them. **By construction, not by accident.** 7 of the 8.
2. **The sentinel collides with a real answer.** The control asserts `100.0`, commented
   as *"a number nothing computes"* — true of nearly every vector, false for one built
   to sit on that boundary:
   `inv(400000, 1200, 300000, 300000, "asymmetric; LTV exactly 100%")` has loan == value,
   so its LTV **really is 100.0%**. The 8th.

Both are now `N/A` with the reason stated, following the crosstool NON-TEST precedent
directly above them. **Proven both directions in-session**, which matters because the fix
touches the same branch the normal path uses:

| run | before | after |
|---|---|---|
| normal full sweep | 170 / 0 / 6, N/A 0 | **170 / 0 / 6, N/A 0** (unchanged) |
| `--mutate expectation` | PASS **8** / FAIL 161 / N/A 7 | **PASS 0** / FAIL 161 / **N/A 15** |

7 pre-existing + 8 newly excluded = **15 exactly** — every one accounted for, not merely
improved. Commit `b40d7d982`.

### 10a. …and that fix was itself WRONG on one axis. Caught in peer review, same day.

A second session on this lane reviewed `b40d7d982` and was right: **I excluded
`raw_contains` from BOTH controls, and only one of them forces it.**

- `expectation` corrupts `want` at a line this branch **never reaches** → immune by
  construction → exclusion **forced**.
- `crosstool` swaps in the **donor's whole check dict, `raw_contains` included** → the
  text check IS genuinely mis-paired → it **must fail**.

Measured across three versions of the file rather than argued:

| oracle.py | crosstool FAIL | N/A |
|---|---|---|
| pre-change (`b40d7d982^`) | 154 | 9 |
| my first fix (`b40d7d982`) | **148** | 16 |
| corrected (`f0eab34e0`) | **154** | 10 |

**The blanket exclusion silenced 6 genuine must-fail comparisons** — compare-loans' four
verdict vectors (they alternate Option A/B between adjacent vectors, so every rotation is
a real mismatch) plus two of overpayment's three prose readings. The third overpayment
reading *is* a true NON-TEST: its `raw_contains` is `"%d Year"` and two of its three
vectors both round to `"2 Year"`, so the borrowed string equals its own — now excluded by
a borrowed==own STRING guard mirroring the float `_true_want` NON-TEST beside it, which is
why N/A lands on 10 and not back at 9. Normal sweep and the expectation control were
re-run unchanged after the correction.

**Two residuals, both stated rather than absorbed:**

1. `--mutate expectation` only ever tested the ~161 checks comparing a parsed NUMBER. It
   never tested the text assertions and it never could. **A `raw_contains` check that has
   silently stopped matching is caught by nothing there.**
2. **`--mutate crosstool` STILL EXITS 1, and did before any of this** — 4 passes
   pre-change, 3 now. A THIRD class, which I did not cause and deliberately did not fix
   inside a batch-closure commit: when the donor vector has **more checks than the
   receiving one**, the rotation pairs by index across vectors of different length,
   `_true_want` is `None` so the float NON-TEST guard cannot fire, and the borrowed
   expectation lands on an **echoed input value** that matches anyway (consolidation
   total-debt-to-clear, bridging net-advance-echoed-back, overpayment whole-years).
   ⚠ **So a red crosstool here is NOT evidence your change broke something** — diff the
   count against a pre-change run first. That is the next contained piece of work on this
   file, and it needs its own change and its own evidence.

**Independent replication:** the peer session ran the whole chain read-only before finding
my work — 5/5 `b2_verify`, the same served md5s, 33/0/0 per-tool, 170/0/6 full sweep — and
derived both mechanisms in §10 **blind**, before seeing the fix. The diagnosis has two
separate derivations, not one.

> ⚠ **Two shell traps fired on me this session; both are in `LANDMINES.md` already.**
> **Backticks in `git commit -m` execute** — two phrases in `b40d7d982`'s message were
> replaced with empty strings, bash printed `command not found` twice next to a success
> line, and forward-only means it stands. Use `git commit -F -` with a **quoted** heredoc
> for any message quoting code. And **`git log -1` may not be your commit** — another
> session committed in between; read your own sha.
