# HANDOFF — 2026-08-03 — the deployed-asset-path lane, and the retraction seam it grew into

**Read this first, then `NOTES_deployed_asset_path.md` (missteps, newest at the bottom) and
`RFC_010`.** Written because the originating session ran long, not because anything is stuck.

**Everything below is committed and live.** Nothing is half-applied. The lane can be picked up
cold, or dropped, without cleanup.

---

## 1. What this lane did, in one paragraph

Took `bugs_open/168` (a shared path helper could not express one asset's published path), found
the filed root cause was too broad and one of its fix candidates actively harmful, fixed the
*class* instead by making one derivation that both the writer and all six readers resolve
through, and closed it live-verified. The council's round-2 objection then exposed a second,
larger problem — the work-item queue has no way to learn that a finding stopped being true —
which became `RFC_010`, two owner rulings, and a retraction seam that is now live.

## 2. State — all verified on chassis `v1.0.1238`, both replicas, 2026-08-03

| thing | state |
|---|---|
| `bugs_closed/168` | **CLOSED.** Live since `v1.0.1229`. Council `abd9b119` APPROVED round 3. |
| `RFC_009` (asset path) | Filed. The code it describes is live. ⚠ **number collides** — see §6. |
| `RFC_010` (retraction) | Filed; **both owner rulings recorded in it**. ⚠ number collides. |
| **Decision 1** (retraction seam) | **LIVE** since `v1.0.1237`. Council `846f4f3d` APPROVED round 1. Registered **WII-009**. |
| Decision 2, safe half | **LIVE** — retraction reaches `unresolved`/`failed`. |
| Decision 2, dedup half | **NOT STARTED, and blocked.** See §4.3. |
| `bugs_open/179` | **OPEN**, unowned. See §4.2. |
| The 11 stale items | Repaired — `cancelled`, reason recorded, pointing at `bugs_open/152`. |

**Verification command that matters** (a positive control only proves the grep works):

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | sed 's|pod/||'); do
  A=$(kubectl -n ai-persona-system exec $POD -- sh -c 'strings /app/agent-chassis | grep -c "Retracted work items no longer reproducing"' 2>/dev/null|tail -1); A=${A:-0}
  C=$(kubectl -n ai-persona-system exec $POD -- sh -c 'strings /app/agent-chassis | grep -c "Phase 2E: derived variant deploy path"' 2>/dev/null|tail -1); C=${C:-0}
  echo "$POD retraction=$A negative_must_be_0=$C"
done
```

⚠ **`N=${N:-0}` is not optional.** `grep -c` prints nothing and exits 1 on a zero count, so a
bare `$(… | tail -1)` captures an empty string. The one control whose zero *is* your proof is
the one that renders as a blank cell and reads like a broken exec. This is in `LANDMINES.md`.

## 3. THE ONE THING TO UNDERSTAND BEFORE TOUCHING ANY OF IT

**The retraction seam is live and has retracted nothing. That is expected, and it is the next
job.**

`CheckResult.Resolved` is opt-in with the unsafe default OFF (owner ruling). Its only adopter,
`check_backend_unreachable`, is enabled by **0** active agent definitions and has produced **0**
work items in all history. So `items_resolved: 0` today means *nobody adopted it*, **not** *it
is broken*. Re-measured after the `v1.0.1238` roll: `site_work_items WHERE result ?
'resolved_at'` → **0**.

If you conclude from a zero that the mechanism is defective, you will "fix" working code. The
estate's standing lesson applies exactly: **a silent mechanism is usually UNDRIVEN, not
missing.**

## 4. What is next, in the order the originating session would take it

### 4.1 Decision 1 ADOPTION — recommended first, highest value, lowest risk

The seam converts a capability into a result only when a check uses it. The job: find checks
that **already compute a positive observation and throw it away**, and have them say so.

- Contract: `platform/orchestration/actions/discovery_checks/registry.go` →
  `CheckResult.Resolved []ResolvedFinding{ItemType, ItemKey|AllOfType, Reason}`.
- Worked example: `check_backend_unreachable.go` (the conversion this lane did).
- **The rule that must not be broken:** retract only on a **positive observation of health**,
  never on an absence of findings. A check that errored or was silently blinded returns an
  empty finding set indistinguishable from a clean site; retraction-by-absence would quietly
  close real defects fleet-wide. The runner enforces the error half structurally (the loop sits
  after `if err != nil { continue }`); the *absence* half is the check author's obligation.
- Use `AllOfType` only when the observation genuinely covers the whole item type for the site.
  It is a separate boolean precisely so the breadth is visible at the call site.
- Each adoption is independently reviewable — do them one or two at a time, and **measure a
  real retraction on a real site** before claiming the seam works.

### 4.2 `bugs_open/179` — the `deploy_path` escape hatch

A caller-supplied `deploy_path` overrides the single derivation entirely and is invisible from
`(asset_key, purpose)`, undermining the contract `168` established. Measured empty across three
populations **including the standing queue** (0 work items, 0 agent definitions, 0
orchestrations carrying a value). Two council seats pressed it at medium.

⚠ **Measuring it empty is not closing it, and this lane has already been wrong once in exactly
this way** — see §5.1. Candidate 1 (make the deployer refuse a purpose it does not own) is
already shipped for the brand-head half; `deploy_path` itself is untouched.

⚠ When measuring `deploy_path` usage, ~~match the **JSON shape** (`'%"deploy_path":"%'`)~~, never
the bare word — a bare `LIKE '%deploy_path%'` over `orchestration_states` returns this lane's
own council submissions.

> **CORRECTED 2026-08-04 by the 179 lane (not this lane's error to start with — the pattern
> originated in `bugs_open/179` and this file inherited it): the JSON-shape pattern is BROKEN.**
> Postgres renders `jsonb::text` with a **space** after the colon, so `LIKE '%"deploy_path":"%'`
> can never match a jsonb column — the census returns 0 whatever the data holds, and it did:
> that structural zero reached a council submission before being caught. Use a spacing-tolerant
> regexp requiring a non-empty value, and **induce a non-zero before trusting a zero**:
> `WHERE collected_data::text ~ '"deploy_path"\s*:\s*"[^"]+"'`. Full account: `WRONG_CALLS.md`
> 2026-08-04 and the footprinted `LANDMINES.md` entry the same day. (§4.2's conclusion survives
> re-measurement; `bugs_open/179` finding A is since FIXED, LIVE and CLOSED — see
> `bugs_closed/179` and `bugfix_179_deploy_path_override/HANDOFF_2026-08-04_continue_here.md`.)

### 4.3 Decision 2's dedup half — real work, do not fold it into anything else

Owner ruled `unresolved` is OPEN, so it must leave `idx_swi_dedup`'s exclusion list. **Two
things make this a project, not an index swap:**

1. **87 duplicate rows block it.** 48 colliding `(site_id, item_key)` pairs across 135 rows —
   `undeployed_asset` alone has 47 rows under 20 keys. `CREATE UNIQUE INDEX` against that
   population **fails**. The cleanup is a prerequisite and needs the same "which copy do I keep,
   and does discarding the rest lose a true finding?" judgement the 11 items needed, at eight
   times the scale.
2. **The ordering is asymmetric and one direction breaks the fleet.** `ON CONFLICT` infers its
   index by requiring the clause to *imply* the index predicate:
   - **Go first** → `NOT IN (6)` does not imply `NOT IN (7)` → **42P10 on every keyed insert,
     fleet-wide.** Not hypothetical: `work_items_common.go`'s own comment records this happening
     when migration 157 added `cancelled`.
   - **Index first** → inference still succeeds, but Go stops treating `unresolved` rows as
     conflict targets, so a colliding insert raises **23505** instead of deduping — precisely
     the case the change exists to fix, and `undeployed_asset` has 20 such keys today.

   So: collapse the 87 → change the index → roll promptly, with a known short window. Council
   gate, reviewed migration, someone watching the roll.

### 4.4 Or just take the next unowned bug from `bugs_open/`

The original standing task. `scripts/who-owns.py <n>` **plus** a grep of live `.jsonl`
transcripts — the script reads commits and is blind to a session mid-fix.

## 5. The five things this lane got wrong (all in `WRONG_CALLS.md`)

These are the reason to read `NOTES` before trusting any measurement in these docs.

**5.1 I told a council twice a clobber path was unreachable. It was reachable.** I measured the
predicate that stops *new* items and the *readers*; the exposure was in a *writer*, and eleven
items **already existed**. *A predicate change stops the tap; it does not empty the bath* — and
nothing in this platform sweeps a queue for items whose defining predicate has since moved.
That finding is the whole of `RFC_010`.

**5.2 I nearly implemented a filed fix candidate that would have caused the drift it claims to
fix.** `bugs_open/168`'s candidate 2. **Read the writer before believing the direction of a
reader/writer mismatch.**

**5.3 My own council submissions became rows in the table I was measuring.**

**5.4 I asserted my result shape "follows the platform convention" without grepping.** It
existed nowhere but my own code. **Compliance claims need evidence attached like any other.**

**5.5 My fix for a silent failure introduced a second silent failure.** The refusal branch fell
through with no signal.

## 6. Traps specific to these files

- **`RFC_009` and `RFC_010` are each AMBIGUOUS — two unrelated papers.** Cite by slug, never by
  number. `CLAUDE.md`'s OWNER RULING 2026-08-02 cites the **other** `RFC_010`
  (`..._who_may_answer_a_page_name_collision.md`), which is RATIFIED. Cause: the number ledger
  in `PROCESS_architecture_review.md` had been dead since `RFC_002`; restored, next free number
  stated there. **Claim your number in the same commit as the paper.**
- **`workItemClosedStatuses` must NEVER reach an `ON CONFLICT` clause** — only
  `workItemTerminalStatuses` matches `idx_swi_dedup`. A source-scan test guards this; if it
  fails, read `work_items_common.go`'s comments before "fixing" it.
- **`CheckResult` is an ambiguous symbol.** `internal/adapters/browserrunner` declares an
  unrelated type of the same name which **is** serialised. The discovery-checks one is not a
  wire type.
- **`check_image_url_404.go` is being actively edited by another lane** (a third `<img src>`
  shape landed 2026-08-03). Co-ordinate before touching it.
- **Use `git commit -F <file>`.** Backticks in `-m` execute; this lane lost two words from a
  commit message to it.
- **Test against `git archive HEAD`** when the working tree is broken by another session — and
  **delete the extraction afterwards**; they are hundreds of MB on a shared tmpfs.

## 7. Correlations, for anyone re-reading the trail

| what | id |
|---|---|
| `168` council (3 rounds → APPROVED) | `abd9b119-d274-43bf-a03f-cf45bfb6b881` |
| `168` diagnosis loop (REFUTED — read why in NOTES) | `ae9404bd-dab7-4606-ade3-c439ebda93af` |
| RFC_010 Decision 1 council (APPROVED r1) | `846f4f3d-8958-4e4c-be81-d5f02e20852d` |

## 8. Where the documents are

`docs/agent_docs/docs024_key_docs_latest/bugfix_168_deployed_asset_path/` — PLAN, RUNBOOK,
NOTES, README_where_we_are (the owner's log, **append only**), two SUMMARYs, the repair SQL,
and this handoff. Register: **IMG-067** (asset path), **WII-009** (retraction seam), plus
corrections to **IMG-066**. Papers: `architecture_review/RFC_009_one_derivation_…` and
`RFC_010_discovery_checks_…`.
