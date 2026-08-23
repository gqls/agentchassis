# HANDOFF — `bugfix_315` — state at 2026-08-23 10:00Z

> **SUPERSEDES `HANDOFF_2026-08-21_continue_here.md`.** That file is still worth reading for the
> build history, but **two of its load-bearing claims were false** and are corrected in place there:
> its "1 true positive, 0 false positives" grading, and its §2 "THE ONLY THING LEFT".

## 0. THE ONE THING TO KNOW BEFORE YOU DO ANYTHING

> **✅ UPDATED 2026-08-23 12:00Z — THE ROLL HAPPENED AND BOTH GUARDS ARE LIVE.** Chassis
> **`v1.0.1328`**, pods started **11:51Z**, stamp **`2dbe12f1d`**. `git merge-base --is-ancestor`
> is **YES** for both `14a50e533` (D11) and `de5d180fc` (D10), and **NO** in both reverse
> directions, so the test discriminates. The first post-roll run reports
> `checks_run: [site_unreachable, page_content_divergence]`, `checks_failed: []`,
> `checks_unregistered: []` — the "what did I break?" query, answering nothing.
>
> **The section below is kept because its REASONING is the reusable part** — for ~2 hours a build
> had been reported as deployed and had not reached the chassis, and the tell was that `IMAGE_TAG`
> still named the tag already running.
>
> **ONE VERIFICATION REMAINS, and it is the lane's closure test.** `vetcomparison.uk` has not been
> swept since `10:14Z` (per-site floor is ~4h), so the guard has not yet met the real case. Both
> false items are now `rejected`, so the dedup slot is FREE — which makes the next pass a genuine
> test rather than a dedup no-op. Check it with the DEMAND CONTROL FIRST, because "no item filed"
> from a check that never ran looks identical to success:
> ```sql
> -- 1. the demand control: did a pass actually run over this site after the roll?
> SELECT created_at, status FROM orchestration_states
>  WHERE owner_agent_type='availability-discovery-agent'
>    AND site_id='72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND created_at > '2026-08-23 11:51Z';
> -- 2. ONLY IF (1) IS NON-EMPTY: did it file anything?
> SELECT id, status, created_at FROM site_work_items
>  WHERE item_type='page_content_divergence' AND created_at > '2026-08-23 11:51Z';
> ```
> Expected: (1) non-empty, (2) **empty**, and the pod log carrying
> `origin object matches the fingerprint; the difference is edge-injected`.
> `[MEASURED 2026-08-23 11:56Z]` the suppression condition still holds — `Accept: */*` returns
> exactly `pages.content_hash` 3/3, browser `Accept` returns the injected body 3/3 — so the guard
> has something real to fire on.

~~**Two guards are committed and NEITHER IS RUNNING. The chassis has not been rebuilt.**~~

`[MEASURED 2026-08-23 09:45Z]`, and this is the whole of the evidence:

```bash
kubectl -n ai-persona-system get deploy agent-chassis \
  -o jsonpath='{.spec.template.spec.containers[0].image}'      # v1.0.1327
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.status.startTime}{"\n"}{end}' # 2026-08-22T19:05Z — 15h old
grep '^IMAGE_TAG' makefile                                      # v1.0.1327  <- SAME as running
```

```sql
SELECT git_commit, count(DISTINCT pod_name) FROM service_binary_capabilities
 WHERE service='agent-chassis' AND name='page_content_divergence' GROUP BY 1;
-- bd454eb93… | 129     (one distinct commit across every pod)
```

`git merge-base --is-ancestor 14a50e533 bd454eb93` → **NO**. The fix is not in the binary.

⚠ **`IMAGE_TAG` still names the tag that is already running.** Per `CLAUDE.md`, a same-tag rebuild
serves the node's **cached** image, and a deploy with an unchanged image spec does not even restart
the pods — which is consistent with what is observed here (deployment revision 885, unchanged since
19:05Z yesterday). **Bump `IMAGE_TAG` before building, or the roll will ship nothing and look fine.**

**So the next action is: bump `IMAGE_TAG`, `make build-agent-chassis`, and let the owner roll it.**
Then verify at the artefact — `service_binary_capabilities` again, plus the ancestry test with its
reverse control.

### ⚠ Expect the false positive to RE-FILE before then, and that is understood, not new

The `vetcomparison.uk` item was set to `rejected` on 2026-08-23. `idx_swi_dedup` excludes `rejected`,
so **the dedup slot is now free while the running binary still lacks the guard** — the check will
re-file the same false item on its next pass over that site (~every 4–5h). It is flag-only
(`handler_agent` is empty), so it costs a triage read and nothing else, and it stops at the roll.
Either re-reject it or leave it; do not "fix" it by reopening the old item.

## 1. THE HEADLINE, CORRECTED — 0 true positives, 1 false positive

The previous handoff reported `vetcomparison.uk/index.html` as a **LIVE CUSTOMER-FACING FAULT**,
serving visitors a stale page for ~21 hours, caught on six consecutive passes. **None of that was
true.** The page was correct throughout.

Cloudflare Web Analytics is enabled on that zone, and Cloudflare injects a ~359-byte
`static.cloudflareinsights.com/beacon.min.js` `<script>` into anything it treats as browser HTML.

`[MEASURED 2026-08-23, 5 fetches per header]`:

| Accept | hash | == stored fingerprint? |
|---|---|---|
| browser | `97fa37ca…` | no, 5/5 |
| `text/html,*/*` (what the check sends) | `97fa37ca…` | no, 5/5 |
| `*/*` | `4dbd143f…` | **YES, 5/5** |

`diff` of the two bodies: **two lines** — the beacon tag and its close. The served `last-modified`
was **15 seconds AFTER `deployed_at`**, so the object was current before any hashing was needed.

**Why it survived a day of scrutiny, which is the part worth carrying:**

- **The previous session was MORE rigorous than usual** — 5 fetches, stated N, browser `Accept`,
  cache-buster — and every bit of that rigour went into confirming that the two hashes **differed**,
  which was never in doubt. **Rigour aimed at the half you already know buys nothing.**
- **`≠ stored` was written down as "8/8 OLD".** "OLD" was never observed; it was inferred, and then
  quoted onward as an observation.
- **"Six consecutive passes, unprompted" felt like accumulating evidence.** On an injecting zone the
  check *cannot* match, so it re-flags every pass for ever. **Repetition of an unconditional result
  is not corroboration.**

**The rule:** a hash comparison can only tell you **THAT** two bodies differ. Before naming the
difference — stale, old, truncated — **`diff` them.** Logged in `WRONG_CALLS.md` and `LANDMINES.md`.

## 2. What was done 2026-08-23

| piece | state |
|---|---|
| **D11** raw-object probe | **COMMITTED** `14a50e533` — 3 mutations, 3 distinct test failures. **NOT LIVE** |
| **D10** measured body floor (2048) | **COMMITTED** `de5d180fc` — 2 mutations, 2 distinct test failures. **NOT LIVE** |
| **D8** 60-minute settle window | ✅ **LIVE**, settled at the artefact |
| `547` + `526` ledger rows | **RECORDED** — both were applied ≤ 2026-08-21 21:53Z by an unrecorded session |
| The false item | **`rejected`**, mechanism + evidence in its `result` |
| Docs | handoff (corrected in place), NOTES, README, RUNBOOK Parts 4–6, PLAN D10+D11, `SUMMARY_2026-08-23`, `WRONG_CALLS`, `LANDMINES`, register `DGH-015` + index row |
| Council | **round `1ceef75a` died with NO VERDICT** — see §4 |

**D8 was settled by a route this lane did not know about, and it is the durable one.** The item
route is dead (no item filed since D8 landed; the one that exists predates it and reads `1800`). The
`build provenance` line had scrolled out of a **full** `kubectl logs`, and probing the binary for
candidate commit shas returned `absent` for every candidate with **no positive control** — proving
nothing, exactly as the old handoff warned. What worked:

```sql
SELECT git_commit FROM service_binary_capabilities
 WHERE service='agent-chassis' AND name='page_content_divergence';
```

`service_binary_capabilities` (RFC_040, `platform/buildcapability`) records what each running binary
actually registers, plus its commit, refreshed continuously. **It has no shelf life.** Two other
lanes (`bugs_open/215`, `299`) were burned by the grep-the-binary route it replaces.

## 3. WHAT IS LEFT — and can this lane close?

**Not yet, and the only thing standing between it and closure is one build.**

| item | state | whose |
|---|---|---|
| ~~Roll the two guards~~ | ✅ **DONE 2026-08-23 11:51Z**, `v1.0.1328` / `2dbe12f1d`, both ancestry-proved with reverse controls | done |
| **Verify at the artefact — THE CLOSURE TEST** | pending the next natural pass over `vetcomparison.uk` (~14:14Z; last swept 10:14Z, ~4h floor). Both false items are `rejected` so the dedup slot is free and the pass is a real test. **Run the demand control first** — see §0 | **this lane's, and it is the last one** |
| **Resubmit the council round** | `submission_315_raw_object_guard.json` covers BOTH guards, validated, ready. `RESUBMIT_CORR=1ceef75a-81ee-4302-8182-69b0f6602bca` | this lane's, when the cap lifts |
| **D9 — escalate on PERSISTENCE across passes** | unbuilt. Convergence times (seconds → ~17 min → 1h20 → 21h) OVERLAP the failure, so no settle-window value separates them | open design question |
| **D6 — unarmed stamper NULLs the hash** | unbuilt; 6 of 6 armed today, so it is a backstop for the NEXT one added | open, low urgency |
| **`RFC_038` close or ratify** | survey done, change shipped | **NOT this lane's — it should not grade its own homework** |
| **`bugs_open/336`'s durable guard** | a test that every key an action READS is declared in its own spec | not this lane's |
| **`collectUniqueValue` extraction** | blocked on the `staged_component_build` lane taking the resolver route | theirs first |

**The honest closure test:** after the roll, one clean sweep in which the check files nothing and its
skip log shows `edge_injected: 1` for vetcomparison. At that point D9 and D6 are *improvements*, not
defects, and the lane can close with them named as follow-ons.

⚠ **What this lane must NOT claim on closing:** that the check has caught a real delivery fault. It
has not. It has **zero** live positives, and that has been true since it was enabled — the one
finding was its own false positive. It is a regression guard proved by induced faults, and saying so
plainly is the whole point after this week.

## 4. The council round died, and it is not this plan

`SUBMISSION_CORR = 1ceef75a-81ee-4302-8182-69b0f6602bca`. Admission passed, the `fix_plan` artifact
was written, and the orchestration terminated at `current_step = complete_invalid` — **an error step,
not a fourth verdict**. No `council_report`, no verdict note.

The discriminating measurement is fleet-wide, not lane-local:

```sql
SELECT count(*) FROM doc_notes WHERE categories ? 'council-gate' AND created_at > '<day>';
-- 2026-08-23: 0        2026-08-22: 54
```

Zero verdicts across **every** lane against 54 the day before is `243-anthropic-cap` (third
occurrence, onset 2026-08-22 18:15:35Z). **Do not resubmit until the cap lifts, and do not read it as
a rejection.** Both commits are therefore **UNREVIEWED**, deliberately and on the record; the
`Council-Submitted:` trailer is honest but will never resolve, because there is no verdict for `098`
to resolve to.

Two client-side validations DID reject earlier drafts and both are reusable: `.plan.risks` must be a
**string**, not an array; and an edit whose `sketch` is comment-only is refused outright ("a fix plan
proposes changes, not observations") — so a documentation-only edit must be folded into another
edit's rationale.

## 5. Traps this session hit for real

1. **Four config rows sharing a timestamp to the microsecond looked like one deliberate transaction.
   `count(*)` on that timestamp returned 204** — the whole table. It was the fleet release, 55
   seconds before the pods started. **The distinguishing query is a `count(*)` and it costs nothing**;
   without it I would have written a confident wrong apply-time into the ledger. It also **destroyed
   the per-row timing evidence**, which is why the ledger `notes` state a proven upper bound rather
   than a time.
2. **`SELECT snapshot_agent(...)` returns the id of the row it SNAPSHOTTED, not of the new snapshot.**
   Seeing that id in `agent_definitions` after a rolled-back rehearsal reads as "my transaction
   leaked". It had not — `count(*) WHERE created_at > CURRENT_DATE` was **0**. **A returned uuid does
   not tell you what it is a uuid OF.**
3. **A migration's own already-applied guard is a state oracle.** `sed 's/^COMMIT;$/ROLLBACK;/'` then
   run it: `547` aborts with *"already applied"*. That settled in one command what the silent ledger
   could not.
4. **The floor caught its own test data.** Five existing fixtures used `bytes: 10` / `100`. Raised the
   fixtures to realistic sizes rather than lowering the floor to fit fiction.
5. **The dedup index excludes `rejected`** — so closing a false item frees the slot and the
   still-unfixed binary will re-file it. Foreseen here rather than discovered later.

## 6. Where the documents are

```
docs/agent_docs/docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/
  HANDOFF_2026-08-23_continue_here.md   this file — START HERE
  HANDOFF_2026-08-21_continue_here.md   the previous one, corrected in place; good build history
  PLAN_2026-08-19_…md                   D1–D11; D10 and D11 decided 2026-08-23
  NOTES_…md                             technical log, read from the BOTTOM. The missteps are the point
  RUNBOOK_…md                           SIX parts. Part 4 = reproducing a divergence by hand (THREE
                                        fetches + the diff). Part 5 = has my Go change rolled.
                                        Part 6 = is a migration applied when the ledger is silent
  README_where_we_are.md                the owner's plain-prose log
  SUMMARY_2026-08-23_…md                the current read-out; the series is the record
  submission_315_raw_object_guard.json  covers BOTH guards, validated, ready to fire
```
Elsewhere: `bugs_closed/315` · `bugs_open/336` · `architecture_review/RFC_038` · register `DGH-013`,
**`DGH-015`** (corrected) · `LANDMINES.md` (new entry: a CDN can ADD bytes) · `WRONG_CALLS.md`
(2026-08-23) · `sql_for_agents/526_*`, `547_*` (**both applied — do NOT re-run**).
