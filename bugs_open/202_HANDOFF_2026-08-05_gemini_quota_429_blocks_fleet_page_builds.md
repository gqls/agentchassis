# 202 — `page-content-writer`'s Gemini provider is out of quota (429), and page builds are failing fleet-wide

**Filed 2026-08-05 by the `webdesign_uk_build_service` lane, whose landing page is
blocked by it.** Not run through `090`, with the reason stated per the 2026-07-31
ruling: **the root cause is asserted by the provider itself in the error body** —
this is a quota/billing state, not a mechanism to diagnose. What IS open is the
decision (below), which is an owner call, not a debugging step.

## Symptom

`page-content-writer` orchestrations fail at
`process_sections_loop_*_generate_content`:

```
failed to execute action execute_llm_prompt: AI endpoint unavailable:
provider=gemini model=gemini-pro-latest ... status 429 ...
"You exceeded your current quota, please check your plan and billing details."
```

## Scope, measured (agent_error_log, `%429%gemini%`)

| domain | hits | latest |
|---|---|---|
| webdesign.co.uk | 12 | 2026-08-05 04:25 |
| webdesign.uk | 6 | 2026-08-05 03:08 |
| gaswholesalers.com | **123** | 2026-08-05 02:00 |
| idea.uk | 21 | 2026-08-05 01:32 |
| vonc.com | 12 | 2026-08-03 23:21 |

Five domains inside ~28h ⇒ **any lane whose build reaches `generate_content` is
affected**. Two active agent definitions name gemini: `page-content-writer`,
`feature-designer` (live-config query, 2026-08-05).

## What this is NOT

- Not `192` (fixed, live on v1.0.1250) — that was a config-shape defect at
  `select_sections`; this fails later, at generation, with a provider error.
- Not content validation — no blocker rows; the writer never produced output.

## The decision needed (owner-level; NOT taken unilaterally from this lane)

1. **Add Gemini quota / fix billing** — restores as-is behaviour.
2. **Re-point `page-content-writer` at an Anthropic model** — a live-DB config
   change to a SHARED agent: council-gate it, and note §7b of the webdesign plan
   already rules intake=Haiku/builds=Fable for that lane's product. A fleet-wide
   writer model change is bigger than one lane's ruling.
3. **Do nothing** — Google quotas often reset daily; the cheap test is re-driving
   one failed item after the reset window and reading `agent_error_log` again.

Whoever picks this up: re-drive is `status='triaged', error=NULL` on the failed
item + the `076` heartbeat (extract from the shebang — the file has notes pasted
above it).

---

## CORRECTED 2026-08-05 (owner dashboard) — the project is TIER 1, not free tier

The filing session then told the owner the key had "no billing attached"
[WRONG — inferred from the 250 limit without checking the account; logged in
`WRONG_CALLS.md`]. The owner's AI Studio dashboard shows: **Tier 1**, project
"agent chassis", monthly spend cap £150 (£21 used), 28-day cost **£46.09**, and
per-model peaks: RPM **8/25**, TPM **36.61K/2M**, RPD **at the 250 cap**.

**So the ONLY binding limit is RPD, per model, per project — 250/day for
`gemini-3.1-pro` at Tier 1.** RPM/TPM have ~3× and ~50× headroom. The fleet
shares one project, so ~250 pro-model calls/day fleet-wide is the whole budget
— yesterday gaswholesalers alone logged 123 failures after the cap.

**The real options (supersede the three above):**
1. **Tier 2 is automatic, no form**: "$100 cumulative Google Cloud spend + 3
   days from first payment" → upgrade "within 10 minutes" of qualifying. At
   ~£46/28d the project crosses $100 in roughly 3–4 more weeks on its own.
2. **Rate-limit increase request form** (no guarantee, free to file):
   https://forms.gle/ETzX94k8jf7iSotH9
3. **Reduce pro-model RPD structurally** — the writer pins `gemini-pro-latest`
   for every section; flash-class models for cheap sections, or a second
   project/key, would multiply capacity. SHARED-writer change ⇒ council, as
   before.
4. The £150 monthly spend cap governs spend, not rate limits — not a lever.

**Timing:** yesterday's 429 said "retry in 20h51m" from 03:08 UTC ⇒ the RPD
window reopens ~**23:59 UTC 2026-08-05**. A re-drive before then fails; after
then, ~250 fresh calls exist for the whole fleet's day.

---

## OWNER RULING 2026-08-05 — option 1: WAIT for the automatic Tier 2 upgrade

No form filed, no writer re-point, no quota purchase. Until the project crosses
$100 cumulative (a few weeks at current burn, sooner with normal use), the fleet
lives inside **250 pro-model calls/day, resetting ~midnight UTC** — so:
- **Re-drive blocked builds just AFTER a reset**, not late in the day; a
  late-day re-drive spends its attempt into an exhausted window.
- A late-day 429 on any lane is THIS bug, not a new one — check the date before
  filing a duplicate.
- The council question (option 3, the shared writer's model) stays OPEN but
  unforced; revisit with real spend numbers if Tier 2's limits still pinch.

This bug CLOSES when Tier 2 is observed active (aistudio.google.com/rate-limit
shows the raised RPD) AND a previously-blocked build has passed through
generation on it.

---

## OWNER RULING 2026-08-05 (later) — TEMPORARY model swap to Sonnet 5, revert when Gemini resets

Supersedes "wait" as the *interim* posture; the Tier 2 wait still stands as the
quota fix. Owner's words: *"swap models from gemini to claude sonnet 5 for now
until tomorrow when gemini resets."*

**Applied 2026-08-05, live immediately (DB config):** `page-content-writer` →
`workflow.steps.process_sections_loop.config.sub_workflow.steps.generate_content.config.ai_service`

| | was | now |
|---|---|---|
| provider | `gemini` | `anthropic` |
| model | `gemini-pro-latest` | `claude-sonnet-5` |
| api_key_env_var | `GEMINI_API_KEY` | `ANTHROPIC_API_KEY` |
| max_tokens | 8000 | 8000 (unchanged) |

Shape copied from the live `fix-proposer`/`grounded-explainer` blocks (1,155
Sonnet calls in 3 days), not invented. `feature-designer` also names gemini and
was **deliberately left alone** — it is not in the failing path and the swap is
minimal-blast-radius.

**THE REVERT (owner's stated intent — "until tomorrow"):** after the Gemini RPD
window reopens (~00:00 UTC 2026-08-06), run:

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
  '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,ai_service}',
  '{"model":"gemini-pro-latest","provider":"gemini","max_tokens":8000,"api_key_env_var":"GEMINI_API_KEY"}'::jsonb),
  updated_at = now()
WHERE type='page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

…and verify with the same `jsonb_pretty` SELECT. **If instead the owner decides
Sonnet output was as good or better, NOT reverting is a real option** — but that
is the council-flavoured "option 3" decision and should be taken as one, with a
build's output in front of him, not by the revert simply being forgotten.
**Either way this file must record which happened.**

Note for anyone reading `llm_call_log` costs this week: writer traffic
2026-08-05→06 is Sonnet, not Gemini — don't attribute it to the wrong provider
when comparing (the `llm-call-log` landmines file also applies).
