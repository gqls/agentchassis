# 327 — the build trigger can publish NOTHING and exit 0: one submission in three vanished with no orchestration, no work item, and no error

> **CLOSED 2026-08-25 — fixed AND live.** Everything this bug produced is shell/Python, so it
> was live the moment it was committed; **no image roll was ever involved**, and a chassis build
> changes nothing here. Verified at close, re-derived not asserted: **live racing queue 0** ·
> **22 callers** on `scripts/kafka-publish-lib.sh` (register **OPP-009**), three of them lanes
> this one never spoke to · library self-test ALL PASS · every migrated script induced-failure
> tested against an unreachable broker.
> **The residual is DATA, not an open defect:** ~55 `scripts/` files carrying the old form but
> untouched for 30d+, plus `docs/` lane one-offs that must NOT be rewritten (doing so falsifies
> the record and risks firing a live trigger). `check_kcat_stdin_race` in `pattern-check.py`
> catches any of them the moment someone picks one up — **the class is bounded by a detector,
> not by finishing a sweep.**
> **Two decisions deliberately left open and NOT lost:**
> `docs/agent_docs/docs024_key_docs_latest/bugfix_327_silent_publish_drop/HANDOFF_2026-08-25_open_decisions.md`
> — a council round (possible on demand via `FORCE=1`; no permanent scope needed) and the
> in-cluster Go submit path (wait; revisit only if a message is observed lost THROUGH the library).
> Story: `SUMMARY_2026-08-25_silent_publish_drop_closing.md`. Diagnosis for a session arriving
> **with the symptom**: `016b` §9, *"I dispatched it and no orchestration row appeared"*.

> **RENUMBERED 327 → 327b on 2026-08-25 (owner decision), following the `347b` precedent.**
> Two unrelated bugs were filed as 327 on the same day, and on close BOTH would have landed in
> `bugs_closed/` — where the directory stops telling them apart, which is how one live handoff
> already conflated them. `327b` is **this** one, the dispatch drop. `327` is the *partial spec
> write*, closed 2026-08-23 by the `copy_quality_two_stage` lane.
> **The number was not reassigned** — no other bug ever held 327b; the suffix only disambiguates.
> ⚠ **Older references say `bugs_open/327`.** `[MEASURED 2026-08-25]` **66** files carry that
> string and were deliberately NOT rewritten — they are explanatory citations, not deferrals, and
> a grep for `327` still finds this file. Only the **4** references naming the full filename were
> updated. This is an estate-wide shape, not a local one: **36 numbers** name two cases.

**Filed 2026-08-19** by the `loanzy_uk_example_site` lane. **Status: OPEN, UNOWNED.**
Sibling of `bugs_open/326` (a re-submission that dedups also reports success) — together they
mean **"I submitted it" is not evidence that anything was submitted**, for two unrelated reasons.

> **On the 090 loop:** not run. The claim is an absence measured against a fleet-wide control in
> the same window, not a theory about a mechanism; the underlying `kcat -P` behaviour is already
> a documented landmine. What is new is that it hit the customer-facing build trigger.

## What happened

`bash scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh loanzy.uk`
printed its correlation and orchestration ids and exited 0. The kcat pod ran and was deleted
normally. **No `orchestration_states` row for that correlation ever appeared, and no work item
was created** — checked immediately, then again after 60s, then again at 10 minutes.

```
correlation 8fa2a4a6-2af1-4675-bae9-bbd59b702160 → 0 rows in orchestration_states
```

**The control that makes this a drop rather than a stalled fleet:** in the same ten minutes,
**29 orchestrations were created fleet-wide**, so the chassis was consuming normally. A second,
identical invocation minutes later (`f7e4dec3-…`) landed and filed
`needs_domain_research` / `research_loanzy.uk` / `triaged` within seconds.

**Rate on this evidence: 1 of 3 submissions from this lane on 2026-08-18.** Small sample, stated
as such — but the failure is silent, so the true rate fleet-wide is unknown and unknowable from
the logs today.

## Why it matters here rather than in a runbook

`kcat -P` publishing nothing at exit 0 is already in `LANDMINES.md`, and every session is told
to verify the item rather than the exit code. That is a fine rule for a session and a bad one
for a product: this is the **entry point of the one-shot build route**. The operator-visible
outcome of a dropped submission is indistinguishable from a successful one — same printed ids,
same "SAVE: CORRELATION_ID=…", same clean exit — and the customer-facing consequence is a build
that never starts while everyone believes it did.

## Fix candidates, ordered by what closes the door

1. **The trigger verifies its own landing.** After publishing, poll for the
   `needs_domain_research` row (or the orchestration row) for ~30s and **exit non-zero** if it
   never appears, printing the correlation to retry with. A submission whose evidence is the
   row it created cannot be silently dropped.
2. **Stop publishing through a throwaway pod.** The kcat-in-a-pod pattern is what makes the
   producer's failure invisible; a small producer that returns the broker's ack (or an HTTP
   endpoint on an existing service) would surface the failure at the source.
3. **Idempotent re-submission** — depends on `bugs_open/326`, and is the reason the two should
   be fixed together: today, "the submission vanished, try again" and "you already submitted
   this, so nothing will happen" both look like success, and the operator cannot tell which one
   they are in.

## How to verify a fix

~~Run the trigger with the broker deliberately unreachable (or the topic renamed) and require a
**non-zero exit** and a message naming what did not land.~~ **A fix verified only against a
healthy broker has tested nothing — the failure mode IS the unhealthy path, and it must be
induced.** (That second sentence stands and is the whole point. The recipe in the first does
not — see the correction.)

> **CORRECTED 2026-08-23 — the recipe above passes on the UNFIXED script, so it cannot tell a
> fix from no fix.** Measured, both arms side-effect free (each publishes zero messages):
>
> | induced condition | published | exit | operator sees |
> |---|---|---|---|
> | broker unreachable — *the recipe above* | 0 | **1** | loud broker errors |
> | **empty / unattached stdin, HEALTHY broker** — *the real arm* | 0 | **0** | **nothing at all** |
>
> `set -euo pipefail` already propagates `kubectl`'s non-zero status when the pod terminates in
> Error, so the unfixed trigger **already exits 1** on an unreachable broker. The two modes are
> **opposite in observability** and this file tested the loud one.
>
> **Induce the silent arm instead:** `… kcat -P -b <REAL broker> -t <topic> < /dev/null`, which
> reproduces the post-race state deterministically and publishes nothing. "Topic renamed" is also
> weak — with auto-create on, the broker accepts the publish and exits 0, testing neither arm.
>
> **What this means for any fix:** on the mode that actually bites, **the exit code carries no
> information**, so a receipt is not belt-and-braces — it is the only instrument that reads the
> silent arm.

## Status 2026-08-23 — FIXED AT THE TRIGGER, and the bug stays OPEN

`082_submit_domain_unified.sh` now publishes through **`scripts/kafka-publish-lib.sh`** (register
**OPP-009**): payload in the container **command** so the stdin race is structurally impossible,
an **asserted** `PUBLISH_OK` receipt, `SAVE: CORRELATION_ID` moved to *after* the receipt (it used
to print three lines *above* the publish, telling the operator to record ids for a message not yet
attempted), a **non-zero exit** naming what did not land, and a landing check that separates
*never published* (retry now, **10**) from *consumed and REFUSED* (**12**, read from
`agent_error_log`) and *published but not landed* (**13**, wait — this is also latency's
signature). A new racing publisher now draws an advisory `check_kcat_stdin_race` at commit time.

**It stays OPEN, deliberately.** `[MEASURED 2026-08-23]` of **218** `kcat -P` publishers, **201**
still use the racing form — but that figure counts COMMENTS. `[MEASURED 2026-08-24]` **183** race
in code and **160** race *and* parse (23 are scrapbooks; 18 more match only inside comments
warning about this very trap, including the ones written into each file as it was migrated). So
**160** is the exposure, not 178 — and
**three** callers have been migrated (2026-08-24: `scripts/trigger-landmine-verifier.sh`, whose
caller counted dispatch failures from the one signal absent on the silent arm; and
`097_TRIGGER_council_review_v1.sh`, closing all three halves of its own landmine — the
pre-publish receipt, the stdin race, and a second-resolution pod name that made two same-second
submissions collide). The customer
path this bug names is closed; the class is not, and **this file tracks the class** (owner,
2026-08-24) — so it stays open until the runnable racing publishers are gone, not until 082 works.

**Also corrected here:** this file's evidence (`8fa2a4a6…` → 0 rows) can no longer be
re-verified — `orchestration_states` retains **~2 days** and holds nothing for 2026-08-18, so that
query now returns 0 regardless. Validity was re-established at the **source** instead (the script
was unchanged since 2026-07-30). The competing explanation *was* excluded: no `agent_error_log`
row names the correlation, and that table recorded **3,761 rows including a real
`VALIDATION_ERROR_DROPPED` on 2026-08-18**, so the recorder was live and its silence is evidence.

**One finding this file did not anticipate:** the racing form is non-deterministic in *both*
directions. In a controlled 10-vs-10 run it published a **duplicate** (11 delivered for 10 sends).
On the real topic a duplicate submission is two builds. `[ONE observation, not a rate.]` In the
same run the **drop did not reproduce** — 0 of 10 — which excludes the historical 4-in-5
(p≈1e-7) but bounds today's rate only below ~26%. **The rate is load-dependent; the silence is
not.**

Full trail: `docs/agent_docs/docs024_key_docs_latest/bugfix_327_silent_publish_drop/`.

## Scope of what remains, and the bar for closing this (2026-08-24)

`[MEASURED 2026-08-24]` **155** runnable racing publishers — but that number is dominated by
files nobody runs, and treating it as a backlog would keep this bug open for ever while nothing
improves:

| slice | count | disposition |
|---|---|---|
| under `docs/` — lane one-offs | 86 | **OUT OF SCOPE.** Rewriting them falsifies the record, and a `*_TRIGGER_*` script **publishes on every invocation**, so editing and testing one risks a live dispatch |
| dormant (untouched 30d+) | 102 | not work |
| literal duplicates `(1)`,`(2)`,`(4)` | 6 | download artefacts |
| match only inside COMMENTS | 18 | not publishers — warnings *about* this trap, including the ones added by each migration |
| **`scripts/` and touched within 30d** | **11** | **the real queue** |

> **ALL THREE CONDITIONS MET, 2026-08-24.** The 11 live `scripts/` publishers are migrated
> and induced-failure tested; the live queue re-derives to **ZERO**; `check_kcat_stdin_race`
> is in place; and the scope statement above stands. **21 callers** now use the library, all
> parsing, self-test green. **The residual is DATA — dormant and historical files — not an
> open defect. Closing is the owner's call.**

**Close this bug when:** (1) those 11 are migrated and induced-failure tested; (2) the
commit-time detector `check_kcat_stdin_race` remains in place — it is what bounds the class, and
it fires only on *newly added* racing publishers; (3) this scope statement stands. The residual
is then **DATA**, not an open defect.

Adoption `[MEASURED 2026-08-24]`: **10 callers**, of which **two are lanes this one never spoke
to** — `140_tool_suggester/077_update_noted_write_tool.sh` (`caa55f04f`) and
`bugfix_380_claims_fail_open/TRIGGER_claims_audit.sh`. That is the evidence for shipping the
remedy as something callable rather than as more documentation: the safe form had been written
down for a month and had **2** assertions to show for it.

Continue here:
`docs/agent_docs/docs024_key_docs_latest/bugfix_327_silent_publish_drop/HANDOFF_2026-08-24_continue_here.md`

