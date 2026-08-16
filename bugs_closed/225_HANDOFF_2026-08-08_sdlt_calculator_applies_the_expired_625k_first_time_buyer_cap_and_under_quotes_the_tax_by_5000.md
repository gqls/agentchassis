# 225 — the SDLT calculator applies the EXPIRED £625,000 First Time Buyer cap, under-quoting a real tax bill by £5,000; and charges the 5% surcharge below the £40,000 floor

**Filed 2026-08-08 by the `loanandmortgagecalculator_couk` lane**, from the
owner-requested arithmetic-validation work
(`docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/HANDOFF_2026-08-08_arithmetic_validation.md`).

**Status: FIXED + LIVE (2026-08-09, both defects, both live sites — see "Fix
landed" section at the bottom).** The owner approved the changed tax figures
and chose to patch the byte-identical twin on mortgagecalculator.co.uk as well
(plan approval, 2026-08-09 — satisfying this file's "What NOT to do" gate).
Oracle: `PASS 17 FAIL 0` live, all three controls re-run green in the same
session. **Kept in `bugs_open/` per owner direction (2026-08-06)** — do not
move to `bugs_closed/`.

## Diagnosis provenance (per the 2026-07-31 owner ruling)

Not run through 090 — **stated substitution**, same ground as
`bugs_open/224:128-172`: the lane DID file the sibling structural claim and the
loop returned NO VERDICT because `code_symbols` indexes one repo
(`gqls/agentchassis`) and one extension (`.go`) — zero rows match any artefact
this bug names (`.html`/`.js` in the `sites` repo). The loop is structurally
unable to look at this bug's evidence. Substituted verification, first-hand
and disconfirmable: the defect was reproduced live at named band-edge vectors
before the fix (`PASS 13 FAIL 4`, the 4 fails carrying the exact expired-rule
DIAGNOSIS lines), the fix was proven locally before shipping and live after
(`PASS 17 FAIL 0`), and the oracle's own controls (`--mutate expectation`,
`--mutate crosstool`, `--selftest-parse`) were re-run green in the same
session so the green could not be a dead checker.

Live page: <https://loanandmortgagecalculator.co.uk/mortgages/stamp-duty.html>
Source: `sites` repo, `loanandmortgagecalculator.co.uk/mortgages/stamp-duty.html`,
inline `calcSDLT()`.

**Two defects, one page, opposite signs. Defect A under-quotes a tax the user
will actually pay. Defect B over-quotes one.** Both are boundary-band defects
and neither is reachable from the page's default inputs.

---

## Why no existing check could ever have caught this

The lane's `GOLDEN_2026-08-05_prechange.json` records what every calculator
answered on 2026-08-05 and `golden_compare_post.py` re-runs it. Those prove
CONSISTENCY. `calcSDLT` has been wrong since it was written, so the golden
recorded the wrong number faithfully and every comparison since has certified
it — [[a-pass-from-a-blind-check-outlives-the-blindness]]. Tier-2
`tool_acceptance` says in its own header that its static checks "CONFIRM, never
refute". Nothing in the estate was asking whether the answer was RIGHT.

It was found by an independent oracle — the HMRC bands recomputed in Python from
gov.uk, never from this page — driven at explicit band edges.

---

## Defect A — First Time Buyers' Relief uses the cap that expired 16 months ago

**The rule, from HMRC** (fetched 2026-08-08):

- <https://www.gov.uk/stamp-duty-land-tax/residential-property-rates> — FTB
  relief threshold £300,000; "5% SDLT on the portion from £300,001 to £500,000";
  above £500,000 buyers "cannot claim the relief" and must "follow the rules for
  people who've bought a home before".
- <https://www.gov.uk/hmrc-internal-manuals/stamp-duty-land-tax-manual/sdltm29805>
  — "From 1 April 2025 the relief applies to purchases of residential property
  for £500,000 or less… If the purchase price is more than £500,000, you cannot
  claim the relief and you must pay the standard rates on the total purchase
  price." **"Between 23 September 2022 and 31 March 2025 the relief applied to
  purchases of residential property for £625,000 or less."**

**What the page does.** The FTB branch is gated on `price <= 625000` — the
temporary cap — and for £500k–£625k charges £10,000 plus 5% of the excess over
£500,000:

```js
if (type === 'ftb' && price <= 625000) {
    ...
    } else {
        // Between 500k and 625k - Rules vary, but often relief is capped or removed.
        // Standard practice: if > 500k, standard rates often apply on the whole or the relief is tapered.
        // Let's use Standard Rates for safety if > £625k, but calculate 5% on the chunk above 300k if <625k.
        tax = (200000 * 0.05) + ((price - 500000) * 0.05);
    }
```

**The page contradicts itself in its own prose.** Immediately above the
calculator it says: *"Following the end of the temporary relief period in March
2025, thresholds have reverted to standard levels."* The standard band table
beside it is correct and current. Only the FTB branch is still on the expired
rule — so the copy is right, the bands are right, and the arithmetic is 16
months stale.

**Measured, live, 2026-08-08** (`oracle.py --tools stamp-duty`):

| price, FTB | page shows | correct (HMRC) | error |
|---|---|---|---|
| £500,000 | £10,000 | £10,000 | — (relief still available AT the cap) |
| £500,001 | £10,000 | £15,000.05 | **−£5,000** |
| £600,000 | £15,000 | £20,000 | **−£5,000** |
| £625,000 | £16,250 | £21,250 | **−£5,000** |
| £625,001 | £21,250 | £21,250 | — (falls through to standard) |

**The error is a flat £5,000 across the whole band** `£500,000 < price ≤
£625,000` — algebraically, the page charges `0.05P − 15,000` where the correct
standard charge is `0.05P − 10,000`. It is the exact value of the relief the
buyer is no longer entitled to.

**Why this one matters most.** It is a false number about a tax a user will
actually pay, in the band where a first-time buyer is most likely to be
stretching, and it errs in the direction that under-prepares them: a buyer
budgeting £15,000 will be asked for £20,000 at completion.

## Defect B — the 5% surcharge is charged below the £40,000 higher-rate floor

**The rule**: <https://www.gov.uk/guidance/stamp-duty-land-tax-buying-an-additional-residential-property>
— "You must pay the higher Stamp Duty Land Tax (SDLT) rates when you buy a
residential property (or a part of one) for £40,000 or more"; property "worth
less than £40,000" is excluded from the higher rates entirely.

The page's `else` branch applies `+ surcharge` to band 1 unconditionally:

```js
let band1 = Math.min(remaining, 125000);
tax += band1 * (0 + surcharge);
```

| price, additional property | page shows | correct | error |
|---|---|---|---|
| £39,999 | £2,000 | £0 | **+£2,000** |
| £40,000 | £2,000 | £2,000 | — |

Below £40,000 the higher rates do not apply, standard rates do, and standard
rates below £125,000 are nil. The page quotes £2,000 of tax on a transaction
that attracts none. Lower impact than A (few residential purchases sit under
£40,000) but the same class and a one-line fix.

## What is CORRECT on this page

Worth stating, because "the SDLT calculator is broken" would be wrong. 13 of 17
boundary vectors pass: every standard-rate band edge (£125,000 / £125,001 /
£250,000 / £925,000 / £1,500,000 / £1,500,001), the FTB nil band and its
ceiling, and the higher rates at and above the floor. The band table and the
5-percentage-point surcharge construction are right and current.

## Fix candidates, ordered by what closes the door

1. **Replace `calcSDLT`'s hand-rolled branches with one banded function plus a
   relief predicate**, and put the thresholds in NAMED, DATED constants
   (`FTB_RELIEF_CAP_2025_04_01 = 500000`). The present code cannot be read for
   correctness because the rule is spread across a gate, three branches and a
   comment that admits uncertainty. This makes the expired-rule state
   unrepresentable rather than merely fixed — you cannot leave a stale cap in
   place if the cap is a dated constant that a reviewer reads at a glance.
2. **Add the `price < 40000` early return** for `additional`.
3. **Regression-lock it**: `oracle.py`'s 17 stamp-duty vectors already encode
   every band edge and both defects. Run it after the fix; it must go to 17/17.

## How to verify

```bash
cd docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
python3 oracle.py --tools stamp-duty
```

Today: `PASS 13   FAIL 4`. After a correct fix: `PASS 17   FAIL 0`.
The oracle's own controls (`--mutate expectation`, `--mutate crosstool`,
`--selftest-parse`) must be re-run alongside, or a green result proves nothing.

## What NOT to do

Do not change the number without the owner seeing this file first. These are
consumer tax figures and a changed answer is a changed claim
(brief §9). The correction is well-evidenced and cited above, but the decision
to publish a different tax number belongs to the owner.

## Related

> **Numbered 225, not 223.** This file was written as 223 and renumbered
> before it was committed: a concurrent session committed a different
> `bugs_open/223` (the landmine verifier's non-Go footprints) in the same hour.
> Theirs was already at HEAD, so it keeps the number.

- `bugs_open/224` — the zero-rate defect family on the same site. Different
  mechanism (duplicated formula), same root discovery method. **Fixed
  concurrently by another session (`71ba7bb76`, sites `ea72609d6`) — no file
  overlap with this fix.**
- Report with the full method, controls and refutations:
  `docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/REPORT_2026-08-08_arithmetic_validation.md`

---

## Fix landed (2026-08-09, this section added by the fixing session)

**The blast radius was three files, not one** — the same 2,775-byte inline
script block, byte-identical in each (verified by assertion before editing):

1. `sites` repo `loanandmortgagecalculator.co.uk/mortgages/stamp-duty.html` —
   this bug's named target. DB-authoritative: the whole document also lives in
   `page_components.rendered_html` (row `55682bc8-0113-4bf1-a10b-08aff6e8ea22`,
   `deploy_mode='verbatim'`), and a reason-less `page_rerender` deploys the DB
   bytes — so the DB row was repaired (repo→DB, via the lane's
   `gate_component_bytes.py --repair`, sha256 re-stamped) BEFORE the sites
   push, closing the revert window.
2. `sites` repo `mortgagecalculator.co.uk/stamp-duty.html` — **a live,
   nav-linked twin this file did not record.** Same defects, same bytes.
   Repo-only (no `pages` row). The adoption lane had routed it to the owner as
   a finding; the owner chose to patch it too (2026-08-09). Their
   framework-built `/tools/stamp-duty/index.html` was already correct — the
   generator never produced the expired rule; only hand-ported pages carried it.
3. `~/projects/domains` repo `mortgagecalculator.co.uk/gemini/02/stamp-duty.html`
   — MORT_SRC, the input `build_site.py` regenerates the LMC page from;
   unfixed it would have resurrected the bug on the next build. Committed
   locally (`c463764`) — that local commit is what protects the build path,
   since `build_site.py` reads the working tree.

   **That commit exposed a separate, pre-existing problem, now fixed** (owner
   decisions, 2026-08-09; full trap in `LANDMINES.md`): the domains repo could
   not push ANYWHERE. Its remote was codeberg, for which no credential exists
   on this machine, and it had been **56 commits ahead** — silently unbacked-up
   for a long time. Even with a credential the push would have been rejected:
   **19 blobs in its history exceed GitHub's hard 100MB per-file limit**, all
   of them `logs-*.json` agent debug dumps (6.6 GB across 369 files, 90% of
   the repo). Remedy: `origin` repointed to `github.com/gqls/domains` (owner
   made it private first; codeberg kept as a named remote), and a
   **log-stripped mirror** pushed — `git filter-repo --path-glob
   '*logs-*.json' --invert-paths`, 1.9 GB → 462 MB, **all 108 commits kept**,
   every non-log file byte-identical. Verified by re-cloning from GitHub: the
   stamp-duty source reads back at `28e04d99…`, matching the local file and
   the live page. The rewrite was done in a MIRROR CLONE, never in place,
   because the working tree carried another session's staged
   `_first_lot_of_domains_/` files that filter-repo's closing `reset --hard`
   would have destroyed.

   **Completed 2026-08-10 (owner instruction): the working repo now tracks the
   clean history and pushes normally.** Done with a **mixed** `git reset
   origin/main`, not `--hard` — mixed moves the branch and rewrites the index
   but never touches the working tree, so all 30,736 files stayed on disk
   byte-identical (verified by hashing every at-risk file either side of the
   switch) while `--hard` would have deleted 87 staged new files and 6.6 GB of
   logs. Safety check before switching: `git diff --name-status c463764
   origin/main` listed exactly 369 deletions, every one a `logs-*.json`, and
   nothing else. The logs remain on disk, now untracked and ignored;
   `.gitignore` (which had existed since March but was never committed, so its
   rules were purely local) is now tracked with a `logs-*.json` glob, closing
   the class. Push proven end-to-end from the working tree (`d219462`).

**The fix** (sites `9d1a17202`, domains `c463764`): the inline block is now a
branch-for-branch JS port of the lane oracle's `oracles.py:sdlt()` — named,
dated constants (`FTB_RELIEF_CAP = 500000` "from 2025-04-01",
`SURCHARGE_FLOOR = 40000`, `SDLT_BANDS`, one `sdltBanded()` helper). The
literal `625000` appears nowhere, comments included — the expired-rule state
is unrepresentable, per fix candidate 1. Golden-pinned strings preserved
(`"First Time Buyer Relief Applied."`; empty breakdown for ftb-over-cap).
`formatGBP` and all element wiring untouched.

**Evidence (all in one session, 2026-08-09):**

| check | result |
|---|---|
| pre-fix live oracle (the checker can see the bug) | `PASS 13 FAIL 4`, all 4 with the expired-rule/floor DIAGNOSIS lines |
| local behavioural proof before shipping (http.server) | `PASS 17 FAIL 0` |
| `gate_component_bytes.py` report → `--repair` → re-run | exactly 1 BYTES DIFFER → `UPDATE 1: 1` → **GATE PASSES** |
| DB stamp check | `content_data->>'sha256'` = recomputed = `817d80c7…` (new file hash) |
| deploy (GitHub Actions run 31304143109) | success; wire poll landed |
| live oracle | **`PASS 17 FAIL 0`** |
| `--mutate expectation` | CONTROL OK — 0 passed, 17 FAIL |
| `--mutate crosstool` | CONTROL OK — 0 passed, 13 FAIL, 4 named refusals |
| `--selftest-parse` | PARSE CONTROL OK |
| golden (3 vectors × 3 phases) | post-fix output **byte-identical** to the recorded pre-fix run |
| golden `--self-test` | PASSED (comparator can still fail) |
| three-way identity | repo = DB = wire `817d80c7…` (LMC); repo = wire `28e04d99…` (twin) |
| negative control | `grep -c 625000` → **0** on both live pages |
| full-estate oracle run | **`PASS 170 FAIL 0 CONVENTION 6`** — with 224's concurrent fix, the whole 176-check estate is green |

**Council**: N/A — the gate refuses site-content submissions client-side
(scope `platform/ internal/ pkg/`); the review of record is the owner gate
this file itself demanded, satisfied by plan approval with the number changes
stated. No platform code was touched; no new mechanism built (no
concept-register entry).

**For the lane** (also in `CONTRIB_2026-08-09_bug225_sdlt_fix.md`): the page's
bytes moved (`c82013b8…` → `817d80c7…`), so
`acceptance/BASELINE_2026-08-05_stored_md5_at_b318a8fad.txt` no longer
describes this page — deliberately left untouched (it is a dated snapshot);
`load_lmc.py`'s guard will now correctly REFUSE on stamp-duty until you
re-baseline, and **`decompose_lmc.py`'s `PINNED_REF=b318a8fad` must move past
sites `9d1a17202` before stamp-duty is decomposed**, or decomposition
re-freezes the buggy bytes. The `--emit-criteria` step your PLAN gated on
"224 and 225 both fixed" is now **unblocked** — both are live and the full
estate is green.

---

## CLOSED 2026-08-16 → `/bugs_closed/` (this section added by the closing session)

**Re-verified live at the wire before moving it, not taken from the section above.**
The fix has been live since 2026-08-09; what changed today is only that the file is
allowed to move.

| check, run 2026-08-16 | result |
|---|---|
| `curl -s https://loanandmortgagecalculator.co.uk/mortgages/stamp-duty.html \| grep -c 625000` | **0** |
| the same page, `grep -o "FTB_RELIEF_CAP *= *[0-9]*"` | **`FTB_RELIEF_CAP = 500000`** — the positive control, so the zero above is an absence and not a failed fetch |
| `curl -s https://mortgagecalculator.co.uk/stamp-duty.html \| grep -c 625000` | **0** (the twin) |
| the DB row behind the LMC page | component `f29254a5-…` (`tool-1` on `mortgages-stamp-duty`) carries `FTB_RELIEF_CAP = 500000` and `SURCHARGE_FLOOR = 40000` |

⚠ **The component id this file names in "Fix landed" (`55682bc8-…`) NO LONGER EXISTS.**
The page was decomposed into `prose-0` / `tool-1` / `prose-2` by the LMC lane's B2 work
after this bug was written. Anyone re-checking this case must resolve the component by
page, not by the id recorded here — and that mortality is itself part of why the class
fix does not address artefacts by component id.

**Why it moves now.** It was held in `bugs_open/` by the owner's direction of
2026-08-06 ("please leave the bugs that you've found in bugs_open"). The owner
**superseded that on 2026-08-12** — *"if it is fixed and live it should be moved"* —
restoring CLAUDE.md's fixed-AND-live bar, which this case has met since 08-09.

**The CLASS is now tracked separately as `bugs_open/288`.** This file's own section
"Why no existing check could ever have caught this" was the finding that outlived the
fix: the evidence register guards COPY, not CODE. Pieces 2+3 of
`PLAN_2026-08-09_facts_into_tool_acceptance.md` were built on 2026-08-16 (`989addb1c`,
council `cff364b8`, register CLM-022 + TL-045) — a tool's criteria fence can now
declare which register facts it encodes, and the daily evidence sweep tells it when
one moves. **That is inert until the next roll and until a fence declares**, which is
why 288 is open and this one is not: the defect this file describes is no longer
reproducible, and the gap it revealed is somebody's live work.
