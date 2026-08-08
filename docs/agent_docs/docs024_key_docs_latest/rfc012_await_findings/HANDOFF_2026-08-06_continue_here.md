# HANDOFF — RFC_012 execution · **START HERE** · updated 2026-08-08

Cold-start for the next session. Read this, then `NOTES_rfc012_await_findings.md` (the
missteps — they are the point) and `RUNBOOK_rfc012_await_findings.md` (every command, with
its gotcha).

**This lane owns three owner rulings** (RFC_012 second sitting, recorded in the RFC itself
at commit `3851e90b5` — read that block, it is the authority, not this file):

1. **(d) → a STANDING check, ONLINE if the framework allows.**
2. **The §3(a) reader census — COMMISSIONED.**
3. **Option B implementation — ASSIGNED here.**

**Two of the three are DONE. The third — (d)'s online half — is the main remaining build.**

---

## STATE

| piece | commit | state |
|---|---|---|
| The rulings, recorded in the RFC | `3851e90b5` | durable — survives any session |
| **B core**: `agenterrors` leaf package | `5f49b4cfd` | **LIVE**, pod-proven v1.0.1259 |
| **B rest**: 18 site conversions + tests | `f930de86b` | **LIVE**, pod-proven v1.0.1262, re-proven **v1.0.1263** 2026-08-08 |
| **(d) detector** (offline half): `--shared-output-fields` | `abf5e8266` | **DONE, proven** — a `cmd/` binary, not in the chassis image, needs no roll |
| **(a)/(a′) reader census** — the commissioned artefact | `40992cbce` | **DELIVERED** → `architecture_review/CENSUS_2026-08-07_rfc012_await_step_readers.md` |
| `domain` `NULLIF` follow-up | `8e786652b` | **CLOSED as NO** — measured; adding it would be a regression |
| Concept register: RSH-008 + WFA-011 | 08-07 | **DONE**; RSH-008 status now `deployed` |
| Council round for the whole code set | corr `5c2bc265-84ac-452b-bd8b-22fd7b875427` | **REJECTED** (r2, guardian hard veto) — see below |

**Nothing is broken and nothing is at risk.** The code has been live since 2026-08-07 and is
confirmed still live on today's build. The rejection blocks a credit, not the work.

### Proving it is still live (run this after any roll)

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c '
  strings /app/agent-chassis | grep -cF "failed to write some discovery check error records"  # POS -> 1
  strings /app/agent-chassis | grep -cF "failed to write discovery check error record"        # NEG -> 0
  strings /app/agent-chassis | grep -cF "content_data envelope: failed to write record"'      # NEG -> 0
```
`f930de86b` reworded that log line singular→plural, so **both halves must hold in the same
binary** — no stale image and no lucky substring can satisfy both. Measured 1/0/0 on both
replicas of v1.0.1262 and again of v1.0.1263.

> **⚠ DO NOT use the count needle this file used to publish.** It said
> `strings … | grep -c "INSERT INTO agent_error_log"` must read **2**. A correct binary reads
> **1** — the two surviving sites hold byte-identical SQL and the Go linker **deduplicates
> identical string constants**, so two sites are one string. Converging the copies is what
> collapsed the count, so the needle reported success as failure. Full class in
> `LANDMINES.md`; `WRONG_CALLS.md` 2026-08-07.

---

## ⚠ FIRST THING — the council verdict is **REJECTED**, and the fix is cheap and named

Round 2 (submitted 08-07 08:29Z, verdict 08:39Z): `guardian` **hard veto**, `editquality`
object (medium), `reuse_agent` **approve**, `tooling_provenance` **approve**, 5 abstained.
Read it in full before acting:

```sql
SELECT left(body, 8000) FROM diagnosis_artifacts
WHERE correlation_id='5c2bc265-84ac-452b-bd8b-22fd7b875427' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
```

**The veto is about SCOPE VISIBILITY, not craftsmanship** — its own notes say the design, the
withdrawn NULLIF correction, the pod-verification and the disclosed dormant third writer are
*"genuinely careful work"*. It vetoed because: *"Guardian cannot assess blast radius on 26
files it cannot see."*

**Understand the catch-22 before you rewrite anything.** Round 1 was gated because my prose
described files **not** in the 8-edit array. I fixed that by declaring the array a
representative sample. Round 2's `editquality` seat **confirms the fix worked** — *"the array
now matches the prose"* — and the guardian vetoed **that same honesty**, because admitting 26
files were out of view is exactly the fact its mandate makes it veto on. **The 8-edit cap and
a 34-file change cannot both be satisfied inside one round.** Do not try to write your way
out; do not go back to hiding the gap.

**The way out is in the verdict's own `missing` field:** *"Full 34-file diff (or at minimum
**the list of all 27 non-test files**) was not supplied."* A file LIST is not an edit and
does not consume the cap. This is **not** the `bugs_open/124` case where a scope veto must
not be answered by resubmitting — that veto was about *how a capability reached production*.
This one names a supplyable artefact.

### The plan for round 3, in order

1. **SPLIT THE SUBMISSION.** Both `editquality` (medium) and `guardian` (low) say edit 8 —
   the (d) detector — is scope creep: *"not on the causal path … makes the round judge two
   different bugs at once."* They are right; it shares an RFC number and an owner sitting
   with B and nothing else. **Round 3a = B only (the conversions). Round 3b = (d) only.**
2. **In 3a, put the full 27-non-test-file list in `plan.summary`**, grouped by shape, and say
   the 8 edits are one exemplar of each group. Generate it, do not type it:
   ```bash
   git show --name-only --format="" 5f49b4cfd f930de86b | sed '/^$/d' | sort -u | grep -v _test.go
   ```
3. **Before 3b, answer `reuse_agent`:** `sharedoutputs.go`'s routing-graph walk was never
   checked against **`relaygaps.go` in the same package**, which also walks
   `agent_definitions` routing config. Extend it or state why not. **I did not look — this is
   an open question, not a settled one.**
4. **Before 3b, answer `tooling_provenance`:** leave a `doc_notes` PLAN/NOTES row for subject
   `cmd/config-key-audit`, carrying the ack-ratchet design and the 13-vs-11 routing-key
   correction, so the next lane does not re-derive them from source.
5. Resubmit with `RESUBMIT_CORR=5c2bc265-84ac-452b-bd8b-22fd7b875427` for 3a; **3b is a new
   submission with its own correlation**, since it is a different change.

**Before submitting anything, assert every `edits[].file` exists in the tree.** Round 1's
array named `cmd/config-key-audit/shared_output_fields.go`, **which is not a file** (it is
`sharedoutputs.go`). One loop over the array against the tree catches it.

**No trailer is owed or writable.** `f930de86b` carries `Council-Submitted:`, which asserts
nothing and is simply uncredited. **Never write `Council-Reviewed:` without an approved
verdict** — 098 buckets that as MISMATCH.

---

## NEXT — the remaining work, in priority order

### 1. The (d) online half — the CronJob. **The last undone owner ruling.**
Precedent to follow is **`component-render-check`**, NOT `single-owner-carriers-check`: ship
the Go binary as its own image rather than re-implementing in Python. RFC_006's Python mirror
exists only because a job cannot `go run` from source (262M clone); an image dissolves that,
and with it both of RFC_006's named drift risks (the `DECLARED_*` literal and the parity test).
- copy `deployments/kustomize/services/component-render-check/` as the shape;
- take a free schedule slot (taken: 02:00, Sun 06:00, 06:20, 06:40, 06:50, 06:55) and say
  which in the manifest;
- **report to `doc_notes`** — one row per run **including clean**, so a missing row means
  THE JOB DID NOT RUN, which must not be indistinguishable from "nothing is wrong";
- exit 1 on NEW findings so the Job shows failed;
- the ack file is in-repo, so it ships with the build.

### 2. The three silent Go breaks the census found — file them or fix them
Not part of any owner ruling; found while doing §3(a) and **they are a live latent defect
regardless of whether (a) ever ships**. `hero_deployed` / `logo_deployed` are read by
two-level direct map access with an `ok` guard:

| file:line | key |
|---|---|
| `platform/orchestration/actions/v3_site_actions.go:1010` | `hero_deployed` |
| `platform/orchestration/actions/v3_site_actions.go:1021` | `logo_deployed` |
| `platform/orchestration/actions/assemble_from_library.go:452` | `logo_deployed` |

Routing them through `datahelpers.ExtractNestedField` makes them wrapper-proof *and* fixes
them for the `call_agent` shape today. **They fail silently** — the page renders with no hero
and no logo and nothing records it — so this is worth a `bugs_open/` file if not a fix.

### 3. The last INSERT site — **still blocked**, re-check before assuming
`store_generated_component_action.go:1353`. Was still ` M` at 2026-08-07 with another
session's one-line `PageWantedLivePredicateFor` change at `:872` (`git diff -U0 <file> |
grep -c agent_error_log` → 0, so it does not touch the INSERT). **Status goes stale within
minutes — re-run it.** When clean, convert with `LogActionEntry` and set
`AgentType:"component-creator"`, `StepName:"store_component"`,
`Action:"store_generated_component"` **explicitly** — those three literals are exactly what
the guardian/edit-quality seats warned would be misfiled fleet-wide, and no test in the
package would catch the slip.

### 4. Not this lane's, but found and recorded
- **`search_results.results.0.url` can never resolve.** `vet-practice-verifier`'s
  `fallback_url_field` uses an array index, and `ExtractNestedField` does map access only —
  the walk aborts at `results`. That fallback has never fired. Belongs to the vet lane.
- **Dead config keys survive indefinitely.** `commit_from` is configured in 6 agents and read
  by nothing; 4 HITL `output_format` templates are never rendered. No drift check notices.

---

## What the census concluded (so you need not re-read it to make a decision)

`architecture_review/CENSUS_2026-08-07_rfc012_await_step_readers.md`. **It does not decide
(a)/(a′)** — it is the artefact they are gated behind.

- **138 of 221 live awaited steps already merge under `.response`** (the
  `call_agent`/`spawn_agent` branch always has). The at-risk set is the other 83.
- **Config side: 0 breaks.** 9 dotted-path readers survive because
  `ExtractNestedField` (`data_helpers.go:1199`) retries every unfound segment through
  `["response"]`; the other 10 are dead config.
- **Go side: 3 breaks, silent** (the table in §2 above).
- **64 whole-key handoffs are UNCLEAR** — and `ExtractFields` carries a **hardcoded**
  `.response` unwrap for one field (`unified_extractor.go:200`), which is evidence someone
  already hit this and patched one call site by hand rather than fixing the general path.
- **The evidence favours the ADDITIVE merge** (reply keys left at top level) over the
  `.response` sub-key: 0 config breaks, 0 Go breaks, residual risk reduced to key collision.
- **(a′) `storeActionResult` is NOT covered** — different write path, own reader set.

---

## Landmines this lane found (all in `LANDMINES.md`, synced to `doc_notes`)

- **A `strings <binary> | grep -c` counts distinct SPELLINGS, never SITES** — the linker
  dedupes identical literals, so a de-duplication refactor moves the number you are using to
  prove it shipped. Use a discriminating pair derived from the commit's own diff.
- **`LogActionEntry`'s merge fills a provenance you meant to set, and the whole suite stays
  green** — the package's tests pin codes and messages, never `agent_type`.
- **A nil `Context` map marshals to `null`, not `{}`** — pin `"null"` in tests.
- **The RFC's own 13-key list is one short** — trust the live query, not the prose.
- **A green live run of the census proves the FLEET, never the DETECTOR** — the mutation test
  (sever the `config.then_step` edge, lose the finding) is what proves the detector.
- **Always run the concept-index uniqueness check, not just the row count** — my first
  `WFA-007` collided with a live relay-gaps entry; a count cannot see a collision.
- **A census of a BEHAVIOUR must enumerate every way the behaviour is EXPRESSED** (NOTES
  misstep 10). Grepping the map-literal `"await_response": true` returns 24 awaiting actions;
  the true answer is 40, because adapter actions signal it through a **typed result struct**
  (`AwaitResponse: true` with a json tag) — so the one-syntax grep dropped *every* adapter
  dispatch.

---

## Milestone read-outs
`SUMMARY_2026-08-06_two_of_three_built.md` — the series' first.
**A second is now genuinely owed** and was not written for lack of session budget: the
read-out has changed materially since ("built" → "shipped, proven live, census delivered,
council rejected on scope-visibility"). Write it as `SUMMARY_2026-08-08_*.md` — a **new
file**, never an edit of the first. The five headings would now answer differently, which is
the test the standing-five rules set.
