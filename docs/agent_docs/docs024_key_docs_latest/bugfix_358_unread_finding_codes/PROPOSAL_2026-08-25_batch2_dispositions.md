# Batch 2 — proposed dispositions for the 25 remaining undecided finding codes, and the seven decisions they leave the owner

> ## RULED IN FULL — owner, 2026-08-25 — and APPLIED the same day
>
> **1** = option (a), all 24 → `human-evidence` (commit `835ab0585`; cap 25 → **0**; live check
> clean, exit 0). **2** = retire — `BUILD_DISPATCH_STALLED` deleted from the registry and migration
> `214` marked `_SUPERSEDED` (it turned out to be an untracked orphan, so it is now tracked for the
> first time as part of being retired). **3** = accept 365 days for the retraction audits, recorded
> in their `why` fields as a knowing acceptance. **4** = ALL FOUR readers commissioned —
> `bugs_open/392` (link context, routed at the 092 lane), `bugs_open/393` (dark_section_audit — the
> census found the entire population is ONE item type), `bugs_open/394` (render-audit truncation —
> prior art showed `242` built the signal and `392`-the-migration raised the cap; the ask is the
> consumption half), and the link-repair pair contributed INTO `bugs_open/071`, whose lane is
> active (contribute, don't compete). **5** = rename + prevent — `HANDLER_VERDICT_UNRECOGNISED`
> (commit `eb7d92371`, council `8d798266`), with the new commit-time prefix guard proven by the live
> case itself: the guard went in first and failed on UNKNOWN vs UNKNOWN_HANDLER_VERDICT, then the
> rename cleared it. **6** = wait — the 13 baseline codes are ruled on first fire. **7** = leave —
> one cap counter, now at its terminal 0: every newborn code is a finding unless declared in the
> same commit, which is the intended behaviour.


**For the owner to ratify or overrule** (his ruling, 2026-08-22: the session proposes with the
evidence attached, he ratifies in batches). ~~**Nothing here is applied.**~~ **Superseded same day
by the banner above: everything IS applied or filed.** The line is struck rather than deleted — the
proposal is the record of what was argued, and this is what it was arguing towards. Every row count is
`[MEASURED 2026-08-25]` over the retained window; every reader claim is a grep of the literal AND
its Go constant across every language, plus a regex over live `agent_definitions` and
`scheduled_tasks`. Batch 1 (7 codes) was ratified 2026-08-23; this is the rest.

---

## The one measurement that decides most of this

**No live workflow or scheduled task selects any of the 25** (`agent_definitions.default_config`
and `scheduled_tasks.pre_query/config`, regex over all 25 names → 0 rows). **Only two have any
reader at all, and both are hand-run:**

- `tool_crosslink_not_emitted` — `cmd/backfill-tool-crosslinks/main.go:90` selects one colon-variant
  by name; a backfill binary with no kustomize service and no makefile target. Plus `LIKE` verifies in
  migrations `211_…_VERIFY` and `602_…_HOLD`.
- `CONTENT_DATA_ENVELOPE` — a SQL query **in a comment** (`render_content_envelope_guard.go:273`),
  i.e. instructions for a human.

So the disposition question for 24 of the 25 has the same answer batch 1's seven had: **a human
wrote this deliberately, and nothing automated reads it.** `human-evidence` is that state's honest
name, not a judgement about the code's worth. The interesting content — which of these *deserve* a
reader, and one code that has no writer at all — is in the decisions section, not the table.

## The proposals, by family

Six of the seven families share a disposition because they share a measurement. Each `why` below
names the retention window it accepts (the checker requires it).

### A. `component_validation_*` siblings — the batch-1 `reader_sink` shape, twice more

| code | rows | what the writer records | proposed |
|---|---|---|---|
| `component_validation_orphan_schema_field` | 79 | a generated component declaring a schema field its template never renders | **human-evidence** |
| `component_validation_unknown_template_var` | 0 | a template referencing a variable the schema does not declare | **human-evidence** |

Both are consumed — **from another sink.** Migration `563` branches the component-creator's prompt
on `last_error_code` for all three `component_validation_*` codes (`grep` of 563/564: each ×4),
reading `site_work_items.retry_feedback`, not this table. Exactly `component_validation_rejected`'s
case, ruled `human-evidence` in batch 1 for this reason. Ruling the siblings the same way keeps the
family consistent, and the `component_validation_%` `LIKE` in the docs keeps working.

### B. Tool birth and regeneration — refusal records

| code | rows | what the writer records | proposed |
|---|---|---|---|
| `tool_crosslink_not_emitted` (+ colon variants) | 50 | cross-links withheld for a tool page, with the reason as a suffix (`:tool_page_will_not_go_live` …) | **human-evidence**, naming the backfill reader |
| `tool_birth_instance_scope_refused` | 9 | a tool birth refused because the generated HTML is not instance-scoped | **human-evidence** |
| `tool_regeneration_hollow_blocked` | 3 | a regeneration refused as hollow — *"the incumbent stands untouched (bugs_open/331; the 012/056 hollow-overwrite class)"* | **human-evidence** |
| `tool_birth_truncation_blocked` | 1 | a birth refused because *"generated HTML is structurally incomplete"* | **human-evidence** |

These are the record that a guard **held**. Their value is countability — "how often does the
hollow-overwrite class try to happen" — which is a hand-run question today. ⚠ `602_HOLD` (another
lane, in progress) reads the `tool_crosslink_not_emitted:%` family; if that lane ships an
automated consumer, the disposition upgrades to `consumed` with `reader_sink: agent_error_log`.

### C. Retraction audits — written *to be* durable, on a 365-day clock

| code | rows | what the writer records | proposed |
|---|---|---|---|
| `ASSET_RETRACTION_AUDIT` | 13 | one info row per asset-retraction run | **human-evidence** |
| `RETRACTION_AUDIT` | 11 | one row per page retraction: what it *"considered, refused and stripped"* | **human-evidence** |
| `RETRACTION_REFUSED` | 5 | a page retraction refused, with the reason | **human-evidence** |
| `ASSET_RETRACTION_REFUSED` | 0 | one warning row per refused asset deletion | **human-evidence** |

The writer's own comment says why these exist: *"after the first live batch run proved the
collected_data copy dies at the await park, this row is the ONLY durable record of what a
retraction considered, refused and stripped."* **Durable — and under migration `567` it lives
365 days.** That is the one place in this batch where accepting the window is a real choice rather
than bookkeeping; see decision 3.

### D. Component collision — the RFC_034 programme's instrumentation

| code | rows | what the writer records | proposed |
|---|---|---|---|
| `COMPONENT_COLLISION_DIVERTED` | 11 | *"function %q is held by component %s, depended on by %s — write diverted to new base component"* | **human-evidence** |
| `COMPONENT_COLLISION_DIVERT_BLOCKED` | 0 | the divert itself refused | **human-evidence** |

Written by the shared-scope guard in `store_generated_component_action.go` (the INSERT path that
deliberately bypasses the seam). Nothing counts diverts today.

### E. Content and validation detail rows — one design, ruled together

| code | rows | what the writer records | proposed |
|---|---|---|---|
| `CONTENT_DATA_LINK_AUDIT` | 49 | *"a THIRD code beside CONTENT_LINK_REPAIR_DETAIL and CONTENT_LINK_REPAIR_SKIPPED"* — link audit on the content-data path | **human-evidence** |
| `CONTENT_DATA_ENVELOPE` | 7 | a refused save, *"so refusals are countable in SQL rather than only visible in pod logs"* | **human-evidence** |
| `CONTENT_DUPLICATE_SECTIONS_COLLAPSED` | 1 | duplicate sections collapsed on save, with the ORDER pattern that discriminates two producers | **human-evidence** |
| `CONTENT_VALIDATION_WARNING_DETAIL` | 1 (first row **today**) | *"what did the gate SEE and leave in place on a build that succeeded"* — sibling of the consumed `BLOCKER_DETAIL` | **human-evidence** |
| `CONTENT_CREATOR_CLAIMS_DETAIL` | 0 | the content-creator's claims guard — the `pgxpool` writer that cannot use the seam | **human-evidence** |

`CONTENT_DATA_LINK_AUDIT` is one design with `CONTENT_LINK_REPAIR_DETAIL` (batch 1, human-evidence,
flagged "most worth a reader"); the same flag applies — see decision 4.

### F. Plan-shape records

| code | rows | what the writer records | proposed |
|---|---|---|---|
| `FIX_PLAN_VALIDATION_REFUSED` | 3 (all 07-31) | a council fix-plan refused, *"findable by someone who does NOT already know to look in diagnosis_artifacts"* | **human-evidence** |
| `PLAN_PAGE_MERGE_LOSSY` | 2 | two plan pages canonicalised to one and authored sections discarded — *"the only artefact that can answer 'did richer-wins ever actually drop content'"* | **human-evidence** |
| `PLAN_PAGE_SAME_NAME_IDENTITY_HELD` | 1 | stored page identities kept over the canonicaliser's re-derivation | **human-evidence** |

### G. Singletons

| code | rows | what the writer records | proposed |
|---|---|---|---|
| `NO_CHANGE_GATE_UNREADABLE_RESULT` | 11 | *"no-change gate could not read the handler's result … item completed UNGRADED by this gate"* | **human-evidence** — ⚠ see decision 4 |
| `RENDER_AUDIT_TRUNCATED` | 4 | *"render audit truncated by max_pages: %d of %d live pages audited — the unaudited tail is the SAME pages every run"* | **human-evidence** — ⚠ see decision 4 |
| `component_write_shared_blocked` | 1 | a write to a shared-scope component refused | **human-evidence** |
| `REVIEW_SUPERSEDED_BY_PASSING_SAVE` | 0 (25 rows deleted by retention 08-22) | the reconciler's OUTPUT — a review superseded by a later passing save | **human-evidence** |
| `BUILD_DISPATCH_STALLED` | 0, **no writer live** | its only writer is migration `214`, never applied | **HELD `unruled`** — decision 2 |

---

## The decisions, in the order they matter

For each: what the thing is, the rule that governs it, and how this case measures against it.

### 1. Ratify batch 2 — 24 codes to `human-evidence`; the cap goes 25 → 1

**What it is.** A disposition is the estate's recorded decision about who reads a finding. `human-evidence`
means: nobody automated does, a human may, and the retention window is accepted knowingly.
**The rule.** Yours, 2026-08-22: the session proposes with evidence; you ratify in batches; the
undecided backlog is capped and the cap comes down with each batch.
**This case.** Twenty-four codes measure identically — no automated reader anywhere — so the
disposition is a measurement, not a taste. One commit applies them and lowers `_unruled_cap` to 1.
**Your sub-choice:** any code you would rather HOLD as `unruled` because you intend to commission
a reader for it (decision 4) — holding is honest; declaring then re-ruling later is also fine.

### 2. `BUILD_DISPATCH_STALLED` — apply migration 214, or retire the code

**What it is.** `214_build_dispatch_watchdog.sql` is a scheduled task that would raise this finding
when a build dispatch stalls; it also reads its own rows to avoid re-raising. It has **never been
applied** (0 rows in `schema_migrations`, no `build-dispatch-watchdog` task), so the code has no
writer in the live estate.
**The rule.** A registered-but-never-observed code is report-only (retention is 30 days, absence
proves nothing) — but this one is not *quiet*, it is *absent*, and the registry should not carry a
code whose mechanism does not exist as if it were merely resting.
**This case.** Migration `506` (applied) says in its own header that honouring `retry_after` leaves
*"no false BUILD_DISPATCH_STALLED risk to mitigate either"* — i.e. the premise 214 guarded may be
moot. **Options:** (a) apply 214 as intended; (b) retire it — mark 214 `_SUPERSEDED`, drop the
registry entry, and the cap goes to 0. **Recommendation: (b)**, unless someone wants the watchdog;
506's note is the evidence that nobody has missed it.

### 3. The retraction audits — accept 365 days, or give them a home that does not expire

**What it is.** Four codes whose writer's stated purpose is to be *"the ONLY durable record"* of
what a retraction did. Under migration `567` a `human-evidence` row lives 365 days.
**The rule.** A `human-evidence` `why` must name the window it accepts; the checker enforces it.
**This case.** For every other code in this batch the window is fine. For an *audit*, 365 days is a
choice with a consequence: after a year, "what did we retract from that site and why" has no
answer. **Options:** (a) accept 365 — declare, and note it; (b) commission a proper audit table (an
architecture-shaped change, not this lane's). **Recommendation: (a) now**, with the note; (b) only
if retractions become something you are asked to account for.

### 4. Which codes get a REAL reader — commissioning work, not ruling

**What it is.** A reader is a build task for some lane: something automated that selects by the
code and *acts*. This registry cannot conjure one; it can only say honestly that none exists.
**The rule.** Batch 1's: rule today's truth, flag the ambition. A disposition can be upgraded to
`consumed` the day a reader ships.
**This case — four candidates, in the order I would take them:**
1. **`LINK_CONTEXT_UNAVAILABLE`** (declared 08-24, `092` lane's code) — two pages served with no
   internal links after a DB timeout, recorded, read by nothing. A real degradation.
2. **`NO_CHANGE_GATE_UNREADABLE_RESULT`** (11 rows) — a gate that could not grade what it was given.
   Rows > 0 means a handler's result shape drifted; that is a drift detector with no alarm.
3. **`RENDER_AUDIT_TRUNCATED`** (4 rows) — the writer says it plainly: a capped sweep reporting
   clean is a false green, and the unaudited tail is the same pages every run.
4. **`CONTENT_LINK_REPAIR_DETAIL` + `CONTENT_DATA_LINK_AUDIT`** — the longitudinal "did a repair path
   stop acting" question, three origins to count.
**Your decision:** whether to commission any, and to which lane. **Recommendation:** file 1 at the
`092` lane now (I can write the bug); hold 2 and 3 `unruled` if you want them built, else declare.

### 5. `UNKNOWN_HANDLER_VERDICT` — rename the code, or narrow the prefix rule

**What it is.** The checker fails on any declared code that is a prefix of another, because live
`LIKE 'family%'` queries exist and a shared prefix silently merges populations.
**The rule.** Unconditional, pairwise, over every declared code — deliberately.
**This case.** `UNKNOWN` (operational, 18k rows) is a prefix of `UNKNOWN_HANDLER_VERDICT` (in
`_scan_baseline`, zero rows). No `LIKE 'UNKNOWN%'` exists, so it is a rule artefact today — but the
day that code is declared, the daily check goes red and its author will not know why.
**Options:** (a) rename the code before declaring it (its writer is
`complete_work_item_verification.go:394`; a local change); (b) narrow the rule to family prefixes —
a change to what the registry guarantees. **Recommendation: (a).**

### 6. The 13 `_scan_baseline` codes — batch 3 now, or when one fires

**What it is.** Codes the actions package writes that are declared nowhere; all have zero rows.
The scan ratchets against this list so it is not red on day one; the list may only shrink.
**The rule.** Same as batch 2: each needs a ruling; a ruling deletes its line.
**This case.** All thirteen are zero-row, so the live check is blind to them by construction and
nothing is being lost unread *today*. **Options:** (a) a batch-3 proposal now, same evidence
standard; (b) rule each the day it first fires — the CronJob will say when. **Recommendation:
(b)**, because ruling a code that has never fired is ruling on the writer's intent rather than on
what it produces, and every one of these is one commit away from being deleted before it ever does.

### 7. The cap — one counter, or two

**What it is.** `_unruled_cap` conflates an INHERITED backlog (should shrink) and NEWBORN codes
arriving from other lanes (must be ruled promptly). At exactly-at-cap, a newborn is an immediate
breach and the only compliant moves are "rule it now" or "raise the cap".
**The rule.** Yours: cap it.
**This case.** After batch 2 the cap is 1 (or 0 after decision 2), so every newborn *is* a breach —
and that is arguably the right behaviour: declare in the same commit, or the check is red. The
day-one events show it works: two newborns, both declared within hours. **Recommendation: leave it
as one counter**; revisit only if newborns start arriving faster than they can be ruled.

---

## If you ratify decision 1

One commit changes 24 `disposition` fields, adds their `why` (each naming the window), lowers
`_unruled_cap` to 1, and the next fleet release carries it into the image. Decisions 2–7 are
independent of it and can land later.
