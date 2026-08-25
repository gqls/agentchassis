# HANDOFF 2026-08-25 — `bugfix_243_provider_cap_resilience` (243-anthropic-cap)

**Read this first. It is the cold-start doc for this lane.**
Bug: `bugs_open/243_HANDOFF_2026-08-10_anthropic_account_usage_limit_reached_every_llm_step_fleetwide_fails_until_september.md`
— **the anthropic-cap slug.** `243` names TWO unrelated bugs; the other (tool-acceptance storage
client) is a different lane and is not ours. **Resolve by slug, never by number.**

---

## 1. One paragraph: what this lane was for

The Anthropic account intermittently refuses calls with a 400 *"You have reached your specified
API usage limits"*. **The cap itself is the owner's to clear and this lane never touched it.**
What it fixed is the platform's *reaction*: the framework detected the condition once and then
threw the fact away at every layer that needed it, so two consumers acted at the wrong altitude
— the work-item claim gate stopped the **whole fleet** for up to an hour on one refusal, and the
council gate discarded a **whole review round** on one seat's transient.

## 2. Status: ALL THREE FIXES ARE LIVE. One proof is outstanding and cannot be forced.

| # | fix | state |
|---|---|---|
| 1 | **MDL-044** — a successful live call clears `ai_endpoint_health` | **LIVE + PROVEN TWICE** (v1.0.1334) |
| 2 | **mig 596** — claude re-probe 3600s → 60s | **APPLIED** 2026-08-24, cadence measured 92–94s |
| 3 | **WFA-023** — `__step_errors` writer + council classifying an errored seat `unreadable` | **LIVE** (v1.0.1337, still in v1.0.1339) |
| 4 | **mig 588** — the 17 seats' `error_step` → their own `next_step` | **APPLIED** 2026-08-25 |

**Re-verified after the v1.0.1339 roll (2026-08-25 19:07Z)** — a roll is the moment to re-check,
because Go can regress by a build from an older ref and DB config can be overwritten by a seed:
ancestry `git merge-base --is-ancestor <commit> a7459a44b` → **IN** for all three commits
(`dbd865ee8`, `e521cde3e`, `893a12d47`); and the DB config survived — probe interval **60s**,
seats repointed **17/17**, `complete_invalid` still exactly **2**.

**Council: APPROVED round 1**, `82f07fa6-1c42-46ad-bdf6-1d58892c44a7`. Every commit carries the
trailer.

### The ONE thing outstanding

**A council round in which a seat's LLM call actually errors.** It must:

- **reach a verdict** (not `complete_invalid`), and
- report that seat under **`unreadable`**, not `abstained`, and
- if the remaining seats would have approved, come back **REVISE** naming the lost seat.

**Negative control — ALREADY OBSERVED, post-migration** `[MEASURED 2026-08-25 09:49:00Z]`: the
first council round to run after mig 588 applied reached **`complete_approved`**, decision
`approved`, **5 abstained, 0 unreadable**. So repointing all 17 seats' `error_step` did **not**
break the ordinary path — a round where nothing errors is unaffected, as intended. That is the
regression check discharged; it is *not* the positive arm.

```sql
-- the positive arm, whenever it lands
SELECT created_at, metadata->>'decision', metadata->>'abstained', metadata->>'unreadable',
       metadata->'unreadable_at'
FROM diagnosis_artifacts WHERE kind='council_report'
  AND (metadata->>'unreadable')::int > 0 ORDER BY created_at DESC LIMIT 5;
```

**This cannot be forced** — it needs a real provider transient. **Do not fake it, and do not
close the bug on its absence — but equally, this lane's work is done.** See §6.

**Progress toward it, refreshed** `[MEASURED 2026-08-25 ~19:15Z, chassis v1.0.1339]`:

- **47 council rounds have completed since mig 588 applied** (27 `complete_approved`,
  20 `complete_revise`, 1 in flight) and **ZERO reached `complete_invalid`**. Before the fix the
  08-19 contribution measured roughly a coin-flip per round dying there.
- **⚠ But that zero is NOT the proof, and the reason is the demand control again:** there were
  **0 cap failures on both 08-24 and 08-25**, so no transient has occurred for the new path to
  handle. All 47 rounds show `unreadable = 0`. **The mechanism is live, un-regressed, and
  UNEXERCISED.** A zero here means "nothing tested it", not "it works".

**Cap recurrence, refreshed — my own earlier prediction has NOT materialised:**

| day | cap failures | ok |
|---|---|---|
| 08-21 | 3 | 1,223 |
| 08-22 | **113** | 1,063 |
| 08-23 | 32 | 1,109 |
| 08-24 | **0** | 1,850 |
| 08-25 | **0** | 1,395 |

I predicted (`[INFERRED]`, from the monthly-reset shape plus month-end clustering) that
recurrence was likely before 08-31. **Two clean days at the highest call volumes in the record
have not borne that out** — consistent with the account having been properly funded once the
wrong-account error was found on 08-23. **Two days cannot settle it either way** (and it is a
monthly limit with six days left), so the prediction stands as unresolved rather than refuted.
Re-run the histogram in §4 before repeating any figure here.

## 3. If you change anything here, these are the traps that will bite you

- **⚠ Probe the WRITER's own literal, never `__step_errors`.** The *reader* mentions that key
  too, so `grep -ac "__step_errors" /proc/1/exe` returns **1** whether or not the writer
  shipped, and reads as "it landed". The discriminating probe is
  `grep -ac "step-error record capped at" /proc/1/exe`. **Always run a known-present AND a
  known-absent control in the same breath** (`diagnose_council_decide` = 15, any invented string
  = 0).
- **⚠ `error_step` sits inside `config`, not at the step level.** Read at step level and the
  census returns `(none) | 29` — clean, confident, wrong. Bit two sessions.
- **⚠ The health PROBE CANNOT SEE THE CAP.** `pingClaude` returns healthy for **any** non-auth
  status and the cap is a **400**. For this condition it is a *timer that clears the flag*, not
  a health check. **Any "require N consecutive failures" hardening belongs on the live-traffic
  writer (`ai_actions.go`), never on the prober, where it would never fire.**
- **⚠ 60s does NOT give a 60s bound.** The probe needs the `ai-endpoint-health-check` task to
  tick **and** the endpoint interval to elapse, and that task is **also 60s**, so they compose.
  **Measured 94s and 92s.** Honest bound: one to two minutes.
- **⚠ An absent seat is an ABSTENTION unless something says otherwise**, and an abstention does
  not gate while `unreadable` downgrades an approval (`diagnose_council_decide_action.go`, the
  `raw == nil` branch and the approval check). This is the whole reason mig 588 could not ship
  alone — it would have converted a lost seat into a silent non-objection.
- **⚠ `psql` renders a bare boolean `t` but a CONCATENATED one `true`.** A verification loop
  comparing to `"t"` on a concatenated row **can never report success**. This cost ~40 minutes
  and nearly produced a false "the fix does not work" (`WRONG_CALLS.md` 2026-08-24). Prefer
  `CASE WHEN healthy THEN 'YES' ELSE 'NO' END` so no driver rendering is load-bearing.
- **⚠ Put the demand control INSIDE the observation query.** Two "still false" results here were
  **vacuous** — zero successful calls in the window, nothing for a clear-on-success writer to
  answer. `(SELECT count(*) FROM llm_call_log WHERE success AND created_at > t0)` alongside the
  row makes a zero-demand window self-labelling.
- **⚠ A council seat reviews a SKETCH, and a sketch is not executable.** Mig 588's reviewed SQL
  used `UPDATE … FROM LATERAL jsonb_each(ad.default_config …)`, which Postgres refuses outright.
  It failed at apply and rolled back. **Compile-check any SQL you submit**; do not expect the
  gate to catch this class.

## 4. How to verify each half is still true

```sql
-- MDL-044 (the signature: last_healthy LATER than last_checked — no other writer can do this)
SELECT name, healthy, last_checked, last_healthy, check_interval_seconds
  FROM ai_endpoint_health WHERE name='claude';

-- mig 588: expect 17 seats routing to their OWN next_step, and exactly 2 keeping complete_invalid
SELECT count(*) FILTER (WHERE v->'config'->>'error_step'='complete_invalid') AS still_invalid,
       count(*) FILTER (WHERE k LIKE 'review\_%' AND v->'config'->>'error_step' = v->>'next_step') AS repointed
FROM agent_definitions, LATERAL jsonb_each(default_config #> '{workflow,steps}') AS s(k,v)
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- is the cap biting? (histogram, NOT max() — an hourly bucket cannot see an outage inside it)
SELECT date_trunc('hour',created_at) hr, count(*) FILTER (WHERE success) ok,
       count(*) FILTER (WHERE NOT success AND error_message ILIKE '%usage limit%') cap
FROM llm_call_log WHERE created_at > now() - interval '14 hours' GROUP BY 1 ORDER BY 1;
```

Full command set with every gotcha attached: `RUNBOOK_provider_cap_resilience.md`.

## 5. What is NOT this lane's, and must not be quietly adopted

- **Candidate 1 — the owner adds credit.** Still the only thing that restores service under a
  real cap. **Not a control**, and this lane does not change that.
- **Candidate 2 — a second provider.** Undecided, owner's call, and a **real build**: `127 of
  127` configured LLM steps across 55 live agents name `anthropic` and the same key.
  ⚠ **That census is dated 2026-08-10; re-run it before quoting.**
- **The wrong-account trap.** The 08-22/23 event cost ~16h because the fleet's key is **not** on
  the org `platform.claude.com` opens by default. Billing read `$0.00 / 0% used` *while* the API
  refused. Decisive column: **Organization settings → API keys → `Last used`** (a live key can
  never read "30+ days ago" — a failed call is still a use).
- **`bugs_open/354`.** Its code is at HEAD (I swept it in, owner-approved, 2026-08-24) **but its
  bug is NOT fixed**: the `outcome` declaration it keys on appears in **no** live agent config
  (checked at all three placements), so its discriminator is present and unreachable, and it has
  **never been council-submitted**. Told them in their bug file. **Not ours to finish.**

## 6. Can the bug be closed?

**This lane's work is complete.** All four changes are live, each proven at the artefact, and
the mechanisms are registered (**MDL-044**, **WFA-023**).

**But `bugs_open/243` should stay OPEN**, and not as bookkeeping:

1. The **positive arm of §2** has not been observed. The estate's bar is *fixed AND live*, and
   for the council half "live" is proven only by a round that survives a seat error.
2. **The bug is a provider/billing condition that recurs and is not fixed by anything here.**
   Candidates 1 and 2 are open owner decisions. Closing 243 would read as "the cap problem is
   solved", which is false — what is solved is how much damage each refusal does to us.

**A fair close condition:** one council round with `unreadable > 0` reaching a verdict, plus an
owner decision on candidate 2. Until then this file is the state.

## 7. Where everything is

| what | where |
|---|---|
| lane docs (5) | `docs/agent_docs/docs024_key_docs_latest/bugfix_243_provider_cap_resilience/` |
| owner's prose log | `README_where_we_are.md` in that dir |
| every command + gotcha | `RUNBOOK_provider_cap_resilience.md` |
| evidence + every misstep | `NOTES_provider_cap_resilience.md` |
| the plan + scope rulings | `PLAN_2026-08-24_provider_cap_resilience.md` |
| register | `docs026_concept_register/register/model-infrastructure.md` (MDL-044), `workflow-authoring.md` (WFA-023) |
| migrations | `sql_for_agents/596_claude_probe_interval_60s.sql`, `588_council_seat_transient_costs_one_seat.sql` (both applied + ledgered) |
| council | `SUBMISSION_CORR 82f07fa6-1c42-46ad-bdf6-1d58892c44a7`, APPROVED r1 |
| contributions sent | `webdesign_uk_build_service/CONTRIB_2026-08-24_…`, `bugfix_302_design_repair_verification/CONTRIB_2026-08-24_…`, notes in `bugs_open/354` |

**Key commits:** `e521cde3e` (MDL-044 + council reader) · `dbd865ee8` (`__step_errors` writer) ·
`893a12d47` (owner-approved sweep of the 354 lane's callee) · `9873bac59` (mig 588 applied +
renamed off `_HOLD`) · `203e5cdf5` (mig 596 split + applied).
