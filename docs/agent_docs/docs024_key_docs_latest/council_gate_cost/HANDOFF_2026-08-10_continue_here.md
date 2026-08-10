# HANDOFF 2026-08-10 — council-gate caching is LIVE and PROVEN. One verdict outstanding.

Start here cold. Read with `NOTES_council_gate_cost.md` (technical trail, both
missteps), `README_where_we_are.md` (owner's plain-prose account) and
`RUNBOOK_council_gate_cost.md` (every query with its gotcha attached).

## 0. One paragraph

The owner asked to switch off the improvement loop to save credits. It turned
out to be already inert and nearly free; **council-gate was ~85% of all fleet
LLM spend**. Root cause was structural: 17 sequential review seats each received
~100k input tokens of *identical* evidence, but each seat's persona sat at the
top of its template, so the measured common prefix was **zero characters** and
Anthropic's prefix-matched caching could never fire. Both halves are now shipped
and proven in production on chassis **v1.0.1283**: an opt-in cache breakpoint in
the shared AI client (LCO-008) and all 17 templates reordered (migration 377).
**Measured saving 68–72%**, verified at the artefact. Nothing is outstanding
except one council verdict.

## 1. State — verified 2026-08-10 ~22:10 UTC

| thing | state | re-check |
|---|---|---|
| Cache seam (LCO-008) | **LIVE + PROVEN**, chassis v1.0.1283, both replicas pod-grepped with +ve and −ve controls | `strings /app/agent-chassis \| grep -c -- "<!--CACHE_BREAKPOINT-->"` on every replica |
| 17 seat templates (mig 377) | **APPLIED**. 17/17 start with the shared block, 17/17 marked, **exactly 1 distinct prefix** | RUNBOOK § "Prove the cacheable prefix is byte-identical" |
| Migration 376 (cache columns) | **APPLIED** live, DO/RAISE verify, control induced first | `\d llm_call_log` |
| Measured saving | **68–72%** and climbing as seats accumulate against each write | RUNBOOK § "Prove caching actually works" |
| Council verdict `b54f173e…` | **APPROVED**, 5 advisory, 2 acted on (ttl removed, marker-strip bug fixed) | done, no action |
| Council verdict `d2c51f41…` | **REVISE** (gated by editquality HIGH). Objections are sound but empirically inapplicable *today* — see §2 | see §2 |
| Improvement loop | 3 discovery rotations disabled on owner instruction; saved ~nothing (it was already inert) | `SELECT name,enabled FROM scheduled_tasks WHERE name LIKE 'site-discovery-rotation%'` |
| Model roster | **Staying on Sonnet 5** (owner ruling). A mixed roster is *counterproductive* — see §4 | — |

## 2. The one open item

`d2c51f41-8095-4d25-b1dd-a5da6340e9b1` — submitted the **prefix-uniformity
standing check** (migration 378, not yet written): promotes migration 377's
one-line invariant (`count(DISTINCT md5(prefix)) = 1`) from a one-shot guard
into a 6-hourly `scheduled_tasks` check. It closes LCO-008's own registered
open question *and* the `bug_historian` seat's "decorative detection, not
enforcement" objection.

```sql
SELECT metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='d2c51f41-8095-4d25-b1dd-a5da6340e9b1' AND kind='council_report';
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

**VERDICT: REVISE**, gated by an `editquality` HIGH. **Seven seats independently
flagged the same thing**, and it is a genuinely good catch: my proposed check
walks `jsonb_each(default_config->'workflow'->'steps')`, which this estate has a
standing landmine against — **that path is TOP-LEVEL ONLY** and cannot see steps
nested under a `sub_workflow`/loop. A check written that way would return zero
rows and **read as healthy while inspecting nothing** — precisely the silent
failure the check exists to prevent, now hidden behind a green tick instead of an
absent one.

**I verified both HIGH objections against the live row before deciding what to
do, and they do not apply to council-gate TODAY:**

- **One active row only** — `be2a7614…`, version 2, `is_active`, not a snapshot.
  So the "four agent types carry two active rows, only the higher version loads"
  landmine does not bite here, and migration 377 updated the row that actually
  executes.
- **All 17 seats are genuinely top-level** — `sub_workflow`, `substeps` and
  `loop` appear **nowhere** in council-gate's workflow.

**So nothing already shipped is wrong.** Independently corroborated by the
strongest possible evidence: caching demonstrably works in production (74.4%
measured), which it could not if 377 had updated a row nothing runs, or missed
nested seats.

**But the objection still stands for the CHECK**, and this is the distinction
worth carrying: migration 377 was a one-shot transform whose result I verified
directly, so a top-level-only walk was adequate. A **standing** check must stay
correct as the config changes around it — if the seats are ever moved under a
loop, the naive walk goes blind exactly when it matters.

A resubmission (`RESUBMIT_CORR=d2c51f41-8095-4d25-b1dd-a5da6340e9b1`) needs:
1. **A nested-aware walk**, or an explicit assertion that seats are top-level
   PLUS a guard that fires if that ever stops being true. Do not simply assert
   "they are top-level today" — that is the assumption the objection is about.
2. **Highest-version scoping** per agent type (cheap, and correct even though
   council-gate has one row today).
3. **Idempotency** — `ON CONFLICT (name) DO UPDATE`/pre-existence guard, so a
   re-applied 378 is a 0-row no-op rather than a duplicate task double-firing.
4. **`target_agent_type` checked against the two precedent rows** — the guardian
   seat notes a `generic`-consumed topic cannot select by agent_type and may
   silently run the wrong workflow, i.e. the check could misfire at the one
   moment it has something to report.

⚠ That submission used `FORCE=1` because it touches only docs/SQL — deliberate,
and stated in its own rationale, because it doubled as the live cache test.

⚠ `agent_definitions.updated_at` for council-gate read **21:42:43**, *after*
migration 377 ran — something else touched the row (a roster sync or the deploy).
**My changes survived intact** (re-verified: 17/17 hoisted, 17/17 marked, 1
distinct prefix). Worth knowing that row is written by more than this lane.

## 3. Traps carried forward (all hit for real today)

- **A zero `cache_read_input_tokens` is THE failure mode, and it looks exactly
  like success.** A single per-seat byte above the marker makes every seat write
  its own entry and read none — *worse* than no caching, no error, correct
  verdicts throughout. Assert a NON-ZERO read on the 2nd+ seat.
- **`input_tokens` now means the UNCACHED REMAINDER**, not prompt size. True size
  is `input + cache_creation + cache_read`. Audited 2026-08-10: only two
  `scheduled_tasks` pre_queries touch `llm_call_log` and **both read
  `output_tokens` only**, so nothing live is affected — re-run that audit before
  assuming it stays true.
- **NULL ≠ 0 in the cache columns.** NULL = binary predates cache support;
  0 = no cache used. Do not backfill.
- **Editing a live agent's config CANNOT reach an in-flight orchestration** —
  it executes from its own `workflow_plan` snapshot. So config edits are safe
  any time, AND a hot-fixed prompt will not affect a run already going. Now a
  fleet-wide LANDMINES entry.
- **Do not reintroduce `ttl:"1h"`** without first confirming the current beta
  header. It needs one; without it the first cached call 400s, and in
  council-gate a 400 takes out the review path for the whole estate.
- **A type-only test assertion sails through a whole class of defect.** My
  marker test asserted the content was still a string (true) and never that the
  marker was gone (false). The council found it by reading code, not by running
  tests.

## 4. Do NOT do this (measured, counterintuitive)

**A mixed model roster costs MORE than staying on Sonnet.** Caches are
model-scoped, so a second model pays a second full-price cache write (~125k
tokens) to save 50% on a handful of 10k-token reads. Per run: all-Sonnet-cached
**$0.43**, mixed 5/5 **$0.49**, all-Haiku **$0.21**. The real choice is
all-or-nothing, and after caching the remaining model saving (~$4/day) is small
against the quality risk. **Owner ruled: stay on Sonnet.**

Supporting data if it is ever revisited — 30 days of per-seat objection rates:
`guardian` 73% (1,207 objections), `bug_historian` 84% (663), `editquality` 56%
(907), `prior_art_librarian` 45% … versus `mission` **1 objection in 494
rounds**, `diagnosis_guardian` 2 in 270, `constitution` 20 in 494. ⚠ A low rate
has two causes with opposite implications (rubber-stamp vs rare-event detector)
and counts alone cannot separate them.

## 5. Next, if anyone picks this lane up

1. Resolve `d2c51f41` (§2). That is the whole of the outstanding work.
2. Optional: a **runtime** check on the tell — alert when a council run's 2nd+
   seat shows `cache_read = 0`. Deliberately not proposed yet; there was no
   production data to calibrate against when 378 was written, and there is now.
3. Optional: the marker is council-gate-only. If a second agent ever adopts it,
   378's check is scoped by name and will silently not cover it.
4. Not this lane's: `getJSONOutputInstructions()` shows a *site-classification*
   worked example to every JSON agent in the fleet, including council reviewers
   emitting verdicts. ~$0.08/day, so no cost case; owner ruled don't remove it
   (it is what stops markdown-fenced JSON). Prompt-quality wart only.
