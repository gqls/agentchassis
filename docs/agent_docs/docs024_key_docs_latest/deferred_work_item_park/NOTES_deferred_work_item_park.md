# NOTES — the `deferred` work-item park

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-25 — lane opened out of `bugs_open/328`, and the first framing was already too big

Opened after 328's closure hit one of these rows. The owner asked for the general case.

### How it started, and the number I led with was three populations stacked

My first report to the owner said **"297 deferred rows, every one with `attempt_count = 0`, 205
naming a real handler"**, and framed all of it as *"jobs nothing will ever pick up"*.

That was too big, and I knew part of it within minutes of looking properly. The 297 is **three
unrelated populations**:

| population | rows | verdict |
|---|---|---|
| `deferred` + **empty** handler, most stamping `spec.not_dispatchable` | 75 (of the 216 unstamped) | **CORRECT BY DESIGN** — the estate's roadmap convention |
| `parked_by = migration_389`, all `contrast_failure` | 87 | **traceable AND owned** (`bugs_open/296`, lane `bugfix_131_contrast_ratio_check`, ACTIVE) |
| named handler, no `parked_by`, no `not_dispatchable` | **118** | the actual question |

**The first population is not damage and saying so would have started an argument with six
well-commented code sites.** `discovery_checks/remit.go` calls it a "double lock"; the convention
has a live consumer (`diagnose_triage_action.go:361`, `fixloop_digest_action.go:358`, both reading
`(item_type='capability_gap' OR status='deferred')`) and a live drain
(`work_item_retraction.go:205`, which counts `parked` separately *because* "the park draining
unnoticed is exactly what this counter exists to make visible"). That is a designed mechanism with
both ends wired, and my headline had it as a leak.

⚠ **The transferable half: a status is not a population.** `WHERE status='deferred'` looked like one
finding and was three. The discriminator that separates them — `spec ? 'parked_by'` and
`handler_agent <> ''` — costs one extra `GROUP BY` and I did not have it in the first query.

### MISSTEP 1 — I ran a test that structurally could not answer its question, and briefly believed it

To ask *"were these rows born `deferred`, or moved there later?"* I measured
`updated_at - created_at` and got **"0 of 205 born deferred"** — a clean, decisive-looking result
that I nearly wrote into a bug file.

**It means nothing.** `trg_site_work_items_updated_at` is BEFORE UPDATE FOR EACH ROW and bumps
`updated_at` on *every* write, so a row born `deferred` and later touched by anything is
**indistinguishable** from one created in another status and deferred later. `site_work_items`
keeps no status history, so the question has no answer from that column at all.

What caught it: recognising the shape from a landmine I had read this morning — *"a periodic write
to an open work item makes it UNREAPABLE for ever; `trg_site_work_items_updated_at` bumps
`updated_at` on EVERY write"* (`bugfix_213`). The same trigger, the same column, a different wrong
conclusion.

**This is the second time in one session I have taken an arithmetic or inferential shortcut on my
own output and had to retract it** (the first: the 328 closure's "12 dead anchors / 13 of 21
pages", both wrong, `WRONG_CALLS.md`). Both were caught, neither by re-reading — by a second
instrument disagreeing.

### The clustering is real, and it is still unexplained

Deferred rows with a named handler cluster into **ten one-minute events**, mostly **one site across
many item types**:

| event | site | rows | item types |
|---|---|---|---|
| 08-12 13:31 | loancalculator.co.uk | 43 | 8 |
| 08-11 12:31 | *fleet-wide, 14 sites* | 87 | 1 (`contrast_failure` — this is migration 389) |
| 08-11 18:20 | loancalculator.co.uk | 17 | 3 |
| 08-02 23:31 | mortgagecalculator.co.uk | 16 | 3 |
| 08-04 22:03 | idea.uk | 14 | 3 |
| 08-03 11:02 | mortgagecalculator.co.uk | 12 | 1 |

None of the non-389 rows carries `handled_by` or `error`. **[UNMEASURED]** whether these are bulk
*parks* or bulk *touches* — see misstep 1; the timestamps cannot tell them apart.

### What IS established first-hand, by reading every site rather than grepping

- **No Go path anywhere does `UPDATE … SET status='deferred'`.** The admin endpoints
  (`site_admin_handlers.go`) only set `complete` and `triaged`.
- **All six Go writers of work-item `deferred` pair it with `HandlerAgent: ""`**, deliberately and
  with comments: `remit.go:202`, `write_audit_findings_action.go:427` and `:584`,
  `load_work_item_actions.go:279`, `check_palette_contrast.go:138`,
  `check_content_duplication.go:251`.
- ⚠ **`plan_sections_action.go`'s four `deferred` hits are a DIFFERENT `deferred`** — a section-plan
  status (`"ready" | "deferred" | "skipped"`, `:906`), not a work-item status. Counting them would
  have made it look as though a Go path *does* produce the shape, inverting the conclusion. Caught
  by opening the file instead of trusting the grep line.
- `refreshOpenWorkItem` (`load_work_item_actions.go:~2116`) updates **description only** — status,
  priority and handler are explicitly untouched — and only evidence/citation paths use
  `refreshOnConflict`, none of the item types in question.

So the shape has **no producer in the codebase**, which is exactly the kind of claim that should not
be filed on my own reading.

### Filed the diagnosis loop rather than the bug

Owner ruling 2026-07-31: a `bugs_open/` file asserting a cross-cutting or structural root cause is
not filed until it has been through `090`, or the session states plainly why it substituted
equivalent first-hand verification. This is squarely that class — cross-cutting, and the cause is
by definition *not* where the symptom is, since the symptom is rows and the cause is whatever wrote
them.

- intake `4623672c-d942-4dfe-a7a4-41bdbf500c5c`
- run `6061299a-cb6a-497f-b5eb-d31b3bb7771c` ← the key artifacts are written under

Symptom authored to the house rules: states the MECHANISM, points at the tables and symbols, asserts
**no counts** (the loop fetches and cites them), no downstream-consequence clauses, one bug, and
explicitly excludes `parked_by=migration_389` as owned by another active lane.

### Migration 389 is the model, and it deserves saying out loud

`389_park_contrast_failures_and_reenable_improvement_sweep.sql` is the one bulk park in the estate
that can be audited after the fact: a precondition that `RAISE EXCEPTION`s if the premise is already
gone, `spec.parked_from_status` / `parked_reason` (naming the bug AND the restore condition) /
`parked_by`, a `GET DIAGNOSTICS` row-count assertion against the pre-count, and a negative control
proving nothing else moved. Every one of its 87 rows is traceable 14 days later. **Whatever made the
other 118 left nothing.**

### While the loop ran: a path that CAN write the shape, and does not

My "no Go path writes `deferred` with a named handler" rested on one grep pattern
(`SET status ... deferred`), which a **parameterised** update would walk straight past. Checked it
properly — every `UPDATE site_work_items … SET status` in the repo, with its value:

```bash
grep -rn -A4 "UPDATE site_work_items" --include=*.go platform/ internal/ pkg/ cmd/ \
  | grep -E "status *= *\\\$"
```

Three parameterised hits. Two are `v3_site_actions.go:6302/6312` — `UpdateWorkItemStatusAction`,
whose `$2` is the caller's `newStatus`. The third is the interesting one:

**`load_work_item_actions.go:1259` — `FailWorkItemAction` honours a step-config key
`status_override`**, and writes it straight into `status`:

```sql
UPDATE site_work_items SET error = $2, status = $3, handled_by = $4 WHERE id = $1
```

It touches `error`, `status` and `handled_by` — and **leaves `handler_agent` alone**. So a step
configured `status_override: "deferred"` produces *exactly* the shape under investigation: parked,
named handler, no `parked_by`.

And the comment immediately above it names the agents that use the key —
*"component-template-fixer ×2, page-build-handler, tool-improver"* — which are **precisely** the
handlers on the untraceable rows (`page_component_status_drift` → `component-template-fixer`,
`improve_tool` → `tool-improver`, `content_rewrite`/`needs_page`/`needs_content_page` →
`page-build-handler`). That is a very good-looking lead.

**And it is not the answer.** Two things refute it, and I checked both rather than stopping at the
resemblance:

1. **`FailWorkItemAction` stamps `handled_by = agentType`.** Every one of the bulk-parked rows has
   `handled_by` NULL/empty. A row written by this path would name its writer.
2. **Every live `status_override` in the fleet is `needs_human_review`, not `deferred`** — read
   with a recursive walk over `agent_definitions`, all four of them:
   `component-template-fixer>judged_refusal`, `component-template-fixer>park_refused`,
   `page-build-handler>mark_needs_review`, `tool-improver>refuse_mangled_write`.

> ⚠ **What this DOES establish, and it is worth a landmine on its own: the black hole is ONE CONFIG
> KEY away.** `status_override` is an ordinary step-config string with no allow-list — nothing
> validates it against the statuses the dispatcher, the promoter and `idx_swi_dedup` actually
> understand. A session setting `status_override: "deferred"` on any refusal step would silently
> mint undispatchable, un-promotable, un-re-filable rows at production rate, and every field on
> them would look healthy. The four live values are `needs_human_review` **by convention, not by
> constraint.**

So the shape still has no live producer, and the resemblance that looked decisive was a near-miss.
**[UNMEASURED]** what actually wrote the 118 — that is the loop's question, and I have deliberately
not guessed at it here. The comfortable answer (earlier sessions running `psql` by hand) remains
untested, and I have no evidence for it beyond the absence of alternatives, which is not evidence.

### `FailWorkItemAction` conclusively ruled OUT — with a control that proves the instrument works

The path stamps `handled_by = agentType` unconditionally. So "do the 118 carry `handled_by`?" is a
decisive test — **provided the column is actually written somewhere**, or a zero means nothing.
Both halves in one run:

| the 118 (deferred, named handler, no stamp) | |
|---|---|
| rows | **118** |
| with `handled_by` | **0** |
| with `error` | **0** |
| with `attempt_count > 0` | **0** |
| ever `triaged_at` | **1** |
| ever `claimed_at` | **1** |

**Positive control — is `handled_by` written at all?** Yes, heavily: **7,114 of 7,329** `complete`
rows carry it, plus 156 of 732 `cancelled`, 131 of 963 `needs_human_review`, 76 of 179 `failed`.
So the zero above is a real absence, not a dead column. (All **303** `deferred` rows carry none,
migration 389's 87 included — consistent, since a migration would not set it.)

**Two things this establishes:**

1. **`FailWorkItemAction` did not write these rows.** Not one carries its fingerprint.
2. **117 of the 118 never entered the dispatch queue at all** — never triaged, never claimed, never
   attempted. So the producer acted on rows in a *pre-dispatch* state, or created them parked. This
   rules out "dispatched, then parked after a failure", which was my second-favourite hypothesis
   and the one the `error` column would have evidenced.

⚠ It still does **not** distinguish born-deferred from moved-to-deferred — `triaged_at` is NULL in
both cases. That question may simply be unanswerable from this table, and the bug file must say so
rather than choosing the comfortable answer.

### Two more candidates raised and killed, and one real correlation left standing

**Candidate: migration `217_site_work_items_handler_agent_not_null.sql` backfilled handler names
onto correctly-parked rows.** This would have dissolved the whole finding — the 118 would be
honest roadmap rows whose empty handler was later filled in, and the "wrong shape" would be an
artefact of the migration rather than a defect. **REFUTED by reading it**: 217 backfills
`handler_agent = ''` **WHERE handler_agent IS NULL** — it collapses NULL onto empty, the opposite
direction, and then sets `DEFAULT ''` + `NOT NULL`. It cannot put a name on anything.

**Candidate: a later router stamps `handler_agent` onto rows born deferred-and-empty.** No such
writer exists — nothing in `platform/`, `internal/`, `pkg/` or `cmd/` does `UPDATE … SET
handler_agent = <a name>`. The only occurrences are `claim_work_item_action.go:173` (setting an
`error` when a handler is *missing*) and test fixtures.

**The correlation that IS real, and is the best lead left.** `agent_error_log` retains to
2026-07-24, so it covers every bulk-park minute. At **08-04 22:03–22:04**, the minute idea.uk's 14
rows were parked, idea.uk's `completeness-discovery-agent`, `design-discovery-agent` and
`quality-discovery-agent` all logged `complete` — **a full discovery run finishing on that exact
site at that exact minute**, and every one of those parked rows carries `source='discovery'`.

⚠ **[UNMEASURED] and I am deliberately not concluding from it.** A discovery run completing at the
same minute is consistent with the discovery write path parking them — and equally consistent with
discovery merely *touching* them (bumping `updated_at`) while they were already parked, which is
misstep 1's trap wearing a new hat. **The same ambiguity, one layer along: I still cannot tell a
park from a touch, and a co-occurring actor is not a writing actor.** Two of my four hypotheses
today have died on exactly this distinction; the third would too if I let it.

What is now excluded, so the loop's answer can be checked against it:

| candidate | verdict | on what evidence |
|---|---|---|
| `FailWorkItemAction` + `status_override` | **OUT** | stamps `handled_by`; 0 of 118 carry it; all 4 live values are `needs_human_review` |
| migration 217 backfill | **OUT** | backfills to `''`, not to a name |
| a later `handler_agent` router | **OUT** | no such writer exists in the repo |
| `refreshOpenWorkItem` | **OUT** | description only; and none of these item types uses `refreshOnConflict` |
| dispatched-then-parked-on-failure | **OUT** | 117 of 118 never triaged, claimed or attempted; 0 carry `error` |
| a discovery-run side effect | **OPEN — best lead** | timing correlation only |
| a hand-run `psql` UPDATE | **OPEN — untested** | no evidence beyond absence of alternatives, which is not evidence |

## 2026-08-25 (later) — the loop came back UNVERIFIABLE, and it was worth every credit

Run `6061299a`, 4 iterations, then the iteration cap. Verdict verbatim:

> **NOT CONFIRMED (stopped: iteration-cap)** … *"That leaves two writers still unidentified and now
> with zero remaining named candidates in the read code"* … *"Hand to a human with the full trail;
> do NOT auto-conclude."*

**A REFUTED or unconfirmed verdict is a result, not a waste** — and this one paid for itself twice.

### What it caught that I had missed — and the method error underneath it

It surfaced a **third** park-provenance convention: `spec.deferred_reason`, carrying an
owner-sanctioned explanation (`canary_replan_407`, corr `b23b19c7`, the 08-12 queue-parking
decision). Four rows, `needs_imagery` + `needs_rerender`.

Numerically that is nothing — 118 → **114**. **The method error is the finding.** I had checked
`spec ? 'parked_by'` and `spec ? 'not_dispatchable'`: *the two conventions I already knew about*.
The right query enumerates:

```sql
SELECT k, count(*) FROM site_work_items w, LATERAL jsonb_object_keys(w.spec) k
WHERE w.status='deferred' AND COALESCE(w.handler_agent,'')<>'' GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **A membership test can only find the members you can name.** I built a discriminator out of my
own prior knowledge and then reported its output as a census of what exists. That is the same
family as this morning's errors and it is the third instance today: **the instrument encoded the
question I already had, not the one I asked.** Now in the RUNBOOK.

And the enumeration immediately paid again, in the *other* direction: `spec.reason` is on 22 of the
114 and looks like provenance. It is not — it records why the item was **detected**
(`cta_links_stale`, `not_built`, `no_style_collection`). Treating it as a trace would have shrunk
the population on a false basis, which is the same error with the sign flipped.

### Where the loop was WRONG, checked rather than inherited

Its "still needed" asks for *"the full body of `HandleUpdateWorkItem` — the only call site the index
shows that SETs `handler_agent` on an UPDATE"*.

**No such symbol exists.** `grep -rn "func Handle.*WorkItem"` returns nothing repo-wide; every
`handler_agent` occurrence under `internal/core-manager/admin/` is an INSERT column list. The loop
had flagged its own caveat — only a signature was indexed, not a body — and that caveat is the
tell: the code index lags the working branch.

⚠ **This is exactly the shape CLAUDE.md warns about in the other direction.** The loop refuted a
thread's confident claim on 2026-07-19 by reading the one function the thread had skipped; here the
thread refuted the loop by grepping for a symbol the loop had only seen a signature of. **Neither
is authoritative. The artefact is.** I would have burned a round on that lead had I taken the
verdict as read.

### So: bug filed with the root cause NOT established, and saying so

`bugs_open/396`. Eight candidates excluded with their evidence; two left open and both labelled —
a discovery-run side effect (timing correlation only) and a hand-run `psql` UPDATE (no evidence
beyond the absence of alternatives, which is not evidence). The header says root cause not
established rather than picking the comfortable one.

**Tally for this lane, since the missteps are the point:** four hypotheses raised, four killed
(`FailWorkItemAction`, migration 217, a `handler_agent` router, dispatched-then-parked), one
measurement retracted as structurally unanswerable (`updated_at - created_at`), one census
corrected by an instrument I had not thought to run, and one inherited lead verified as void. The
useful residue is not the cause — it is the eight exclusions and the two controls, which are what
the next reader would otherwise re-derive.

## 2026-08-25 (later still) — an independent Fable review refuted the filing within the hour, and both errors were mine to have avoided

The owner commissioned a second model to reconsider the whole lane. It refuted the central claim
and **answered the question that had defeated both me and the diagnosis loop**. I re-verified every
load-bearing point first-hand before accepting any of it.

### MISSTEP 2 — I wrote the lesson, then applied it to one column

`396` §5 records, in my own words, the lesson the diagnosis loop taught me hours earlier: *"a
membership test can only find members you can name — do not test for the stamps you know,
ENUMERATE the stamps that exist."* I put the `jsonb_object_keys` query in this lane's RUNBOOK,
added the pattern to `016b` §9, and saved it to memory.

**I ran that enumeration against `spec`. I never ran it against `result`.**

```sql
SELECT k, count(*) FROM site_work_items w, LATERAL jsonb_object_keys(w.result) k
WHERE w.status='deferred' AND COALESCE(w.handler_agent,'')<>'' GROUP BY 1 ORDER BY 2 DESC;
--  deferred_by 62 | deferred_reason 60 | deferred_from_status 60 | reason 2 | deferred_at 2
```

**62 of the 114 rows I had just published as *"carrying no trace of any kind"* are fully stamped** —
`loancalculator_rebuild_thread` (60, with `deferred_from_status='detected'`, a reason naming the
owner-ordered rebuild, and an explicit release condition *"un-park after rebuild verify"*) and
`apis-uk-bees-lane` (2, with `deferred_at` and a paragraph of reasoning). Genuinely unstamped: **52**.

`result` is not an exotic place to look. Migration `442` stamps `result.repair_284` on this same
table for exactly this purpose.

⚠ **The shape is worse than not having known the lesson: I generalised the fix to the INSTANCE that
taught it (the `spec` column) instead of to the CLASS (provenance can live in any JSONB column on
the row).** A rule written down after a miss is not learned until you ask *where else does this
apply?* and answer it with a query. Two `jsonb_object_keys` calls instead of one.

### MISSTEP 3 — I searched the code for a writer, on a tree whose writers leave notes

`396` §4 listed *"a hand-run `psql` UPDATE by an earlier session"* as **OPEN, untested, "no evidence
beyond the absence of alternatives, which is not evidence."**

The evidence was in the repo, dated, and weeks old:

- `mortgagecalculator_couk_adoption/HANDOFF_2026-08-03_continue_here.md:81-90` — *"every 15s, defer
  anything dispatchable"*, verbatim `UPDATE site_work_items SET status='deferred', updated_at=NOW()
  WHERE … AND status IN ('triaged','approved')`. ⚠ **It sets nothing else — which is exactly why
  those 38 rows carry no stamp.** Corroborated by that lane's own NOTES at `:305`, `:462-473`,
  `:526`, whose named counts (3 `needs_page` + 1 `needs_rerender`) match the surviving residue.
- commit `90a4fb812` (08-04, idea.uk) — *"Holds: 12× undeployed_asset + footer needs_rerender +
  deactivated_component → `deferred` (UPDATE 14)"*, matching idea.uk's 14 rows type-for-type.

I ran the Go grep a dozen times. **I never ran `grep -rn "SET status='deferred'" docs/`**, which
lands on the answer immediately. **"No code does this" and "nobody did this" are different
statements**, and CLAUDE.md's opening fact — many sessions, one tree, one live database, worked by
hand — is precisely what makes them different here.

### And my "best lead" was refuted by two timestamps

idea.uk's 14 were **created 07-31 19:18** and **parked 08-04 22:03** [MEASURED]. The discovery run I
found completing at 22:03 was co-occurring with the **park**, not the birth — and the park was a
human's `UPDATE 14`. I had already written in this file that *"a co-occurring actor is not a writing
actor"* and then leaned on the correlation anyway, because it was the only lead I had left.

### The detail worth keeping longest

`mortgagecalculator_couk_adoption/NOTES…:2844` reads: **"[UNVERIFIED] what deferred them … a
hand-park at adoption is the obvious guess and I did not establish it."** Written by the lane whose
own handoff, in the same directory, carries the recipe it ran. **The lane forgot its own backstop,
and I inherited the marker and re-derived the mystery from scratch.**

⚠ **Someone else's `[UNVERIFIED]` is a claim about what THEY checked — a lead to close, not a
boundary to respect.** The cheapest first move is to check whether the marker's own lane already
answered it. Contributed back to them today.

### What survived the attack, and what the lane is now worth

Unchanged and re-verified: §3's three predicates; the `idx_swi_dedup` reading; migration 217; the
`claim_work_item_action.go:102` clause; `FailWorkItemAction`'s exclusion; and — importantly — that
**no Go path produces the shape**. But that last one now rests on the right evidence: a *config
walk* over every `agent_definitions` row (two actions take a status from step config —
`create_work_item_action.go:194-222` and `UpdateWorkItemStatusAction`), not on a source grep, which
structurally could not have falsified it.

**The reframing is the useful output**: the cause is not a defective writer, it is a **missing
verb**. Four lanes each needed to hold a site's queue, none had a supported way to do it, so each
improvised the same `UPDATE` — and only the ones who thought of it left a stamp. The estate now
carries **six** competing ad-hoc provenance conventions for one act. That is a better argument for
building the verb than the 52 unstamped rows ever were.

**Tally for this lane: 3 hypotheses killed by me, 1 of my own claims killed by a peer, 1 measurement
retracted as unanswerable, 2 census errors — and the answer found in a document I never opened.**

## 2026-08-25 — BOTH FIXES BUILT AND SHIPPED: the park verb (live) and the status_override allow-list (inert until a roll)

Owner ruled "2 then 3": build the verb, then close the config door.

### The verb — migration 621, LIVE ON APPLY

`park_work_items(...)` / `unpark_work_items(...)`, register **WII-034**, council `ed821065`.

**A SQL function, not a Go action, and that is the load-bearing design decision.** The users are
sessions at a psql prompt — all four recorded parks were performed that way. A registered action
would not have been called by any of them, and would have been inert until an image rolled. This is
DB config: live the moment it applies.

**The structural property is that the stamp is an ARGUMENT.** `p_parked_by`, `p_parked_reason` and
`p_release_condition` have no defaults, so a park cannot be performed without saying who, why and
what would release it. Everything else in the file is guardrails.

⚠ **It stamps `result`, not `spec`, and the reason is a real finding rather than a preference.**
`refreshOpenWorkItemSQL` does `SET summary = $3, spec = $4::jsonb` — it **replaces `spec` wholesale**
whenever a `refreshOnConflict` producer re-detects the same `item_key`. A `spec`-based park stamp is
therefore destroyed by exactly the re-detection a parked row is most exposed to. `result` is merged
by the estate's main status writer (`v3_site_actions.go:6306`) and untouched by the refresh, and
migration `442` already stamps `result.repair_284` there. **That explains why the two hand-conventions
diverged and which of them was right.**

**Verified before applying, then at the artefact after.** Ran the whole file with `COMMIT` replaced
by `ROLLBACK`: 5 refusals induced, positive control passed, negative control clean, and a check
afterwards confirmed the rollback left nothing behind. Then applied; then read back independently —
both functions present in `pg_proc` with the right signatures, and a bare call on a real site
(`mortgagecalculator.co.uk`) returned:

```json
{"applied": false, "would_park": 30, "domain": "mortgagecalculator.co.uk",
 "by_item_type": {"head_essentials_missing": 29, "dead_internal_link_live": 1},
 "note": "DRY RUN — nothing changed. Re-run with p_apply => true to write."}
```

…and wrote **0** rows. Recorded via `--record-only` with the reason stated: applied by hand rather
than through an unscoped runner pass, which would have taken every other session's pending file.

### The allow-list — Go, inert until an image rolls

`workItemStatusOverrideAllowed` + `statusOverrideAllowed()` in `work_items_common.go`, wired into
`FailWorkItemAction`. Council `9c16eb83`.

**The rule each entry must meet is stated and TESTED:** a status may be written here only if the row
can still *leave* it — terminal (releases the dedup slot, so the detector can re-file) or with a
named live consumer. `needs_human_review` has `HandleRetryWorkItem` and the resolve endpoint;
`blocked` has `feasibility-recheck`. **`deferred` has neither, is excluded, and the refusal message
points at the park verb** — so the narrowing names the good way to park instead of only saying no.

**Blast radius measured, not assumed:** a **recursive** walk over every `agent_definitions` row,
snapshots and soft-deleted included, finds `status_override` on 4 steps in 3 agents and **every value
is `needs_human_review`**. No other value in the estate's history.

### The tests were mutation-proven, and one mutation is the interesting one

All five made to fail, then the tree restored **byte-identically** (verified with `diff -q` against
copies taken before mutating — a shared tree must not be left mutated even for a minute):

| mutation | result |
|---|---|
| add `"deferred"` to the list | 2 tests FAIL ✓ |
| remove `"needs_human_review"` (over-narrowing) | over-narrowing guard FAILS ✓ |
| replace the guard condition with `if false`, **every comment left in place** | wiring test FAILS ✓ |

⚠ **The third is why that test parses the AST instead of grepping the source.** The estate's existing
test for this same function is a source scan, and LANDMINES records that a source-scanning test makes
your own comments load-bearing — the needle matches the comment explaining the guard and the test
passes vacuously. `parser.ParseFile(..., 0)` drops comments, so it can only pass on real code. The
`if false` mutation is precisely the one a grep would have missed.

### Housekeeping, and one thing that went sideways

Register entry WII-034 written in the **same commit** that ships the verb, per the 2026-07-28 ruling
condition (2). The lane's `102_coverage_ratchet.txt` line retired — it has now built rather than
diagnosed, so it is registrable, not ratchetable.

⚠ **My `000_concept_index.md` row for WII-034 was swept into another session's commit (`ce969cf9c`)
between my write and my commit** — the same-file passenger case CLAUDE.md says no hook can prevent.
Nothing lost, forward-only holds, the row is correct in HEAD, and my commit message says so. Worth
noting as the second concurrency event this lane has hit today (the first was a transient
`index.lock` that failed a commit outright).

### What is NOT done, stated plainly

- **Neither fix releases the 52 unstamped rows or the 62 stamped ones.** That is deliberate: 60
  carry another lane's live *"un-park after rebuild verify"* condition, and both sites holding the
  52 now show a zero live queue, so the holds look expired — **but expired is the holder's call, not
  mine.**
- **Nothing stops a session writing a raw `UPDATE` instead of calling the verb.** Short of a trigger,
  nothing can. Stated as the open residual in WII-034 and in `bugs_open/396` §6a rather than left
  implied.
- The Go half is **inert until an image rolls**. The SQL half is live now.

## 2026-08-25 — council REVISE on the verb, and the objection is right: a site lock already exists

`ed821065` → **REVISE**, gated by `prior_art_librarian` (HIGH). The allow-list round (`9c16eb83`)
→ **APPROVED**, 1 advisory + 3 low/medium.

### The gating objection, verified rather than accepted

> *"Rationale's core premise — 'The platform offers NO verb for holding a site's work queue' — is
> contradicted by an existing mechanism already in the Schema: `sites.locked_at` / `sites.locked_by`."*

**It is right, and I checked every part of it myself:**

- `sites.locked_at timestamptz` / `locked_by text` exist.
- **Live on 3 of 51 sites right now**, and `locked_by` carries a real reason —
  *"portfolio_positioning: owner HALT 2026-08-18 pending classifier register-input (RFC)"*.
- `build-pipeline-trigger > find_dispatchable_site` gates on **`s.locked_at IS NULL`**.
- `internal/core-manager/admin/site_admin_handlers.go:425/454` are lock/unlock endpoints.
- **It mutates no work-item row.** Nothing is stranded, no dedup slot is held, no 23505.

⚠ **So the 22-day stall that justified my whole submission could not have happened under the
mechanism that already existed.** My verb makes the *wrong* tool tidier. Third asserted-absence
error today — and the first one to reach an owner-approved plan and a live migration.

### What is ACTUALLY missing, which is much narrower

**"Hold this site's queue EXCEPT these N items."** `sites` has only `locked_at` and `locked_by` —
no exception list, so the lock is all-or-nothing. That is exactly why the mortgagecalculator lane
built its 15-second backstop; its own handoff labels the site lock **(a)** and item status
**"(b) ITEM STATUS — the finer control"**, and its comment even notes *"'deferred' is NOT in
workItemTerminalStatuses, so the row keeps its idx_swi_dedup slot and release is one UPDATE"* —
**they knew about the slot and treated it as a feature.**

### ⚠ AND THE OBVIOUS FIX IS A TRAP — the finding that makes this worth writing down

The natural change is one clause on the dispatch gate:

```sql
-- find_dispatchable_site, today:
WHERE s.locked_at IS NULL AND wi.status IN ('triaged','approved') AND ...
-- the "obvious" extension:
WHERE (s.locked_at IS NULL OR wi.id = ANY(s.lock_except_item_ids)) AND ...
```

**That would unlock the entire site.** `find_dispatchable_site` selects a **site**, not an item.
The loop then calls `load_items` → `LoadWorkItemsAction(site_id, max_items: 5)` — and
**`LoadWorkItemsAction` does not check the site lock at all**: its WHERE is `wi.site_id = $1 AND
wi.status IN ('triaged','approved') AND …`, with no `locked_at` anywhere. So the moment one
excepted item makes the site selectable, every dispatchable item on it is loaded and claimed.

⚠ **The lock is enforced at exactly ONE gate today** (`find_dispatchable_site`). The Go check at
`load_work_item_actions.go:134` is **not** a second gate — it is inside **`WriteBuildItemsAction`**
(minting new build items), and **its log line misnames itself as `"LoadWorkItemsAction: site is
locked, skipping"`**, which is how I misread it the first time. Do not trust that log string.

### So the correct design, with its ordering constraint

1. `sites.lock_except_item_ids uuid[]`, default NULL → `wi.id = ANY(NULL)` is NULL, never true, so
   behaviour is **byte-identical to today** until someone lists ids. Opt-in, unsafe side OFF
   (RFC_010 §2).
2. **Go**: `LoadWorkItemsAction` becomes lock-aware — on a locked site it returns **only** the
   excepted items instead of everything.
3. **Config**: `find_dispatchable_site` gains the exception clause.
4. ⚠ **ORDER MATTERS AND THE CONFIG HALF MUST BE HELD.** If the pre_query ships before the binary,
   a locked site with an exception list dispatches its whole queue — the exact failure the lock
   exists to prevent. Go first, roll, *then* apply the config as a `_HOLD.sql`. This is migration
   `575`'s pattern and the same reason it was a `_HOLD` file.
5. `WriteBuildItemsAction`'s lock check stays all-or-nothing: an exception list is about
   dispatching named EXISTING items, not about minting new ones.

**This is architecture-scope**: it changes what a shared dispatch gate GUARANTEES for every site,
so it needs its own council round on its own merits, not a resubmission of the verb's.

### The other three objections, all checked

- **`blocked` has a live consumer** — the guardian asked and was right to. Verified with the query,
  not the landmine I originally took it from: `scheduled_tasks` row `feasibility-recheck`,
  `enabled = t`, `pre_query`: `UPDATE site_work_items SET status='triaged' WHERE status='blocked'
  AND EXISTS (SELECT 1 FROM agent_definitions WHERE type = wi.handler_agent …)`. ⚠ Note the
  condition — a `blocked` row whose `handler_agent` names no live agent is NOT promoted.
- **A verb-park still holds the dedup slot** (editquality, medium) — correct, and I had explicitly
  deferred that as candidate 5 while citing the 23505 as the urgency. The site-lock route fixes it
  properly, because no row changes status at all.
- **Retraction can silently drain a verb-park** (prior_art, medium) — correct.
  `work_item_retraction.go:205` counts `parked += n` when it closes a `deferred` row, so an
  automated sweep can release a deliberate hold without going through `unpark_work_items`. That
  breaks the single-releaser guarantee the verb was built around, and it is another reason the
  lock — which touches no row — is the better mechanism.

## 2026-08-26 — the roll landed, both halves are LIVE, and the council round died on a fleet outage (not on its content)

### The binary carries both changes — proven, both replicas, with controls

Chassis **`v1.0.1341`**, pods 9 h old. ⚠ The `build provenance` line had **scrolled** — an empty
`grep` there means *"not in range"*, **not** *"unstamped"* — so the binary probe is the evidence:

| probe | replica 8h8b7 | replica gmj4b | meaning |
|---|---|---|---|
| `honour_site_lock` | PRESENT | PRESENT | the site-lock arm |
| `WORK_ITEM_STATUS_OVERRIDE_REFUSED` | PRESENT | PRESENT | the allow-list |
| `lock_except_item_ids` | PRESENT | PRESENT | the exception clause |
| `repairOutboundPageLinks` | PRESENT | PRESENT | **positive control** |
| `zzzNotARealSymbol396zzz` | absent | absent | **negative control** |

The negative control is what makes the rest mean anything — without it, five PRESENTs are
indistinguishable from a grep that matches everything.

### So `633` was applied, and verified at the artefact

Hold condition met, and **inert on apply**: 0 sites carry a non-NULL `lock_except_item_ids`, so the
new clause evaluates identically to the old for every row. Read back with a query the migration does
not contain: gate **names** the column ✓, gate **KEEPS** `s.locked_at IS NULL` ✓ (the arm whose loss
would switch the lock off), `load_items` `honour_site_lock=true` ✓, and the negative control —
`site-work-orchestrator` steps carrying the key — **0** ✓.

⚠ **`--record-only` REFUSES a `_HOLD` file** ("an UPPERCASE-suffixed sidecar … recording one is
meaningless"), so the ledger row was written by hand, the same way the runner writes one and the
same workaround `610_..._HOLD` used the previous day. **Held migrations are otherwise invisible to
`schema_migrations`, so "was the held half applied?" can only be answered from the live config.**

### ⚠ MISSTEP 4 — I read a live outage backwards, and the discriminator was one word

The r2 council round produced no verdict. I found `ai_endpoint_health` saying the Anthropic
endpoint was **unhealthy**, and `llm_call_log` showing **569 calls since**. I was one step from
writing *"the health row is stale while calls are succeeding"* — the fleet-stopper landmine's exact
shape.

**They were not succeeding.** `success = false` on **every one**. The column was right; I had
counted *calls*, not *successes*.

> **The check: `count(*)` answers "did it happen", not "did it work". For any liveness claim, group
> by the success column — and if the table has one, an ungrouped count is not evidence.**

The truth: last successful Anthropic call **2026-08-25 23:46:29Z**, first failure **23:47:10Z**,
**631 consecutive failures** since — ~8.6 h. Cause already diagnosed by other lanes hours earlier:
**fleet credits exhausted** (`400 credit balance too low`), owner already push-notified,
`bugs_open/243`'s class. **Not re-filed** — I grepped first this time.

That is why the r2 run ended `complete_invalid`: *"no reviewer produced a readable opinion (5
abstained, 12 unreadable) — a council with no opinions cannot decide"*. **The submission was never
judged.** ⚠ I had also half-concluded it was refused for its duplicate file paths — plausible,
adjacent, and wrong. The duplicates are real and worth fixing before resubmitting, but they were
not the cause.

### The queue is stalled, and the number is the tell

**1,399 `triaged`, 0 `claimed`**, oldest triaged **2026-08-18**. That is the documented signature —
`claim_work_item` gates on `ai_endpoint_health`, so a false row releases every claim fleet-wide.
Nothing this lane can do about it; it clears when the credits do.

### What that means for the site-lock work

The config is live and provably inert. **The end-to-end exercise is OWED and cannot be done today**:
lock a site, except one item, confirm that item dispatches and its siblings do not — all of which
needs a working dispatch queue. Recorded in the ledger note and in the handoff, not left implied.

## 2026-08-26 09:0xZ — fleet back, council resubmitted, and the exception list EXERCISED against live data

### The fleet recovered, and the grouped query is what showed it

`GROUP BY success` over the last 30 minutes: failures stop at **08:57:45**, successes start at
**08:58:28**. Health row flipped to `t` at 09:00. Queue moved `claimed = 0 → 2`. **The bare
`count(*)` I used yesterday would have read "healthy" through the whole outage** — see misstep 4.

### ✅ ACCEPTANCE: both gates exercised across three states, on live data, rolled back

The end-to-end scheduler path is not reachable on demand — `find_dispatchable_site` is
`ORDER BY wi.created_at ASC LIMIT 1` **across the whole fleet**, and with **1,398** triaged items
outstanding (oldest 08-18) a site whose items were created this morning will not be selected for
days. Waiting for it is not a test, it is a hope.

So both predicates were run **verbatim against live data** inside a transaction that was then
**rolled back** — `cv1.co.uk`, 6 dispatchable items:

| state | selector picks the site? | loader returns |
|---|---|---|
| **A — unlocked** (baseline) | YES | **6** (all of them) |
| **B — locked, NO exception** | **no** ✓ | **0** ✓ |
| **C — locked, ONE exception** | **YES** ✓ | **exactly 1** ✓ — and **the right one** ✓ |

**6 → 0 → 1 is the whole result.** State B is the important one: it proves the lock still holds
with an exception column present, which is the arm whose loss would silently switch the lock off
fleet-wide. State C proves the exception is scoped to the named id and does not leak to its five
siblings. Negative control after `ROLLBACK`: `locked_at` NULL, `lock_except_item_ids` NULL —
**zero production impact, and cv1's queue was never actually held.**

⚠ **What this does NOT prove, stated so nobody reads it as more:** that the *scheduler* picks the
site in production. The predicates are proven; the dispatch loop's use of them is proven only by
the binary probe and the config read-back. **The honest gap is the tick, not the logic** — and it
closes on its own the first time a locked site with an exception list wins the fleet ordering.

### Council resubmitted

`175df761` r2, run `e74cc1f3`. The JSON was corrected first: it had **8 edits with two file paths
listed twice**, because the r2-only additions were appended for files already in the plan. Merged
to **6 edits, 0 duplicates**. ⚠ That was **not** why the previous round died — it died
`complete_invalid` with every reviewer unreadable on the outage, and was never judged — but a plan
listing one file twice is harder to review and sat at the 8-edit cap for no reason.

## 2026-08-26 (afternoon) — the production proof arrived for free, and the guard I left behind was blind

Picked the lane back up at "nothing blocked, nothing owed". Both of the things below were found by
**re-checking claims this lane had already written down**, not by new work.

### 1. ✅ RESIDUAL CLOSED: the scheduler honours the lock IN PRODUCTION — observed, not inferred

This morning I recorded the honest gap as *"the tick, not the logic"*: the three-state acceptance
(6 → 0 → 1) ran the predicates verbatim but inside a rolled-back transaction, and I wrote that the
scheduler path *"cannot be forced"* because `find_dispatchable_site` is `ORDER BY wi.created_at ASC`
fleet-wide with ~1,400 older items queued. **That was true, and it did not need forcing — the fleet
ordering did it on its own.**

As of 2026-08-26 ~15:00Z the **eight oldest dispatchable rows fleet-wide** all sit on
`adversecreditmortgage.co.uk`, which is **locked**, `locked_by = "portfolio_positioning: owner HALT
2026-08-18 pending classifier register-input (RFC) + builder-flow decision"`. So a locked site heads
the queue and the lock is exercised on **every tick**.

Two arms, both read-only `SELECT`s, no mutation of anything:

| arm | query | returns |
|---|---|---|
| **guard** | the live `find_dispatchable_site` text, **verbatim** | `agritec.uk` |
| **control** | the same text with **only the lock clause deleted** | **`adversecreditmortgage.co.uk`** |

**The two queries differ in exactly one clause, so that clause is what moved the answer.** The
control is the half that matters: without it, "arm 1 did not return the locked site" is equally
consistent with the query simply being broken — the [[a-post-fix-zero-needs-a-demand-control]]
shape, and the same two-sided discipline the `sites.locked_at` entry already demands ("a guard that
never lets anything through is indistinguishable from a broken pipeline").

⚠ **This test is not always available**, and that is worth stating so nobody records a false pass:
if the control returns an *unlocked* site, no locked site currently heads the ordering, the lock is
not being exercised, and **neither arm means anything**. Unavailable, not passed.

**67 dispatchable items across 3 locked sites** are held by this clause today. It is not a latent
guard; it is load-bearing right now, and on an owner HALT.

### 2. ⚠ THE GUARD I NOMINATED FOR MIGRATION AUTHORS IS BLIND — and my own migration is what blinded it

The approving council's one gating-level advisory was that the Go test cannot reach a **migration**
author. I answered it in the handoff by nominating the `sites.locked_at` LANDMINES entry as the
guard. **I never ran that entry's check against the failure it was now supposed to catch.**

Its check is `... ->>'query' LIKE '%locked_at%'`. Four spellings, one `VALUES` list, one query:

| spelling | check says | what it actually does |
|---|---|---|
| **A** the live clause | `HONOURS` | correct — **1,104** rows admitted |
| **B** outer parens dropped | `HONOURS` | `AND` binds tighter than `OR` → status/attempt/retry/deps gates stop applying to every unlocked site. **15,683** rows — re-dispatches `complete`, `failed`, `cancelled` |
| **C** `OR COALESCE(...) IS NOT NULL` | `HONOURS` | `COALESCE` is never NULL → **lock off on every site**, releases the 67 held items |
| **D** exception arm deleted | `HONOURS` | kills `lock_except_item_ids` silently — **and no row count changes today**, because all 3 locked sites have an empty exception list, so data cannot tell you |

**The check was not wrong when written.** On 2026-08-03 the clause was *absent*, and a substring test
detects absence perfectly. **Migration `633` — mine — made presence insufficient**, because the
clause became conditional. I inherited a check across the exact change that invalidated it.

C is the one with teeth: it would dispatch onto the owner-HALTed site on the **next tick**, and the
check every session is told to run would print `HONOURS`.

**Corrected in `LANDMINES.md`** — original left visible per convention, with the four-spelling table,
a two-sided behavioural check, and an always-available `DO` block that executes the **live text** so
it cannot drift from what runs. ⚠ Recorded honestly there: that block catches **C** and **D**, and
**cannot catch B** — B's damage lands on *unlocked* sites, so the site it returns is legitimately
unlocked. Nothing short of reading the parens catches B.

Logged as `WRONG_CALLS.md` 2026-08-26. The generalisation is the transferable part: **when you
nominate an existing check as the guard for a new failure mode, feed the new failure mode to the
check and watch it fail** — the tell is that you changed the thing the check inspects.

## 2026-08-26 (evening) — chassis rolled to v1.0.1345; everything re-proven, and the fixed guard immediately earned itself

### The roll obliged a re-proof, and the provenance line was useless

Chassis `v1.0.1341` → **`v1.0.1345`**, pods up 20:24:56Z / 20:25:20Z. The lane's Go half is in that
binary, so none of this morning's liveness claims transfer.

⚠ **`build provenance` was NOT in range on either pod, ten minutes after start** — chassis emitted
**2.4 MB of logs in ten minutes** and the startup line had already scrolled past `--tail=20000`.
Empty means *not in range*, never *unstamped*. The binary probe has no shelf life; that is the tool.

⚠ **The absent-control timed out at 120 s and I nearly shipped a probe without it.** Scanning a whole
binary for a string that is *not* there is the slowest case — it cannot stop early. The first run
returned four `PRESENT` rows and died before `zzzNotARealSymbol396zzz`. **Four PRESENTs with no
absent-control are unfalsifiable** (a grep matching everything looks identical). Re-ran it alone at
240 s: **absent** ✓. Both replicas then came back clean on all five symbols.

### The config half survived, and I checked the SHAPE this time, not the substring

Live selector names `lock_except_item_ids` ✓ **and carries the exact parenthesised form** ✓. That
second check is the whole lesson of this afternoon — the substring alone is what made the old
landmine check useless. **I wrote the check that way because I had just been caught by the other.**

Ran my own new `DO` block against the live row for real: `PASS: selector returned an UNLOCKED site`.
The two-arm production proof still holds post-roll — a locked site (`adversecreditmortgage.co.uk`,
owner HALT) still heads the ordering. **70 held items across 4 locked sites now** (was 67/3 — another
site was locked in the interim, so this number moves; do not quote it without re-counting).

### The fixed guard paid for itself within the hour — migration 657

Having corrected the guard, I asked the obvious next question: **is any pending migration about to
edit that query?** One is — `657_selector_ranks_sites_by_loadable_work_HOLD.sql` (`bugs_open/413`,
`dispatch_throughput`), council-APPROVED and hand-applied ~12:00Z tomorrow.

**Their query is correct** — lock clause present and properly wrapped, header names `config.query`
not `pre_query`. They had read the landmine. **But their guard has exactly the blind spot I had just
documented**: `657:201-209` tests eligibility fragments with `position()`, and **four of the seven
are OR-bearing and listed WITHOUT their wrapping parens.** Their comment says each clause "widens
dispatch if dropped" — the precedence break widens dispatch *without dropping anything*.

Also confirmed their md5 precondition still holds (`d6f98acdb5aec385d5eb4077eac530fc`), so `657`
applies cleanly — worth telling them, since a refusal would otherwise read as a bug in their file.

CONTRIB written into **their** directory, decision left to them, explicitly not a reason to delay
their apply. This is the owner ruling of 2026-07-29 §3 in practice: *a shared mechanism's other
consumers must be TOLD, not merely measured.*

### One thing I could not attribute, recorded rather than guessed

The selector's `agent_definitions` row shows `updated_at = 20:24:17`, ~40 s before the pods started.
**I did not identify the writer.** Established: the `633` clause is present in its exact shape
afterwards; exactly one active non-snapshot row for this type; md5 matches `657`'s baseline. So the
fix survived it. Marked `[UNRESOLVED]` in the handoff rather than narrated into a cause.

Separately: `sql_for_agents/052_build_pipeline_trigger.sql` — the seed — still carries the **OLD**
query with no `lock_except_item_ids`. A re-seed would silently revert `633`. It is not in
`schema_migrations` and did not fire here. `[UNVERIFIED]` whether any path re-applies seeds; stated
as an open question, not a finding.

## 2026-09-02 — the standing residual is CLOSED IN CODE: migration 690 refuses an untraceable park

Picked this up on the owner's instruction to "fix and test the trigger". A week had passed, so I
re-checked ownership first: nobody had built it, nobody else had touched the lane.

### The census is what designed the guard, and it contradicted my starting assumption

I expected to require `parked_by`/`parked_reason` on every write of `status='deferred'`. **That
would have been a fleet-breaking mistake**, and the write-history census caught it before I wrote a
line of SQL:

| shape | rows | with provenance |
|---|---|---|
| `deferred` + EMPTY handler — the `bugs_closed/077` shelf | **2,656** | **0**, correctly |
| `deferred` + NAMED handler — this bug's shape | **257** | 87, **170 without** |

The shelf class is a *different mechanism* with five live producers, deliberately provenance-free.
The discriminator was already written down in the codebase — `write_audit_findings_action.go:95`
warns against *"the other shape — `deferred` WITH a named handler"* — and this bug's own title says
"with a named handler". **I had read that title many times and still nearly built the wrong guard.**
[[census-the-write-history-not-the-bug-file]] is the lesson, and it earned its keep here.

### Two more things only the source gave me

**`park_work_items()` stamps `result`, NOT `spec`** (621:147-154), while migration `389` stamps
`spec`. A guard reading only `spec` would have **refused the sanctioned verb** — and reading only
`spec` is precisely this lane's §8.1 misstep, the one that called 62 stamped rows "no trace of any
kind". I very nearly repeated my own recorded error, in the fix for the bug that recorded it.

**The two files that name `deferred` beside a handler are READERS, not writers.**
`work_item_failure_ladder.go` has it in a guard list of statuses a write must NOT overwrite;
`work_item_retraction.go` reads it to count parks being drained. Had I grepped and stopped at the
hit, I would have concluded there were live writers and built something far more cautious.

### The dry run found a defect in my TEST, not my guard

First dry run failed on CHECK constraint `swi_no_handlerless_promotable`: an empty `handler_agent`
cannot be in `triaged`/`approved`/`claimed`. My assertion staged a shelf row at `triaged` and
updated it to `deferred` — **a shape production can never produce.** Shelf rows are *born*
deferred. Corrected, and the correction made the test more faithful, not less.

### The mutation proof is the part that makes the clean run mean anything

Three clean dry runs (COMMIT swapped for ROLLBACK): 6 assertions, exit 0, nothing installed. Then
two deliberate breakages, each caught by a **different** assertion:

- guard made inert → assertion 1, *"an untraceable park was ACCEPTED"*, exit 3
- shelf exemption removed → assertion 4, *"the SHELF class was REFUSED"*, exit 3

**The second mutation is the one that matters.** A one-sided "did it refuse?" test passes a guard
that refuses everything — which would have broken 2,656 live rows. Same two-sidedness this lane
learned the hard way on 08-26 with the blind `LIKE '%locked_at%'` check.

### ⚠ NOT APPLIED — and the reason is a permission boundary, not a doubt

The production apply was **blocked by the session's own harness classifier**. I did not work around
it. `schema_migrations` has no `690` row and the trigger is not attached, so **the fix protects
nothing today**. Everything up to the apply is done: committed `a027bf03b`, council SUBMITTED
`dcd2b3c9`, registered **WII-037**, recipe in `396` §6f.

⚠ **And a trap found while looking for the apply path:** `run-migrations.sh --apply` takes EVERY
pending file — **271 of them today** — so using it would sweep ~20 other lanes' migrations into
production. The scoped path is `psql -f` by hand, then `--record-only`. Also noticed `668` is
duplicated on disk right now (two different files, same number): the documented same-number trap,
live.

### Register corrections made in passing

WII-034 and WII-036 both stated the trigger residual as open, and WII-036's status still said the
config half was HELD and the scheduler proof unobtained — both untrue since 08-26. Council seats
read the register as ground truth, so all five stale claims were struck through and corrected in
place. WII-036 was also **missing from the concept index entirely**; added, with WII-037.

### ✅ 2026-09-02 16:16Z — APPLIED by the owner, and proven at the artefact

The owner ran the apply by hand (my session's harness classifier gates live schema changes). It
committed cleanly, with the post-check's own NOTICE in the output: **6 assertions passed before
COMMIT**.

Then the independent `_VERIFY` sidecar against the **live** trigger: **all 6 passed, exit 0**,
ending in `ROLLBACK`. Structural confirmation in the same breath —
`trg_site_work_items_park_provenance` attached and **enabled** (`tgenabled='O'`), sitting beside the
pre-existing `trg_site_work_items_updated_at` with no interference; ledger row recorded
(`applied_by='record-only'`, with the note saying why `--apply` was not used); and **zero litter
rows** (`item_key LIKE 'MIGRATION_690_%'` = 0), so the self-test cleaned up after itself as designed.

⚠ **Council `dcd2b3c9` verdict is still UNREAD.** The commit carries `Council-Submitted:`, which
asserts nothing and lets `098` credit it automatically once the verdict turns approved. **Nobody
should write `Council-Reviewed:` for this until they have actually read the verdict** — `098`
buckets an unread claim as MISMATCH, which is the report's dishonesty surface.

**Worth noting about the apply itself:** the migration is self-proving, so "it applied" and "the
guard works" were established by the same command — the post-check would have aborted the
transaction before `COMMIT` if the refusal had not fired. That is the property that made handing the
command to someone else safe: they did not have to interpret anything, and a broken guard could not
have installed itself quietly.

### ⚠ 2026-09-02, later — the council APPROVED 690 and found a real hole in it, and my own test was defending the hole

Verdict on `dcd2b3c9`: **APPROVED, 4 advisory objections, none high.** I read it rather than
stopping at the decision field, which is the only reason the rest of this entry exists.

**The `editquality` seat was right.** `690`'s third exit exempts every update to an already-deferred
row, including one that changes `handler_agent` — so a shelf row can be re-pointed to a named
handler without `status` ever moving, landing on `deferred` + NAMED + no provenance. I did not take
the seat's word for it: **induced against the live trigger inside a rolled-back transaction —
ACCEPTED.**

**The part worth writing down is that my own `_VERIFY` asserted that exact write as CORRECT.**
Assertion 5 required it to succeed, and I had described it in the file as *"the sharpest form"* of
proving already-deferred rows stay writable. The assertion and the exploit were the same statement.

**And mutation-proving did not catch it.** Both my mutations were applied to the *guard* — inert,
and over-broad — and the test agreed with the guard's blind spot in both. **A mutation test proves
your test can detect a change to the thing you thought of; it cannot tell you the thing you did not
think of is missing.** That took an outside reader with a different frame. This is the concrete
argument for the council gate being worth ~30 minutes, and it is stronger than the abstract one:
the round APPROVED the change and still found a defect that would have shipped.

**Also worth keeping:** the fix's post-check nearly repeated a different mistake. My first draft
wrapped the whole `DO` block in one `EXCEPTION WHEN OTHERS`, which would have **swallowed a failed
assertion and let the synthetic rows COMMIT as litter**. Caught by reading it back before running
it. The final version has no outer handler — every assertion is its own sub-block.

`700` is committed (`1f0cd8ae2`) and **not applied** — same permission boundary as `690`. Until it
is, the hole is open in production, and I have said so at the top of the handoff rather than
letting "690 is live" read as "the guard is complete".

### ✅ 2026-09-02 — migration 700 APPLIED; the hole is closed and the proof is symmetric

The owner applied it. Post-check: 4 assertions before `COMMIT`. Ledger row recorded.

**The confirmation I care about is the symmetric one.** The same induction probe I wrote to test the
council's claim — insert a shelf row, re-point its handler, see whether it is accepted — returned
**"HOLE CONFIRMED: re-point ACCEPTED"** against `690` alone, and now returns **"HOLE CLOSED: the
re-point is now REFUSED"**. Same probe, same rows, opposite answer. That is stronger than any
assertion count, because the probe was written *before* the fix existed and could not have been
shaped to agree with it.

Second confirmation: the corrected `_VERIFY`, which **failed at assertion 5b** before `700` with
*"a shelf row was re-pointed to a NAMED handler with no provenance and ACCEPTED"*, now passes **all
7, exit 0**, `ROLLBACK`, **zero litter rows**. That pre-fix failure was itself the demand control —
5b was proven able to fail before it was allowed to pass.

**Guard is now complete on both entry paths:** the transition INTO `deferred`, and the handler
re-point while already `deferred`. Both `690` and `700` are in `schema_migrations`.

**What I would carry to the next lane from this whole sequence:** the mutation testing I was pleased
with was real but bounded — it proved the test could detect the failures *I had imagined*. The
failure I had not imagined was sitting inside my own test as a passing assertion. Two things caught
it and neither was mine: an outside reader with a different frame, and then a probe written to test
*their* claim rather than my code.

### 2026-09-02, close — a `/code-review` at the end, and what it actually reviewed

Ran `/code-review` after the lane was complete. **15 findings, none of them this lane's.** They named
`save_page_meta_description_action.go`, `voicetells.go`, `registerwords.go`,
`check_unrendered_page_imagery.go` and migration `694` — files this session never opened. Nothing
about `690`, `700`, their sidecars or the docs.

**The cause is structural, not a reviewer fault:** a bare diff review on this tree reads the working
directory, which holds ~10 concurrent sessions' WIP at once. So it reviewed the union of everyone's
half-finished work and reported it to me in the second person. The real hazard is that acting on a
finding means **editing another lane's code mid-flight** — the one thing CLAUDE.md most consistently
forbids. Written up as a LANDMINE, because a session with no symptom would reasonably assume the
findings were theirs.

**One check that stopped me over-reacting:** two findings named files that turn out to be
**untracked**, so the red `TestNoHandSpelledTombstonePredicate` is a working-tree failure belonging
to whoever is writing them — **not a red HEAD**. `git ls-files --error-unmatch` settled it in one
command. I have NOT independently tested HEAD as a whole and said so rather than implying I had.

**Two findings are worth routing** and are recorded in the handoff §4b as not-done: the untracked
`BANNED_REGISTER_v2.json` that `registerwords.go` already references (a pathspec commit of the Go
alone breaks HEAD for everyone), and a deleted false-positive guard case that turned the suite green
while leaving the `plain_words` pattern flagging ordinary prose. Both belong to their lanes; CONTRIBs
not written.

**What I would keep from this:** "run a reviewer over your work" and "run a reviewer over the diff"
are the same command here and different actions. On a single-session repo they coincide; on this one
they do not, and the report gives you no signal that they diverged.
