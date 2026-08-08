# NOTES — `bugs_open/201` lane

Running record, append-only, newest at the bottom. Evidence, what the system actually said,
and every misstep.

## 2026-08-05 — lane opened from the 194 lane, fix-1 committed

### How this lane started, which is itself the lesson

I was on `bugfix_194` running its two owed checks. Check 3b needed a live dispatch that reaches
`save_page_sections` inside `site-work-orchestrator`'s build loop. Chasing a target for it, I
hit three walls in succession — and the third one was this bug, already filed by another lane
the day before. **Grepping `bugs_open/` for the mechanism before filing anything is the only
reason this lane is a contribution rather than a duplicate investigation.**

### What I did NOT re-derive

201 §3 asks that its failure mechanism not be re-verified — error text, `087` citation and the
artefact-level proof are all confirmed with correlation ids. I took that at face value and
spent the budget on its **open** question (§1) instead. Worth saying because the temptation to
re-run someone else's evidence is strong and it buys nothing.

### The cause is sharper than the error message says — and that matters

The error reads *"planned its own sections and none are ready"*, which sounds like a
data-readiness problem: the sections exist, their inputs are not available. **It is not.**

`plan_sections_action.go:867-875` has an **empty-input early return**:

```go
if len(sectionNames) == 0 {
    return map[string]interface{}{ ..., "ready_count": 0, "reason": "no sections to plan" }, nil
}
```

`sectionNames` comes from `inputs.GetRaw("sections")`, and `page-content-writer`'s self-plan
step binds that to `input_data.current_page.sections`. So the real failure is **"no sections
were supplied"**. Measured, all 14 `page-content-writer` items in history:

| item_type | `spec ? 'sections'` | keys | n |
|---|---|---|---|
| `literal_markdown` | **false** | check, findings, fix, original_pipeline, page_id, page_name, page_url | 12 |
| `placeholder_contact` | **false** | check, findings, original_pipeline, page_id, page_name, patterns | 2 |

A message that describes the *symptom state* rather than the *input state* sent the reader
looking at component `input_schema` data requirements. The disconfirming question was one
`spec ? 'sections'` away.

### §1's open question, answered — empirically, which beats the code-read

201 §1 asked whether `page-build-handler` returns `ready_count > 0` for an already-built page,
and proposed reading `load_spec_sections` to find out. I read it (it sources from
`site_specs.site_plan`, authoritative, `pages.sections` fallback — so it never depends on the
caller's spec), **but the stronger evidence was in the outcomes table**:

`page-build-handler`, items whose page already has `page_components` rows —
`content_rewrite` **19 complete**, `empty_section` **12 complete**, `empty_internal_href` **1**.
Against `page-content-writer`'s **zero** genuine successes in 14 items.

A code-read tells you what should happen; 32 completed rows tell you what does. Both were
cheap; only one of them can be wrong about a live system.

### The third check was an outright inconsistency, and I nearly missed it

`check_component_standards:477` files `ItemType: "needs_content_page"` — **the same item type
`write_build_items` produces** — and `write_build_items` routes that type at
`page-build-handler` (`load_work_item_actions.go:242-243`, where *every* entry in
`availableBuilders` resolves to `page-build-handler`). One item type, two handlers, decided by
which producer happened to file it. That one needs no argument about writers at all.

### A candidate 201 did not list, which I chose against — recorded so it is not re-proposed blind

`section-editor` is live and is exactly the right *shape*: `apply_section_edit`, contract
`domain + edit_type + (page_component_id | page_name+slot_name) + edit params`, and
`fix_component_template_action.go:27` says content changes are supposed to go through it. The
`literal_markdown` item even carries `findings[].slot_name` (`hero`) and `field`
(`subheadline`), so the target maps cleanly.

**It cannot serve, for a reason visible only by reading its step list:** `ensure_site_record →
spawn_deployer → load_edit_context → apply_edit → deploy_page → update_page_status →
trigger_deploy → complete`. **No LLM step.** It applies an edit someone else composed. And the
item's `fix` is an *instruction* — "Rewrite the affected fields WITHOUT markdown syntax… if the
writer wanted emphasis, re-word so the words carry it" — not a replacement string. Routing
there would have needed a new compose-then-edit agent: a new shared mechanism, architecture
scope, to fix a routing bug. Full trade-off in PLAN.

### Missteps, in order

- **I wrote SQL before reading the schema, twice, and both times it errored rather than
  misled.** `orchestration_states` has no `agent_type` (it is `owner_agent_type`);
  `jsonb_object_keys` cannot sit bare beside an aggregate. CLAUDE.md says `\d <table>` first
  and I skipped it. Cost: two round-trips. **The reason this was cheap is luck, not care** — an
  errored query is loud. The same haste against a column that *exists but means something else*
  is the expensive version, and this lane's sibling (194) has a `WRONG_CALLS` entry for exactly
  that (`created_at` vs `occurred_at`).
- **`who-owns.py 201` named MY OWN lane as the likely owner.** It reads commits and doc
  mentions; I had just written 12 mentions of 201 into `bugfix_194`'s files an hour earlier. The
  tool is honest and the signal was an artefact of my own action. **An ownership check reflects
  what has been WRITTEN, not who is working** — I checked live `.jsonl` transcripts as well,
  which is what actually answered it.
- **Two other sessions were live in `page-content-writer` code while I worked** (transcripts
  modified ~15 min before I looked). Neither mentioned 201 (`grep -c` → 0): one is the `156`
  dedup lane in `save_page_sections_action.go`, the other the concept-register drift lane. So
  201 was genuinely free — but I only know that because I looked at the *tree and transcripts*,
  not just at `git log`.
- **My first instinct on finding the site lock was to ask whether it could be released.** It
  can, and doing so would have re-run a filed incident (`aee11cb90`, a live homepage rebuilt
  under a held lock on that same site). The lock is the control added *because of* that. Noted
  because the instinct was wrong and immediate.

### The consequence I am declaring rather than burying

After this change, **no producer anywhere files a `page-content-writer` work item** — grep
across `platform/ internal/ pkg/` returns zero. `site-work-orchestrator`'s `build_items_loop`
is gated on exactly that handler, so it is now permanently unfeedable.

I am **not** repointing that filter at `page-build-handler`. `build-dispatch-loop` already
consumes those rows; a second consumer of the same rows is a double-dispatch invitation, and
this estate has filed history in that class. The loop was already unreachable in practice — it
has never run (absent from `agent_run_stats`, which has no reaper and does track
orchestrator-shaped agents), and its only possible inputs were the items that hard-fail. **This
makes an existing deadness explicit; it does not create one.** Written into the bug file as an
explicit do-not-tidy, because it is precisely the kind of thing a later session "cleans up".

> **[INFERRED — not measured]** That loop being unfeedable *may* mean seed 312's
> `site-work-orchestrator` mapping (from `bugs_closed/194`) can never be exercised on that
> route. I have not run it and am not asserting it. Flagged in 194's NOTES and in the CONTRIB.

### The council verdict — APPROVED, and two of its five objections were substantive

`decided_by: "approved with 5 advisory objection(s) — none high-severity"`, 15 reviewers,
2 abstained, 0 unreadable, `gated_by_truncation: false`. The run dispatched in **under a
minute** (submitted ~11:41Z, first seat executing 11:41:45Z, verdict by ~11:47Z) — nothing like
the 29-minute queue CLAUDE.md warns about, so budget by the worst case but do not assume it.

Approval is not the interesting part. Acting on the objections is.

**1. `editquality` [medium] — WAS RIGHT, and I had understated the cost in my own PLAN.**
It cited a landmine I had not read: `page-build-handler`'s writer never sees the page's own
stored prose unless `spec.mode="recreate"`, which none of these three checks sets. Read it
(`LANDMINES.md:4433`) and it is confirmed root cause on `bugs_open/178`, not a suspicion:
`load_existing_content_action.go:64-69` no-ops with `{"has_existing": false, "reason":
"not_recreate"}`, and `load_page_record` carries only sections/title/page_type — no prose. **So
the repair rewrites the section from scratch and prior prose is lost.** My PLAN had called this
"heavier than the ideal repair", which is too soft; corrected in place there with the reasoning
for why the decision still stands (the alternative is a permanent defect, not a field edit).
**And setting `mode=recreate` is explicitly the wrong fix** — it sources the adoption-crawl
snapshot, i.e. stale content rather than none.

**2. `guardian` [medium ×2] + `bug_historian` [medium] — the SECOND dispatch gate. Checked; clear.**
The objection was the one this session's own SessionStart landmine warns about: a dispatcher has
two gates, and fixing the visible one leaves the key dropped. Verified rather than argued:

- `scripts/audit-relay-gaps.sh` — 175 agents decoded, **0 relay-gap findings**;
  `build-dispatch-loop` is not even among the 2 uncovered dispatcher-shaped relays, so it is
  asserted.
- `build-dispatch-loop.call_handler` `input_mapping` forwards `domain` ← `input_data.domain` and
  `site_id` ← `current_item.site_id` — **both** of `page-build-handler`'s `input_contract`
  required fields (`{required: [site_id, domain], optional: [page_name, page_id, sections]}`) —
  plus `spec`, `current_page`, `item_type`, `work_item_id`, `page_name?`. It is
  **item-type-agnostic**: there is no per-`item_type` allow-list to exclude the new types.
- `spawn_handler` uses `agent_type_field: current_item.handler_agent`, so the spawn follows the
  ROW. New rows carry the new handler; **existing rows do not** (RUNBOOK R6 trap 1).
- ⚠ **Gotcha worth recording:** reading these steps with a top-level `#>> '{workflow,steps,…}'`
  returned EMPTY for both, because they are nested in a loop `sub_workflow`. An empty result
  there reads exactly like "the step does not exist". `jsonb_path_query(…, '$.**.steps')` — the
  same trap as R2, hit again in the same session.

**3. `bug_historian`'s deeper point stands and is NOT closed by the above.** It observed that
this trades a LOUD hard-fail for a pipeline with filed history of *silent* partial success
(016b §9: sections deferred and dropped; "page build completes having built nothing"). My
evidence for `page-build-handler` was **by analogy** — different item types succeeding — not
these item types. That is a fair characterisation and I have not closed it. It is why RUNBOOK
R6 trap 3 requires `content_data` to change rather than accepting `complete`.

**4. `debug_historian` [medium] — no post-deploy pod-grep was proposed.** Correct. I had written
"inert until a roll" but never named the artefact check. Added to RUNBOOK R6 and the bug file:
grep the RUNNING pod's binary for the new literal, every replica, never git or the tag.

**5. `prior_art_librarian` [low] — my "site-work-orchestrator has never run" claim.** It rests
on absence from `agent_run_stats`, which the council cannot see, and there is a standing
landmine about "has agent X ever run" claims being false negatives from `orchestration_states`'
~24h retention. **I used `agent_run_stats` deliberately BECAUSE of that retention trap** (it has
no reaper) — but the seat is right that the claim is unverifiable from its side and should not
be repeated in a close-out unchecked. Marked `[UNVERIFIABLE-BY-COUNCIL]` where it appears.

**6. `architecture` [medium] — the RFC-shaped finding, filed as RFC_014.** This is the FIFTH
site of the same defect class, and the only guard checks that the string names a KNOWN agent,
not that the agent can CONSUME the filed spec shape. Seat's verdict: "Approve the edits as
written … ship it", but on record. `RFC_014_handleragent_is_a_stringly_typed_routing_contract.md`
costs three options, recommending the cheap floor (a narrower legal-direct-dispatch set) over
continuing to patch strings one site at a time.

`editquality`'s low objection — that edit 4 is comment-only and should not count as an edit — is
accepted without argument. It was documentation upkeep bundled into the count.

### State at end of session

Committed `37afbb847`; council **APPROVED** (`71523705-07d1-4067-9c5d-af371ba84b89`), verdict
read and its objections acted on above. The commit carries `Council-Submitted:` — correct for a
pre-verdict commit, and `098` credits it automatically now the correlation is approved; **no
amend, forward-only.** I have deliberately NOT written a `Council-Reviewed:` trailer onto a
later commit, because that trailer belongs to the commit carrying the reviewed code and
back-dating it is the report's MISMATCH surface.

**Inert until a chassis roll** — the fleet is on `v1.0.1252`, which predates the commit.
Symptom 2 untouched by 201 §2's own instruction. Verification traps for the next session are in
RUNBOOK **R6** (now four: the stale `handler_agent` on existing rows, the locked site returning
success-with-zero-items, `complete` not being proof, and **prior prose legitimately vanishing**)
plus **R7**, the post-deploy pod-grep the council asked for.

Filed out of this round: **RFC_014** (the stringly-typed routing contract, five recurrences).

### One last thing I got wrong today, worth its own line

I ran `who-owns.py 201` and it named **my own lane**. I noticed and discounted it, which is the
only reason it did not mislead me — but the mechanism is worth stating: I had written 12
mentions of 201 into `bugfix_194`'s files an hour earlier, and the tool ranks by mentions and
recent commits. **An action I took to be diligent (cross-referencing the bug thoroughly) made
the ownership check point at me.** The generalisable form is already in memory as "your action
moves you to the back of the selector"; this is the same shape pointing forward instead of back.
The check that actually answered ownership was grepping live `.jsonl` transcripts and finding
two concurrent sessions in `page-content-writer` code, neither of which mentioned 201.

---

## 2026-08-06 — symptom 1 PROVEN, by inducing the test instead of waiting for it

### The waiting plan was wrong, and finding that out was the day's real work

Yesterday I wrote "the proof arrives on its next run" in three documents. **There is no next
run.** `quality-discovery-agent` has one `scheduled_tasks` row, `enabled=false`, a 07-30
one-shot — and `SELECT count(*) … WHERE target_agent_type ILIKE '%discovery%' AND enabled`
returns **0**. Nothing in that whole class is driven. I had inferred a cadence from
`agent_run_stats.run_count` = 22 without opening `scheduled_tasks`; **22 matches no cadence** —
daily since 07-26 is ~11, hourly ~250 — and I read that as "infrequent" instead of "unscheduled".
Logged in `WRONG_CALLS.md`. Owner authorised a deliberate sweep at `gaswholesalers.com`.

### Pre-flight — three checks, each of which could have made the run worthless

1. **Is the defect still there?** If the page is clean the check files nothing and the result is
   zero rows, which I had already documented as "not a pass". Queried the check's own bold
   pattern against live `page_components`: `how-pricing-works` / slot `pricing`, `content_data`
   matches, `rendered_html` does not. **Defect live.** So the sweep had something to find.
2. **Can a filed item start a repair?** I had told the owner "detects, does not repair" and
   promised not to trust my own sentence. Read it: `build-dispatch-loop.load_items` → action
   `load_work_items` → `status IN ('triaged','approved')` (`load_work_item_actions.go:633`).
   A `detected` item is inert. **Confirmed before dispatching, not after.**
3. **Is the site free?** `gaswholesalers.com` UNLOCKED, no lane rebuilding it, nothing in flight.

### ⚠ The trigger script is a weapon — do not run it

`075_trigger_discovery.sh` looked like the obvious tool. Reading it first (this lane's second
such catch) found two things: its `case "$2"` accepts only `design|completeness`, so it **cannot
fire quality discovery at all**; and **below the `CORRELATION_ID=` echo, where reading normally
stops, it runs an unconditional `UPDATE site_work_items SET status='triaged' … WHERE domain =
'finetuning.uk' AND status='detected'`** — hardcoded, ignoring `$1`. That is the exact status
transition that makes a backlog **dispatchable**, on another lane's customer site, where a
page-build repair regenerates prose rather than editing it. Landmine filed + synced. I copied its
kcat envelope into a standalone script and left the tail behind.

### The result

Corr `35e24460-d3f9-4d0e-a4bb-28bb9bc82a5c`. `generic` → `quality-discovery-agent`, both
COMPLETED, 11:34:22Z. `agent_run_stats` 22 → **23** (so the run is attributable, not inferred).

```
d2a6117d-8840-4ee1-af97-6ff688c2758c | gaswholesalers.com | literal_markdown |
handler_agent = page-build-handler | detected | created_by = quality-discovery-agent |
page how-pricing-works | 2026-08-06 11:34:22.809653+00
```

**All 14 pre-fix items carry `page-content-writer`; this one carries `page-build-handler`.** The
discriminator is the value itself, which is why this works where a pod-grep could not: the change
adds no string, but it *writes a different value into a row*, and rows are observable.

It filed at `detected`, so nothing was repaired — pre-flight 2 held.

### An unlooked-for corroboration of symptom 2

Pre-flight 1 turned up something I was not looking for: gaswholesalers' **existing**
`literal_markdown` item is `status='complete'` **while the markdown is still in `content_data` on
that page.** That is symptom 2 — a handler reporting success having written nothing —
reproduced independently, two days after 201's filer proved it at artefact level. It is also a
neat demonstration of why `complete` is not evidence, on the very site being used to prove the
other half.

**State: symptom 1 fixed, live, proven. 201 stays OPEN for symptom 2, which is now the next
work in this lane.**

---

## 2026-08-07 — symptom 2 DEPLOYED, and the live behavioural test (owner-authorised)

### Deployment, pod-verified — and this one COULD be greppable

`v1.0.1262`, both replicas up 05:47Z; commit `7e62f4a07` at 01:37Z UTC predates it.

| replica | `VerifyLiteralMarkdownResolved` | `…ResolvedV2` (fabricated) |
|---|---|---|
| `…-5ghft` | **4** | 0 |
| `…-dfk4b` | **4** | 0 |

**Unlike symptom 1's fix, this one is pod-greppable** — a genuinely new symbol rather than a swap
between two literals that were both already in the binary. The fabricated near-miss returns 0, so
the grep discriminates; the positive alone would only prove the pipeline.

### Choosing a canary — the obvious one was the confounded one

The natural canary was gaswholesalers.com/how-pricing-works. **It is the envelope-guard page**
(re-confirmed 08-07: `type='text'` + `jsonb_typeof(result)='string'`), so any verdict on it is
uninterpretable. Measured the fleet for the alternative instead of assuming one existed:

```
gaswholesalers.com | how-pricing-works | pricing      | envelope_confounded = true
webdesign.co.uk    | news              | news-listing | (not envelope-shaped)
```

**Exactly two live literal-markdown instances exist fleet-wide**, and only one is clean. That one
is also one of `bugs_open/184`'s three originals.

### Method, and the three pre-flight checks

`bugs_open/201` §"How to verify a fix" prescribes re-arming **one** of the failed items rather
than dispatching fresh, and its precondition ("only once a real fix is ready to test") is now met.
The `news` page's own item — `efaa39a2`, `failed` 3/3, error `fail_no_ready_sections`, i.e. 201's
symptom-1 signature — is the ideal subject.

1. **Site state:** `webdesign.co.uk` UNLOCKED; page `news` is `rebuild_policy='generic'` (not
   `owned`, so `bugs_open/208`'s commit-before-guard trap cannot fire here).
2. **What else would ride along:** the exact `load_work_items` predicate returned **zero**
   triaged/approved items on that site, so triaging mine picks up one thing and nothing else.
   (The 075 trigger's sin was doing this blind, on a hardcoded other domain.)
3. **Artefact baseline BEFORE touching anything** — without it, "the markdown is gone" is
   unfalsifiable:

```
hero          | md5 3e77770ccea4619f6d7b8c78c733e3a9 |    304 B | 08-05 14:21:41Z | bold_md=false
news-listing  | md5 9df6c43d4eab12ab5600ca0f760daacc | 10 232 B | 08-05 14:21:41Z | bold_md=TRUE
call-to-action| md5 45f6f2b8154d441af30e7c334f7c8af1 |    331 B | 08-05 14:21:41Z | bold_md=false
```

**Re-armed exactly one row** (`status=triaged, attempt_count=0, claimed_by=NULL, error=NULL`) —
**and `handler_agent='page-build-handler'`**, which is RUNBOOK R6 trap 1: omit it and the re-arm
re-runs the broken pre-fix route and looks like the fix failed.

Dispatched `build-dispatch-loop` (corr `c0c88fcb-cd68-4560-bb42-2f009dededb5`). Confirmed it
landed rather than trusting kcat's exit 0: the loop went `AWAITING_RESPONSES` at
`process_item_iter_0_call_handler` and a **`page-build-handler` reached `spawn_content_writer`**.
⚠ A *different* `build-dispatch-loop` row COMPLETED 19s in — that is the scheduled
`build-pipeline-trigger`'s own instance, not this one. Reading it as "the run finished
instantly" would have been the wrong conclusion from a correct query.

### What either outcome proves

- **Repair worked** → markdown gone from `content_data`, md5 of `news-listing` changed, verifier
  returned `Resolved:true`, item `complete`.
- **Repair wrote nothing** → verifier returns `Resolved:false` and **completion is REFUSED**, item
  goes to attempts rather than being stamped `complete`. *This* is the outcome that demonstrates
  symptom 2 is actually closed.

⚠ **A third outcome is now possible and is BY DESIGN:** the item stays `claimed` indefinitely.
Migration 331 removed `literal_markdown` from the claimed-item-timeout sweep, so a stuck item is
no longer auto-completed at 15 minutes on the handler's own evidence. Do not read "still claimed"
as a hang without checking that first.

### OUTCOME — attempt 2 (corr `78e15724…`): **the verifier fired and BLOCKED completion.** Symptom 2 proven closed.

```
completion blocked: post-fix verification found the defect still present:
18 finding(s) still present across 3 component(s); first: slot "news-listing"
field "items[1].summary" pattern bold in content_data — "**the `animation`**"
```

That is `VerifyLiteralMarkdownResolved`'s own `Detail` format verbatim, wrapped by
`CompleteWorkItemAction`'s block message. **The verifier was consulted, returned
`Resolved:false`, and completion was refused.** The item went to `failed` instead of `complete`.

**Before this change that run would have been stamped `complete`.** That is not a hypothetical:
it is exactly what happened to the gaswholesalers item, which is 201's original symptom 2.

**Artefact, against the baseline — the page WAS rebuilt, and the defect SURVIVED:**

| slot | md5 before → after | bytes | `updated_at` | bold md |
|---|---|---|---|---|
| `hero` | `3e77770c…` → `23d29af3…` | 304 → 347 | 08-05 → **08-07 08:37:26Z** | false |
| `news-listing` | `9df6c43d…` → `d90effc0…` | 10 232 → 10 157 | 08-05 → **08-07 08:37:26Z** | **still TRUE** |
| `call-to-action` | `45f6f2b8…` → `34bf533c…` | 331 → 308 | 08-05 → **08-07 08:37:26Z** | false |

All three components changed, so the repair genuinely ran and wrote. **It did not cure the
defect** — the writer re-emitted literal markdown into the very field it was dispatched to clean
(`items[1].summary`, `"**the `animation`**"`). 18 findings remain across 3 components.

### The second finding, which is 184's and not 201's

**Symptom 1's fix makes the dispatch WORK; it does not make the repair EFFECTIVE.** Routing
`literal_markdown` at `page-build-handler` stops the hard fail (proven 08-06) and the handler now
rebuilds the page — and `page-content-writer`, spawned behind it, writes markdown syntax back
into a text field. That is `bugs_open/184`'s underlying defect, untouched by anything this lane
did, and it is now **visible instead of hidden**. Noted on 184.

That is the verifier earning its place on its first real run: it converted a silent false success
into an attributable failure naming the field and the matched text.

### Honest notes on the run

- **`attempt_count` went 1 → 3 in a single dispatch**, so the item is now exhausted and terminally
  `failed`. `[UNMEASURED]` why it jumped two rather than one — possibly an internal retry inside
  the loop, possibly `fail_work_item` incrementing alongside the verification block. Not chased;
  recorded so nobody reads "3" as three separate operator dispatches. A further attempt needs
  `attempt_count` reset as well as `status`.
- **Terminal `failed` is the correct destination**, not a regression: three attempts, defect not
  cleared, so it goes to human review — which is where an uncleaned markdown defect belongs, and
  what the deferral note meant by writing against the handler's remit.
- **Attempt 1's failure was NOT the verifier** (spawn→call handshake race, upstream of
  `complete_work_item`). Recorded above as inconclusive at the time rather than claimed as a pass,
  which is the only reason attempt 2 was run at all.

---

## 2026-08-06 (evening) — symptom 2: the mechanism already existed, and the work was mostly not writing it

### The fix was a REGISTRATION, and finding that out was most of the job

Symptom 2 reads like "complete_work_item needs a new check". It does not. `verifiers.go` already
declares `ItemVerifier` / `RegisterVerifier` / `GetVerifier`; `CompleteWorkItemAction` already
consults the registry before stamping; **seven verifiers were already registered**; and
`literal_markdown` was **already listed** in `itemTypesWithoutVerifiers` with a written deferral
note explaining exactly how to write it. The header of `verifiers.go` even names symptom 2's
general form — *"a saga can 'succeed' without touching the defect"*.

So this is filling a declared gap with the estate's own machinery. **What I deliberately did NOT
do is change `complete_work_item`'s general trust of `handler_result`** — 201 §2's observation
stays true for every other unverified type, and narrowing it fleet-wide is a different, much
wider change than this bug licenses.

### The one way to get it wrong, and the note that stopped me

`verifier_coverage_test.go`'s `page_rerender` entry is the estate's cautionary tale: a working
verifier was written, tested, and then **deliberately held**, because re-running the detector's
predicate over the whole page is *stricter than the handler's remit* — it would mark
correctly-handled items unresolved, burn their attempts, and strand **1,849** of them in
`failed`, destroying a designed escalation.

`literal_markdown`'s deferral note carried that warning forward verbatim: *"write it against the
REPAIRING agent's rewrite remit, not the detector's predicate"*. That instruction is the reason
whole-page scope is defensible here: the handler is `page-build-handler` (symptom 1's own fix),
whose `build_pages_loop` runs `load_spec_sections → plan_sections` and rewrites **all** of the
page's spec sections. Whole-page **is** the remit. And unlike `page_rerender` there is no
two-strike escalation on this type to destroy — a failed verify goes to attempts, then human
review, which is where an uncleaned markdown defect belongs.

**Had I not read that note I would have written the same verifier and been right by luck.**

### The zero I refused to certify

A page with no scannable components scans clean. That observation cannot distinguish *repaired*
from *content lost* — and this lane has already measured the second: 31 of 106 components NULL
across 10 pages on ai-agent-orchestration.com, every component on 9 of them. Returning `Resolved`
there stamps `complete` over a destroyed page.

So `scanned == 0` returns an **error**, not a verdict. The caller records "could not verify"
rather than treating it as success. **I am aware that CompleteWorkItemAction fails open on an
error** (`verifiers.go:60-63`) — that is the registry's existing documented policy, and inventing
a different one for this verifier alone would be a silent divergence, so I left it and said so in
the submission instead.

### Drift, and why the scan is now one function

`verifiers.go`'s contract is that a verifier re-runs *"the SAME predicate the discovery check
used to create the item"*. Two hand-kept copies of a five-line scan is how that quietly stops
being true — and a verifier drifted **stricter** is the page_rerender disaster. So the per-row
scan is extracted into `scanComponentRowForMarkdown` and both call it. The row *selection* is
still written twice (site-wide vs page-scoped) because the scopes genuinely differ; a comment
records that their WHERE clauses must otherwise stay identical.

### Both lockstep obligations, and the one no test can check

Registering the verifier immediately failed **two** guards — the designed catch, working:

1. `TestEveryItemTypeIsVerifiedOrAnAcknowledgedGap` — *"HAS a verifier but is still listed as a
   gap … remove it from itemTypesWithoutVerifiers"*.
2. `TestRegisteredVerifiersMatchClaimTimeoutExclusion` — *"HAS a Go verifier but is NOT excluded
   in 220 … the claimed-item-timeout sweep will auto-complete it on handler-orchestration
   evidence alone, bypassing the verifier"*.

**Registering a verifier is not what makes it gate.** The sweep auto-completes at 15 minutes
unless the item_type is in its `item_type NOT IN (...)` list. And per the LANDMINE, `220` is only
the **declared** list — the LIVE `scheduled_tasks.pre_query` is a separate edit that no test can
check, and a lane has already left a verifier bypassable for two days by doing one and not the
other. Read the live column first (the landmine's own instruction): **one row, seven entries,
byte-identical to 220 — no drift to carry.** Migration `331` written against that exact string,
asserting both directions in `DO`/`RAISE` because plain `SELECT`s cannot stop a `COMMIT`.

### Two things I proved rather than assumed

- **The test bites.** Mutating `if len(remaining) == 0` to `>= 0` fails
  `RefusesWhenMarkdownSurvives` with the live shape (`**Decision Engine**` still in
  `content_data`). A happy-path test alone would pass for a verifier that always certified.
- **The lockstep test really reads `220` from disk.** My first archive-of-HEAD run **failed** —
  because I had copied only the Go files into the archive. Annoying, and it is the cleanest
  possible evidence that half of the lockstep is live rather than decorative.

### State

Committed `dc4f4e6b2`, council submitted `f14a8b64-4f71-4915-88d0-9587db845052` (verdict unread).
**Inert until a chassis roll** — Go change; the fleet is on `v1.0.1261`, which predates it.
**Migration 331 not yet applied.** The post-roll canary already exists and needs no new dispatch:
gaswholesalers.com's `literal_markdown` item on `how-pricing-works`, whose markdown is still in
`content_data`; a repair attempt on it must now be refused completion rather than stamped
`complete`.

---

## 2026-08-08 — RFC_017's missing number, measured. And it says option 4, which is not what I expected.

Picked this up because it is the only thing the lane had left that was ours to do: `RFC_017` asked
*"how often do registered verifiers actually error in production?"* and said the choice between
options 3 and 4 should not be made without it. Full write-up is now in the RFC itself; queries in
`RUNBOOK` R8. Evidence and the wrong turns here.

### The number

11 verifier consultations across the whole life of the gate (live `e1b8e1f84`, 2026-07-14,
v1.0.1116): **8 `verified`, 2 `error`, 1 `defect_persists`** — the last being this lane's own proof
of 08-07. Per verifier: `hardcoded_section_colors` 8/0 errors, `empty_section` **2/2**,
`literal_markdown` 1/0, and **the other five registered verifiers have never been consulted at all.**

`[MEASURED 2026-08-08]` and it could have come out otherwise — the query returns three distinct
statuses, so it demonstrably distinguishes them; a blind version would have returned one bucket.
**But two caveats make `2` a floor, not a count**, and they belong next to the figure every time it
is repeated: `result` is overwritten on each completion attempt, so this is current-state, not
history; and `n=11` over five days.

### The shape matters more than the rate, and that is the actual finding

**Zero infrastructural errors.** Both errors are one deliberate branch —
`check_empty_sections.go:412`, which is `bugs_closed/032`'s own accepted fix, choosing `error` over
`Resolved:false` precisely so the gate fails open and records a *visible unknown* instead of a false
green. The same branch is duplicated at `check_truncated_component.go:272` (never fired). So the
error path is armed on ~~2 of 8 verifiers by two authors~~ **[CORRECTED 2026-08-08 after the council
round: 4 of 8, by four authors — `content_duplication:631` and `page_canonical_collision:382` carry
it too, under different wording. I grepped the spelling instead of the class. See the triage entry
at the foot of this file.]**, both citing the documented policy correctly.

Then the part that inverted my expectation. 032 named its own disambiguator — *if the page still
expects the component, absence is deletion, not ambiguity* — and **both items pass that test**:

```
177bbb2e… (page ai-guides) and 8c4b10f1… (page insights), site 1368e337…, slot featured-article
  → both stamped 'complete' 2026-08-03, attempt_count 0, _verification.status='error'
  → pages.sections still lists "featured_article" on BOTH
  → both pages now serve a deployed 334-byte shell in slot featured-content:
    <section …><article class="featured-article"><div …><h1 …></h1></div></article></section>
```

The defect is still live on two production pages, recorded as `complete`. And the backstop the
policy leans on has not fired: **no item for a `featured-content` slot has ever existed
fleet-wide**, although `findEmptySections`' predicate run verbatim returns both components right now
(`empty_heading`, unlocked, unsuppressed, `build_status='deployed'` — so `bugs_open/185` is not the
explanation), and the check demonstrably ran on that site four times afterwards, retracting 10 other
items. So fail-closed would have been right on 2 of 2, and its feared harm — stranding a
legitimately-removed item — happened 0 of 2 times.

**Why detection never re-filed is NOT established and I am not asserting it.** Ruled out by
measurement, not by argument: the dedup index excludes `unresolved`, so April's zombies cannot block
a new key; and `bugs_closed/041` is cleared too — the site's four `needs_new_component` rows are
`category_section`/`article_grid` with `already_exists=f`, genuinely absent components rather than
041's snake-case miss. That is a `090` job, not another hour of grep.

### Three wrong turns, in the order I made them

1. **I checked component existence in the wrong table and nearly wrote it down.** Queried
   `site_components` for the two vanished `component_id`s, got `0 rows`, and had "the slot has no
   component at all" drafted. The verifier reads **`page_components`** — I only caught it by opening
   the function to quote its error branch. `site_components` is slot-level and `UNIQUE (site_id,
   slot_name)`; those ids were never eligible to be in it. An absence is only ever an absence *in the
   table you named*. Now R8 trap 3.
2. **Then I over-corrected into the opposite wrong answer.** With the right table I found both pages
   rebuilt seconds before the verification, carrying a `featured-content` slot where the item named
   `featured-article`, and concluded "legitimately replanned, so fail-open completed them correctly"
   — and said so out loud before checking. `pages.sections` reversed it: both pages still declare
   `featured_article`. Two confident readings in ten minutes, opposite conclusions, same data. The
   only thing that settled it was 032's disambiguator, which was written down a fortnight ago and
   which I had already quoted without applying.
3. **I read a `resolved_by` batch as a bypassed gate.** Ten `empty_section` completions carry no
   `_verification`; I took that as the gate being walked around, which would have been a much bigger
   claim than the RFC's. They are the check **retracting its own items** after re-observing them
   healthy (`work_items_common.go:287-301`) — verification by construction. The tell was
   identical-microsecond `updated_at` across each batch: a sweep, not a saga. Now R8 trap 2.

All three have the same shape as the pattern this lane's 08-06 summary already named: the check that
would have caught it was written down, by me or by someone else, and not applied to the case in hand.

### Same day, later — the owner ruled fail-CLOSED, and it is built

Ruling: **option 4 (fail-closed), option 2 (the lint) declined, option 3 (park) explicitly not
taken.** The reasoning for declining 2 is worth keeping: under fail-closed, `return VerifyResult{},
err` means "do not complete", which is the safe direction, so the guard it provides largely
evaporates. Option 3 stays open in RFC_017 with the retry cost named as the evidence that would buy
it.

Built, tested and committed the same day. Shape:

- `RegisterVerifier` keeps its signature and becomes the SAFE registration — the flip reaches all 8
  verifiers without editing 8 files. Fail-open is now `RegisterVerifierWithPolicy(t, v,
  VerifierPolicy{FailOpenOnError: true})`: **an opt-in field, unsafe default OFF**, per the
  2026-08-02 shared-seam ruling, because the old authority was a doc comment no reviewer of a
  calling check could see. **Nothing opts in.**
- `GetVerifier` → `(ItemVerifier, VerifierPolicy)`, so a caller cannot take the verifier and forget
  the terms; the zero policy is the safe one, so an ignored return lands safe.
- `verificationDecision` extracted as a **pure** function. Not tidiness — the policy branch was
  reachable only through a live `*sql.DB`, which is exactly why no test asserted the behaviour
  RFC_017 was written about. **Mutation-proven:** restoring `}, true` fails
  `TestVerificationDecision` on the two rows that encode the ruling, and passes again on restore.
- **A message that was silently wrong the moment errors could block.** The blocked-completion text
  was hard-coded to "post-fix verification found the defect still present" and read from
  `payload["detail"]` — on an error payload that is a claim the verifier never made, with an EMPTY
  body (error text lives under `"error"`). `blockedCompletionReason` splits it:
  `verification_unavailable` vs `verification_failed`. I found this by writing the caller edit, not
  by reasoning about it; a flip that only touched the policy would have shipped the false sentence.
- Payload gains `fail_open`, so a future census can tell a completed error from a blocked one — the
  census that produced RFC_017 could not, and would have read the post-roll rows backwards.
- The **unparseable-spec** branch takes the same policy. Same class; exempting it leaves a second
  silent completion path behind the one being closed. **Honest gap: I did not measure how many live
  specs fail to parse** — 0 such payloads appear in the census, which is weak evidence, and it is in
  the submission's risks block rather than glossed.

Council `a104d454-a4ff-4c95-a578-9a7e48c95100`, submitted before committing, so the commit carries
`Council-Submitted:` — the trailer that asserts nothing. Register entry `WII-011` in the same commit
as the seam (ordering-exemption condition 2), index row added, count re-grepped (1,794) and
uniqueness checked in the same pass.

**Misstep, and it is the multi-session one rather than a technical one.** While I was correcting my
own LANDMINES entry, another session committed that file (`d485b60ac`) and swept my in-progress
correction into their commit. Nothing was lost — it is at HEAD, forward-only holds — and it is the
exact scenario CLAUDE.md describes. Worth recording because the *entry itself* inverted within six
hours of being written: `_verification.status='error'` meant "completed" this morning and means
"blocked" after the next roll, with **both readings live in one table** and nothing in the row
saying which era it came from. `fail_open` (absent on every pre-roll row) is the discriminator.

### Council verdict read — APPROVED round 1, and one objection was RIGHT

`a104d454-a4ff-4c95-a578-9a7e48c95100` — **approved, 2 advisory objections, none high-severity**,
11 seats reviewing, 6 abstained, `gated_by_truncation:false`. Verdict read in full, not taken from
the decision word. Triage of every objection, including the ones that turned out to be answered:

1. **`guardian` [medium] + `editquality` [low] — "confirm no other caller of `GetVerifier` exists".**
   **Answered, and it was already evidence I held:** exactly one caller
   (`complete_work_item_verification.go:57`), and `go build ./...` is clean — assigning a two-value
   return to one variable is a compile error, so the build *is* the sweep. Recorded rather than
   waved: the seats could not see the build from the submission.
2. **`guardian` [medium] — "the other 7 verifiers' owning pipelines … asserted ('I assert none
   does') rather than enumerated".** **THIS ONE WAS RIGHT AND IT CAUGHT A REAL ERROR OF MINE.** The
   enumeration I then did says the ambiguity branch is on **4 of 8** verifiers, not 2:
   `content_duplication:631` (*"page no longer exists — cannot distinguish a fix from content
   loss"*) and `page_canonical_collision:382` (*"site no longer exists"*) carry it under different
   wording. I had grepped the one spelling I had already read. Corrected in `RFC_017` (visible box),
   `WII-011`, the `LANDMINES` entry and above; logged in `WRONG_CALLS.md`. **The ruling is
   unaffected** — blocking is right for all four — but the reach was twice what I told the owner and
   the council, and I told them a measured-sounding number.
3. **`prior_art_librarian` [medium] — does the restructure break what `verifier_coverage_test.go`
   enforces?** **Answered:** that test walks `RegisteredVerifierItemTypes()`, which I did not touch;
   `RegisterVerifier` still populates `verifiers` (via the policy overload), so its
   *"zero verifiers registered — either init() ordering broke or the registry moved"* fatal cannot
   trip, and the package tests pass.
4. **`prior_art_librarian` + `reuse_agent` + `guardian` — the stale-error-text landmine on
   `failUnverifiedCompletion` / `result._verification`.** **Real, pre-existing, and my change makes
   it MORE visible rather than causing it.** That entry (added today by the
   `bugfix_071_fragment_blindspot` lane) records that the success UPDATE never clears `error`, so a
   row refused and later completed keeps the refusal text — measured there as `status='complete'` +
   `_verification='verified'` + a "completion blocked" `error`, simultaneously. With fail-closed,
   refusals get commoner, so more rows will carry a stale `verification_unavailable` sentence beside
   a later success. **Not fixed here** — it is another lane's filed defect and it lives on the
   success path, outside the ruling I was executing. Flagged to the owner as available: it is
   plausibly one line (`error = NULL` on the success UPDATE).
   Its second half also **independently corroborates my own caveat**: `result._verification` is
   replaced per attempt, so a census of it "counts surviving verdicts, not verifications performed,
   and systematically under-counts exactly the refusals". Two sources, same mechanism, arrived at
   separately on the same day.
5. **`debug_historian` [medium] — no post-roll pod-verification step.** **Answered by the artefact,
   not by the submission:** `WII-011`'s `verify-later` carries the pod-grep with a **negative
   control** (positive `verification could not run, and this item type fails closed` ≥1; negative
   `verifier error — failing open`, the removed string, expect 0) plus the behavioural check. Fair
   objection against what the council could see — the submission JSON did not include it.
6. **`editquality` [low] — the register edit is documentation, shouldn't count as plan substance.**
   Noted, no action: it is a required artefact under the platform-seams ordering exemption, and the
   seat says so itself.
7. **`architecture` — recorded `DEFLECTIONS: unknown`**, having run out of budget before querying
   whether these files had been bounced before. Not an objection; recorded so nobody reads the blank
   as a clean bill.

**The pattern, for the fourth time today:** the thing that caught the error was a check already
written down — this lane's own R8 trap 3 (*"an absence is only ever an absence in the table you
named"*, and its sibling for spellings) — applied by someone else, after the fact.

### 2026-08-08, post-roll — the flip is LIVE on v1.0.1268, and what that does NOT mean

Fleet rolled ~18:57Z. **All 43 chassis-imaged pods are on `v1.0.1268`** — uniform, so today's
mixed-fleet landmine (another lane, same day: *"after a whole-fleet release the fleet is MIXED for
hours"*) has no window here; checked rather than assumed.

**Pod-grep, both replicas, one exec each, positives AND a negative in the same command:**

```
fails closed (RFC_017)            : 1     blocking completion (fail-closed) : 1
failing OPEN by explicit policy   : 1     verification_unavailable          : 1
failing open   (removed spelling) : 0   <- negative control
```

The negative needle discriminates: `failing open` (lowercase) is in the **pre-change source twice**
(`1c5d9ceb5^`, lines 64 and 87) and **zero** times at HEAD; the new code says `failing OPEN by
explicit policy`, so the case change is what makes it a clean 1→0.

> **⚠ Honest limit, and it is my own miss.** The "before" half of that transition is **source-derived,
> not binary-derived** — the fleet is uniformly 1268, so no pre-roll chassis binary was left to read
> the needle at 1 on. **The baseline should have been taken before the roll.** Another lane logged
> precisely this in `WRONG_CALLS.md` today (a needle that read 6 on a binary without the change), and
> I read that entry hours before the roll and still did not take my own baseline.

**What is proven and what is not, kept apart on purpose:**

- **Proven:** the binary carries the change, on every replica.
- **Proven:** the gate runs post-roll — one `hardcoded_section_colors` item verified at **18:58:44Z**,
  63 seconds after the pods came up.
- **NOT proven:** the fail-closed branch has **never executed in production**. That item returned
  `Resolved:true`, so no error was raised. It cannot be forced without the absent-target case
  recurring, and that has happened **twice in the registry's entire life**. Anyone reporting this as
  "behaviourally proven" is repeating the mistake this lane spent a week catching in symptom 2.

**A gap in my own era-marker, found by running the census rather than by thinking about it.**
I added `fail_open` to the payload so a future census could tell a completed error from a blocked
one — but it is written **on the error branch only**. So `verified` and `defect_persists` rows carry
no era marker in *either* era, and the first post-roll row duly has none. My census filter
`result->'_verification' ? 'fail_open'` therefore returned **0 post-roll-shaped rows** on a day when
a post-roll row demonstrably exists — a filter that answers a narrower question than the one I asked
it. Corrected in the `LANDMINES` entry and `WII-011`: date a non-error row by `updated_at` against
the roll, never by payload shape.

---

## 2026-08-08 (late) — picking up §4's "single most valuable next step", and correcting my own handoff first

`HANDOFF_2026-08-08_continue_here.md` §4 names two live pages on site
`1368e337-dd1d-4799-bbb3-8221a1b79bcc` still serving empty sections, and asserts:
*"no `featured-content` item has ever existed fleet-wide although `findEmptySections`' predicate
matches both right now"*, with the whole thing wanting a `090` run.

> **CORRECTED 2026-08-08 — the handoff conflated TWO slot names, and the absence claim is an
> artefact of the spelling it searched.** The slot in `page_components.slot_name` is
> **`featured-content`**. The slot the work items name is **`featured-article`**. They are different
> strings, and items for `featured-article` on exactly these two pages have existed **twice** — an
> April batch that went `unresolved` after 2 attempts, and an 08-03 batch that went **`complete`**.
> Caught by widening the search off `item_key` onto `summary`/`spec`:
>
> ```sql
> SELECT count(*) FILTER (WHERE item_key LIKE '%featured-content%') AS key_hyphen,   -- 0
>        count(*) FILTER (WHERE summary ILIKE '%featured%')          AS summary_hit,  -- 7
>        count(*) FILTER (WHERE spec::text ILIKE '%featured%')       AS spec_hit      -- 10
> FROM site_work_items;
> ```
>
> This is `MEMORY`'s *"a grep proves absence only for the SPELLING it searches"*, committed by me,
> into a handoff, as a fact. The zero was real; it just answered a question nobody asked.

**The site is `finetuning.uk`** (`sites.domain`, status `deployed`) — not an anonymous test site.
Two other lanes hold it (`finetuning_uk_repair`, `finetuning_uk_service`), so nothing gets dispatched
at it from here without asking.

### What is actually true, and it is two separate mechanisms

**(1) The 08-03 items closed wrongly, and it is exactly `bugs_closed/032`.** Both items carry:

```
result->'_verification' = {"status":"error","item_type":"empty_section",
  "error":"cannot verify: component a390860e-… no longer exists
           (genuinely fixed or silently deleted — indistinguishable here)"}
status = complete
```

The fix replaced the component rather than filling it: old components `a553f25f` / `a390860e`
(slot `featured-article`) are gone; a new component `b3e0c2c0` sits in slot **`featured-content`**
on both pages at **334 bytes** — i.e. still empty. The verifier looked for the old id, found nothing,
errored, and **fail-OPEN stamped it `complete`.** These two rows ARE the "fired twice in the
registry's entire life" absent-target cases §3 refers to. Under `RFC_017`'s fail-closed flip
(live on `v1.0.1268`) this same case now lands `triaged`/`failed` instead. `[MEASURED]`

**(2) `featured-content` has never been filed because NOTHING HAS LOOKED SINCE 08-03.** Every
`empty_section` item for this site was created in exactly two batches (`2026-08-03 10:15:22`, seven
items sharing one timestamp; and April). The predicate matches *right now* — I ran
`findEmptySections`' WHERE clause verbatim against the live DB and it returns both pages plus
`tools/tool-list`; `page_type` is `content` (not `blog-index`), `locked_at` is NULL, `build_status`
is `deployed`, and neither slot is in `suppressed_sections`. So detection is not blinded. It is
**undriven**: `[MEASURED]`

```sql
SELECT count(*) FILTER (WHERE enabled) AS enabled_site_discovery,  -- 0
       count(*)                        AS total_site_discovery     -- 5
FROM scheduled_tasks
WHERE target_agent_type IN ('quality-discovery-agent',
                            'completeness-discovery-agent','design-discovery-agent');
```

**All five rows are `oneshot-*` and all five are disabled** — including
`oneshot-completeness-discovery-fai-20260803`, this very site's 08-03 run, switched off after it
fired. No cluster CronJob fires them either (`kubectl get cronjobs -A` — the eight that exist are
cleanup/backup/drift checks). The only enabled rows matching `%discovery%` target
`adoption-researcher` / `directory-researcher`, which are the model-directory research agents, a
different subsystem entirely.

Fleet-wide the shape holds: `empty_section` last filed **08-04 19:36**, while sibling checks in the
same family filed on **08-08** (`hardcoded_section_colors` 18:11, `literal_markdown` 18:09,
`voice_tells` 17:13) — and those three landed on `webdesign.co.uk` and `leopardessconsulting.co.uk`,
the two sites other lanes were hand-driving today. Detection follows attention, not a clock.

> **This CORROBORATES rather than discovers.** `MEMORY` already carries *"a silent mechanism is
> usually UNDRIVEN, not missing"* and *"detection works; SCHEDULE and DISPATCH do not"*. I did not
> derive that from first principles — I checked the schedule because the memory line told me to.
> Recording it as a fresh finding would overstate it.
