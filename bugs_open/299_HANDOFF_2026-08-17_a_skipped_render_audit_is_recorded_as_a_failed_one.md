# 299 — a skipped render audit is recorded as a FAILED one: the drain has no case for its own upstream's honest no-op

> **⚠ NUMBER COLLISION — `299` names TWO unrelated cases.** A concurrent lane filed
> `299_HANDOFF_2026-08-17_home_page_cta_names_the_brief_starter_tool_and_dials_the_phone_instead.md`
> at 17:19 (this file was committed ~13:00). Neither is renumbered — per CLAUDE.md
> numbering is never reassigned. **Resolve by SLUG, and `git log` the FILE PATH.**
>
> **STATUS 2026-08-17 (evening) — fix COMMITTED (`89b3e582b`), council APPROVED round 1
> (`Council-Reviewed: eaa043d7-867f-4d40-a0d9-c41b41e56cf9`, verdict and all six
> advisories READ and answered in §8) — and NOT LIVE.** A chassis roll landed at ~14:43Z
> and **shipped none of this**: the deployed tag is still `v1.0.1305` (unchanged, fleet-wide)
> because `IMAGE_TAG` was not bumped, so the same-tag rebuild served the node's cached
> image. **Proven at the artefact, both replicas, with controls** — see §7b. Independently
> confirmed the same hour by the `295` lane (`dcf3dc7d6`, `b9942d449`): same cause, a
> different fix, also unshipped. **So the close criterion below is blocked on a
> TAG-BUMPED build, not on the next roll.** Lane:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_242_render_audit_truncation/`.

**Filed** 2026-08-17, found while re-verifying `bugs_closed/242`'s close criterion (this
file's evidence is what that verification turned up on the live DB). **Severity** low —
the damage is a false failure record per occurrence, not lost work — but it fires on a
state every new site passes through, and the record it writes is exactly the kind the
estate's own rule forbids: a run that measured nothing, stamped in a way a reader cannot
distinguish from a run that *broke*.

**Scope note:** everything here is first-hand and disconfirmable (queries and file:line
given). No cross-cutting mechanism is asserted — see §6 for the `090` substitution
statement.

## 1. What I measured (2026-08-17)

Rotation fire 12:10:14Z, orchestration `dc0233ab-d5fb-490f-8f8a-6587738e2f3b`, site
`remortgagecalculator.uk` — created **the same morning** (11:14:54Z), zero deployed pages.

| fact | value | how |
|---|---|---|
| `collected_data->'render_audit'` | `{"domain": "remortgagecalculator.uk", "reason": "no_deployed_pages", "skipped": true}` | keys **enumerated** with `jsonb_object_keys`, then read |
| `collected_data->'__step_error'` | `step write_findings failed: … write_render_audit_findings: "render_audit" carries neither a result nor a .response — the audit is still awaited or failed` | same row |
| run status | `COMPLETED` — via the `complete_error` edge (its message: "Render audit did not complete… Nothing was measured.") | `orchestration_states` |
| `agent_error_log` residue | 1 row, `severity=error`, `error_code=UNKNOWN`, step `write_findings`, joined to the run | `WHERE orchestration_id LIKE 'dc0233ab%'` |
| sites currently able to trigger this | **1** (`remortgagecalculator.uk`) | `sites` × `pages` group-by, `status IN ('active','deployed')` HAVING zero `build_status='deployed'` pages |
| occurrences ever | **2** — 2026-08-11 08:02Z and this run | `SELECT count(*) … FROM agent_error_log WHERE error_message LIKE '%carries neither a result nor%'` |

The 2026-08-11 occurrence is the run the `bugfix_242` lane used as its *persistence
control* ("the eighth was a no-await `skipped` run and keeps its full result") — the
persistence observation was correct, and the same run's `write_findings` failure went
unremarked. Both occurrences postdate the weekly rotation (migration `369`, VIZ-015);
before it, nothing ever pointed the audit at a page-less site.

## 2. The mechanism, first-hand

Producer — `platform/orchestration/actions/request_render_audit_action.go:146-156`
(present since `def22cd4d`, the action's first commit). When the site has zero deployed
pages it deliberately returns a **no-op, not a failure**, and says so in the code:

> `// No await, and NOT a failure: a site with nothing deployed has nothing`
> `// to measure. Declaring it clean would be the lie.`
> `return map[string]interface{}{"skipped": true, "reason": "no_deployed_pages", …}`

Consumer — `platform/orchestration/actions/write_render_audit_findings_action.go`,
`extractRenderAuditPayload` (`:607-631`, present since `f2a222964`, the drain's first
commit). It accepts exactly two shapes: a direct result (has `contrast`) or an awaited
envelope (has `response`). The skip shape has neither → error at `:619`, the exact
string in §1's `__step_error`.

Workflow — the live `render-audit-agent` row is linear: `audit` → `next_step:
write_findings`, no conditional edge. `write_findings`' `error_step: complete_error`
stamps *"this is NOT a clean result — a failed audit and a clean audit must never be
read the same way."*

So the two halves were **born incompatible in the same build** (A0.3 / VIZ-012+013) and
the edge was simply never exercised until the rotation started sweeping every site
unattended. The producer's comment and the drain's own file-header claim ("Fails loud
when render_audit is absent or awaited") are both individually right; the state neither
of them owns is *skipped*.

## 3. Why it matters (and why it is only low severity)

- The estate's rule — a failed audit and a clean audit must never be read the same way —
  has a **third state**, *skipped*, and the current behaviour collapses it into
  *failed*. The weekly record for the site reads "the audit errored, nothing was
  measured" when the truth is "there was nothing to measure". Both are non-clean, but
  one asks an operator to investigate and the other does not.
- One `agent_error_log` row at `severity=error` per occurrence — the immune system
  sweeps recorded failures, so each is a candidate for spurious triage. (No
  `site_work_items` row has resulted yet — checked, zero open render-audit items.)
- The trigger state is not exotic: the rotation's `pre_query` selects any
  `active`/`deployed` site with a >7-day-old stamp and no *claimed* build work item —
  a site created an hour earlier qualifies immediately, so **every new site passes
  through the window between creation and first page deploy**. Today that window
  produced a failure record one hour after site creation.
- Cost per occurrence is otherwise negligible: the run completes, the stamp is written,
  the rotation moves on. Nothing retries, nothing accumulates except the false record.

## 4. Fix candidates, ordered by what closes the door

1. **Teach the drain its own upstream's vocabulary** (the fix): in
   `WriteRenderAuditFindingsAction`, recognise `skipped: true` at the top level of the
   audit field and return an honest no-op result — `{"skipped": true, "reason": …,
   "inserted": 0, "deduped": 0}` — so the workflow completes via its NORMAL edge and
   `findings_written` says exactly what happened. Crucially it must **not run
   retraction**: standing findings may only be closed by a run that re-measured their
   pages, and a skip measured nothing. A skip result cannot be mistaken for a
   measured-clean sweep (those carry no `skipped` key), so the 242 honesty property is
   preserved, not weakened.
2. Conditional workflow edge (audit-skip → complete directly). **Rejected**: the
   workflow schema has no conditional `next_step`; inventing one is a shared-seam
   change orders of magnitude beyond this defect.
3. Exclude page-less sites in the rotation `pre_query`. **Rejected even as
   mitigation**: it duplicates the deployed-pages predicate into SQL config (the
   lockstep-drift class — cf. the dedup-index/Go-list landmine), inherits
   `bugs_open/185`'s predicate gap, and deletes the honest weekly "nothing to measure"
   record for the price of a ~0.3s no-op run.

**Do NOT** "fix" the guard by relaxing it generally: the `has not run` and
`awaited or failed` errors are load-bearing (they are what makes an absent or
still-parked audit unable to read as a clean write — tests pin both).

## 5. `[UNMEASURED]` — the class, noted and not asserted

`"skipped": true` no-op returns are an estate idiom (a dozen producers in
`platform/orchestration/actions/` alone: tool_acceptance, rerender_single_page,
rebuild_blog_listing, …). Whether every consumer downstream of each one tolerates the
shape is a per-workflow question **this file does not answer and does not assert either
way**. If someone audits that class, this bug is one worked example of the failure
shape: a linear workflow whose consumer predates/ignores its producer's no-op contract.

## 6. Why `090` was not run (OWNER RULING 2026-07-31 escape hatch, stated)

The root cause is local and self-evidencing: one function
(`extractRenderAuditPayload`) missing one case of its direct upstream's declared
output. Verified first-hand by: reading both functions in full (not grepped), reading
the live workflow config from `agent_definitions`, two live failure rows carrying the
exact error string, and `git log -S` dating both halves to their files' first commits.
No claim in this file reaches beyond that edge; §5's class question is explicitly left
open rather than asserted.

## 7. Verify a fix

Unit: the skip shape returns `{skipped: true, reason, inserted: 0}` with **zero**
queries executed; the `has not run` and `awaited or failed` guards still error (the
skip check must not swallow the await-signal shape `{"success": true, "request_id": …}`).

Live (post-roll close criterion): dispatch `render-audit-agent` at a site with zero
deployed pages (remortgagecalculator.uk while it still qualifies, else force one) —
expect `COMPLETED` via the **normal** edge, `findings_written` = the skip no-op, **no**
`__step_error`, **no** new `agent_error_log` row. The rotation stamp from 2026-08-17
means the organic re-fire is ≥7 days out; use the manual dispatch recipe
(`bugfix_242_render_audit_truncation/RUNBOOK` → RUNBOOK_oufe §14).

## Related

- `bugs_closed/242` — the sibling honesty defect in the same pipeline, opposite
  direction: 242 was a *partial* result reading as *complete*; this is a *deliberate
  no-op* reading as a *failure*. Found while verifying its close criterion.
- `bugs_open/185` — why the rotation selector must not grow its own deployed-pages
  predicate (§4 candidate 3).
- `bugs_open/296` — parked contrast findings and retraction; different defect, same
  drain. §4's "no retraction on skip" is written so a fix for either cannot regress
  the other.

## 7b. The 2026-08-17 roll did NOT ship this fix — measured, not inferred

The chassis pods restarted ~14:43Z, **after** the fix commit (12:53Z), and the fix is
still absent. The binary is the authority: the startup `build provenance` line had
already rotated out of reach (`--limit-bytes=400000` from the log start reaches only
16:08Z — "not in range", not "unstamped", exactly as `LANDMINES.md` warns).

Probe, run on **both** replicas, each with two controls in the same breath:

```bash
POD=agent-chassis-5bd56bdd9b-6sb8t   # and -jzmns
kubectl -n ai-persona-system exec $POD -- grep -ac "audit was skipped upstream" /proc/1/exe
kubectl -n ai-persona-system exec $POD -- grep -ac "carries neither a result nor a .response" /proc/1/exe
kubectl -n ai-persona-system exec $POD -- grep -ac "zzz_this_string_is_not_in_any_binary_299" /proc/1/exe
```

| probe | expectation | 6sb8t | jzmns |
|---|---|---|---|
| this fix's new literal | present if shipped | **0** | **0** |
| positive control (guard literal, in old AND new) | must be present | 1 | 1 |
| negative control (nonsense string) | must be absent | 0 | — |

So the instrument discriminates (control present, negative absent) and the answer is
**the fix is not in the running binary**. Cause: `IMAGE_TAG` in the makefile still reads
`v1.0.1305` and every deployment reads `v1.0.1305` — a same-tag rebuild ships the node's
cached binary (CLAUDE.md, "Bump `IMAGE_TAG` for every build"). **A roll is not a build,
and a fresh-looking pod is not new code.**

## 8. Council advisories (eaa043d7) — answered with measurements, not argument

Verdict: **approved**, 12 reviewers, 5 abstained, 0 unreadable, not truncation-gated.
Three seats objected at medium, two at low. Answers:

- **bug_historian (medium) — "the plan does not audit for siblings."** Measured: the
  exact strict shape (*error* when the field carries neither a result nor a `.response`)
  exists in **exactly one file fleet-wide** — this one (`grep -rln "neither a result
  nor" platform/ internal/` → 1 hit). The named sibling `write_audit_findings_action.go`
  is the **opposite** design: it falls back through three alternative field paths and, on
  a miss, returns `{items_created: 0, reason: "no findings in …"}` with a `Warn` — it
  cannot turn a skip into a failure. **So the exposure does not replicate in this
  shape.** `[UNMEASURED]` and deliberately left so: whether some *other* consumer is
  strict in a different way, and whether `write_audit_findings`' permissiveness is the
  242-class risk in the other direction (a silently-empty write where a loud failure
  might be right). That is that drain's owner's call, not this bug's.
- **reuse_agent (medium) — "did you look for an existing shared skip helper?"** Looked:
  `grep -rn "skipped" platform/orchestration/datahelpers/*.go` returns six hits, all
  unrelated prose (voicetells, page_identity, claims, runtime_fill, unified_extractor).
  **There is no shared skip-shape helper to extend**, so the inline check duplicates
  nothing. The divergence the seat suspected between the two drains is real and is
  recorded in the bullet above.
- **guidelines (medium) — a nested-field addition must be named in the concept register
  in the SAME commit (2026-08-11 ruling).** **Partly missed, stated rather than
  glossed:** the register *was* updated — VIZ-013 now names the `skipped` key and its
  producer — but in `4fe7cc519`, a docs commit ~15 minutes after the code commit
  `89b3e582b`, not in it. The ruling asks for the same commit. Logged in `WRONG_CALLS.md`.
- **guardian (low) — "`ExpectationsWereMet()` with no expectations passes vacuously."**
  Correct about that line, and the test does not rest on it. **Mutation-proven**:
  inserting `params.DB.QueryContext(ctx, "SELECT 1")` before the skip return fails the
  test — `a skipped audit must not error: stray query: … call to Query 'SELECT 1' … was
  not expected` — because sqlmock errors on an unexpected query and
  `loadLockedComponentHTML` propagates rather than swallows. The load-bearing assertion
  is the `err` check; `0c65dc131` writes that into the test so the next reader does not
  trust the vacuous line.
- **guardian (low) — "confirm this action has exactly one caller."** Query, not
  assertion: `SELECT type … FROM agent_definitions, jsonb_each(default_config->'workflow'
  ->'steps') s WHERE … s.value->>'action' = 'write_render_audit_findings'` →
  **`render-audit-agent`, 1 step. Single consumer.**
- **debug_historian (low) — "no step verifies the fix reaches the running pod; same-tag
  rebuilds ship stale binaries."** The seat called it before it happened. Done in §7b —
  and it is exactly what the probe caught. This advisory is now the close-out step.

## 9. CLOSED 2026-08-17 — LIVE on `v1.0.1307` and PROVEN by a forced run

**Deploy proven at the artefact, per service, both replicas, with controls.** Tag bumped to
`v1.0.1307` (makefile and every deployment agree); pods started ~17:05Z. Binary probe of
`/proc/1/exe` on `agent-chassis-6d6d7b9996-9z8rp` **and** `-z7rj2`:

| probe | expectation | 9z8rp | z7rj2 |
|---|---|---|---|
| this fix's literal (`audit was skipped upstream`) | present if shipped | **1** | **1** |
| positive control (the old guard literal) | must be present | 1 | 1 |
| negative control (nonsense string) | must be absent | 0 | 0 |

**The forced run** (the organic trigger was gone — no site had zero deployed pages any
more, `remortgagecalculator.uk` having deployed in the interim). Forced **without creating
or mutating anything**: `pool-web-tech.internal` (`8f02310c…`) is an existing internal pool
row with zero pages, so `request_render_audit` takes its no-deployed-pages branch exactly as
the rotation did this morning. Manual dispatch on `system.agent.scheduled.requests`
mirroring the known-good 17:12 rotation payload shape (`PUBLISH_OK` seen; corr
`c920fffe…`, orch `132b38b9-254e-49d0-86a3-a21d295eb0ac`, COMPLETED in ~90s).

All four criteria from §7 held:

1. **`current_step = complete`** — the NORMAL edge, not `complete_error`.
2. **`findings_written = {"skipped": true, "reason": "no_deployed_pages", "inserted": 0,
   "deduped": 0}`** — the honest no-op, carrying the upstream's reason verbatim.
3. **No `__step_error` and no `complete_error`** (`collected_data ? …` → both false).
4. **Zero new `agent_error_log` rows** for the run; the fleet-wide count of the guard error
   is still **2** — the same two pre-fix occurrences, unmoved.

**Why this grades the fix rather than the site:** the disconfirming result was available and
is on record. The *same* input shape against the *pre-fix* binary produced the opposite
outcome nine hours earlier — orchestration `dc0233ab…` (§1): `__step_error`, an error-log
row, and the `complete_error` edge. Same case, different binary, opposite result.

**Also confirmed post-roll, for `bugs_closed/242`:** the 17:12 organic rotation run on
`cookly.uk` (`2bb6873a…`) shows `pages_total: 5` alongside `pages: 5` in its summary and a
full `findings_written` — 242's honesty fields are live in an unforced production run, not
only in the forced 08-11 test.

**Residue:** one COMPLETED orchestration row against an internal pool site, no work items,
no retraction (the skip returns before any DB work), no rotation stamp touched.
