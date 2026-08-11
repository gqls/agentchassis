# HANDOFF — `bugs_open/223` lane, cold start for a fresh chat

Written 2026-08-10, superseded 2026-08-11 ~10:15Z, **superseded again 2026-08-11 ~12:45Z**
(after the RFC_022 ruling landed and was implemented, and the staleness 090 was filed).
**BOTH PHASES ARE LIVE AND BEHAVIOURALLY PROVEN; RFC_022 IS RULED AND ITS INTERIM IS LIVE
IN BOTH ROSTERS.** §4's acceptance has been run and passed — do not re-run it, and do not
trust its two numbers (both corrected below). Verified through the v1.0.1286 roll
(census intact; phase 1 marker 1/0 with negative control; image digest matches the pods;
phase 2 commit `027bf28a0` is an ancestor of the image revision `c3b424c8e`).
Read this file, then `NOTES` (newest at the bottom) and `SUMMARY_2026-08-11_*`.
Everything below is verified unless marked `[UNMEASURED]`.

---

## 0. WHAT IS ACTUALLY LEFT (read this and you can stop)

1. **READ THE 090 VERDICT on the staleness finding — it is IN FLIGHT.**
   Filed 2026-08-11 ~12:40Z, `RUN_CORRELATION_ID=520b2f7e-5473-4655-8f41-9a04b7b9eab1`
   (the run key — the intake correlation is different and joins to nothing). Find it:
   `SELECT current_step, status FROM orchestration_states WHERE
    collected_data->'input_data'->>'fix_correlation_id' LIKE '520b2f7e%';` — or the
   diagnosis artifacts under that correlation. The symptom: `diagnose_code_lookup`'s
   answerability layer reports kind + extension censuses but never the indexed COMMIT, so
   a symbol committed after the last push returns 0 rows explained in kind/extension
   vocabulary (§8 below, mechanism proven both ways). A CONFIRMED verdict → implement on
   its cited seam (likely: surface `commit_sha`/`commit_time` in `codeIndexScope` and the
   evidence line). REFUTED → record the correction in NOTES + LANDMINES, visibly.
2. **BUILD THE RFC_022 COUNTER — it is what closes the RFC** (owner ruling 2026-08-11:
   option 3 with option 1 as interim; interim is LIVE via migrations `381`+`383`).
   The counter: a sweep over `RegisterActionInputSpec` declarations per action reporting
   each shared action's optional-key count, so the seat can trigger on ACCUMULATION
   rather than on any single addition. Precedent shape: `cmd/config-key-audit` grew
   `--single-owner-actions` for RFC_006 — extend that command rather than birthing a new
   one. Until it exists the estate runs with a **stated** blind spot (the seat prompt
   itself says what it no longer watches).
3. **`derive_checks` should not emit a bare `content` query for a footprint it can see is
   non-Go** — RFC_005's mechanism, route it, do not patch it here. (This is also what
   would make the never-fired `verify_unverifiable` gate reachable — §3a.)
4. **`099_SYNC_gate_roster.py` needs to learn migration 377** — owned by the
   `council_gate_cost` lane (told in their NOTES, 2026-08-11). Until fixed, `--apply` is
   SUSPENDED (CLAUDE.md + LANDMINES): it would rebuild all 17 gate prompts pre-hoist and
   destroy the measured 68% caching saving. Mirror seat changes with a surgical anchored
   migration instead — worked pair `381`/`383`; health check = 17 seats marked, **1**
   distinct prefix.

## 1. State in one paragraph

The landmine-verifier used to report footprints as non-existent when the truth was that
`code_symbols` **could not represent** them. Phase 1 makes the shared lookup state what it
cannot see (computed from a live census, never hardcoded), carries that mechanically into
every persisted verdict, and gates the STALE-bearing prompt behind a workflow branch. Phase 2
removed two of those blind kinds outright by indexing Go `var` and `const` **with their
bodies**. Both are live, artefact-verified, and proven on the exact entry that produced the
original false verdict. The bug file stays in `bugs_open/` per the owner ruling of 2026-08-06
(a finished bug stays there), not because work is outstanding.

## 2. What is live (do not redo any of this)

| thing | where | proof |
|---|---|---|
| census + `NOT ANSWERABLE BY THIS INDEX` + the three caveats | `diagnose_code_lookup_action.go` | pod-grep 0→1 on v1.0.1279, both replicas, positive + never-added negative control |
| four additive return keys (`checks_with_rows`, `checks_unanswerable`, `no_code_evidence`, `evidence_line`) | same file | live runs show them populated |
| the runtime lane adopts the same census | `diagnose_load_runtime_action.go:~484` | compiles + `codeEvidenceLine` in the bundle |
| opt-in `note_body_suffix_field` | `append_doc_note_action.go` (`applyBodySuffix`) | persisted verdicts carry `[code-lookup evidence: …]` |
| the evidence gate | `agent_definitions` `landmine-verifier`, seed `365_*.sql` | `evidence_gate` recorded in live runs |
| **PHASE 2 — Go `var`/`const` indexed with bodies** | `internal/analysis/{types,analyse}.go`, `code_symbols_actions.go` | **census + verdict flip, 2026-08-11 09:55Z, v1.0.1284 — §4a below** |
| register | `diagnosis-loop.md` DIAG-036 + DIAG-042; `documentation-system.md` DOC-067 | committed with the code |

Commits: `1058b5366`, `362c7c091`, `bb35e8337`, `50d2ba9cb` (phase 1); `027bf28a0`,
`c7c9dd87f` (phase 2). Councils: `495df717-…` and `3af67677-…`, both **APPROVED**.

## 3. The two things a fresh session will most likely get wrong

**(a) Do NOT credit the gate with work it has not done.** `verify_unverifiable` **has never
executed in production.** `no_code_evidence` is `checks_with_rows == 0`, and it keeps not
happening: 1, 3, 1, 1 across the phase-1 acceptance runs, and **8 of 8 / 5 of 10** on
08-11. One accidental match is enough, because `content` is an ILIKE **substring** search.
The gate is correct (resolution proven at build time by
`TestSeedConditionResolvesAgainstTheActionsReturnShape`, FALSE branch proven live) and
**unexercised**. The protective work is being done by the evidence layer. Item 2 in §0 is
what would make it reachable.

**(b) Do NOT test a rendering change by testing its helper.** A mutation survived in this
lane precisely that way (`WRONG_CALLS.md`, 2026-08-10): silencing
`b.WriteString(scope.lsReachNote())` to `b.WriteString("")` left twelve green tests green.
**Which test fails if I delete the CALL, not the function?**

## 4. Phase 2 — BUILT 2026-08-10, **LIVE AND PROVEN 2026-08-11**

**Index Go `var` and `const`.** `027bf28a0`: `ValueDef` + `FileInfo.Values`
(`internal/analysis/types.go`), the `token.VAR`/`token.CONST` arm and `valueDefs()`
(`analyse.go`), one loop in `code_symbols_actions.go`, `values_test.go`. Council objections
closed in `c7c9dd87f` (block-level docs no longer inherited).

### §4a — ACCEPTANCE: RUN AND PASSED, 2026-08-11. Two corrections to what this section said

> **⚠ CORRECTED — "var+const should appear near 1,371" was WRONG. The figure is 1,204.**
> 1,371 was measured by running the analyser **without** the `exclude_patterns: ["docs/"]`
> that `analyse_repo_local` actually passes it — a proxy for the third time in this lane, and
> this one had been written down as a deploy's pass mark. Comparing a healthy index against
> it reads as 12% of rows silently dropped, which is the exact symptom of the identity
> collision the kind census exists to detect. `WRONG_CALLS.md`, 2026-08-11.

> **⚠ CORRECTED — "it cannot be pod-grepped" is true but no longer the constraint.**
> `bugs_open/153`'s build stamp landed in the same window:
> `kubectl logs <pod> | grep 'build provenance'` → `{"git_commit":"…"}`, then
> `git merge-base --is-ancestor <your-commit> <that-sha>`. **Test ancestry, not equality.**
> This retires the "date it by a descendant's literal" workaround estate-wide. Landmine filed.

**All five criteria met.** The prediction was computed independently — the deployed analyser
(`internal/analysis` is byte-identical `027bf28a0`→`55fc8fc35`→HEAD) run over the exact tree
the indexer fetches — *before* the census was read, so it could have come out otherwise:

| kind | predicted | live | reconciliation |
|---|---|---|---|
| var / const | 694 / 510 | **694 / 510** | — |
| func | 3702 | **3700** | −2: `init` twice in `directory_claims.go` and `check_phantom_internal_links.go` |
| method / struct / interface / alias | 1135 / 1001 / 36 / 42 | **unchanged** | criterion (b): no pre-existing kind moved |

- **`metaCommentaryPatterns`** — verdict flipped from *"no longer resolves as a standalone
  symbol (possibly inlined or renamed)"* to **`STILL_VALID`, "confirmed present at expected
  line ranges"**. Body indexed, 2,853 chars, `validate_page_content.go:1229-1275`.
- **`DeployImageAssetInputSpec`** (`bugs_open/231`'s repro) — one `var` row,
  `deploy_image_asset_action.go:32-44`, body carries the `Defaults` map and the `Deprecated`
  aliases. 231 is unblocked; **whether its own `UNVERIFIABLE` resolves is that lane's run.**
- **The var/const warning retired itself**, live: the evidence line now reads
  `kinds with NO rows: type`.
- **0 cross-kind collisions** measured directly, so phase 2 destroyed nothing.

## 5. Two traps in re-running any of this

1. **The indexer reads the REMOTE tip, not your HEAD.** Use `git ls-remote origin <branch>` —
   `git rev-parse origin/<branch>` reads a local cache. Re-indexing an **unchanged** tree is
   what makes the kind census a controlled comparison; pushing first confounds criterion (b),
   because a pre-existing kind then moves for ordinary churn and nothing can tell that from a
   collision.
2. **`landmines-sync.py --apply` consumes the `NEEDS_VERIFICATION` signal.** Use
   `./scripts/landmines-verify-dispatch.sh`, or dispatch the printed slugs by hand. And no
   orchestration dispatch within ~300s of a chassis pod restart.

## 6. Commands (full set, with gotchas, in `RUNBOOK_223_index_answerability.md`)

```bash
kubectl -n ai-persona-system logs <pod> | grep -m1 'build provenance'   # what is running
git merge-base --is-ancestor <commit> <sha> && echo LIVE                 # is my change in it
./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug>' <branch>    # fire one entry
```
```sql
SELECT kind, count(*) FROM code_symbols GROUP BY 1 ORDER BY 1;            -- the deploy proof
SELECT DISTINCT commit_sha, ref FROM code_symbols;                        -- how stale is it
SELECT status, current_step FROM orchestration_states WHERE correlation_id='<corr>'::uuid;
```
⚠ `code_symbols` has `line_start`/`line_end`, **not** `start_line`; no `indexed_at`.
⚠ Never `ORDER BY created_at DESC LIMIT 1` on a shared table — filter by correlation id.

## 7. Where the paperwork is

`docs/agent_docs/docs024_key_docs_latest/bugfix_223_index_answerability/` — PLAN, NOTES
(append-only; the missteps are the point), RUNBOOK, README_where_we_are (owner's log),
`EVIDENCE_2026-08-10_prefix_run_verbatim.md` (**the banked BEFORE artefact — keep it**), two
SUMMARYs (08-10 and 08-11; the series is the record), and both council submission JSONs.
Fleet-wide: `016b` §9, `WRONG_CALLS.md`, `LANDMINES.md` (three new entries 08-11),
`bugs_open/223` (status banner at the top).

## 7b. RFC_022 — RULED 2026-08-11, interim LIVE, and what the implementation flushed out

Owner ruling: **option (3), a budget on the accumulated optional-key count, with option (1)
as the interim.** An opt-in field, unsafe default OFF, named by no live consumer, is NOT
architecture-scope — all three conditions must hold, and an author asserting zero consumers
without enumerating them is itself the objection. Live in both rosters as a byte-identical
clause: `381` (fix-proposer, 10,442→11,829 chars) + `383` (council-gate, 10,479→11,866),
both applied, recorded, and verified from the rows. The clause deliberately names the blind
spot it opens (accumulation) and keeps the reduced signal (report an observed optional-key
count as `insufficient`). CLAUDE.md carries the ruling as its own section.

**The implementation nearly caused a regression worth more than the ruling** — `099`'s
mirror would have reverted 377 (item 4 in §0; full story in `LANDMINES.md` and RFC_022's
"What shipped" section). The tell: its dry-run drift report listing all 17 seats means the
gate is AHEAD of the mirror, not behind it.

## 8. The staleness finding — now proven BOTH WAYS, 090 filed (was: NOT diagnosed here)

**The code index is only as fresh as the last PUSH, and nothing in the lookup's caveat says
so.** Found by verifying this lane's own new landmine entries: asked about `ValueDef`, the
verifier answered that it was "of kinds not indexed" — **false**, it is a `struct`, and its
three siblings from the same file were indexed at that moment. It was absent because its
commit had not been pushed. Measured 08-11: the index sat **246 commits and 88 changed `.go`
files** behind the working tree.

This is `223`'s own failure mode one level up — phase 1 taught the lookup to state which
**kinds** it cannot represent, and it does; nothing states which **commits** it has not seen,
so a model fills that gap with a kind-shaped story, because kinds are the only vocabulary it
was given. **Not a regression of `bugs_closed/108`**: 108's fix (pin to the live working
branch) is working; this is its residual on a tree where commits outpace pushes.

`[UNMEASURED]` how many of the 392 entries this has actually mis-verdicted — one instance,
found by looking at my own two. Written up in `LANDMINES.md` so nobody is caught by it
meanwhile.

**Update, same day:** the owner's push turned this into a controlled both-ways proof — the
same entry, re-verified after re-indexing at the pushed tip, flipped `NEEDS_HUMAN_REVIEW`
("of kinds not indexed" — false, they are a `struct` and a `func`) → `STILL_VALID` ("are
present"), 2 NOT ANSWERABLE → 0, with the commit the only variable (NOTES, 08-11). That
strengthened the finding: nothing reported the staleness then and nothing reports it now.
**The 090 is now FILED and in flight** — §0 item 1 has the run correlation and the queries.
The remedy remains the diagnosis loop's to cite, not this handoff's to prescribe.
