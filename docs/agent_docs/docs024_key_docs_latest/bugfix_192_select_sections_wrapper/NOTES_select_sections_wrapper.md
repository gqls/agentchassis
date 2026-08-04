# NOTES — `bugs_open/192`, the section_plan wrapper

Append-only, newest at the bottom.

---

## 2026-08-04 ~09:35 — picking the bug, and the two candidates I put down

Swept `bugs_open/` against live session transcripts rather than `who-owns.py` alone
(that script reads COMMITS, so a session mid-fix is invisible to it — the standing
lesson). 30-odd sessions are live on this tree; nearly every recent bug is claimed.

Two candidates I opened and deliberately put down, recorded because the reasons are
reusable:

- **`132`** (every B2 site serves a raw JSON blob instead of a 404) — genuinely
  unowned, still reproducible (I re-curled webdesign.co.uk, robot-hands.com,
  oufe.com today: all still raw JSON; relojistas.com still the nginx control). Put
  down because its fix is a **Cloudflare-side edit with no deploy path from this
  tree** — no `wrangler.toml`, and the repo's `scripts/cloudflare/worker.js` is
  provably NOT what is deployed (its miss branch returns plain "Not found", the edge
  returns JSON). I could have written a fix I could neither ship nor verify.
- **`114`** (imagery generated, deployed, never referenced) — unowned, but its
  load-bearing candidate is "drain or stop filing `image_landed`", and those rows
  are `item_type='needs_page'`, which is exactly the seam the `187` lane fixed and
  committed **today**. Taking it would have been competing, not contributing.

Took **`192`** instead: filed 30 minutes earlier by the `154` lane while
live-verifying `178`, explicitly *"not yet diagnosed"* and handed off. Confirmed
only the filing lane and mine had it in transcript.

## 2026-08-04 ~09:40 — filed the `090` run before asserting anything

Per the CLAUDE.md default (durable, cross-cutting claim ⇒ file first). Intake
`b45144fa-2051-4ff8-9318-e052a9c3a084`, run `aea3cc68-b274-4df4-a1c1-f60ba47bf09e`.
Checked the diagnosis queue first — no open `needs_diagnosis` touched this. Then
diagnosed first-hand in parallel rather than waiting the ~30 minutes.

## 2026-08-04 ~09:45 — the diagnosis, and the measurement that turned it

The bug file's own open question was whether the fallback fails for an **ordering**
reason (the step running before `input_data.section_plan` is merged). It is not
ordering. The step that decided it was enumerating the KEYS rather than reading the
path:

```sql
SELECT (SELECT string_agg(k, ',' ORDER BY k)
        FROM jsonb_object_keys(collected_data->'input_data'->'section_plan') k) AS keys,
       jsonb_array_length(coalesce(collected_data->'input_data'->'section_plan'->'sections_ready','[]')) AS direct,
       jsonb_array_length(coalesce(collected_data->'input_data'->'section_plan'->'section_plan'->'sections_ready','[]')) AS nested
FROM orchestration_states WHERE owner_agent_type='page-content-writer' ORDER BY updated_at DESC;
```

```
3b692317 | applied,reason,section_plan              | direct 0 | nested 7   <- FAILING
0883b1aa | applied,matched,section_plan             | direct 0 | nested 1   <- FAILING
df69efd6 | applied,matched,section_plan             | direct 0 | nested 1   <- FAILING
4edcdaeb | deferred_count,…,sections_ready          | direct 3 | nested 0   <- flat, pre-seed
```

Two shapes. The failing ones carry a **wrapper**; the data was never missing, it was
one level down. `sections_for_render` on all three is `{}`.

The producer is `load_current_section_content_action.go` (`178`'s fix, `08d0515f3`),
wired by seed `299` with `output_field: section_plan` — and `coordinator.go:1859-61`
(`storeActionResult`) stores an action's return value **wholesale** under
`output_field`. The action returns `{section_plan, applied, reason|matched}` on
**every** path including its eight "pass-throughs", so the flat plan is replaced by
the wrapper on every build in every mode.

**The part I nearly missed:** this kills BOTH of `select_sections`' fallback paths
from one cause. Path 2 directly; path 1 because `resolve_links`' input_mapping is
`"sections?": "input_data.section_plan.sections_ready"` — the resolver child is
handed no sections and returns `sections_ready: null`, which is the "present but
explicitly null" value the bug file reported as if it were an independent second
fault. I would have written that up as two problems. It is one.

## 2026-08-04 ~09:50 — MISSTEP AVOIDED, and the one I nearly shipped

I had drafted the seed to fix `resolve_links`' input_mapping too, pointing it at the
nested path so link resolution worked again immediately. **That would have been a
silent time-bomb.** `input_mapping` has no ordered fallback: the moment the Go fix
rolls and the plan is flat again, a nested-path mapping resolves to nothing, the
resolver silently gets no sections, and there is no error to notice — the exact
shape of defect I am here to fix. Dropped it. `select_sections`' shim is safe only
because its paths are tried **in order** and the flat path sits ahead of the shim, so
the shim retires itself on the roll. Recorded in the seed header as a deliberate
non-change.

## 2026-08-04 ~09:55 — the filing's onset is wrong, and it pointed away from the cause

`192` says "failing broadly since ~2026-08-03 21:00" and cites hourly FAILED spikes
of 11/14/12. Re-measured over the full retained window, splitting on the step name:

```
08-04 08:00 | select_sections_miss  | FAILED |  3
08-04 01:00 | iter_generate_content | FAILED |  1
08-03 23:00 | iter_generate_content | FAILED | 12
08-03 22:00 | iter_generate_content | FAILED | 14
08-03 21:00 | iter_generate_content | FAILED | 11   (+ 12 COMPLETED the same hour)
```

The overnight spike is `process_sections_loop_iter_N_generate_content` — a step only
reachable **after** `select_sections` has already succeeded. It is a different,
still-undiagnosed defect. This bug's mode (`current_step='process_sections_loop'`
exactly) starts at **08:20 on 08-04**, which is `agent_definitions.updated_at` for
both agents — i.e. when seed `299` was applied.

This is not pedantry: the wrong onset is what made "178 cannot be the cause" look
sound, because 08-03 21:00 predates the trigger. Logged in `WRONG_CALLS.md`, and
corrected in place at the head of the bug file.

## 2026-08-04 ~10:05 — the test that passed on the broken code

`load_current_section_content_action_test.go` asserted `result["applied"]`,
`result["reason"]`, `result["section_plan"]` — i.e. it encoded the wrapper as the
contract, and so **passed on the code that broke every page build in the fleet**.
Its own header, two lines above, says the step "must leave section_plan
byte-identical". The test and its comment disagreed and the comment was right.

Rewrote it to assert the contract instead of the observed shape, and added
`TestLoadCurrentSectionContent_NeverWraps` walking all four reachable return paths.

**Then proved it could fail**, because a passing test proves nothing on its own:
mutated the action back to the wrapper form, re-ran, confirmed **all five** cases
fail with the intended message, restored, and diffed to confirm the restore was
clean. (Short window on a shared tree — copy first, restore in the same command.)

```
--- FAIL: TestLoadCurrentSectionContent_NoMode_Passthrough
    pass-through must return section_plan unchanged.
    bugs_open/192 regression: the plan is nested under its own key …
--- FAIL: …ModeRecreate_Passthrough / …EditLive_AttachesMatchingSlot /
    …NoReadySections_SkipsQuery / …NeverWraps (all 4 sub-cases)
```

## 2026-08-04 ~10:15 — seed 308 applied, deliberately before its image

Inverts the house "image first, then seeds" rule, and the header says why: that rule
exists because a seed naming an **unregistered action** fails at runtime, and this
seed names no action. The third fallback path is plain data; `required` is a config
key the running binary provably ignores (`ExtractFieldsAction` reads only `fields`,
`field_map`, `defaults`). So it restores builds on the binary already deployed.

Applied clean: `UPDATE 1`, both `DO` verification blocks passed, `COMMIT`. Wrote the
verification as `DO`/`RAISE` rather than bare `SELECT`s — a `SELECT` below an
`UPDATE` cannot stop a `COMMIT` under `ON_ERROR_STOP`, which is a landmine already on
file.

## 2026-08-04 ~10:25 — V2 PASSED end to end, and it could have failed

Re-queued `18bc832c` (vetcomparison `content_rewrite`) — the item `192`'s filer
explicitly nominated for this ("both parked items … can retry once this is fixed").
The dispatcher takes `status IN ('triaged','approved')`, so `failed` → `triaged`.

```
[1] 0511e4d1 | EXECUTING_STEP | spawn_link_resolver                           | -    | -
[2] 0511e4d1 | EXECUTING_STEP | process_sections_loop_iter_0_generate_content | true | 1
[3] 0511e4d1 | COMPLETED      | complete                                      | true | 1
```

Work item `18bc832c` → **`complete`**. The columns are
`sections_for_render ? 'sections_ready'` and its length: **true / 1**, against **false**
on the three runs immediately before. This is falsifiable — it fails at
`process_sections_loop` if the wrapper diagnosis is wrong.

**A near-miss worth recording.** `0733d0cc` FAILED at 09:02:08, which at first looked
like a counter-example *after* my seed. It was created at **09:01:33**, and the seed
committed at **09:01:35** — two seconds earlier. It had already loaded the old config.
Checking `created_at` rather than `updated_at` is what settled it; `updated_at` on a
failed run is when it died, not when it started, and reading the wrong one would have
had me chasing a fix that had actually worked.

## 2026-08-04 ~10:30 — the `090` run returned NOTHING, and that is worth stating plainly

The diagnosis loop I filed at the start completed, but produced **no verdict**:

```
kind   | count      diagnosis_artifacts for aea3cc68-…
bundle |     3      (08:42:31 · 08:45:01 · 08:48:28; truncated = false/true/false)
final_result: empty          work item: complete, spec carries no verdict
```

Three bundle rounds, one of them truncated, then the run ended with no
diagnosis/verdict artifact and no `final_result`. So it neither confirmed nor refuted
anything. **This lane's diagnosis therefore rests entirely on first-hand
verification** — reading the code, both shapes in live rows, the fleet census, and the
end-to-end re-dispatch above — which the owner ruling of 2026-07-31 permits provided
the substitution is **stated, not silently omitted**. It is stated here, in the bug
file, and in the commit.

Same failure mode `bugs_open/161` recorded ("UNVERIFIABLE at the iteration cap … blocked
by a harness truncation"), so this is a second instance rather than a one-off. Not
diagnosed here and **not** filed as a new bug by this lane — but noted, because "I ran
the loop" is worth nothing if nobody checks that the loop produced anything, and it
would be easy to cite the run id and imply corroboration it never gave.

## 2026-08-04 ~10:40 — state at handover

- **Committed:** `2b9d84072`, `Council-Submitted: 7afbf531-5ddd-484e-88c8-091994a0f51f`
  (verdict not yet read — trailer is deliberately `Submitted`, never `Reviewed`).
  The trailer gate refused an earlier attempt that carried `Council-Submitted: pending`,
  correctly: a non-UUID resolves to nothing and forward-only forbids the amend.
- **Live now:** seed `308` only. The outage is over and proven.
- **Inert until the next roll:** the unwrap, the `required` opt-in, the loop message.
- **Owed after the roll:** a cleanup seed removing the shim path (path 3 of
  `select_sections`), and the WFA-009 `verify-later` pod-greps.
- **Deliberately left alone:** `resolve_links`' `input_mapping`, so internal CTA
  resolution is degraded on every build until the roll. Stated in the seed header;
  fixing it would have re-broken silently on the roll.
- **Not this bug, still open, nobody on it:** the overnight
  `process_sections_loop_iter_N_generate_content` failures (21:00–01:00, ~38 runs).
  `192`'s filing counted them as this bug; they are not.

## 2026-08-04 ~10:55 — second proof, because n=1 is not a rate

The first pass left the post-seed sample at **one** successful run, which is a fact and
not a rate. Re-queued the second item `192`'s filer nominated (`9e9ec430`, vetcomparison
`guide-cma-compliance`). Every `page-content-writer` run created since the seed:

```
orch     | status    | current_step          | has_sections | created_at
25652dd0 | COMPLETED | complete              | t            | 09:21:02   <- 2nd re-dispatch
0511e4d1 | COMPLETED | complete              | t            | 09:03:47   <- 1st re-dispatch
0733d0cc | FAILED    | process_sections_loop | f            | 09:01:33   <- started 2s BEFORE the seed
```

**2 of 2 post-seed runs COMPLETED, 0 failures**, both with
`sections_for_render ? 'sections_ready'` true; the single failure in the window
pre-dates the seed by two seconds on `created_at`. Still a small sample — traffic is
low — so the honest claim is "the mechanism is proven on both items that were parked on
it", not "the fleet failure rate is now zero".

## 2026-08-04 ~11:20 — council REVISE, and the gating objection was RIGHT (there is a second live instance)

Verdict on `7afbf531`: **REVISE**, 12 reviewers, 5 abstained, 0 unreadable, not
truncation-gated. Decided by a **high**-severity objection from `bug_historian`.

**The gating objection.** My edit 1 fixes the one action that tripped over
`storeActionResult`'s wholesale write under `output_field` and documents the trap in
prose, but leaves the *generic mechanism* unguarded — "any other action that reuses an
upstream output_field name has the same silent-replacement exposure". It also flagged
that a related instance of this exact `storeActionResult` class is **already** a landmine
here, so this is a recurring class, not a one-off. And under `missing`: **no fleet-wide
audit of steps whose `output_field` reuses an upstream key was proposed or run.**

**I ran it, and it found a second live instance.** Not a hypothetical:

```
24 (agent, output_field) pairs shared by 2-5 steps.
Most are MUTUALLY EXCLUSIVE BRANCHES — mark_complete/mark_failed,
notify_scheduler/notify_scheduler_idle, 4x finalize_*_result, 5x store_*_asset.
Only one runs per execution; no overwrite, no hazard.
SEQUENTIAL refiners — step B rewrites step A's key — are exactly TWO:
  page-build-handler / section_plan      plan_sections -> load_current_section_content   (= 192)
  site-adoption-agent / design_fingerprint  extract_fingerprint -> enrich_fingerprint    (NEW)
```

`enrich_fingerprint_with_css_action.go`: both success paths correctly return `fp`, but
both **early-outs** returned `{"status":"no_fingerprint"}` / `{"status":"invalid_fingerprint"}`
— and with `output_field: design_fingerprint` the second **overwrites a real fingerprint
with a status stub**. Fixed the same way (return the caller's value unchanged; reason to
the log). **Nobody had reported a failure from it. The census found it, not a symptom** —
which is the argument for making the census standing rather than one-off.

**The count alone would have been alarmist.** 24 → 2 only because branches are excluded,
and that distinction is the whole difficulty. Worth remembering before quoting a census.

**Routed the architectural half to `RFC_012`, not a new RFC.** 012 is the *same two lines*
from the other side (`applyResponseToState` replacing the record when the awaited reply
lands) and is OPEN needing a human. Filing RFC 013 would have forked one mechanism into two
accounts. Added: the second face, the census 012's own §3(a) asked for and nobody had run,
and two new questions — should `storeActionResult` warn on an `output_field` collision
(cheaper here, because the collision is statically knowable), and should the shared-
`output_field` census become a standing `config-key-audit` mode alongside `SingleOwner`.
Also recorded why I did **not** just ship a warn: it would fire on all 24 pairs today,
including the 22 legitimate branches, converting a real signal into ignorable noise.

**The medium objection — `research-agent.extract_topic` left exposed — answered with a
measurement, not a shrug.** The seat is right that "one call site guarded, sibling
heuristic" is a known bad shape. But opting it in blind would be the unmeasured change
this same council objects to elsewhere, and the discriminator is real: `select_sections`'
missing field is **fatal-but-silent** (the run dies two steps later either way), while
`extract_topic`'s looks **genuinely degrading** (no `defaults`, and research continues on
a weaker query). Tried to measure which:

```sql
SELECT count(*) FROM orchestration_states WHERE owner_agent_type='research-agent';  -- 0
```

**Zero runs in the retained window**, so there is no evidence either way and I will not
opt in a path I cannot observe, on another lane's agent, to make it fail where it
currently degrades. Stated as an open item rather than closed.

**Two low-severity objections converted from assertion to attached check** (`editquality`
was right that they rested on claims, not checks):
- *nothing iterates the plan's keys* — `DisallowUnknownFields`: **0 occurrences
  platform-wide**; no `range` over the section-plan map (the one `for k := range plan` hit
  is `ValidateSitePlanAction` walking a **site** plan on an error path); consumers read
  `sections_ready` by name. `edit_live_meta` appears **0** times outside its own file.
- *exactly 2 `extract_fields` steps* — re-run and shown above, with `has_required` per row.

`reuse_agent`'s soft point (is `required` a second `check_required_fields_missing.go`?):
**no, different layers.** That is a *discovery check* over `content_data` of deployed
component instances — an audit sweep, flag-only, emitting `needs_human_review`. Mine is a
*runtime step contract* over orchestration `collected_data` that fails the step. Different
data, different time, different consequence.

## 2026-08-04 ~12:10 — council round 2: APPROVED, and the five advisories worked through

`7afbf531`, round 2: **approved with 5 advisory objections, none high-severity**,
4 abstained, 0 unreadable, not truncation-gated. The `bug_historian` seat explicitly
credited the round-1 response: *"it ran the fleet census the round-1 objection demanded and
found a real second casualty (design_fingerprint), which it fixed. That is the correct
shape."* Trailer earned: `Council-Reviewed: 7afbf531-…`, on a verdict actually read.

Four advisories acted on, one recorded as a human decision:

**1. `editquality` (medium) — seed 308 was applied by hand and the NUMBER was never
claimed.** Correct, and it is the standing landmine ("a migration number is not yours
because you named a file — it is yours when the LEDGER says so"). I had picked 308 with
`ls | tail`, which answers *what existed a moment ago*, never *what is free*. Asked the
ledger instead:

```
schema_migrations, highest recorded: 302_load_existing_pages_has_shipped.sql
303-307 exist on disk, recorded: NONE (other lanes' pending files)
```

So 308 collided with no *claimed* number — but mine was pending, and the next `--apply`
would have replayed it. Recorded it the same minute with the by-hand provenance in the
note, which is the only place that survives:
`./scripts/migration/run-migrations.sh --record-only 308_… --note '<applied at, verified how>'`

**2. `reuse_agent` (missing) — no `code_checks` on the new config key.** The code index is
stale (known landmine), so `code_checks` would return NEEDS_HUMAN_REVIEW. Ran the offline
audit instead, which is the tool that actually answers it:
`./scripts/audit-config-keys.sh` → `extract_fields` is an **UNDECLARED ACTION** (no
`RegisterActionInputSpec`), so nothing knows its key set and `required` cannot be flagged
as unknown. The three UNKNOWN KEYS it does report (`plan_sections: domain` et al.) are
pre-existing and belong to `bugs_open/136`. **No breakage from the new key.**

**3. `editquality` (low) — `required` was not cross-checked against `fields`.** Fixed: a
name in `required` matching no configured target now fails as a *step-config error*,
naming what WAS configured. Without it, a typo fails **every** run with a message about a
field the reader cannot find in `fields` — maximally confusing, and exactly the
wrong-file-named failure this whole bug is about. New test; the union covers
`fields`/`field_map`/`defaults` so a default-satisfied requirement still validates.

**4. `bug_historian` (low) — "absence of evidence is not evidence of absence" on
`research-agent`.** The seat was right to push: my 0 came from `orchestration_states`,
which reaps at ~24h. Re-measured in a table with a 4.5-month memory, and **checked the
relabelling trap first** (`llm_call_log.agent_type` is a known silent trap — a filter on
one spelling can read 0 while the rows sit under another):

```
llm_call_log, ALL agents:      48,525 rows, 2026-03-25 → 2026-08-04
llm_call_log, research-agent:       0 rows
agent_type ILIKE '%research%':  domain-research-classifier 55, vertical-exemplar-researcher 19,
                                adoption-researcher 2, directory-researcher 1  — none is this agent
page-content-writer:           18,590 rows
```

So the conclusion strengthens rather than flips: **`research-agent` is spawned on every
page build** (`page-content-writer.spawn_research_agent`, the only live spawner — I watched
that step execute in both verification runs) **and has never made a single LLM call in the
4.5 months `llm_call_log` retains.** Opting `extract_topic` into `required` would change a
path that has produced no observable LLM work in that whole period — still the right call
not to touch it, but now for a properly evidenced reason.

> **Flagged, NOT chased, and not this lane's:** an agent spawned on every page build that
> has never made an LLM call is worth someone's attention on its own. Two candidate
> readings — spawned but never actually called (cf. the known spawn→call handshake race),
> or called but never reaching an LLM. **Stated precisely: `llm_call_log` measures LLM
> calls, so this says the agent has never made one, NOT that it has never run.** No bug
> filed by me: I have not diagnosed it and would be filing a symptom.

**5. `bug_historian` (medium) — no shipped guard on `storeActionResult` itself.** Not
acted on, deliberately, and the seat's own recommendation is *"the router treat the
RFC_012 addendum as a required follow-on, not optional reading"*. That is a human decision
about a shared-mechanism guarantee, which by the 2026-07-29 ruling §1 is architecture
scope, not council-gate scope. It is recorded in RFC_012 with the working detector
specified — including that **both naive versions return 0 on the bug that motivated them**.

**One more claim converted to a check** while waiting on the verdict: I told the council
that failing loud after the roll cannot break a legitimate build because
`check_has_ready_sections` gates upstream. Verified — it is a `conditional` on
`section_plan.ready_count > 0` whose `then_step` is the only route onward, and it runs
**before** `load_current_section_content`, so it reads the FLAT plan and was never itself
broken by the wrapper. Consistent with builds failing at `select_sections` and not earlier.
