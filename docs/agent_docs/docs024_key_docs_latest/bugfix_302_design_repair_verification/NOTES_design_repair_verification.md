# NOTES — `bugs_open/302` design-repair completion verification

Append-only, newest at the bottom. Evidence, commands, and every misstep.

---

## 2026-08-18 — session opens: scoping, ownership, and what the filing got wrong

### Ownership checks, before touching anything

- `scripts/who-owns.py 302` → OWNED/recently active, `finetuning_uk_service`. **Checked the live
  transcript rather than trusting the commit log** (per the landmine "an ownership check that
  reads COMMITS cannot see the session that is fixing it right now"): that session filed 302,
  appended its SCOPING UPDATE (`b6f869676`), and wrapped up with a docs summary. Its current
  work is GPU/training Phase 0, not this bug. **Not competing — it is the FILING lane, not a
  fixing lane.**
- `scripts/who-owns.py 201` → `bugfix_201_page_content_writer_dispatch`, last commit 2026-08-09.
  Its `HANDOFF_2026-08-09` says in terms: *"there is nothing left to do on 201 itself"*.
- Grepped all 40+ live `.jsonl` session transcripts for `bugs_open/302` and `bugs_open/201`.
  One session had substantive hits (the filing lane, above); every other hit was an incidental
  `ls bugs_open/` listing.
- **The seam's own lanes:** `bugfix_213_verifier_producer_join` built gate 1b and is FINISHED
  (its handoff: "There is no queued work… Do not pick this up here", and it explicitly hands the
  `NO_CHANGE_GATE_UNREADABLE_RESULT` stream to the `RFC_029` lane as their before/after).
  `work_item_completion_integrity` is the historical registry lane, dormant since 08-08.

### The machinery, read first-hand

Two completion gates, both consulted from `verifyBeforeComplete`
(`complete_work_item_verification.go`), called at `load_work_item_actions.go:982`:

| gate | file | keyed on | opt-in roster | on "cannot read/run" |
|---|---|---|---|---|
| 1b no-change | `complete_work_item_no_change.go` | handler's returned payload counters | `noChangeGates` — **1 type**: `dark_section_audit` | **abstain and COMPLETE** (+ `agent_error_log` row) |
| 2 verifier | `discovery_checks/verifiers.go` | `item_type` | 11 registered types, all discovery shapes | **fail CLOSED** since RFC_017 (owner ruling 08-08) |

The two sibling gates hold **opposite policies on the same question**. Gate 2's own comment
refuses to exempt an unparseable spec because doing so "would leave a second silent completion
path behind the one RFC_017 closed" — and gate 1b, written five days AFTER that ruling, is
exactly such a path for its own unreadable case.

### [MEASURED 2026-08-18] the arm is live and it fired — with a demand control

```sql
SELECT error_code, count(*), min(occurred_at), max(occurred_at)
FROM agent_error_log WHERE error_code='NO_CHANGE_GATE_UNREADABLE_RESULT' GROUP BY 1;
--  11 rows, 2026-08-14 14:24Z → 2026-08-17 12:44Z
```

Not a theoretical arm, and **not a broken gate**: in the same window gate 1b correctly BLOCKED
4 rows (`status='failed'`, `_verification.status='handler_reported_no_change'`). The gate works
when it can read the payload. Both directions observed, so the measurement could have come out
otherwise.

### [MEASURED] the filing's account of the 11 payloads is WRONG, and this matters

302 attributes the population to the handlers returning analysis blobs. The actual shapes:

| result top-level keys | rows | what it is |
|---|---|---|
| `agent_id,agent_type,role,topics` | **7** | a **SPAWN RECORD** — `bugs_closed/287`'s defect |
| `color_scheme,design_notes,spacing,typography` | 3 | the design-token blob 302 names |
| `add_to_page,approach,new_page,not_actionable,reasoning,retype_existing,update_spec` | 1 | an unrelated child-page triage decision |

`bugs_closed/287` is **fixed, live and proven** on chassis `v1.0.1307` (roll 2026-08-17 17:05Z;
fleet now `v1.0.1309`): its own close-out measures `field=result` resolver rows at **0**, down
from ~455/day, with 11/11 loop completions carrying the handler's reply. So the majority cause
of unreadable payloads was removed at source yesterday.

### [MEASURED] and therefore: the arm has had ZERO demand since that roll

```sql
SELECT count(*) FROM site_work_items
WHERE item_type='dark_section_audit' AND updated_at > '2026-08-17 17:05:00+00';   -- 0
SELECT count(*) FROM site_work_items
WHERE status='complete' AND updated_at > '2026-08-17 17:05:00+00';                -- 1862
```

The fleet is busy; this item type specifically is not. **All 11 abstentions predate the roll.**
So the honest position, and the one the fix must be argued on:

- the **structural** hole is real and untouched by 287 — an opted-in type whose payload cannot be
  read is silently exempted from its own opt-in, by construction, for ever;
- the **observed rate** post-roll is **unmeasured, with zero demand**, and I must not claim a
  continuing flood. 7 of 11 are attributable to a bug that is now closed;
- whether the 3 blob rows and the 1 triage row share 287's cause (the recursive `$.**` search
  binding the wrong value into `result`, `RFC_029`'s subject) is **[INFERRED], not established** —
  `bugs_closed/213` §D records it as NOT ESTABLISHED and `RFC_029` owns it. Do not assert it.

### [MEASURED] 302's own working candidate is trap-laden — the producer-split check refutes it

302's SCOPING UPDATE names "Gate-2 artefact verifiers for the design-repair family" as the
working candidate. `LANDMINES.md` mandates a producer split before registering any verifier
(`spec->>'audit_source'` is the only thing that names a producer; `created_by` bottoms out at
`generic`):

```sql
SELECT item_type, count(DISTINCT COALESCE(spec->>'audit_source','<none>')) AS producers,
       string_agg(DISTINCT COALESCE(spec->>'audit_source','<none>'), ' | ') AS which, count(*) n
FROM site_work_items WHERE item_type IN (...) GROUP BY 1 ORDER BY producers DESC;
```

| item_type | producers | which | rows |
|---|---|---|---|
| `needs_design_review` | **4** | brief-fidelity-audit, design-audit, visual-design-audit, `<none>` | 75 |
| `responsive_fix` | 3 | design-audit, visual-design-audit, tool-acceptance-tier4 | 19 |
| `dark_section_audit` | 2 | design-audit, visual-design-audit | 30 |
| `spacing_fix` | 2 | design-audit, visual-design-audit | 30 |
| `contrast_failure` | 1 | `<none>` | 284 |
| `hardcoded_section_colors` | 1 | `<none>` | 9 |

One verifier per `item_type` over a 4-producer population **is** `bugs_closed/213`'s defect: the
verifier is correct about the wrong question, returns `Resolved:true`, and the item closes
untouched. So any gate-2 route for this family owes a `VerifierPolicy.Grades` remit function per
type (contract WII-013) — a materially bigger job than the filing implies, and not the cheapest
thing that closes the door.

### `bugs_open/201` — checked for validity, and it holds up

Not a fix task: both symptoms were fixed, live and proven before 08-08, and RFC_017 (the flip
this lane surfaced) is decided, built and proven. Re-verified today at the artefact rather than
from the lane's own account:

```sql
SELECT status, result->'_verification'->>'status', count(*), max(updated_at)
FROM site_work_items WHERE item_type='literal_markdown' GROUP BY 1,2;
```

- **15** rows `failed` + `_verification.status='defect_persists'` — the verifier REFUSING a
  completion, most recent 2026-08-17;
- **1** row `complete` + `verified` — the verifier CERTIFYING a real repair (08-15);
- both directions present, so this is a discriminating check, not a one-sided pass.

Two `complete` rows carry no `_verification` at all (08-17 13:20Z). **Chased rather than
assumed** — they closed through the discovery-check retraction seam (WII-009), not a handler:
`result.resolved_by='literal_markdown'`, `reason='literal_markdown re-scan: page's unlocked
components carry no markdown syntax on either surface'`. A retraction is the detector's own
measurement, so it legitimately does not run the completion gate. **No residual hole.**

---

## 2026-08-18 (cont.) — THREE "I could not read it" arms, and the boundary that stops the fix over-reaching

Read the whole of `complete_work_item_verification.go` + `complete_work_item_no_change.go` rather
than the two functions 302 quotes. The completion path has **three** arms that cannot read what
they were asked to judge, and they do not all behave the same way:

| # | arm | trigger | outcome today | recorded as |
|---|---|---|---|---|
| 1 | `handlerReportedFailure` | `response.status` is a value the guard does not recognise | **completes** | `recordUnknownVerdict` |
| 2 | `handlerReportedNoChange` (gate 1b) | none of the declared counters resolve | **completes** | `recordUnknownNoChangeShape` |
| 3 | gate 2 registered verifier | verifier errors, or spec unparseable | **fails CLOSED** (RFC_017) | `_verification.status='error'` |

**Arm 1 must NOT be changed, and this is the load-bearing distinction.** Its own header records the
measurement: on the 2026-07-18 sweep `failed` was the only value `response.status` had ever held,
and **2,905 completed items carried no `response.status` at all**. An unrecognised status there
genuinely is not evidence of failure, and inverting it would block essentially every completion on
the fleet. It abstains correctly.

**Arm 2 is different in one specific way, and that difference is the whole licence for the fix.**
Gate 1b's roster is **opt-in with a per-type assertion attached** — an entry means its author has
asserted, with a measurement in the `Why` field, that *for this type a zero-change run cannot be a
repair*. When the payload cannot be read, the type is silently exempted from **an assertion it
made about itself**. Arm 1 carries no such assertion; arm 3's fail-closed rule was ruled on
precisely because "I could not check" was being read as "I checked and it is fixed".

So the defensible scope of any fix here is **arm 2 only, and only for types on the opt-in roster**
— not a fleet-wide policy, not arm 1, and not "unregistered verifier types fail closed" (302's
candidate 1, which revisits a recorded ruling and needs the RFC route by its own scoping).

**What refusing actually costs — measured, and it is NOT an item stranded for ever.** A blocked
completion goes to `failUnverifiedCompletion` → `attempt_count+1` → `triaged` for retry →
`failed` at `max_attempts`. For `dark_section_audit` specifically the escape hatch already exists
and is live: `silenceRetractionGates` (WII-018) closes the site's rows after **3 consecutive audit
silences**, and its status filter excludes only `triaged`/`claimed` — a `failed` row can accrue
silence and be retracted. So refusal for this type costs at most three rebuilds and then hands the
row to a retraction path built for exactly this population. That is the same cost the owner
knowingly accepted for RFC_017 (`bugs_closed/201` lane recorded the first live instance of it).

`blockedCompletionReason` already discriminates four blocking causes, each with its own message and
reason code (`verification_unavailable`, `verifier_scope_mismatch`, `handler_reported_no_change`,
`verification_failed`). A fifth cause would follow an established pattern rather than invent one —
and per that function's own comment, the message a human reads off `site_work_items.error` must say
which cause it was, so reusing `handler_reported_no_change`'s text for an unreadable payload would
be recording a finding the gate never made.

---

## 2026-08-18 (cont.) — the implementation surface, and the forcing function ALREADY EXISTS

Baseline pinned before touching anything: HEAD `fe83e9430`,
`go test ./platform/orchestration/actions/ -run 'NoChange|Verification|Verify' -count=1` → **ok**.
⚠ Baseline noted with a sha because committing kills a `git show HEAD:` baseline (LANDMINES).

⚠ **Concurrency check, and it moved under me mid-command.** `git status` showed
`discovery_checks/verifier_coverage_test.go` and `check_misdirected_cta.go` MODIFIED — the
strongest ownership signal there is, in my exact package. By the time I diffed them the diff was
empty: another session had committed (`53a8d3c1d`, `bugs_open/248` — CTA exclusion sets). **Not a
collision**, different concern. My two target files were and are clean. Re-run `git status` before
each edit; the session-start snapshot is worthless here.

Everything this change needs already has a home and a convention to follow:

| what | where | note |
|---|---|---|
| the per-type declaration | `noChangeRule` struct, `complete_work_item_no_change.go` | sits beside `Why` / `CounterPaths` |
| the roster forcing function | **`TestNoChangeGatesRosterCarriesItsEvidence` already exists** | it already fails an entry with no `Why` and no `CounterPaths` — a third assertion extends it; no new mechanism |
| the fifth blocking cause | `blockedCompletionReason`, `complete_work_item_verification.go` | four causes already discriminated, each with its own sentence + reason code |
| its test | **`TestBlockedCompletionReasonDistinguishesNoChange` already exists** | it asserts the causes are DISTINCT via a `seen` map, so a reused message fails |
| threading | `verifyBeforeComplete` → caller at `load_work_item_actions.go:982` | the abstain path (`recordUnknownNoChangeShape`) stays for types that keep today's behaviour |

**This matters for the scope argument:** the change adds a field to an opt-in roster and a fifth
arm to a function that already has four, and both guards that would enforce it are already written
and passing. It is not a new seam; it is a declaration made explicit on a seam that already exists.

---

## 2026-08-18 (cont.) — **CORRECTION to my own claim two sections up: the escape hatch is BUILT but NOT EXERCISED**

> **CORRECTED 2026-08-18, same session.** Two sections above I wrote that refusing a completion for
> `dark_section_audit` "costs at most three rebuilds and then hands the row to a retraction path
> built for exactly this population", and called that path **live**. **The mechanism is deployed;
> it is not running.** What caught it: I went to check whether the retraction had ever written a
> streak, instead of reasoning from the code being merged.

```sql
SELECT count(*) FROM site_work_items
WHERE item_type='dark_section_audit' AND result ? 'retraction';        -- 0 rows, ever
```

```sql
SELECT name, enabled, interval_seconds, target_agent_type, last_triggered_at
FROM scheduled_tasks WHERE name ILIKE '%design%' OR name ILIKE '%audit%';
--  site-discovery-rotation-design   | f | 3600 | design-discovery-agent | 2026-08-11 12:42Z
--  site-render-audit-rotation       | t | 3600 | render-audit-agent     | 2026-08-18 16:22Z
```

**The design audit's own carrier has been `enabled=false` since 2026-08-11.** The enabled hourly
rotation is `render-audit-agent`, which is WII-016's contrast path — a *different* producer. So
WII-018's silence streaks cannot accrue for this type, and none ever have. The `bugs_closed/213`
lane's own handoff says the same thing in terms — *"LIVE IS NOT EXERCISED… both carriers still
`enabled=false`"* — and I had read that file without carrying the fact into my own reasoning.

⚠ **Note the second-order trap in this, which is why "disabled" is not the end of the answer:**
items were still FILED on 08-14 (7), 08-15 (5) and 08-17 (2) with the rotation off, so the audit is
reaching sites by some other route (one-shot tasks or direct dispatch). **A disabled scheduler is
not a dead subsystem — count the rows.** So the type has live traffic AND a dead retraction path
simultaneously.

**What this does to the fix's cost, stated honestly.** Refusing an unreadable completion today
means: `attempt_count+1` → `triaged` → retry → `failed` at `max_attempts`, and then the row **stays
`failed`** because nothing is accruing the silence that would retract it. It does not vanish —
`failed` is a human-review destination, and that is the destination `bugs_closed/201`'s lane
recorded as *"the correct destination"* when RFC_017's fail-closed branch first fired. But it is
**three wasted rebuilds and a row a human must look at**, not "the retraction tidies it up". That
is the real price, and it is the same price the owner knowingly accepted for RFC_017 — which is the
honest way to argue for it, and the way it goes into the submission.

This also means **the fix must not be argued on "the retraction will catch it"**, and any plan that
leans on that (including anything fable returns that does) needs this correction applied first.

---

## 2026-08-18 (cont.) — the payload class 302 blames is GONE post-roll, and the release valve is PROVEN once

### [MEASURED] 287's fix removed the malformed-payload class entirely, with a real demand control

```sql
SELECT CASE WHEN updated_at > '2026-08-17 17:05:00+00' THEN 'after roll' ELSE 'before' END AS era,
       count(*) AS completions,
       count(*) FILTER (WHERE result ? 'response') AS has_envelope,
       count(*) FILTER (WHERE result ?& array['agent_id','agent_type','role','topics']) AS spawn_shape
FROM site_work_items WHERE status='complete' AND updated_at > '2026-08-14' GROUP BY 1;
```

| era | completions | has handler envelope | **spawn-record shape** |
|---|---|---|---|
| before (08-14 → roll) | 2,694 | 1,505 | **939** |
| **after roll** | **1,880** | 1,813 | **0** |

939 → **0** against 1,880 completions of demand. ⚠ The 939 independently reproduces the `287`
lane's own figure for rows that stay wrong for ever — two different queries, same number, which is
the sort of agreement worth noting because it could easily have disagreed.

### [MEASURED] and the 67 post-roll completions with no envelope are all LEGITIMATE non-handler closes

Not one is a malformed handler reply:

| keys | rows | what it is |
|---|---|---|
| `reason,resolved_at,resolved_by` | 47 | the **retraction** seam (WII-009/016) — the detector's own re-scan |
| `revalidation` | ~10 | the revalidation sweep |
| `applied,closed_by,gate,owner_decision,verified_at_served_page` | 4 | owner-decision closures |
| `completed_at_iso,completed_by_*` | 2 | orchestration bookkeeping |

**So the population 302 blames on the handlers has no post-roll instance at all.** The hole in gate
1b is now a *latent* one: real by construction, with no current traffic exercising it. That is the
honest basis for the fix and it must be argued that way — as a guard that closes a door, not as a
leak being stemmed.

### The safety question this raises, and the answer — the close paths are DISJOINT from the gate

If a legitimate non-handler closure went *through* gate 1b, an unreadable payload would be
indistinguishable from a retraction, and making the gate refuse would block real closures. It does
not:

```sql
SELECT count(*) FILTER (WHERE result ? 'resolved_by' AND result ? '_verification') AS both,
       count(*) FILTER (WHERE result ? 'resolved_by')   AS retraction_closes,
       count(*) FILTER (WHERE result ? 'revalidation' AND result ? '_verification') AS both_reval,
       count(*) FILTER (WHERE result ? 'revalidation')  AS reval_closes
FROM site_work_items;
-- both = 1 | retraction_closes = 63 | both_reval = 0 | reval_closes = 811
```

**811 revalidation closes, zero of which ever carried a `_verification` payload**, and 62 of 63
retraction closes likewise. These paths write the row directly; they do not call
`complete_work_item`, so `verifyBeforeComplete` never sees them and no change to it can block them.

### The single overlapping row is not a counterexample — it is the release valve WORKING

`8ab3a32b…`, `empty_section`, created 08-09, closed 08-14:
`_verification.status='error'` **and** `resolved_by='empty_sections'`.

Read in order, that row is the whole design functioning: the gate **refused** the completion (a
verifier error, fail-closed under RFC_017), the item went into attempts, and days later the
detector's own re-scan found the defect gone and **retracted** it to `complete`. The
`_verification` payload is the earlier refusal, still on the row.

> **This REFINES my correction two sections up rather than reversing it.** The
> refusal→attempts→retraction release valve is **proven live, once, on `empty_section`.** What is
> dead is specifically the **design** audit's carrier (`site-discovery-rotation-design`,
> `enabled=false` since 08-11), so for `dark_section_audit` the valve cannot fire. The architecture
> is sound; the gap is **operational**, in a different subsystem, and plausibly
> `bugs_open/230`'s rotation work rather than this lane's. Both statements are now measured: the
> pattern works, and it is not switched on for this producer.

---

## 2026-08-18 (cont.) — the registry is BIGGER than 302 says, and the archive re-dates two recorded figures

### 302's "eleven item types" is wrong: there are THIRTEEN

The filing (and my own first grep) counted `RegisterVerifier(` and missed
**`RegisterVerifierWithPolicy(`**, which is how the two most interesting ones are registered:

```bash
grep -rn "RegisterVerifier(\|RegisterVerifierWithPolicy(" platform/ --include=*.go \
  | grep -v _test.go | grep -v 'func Register'
```

The 11 named in 302, **plus `hardcoded_section_colors`** (with a `Grades` remit test) **and
`needs_brand_head_assets`**. ⚠ This matters beyond arithmetic: `hardcoded_section_colors` is a
*design* item type WITH a verifier, which weakens 302's framing that "no design-repair item type is
registered". The accurate statement is that the design-AUDIT family (`dark_section_audit`,
`needs_design_review`, `spacing_fix`, `responsive_fix`) has none, while the design-DISCOVERY
aggregate does.

### I nearly filed a duplicate landmine — the archive trap is already recorded, TODAY, by another lane

`site_work_items` is a **~7-day window**, not a history: `work-item-archiver` (enabled, daily)
moves terminal rows older than 7 days to `site_work_items_archive`. **[MEASURED 2026-08-18] 10,689
live vs 20,184 archived** — two thirds of the history is in the other table. Already in
`LANDMINES.md` (`site_work_items` is a ~7-DAY WINDOW…, added today by the migration-465 lane,
measured at 8,702 live in the morning). **Grepped before filing; cited instead of duplicating.**

**My demand control survives this**, and I checked rather than assuming: the archive's newest row is
**2026-08-11**, well before the 08-17 roll, and **zero `dark_section_audit` rows are archived** — so
the 30-row population and the post-roll zero both stand as complete.

### Applying it: two recorded figures about WII-013's remit test are 7-day figures

`gradesHardcodedColourAggregate`'s own note says *"of the 21 rows ever filed under this item_type,
all 10 from this check carry spec.check and all 11 from the design audit do not"*. Archive-inclusive
the population is **564**, 27× that. Re-run over all of it:

| source | producer | carries `spec.check` | rows |
|---|---|---|---|
| archive | `design-audit` | no | 519 |
| archive | `<none>` (the check) | yes | 36 |
| live | `<none>` (the check) | yes | 9 |

**The partition is CLEAN at 27× the sample its licence was measured on** — 519 correctly disclaimed,
45 correctly graded, no crossover in either direction. That is a *confirmation*, and a stronger one
than the mechanism claims for itself. WII-013 is in better shape than its own comment says.

### And the exposure figure, split by era — I was about to overstate this by 30×

My first instinct was "468 rows closed under a verifier that did not describe them, not 11". That
would have been wrong, because most of them closed before the verifier existed at all. Split at the
two real boundaries (verifier registered 2026-07-24, `Grades` added 2026-08-10):

| era | complete | failed | wont_fix |
|---|---|---|---|
| before the verifier existed (→ 07-20) | 453 | 47 | 4 |
| **verifier live, NO `Grades` — the actual exposure window (07-24 → 08-09)** | **15** | 0 | 0 |
| `Grades` live (08-10 →) | 0 | 0 | 0 |

- **The recorded "11 of 11" is really 15 of 15** — same shape, understated by four. A modest,
  in-direction correction; the mechanism's story does not change.
- **"not one has ever failed to close" needs its scope stated.** Within the exposure window it is
  exactly true (0 of 15 failed). Read as written — *ever* — it is false: **47 design-audit rows
  failed, all on 2026-07-18**, before the verifier existed, so those failures say nothing about the
  verifier. The word "ever" over-reaches; the argument it supports does not.
- Zero design-audit rows have closed under this type since `Grades` landed, consistent with
  `bugs_closed/213` having re-routed that producer to `dark_section_audit`.

**The lesson I am taking into the fix:** every figure I quote in the submission must say which
table(s) it came from and which era it covers, because on this schema "ever" and "the last seven
days" are different questions that look identical in SQL.
