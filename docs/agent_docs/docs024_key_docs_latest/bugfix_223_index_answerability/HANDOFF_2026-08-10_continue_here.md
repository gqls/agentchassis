# HANDOFF — `bugs_open/223`, cold start for a fresh chat

Written 2026-08-10 ~15:00Z, **updated ~17:00Z**. **Phase 1 is done, live and proven on
`v1.0.1279`/`v1.0.1283`. PHASE 2 IS NOW BUILT, TESTED AND COMMITTED** (`027bf28a0`,
council `3af67677-601e-4181-ad09-17c7a789f995`) — inert until the next roll, and §4 below
is therefore a record of what was done rather than a task list. **What remains is the
acceptance run after that roll** (§4a).
Read this file, then `PLAN_2026-08-10_223_index_answerability.md` §5 (phasing) and §7
(mutations). Everything below is verified unless marked `[UNMEASURED]`.

---

## 1. State in one paragraph

The landmine-verifier used to report footprints as non-existent when the truth was that
`code_symbols` — a Go-only index — **cannot represent** them. Phase 1 makes the shared
lookup state what it cannot see (computed from a live census, never hardcoded), carries that
mechanically into every persisted verdict, and gates the STALE-bearing prompt behind a
workflow branch. It is **live on `v1.0.1279`, artefact-verified on both replicas, and
behaviourally proven** on the exact entry that produced the original false `STALE`. Council
**APPROVED** round 1, and all five actionable objections were closed the same day. The bug
file stays in `bugs_open/` because phase 2 — indexing Go `var`/`const` — is not done.

## 2. What is live (do not redo any of this)

| thing | where | proof |
|---|---|---|
| census + `NOT ANSWERABLE BY THIS INDEX` + the three caveats | `platform/orchestration/actions/diagnose_code_lookup_action.go` | pod-grep 0→1 on v1.0.1279, both replicas, with a positive and a never-added negative control |
| four additive return keys (`checks_with_rows`, `checks_unanswerable`, `no_code_evidence`, `evidence_line`) | same file | live runs show them populated |
| the runtime lane adopts the same census | `diagnose_load_runtime_action.go:~484` | compiles + `codeEvidenceLine` in the bundle |
| opt-in `note_body_suffix_field` | `append_doc_note_action.go` (`applyBodySuffix`) | 4 of 4 persisted verdicts carry `[code-lookup evidence: …]` |
| the evidence gate | `agent_definitions` `landmine-verifier`, seed `docs/agent_docs/sql_for_agents/365_*.sql` (applied + recorded in `schema_migrations` via `--record-only`) | `evidence_gate` recorded in live runs |
| register | `diagnosis-loop.md` DIAG-036 (capability change) + DIAG-042; `documentation-system.md` DOC-067 cross-ref | committed with the code |

Commits: `1058b5366` (the fix), `362c7c091` (council objections), `bb35e8337` (acceptance),
`50d2ba9cb` (register), plus docs. Council correlation
`495df717-4010-491f-aec0-92c13aaf3809` — APPROVED, and `362c7c091` carries
`Council-Reviewed:`.

## 3. The two things a fresh session will most likely get wrong

**(a) Do NOT credit the gate with work it has not done.** `verify_unverifiable` **has never
executed in production.** `no_code_evidence` is `checks_with_rows == 0` and across 4 of 4
acceptance runs the count was 1, 3, 1, 1. One accidental match is enough: `content` is an
ILIKE **substring** search, so `content: VECTORS` — aimed at a *Python* constant — matched 8
Go rows via `vectorSearch`/`pgvector`. The gate is correct (resolution proven at build time
by `TestSeedConditionResolvesAgainstTheActionsReturnShape`, FALSE branch proven live) and
**unexercised**. The protective work is being done by the evidence layer.

**(b) Do NOT test a rendering change by testing its helper.** A mutation survived in this
lane precisely that way (`WRONG_CALLS.md`, 2026-08-10): silencing
`b.WriteString(scope.lsReachNote())` to `b.WriteString("")` left twelve green tests green.
And the *test written to close a council objection about wiring* had the same hole one level
up. **Which test fails if I delete the CALL, not the function?**

## 4. Phase 2 — BUILT 2026-08-10, awaiting a roll

**Index Go `var` and `const`.** Committed in `027bf28a0`: `ValueDef` + `FileInfo.Values`
(`internal/analysis/types.go`), the `token.VAR`/`token.CONST` arm and `valueDefs()`
(`analyse.go`), one loop in `code_symbols_actions.go`, and `internal/analysis/values_test.go`.
Verified in a clean tree from HEAD because another session had the actions package
temporarily broken.

**§4a — WHAT REMAINS, after the next roll:** run the indexer, then
`SELECT kind, count(*) FROM code_symbols GROUP BY 1` — `var`+`const` should appear near
1,371 and **every pre-existing kind's count must be unchanged** (a drop means an identity
collision, and the kind census is the detector). Then re-fire the landmine-verifier on the
`221` entry: `metaCommentaryPatterns` must now resolve **with its body**, not
"possibly inlined or renamed". Then `bugs_open/231`'s reproduction: a check for
`DeployImageAssetInputSpec` must return the DECLARATION with its `Defaults` map, not only
its two use sites. Finally, `TestMissingKindNoteDisappears` predicts that phase 1's
var/const warning **retires itself** — confirm the live `emptyAnswer("symbol")` stops
naming them.

Original notes on the work, kept: This is what remains of the bug, and it is what unblocks
`bugs_open/231`'s class: the diagnosis loop can find every *use* of a package-level `var`
and never the declaration, so it stops at `UNVERIFIABLE` naming exactly what it cannot see.

- **The kinds are already legal.** `code_symbols_kind_check` permits `type`, `var`, `const`;
  the reader's `codeKindList` already treats them as code. **This is an unfinished write
  path, not a design gap** — which is why it is not RFC-scope.
- **Where:** `internal/analysis/analyse.go:140` walks `*ast.GenDecl` **only** when
  `d.Tok == token.TYPE`; `internal/analysis/types.go`'s `Output` has no field for values;
  `code_symbols_actions.go` (~line 594–615) builds rows from `Functions` and `Types` only.
  Use **spec-level line spans** so the body slice captures the literal — the declaration
  *content* is the evidence 231 needs, not just the name.
- **Size:** **1,371 entries (var 795, const 576)**, ~+23% of the analyser's output.
  > **CORRECTED 2026-08-10:** earlier drafts said 930 (a grep over declaration *openers*,
  > blind to block members) and then 1,173 (an `awk` over block members). Both are counts of
  > declaration TEXT. The figure above comes from building the analyser and running it, which
  > is what actually decides the number. `WRONG_CALLS.md`, same day.
- **Blast radius to measure before submitting, not argue:** the per-kind prune cohorts
  (`prune_floor.go` — a brand-new kind arrives 0%-confirmed and an old binary indexing
  against a DB that already has `var` rows will **refuse the whole prune**; safe direction,
  self-healing, visible in `doc_notes`), one-off embedding cost (~1,371 calls, non-fatal),
  and every diagnosis run's retrieval search space.
- **Pre-registered proof it lands:** `TestMissingKindNoteDisappears` already asserts the
  var/const warning **retires itself** when the census shows those kinds. It should start
  passing for the live-shaped fixture's successor, and the live `emptyAnswer("symbol")`
  should stop naming them.
- **Own commit, own council round** (use `operation: "config_change"` for any workflow-JSON
  edit — the guardian's accepted process objection from round 1).

## 5. Smaller open threads, in priority order

1. **`derive_checks` should not emit a bare `content` query for a footprint item it can see
   is a non-Go file.** This is the fix for §3(a): it would cut wasted checks *and* make
   `no_code_evidence` reachable when a footprint really is entirely unverifiable. **It is
   RFC_005's mechanism (`architecture_review`), not this lane's** — route it, do not patch it.
2. **`RFC_022`** is filed and awaiting an owner ruling: an opt-in, default-OFF field on a
   shared action is simultaneously the owner's prescribed remedy (2026-08-02 §2) and the
   architecture seat's RFC trigger. Three costed options; this lane recommends triggering on
   the accumulated optional-key **count**, not on any single addition.
3. **The slug-cap landmine** filed today (`slugify` cuts at 80 chars; 333 of 356 entries sit
   exactly on the cap; 0 collisions **today**). Nothing to fix; it is a trap, and the check
   is in the entry.
4. **Verdicts written before 2026-08-10 must be read under the old caveat** — for a non-Go
   footprint, `STILL_VALID` was never evidence *for* an entry any more than `STALE` was
   evidence against it. Recorded in DOC-067.

## 6. Commands you will want (full set in `RUNBOOK_223_index_answerability.md`)

```bash
# is the fix in the running binary? (0 before, ≥1 after — and grep a positive control too)
kubectl -n ai-persona-system exec <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "NOT ANSWERABLE BY THIS INDEX"'

# fire one entry
./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug>' <branch>
```
```sql
-- which branch ran, and what the round established
SELECT collected_data->'evidence_gate'->>'next_step_override' AS branch,
       collected_data->'verdict'->'result'->>'status'          AS status,
       collected_data->'lookup'->>'checks_with_rows'           AS with_rows,
       collected_data->'lookup'->>'checks_unanswerable'        AS unanswerable
  FROM orchestration_states
 WHERE collected_data->'input_data'->>'source' LIKE '%<slug>%' AND collected_data ? 'verdict'
 ORDER BY created_at DESC LIMIT 1;
```

⚠ **`landmines-sync.py --apply` consumes the `NEEDS_VERIFICATION` signal** — run
`./scripts/landmines-verify-dispatch.sh` instead if you want new entries verified.
⚠ **No dispatch within ~300s of a chassis pod restart** — the spawn is silently dropped.

## 7. Where the paperwork is

`docs/agent_docs/docs024_key_docs_latest/bugfix_223_index_answerability/` — PLAN (design +
ranked candidates + the mutation table), NOTES (append-only, the missteps are the point),
RUNBOOK (commands with their gotchas), README_where_we_are (owner's plain-prose log),
`EVIDENCE_2026-08-10_prefix_run_verbatim.md` (**the banked BEFORE artefact** — keep it; the
paired before/after is what makes the acceptance mean anything), and the council submission
JSON. Fleet-wide: `016b` §9 (the transferable pattern), `WRONG_CALLS.md` (the surviving
mutation), `LANDMINES.md` (the slug cap), `bugs_open/223` (status banner at the top,
acceptance section at the foot).
