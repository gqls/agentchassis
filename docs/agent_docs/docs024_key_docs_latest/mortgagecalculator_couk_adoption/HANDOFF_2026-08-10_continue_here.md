# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-10)

> **SUPERSEDED the same evening by `HANDOFF_2026-08-10b_continue_here.md` — start
> there.** §4 actions 1 and 4 are DONE, and §4 action 1's stated test ("read the
> generated JS for £500,000 rather than £625,000") **could not have failed**: the
> pre-366 build already encoded £500,000. See 10b §1(a) for what was measured
> instead, and §1(b) for what the rebuild deleted.

**Supersedes `HANDOFF_2026-08-08c_continue_here.md`** (read it second: its §0 owner
rulings and §3 landmines all stand; its §1 verdict table is superseded by §2 below,
and its §2 items 1 and 1b are DONE).
Site: UNLOCKED. The design this lane is now executing:
**`PLAN_2026-08-09_facts_into_tool_acceptance.md` — read it before writing code.**

## 0. Owner rulings in force (unchanged)

1. Correctness beats fidelity — never copy a wrong original; improve past it.
2. The checker proves results don't differ on identical inputs, and catches wrong
   results. 3. Site stays unlocked. 4. Everything runs from the framework.
5. Both-right-differently → supply BOTH, the alternative on its own signposted page.

## 1. What landed on 08-10

**(a) The legislation watch is PROVEN, not merely armed.** The daily
`evidence-freshness` sweep ran 09:02Z and re-verified all four SDLT facts:
`verified_at` AND the citation's `accessed` both moved 08-09 → 08-10, with **zero
`citation_lost` items**. That is proof *this site* was swept (the task's own
timestamp covers the fleet and proves nothing about us — RUNBOOK §11). It also
closes 08-09's `[UNVERIFIED]`: our python quote extraction and the Go
`VisibleTextFromHTML` extractor agree on all four quotes. **The day-one gotcha is
now spent — a `citation_lost` from here on is a real signal.**

**(b) The three supply-both rebuilds landed, and two are verified to the penny.**
Comparator re-run: `acceptance/COMPARE_2026-08-10_after_supply_both_builds.txt`.

| tool | verdict | what it actually means |
|---|---|---|
| bridging-loan | **VERIFIED** | retained-interest gross-up adopted; 16 rounding-equal, matches the original outright |
| rate-forecaster | DOMAIN-DIFF | **3-phase model landed exactly** — defaults drive to 1,389.58 / 1,525.78 / 1,286.39, the spec's worked check to the penny. Only the `double` vector diverges: it is a **50-year term**, and the rebuild refuses it with a visible *"Please enter a term of 40 years or less."* At 40 years it computes. A stated domain cap, not an arithmetic fault |
| fee-analyser | DIVERGED **(intended — see the trap below)** | one page, both definitions: `tcTotal` £17,384.79 (= the spec's worked check exactly) and `tcOutlay` **£26,841.44 — the original's figure to the penny**. Driven and read directly, since the comparator cannot see it |
| simple | VERIFIED | unchanged |
| repayment | DOMAIN-DIFF | unchanged — rebuild refuses fractional terms, stricter domain, defensible |
| equity-release | DIVERGED | golden `£0` debts are the **unpressed-second-button** artefact (known landmine); rebuild is penny-exact on 100000×1.065^n. Real diff stays max-cash £124k vs £120k — improvement-loop call |
| investor | DIVERGED | golden LTV `0%` — same unpressed-button artefact; rebuild 75.0% exact |
| overpayment | DIVERGED | "0" vs "6 months" — units, rebuild finer |
| stamp-duty | REPLAY-FAIL | still blocked on `#buyerType` option VALUES. But its visible breakdown now shows **£19,750 at £595k FTB** — the correct post-Apr-2025 figure the original under-quotes by £5,000 |

**Nothing regressed. No rebuilt tool computes a wrong number.**

> **NEW TRAP, and it convicted us today: a shared-id comparator CANNOT judge a
> rebuild that was told to ADD outputs.** `compare_rebuilt.py` judges only ids
> present on BOTH sides, so fee-analyser — whose whole specification was "supply
> both figures" — is **structurally guaranteed** to read DIVERGED: the id that
> agrees with the original (`tcOutlay`) is new, therefore invisible to the
> comparison, and the id that is judged (`tcTotal`) is the one we deliberately
> changed. A "supply both" or "add a breakdown" item will always come back red
> here. **Drive the new ids directly before believing a DIVERGED on any tool whose
> spec added outputs** — that is how £26,841.44 was confirmed.

**(c) Piece 1 of the plan is written as migration `366`** — see §3 for its state.

## 2. Where the plan stands

`PLAN_2026-08-09_facts_into_tool_acceptance.md`. Four pieces; the phases matter.

- **Phase A — data and config, no Go, no image roll.**
  - A1 **TODO** — re-shape the SDLT facts to **one fact per threshold/rate**
    (~12 instead of 4). No schema change; each band edge gets its own verbatim
    GOV.UK quote; `value` stays scalar. **This is the prerequisite for Phase C** —
    today `sdlt-standard-bands` is `value: 12` with the bands in prose, and no
    oracle can derive a band table from a `claim` string. Extract quotes
    PROGRAMMATICALLY from the fetched HTML; never retype (emission-rewrite trap).
  - A2 **DONE** — first sweep checked, §1(a).
  - A3 **PART DONE — migration 366 APPLIED (tool-recreation-handler only).**
    `tool-generator` and `tool-improver` still TODO: they need a `read_site_spec`
    step added first (the action exists and is registered, so still config, not
    Go). **366's effect on a real rebuild is UNPROVEN** — see next action 1.
  - A4 **TODO** — create tool PLANs for the twelve recreated tools. They have
    **no `doc_plans` row**, so no criteria, so no Tier 2 and no Tier 4, and
    **zero acceptance runs have ever happened on this site.** This blocks fences,
    blocks Phase B, and turns the existing ladder on here for the first time.
  - A5 **DONE** — comparator re-run, §1(b).
- **Phase B** — Pieces 2 and 3 (the `facts` declaration in the criteria doc; the
  drift fan-out). Go + validator + one seed. Council gate.
- **Phase C** — Piece 4, the oracle. **RFC first** (it changes what a green
  arithmetic check claims), then council.

## 3. Migration 366 — state, and how to finish it

`docs/agent_docs/sql_for_agents/366_tool_recreation_reads_the_evidence_register.sql`
(+ `_ROLLBACK`). Prompt text only: inserts a "Verified facts — these OVERRIDE the
original tool AND the specification" section before `## Design Context`, injecting
`{{.site_specs.specs.evidence_base.writer_block}}`.

**Why it is safe on today's binary, verified not assumed:** no workflow step, no
action, no config key — nothing in it is read by Go. The two-level access on a
possibly-absent aspect is **already proven in the same template**
(`{{if .site_specs.specs.site_archetype.visual_character}}` ships today), so a
site with no register takes the else branch. Additive on the 10 sites with a
`writer_block`, inert on the rest. It ends with a `DO $$ … RAISE EXCEPTION $$`
guard asserting one changed row, the anchor still present exactly once, the
`writer_block` reference exactly once, and the section landing *before* the anchor.

**APPLIED AND RECORDED 2026-08-10** (snapshot `8701375f`, `UPDATE 1`, guard
passed, ledger row written). Two things about *how*, because both will bite the
next person:

- **`--apply` takes EVERY pending file and 11 others were pending**, including
  `324` which refuses by design ("on an older binary this config deploys the
  WRONG asset bytes") and four with pre-state mismatches belonging to other
  lanes. A bare `--apply` would have run them. **Scope it:**
  `MIGRATIONS_DIR=<dir-with-only-your-file> ./scripts/migration/run-migrations.sh --apply`
  — copy your file to a scratch dir, `md5sum` both to prove the copy is the repo
  file, then point the runner at it. The ledger records filename + checksum, so
  the record is identical to an in-place run.
- **The guard caught its own author.** The first draft asserted the
  `writer_block` reference appeared once; it appears **twice** (the `{{if}}` and
  the interpolation) and the probe refused the file. The expectation was wrong,
  not the edit. It is fixed to `= 2` rather than loosened to `>= 1`, because the
  exact count is what makes a double-application visible.

**VERIFIED BEYOND "the row changed"** — a malformed template would have broken
tool recreation for the whole fleet, so the LIVE prompt was pulled from the DB
and parsed + executed through the same engine and funcMap as
`datahelpers.RenderPromptTemplate`, across four data shapes:

| case | result |
|---|---|
| register + `writer_block` | renders the block |
| register, no `writer_block` (the 11th register) | else branch |
| **no `evidence_base` aspect at all (the 6 registerless live sites)** | **else branch, no error, no `<no value>`** |
| `specs` map entirely empty (not a real state) | else branch; the one `<no value>` comes from the pre-existing `identity.industry` line, not this section |

That third row is the no-op case, and it is the one that mattered: a chained
access through a missing map key was the plausible way to break six sites, and it
does not.

**What 366 does NOT do, and must not be described as doing:** an LLM shown a fact
may still ignore it. It lowers the odds of a tool being *built* with a stale
constant; it closes no door. The code comment it asks for beside each registered
constant is a **trace for a human reader** — it must never become the machine
declaration of which facts a tool encodes (Piece 2), because a comment enforces
nothing and a source-scanning consumer would make every comment load-bearing.

## 4. Next actions, in order

1. **Finish A3**: confirm/apply 366 (§6), then **prove it changed behaviour** —
   re-file ONE recreation item (stamp-duty is the right one) and read the
   generated JS for £500,000 rather than £625,000. A prompt change with no
   observed output is a claim, not a result.
2. **A1** — the per-threshold fact re-shape. Prerequisite for Phase C.
3. **A4** — the twelve tool PLANs, then fences. Turns the acceptance ladder on.
4. **Stamp-duty option values** (unchanged from 08-08c §2.2): fold option VALUES
   into the id contract and re-file the recreation so `#buyerType` carries
   `ftb`/`next`/`additional`. Blocks automated replay AND any emitted criteria.
5. **Report the stamp-duty ORIGINAL defect to the owner** — under-quotes by £5,000
   at £595k FTB. `CONTRIB_2026-08-09_original_stamp_duty_fixed_by_owner_decision.md`
   suggests this is already in hand; confirm before re-raising.
6. **Id-alignment on the three stragglers** (affordability, fact-finder, portfolio).
7. Phase B, then the Phase C RFC.

## 5. Landmines live on this work

- **All of 08-08b §4 and 08-08c §3** still stand.
- **NEW (§1b):** a shared-id comparator cannot judge a rebuild specified to ADD
  outputs — DIVERGED is the default verdict there, not a finding.
- **NEW, filed fleet-wide to `LANDMINES.md` 08-09:** parsing `evidence_base`
  through its own typed struct and writing it back **deletes every citation,
  `writer_line`, `unit` and `staleness_days`** — the structs model a subset,
  both live writers use raw maps precisely because of it, and the write succeeds
  with every gate still green. 104 of 115 facts fleet-wide would lose something.
  **Read this before writing any consumer of the register** (i.e. before Phase B).
- **`doc_plans` has no `site_id`** — the criteria document is fleet-global, keyed
  `(subject_type, subject_key)`. A fact id is per-site, so it must resolve against
  the site of the PAGE being driven. `mortgages-stamp-duty` (loanandmortgage) and
  our `tool-stamp-duty` are the same calculator under two keys today; that is luck.
- **A neighbouring lane is active in the same machinery.** It installed 9
  `computed_values` fences on 08-09 at 14:33–14:40Z, minutes after committing a
  message saying they were "NOT installed". Re-measure `doc_plans`; do not carry
  a count from any commit message, including this handoff's.

## 6. Files of record

This dir: `PLAN_2026-08-09_facts_into_tool_acceptance.md` (**the design — read
first**) · `NOTES` (08-10 entry = today's measurements) · `README_where_we_are`
(owner's log) · `RUNBOOK` §11 (the legislation-watch queries) ·
`acceptance/COMPARE_2026-08-10_after_supply_both_builds.txt` (report of record;
the 08-08 one is superseded but kept).
Migrations: `sql_for_agents/366_*` (+ ROLLBACK).
Bugs: `bugs_open/218`, `bugs_open/222`, `bugs_open/225`, `bugs_open/178`.
