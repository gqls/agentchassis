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

---

## RECURRENCE — 2026-08-14 15:35:51Z. Third exhaustion in 15 days. (contributed by the bugfix_209/231 lane)

Same signature, same message, same stated reset date:

> `API request failed with status 400: {"type":"error","error":{"type":"invalid_request_error",
> "message":"You have reached your specified API usage limits. You will regain access on
> 2026-09-01 at 00:00 UTC."}}`

**Boundary `[MEASURED 2026-08-14 16:40Z]`, using this file's own §6 liveness query:**

| | time (UTC) |
|---|---|
| last SUCCESSFUL llm call | `2026-08-14 15:35:51.922` |
| first usage-limit failure | `2026-08-14 15:36:30.499` |
| state at 16:40Z | **0 ok / 10 failed** in the preceding 20 minutes |

Sustained, not a blip: **zero** successful calls in the ~65 minutes since onset, across 5
distinct agent types.

⚠ **A CORRECTION I MADE MID-DIAGNOSIS, because the shape is worth knowing.** A
`GROUP BY model_resolved` over "the last 90 minutes" showed `claude-sonnet-4-6` 9 ok / 18
failed and `claude-sonnet-5` 21 ok / 9 failed, and I briefly read that as *successes
interleaved with failures, so the cap trips and releases rather than blocking*. **Wrong:
every one of those successes predates 15:36Z.** The window straddled the onset, so a
per-model split across it mixes two regimes and manufactures an interleaving that does not
exist. **Bucket by MINUTE and check `max(created_at) WHERE success` before reading any
split that spans an incident boundary.**

**This lane contributed to the exhaustion and should say so.** I ran **three council
rounds** on corr `41a01378` today (231 candidate 2). Fleet-wide `council-gate` spend on
2026-08-14 up to onset: **143 calls, 525,277 input tokens.** `bugs_open/244`'s measurement
— council-gate at 87.8% of August spend, ~790k input tokens per round — is the mechanism,
and my rounds are three instances of it. Round 3 died at
`review_editquality` on this cap, so **the round is lost and cannot be resubmitted until
service returns**; the submission itself was valid (it had passed `persist_submission`).

**What this changes about this file's conclusion: nothing, and that is the point.** §5's
candidate 1 (owner adds credit) is again the only thing that restores service, and it is
again not a control. **Third occurrence in 15 days (07-31, 08-10, 08-14)** — the interval
is shortening, and `bugs_open/244`'s caching fix is the prevention that keeps not landing.

**For the next session that finds its dispatch silently stuck:** this is what it looks
like. An orchestration terminates at `complete_invalid` with no `council_report` artifact,
and `collected_data->'__step_error'->>'failed_step'` names a *review seat*, not your
submission. **Do not edit your JSON** — the runbook's warning applies exactly. Check
liveness first:
```sql
SELECT max(created_at) FROM llm_call_log WHERE success;   -- more than ~2 min stale ⇒ this bug
```

---

## ADDENDUM 2026-08-17 (webdesign_uk_build_service lane) — the cap RECURRED, and it stops the BUILD QUEUE by a path this file does not describe: one sampled 400 wedges `claim_work_item` for a FULL HOUR while the fleet's own liveness query says healthy

**Contributed, not competing** — 243 is the cap; this is a consequence of the cap that
nothing in 243, 244 or any other bug file records (`grep -rln ai_endpoint_health
bugs_open/ bugs_closed/` → one CLOSED file, 030). Found because it blocked a
webdesign.uk acceptance run, not by looking for it.

**The mechanism, end to end, all `[MEASURED]` 2026-08-17 11:42–11:58Z:**

1. `ai_endpoint_health` holds one row per endpoint. The `claude` row
   (`https://api.anthropic.com/v1/messages`) went `healthy=false` at **11:09:53**,
   `last_healthy 11:07:15`, error = the same *"You have reached your specified API usage
   limits … regain access on 2026-09-01"* 400 this bug is about.
2. `claim_work_item_action.go:~218-255` gates **every work-item claim fleet-wide** on that
   row: handler's endpoint unhealthy → the claim is RELEASED and the item set back to
   `triaged`. Returns `{"claimed": false, "reason": "ai_endpoint_unavailable"}`.
3. So `build-dispatch-loop` runs to COMPLETION every ~90s, loads the same item
   (`pending.has_items = true`, `items[0].id` unchanged run after run), fails the claim,
   and completes. Path taken, read from `collected_data` keys:
   `claim → check_claim → done`; `spawn_handler` never reached.
4. `build-pipeline-trigger`'s `find_dispatchable_site` is
   `ORDER BY wi.created_at ASC … LIMIT 1` across **all sites**, so the whole fleet queues
   behind the single oldest item, which can never clear. **Zero claims fleet-wide since
   10:32:33**; four sites' items sat `triaged` (webdesign.co.uk, leopardessconsulting.co.uk,
   dartsonline.com, webdesign.uk).
5. **`check_interval_seconds` on that row is 3600.** The endpoint is re-probed ONCE AN HOUR.
   So a single sampled 400 stops all build dispatch for **up to an hour after real service
   has recovered**.

**The part that makes it a trap rather than merely a wedge — the fleet looks HEALTHY the whole
time.** While dispatch was fully stopped, real Anthropic traffic was succeeding:
**93 of 99 calls OK in the last 2 hours, latest 11:52:32Z** (`claude-sonnet-5` 66/67,
`claude-opus-4-6` 27/30) — i.e. *after* the health row went false. **243 §6's own liveness
query (`SELECT max(created_at) FROM llm_call_log WHERE success`) therefore reports the fleet
UP while nothing can be claimed.** Two true facts that read as contradictory:

- the cap is real and intermittent — `usage limit` errors in the last 24h: 3 on
  `claude-opus-4-6`, 1 on `claude-sonnet-5`;
- most calls still succeed, so the fleet is *not* down in the 08-10 sense.

The health probe samples ONE call (`check_endpoint_health_action.go:203`,
`claude-haiku-4-5-20251001`, `max_tokens 1`). Under an intermittent cap, a single sample is a
coin flip, and losing it costs an hour of fleet dispatch. **NOTE the probe uses haiku while the
fleet's live traffic is sonnet/opus** — `llm_call_log` holds **zero** haiku rows in 24h, so the
probe's model is exercised by nothing else and its result cannot be cross-checked from that
table. Whether the cap is account-wide or per-model is **[UNMEASURED]** here and matters: if
per-model, this is not transient at all.

**How to tell this apart from ordinary queue latency** (the question 243 §6 answers for LLM
steps; this is the dispatch equivalent):
```sql
SELECT healthy, last_checked, last_healthy, check_interval_seconds, left(error,120)
  FROM ai_endpoint_health WHERE name = 'claude';
SELECT max(claimed_at) FROM site_work_items WHERE claimed_by = 'build-dispatch-loop';
```
`healthy=f` + no claim since roughly `last_checked` = this. **A dispatch stop is invisible in
`llm_call_log`** — do not conclude from succeeding LLM calls that work is flowing.

**NOT fixed here, and deliberately not:** `check_endpoint_health_action.go` is **dirty in the
working tree right now** (another session, on an unrelated `CheckConfig`/input-spec concern),
so editing it would make me a same-file passenger on their commit. Recorded for whoever owns
this surface. Shapes worth considering, in the order they close the door: require N
consecutive failures before marking unhealthy (a single sample should not gate a fleet);
re-probe on a short interval **while unhealthy** rather than the healthy-state 3600s; and let
`claim_work_item` distinguish "endpoint capped" from "endpoint unreachable" — under a cap,
letting the claim through costs one failed step, while blocking it costs the whole queue.

**Falsifier for this addendum:** if the 12:09:53Z probe succeeds and claims resume within
minutes, the "up to an hour" figure is confirmed as the *upper* bound of one cycle, not a
permanent stall. If it fails again and dispatch stays stopped, the cap is biting the probe's
model specifically and this is the more serious per-model case.

### Falsifier resolved, 2026-08-17 12:10:18Z — the FIRST branch: one re-check cycle, not a permanent stall

The addendum above named two outcomes. The measured one is the first: the
**12:09:53Z-due probe ran at 12:10:18Z and returned `healthy=true`**, so the cap is
intermittent and its bite on the probe was a lost coin-flip, not a per-model block.

**What that confirms:** the "up to an hour" figure is the **upper bound of ONE re-check
cycle**, and this instance ran nearly the whole of it — `healthy=false` from **11:09:53Z**
to **12:10:18Z**, i.e. **60 minutes 25 seconds of fleet-wide dispatch stop from a single
sampled 400**, during which Anthropic traffic kept succeeding throughout.

**What it does NOT close.** The severity is unchanged and arguably worse for being
self-healing: nothing alerted, nothing failed, no item was marked failed, and the only
durable trace is `claim_result.reason` inside orchestration `collected_data`. An hour of
whole-fleet dispatch loss that leaves no error row is precisely the shape that goes
unnoticed until someone happens to be watching a specific item — which is how this was
found. The three fix shapes in the addendum stand, and the cheapest (require N consecutive
probe failures before marking unhealthy) would have prevented this instance outright.

**Still [UNMEASURED]:** whether the cap is account-wide or per-model. This round is
consistent with either — the probe's `claude-haiku-4-5-20251001` succeeded on the retry, but
`llm_call_log` still holds zero haiku rows in 24h, so the probe's model remains exercised by
nothing but the probe.


---

## Third occurrence, 2026-08-17 — and it lasted THREE MINUTES, with the same message

Contributed by the `bugfix_281`/loop-engine lane, which hit this, **did not read this file**, and
reported a 15-day fleet outage to the owner off the back of the API's message text
(`WRONG_CALLS.md` 2026-08-17). Recording it here because it refines what this file already says.

**The measurement `[MEASURED 2026-08-17 16:35Z]`:**

| | time (UTC) | agent |
|---|---|---|
| last success before | `11:08:03.753` | council-gate |
| first failure | `11:08:37.083` | council-gate |
| last failure | `11:09:53.482` | landmine-verifier |
| first success after | `11:13:02.200` | landmine-verifier |

**Outage ≈ 3 minutes** (4 failed calls in 76 s), across 2 agent types and 2 models
(`claude-sonnet-5`, `claude-opus-4-6`) — so *width* looked exactly like the 08-10 event while
*duration* was two orders of magnitude smaller. Recovery is confirmed by sustained traffic, not one
call: hourly `ok/failed` for the day — 11:00 **101/4**, 12:00 **255/1**, 13:00 **99/0**, 14:00
**23/0**, 15:00 **3/0**, 16:00 **55/1**; **462 successes in the five hours after**.

**The refinement this adds to §the-message-is-not-the-calendar.** This file already establishes
that the "regain access 2026-09-01" text does not predict when access returns. 2026-08-17 shows
something stronger and more dangerous: **the same message accompanies outages of wildly different
severity**, so the text carries no information about duration *at all* — not even order of
magnitude. A 3-minute blip and a 3h20m exhaustion are indistinguishable at the moment of failure.
**Therefore the only sound reading of a usage-limit 400 is "right now, calls are failing".**
Anything about *how long* requires a second observation separated in time.

**Practical consequence for the next lane, since this is the third time:** do not announce an
outage, and above all do not defer work on account of one, until you have re-run this file's own
liveness query at least a few minutes later:
```sql
SELECT date_trunc('hour', created_at) AS hr,
       count(*) FILTER (WHERE success) AS ok,
       count(*) FILTER (WHERE NOT success) AS failed
FROM llm_call_log WHERE created_at > now() - interval '8 hours' GROUP BY 1 ORDER BY 1;
```
The histogram distinguishes a blip from an exhaustion on sight; `max(created_at) WHERE success`
alone does not tell you whether the failures are still arriving.

**Does not reopen or close anything here.** The bug's own point stands unchanged and this is
evidence for it: adding credit restores service and prevents nothing, and `bugs_open/244` remains
the actionable prevention.

---

## CONTRIBUTION 2026-08-19, from the `bugfix_302_design_repair_verification` lane — this bug's cost on the COUNCIL GATE is roughly a coin-flip per round, and the reason is one config line ×17

Not a new instance; a measured consequence on a specific consumer, contributed rather than
filed separately.

**What happened.** Two council submissions ended `complete_invalid` within minutes of each other,
neither reviewed. The step error is this bug verbatim: *"step review_editquality failed: failed to
execute action execute_llm_prompt: AI endpoint unavailable … status 400 … You have reached your
specified API usage limits. You will regain access on 2026-09-01"*.

**Read correctly, per this file's own §the-message-is-not-the-calendar** — which stopped me
reporting a fleet outage. [MEASURED 2026-08-19 10:33Z] `llm_call_log` for the current hour reads
**57 ok / 7 failed**, and over three hours only **5** of the failures are the cap. **The fleet is
not down; the cap is intermittent at roughly 5% of calls.** The September date carried, as this
file establishes, no information at all.

**The finding: the council gate has no tolerance for a per-seat transient, and that is a config
line repeated 17 times.**

```sql
SELECT COALESCE(v->'config'->>'error_step','(none)'), count(*)
FROM agent_definitions, LATERAL jsonb_each(default_config #> '{workflow,steps}') AS s(k,v)
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND (k LIKE 'review_%' OR k LIKE 'gate_%') GROUP BY 1;
--  complete_invalid | 17
--  (none)           | 12
```

⚠ **Note the nesting, because I got it wrong first time**: `error_step` sits inside `config`, not at
the step level. Read at the step level the same query returns "(none) | 29" — a clean, confident,
completely wrong answer that would have hidden this.

**The asymmetry is the point.** A seat that returns *unparseable* output is tolerated — the verdict
metadata carries `unreadable` and `abstained` counters for exactly that, and a round with 7
abstentions still decided. A seat whose LLM call *errors* takes `error_step: complete_invalid` and
**discards the entire round**, including the 9 seats that already answered. Garbage is survivable;
a transient is fatal.

**The arithmetic, and its limits.** A round fires ~10 relevance-gated seats, each making at least
one call. At the measured ~5% per-call cap rate, P(at least one seat 400s) ≈ 1 − 0.95¹⁰ ≈ **40%**.
[MEASURED, n=4 — a small sample and stated as such] of this lane's four rounds over two days, two
completed (one APPROVED, one REVISE) and two died `complete_invalid`. Consistent with ~40%, and not
evidence for a precise rate.

**Why it matters beyond annoyance:** a discarded round costs the full credit spend of every seat
that had already answered, and it is indistinguishable at a glance from a submission being
*rejected* — the orchestration reaches `complete_invalid`, which reads like a verdict and is not
one. **A `complete_invalid` with a 400 in `__step_error` is a TRANSIENT: resubmit.** That is the
opposite of the standing advice for a missing orchestration row (which is latency — do not retry),
so the two cases need telling apart before acting:

```sql
SELECT collected_data->>'__step_error' FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

**Cheapest fix, in this file's existing idiom:** point the seats' `error_step` at a
tolerate-this-seat path rather than `complete_invalid`, so a transient costs one seat's opinion
instead of the round — the machinery already exists, since abstention is a first-class outcome the
decision rule handles. A retry on the seat would also work and is more expensive. Both are config,
not code.

⚠ **And one trap for whoever fixes it:** resubmitting without `RESUBMIT_CORR` mints a NEW submission
correlation, which silently orphans any commit already carrying the old one in a
`Council-Submitted:` trailer — the `098` report joins on that id and the old run has no verdict to
resolve to, for ever. I did this on one of the two resubmissions and had to record the new
correlation in a follow-up commit.

---

## ⚠ IT HAS HAPPENED AGAIN — THIRD OCCURRENCE, ONSET 2026-08-22 18:15:35Z

Observed by the `bugfix_305_negation_gate` lane at 18:27Z, not by any alert — found while chasing why
a page build failed.

**The boundary, measured either side** `[MEASURED 2026-08-22 18:27Z]`:

| window | successful LLM calls | usage-limit failures |
|---|---|---|
| 17:15:00 → 18:15:51 | **116** | 3 (the first arrive at 18:15:35, at the edge) |
| after 18:15:51 | **1** | **8** (and 0 failures of any other kind) |

**Last successful call of any kind: `18:15:51`.** Same 400 as before, verbatim: *"You have reached
your specified API usage limits. You will regain access on 2026-09-01 at 00:00 UTC."* Agents hit so
far: `council-gate`, `landmine-verifier`, `page-content-writer`.

⚠ **HOW I NEARLY MISREAD IT, because the next person will hit the same trap.** My first look used a
ONE-HOUR bucket and reported "84 successes alongside 7 failures — intermittent, not a wall". That
window STRADDLED the cutover: every one of those successes predates 18:15:51. **An hourly bucket
cannot see an outage that started inside it** — split at the last success and count both sides, which
takes one query and gives an unambiguous answer. The same shape as this file's own §"how to tell it
apart from queue latency".

**This is the third exhaustion in 22 days** (07-31, 08-10, 08-22), which strengthens rather than
changes this file's standing conclusion: adding credit is the right emergency action and a poor
control, and `bugs_open/244` (council-gate = 87.8% of August spend, ~76% reduction costed) is still
the cheaper prevention. Note `council-gate` is again among the first agents to hit it.

**Consequence for this lane, recorded so it is not re-diagnosed as a gate fault:** no `copy_gate` run
has occurred on chassis `v1.0.1326` (deployed 15:10Z) — page builds now fail at
`generate_content`, upstream of the gate. `ai-agent-orchestration.com/adoption-tracker`, **one of the
three pages `bugs_open/305` is about**, failed at 18:22:02 on exactly this
(orchestration `4774680e-d16e-4559-9dca-54ecdfe56eeb`). The gate is binary-probed present on both
replicas; it simply has not been reached.

**Recovery check (from §6, unchanged):** `SELECT max(created_at) FROM llm_call_log WHERE success;`
— it must move, and keep moving, before declaring it over. Last time the owner added credit and the
fleet came back in ≈3h20m, 21 days before the stated calendar date.

---

## 2026-08-24 — THIS FILE NOW HAS AN OWNER, and three of its standing claims need correcting

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_243_provider_cap_resilience/`
(opened today, commit `9dd907e20`). Taken up because the file says "OPEN, unowned" and six
lanes have contributed occurrences without any of them owning the *platform's response* to the
condition. The cap itself remains the owner's to clear; nothing in this lane touches it.

### 1. The rate is roughly double what this file's narrative implies

This file counts **narrated incidents** — the occasions on which somebody happened to be
present. Counting **days on which the refusal actually arrived** `[MEASURED 2026-08-24]`:

| day | cap failures | ok | all failures |
|---|---|---|---|
| 08-10 | 7 | 457 | 7 |
| 08-14 | 28 | 751 | 38 |
| 08-17 | 4 | 867 | 6 |
| 08-19 | 5 | 914 | 13 |
| 08-21 | 3 | 1223 | 6 |
| 08-22 | **113** | 1063 | 116 |
| 08-23 | 32 | 1109 | 36 |
| 08-24 | 0 | 750 | 2 |

**Seven of the last fifteen days**, not three occurrences in 22. On 08-22 the cap was 113 of
the 116 failures of *any kind* that day. The query is in the lane's RUNBOOK §3.

Also: the refusal always quotes a reset on the **1st**, so this is a **monthly** limit, and the
events cluster toward month-end as spend accumulates `[INFERRED from the reset-date shape plus
the clustering, not from any billing API]`. It makes a falsifiable prediction — **expect
recurrence between now and 08-31**, and expect early September to be quiet regardless of what
anyone fixes.

### 2. ⚠ `bugs_open/244` IS FIXED AND LIVE — and it did NOT prevent this

This file tells the reader in **four** places that 244 is "the actionable prevention", "the
cheaper prevention", and that fixing it makes this bug's trigger "largely go away". **244 has
been fixed and live since 2026-08-10 evening** (its own banner; `3d6851d9b`, `071adc44c`,
migration 376). It worked, and it was not enough:

| day | full-price input | cache read | cache write |
|---|---|---|---|
| 08-09 (pre-cache) | 42,716,865 | 0 | 0 |
| 08-21 | 5,708,697 | 121,127,386 | 12,338,987 |
| 08-22 | 6,664,088 | 102,528,376 | 9,972,791 |

Full-price input fell ~7×. But total prompt volume grew several-fold, so on the standard
weighting (reads 0.1×, writes 1.25×) **effective input spend fell only ~30%** — and **the two
worst cap days in the entire record are after the fix.** So: the prevention landed and the bug
got worse. **Anyone reading this file to decide what to do next should stop waiting for 244.**

### 3. ⚠ CORRECTION to the 08-17 "Falsifier resolved" section — the probe cannot see this condition

That section poses two branches — (a) intermittent cap, the probe lost a coin-flip; (b) the cap
bites the probe's model specifically — and resolves to (a) on the evidence that *"the
12:09:53Z-due probe ran at 12:10:18Z and returned `healthy=true`"*.

**That evidence cannot discriminate between the branches.** `ai_endpoint_health.claude` has
`ping_path='claude_ping'`, so the probe runs `pingClaude`, whose status switch
(`check_endpoint_health_action.go:220-231`) is: 200→true, 402→false, 401→false, 529→true,
**default→true** (*"any non-auth error means API is reachable"*). **The cap is a 400** — this
file's own §1 emphasises that. So a capped probe and a healthy probe write the identical row.
**The probe cannot lose that coin-flip**, because it cannot fail on this condition at all.

The conclusion may still be true — the *other* evidence in the same addendum (93 of 99 live
calls succeeding) does support an intermittent cap — but the probe result contributes nothing
to it, and the "lost coin-flip" framing is wrong. Two consequences:

- **The probe is a TIMER, not a health check**, for this condition: it clears the flag on its
  next tick whether or not the provider is still refusing us. That is why the wedge is bounded
  by `check_interval_seconds`, not by the outage.
- **The addendum's first suggested fix — "require N consecutive failures before marking
  unhealthy" — is aimed at the wrong place.** The probe is not the writer of `false`. Live
  traffic is, via the single `update_endpoint_health` caller at `ai_actions.go:634`. Building
  that suggestion as written would harden a path that never fires for this condition.

### 4. The asymmetry that makes seam A worse than the addendum states

`grep -rn "update_endpoint_health" --include=*.go platform/ internal/ pkg/` → **exactly one
hit**, `ai_actions.go:634`, which **only ever passes `false`**. No SQL trigger on the table
(`pg_trigger` → 0 rows) and it is the only `pg_proc` mentioning it. The **sole** writer of
`healthy=true` is the probe (`check_endpoint_health_action.go:138`).

**So live traffic can mark the fleet's claim gate DOWN, and nothing that succeeds can mark it
UP.** The 08-17 addendum attributes the wedge to the probe's lost sample; the more common
trigger is ordinary traffic, and the recovery path is unrelated to the failure that set it.
`check_interval_seconds` for claude is **still 3600 today** — this is unfixed.

### 5. Submitted, with the fix shaped around one non-obvious hazard

Council `SUBMISSION_CORR = 82f07fa6-1c42-46ad-bdf6-1d58892c44a7` (2026-08-24). Four edits:
a symmetric health writer on the success path; per-step failure records in `routeToErrorStep`;
the council classifying an errored seat as **unreadable**; and a `_HOLD` migration repointing
the 17 seats' `error_step` to their own `next_step` and cutting the claude probe interval
3600→60.

**The hazard, recorded here because it is the trap in the obvious fix.** The 08-19 contribution
proposes pointing the seats' `error_step` "at a tolerate-this-seat path". Doing *only* that
makes the failed seat's field **absent**, and `diagnose_council_decide` counts an absent field
as an **abstention**. Its own comment (`:311-318`) says why that is wrong: *"an abstention is a
seat the relevance filter skipped, which is information; an unreadable seat is an opinion we
were owed and lost… Conflating them would let a lost opinion read as a considered
non-objection."* An `unreadable` seat downgrades an approval to REVISE (`:460`); an abstention
does not. **So the config-only fix trades "the round dies" for "the round can APPROVE with a
seat we never heard from, silently"** — worse, because the first failure is loud. The Go half
must be live before the config half is applied; hence `_HOLD`.

### 6. What a 090 on this returns, so the next lane does not spend one

Run `6c834cc7-de0d-4b1f-b283-d6b82b8dffda`: **UNVERIFIABLE**, stopped on its **iteration cap**,
5 bundles, **0 verdict artifacts**. Neither confirmed nor refuted. Note the LANDMINES
discriminator (the body-omission line) came back **clean on all five bundles** — this is a
third no-verdict shape, and the conclusion lives on `site_work_items.result`, not in
`diagnosis_artifacts`. Declared substitute taken per the owner ruling of 2026-07-31; every link
cited first-hand in the lane's NOTES.

### 7. Still not fixed here, and still the owner's

Candidate 1 (add credit) remains the only thing that restores service, and candidate 2 (a
second provider) remains undecided — **127 of 127 configured LLM steps across 55 live agents
name `anthropic` and the same key**, so it is a real build, not a config flip. What this lane
changes is only how much damage each refusal does on our side.

**And add to §5 candidate 1 the finding that cost ~16 hours on 08-22/23:** the fleet's key is
**not** on the Anthropic org that `platform.claude.com` opens by default for the owner's
address. Billing read `$0.00 spent / 0% used` against a $2,000 limit *while* the API was
refusing. The decisive column is **Organization settings → API keys → `Last used`** — a live
key can never read "30+ days ago", because a failed call is still a use. Full signature:
`~/.claude/projects/*/memory/the-fleet-key-is-not-on-the-default-console-org.md`.

---

## 2026-08-24 (evening) — TWO OF THE THREE FIXES ARE LIVE, ONE IS PROVEN, AND THE PROBE INTERVAL IS APPLIED

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_243_provider_cap_resilience/`.
Council `82f07fa6-1c42-46ad-bdf6-1d58892c44a7` **APPROVED round 1** (4 advisory objections,
none high; two of them changed the code).

### Live on chassis `v1.0.1334`, capability-probed on BOTH replicas with both controls

The `build provenance` startup line had already scrolled out of `--tail=3000` — the documented
case, where an empty result means "not in range", not "unstamped" — so presence was established
at the binary, which has no shelf life:

| probe | fr8dn | xl2zk |
|---|---|---|
| `failed to clear endpoint health after a successful call` (**MDL-044**) | 1 | 1 |
| `recorded as unreadable, NOT abstained` (council reader) | 1 | 1 |
| `diagnose_council_decide` (present-control) | 15 | 15 |
| a deliberately absent string (absent-control) | 0 | 0 |

### MDL-044 is PROVEN on live traffic `[MEASURED 16:51:32 → 16:52:34 Z]`

Forced `healthy=false` by hand; **one successful call cleared it in ~39s.** Attribution rests
on four independent grounds and the disconfirming result was available:

- **`last_healthy` (16:52:34.402) is 61 seconds LATER than `last_checked` (16:51:33.504)** — an
  ordering **no other writer can produce**, because the prober and `update_endpoint_health`
  both assign the two columns in a single `NOW()`;
- `last_checked` never moved from the forced write, so **no probe ran** — the prober cannot be
  the cause;
- `error` went to `NULL`, which is this writer's own statement;
- the **demand control** read exactly 1 successful call, in the same query as the row.

Had the prober done it, `last_checked` would have moved and `last_healthy` would equal it.
Observed a **second** time independently at 16:54:11 on a `council-gate` call.

**So the specific damage this bug's 08-17 addendum measured — up to an hour of fleet-wide
dispatch loss from one refusal — is closed at the mechanism, not merely shortened.**

### Migration applied: the probe interval, 3600s → 60s

`596_claude_probe_interval_60s.sql`, applied by hand on the owner's instruction and recorded
`--record-only` (**not** `--apply`, which takes every pending file fleet-wide). Split out of
`588_..._HOLD.sql` because it has no dependency on the Go half that file waits for.

⚠ **CORRECTED before it could be quoted: this does NOT give a one-minute bound.** The probe
fires only when the scheduled task `ai-endpoint-health-check` ticks **and** the endpoint's own
interval has elapsed — and that task's `interval_seconds` is **also 60**, so the two compose.
Measured ticks 16:22:38 → 16:24:12 → 16:25:44: gaps of **94s and 92s**. The honest bound is
**one to two minutes**, still ~39× better than 3600s. If a tighter one is wanted the lever is
this row at 30s against the task's 60s tick, not lowering the row alone.

### ⚠ STILL HELD: the council half. Do not apply `588`.

The `__step_errors` **writer** is not in the build — probed, not assumed, on both replicas:
`step-error record capped at` = **0**. Its reader IS live and **inert by design** (fails closed
with no key to read).

**Probe the writer's own literal, never `__step_errors`** — the reader mentions that key too,
so grepping it returns **1** either way and reads as if the writer landed.

The writer is ~20 lines in `routeToErrorStep`, written and green against HEAD in isolation, and
**deliberately uncommitted**: `coordinator.go` carries the **`bugs_open/354`** lane's
uncommitted call to `errorRouteTermination`, whose definition is in an **untracked** file. A
pathspec commit takes the whole file, so committing it would break HEAD for the estate. Note in
`bugs_open/354` with two ways out. **Any session that commits `coordinator.go` while that
stands breaks HEAD**, and nothing at commit time would warn them.

### What is still the owner's, unchanged

Candidate 1 (add credit) is still the only thing that restores service under a real cap, and
candidate 2 (a second provider) is still undecided — **127 of 127 configured LLM steps across
55 live agents name `anthropic` and the same key.** This lane only reduces what each refusal
costs us.

---

## 2026-08-25 — ALL FOUR PLATFORM FIXES ARE NOW LIVE. The lane is complete; this bug stays OPEN, and §6 says why.

**Cold-start for this work:**
`docs/agent_docs/docs024_key_docs_latest/bugfix_243_provider_cap_resilience/HANDOFF_2026-08-25_continue_here.md`

| fix | state |
|---|---|
| **MDL-044** — a successful live call clears `ai_endpoint_health` | **LIVE + PROVEN TWICE** (v1.0.1334) |
| **mig 596** — claude re-probe 3600s → 60s | **APPLIED** 08-24 (real cadence 92–94s, see below) |
| **WFA-023** — `__step_errors` + the council classifying an errored seat `unreadable` | **LIVE** (v1.0.1337) |
| **mig 588** — the 17 seats' `error_step` → their own `next_step` | **APPLIED** 08-25 |

Council `82f07fa6-1c42-46ad-bdf6-1d58892c44a7`, **APPROVED round 1**; every commit carries the
trailer.

**So the two costs this file documented are now closed at the mechanism**: a refusal no longer
pins the fleet's dispatch gate (a successful call clears it in seconds — proven, twice), and a
seat's transient no longer discards a whole council round.

### The gate for 588, and how it was checked

Ancestry against the running binary's **own** provenance stamp —
`git merge-base --is-ancestor dbd865ee8 4c996e1b5` → IN — plus a binary probe on **both**
replicas with **both** controls (`step-error record capped at` = 1/1; known-present = 15/15;
known-absent = 0/0). ⚠ `__step_errors` also reads 1 and **is not evidence**: the reader mentions
that key too, so it says "landed" either way.

### ⚠ Two corrections to things this lane itself wrote

1. **mig 596 does NOT give a one-minute bound.** The probe needs the `ai-endpoint-health-check`
   task to tick **and** the endpoint interval to elapse, and that task is **also 60s**, so they
   compose. Measured ticks: gaps of **94s and 92s**. Honest bound: **one to two minutes** —
   still ~39× better than 3600s. Corrected in the migration header before it could be quoted.
2. **mig 588's first apply FAILED and rolled back.** The council-reviewed *sketch* used
   `UPDATE … FROM LATERAL jsonb_each(ad.default_config …)`, which Postgres refuses (the UPDATE
   target cannot be referenced from a `LATERAL` in its own `FROM`). Nothing changed — verified
   19 seats still `complete_invalid` afterwards. Rewritten to a correlated scalar subquery.
   **A council seat reviews a sketch, which is not executable; this class cannot be caught at
   the gate.** LANDMINE filed.

### What is proven and what is not

**Discharged (negative control)** `[MEASURED 2026-08-25 09:49:00Z]`: the first council round
after the migration reached **`complete_approved`** — approved, 5 abstained, **0 unreadable**.
Repointing all 17 `error_step`s did **not** break the ordinary path.

**Still owed (positive arm), and it cannot be forced:** a round in which a seat's call *errors*
must reach a verdict, report that seat under **`unreadable`** (not `abstained`), and return
**REVISE** if the rest would have approved.

```sql
SELECT created_at, metadata->>'decision', metadata->>'unreadable', metadata->'unreadable_at'
FROM diagnosis_artifacts WHERE kind='council_report'
  AND (metadata->>'unreadable')::int > 0 ORDER BY created_at DESC LIMIT 5;
```

It will arrive on its own — cap failures hit **7 of the 15 days** to 08-24. **Do not fake it.**

### Why this bug stays OPEN

Not bookkeeping. **Nothing in this lane fixes the cap**, which is what this file is actually
about. §5 candidate 1 (the owner adds credit) is still the only thing that restores service, and
**candidate 2 (a second provider) is still an open owner decision** — `127 of 127` configured LLM
steps across 55 live agents name `anthropic` and the same key (⚠ census dated 2026-08-10;
re-run before quoting). Closing this would read as *"the running-out-of-credit problem is
solved"*, which is false. **What is solved is how much damage each refusal does to us.**

**Fair close condition:** one council round with `unreadable > 0` reaching a verdict, **plus** an
owner decision on candidate 2.

### Addendum, same day (2026-08-25 ~19:15Z) — post-roll re-verify, and the trigger has now been absent for 48h

Chassis rolled again to **v1.0.1339**. Re-checked rather than assumed (a roll can regress Go via
a build from an older ref, and a seed can overwrite DB config): ancestry against the running
binary's own stamp `a7459a44b` → **IN** for all three commits; DB config survived — probe
interval **60s**, seats repointed **17/17**, `complete_invalid` still exactly **2**.

**47 council rounds since mig 588 applied, ZERO `complete_invalid`** (27 approved, 20 revise).
⚠ **Not the proof**: `cap = 0` on both 08-24 and 08-25, so no transient has exercised the new
path, and all 47 report `unreadable = 0`. **The zero means "nothing tested it".** What it does
establish is **no regression on the ordinary path**, which was the real risk of repointing 17
`error_step`s.

**The trigger itself has been quiet for 48 hours, at the two highest call volumes in the record:**

| day | cap | ok |
|---|---|---|
| 08-22 | **113** | 1,063 |
| 08-23 | 32 | 1,109 |
| 08-24 | **0** | 1,850 |
| 08-25 | **0** | 1,395 |

Most likely the account was properly funded once the **wrong-account** error was found on 08-23
(§ the 2026-08-24 section). **This does NOT close the bug**: two days cannot establish stability,
six days of the month remain, and §5 candidate 1 restoring service has always been the pattern —
it is what happened on 08-10 too. Recorded as **unresolved, not fixed**. Re-run the histogram
before repeating any of these figures.

---

## 🔴 LIVE NOW — 2026-08-26: A NEW VARIANT. Not the usage cap — **CREDIT BALANCE EXHAUSTED**. Fleet LLM capability has been down ~9 hours.

**Found at 08:50Z** by this lane, while checking a council verdict — **not by any alert.** Same
family as this file (the provider refusing us on a billing state, HTTP 400), **different
condition, different message, and a different owner action.**

```
API request failed with status 400: {"type":"error","error":{"type":"invalid_request_error",
"message":"Your credit balance is too low to access the Anthropic API. Please go to Plans &
Billing to upgrade or purchase credits."},"request_id":"req_011CeQqahpusvb1cHm6UmUK3"}
```

**⚠ This is NOT the message this file has documented five times.** The cap says *"You have
reached your specified API usage limits… You will regain access on <date>"* and is fixed by
raising a limit. This says **"Your credit balance is too low"** and is fixed by **purchasing
credits**. A session pattern-matching on "400 + billing" will reach for the wrong remedy — and
the wrong-account trap (§2026-08-24) applies to both.

### The boundary, measured to the second `[MEASURED 2026-08-26 08:50Z]`

| | time (UTC) |
|---|---|
| last SUCCESSFUL LLM call, fleet-wide | `2026-08-25 23:46:29.930` |
| first credit failure | `2026-08-25 23:47:10.233` |

**691 credit failures across 18 agent types**, and **zero successful calls in every hour from
00:00 to 08:00**. This is a wall, not an intermittent — unlike the cap, which ran at ~5%.

Worst-hit: `page-content-writer` (259), `tool-improver` (100), `content-quality-auditor` (69),
`council-gate` (55), plus 14 more.

`ai_endpoint_health.claude` reads **UNHEALTHY** carrying this 400 — correctly, and note **MDL-044
cannot clear it**, which is right: there are no successful calls to clear it with. The flag is
telling the truth.

### What the owner needs to do

**Purchase credits** (Plans & Billing). ⚠ **Check you are on the right organisation first** —
the 08-22/23 event cost ~16 hours because the fleet's key is not on the org
`platform.claude.com` opens by default. Decisive column: **Organization settings → API keys →
`Last used`**; a live key can never read "30+ days ago".

### Recovery check

```sql
SELECT date_trunc('hour',created_at) hr, count(*) FILTER (WHERE success) ok,
       count(*) FILTER (WHERE NOT success AND error_message ILIKE '%credit balance%') credit
FROM llm_call_log WHERE created_at > now() - interval '6 hours' GROUP BY 1 ORDER BY 1;
```
It must move **and keep moving** before declaring it over.

---

## ✅ …AND THIS OUTAGE DELIVERED THE PROOF THIS LANE WAS WAITING FOR — partly

The council round I submitted at 08:2xZ (`dfde47a4-a64b-4fe8-ba80-b9be88da0e21`) ran straight
into it. Its `__step_error` reads:

> `step council_decide failed: … no reviewer produced a readable opinion (6 abstained, **11
> unreadable**: review_editquality.result, review_reuse_agent.result, review_guidelines.result,
> review_guardian.result, review_di…)`

**ELEVEN SEATS WERE CLASSIFIED `unreadable`, NOT `abstained`.** That is the mechanism firing in
production for the first time: the coordinator wrote `__step_errors`, `reviewStepFailed` read
it, and the council correctly recorded eleven opinions as **owed and lost** rather than as
considered non-objections. **Before this fix those seats would have been counted as abstentions
and an approval could have stood on six.**

### But the round still died — and that is CORRECT, not a defect

11 unreadable + 6 abstained = **all 17 seats**. With no opinion at all, `council_decide` hit its
own long-standing guard — *"a council with no opinions cannot decide"* — and took
`council_decide`'s `error_step`, which migration 588 **deliberately left** as `complete_invalid`.
Every branch behaved as designed.

**So the honest statement of what is now proven, and what is not:**

| claim | status |
|---|---|
| an errored seat is recorded `unreadable`, not `abstained` | **PROVEN in production**, 11 seats |
| the writer→reader contract holds end to end on live data | **PROVEN** |
| a round SURVIVES a lost seat and returns REVISE naming it | **STILL NOT PROVEN** |

The last one needs a **PARTIAL** outage — some seats answering, some failing. This was a
**TOTAL** one, where no fix could have produced a verdict.

**The limit this exposes, stated plainly: the fix converts a partial provider failure from fatal
to survivable. It cannot save a round when the provider is 100% down, and it was never intended
to.** That belongs in the record next to the win.

---

## RECURRENCE 2026-08-26 (`idea_uk_vm_site` lane) — same class, new message: "credit balance is too low", 23:47Z onward, OPEN as of 08:50Z

Found while reading idea.uk's overnight failures after the v1.0.1341 roll; recorded here because
this file is the class and carries the proven recovery procedure.

- **Message differs from 08-10:** `{"type":"invalid_request_error","message":"Your credit balance
  is too low to access the Anthropic API. Please go to Plans & Billing to upgrade or purchase
  credits."}` (sample `req_011CeQrxFHTYb4BE9SeTB5Kd`) — credits exhausted, not the monthly usage
  cap; no "resets September" clause this time, so there is no calendar to even wrongly wait for.
- `[MEASURED 2026-08-26 ~08:55Z]` `agent_error_log`: **1,884** rows matching `%credit balance%`
  from **2026-08-25 23:47:10Z** to **08:50:26Z** — still firing at last read, ~9 h in.
- **Attempt burn is the compounding damage:** 20 `site_work_items` rows driven to terminal
  `failed` (attempt_count 3) across 6 sites so far — loanzy 7, dartsonline 6, cookly 4,
  ai-agent-orchestration 1, idea.uk 1 (`ade31076`, `dead_fragment_link`), system.internal 1 —
  and last night's fleet audit wave left large `triaged` backlogs (idea.uk alone: 34) queued to
  dispatch into the outage.
- **Recovery = the owner adds credit** (what actually ended the 08-10 instance, 21 days before
  the calendar). Recovery check unchanged from §6: `SELECT max(created_at) FROM llm_call_log
  WHERE success;` — sustained successes, not one probe. Post-recovery, sweep the burned rows:
  `SELECT id, item_type, site_id FROM site_work_items WHERE status='failed' AND error LIKE
  '%credit balance%';` — these are provider casualties, not code bugs; re-read before anyone
  diagnoses them as regressions of the fresh build.
- One console pointer for the owner, from the fleet's own memory: if the billing page shows
  credit/0% used, check which ORG the fleet key belongs to via the key's "Last used" — the
  default console org has not been the fleet's before.

### RECURRENCE RESOLVED 2026-08-26 ~08:58Z — the owner added credit; boundary measured; the 20 burned rows RESET

- **Recovery boundary `[MEASURED]`:** last `%credit balance%` error **08:57:46Z**; first
  `llm_call_log` success **08:58:29Z**; by 09:02:11Z **14 successes across 6 agent types**, zero
  new credit errors — sustained, per this file's own bar. Outage span 23:47:10Z → 08:57:46Z
  (~9h11m), 1,884 errors.
- **The platform self-healed everything below max_attempts:** 33 rows that hit the error on
  attempts 1–2 returned to `triaged` on their own and 2 had already completed by 09:02 — so the
  designed retry behaviour needed no help. Only the **20 rows at attempt_count 3** were stuck.
- **Those 20 were RESET 2026-08-26 ~09:03Z** (`idea_uk_vm_site` lane): backup first
  (`CREATE TABLE bak_credit_burn_20260826 AS SELECT * … WHERE status='failed' AND error LIKE
  '%credit balance%'` — holds the full rows incl. error text), then
  `status='triaged', attempt_count=0, error=NULL, retry_after=NULL, claimed_by=NULL,
  claimed_at=NULL` on the same predicate. Rationale: they differ from the 33 self-healed rows
  only in WHEN their third attempt landed; the cause was environmental and uniform. A row a lane
  wants dead can simply re-fail; the backup preserves the record. By site: loanzy 7, dartsonline
  6, cookly 4, ai-agent-orchestration 1, idea.uk 1 (`ade31076`), system.internal 1
  (`1d60fc7b`, a `needs_diagnosis`).
- Post-reset check: `SELECT count(*) FROM site_work_items WHERE status='failed' AND error LIKE
  '%credit balance%';` → **0**. Drop `bak_credit_burn_20260826` once nobody needs the record.

---

## RECURRENCE — 2026-09-04 11:20→11:56Z, ~35 min, fleet-wide (contributed by the bug sweep lane, 442)

Reporting an episode and, more usefully, a **measured split of this file's own subject matter into
two failure modes that never co-occur.**

### The episode

`[MEASURED 2026-09-04]` Every LLM call in the estate failed with HTTP 400
`"Your credit balance is too low to access the Anthropic API"`. Not council-only, not one agent:

| 5-min bucket | ok | credit-400 | distinct agent types |
|---|---|---|---|
| 11:15 | 13 | 0 | 2 |
| 11:20 | 6 | **19** | 5 |
| 11:25 | **0** | 9 | 2 |
| 11:30 | **0** | 30 | 3 |
| 11:35 | **0** | 22 | 2 |
| 11:40 | **0** | 13 | 3 |
| 11:45 | **0** | 14 | 2 |
| 11:50 | **0** | 5 | 1 |
| 11:55 | 6 | 5 | 3 |
| 12:00 | 12 | **0** | 4 |

**Zero successful LLM calls of any kind between 11:25 and 11:50.** 117 failures in the 11:00 hour.
Last failure of any kind 11:56:47Z; recovered without intervention from this lane.

### ⚠ THE SECOND-ORDER DAMAGE, which is what makes this worth reporting rather than logging

**Three council rounds — three different lanes — were converted into `revise` verdicts by the
outage**, and the verdicts are indistinguishable from real ones at the column everyone reads:

| corr | time | decision | unreadable | readable |
|---|---|---|---|---|
| `3e9e8ce8` | 11:22:23 | revise | 6 | 5 |
| `5de01fd3` | 11:28:44 | revise | 8 | 5 |
| `8bf83b59` (this lane, mig 773) | 12:02:36 | revise | 6 | 3 |

On ours, **the three seats that WERE readable all approved** (`guardian`, `debug_historian`,
`architecture`) and the round contains **not one objection**. A submitter who reads
`metadata->>'decision'` starts revising a change nobody objected to. The existing landmine
("A council-gate run that ends `COMPLETED` at `complete_invalid`…") covers the **total** outage,
where no verdict is produced at all — that at least leaves an absence. **This is the partial case
and it leaves a positive artefact.** Addendum appended to `LANDMINES.md` with the discriminating
query. It also killed a `landmine-verify-dispatch.sh` run (4 attempts, all credit-400), whose
verdict then simply never arrives — indistinguishable from queue latency.

### ⚠ THE FINDING: this file's title and its evidence are TWO different failure modes

`[MEASURED 2026-09-04, all history — `llm_call_log` retains from 2026-03-25, 90,037 calls, so this
is the whole record and not a window]`

| mode | message | failures | first | last |
|---|---|---|---|---|
| **B** | `usage limit` (this file's title) | **15,315** | 2026-04-05 | **2026-08-31** |
| **A** | `credit balance is too low` | **934** | 2026-04-10 | **2026-09-04** |
| other | — | 463 | 2026-03-31 | 2026-09-04 |

**They are mutually exclusive by day. Across 31 affected days, not one shows both** — every day is
either all-B or all-A. `[INFERRED, not established]` that this is two states of one billing
arrangement rather than two independent faults; the exclusivity is measured, the cause is not, and
somebody with account access can settle it in a minute where I cannot (and per the estate's own
rule I did not go looking for keys).

**The load-bearing consequence: mode B STOPPED on 2026-08-31 and mode A is what has fired since.**

```
08-27  usage 142 | 08-28  usage 482 | 08-29  usage 757 | 08-30  usage 719 | 08-31  usage 288
09-02  credit  8 | 09-04  credit 117
```

So whatever changed around **2026-09-01** — and something did — **retired mode B and left mode A
live.** Any remediation in this file justified by "the cap resets on <date>" or aimed at usage-cap
behaviour should be re-checked against that: a credit balance does not reset on a billing date, it
runs out again.

**Two mode-A episodes are not in this file** (`grep` for both returns 0): **2026-04-10** (78
failures — *earlier than the 07-31 this file calls the first*) and **2026-09-02** (8). Both are
mode A, and this file's narrative is overwhelmingly mode B, so this is a gap in coverage rather
than an error by anyone. Full mode-A episode list: 04-10 (78) · 08-08 (20) · **08-25/26 (711, ~10 h
— the shape to plan for)** · 09-02 (8) · 09-04 (117).

### The query, so nobody has to re-derive it

```sql
SELECT created_at::date AS day,
       count(*) FILTER (WHERE error_message LIKE '%credit balance is too low%') AS credit_balance,
       count(*) FILTER (WHERE error_message ILIKE '%usage limit%')              AS usage_limit
FROM llm_call_log WHERE NOT success AND provider='anthropic'
GROUP BY 1 HAVING count(*) FILTER (WHERE error_message LIKE '%credit balance is too low%') > 0
              OR count(*) FILTER (WHERE error_message ILIKE '%usage limit%') > 0
ORDER BY 1;
```

⚠ **Do not collapse the two modes into one `NOT success` count.** They have different messages,
different remedies and — on this evidence — different lifetimes, and a combined count would have
shown a healthy-looking decline from August into September while mode A was in fact taking over.
