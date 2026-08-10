# 243 — the Anthropic account hit its usage limit at 14:51:47Z today; EVERY LLM step fleet-wide now fails, and the API says access returns 2026-09-01

> ## ✅ RESOLVED SAME DAY, 2026-08-10 ~18:12Z — the owner added credit. Candidate 1, exactly as in §5. NOT the calendar.
>
> **This matters because §6 said to record which of the two ways it ended.** It did **not**
> auto-restore on 2026-09-01; **the owner acted**, and the fleet came back **21 days early**.
>
> **The recovery boundary `[MEASURED]`:**
>
> | | time (UTC) | agent |
> |---|---|---|
> | last failure | `2026-08-10 17:02:12.066` | council-gate |
> | first success after | `2026-08-10 18:12:11.065` | council-gate |
>
> **Outage duration: 14:51:47Z → 18:12:11Z ≈ 3h 20m.** Since the cutover: 7 failures across 4
> agent types, then **43 successes across 3** — so the recovery is confirmed by sustained
> traffic, not by a single lucky call. The fleet's own liveness query from §6
> (`SELECT max(created_at) FROM llm_call_log WHERE success`) has moved and keeps moving.
>
> **What this does and does not close.**
> - **The outage is over.** Councils are running; this lane resubmitted on corr
>   `44fa6a98-acaa-46b5-9ada-f0c34ca5475d` immediately after and it dispatched normally.
> - **The bug stays OPEN**, and not as bookkeeping: candidate 1 restores service, it does not
>   prevent recurrence. **This is the second Anthropic exhaustion in eleven days** (07-31,
>   08-10) and the third single-provider outage counting `bugs_open/202`'s Gemini 429 on 08-05.
>   Adding credit is the right emergency action and a poor control.
> - **The actionable prevention now exists and is not mine:** `bugs_open/244` measured that
>   `council-gate` is **87.8% of the fleet's August spend** (165.2M tokens), at **790,551 input
>   tokens per round**, because it sends 11–15 near-identical ~106k prompts with no caching and
>   in an order that defeats caching anyway — with a costed **~76% reduction** available. **Fix
>   that and this bug's trigger largely goes away.** Candidate 2 here (a second provider) is
>   still worth a decision, but 244 is cheaper and closer.
>
> Everything below is the record of the outage as it happened, and is left unedited.

> **NUMBER COLLISION (2026-08-10, same day):** another thread filed a different `243`
> (`tool_acceptance_look_step_has_no_storage_client_on_either_execution_path`). Numbers are
> never reassigned — **resolve by slug**. Cite this case as **243-anthropic-cap**.

> **READ `bugs_open/244` FIRST IF YOU WANT THE CAUSE.** This file records the **outage**: when
> it started, how wide it is, how to tell it apart from queue latency, and what unblocks it.
> `244` — filed the same afternoon by the `bugfix_168_deployed_asset_path` lane — has the
> **mechanism that spent the budget**, and it is far more actionable than anything here:
> the council gate sends 11–15 near-identical ~106k-token prompts per round with no prompt
> caching, and orders the prompt so caching could not help even if enabled. Its measurements:
> **790,551 input tokens per round**, and `council-gate` is **165.2M = 87.8% of the fleet's
> entire August spend**, with a costed ~76% reduction available.
>
> So the two files are complements, not duplicates: **244 is why we ran out; 243 is what
> running out looks like and how to recognise it.** §5 candidate 1 here (owner raises the cap)
> and 244's caching fix are both wanted — the cap gets the fleet working again today, the
> caching fix stops it recurring. **Neither alone is sufficient**, and on the evidence of two
> Anthropic exhaustions in eleven days, raising the cap alone buys weeks rather than a fix.

**Filed:** 2026-08-10 ~17:00 BST by the `bugfix_153_build_provenance` lane, which found it
by having its own council submission die on it. **Status: OPEN, unowned, LIVE RIGHT NOW.**
**Class:** account/billing state, not a code defect — the same family as `bugs_open/202`
(Gemini quota 429 blocking page builds), one provider along.
**Severity:** the fleet's entire LLM capability is down. This is an **owner action**, not a
debugging task — see §5.

> **THIS IS NOT A CODE BUG AND MUST NOT BE "FIXED" IN CODE.** The provider is asserting a
> billing state about our account in its own error body. Per CLAUDE.md's diagnosis norm, a
> `090` run was **not** performed and the reason is stated: the root cause is named by the
> upstream service, there is no mechanism to diagnose. What IS open is a decision (§5), which
> is the owner's.

> ## ⚠ CORRECTED BY THE FILER, within the hour, before anyone read it — THIS IS A RECURRENCE, AND I FILED IT AS IF IT WERE NEW
>
> The original text of this file said the outage was "not filed" and implied a first
> occurrence. **Both readings are wrong in the way that matters.**
>
> - **It has happened before, on 2026-07-31**, same signature, same `review_editquality`
>   seat, stated reset `2026-08-01`. It is recorded in `LANDMINES.md` under *"An API
>   USAGE-LIMIT death looks exactly like a transient seat fault…"* and narrated inside
>   `bugs_closed/130` (which records the resolution: **the owner raised the limit the same
>   day, ~14:50 BST**, and the resubmitted council round was APPROVED). `bugs_closed/128`
>   and `bugs_closed/137` also record being hit by it.
> - **Another lane got here first today.** `bugfix_236_site_availability` had already
>   appended today's recurrence to that landmine — including a detail I did not have and
>   which is the strongest evidence in the case: they reproduced the refusal **from a
>   standalone service outside the cluster** (`6a4fbab21`, the webdesign.uk chat lane), which
>   is what makes it ACCOUNT-level rather than a chassis credential fault.
>
> **How I got it wrong, because the mechanism is the reusable part:** I did grep both bug
> directories, exactly as CLAUDE.md says to. The grep **returned all three closed files**. I
> read three closed-bug hits as coincidental matches and never opened one. So the check ran,
> produced the right answer, and I discarded it — which is worse than not running it, because
> it let me write "not filed" with a clean conscience. Logged in `WRONG_CALLS.md`.
>
> **What survives, and why this file still exists rather than being deleted:** there is
> genuinely no *bug file* for this class — it lives only in a landmine and in other bugs'
> narratives — and `bugs_open/202` sets the precedent that a provider outage blocking the
> fleet gets one. What this file adds over the landmine is the measured cutover (§2), the
> blast-radius census (§3), and the fix candidates (§5). **What it must not do is compete:**
> the landmine is the system of record for the *trap*; this file is the record of the
> *outage*. Contribute to whichever you are actually adding to.
>
> **And the recurrence is the finding.** Twice in eleven days, plus `bugs_open/202`'s Gemini
> 429 five days ago, is not bad luck — it is a single-provider, single-key estate meeting its
> own cap. That is the argument for candidate 2 in §5, and it is much stronger than the
> version I originally wrote, which had one data point.

## 1. The error, verbatim

```
step review_editquality failed: failed to execute action execute_llm_prompt:
AI endpoint unavailable: provider=anthropic model=claude-sonnet-5 endpoint=<nil>:
API request failed with status 400: {"type":"error","error":{"type":"invalid_request_error",
"message":"You have reached your specified API usage limits. You will regain access on
2026-09-01 at 00:00 UTC."},"request_id":"req_011CduR3gLr5nj1nr21HVhun"}
```

Note **400**, not 429. Nothing in our retry/classifier machinery treats a 400 as transient,
which is correct here — retrying cannot help for three weeks — but it means the failure is
**terminal on first contact** and burns the step's attempts immediately.

## 2. The cutover, measured to the second `[MEASURED]`

`llm_call_log` gives an unusually clean boundary — two calls 2.7 seconds apart:

```sql
SELECT created_at, agent_type, success, left(error_message,70)
FROM llm_call_log WHERE created_at > now() - interval '5 hours' ORDER BY created_at DESC;
```

| time (UTC) | agent | success |
|---|---|---|
| 2026-08-10 14:51:45.067 | council-gate | **t** — the last successful LLM call in the fleet |
| 2026-08-10 14:51:47.814 | council-gate | **f** — the first refusal |

Since that boundary, **zero successes**:

```sql
SELECT success, count(*), count(DISTINCT agent_type)
FROM llm_call_log WHERE created_at > '2026-08-10 14:51:46+00' GROUP BY 1;
-- f | 5 | 3      (no 't' row at all)
```

Affected agent types so far: `council-gate` (3), `experience-planner` (1),
`tool-recreation-handler` (1). That list is small only because **fleet LLM volume has
collapsed** — 40 calls in the 14:00 hour, 3 in the 16:00 hour — not because the failure is
selective.

## 3. Blast radius — every LLM step we own `[MEASURED]`

```sql
SELECT s.v->'config'->'ai_service'->>'provider' AS provider,
       count(*) AS steps, count(DISTINCT type) AS agents
FROM agent_definitions, LATERAL jsonb_each(default_config->'workflow'->'steps') s(k,v)
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND s.v->'config' ? 'ai_service' GROUP BY 1;
-- anthropic | 127 | 55
-- (null)    |  11 | 10
```

**127 configured LLM steps across 55 live agents, and every single one is `anthropic`.** All
127 name the same credential, `ANTHROPIC_API_KEY` — there is no second key and no non-Anthropic
step anywhere in live config. The 11 null rows are steps whose `ai_service` block omits a
provider, not steps using a different one.

> **TRAP for anyone re-measuring this.** `provider` is nested inside `ai_service`, NOT a
> direct child of `config`. The obvious query —
> `WHERE v->'config'->>'provider' = 'anthropic'` — returns **0 rows and no error**, which
> reads exactly like "nothing uses Anthropic". I wrote that query first and briefly believed
> it. The correct path is `v->'config'->'ai_service'->>'provider'`. Landmine filed.

The council gate specifically: all **17** of its `review_*`/`gate_*` steps use
`{"model":"claude-sonnet-5","provider":"anthropic","max_tokens":16000|8000,
"api_key_env_var":"ANTHROPIC_API_KEY"}`. There is no seat that could still run.

## 4. What this breaks, and the part that is genuinely dangerous

- **Every page build, content write, classifier, planner and reviewer** — 55 agents.
- **The council gate is DOWN**, which means CLAUDE.md's standing review norm is currently
  unsatisfiable. Every thread that submits from now gets an orchestration that ends
  `current_step='complete_invalid'`.
- **THE FAILURE MODE LOOKS LIKE LATENCY, AND THAT IS THE TRAP.** CLAUDE.md tells submitters to
  budget ~30 minutes and explicitly says *"a missing orchestration row is almost always
  latency, not a dropped dispatch — do not retry on that evidence"*. That advice is correct in
  general and **actively misleading today**: the run is not slow, it is dead, and the thread
  that waits patiently will wait for ever. Three submissions have already died this way:

```sql
SELECT current_step, status, count(*), min(created_at), max(created_at)
FROM orchestration_states WHERE created_at > now() - interval '6 hours'
  AND collected_data->'input_data'->>'fix_correlation_id' IS NOT NULL GROUP BY 1,2;
-- complete_rejected | COMPLETED | 1 | 11:04 | 11:04
-- complete_approved | COMPLETED | 1 | 11:20 | 11:20
-- complete_invalid  | COMPLETED | 3 | 14:42 | 16:51
```

Correlations that died: `b67eb26a-14ef-45d7-b755-3e489fd57ef0` (14:42, `review_architecture`),
`7177fb02-51c5-4c2a-bb02-10aa27ae85ca` (16:50), `44fa6a98-acaa-46b5-9ada-f0c34ca5475d`
(16:51, this lane's). **Note the 14:42 one predates the measured cutover** — its
`review_architecture` step ran later in the workflow than the 14:51 boundary, so it is the
same cause, not a second one.

**How to tell, in one query** — do this before assuming your council run is merely queued:

```sql
SELECT current_step, status, collected_data->'__step_error'->>'message'
FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<YOUR_CORR>';
```
`complete_invalid` + a `usage limits` message = this bug, not your submission. Your `fix_plan`
artifact will have persisted normally, which is why the submission itself looks fine.

## 5. Fix candidates — ordered, but the first is the only real one

1. **Owner raises or resets the Anthropic account spend limit.** The message is
   *"your specified API usage limits"* — a self-imposed cap on our own account, not an
   Anthropic-side suspension, so it is very likely a console setting the owner can change in
   minutes. **This is the whole fix and everything below is contingency.** It needs the
   owner: no thread has console access.
   **PRECEDENT, and it is exact:** on 2026-07-31 this same condition was resolved this same
   way — `bugs_closed/130` records the owner raising the limit ~14:50 BST, the council round
   resubmitted on the same correlation, and an APPROVED verdict at 14:34 UTC. So this is a
   known-good, same-day fix that has already worked once. **The only thing different today is
   the horizon:** July's stated reset was six hours out, so waiting was a real option; this
   one is three weeks out, so it is not.
2. **Add a second provider as a fallback and let the fleet fail over.** The platform already
   supports multiple providers (`bugs_closed/`'s gemini work, and `platform/aiservice` has
   per-provider clients), and `bugs_open/202` records the mirror-image situation with Gemini.
   **But note what §3 measured: 127/127 steps name one provider and one key.** There is no
   fallback configured anywhere, so this is a real build, not a config flip — and doing it
   under an outage, to live config, on a shared tree, is how mistakes ship. Propose it as
   follow-up work, do not improvise it today.
3. **Make the failure legible rather than silent** — the genuinely useful code-adjacent
   follow-up, and the only one this lane would recommend building. A 400 whose body contains
   `usage limits` is a distinct, non-retryable, fleet-scope condition and should be surfaced
   as such: a named error class, and ideally a work item or alert, so the next thread does not
   spend 30 minutes reading a dead run as a slow one (§4). Today the only way to learn this is
   to query `__step_error` by hand, knowing to.
4. **Do NOT** bump retries, widen the transient classifier, or "just resubmit". A 400 with a
   three-week horizon is correctly terminal; making it retryable would burn quota we do not
   have and turn one clean failure into a storm.

## 6. How to verify it is over

```sql
-- the fleet's own liveness signal: any successful LLM call at all
SELECT max(created_at) FROM llm_call_log WHERE success;
```
If that is newer than the moment the owner changed the setting, the fleet is back. A positive
control costs one council submission — but do not spend one until the query above has moved,
because a submission during the outage is a guaranteed `complete_invalid`.

**Do not close this on the calendar.** `2026-09-01` is the API's stated auto-restore date, so
this bug will appear to fix itself in three weeks whether or not anyone acts. If it closes
that way, candidate 1 was never done and the same cap will be hit again — record which of the
two actually happened.

## 7. Related — read these BEFORE adding to this file

**The system of record for the TRAP is the landmine, not this file.**
`LANDMINES.md` → *"An API USAGE-LIMIT death looks exactly like a transient seat fault, and
the runbook's advice for the transient case — 'resubmit unchanged' — is actively wrong for
it"* (filed 2026-07-31, recurrence appended 2026-08-10 by `bugfix_236_site_availability`). It
carries the three-way message triage, the `usage limit` vs `spending limit` needle trap, and
the low-volume measurement trap. If you are adding a *check*, add it there.

Prior occurrences of the same condition:
- **2026-07-31** — first fleet-wide hit; killed a `bugs_open/149` A6 council round at
  `review_tooling_provenance` and another session's at `review_editquality`. Reset was
  2026-08-01, six hours out. **`bugs_closed/130` records the resolution: owner raised the
  limit the same day, resubmit APPROVED.** Also narrated in `bugs_closed/128` and
  `bugs_closed/137`, both of which were merely *interrupted* by it.
- **2026-08-10 (this one)** — reset 2026-09-01, three weeks out.

Sibling provider outage:
- `bugs_open/202` — the Gemini 429 equivalent, filed 2026-08-05: a quota state blocking page
  builds fleet-wide, decision belonging to the owner. **Three single-provider outages in
  eleven days (07-31 Anthropic, 08-05 Gemini, 08-10 Anthropic) is the case for candidate 2**,
  and it is a rate, not an anecdote.

Standing docs whose advice interacts badly with this state:
- CLAUDE.md § "Council review of platform changes" — the ~30-minute latency guidance (§4).
- `RUNBOOK_council_gate.md`'s `complete_invalid` note — written for the transient case, and
  its "resubmit unchanged" is wrong here; the landmine above is the counter-example.
