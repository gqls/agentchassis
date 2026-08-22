# 355 — eight of the nine `page_components.content_data` writers cannot be observed losing keys

**Filed** 2026-08-22 by the `bugfix_238_regeneration_key_loss` lane (which owns this territory —
`scripts/who-owns.py 238`), at the owner's direction, to scope RFC_042 option **(c)**: the unified
content-loss detector.

**Status** OPEN — but read §2 before assuming there is damage to chase. **The interim census run
today found ZERO losses at the eight uncarried writers.** What is open is not a known loss; it is
that the instrument capable of *finding* one is blind in four specific, quantified ways, so the
zero cannot yet be trusted as an answer. This file exists so the detector is built from measurement
rather than from the census's silence — in either direction.

**Decision this depends on** ~~RFC_042 §6 is still with the owner. This file is the scoping of option
(c) so that decision can be taken on costed work rather than on a sketch. **Nothing here should be
built before that decision.**~~
**DECIDED 2026-08-22 — OWNER RULED option (c)** (decision record in RFC_042 §6). The build order is
§4's: A1 first, then A2+A3 in one commit, A4 not before a measured population exists. Per §8 this
file is now the implementation record; the joint-with-RFC_008 half of the §6 question was NOT ruled
and stays open.

> **Prior art, read first:**
> `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_042_content_data_has_nine_writers_two_write_disciplines_and_one_carried_funnel.md`
> · `bugs_closed/238_HANDOFF_2026-08-09_content_regeneration_drops_template_required_image_url_fields_leaving_empty_src_live.md`
> · `bugs_open/178_HANDOFF_2026-08-02_crosslink_item_regenerates_whole_section_and_loses_content.md` (the stop sign that
> names this detector as the sanctioned alternative to a sixth refusing floor)
> · `bugs_closed/194_HANDOFF_2026-08-04_four_of_six_save_page_sections_callers_never_map_sections_metadata_so_every_save_nulls_content_data.md`
> (the detector's embryo, `writeContentDataRegressionLog`)

> **On the 2026-07-31 owner ruling** (a `bugs_open/` file asserting a cross-cutting root cause is not
> filed until it has been through the `090` loop, or the filing session says plainly why it
> substituted equivalent first-hand verification). **Stated plainly: `090` was not run for this file,
> deliberately.** Its structural half — nine writers, two disciplines, one carried funnel — has
> already been through the loop as part of RFC_042 (run `68b3f9b6-1674-41a0-bc9e-c251192daaa1`,
> verdict UNVERIFIABLE, which independently re-read `planSection` and found the opposite of the
> hypothesised scenario). This file's own new claim is not a root cause at all: it is a **measured
> absence with a demand control**, every query of which is recorded inline in §2 and can be re-run
> in one paste. A diagnosis loop grades a theory about a cause; there is no theory here to grade.
> If the owner wants the *blindness* claim of §3 graded independently, `090` is the right tool and
> the symptom to file is in §3.5.

---

## 1. The defect, in plain terms

A page's sections live in `page_components`. Each row carries `content_data` — the field values that
section was built from. Some of those values are written by a language model; the rest are
**resolved from elsewhere in the system**: an image URL, a link target, a contact address, a query
result. The component's `input_schema` says which is which.

**Nine different pieces of code write that column, and exactly one of them is protected against
dropping a resolved value.** The protected one is the plan→save funnel, which every ordinary build
and re-render flows through; `PBP-039`'s carry sits inside it. The other eight inherit nothing, and
nothing tells their authors that.

That is RFC_042's finding and it is not new. **What this file adds is the attempt to measure the
exposure, and what that attempt ran into.** The short version:

> The system keeps a before-and-after archive of every `content_data` change, which is enough to
> count losses — and it does count them, correctly, for the funnel. But it **cannot say which piece
> of code made any given change**, it is **structurally blind to a writer that creates a row rather
> than changing one**, and for **58% of the changes it recorded it can no longer find the schema**
> needed to judge whether a lost key mattered. So a zero from it is a weak zero.

## 2. What was measured today, 2026-08-22 — and it is a clean zero

All figures `[MEASURED]` this session against the live database. Every query is reproducible from
what is written here; the shapes are in the lane RUNBOOK.

### 2.1 The instrument

`page_component_history` — an archive row written by a trigger on `page_components`. **6,210 rows,
all-time, beginning 2026-08-09** (`source='artefact_archive_trigger'`); every one carries
`slot_name`, so rows can be paired into consecutive generations of the same (page, slot).

Its `op` column splits the population, and the split is the load-bearing part:

| `op` | rows | pages | sites | which writers land here |
|---|---|---|---|---|
| `delete` | 5,830 | 470 | 24 | **the funnel** — `save_page_sections` is DELETE+INSERT, so its old rows arrive as deletes |
| `overwrite` | 380 | 213 | 19 | **the non-funnel writers** — every in-place `UPDATE page_components … SET content_data` |

The re-render path is *not* a separate writer: `rerender_page_sections_action.go` emits sections that
`save_page_sections` ingests (its own comments say so at :30, :1091, :1149; it holds no
`UPDATE page_components` of its own). So `op='overwrite'` really is the uncarried population.

### 2.2 The census — 0 losses at the non-funnel writers

Pair each `op='overwrite'` archive row (the pre-image) with the state that replaced it — the next
archive row for that (page, slot), else the live row. Resolve the component's schema, and count keys
that were **present and non-blank before** and are **absent or blank after**, restricted to fields
whose declared `source` is not `llm*`.

```
pairs total                        380
judgeable (schema + successor)     279      (73%; see §3.1 for the other 101)
NON-LLM KEY LOSSES                   0
LLM key losses (control)             0
```

### 2.3 The demand control — the same query finds 72 losses where losses are known to exist

**A zero whose control is also zero is not a measurement**, and this one nearly went out that way:
the LLM control returned zero too, which is exactly what a blind query looks like. So the same query
was run against the funnel population, where RFC_042 §4.6 had already measured losses by a different
method:

```
pairs total                       5,830
judgeable                         5,532
NON-LLM KEY LOSSES                   72     static=24  renderer=48
dates                          2026-08-09: 4   2026-08-11: 63   2026-08-12: 5   — and none since
```

**The query can see losses.** It finds them in the population where they are known to be, in the
class they are known to be (`static`/`renderer` — precisely what the `bugs_open/268` carry extension
closed on 2026-08-14), and it finds none in the ten days since. So §2.2's zero is a real zero within
its coverage, not an artefact of a broken query.

**This also independently re-confirms 238's closure** by a different route than the one used to close
it: 5,532 judgeable funnel transitions since 2026-08-12, zero losses.

### 2.4 The writer census, re-verified at HEAD

RFC_042's nine still stands, but **three files newly matched the census grep since it was written**
(the register's own landmine warns that this set grows while you are not looking) — each was opened
and each is **excluded**, for the record:

| file | why excluded |
|---|---|
| `v3_site_actions.go` | writes `sites.content_data` (a different table) and `page_components.build_status` |
| `store_generated_component_action.go` | `UPDATE page_components SET build_status` only (:1177) |
| `create_tool_component_regenerate.go` | `UPDATE page_components SET rendered_html` only (:316) |

And **two paths destroy `content_data` without writing it**, so no differ over writes will ever see
them — they delete the row: `remove_duplicate_page_sections_action.go:297` and
`internal/core-manager/admin/tool_admin_handlers.go:184`. Deliberately out of scope (§5), noted so
the next census does not "discover" them as a gap.

### 2.5 The surface a differ would protect

**2,627 field declarations across 174 active components; 1,042 declare an `llm*` source; 1,585 do
not.** Those 1,585 are what "schema-declared non-LLM key" means in practice. Every active field
declares a `source` (two exceptions fleet-wide), so the LLM/non-LLM discrimination is a string
prefix test, not a heuristic.

## 3. Why that zero cannot close the question — four blind spots, each quantified

### 3.1 The archive loses the schema pointer on 58% of rows

`page_component_history.component_id` is `REFERENCES page_components(id) ON DELETE SET NULL`. A row
archived by an in-place overwrite and *later* deleted by a regeneration has its pointer nulled — the
history keeps the content and loses the way to judge it.

```
history.component_id NULL (FK SET NULL)   221 of 380     ← the dominant cause
no `fields` in input_schema                57
page_components.component_id NULL           10
JUDGEABLE via the FK                        92           ← 24%
JUDGEABLE via the slot fallback            279           ← 73%, what §2.2 used
```

A census keyed on the FK alone reports **92** and reads as complete. The fallback — resolve the
schema from the live row at that (page, slot) — recovers most of it, at the cost of being wrong
where a component was swapped at that slot. **101 pairs remain unjudgeable by any route.**

### 3.2 The archive cannot name the writer

`page_component_history` has an `application_name` column, and it looks exactly like the answer. It
is not. Measured over the last 14 days, every application-side write carries the pgx connection
default — `app - 10.20.99.74:35564` and 3,000-odd siblings — and hand-run SQL carries `psql`. A pod
IP is not a writer; several of the nine live in the same binary.

**So even a positive result here could not be attributed.** "Which writers actually lose keys in
practice" — RFC_042 §4.3's stated reason for preferring the detector over the guard — **is not
answerable from the archive at all**, at any sample size. That is the single strongest argument in
this file for building something, and it is also the cheapest thing to fix (§4, A1).

### 3.3 The archive never sees a row being born

`pch_op_check` permits `op IN ('overwrite','delete')`. **The trigger does not fire on INSERT.** A
writer that creates a row with incomplete `content_data` is invisible to every pairing method,
for ever — and this is not hypothetical: the tool writers deliberately write `'{}'`, and
`adopt_verbatim`, `create_report_page_action` and the CLI importer all create rows.

Bound on the class today: **119 of 1,850 deployed rows (6.4%) carry no `content_data` at all** — 77
NULL, 42 `{}`. Not all of those are defects (a tool widget legitimately has no structured content),
which is precisely why it needs a detector that knows the schema and not a census that counts nulls.

### 3.4 The window is 13 days

Trigger rows begin **2026-08-09**. The older `save_page_sections_overwrite` archive goes back to
**2026-03-16** (20,664 rows) but **carries no `slot_name`** — it cannot be paired by slot, only by
page, which is the page-level all-or-nothing granularity that made `writeContentDataRegressionLog`
unable to see 238 at all (a page that lost 11 of 58 keys while every section still had
`content_data`).

### 3.5 If §3 is to be graded independently

Symptom to file at `090`, phrased to earn a verdict: *the archive trigger on `page_components`
records `op='overwrite'` and `op='delete'` pre-images with `slot_name`, and `application_name` is
populated from the connection rather than the caller; determine whether any surviving column,
table or index can attribute an archived `content_data` transition to the Go call site that caused
it.* Point it at `page_component_history`, the trigger definition, and the nine writers in §2.4.
Assert no counts — the loop fetches them.

## 4. Fix candidates, ordered by what closes the door

Ranked by the estate's rule: prefer what makes the bad state unrepresentable over what asks an
operator to remember something.

### A1 — make the write self-attributing (transaction-scoped `set_config('application_name', …)`) · **do this one first**

One statement per write site, before the write, inside the existing transaction:

```go
// so page_component_history.application_name names the CALLER, not the socket
_, _ = tx.ExecContext(ctx, `SET LOCAL application_name = $1`, "action:section_editor.replace")
```

> **CORRECTED 2026-08-22, hours after filing — the sketch above DOES NOT RUN, twice over.**
> Caught by the parallel 238 session before anyone shipped it. (1) `SET`/`SET LOCAL` take **no bind
> parameters** in the extended protocol, and the `_, _ =` would have discarded the refusal — a
> silent no-op on every write. The working form is
> `SELECT set_config('application_name', $1, true)` (third argument = transaction-scoped).
> (2) "inside the existing transaction" assumed transactions that mostly do not exist: **only 2 of
> the 9 write sites hold one** — on a bare autocommit statement, `set_config(..., true)` is a
> statement-scoped no-op, and a session-level `SET` is forbidden outright behind pgbouncer
> `pool_mode=transaction` (it leaks the stamp onto other clients' work on the shared server
> connection). So the shipped form wraps each bare write in a short transaction
> (`stampedExecContext`, `content_write_stamp.go`) and stamps the two real transactions in place
> (`stampWriterTx`). Best-effort both ways: a Begin or stamp failure falls back to the exact
> unstamped statement — attribution must never break a write.

**No schema change, no new column, no config key, no migration.** The trigger already captures the
value; today it captures a default nobody chose. It closes §3.2 permanently and **retro-actively
makes every future census attributable**, including censuses nobody has thought of yet.

It is also the honest precondition for A2: a detector that reports "a key was lost" without naming
the writer produces work nobody can route. Worth shipping **even if the owner picks RFC_042 option
(a)** — it costs nine lines and it is the difference between a measurable estate and an unmeasurable
one.

⚠ Verify at the artefact, not at the diff: after a roll, `SELECT DISTINCT application_name FROM
page_component_history WHERE created_at > now() - interval '1 day'` must show `action:*` values.
`SET LOCAL` outside a transaction is silently scoped to the statement — if the write is not in a
`tx`, the stamp does not reach the trigger, and both the code and the test can look right.

### A2 — the per-key differ (the detector proper)

At the write seam, in-process, with the schema already loaded: compare the schema-declared non-LLM
keys of the row being replaced against the row being written; record a durable finding per lost key.

**Why in-process rather than as a SQL sweep over the archive:** it is the only version that answers
§3.1 (the schema is in hand, no FK to survive), §3.2 (the caller is known by construction) and §3.3
(an INSERT is a write like any other). A sweep can only ever re-run §2.2 with better joins.

**Extend, do not invent.** `writeContentDataRegressionLog` (`save_sections_metadata_source.go:250`)
already writes exactly this shape of finding — it is page-level and all-or-nothing, which is why it
could not see 238. Per-key resolution of an existing record is a smaller change than nine call-site
conversions, and `LogActionEntry`/`agenterrors` is already the leaf package both sides of the
coordinator edge can import (the cycle that once blocked this was closed on 2026-08-06, RFC_012
option B — the correction is in that function's own comment).

### A3 — the consumer, shipped in the same commit as A2 · **non-negotiable**

**This estate has now written two loss-finding codes that nothing reads:**

| code | rows | window | resolved | readers in Go/SQL/py |
|---|---|---|---|---|
| `CONTENT_DATA_REGRESSION` | 41 | 2026-08-08 → 08-21 | 0 | **0** |
| `STRUCTURAL_KEY_CARRY_MISS` | 28 | 2026-08-11 → 08-17 | 0 | **0** |

Grep-verified: both codes appear only at their write sites and in prose. **A third unread code is
not a detector; it is a way of feeling detected.** So A2's acceptance includes a consumer — the
cheapest honest one is a work item per finding through the existing router, or a daily check that
writes one `doc_notes` row per run *including on clean results*, so that a missing row means the job
did not run rather than "nothing is wrong" (the shape the optional-key-budget cron already uses).

Either way it must be **the same commit**. A2 without A3 is the failure this row records.

### A4 — refusal, per-caller opt-in, default OFF · **only after A2/A3 produce a population**

Per the owner's 2026-08-02 ruling, new authority on a shared seam ships as an opt-in field whose
unsafe side is the default. Budget context, `[MEASURED]` today: no shared action is over the RFC_022
limit of 10; `plan_sections` sits at 7. **A2 and A3 add no config key; A4 adds one per opting
caller** and must re-run `scripts/audit-optional-key-budget.sh` in its own submission.

`bugs_open/178`'s stop sign against a sixth refusing floor in `save_page_sections` still stands and
this does not breach it — A4 refuses at the *opting* caller, not in that function.

### A5 — accept the zero, re-run the census on a schedule

Entirely legitimate given §2, and cheaper than all of the above. It costs the four blind spots
staying open, and it is only a decision if the schedule exists: the census in §2.2 plus its §2.3
demand control, monthly, recorded where a reader will meet it. **A5 without the demand control is
worse than nothing**, because it manufactures a reassuring zero every month.

## 5. Explicitly not in scope

- **The two DELETE paths** (§2.4). A row that is removed is not a key that was lost; conflating them
  makes every legitimate dedup look like damage.
- **Repairing anything.** Nothing found in §2 needs repair — there is nothing there. If A2 later
  finds losses, remediation routes through the existing `required-fields-missing-handler` widening
  (238 §Phase B), not a new router: **RFC_030 is RULED — no fourth bespoke single-type router.**
- **`rendered_html`.** That is `RFC_008`'s column and its own seam question. RFC_042 asks to be
  answered jointly with it; that is a decision about *both*, not a licence for this file to widen.
- **The `planSection` resolved-side gaps** (blank-beats-stored, `query.*` error). Real in the code,
  **0 observed instances**, recorded in RFC_042 §4.6 with the exact evidence that would justify
  shipping them. Still recorded, still not shipped.

## 6. Acceptance — how anyone tells whether this worked

| candidate | acceptance test | the control that makes it falsifiable |
|---|---|---|
| A1 | `SELECT DISTINCT application_name … WHERE created_at > now() - interval '1 day'` shows `action:*` | a write site deliberately left unstamped must still show `app - <conn>` — otherwise the stamp is coming from somewhere else |
| A2 | a unit test per writer: mutate the write to drop one schema-declared non-LLM key → the named test must fail | **mutation-proved, not assertion-proved**: a mock's own bookkeeping cannot assert a negative |
| A2 | on real traffic: a finding whose (page, slot, key) can be confirmed absent at the served artefact | grep the live page for the value, never the row — a repaired row is not a repaired page |
| A3 | one consumer row exists per finding, and one exists on a clean run too | delete the finding source and re-run: the clean-run row must still appear |
| A4 | an opting caller refuses; a non-opting caller does not | the non-opting caller is the control, and it must be exercised in the same run |
| A5 | the monthly census's demand control returns non-zero | if the control ever returns zero, the census is blind and its zero must be discarded |

## 7. Traps for whoever picks this up

- **A zero here is nearly always blindness until proven otherwise.** §2.2 and its LLM control both
  returned 0 on the first run; only the funnel demand control (§2.3) distinguished "no losses" from
  "no vision". **Never report a loss census without running the `op='delete'` control in the same
  breath.**
- **`application_name` looks like attribution and is a socket** (§3.2).
- **`op` is a proxy for the writer, not the writer.** It says DELETE+INSERT versus in-place UPDATE.
  A future non-funnel writer that deletes and re-inserts lands in the funnel bucket and would be
  mis-attributed; a funnel change to in-place UPDATE would invert the whole split.
- **Joining history to the schema through `component_id` silently drops 58% of rows** (§3.1) and
  reports a clean, plausible number.
- **`save_page_sections_overwrite` rows carry no `slot_name`** — they widen the window back to March
  but only at page granularity, which is the granularity that could not see 238.
- **INSERT is not archived** (§3.3) — no pairing method will ever see a row born empty.
- The `page_component_history` entry in `LANDMINES.md` already carries three related traps. Read it
  before writing any query against this table.

## 8. Obligations if any of this is built

- **Council gate** — `platform/` and `internal/` are in scope; A1 and A2 are both platform code. One
  run per coherent task; `Council-Submitted:` trailer if committing before the verdict.
- **Concept register** — A2 is a new reusable mechanism and needs an entry the moment it exists, in
  `docs026_concept_register/register/page-build-pipeline.md`, beside PBP-039 and PBP-040. A1 is a
  convention that every future writer must follow, which makes it register-shaped too.
- **RFC_042** — whichever option is taken, record the decision in that file. If (c) is taken, this
  file becomes its implementation record.
- **The standing five** live in `docs/agent_docs/docs024_key_docs_latest/bugfix_238_regeneration_key_loss/`.

## 9. Closing bar

This bug is closable **without any code** if the owner takes RFC_042 (a) or (e) and A5's schedule is
put in place with its demand control — the honest finding today is that the eight uncarried writers
show no losses. It is closable **with** code once A1 is live and A2+A3 have produced either findings
or a demonstrably non-blind zero over a full window.

**What must not happen is the third silent code.** Two already sit unread; the whole argument for a
detector is that it measures something, and a measurement nobody reads is not one.

---

## 10. IMPLEMENTATION RECORD — built, deployed and behaviourally proven 2026-08-22, the day of filing

The owner ruled option (c) and separately directed: *"fix that and also ship both the detector and
the solution that reads its output."* All four pieces shipped the same day:

| piece | commit(s) | state |
|---|---|---|
| **mig 552** — content-only UPDATEs archive too (closes a FIFTH blind spot found after filing: 357's trigger is gated on `rendered_html` changing, so the one change this class is about was the one change the archive could not see; the admin handler's dynamic SET is the live producer) | `e7567d1fc` | committed; **apply follows council corr `f5550f04`** |
| **A1** — writer stamps, in the corrected form above | `8552e621d` + `0702fb9cb` (the untouched-twin advisory was a real catch: the 344 chrome archive captures `application_name` too) | committed, INERT until the next roll |
| **A2+A3** — `cmd/content-loss-check`, detector and reader ONE binary (register **PBP-046**; A1 is **PBP-047**) | `cba51ad1d` (+`d56fd6b11` makefile fix) | **LIVE**: CronJob daily 07:05 UTC, image `v1.0.1324` = that commit (provenance label verified before push) |
| **A4** — refusal | — | correctly unbuilt: no measured population |

**One deviation from this file's own A2, recorded not silent:** the detector does NOT extend
`writeContentDataRegressionLog` — that function sits in the FUNNEL, so per-key-ifying it upgrades
the one writer PBP-039 already carries and does nothing for the eight uncarried ones, which never
route through it. The shipped shape is a daily sweep over `page_component_history` (with 552
closing its content-only hole): structural coverage of all nine writers plus psql plus every future
writer, and the reader in the same binary so it cannot be shipped without it.

**First writing run (job `content-loss-check-manual-20260822-113722`, pod exit 1 — the honest
steady state while parked damage stands):**
- instrument: canary ok; demand control re-found exactly **72** (0 would have refused the run);
- coverage 6,216 pairs / 5,820 judgeable over 21 days;
- **72 findings filed** (all pre-fix funnel, as §2.3 predicted), deduped for every future run;
- **48 findings stamped resolved — the first `resolved=true` rows in `agent_error_log`'s entire
  history** (40 `CONTENT_KEY_LOSS` healed, 7 carry-miss healed, 1 row-gone); **93 genuinely still
  open**, including the gated-field class 238 said had no producer — it has one now;
- state census: **32 required non-llm blanks across 13 (page, slot)s** — aao's parked grid (11),
  leopardess `who-we-help` (6), gamesdesign's honest no-email gap (1), and NEW visibility on
  dartsonline `brands-index`/`shop-index` `category-listing` (3+3), finetuning
  `ai-guides`/`insights`/`ai-readiness-quiz`, gaswholesalers `client-case-studies`/
  `fuel-industry-insights`, leopardess `ai-readiness-quiz` (1 each) — surfaced for owners, not
  fixed by this lane;
- heartbeat `doc_notes` row written (`subject_key='content-loss-check'`).

**§3 scorecard after shipping:** §3.1 (58% FK loss) — mitigated by the slot fallback, residual
honest in the judgeable count; §3.2 (attribution) — CLOSED by A1 at the next roll; §3.3 (INSERT
blind) — covered by the state census each run; §3.4 (13-day window) — grows a day per day and the
detect window is 21; the FIFTH (content-only UPDATEs) — CLOSED by 552 on apply.

**Retention (from `bugs_open/358`, filed the same day as the class file this section's A3 rule
anticipated):** finding rows expire — 30d unresolved, 14d RESOLVED (mig 466), so resolving halves a
row's remaining life. Accepted deliberately: nothing depends on the rows persisting — the durable
records are the heartbeat and the state census, both re-derived every run.

**What closes this file:** the council verdict on `f5550f04` read and 552 applied; one scheduled
run (07:05 UTC) producing its heartbeat unattended; and after the next roll, one archive row
carrying an `action:*` writer. Then this moves to `bugs_closed/` with the fixed-AND-live bar met.
