# 358 — `agent_error_log` finding codes are write-only, and retention deletes them unread

**Filed** 2026-08-22 by the session working `bugs_closed/238`'s succession, at the request of the
`bugfix_238_regeneration_key_loss` lane (which is building `cmd/content-loss-check`, `bugs_open/355`
candidates A2+A3, and whose docs point here as the class owner). 355 §A3 found two loss-finding
codes nothing reads; scoping that found the CLASS: **it is most of the table.**

**Status** OPEN. Nothing here is a lost page or a broken build — the damage is that the estate keeps
paying for evidence it never collects: detectors write durable-looking findings, no automated
consumer ever acts on them, and a retention job deletes them on a 30-day clock. Every number below
is `[MEASURED]` this session; every query and grep is restated inline.

> **On the 2026-07-31 owner ruling** (a `bugs_open/` file asserting a cross-cutting root cause is
> not filed until it has been through the `090` loop, or the filing session states plainly why it
> substituted equivalent first-hand verification). **Stated plainly: `090` was not run,
> deliberately** — this file asserts no causal theory. It is a census of readers and writers, every
> row of which is a grep or a query reproduced inline, re-runnable in one paste. Precedent:
> `bugs_open/355` filed on the same grounds, same day, and its reasoning survived. If independent
> grading is wanted, the symptom to file is in §7.

> ## CONTRIBUTION 2026-08-22 (later) — lane picked up; THREE of this file's claims corrected
>
> Owning lane from now: `docs/agent_docs/docs024_key_docs_latest/bugfix_358_unread_finding_codes/`
> (`PLAN_2026-08-22_unread_finding_codes.md` carries the design and the dated corrections;
> `NOTES_unread_finding_codes.md` the evidence log; `RUNBOOK_unread_finding_codes.md` the queries).
> Council submission `be1fd678-0836-4f32-90a6-8927b2463fee`. `090` filed on the §3 claim below,
> run correlation `c965bfec-993a-4b2b-88ba-d44549c81df1`.
>
> **CORRECTION 1 — "the `resolved` workflow has never been used once" (§1) expired within the
> day.** Live now: **48 resolved rows**, all stamped 2026-08-22 10:40 UTC by `resolved_by =
> 'content-loss-check:healed'` / `':row_gone'`. The `bugfix_238_regeneration_key_loss` lane's
> checker (`cba51ad1d`) is the first user of the triple in the table's history, and it also added
> a code, `CONTENT_KEY_LOSS` (72 rows), which it writes AND reads AND resolves. §2.2's
> "reader-with-writer from birth" claim gains a fourth instance and a working exemplar.
>
> **CORRECTION 2 — `RESOLVER_CONFLICTING_CANDIDATES` is NOT unruled, and it is this file's
> headline example.** §4 B1 says *"nobody has ruled which"*. Somebody has: it and
> `RESOLVER_MAPPING_BYPASSED` are Phase-1 instrumentation under
> `architecture_review/RFC_029`, with an owner, a stated observation window, **six dated reads**
> (§10.5, §10.6, §10.7, §10.9, §10.11, §10.12) and an **owner ruling of 2026-08-18** (§10.13)
> sequencing Phase 2 on their evidence. The concept register carries the architecture seat's
> scope note verbatim (`register/contracts-and-standards.md:511`). The no-dedup design this file
> reads as waste is deliberate — *"frequency is the population §9's disconfirmation clause
> needs."* **Consequence for §4 B1:** its three outcomes (consume / demote / keep-as-human-
> evidence) have no slot for a time-boxed instrument with an owner. A **fourth disposition** is
> needed, and the triage must join against OWNERS, not only against automated readers.
>
> **CORRECTION 3 — the load-bearing one, and it changes where the fix goes.** This file assumes
> (reasonably, from its own §2.1 grep method) that the Go write sites are the population.
> `platform/orchestration/agenterrors/agenterrors.go:3` declares itself *"The ONE writer against
> `agent_error_log`"* and RFC_012 did retire nineteen hand-copied INSERTs into it — but
> `grep -rn "INSERT INTO agent_error_log"` across every language finds **five** paths, not one:
> the seam; `store_generated_component_action.go:1439` (own INSERT, kept deliberately — a prior
> council round's edit-quality and guardian seats objected to consolidating it, recorded at
> `:1428`); `internal/agents/contentcreator/claims_guard.go:184`, which **cannot** use the seam
> (`*pgxpool.Pool` at `agent.go:92` vs `agenterrors.Write`'s `*sql.DB` — a type-level barrier);
> `sql_for_agents/214_build_dispatch_watchdog.sql:108`; and `cmd/content-loss-check/main.go:292`.
> **So the obvious home for B2's ratchet — a guard inside the one writer — would be blind to four
> of the five writers while looking complete.** The fix is DB-authoritative instead:
> `SELECT DISTINCT error_code` sees every writer regardless of language, seam, or whether the
> value is a literal, and parses nothing. Also worth knowing before building a per-code consumer:
> **there is no index on `error_code`** (`\d agent_error_log` — four indexes, none on it).
>
> Two figures re-measured, both still standing: 45,507 rows (was 45,426), oldest still
> 2026-07-23, and retention confirmed live — `scheduled_tasks` row `database-cleanup`, enabled,
> `interval_seconds` 3600, last triggered 2026-08-22 10:07 UTC.

## 1. The defect, in plain terms

When a platform mechanism notices something wrong that it will not or cannot fix — a review being
superseded, a plan section name dropped, a deploy of an archived page refused, two resolver
candidates in conflict — it records a **finding**: a row in `agent_error_log` with a named
`error_code`. Writing that row is cheap, it feels like detection, and the codebase has grown at
least **nineteen distinct finding codes** this way.

**For sixteen of those nineteen, the row is never read by anything automated.** And the table has
a retention job, so the rows do not even accumulate as evidence for a human to find later — an
unresolved row is deleted at 30 days, a resolved one at 14. The system's memory of its own
detections is a 30-day sliding window that, for most codes, nothing looks at before it slides.

Two aggravating facts, both measured:

- **The `resolved` workflow has never been used once.** 45,426 rows all-history, `resolved = false`
  on every single one; `resolved_at`/`resolved_by` never written. The triple exists (schema
  `022_agent_error_log.sql:38-41`), and no code path or operator has ever set it.
- **The one code with a fully closed automated loop has zero rows** — `DEPLOY_STAMP_REFUSED_ON_SKIP`
  is read back by its own writer for strike-counting, and it has never fired in the window. The
  codes that fire the most are read the least (top unread code: 9,615 rows in five days).

## 2. The census, 2026-08-22

Totals: `SELECT count(*), count(*) FILTER (WHERE resolved), min(occurred_at) FROM agent_error_log;`
→ **45,426 / 0 / 2026-07-23** — the oldest row sits exactly on the 30-day retention boundary, which
is the live confirmation that the cleanup runs.

> **CORRECTED 2026-08-22 10:40Z, ~1 hour after filing — the "resolved has NEVER been used" claim
> stopped being true while this file was being written, and it is worth reading as a dated event
> rather than as an error.** `cmd/content-loss-check` (the `bugfix_238_regeneration_key_loss` lane's
> `bugs_open/355` A2+A3, shipped the same day) ran and **stamped the first `resolved = true` rows in
> the table's history**: re-verified independently for this correction —
> `count(*) FILTER (WHERE resolved)` = **48**, first and last `resolved_at` 10:40:21–10:40:22Z, two
> distinct `resolved_by` values (`content-loss-check:healed`, `content-loss-check:row_gone`).
> So §2's figure is now the **before** half of the sentence, and it keeps its evidential force:
> nothing had ever used the workflow in the 30 days the table retains, until a check shipped *with a
> reader attached*. What this does NOT license is repeating "45,426 rows, 0 resolved" as current
> state — cite it with its timestamp, as the state the estate ran in until 10:40Z on 2026-08-22.
> The general lesson is the estate's own: a `[MEASURED]` claim about STATE expires; a dated EVENT
> does not.

Retention mechanism: migration `466_orchestration_status_vocabulary.sql:184-189` embeds a
`pre_query` that `DELETE`s `resolved AND older than 14 days` OR `NOT resolved AND older than 30
days`. **Note the inversion: marking a row resolved makes it die faster.**

### 2.1 Finding-shaped codes with NO automated reader — 16 codes, 10,673 rows live today

"Finding-shaped" = written deliberately by a detector, guard, or audit as its record of something
noticed (as opposed to the operational-failure codes in §2.3). "No automated reader" = no Go, SQL,
Python or shell in the repo, and no live `agent_definitions` workflow, selects rows by this code —
verified by grepping the **literal AND the Go constant that carries it** (see §3 for why both).

| code | rows | window | writer |
|---|---|---|---|
| `RESOLVER_CONFLICTING_CANDIDATES` | 9,615 | 08-16→08-21 | `datahelpers/resolver_findings.go:55` const, emitted `datahelpers/unified_extractor.go:589` |
| `RESOLVER_MAPPING_BYPASSED` | 485 | 08-16→08-17 | `resolver_findings.go:59`, emitted `datahelpers/action_inputs.go:932` |
| `CONTENT_LINK_REPAIR_DETAIL` | 401 | 07-27→08-22 | `validate_page_content.go:644` const |
| `VALIDATION_ERROR_DROPPED` | 202 | 07-26→08-22 | TWO writers: `messaging/validation_drop.go:109`, `agentbase/agent.go:1489` |
| `PLAN_SECTION_NAME_DROPPED` | 140 | 08-17→08-20 | `component_name_resolver_menu.go:183` |
| `component_validation_rejected` | 100 | 08-15→08-22 | `store_generated_component_action.go` |
| `ARCHIVED_PAGE_DEPLOY_REFUSED` | 80 | 08-14→08-21 | TWO writers: `git_deployer_actions.go:103`, `v3_site_actions.go:911` |
| `CONTENT_CLAIMS_FLOOR_DETAIL` | 59 | 07-31→08-21 | `save_sections_claims_guard.go:104` const |
| `component_validation_orphan_schema_field` | 57 | 08-03→08-18 | `store_generated_component_action.go` |
| `TRUNCATION_DEGRADED_REVIEW` | 41 | 07-26→08-02 | `diagnose_council_decide_action.go:625` — **dies ~08-25 on the 30-day clock** |
| `tool_crosslink_not_emitted:*` | 37 | 07-31→08-22 | `create_tool_cross_link_items.go` |
| `CONTENT_DATA_LINK_AUDIT` | 34 | 08-02→08-18 | `save_sections_content_data_links.go:62` const |
| `REVIEW_SUPERSEDED_BY_PASSING_SAVE` | 25 | 07-23 only | `reconcile_superseded_reviews_action.go:181` — **oldest rows in the table; days from deletion** |
| `NO_CHANGE_GATE_UNREADABLE_RESULT` | 11 | 08-14→08-17 | `complete_work_item_no_change.go:504` |
| `COMPONENT_COLLISION_DIVERTED` | 11 | 08-19→08-22 | `store_generated_component_action.go` |
| `RETRACTION_AUDIT` + `RETRACTION_REFUSED` | 14 | 08-04→08-20 | `retract_asset_files_action.go`, `retract_page_deployment_action.go` |

Smaller same-shape codes, same verdict: `CONTENT_DATA_ENVELOPE` (7, `content_data_envelope_guard.go:142`),
`RENDER_AUDIT_TRUNCATED` (3), `tool_birth_instance_scope_refused` (9), `tool_regeneration_hollow_blocked`
(3), `PLAN_PAGE_MERGE_LOSSY` (2, `write_site_plan_action.go:391`), `component_write_shared_blocked` (1).

**Excluded from this list, owned elsewhere:** `CONTENT_DATA_REGRESSION` (41) and
`STRUCTURAL_KEY_CARRY_MISS` (28) — the two 355 §A3 named; `cmd/content-loss-check` (in progress,
`bugfix_238_regeneration_key_loss` lane, owner directive 2026-08-22) ships as their consumer.

### 2.2 The three codes that DO have automated readers — the positive pattern

| code | rows | the loop |
|---|---|---|
| `CONTENT_VALIDATION_BLOCKER_DETAIL` | 215 | `reconcile_superseded_reviews_action.go:228` reads the last 20 per page and merges blocker values into its supersede finding |
| `DEPLOY_STAMP_REFUSED_ON_SKIP` | **0** | `page_build_failure_guard.go:131` counts its own rows for the 7-day strike ladder — writer and reader are one mechanism |
| `BUILD_DISPATCH_STALLED` | **0** | the watchdog's own CTE (`214_build_dispatch_watchdog.sql:103-105`) reads it for self-dedup before raising again |
| `CONTENT_KEY_LOSS` | 72 (40 resolved) | **added 2026-08-22** — `cmd/content-loss-check` writes it and re-reads its own findings each run, stamping `healed` / `row_gone`. Shipped reader-with-writer in one commit, per `bugs_open/355` A3 |

All four were built as **read-back loops from birth** — the reader shipped with (or before) the
writer. **No code has ever acquired a reader after the fact**, and the fourth is the first test of
that claim run deliberately rather than observed retrospectively: 355 §A3 made "same commit"
non-negotiable *because* of this file's census, and the resulting code is the only entry here whose
loop closed on the day it was written. That is the strongest evidence available for the same-commit
rule being the general law, not a lane preference.

> **Also corrected 2026-08-22:** `STRUCTURAL_KEY_CARRY_MISS` (28 rows, listed in §2.1's exclusions as
> *owned, consumer in flight*) **now has that consumer** — 8 of its 28 rows carry `resolved = true`
> from the same run.
>
> > **CORRECTED AGAIN, same day, and the error is instructive enough to keep rather than tidy.** This
> > block first said: *"`CONTENT_DATA_REGRESSION` (41) remains at 0 resolved: it is page-level and
> > all-or-nothing, so the per-key check cannot adjudicate it, and it stays in §2.1's population."*
> > **Both halves were wrong.** The check reads it explicitly — `cmd/content-loss-check/main.go:327`
> > selects unresolved rows of three codes including `codeRegress`, and `:423` is its own disposition
> > arm (page gone → `page_gone`; page holds ≥1 component with non-empty `content_data` → `healed`;
> > else `open` with a stated reason). All 41 graded `open` on the first run. And the reason given was
> > wrong too: the granularity of a finding and the granularity of its heal predicate are independent
> > — **readable ≠ per-key**; a page-level claim is adjudicated by a page-level query.
> > **The inference that caused it — "0 resolved ⇒ 0 readers" — is the exact discrimination this file
> > exists to make, and it fails in the one direction that matters: a finding that is read and
> > correctly left OPEN is indistinguishable, in the rows, from a finding nobody reads.** Caught by
> > the lane that wrote the check. Logged in `WRONG_CALLS.md` (2026-08-22).
> > So `CONTENT_DATA_REGRESSION` **has a consumer** and leaves §2.1's population; what it has instead
> > is a standing open population, which is §4's B-track lead below.

### 2.3 Scope boundary: what "unread" does NOT mean here

- **Generic readers exist and are healthy.** `diagnose_load_runtime_action.go:278` (newest-N context
  dump per site), `work_item_failure_ladder.go:611` (burst probe, by message signature not code),
  the `dbcontext` toolkit, and five live agent workflows (`council-gate`, `fix-proposer`,
  `feature-designer`, `report-builder`, `diagnose-agent` — `default_config` references measured in
  `agent_definitions`) all read the table **as diagnostic context, any code, newest first**. So the
  honest claim is: *no automated consumer acts on these codes' semantics* — the rows serve as
  scrollback, never as findings.
- **Operational codes are not in the class.** `UNKNOWN` (18,793), `PROCESSING_FAILED` (5,877),
  `LLM_API_ERROR` (5,681), `TIMEOUT` (2,165), `CHILD_ORCHESTRATION_FAILED` (1,035) are failure
  plumbing — classifier defaults and error-step routing. Generic diagnostic reading is their
  correct consumption; they are why the table exists.
- **Hand-run verification queries** (e.g. `516_..._HOLD.sql:112-114`, `547_...sql:109`) read codes
  once, at ship time, by a human. Real, but not a consumer: they stop when the session ends.

## 3. Why the class is self-sustaining — three mechanisms, all observed

1. **Writing a finding feels like detection.** 355 said it once: *"a third unread code is not a
   detector; it is a way of feeling detected."* Each new guard cites the previous guard's record
   shape as precedent (`save_sections_content_data_links.go:61` cites `CONTENT_LINK_REPAIR_DETAIL`;
   the claims guard cites both) — the record shape propagates; the missing reader propagates with it.
2. **The census that would notice is itself easy to get wrong.** A reader that consumes via a Go
   constant is invisible to a grep near the literal: `DEPLOY_STAMP_REFUSED_ON_SKIP`'s reader
   (`page_build_failure_guard.go:131`, `WHERE error_code = $1` bound to a const declared at `:65`)
   was missed by exactly that method during this file's own preparation, caught only by re-grepping
   every constant name. A "zero readers" verdict produced by literal-grep alone answers the
   question it encoded, not the one asked.
3. **Retention erases the backlog that would otherwise embarrass someone into reading it.** A code
   can run at 2,000 rows/day for ever and the table never grows past 30 days of it. The cost is
   invisible on every dashboard the estate has.

## 4. Fix candidates, ordered by what closes the door

### B1 — per-code triage, owner-decided (the fixing session's deliverable)

For each §2.1 code, one ruling: **consume** (name the reader and build it as its own task — the
`content-loss-check` heartbeat shape, or a work item via `insertWorkItem`), **demote** (it is a log
line wearing a finding's clothes — rewrite the write site as `logger.Warn` and delete the code), or
**keep as human evidence** (legitimate for hand-run forensics — say so in a comment at the write
site, with the retention window named, so the next reader knows the 30-day clock is accepted).
`RESOLVER_CONFLICTING_CANDIDATES` first: 9,615 rows in five days is either the estate's loudest
unheard alarm or its most expensive no-op, and nobody has ruled which.

### B2 — the ratchet: no NEW code ships unread (makes the bad state unrepresentable)

> **Independent support, 2026-08-22:** the council's `architecture` seat, reviewing
> `content-loss-check`, objected on the *writer* side of the same gap — *"the attribution contract is
> enforced by convention only; no lint ties a NEW writer to the helper"* — and the lane's named
> follow-up is a check in the `check_unrepaired_component_write` family (`bugs_open/355` §10). That
> is this candidate approached from the opposite direction, and the pair is the complete shape:
> **a reader attached at birth (A3's law) plus a lint tying new writers to the seam (B2).** Whoever
> takes B2 should build it once for both, not twice.

The acks-file + source-scan-test pattern already proven by `optional_key_budget_acks.json` +
`optional_budget_cron_parity_test.go`: a registry mapping each `error_code` literal to its declared
consumer (or `human-evidence-only` + reason), enforced by a test that greps `ErrorCode:` literals
and constants in Go and fails on an unregistered one. New-code-without-reader then breaks the
build instead of becoming row 17 of §2.1. **Not built here** — it is platform code, council-gated,
and it should be sized by whoever takes B1 (the registry's initial content IS B1's output).

### B1a — the first triage case is already staged: 41 rows where the WRITER over-reports

**A worked lead for whoever takes B1, handed over by the `content-loss-check` lane rather than
shipped by it.** `CONTENT_DATA_REGRESSION`'s 41 open findings are **one uniform class**
`[MEASURED 2026-08-22, re-verified independently for this file]`: every row is a `tool-*` page,
every one `metadata_field_origin=configured`, `incoming_sections=1`,
`metadata_field=rerender_sections.sections_metadata`. They are almost certainly the
`bugs_closed/194` shape where the caller legitimately has no sections metadata — in which case the
finding should never have been filed, and the fix is a **declaration**
(`expects_no_sections_metadata`, `save_sections_metadata_source.go:90`, whose `declared_absent`
origin `shouldReportContentDataLoss` already exempts), not a reader. Precedent: migration `313`
declared it for tool *recreation*; these 41 arrive by a path `313` did not cover.

**This is the triage outcome B1 calls "demote", and it names a defect class B1's three options did
not: the writer over-reporting.** Under-reading and over-reporting both end as rows nobody acts on,
and they need opposite fixes — so B1's per-code ruling should ask *"is this finding true?"* before
*"who reads it?"*.

> ⚠ **The obvious completion is a trap, and it is a design decision, not a config one-liner.**
> `rerender_sections` is a step of exactly **one** agent — `page-rerender` — which is the shared
> re-render path **every ordinary page flows through** (`[MEASURED]`: 1 agent carries the step).
> Declaring `expects_no_sections_metadata` *on the step* would therefore suppress the regression
> record **fleet-wide**, silencing the genuine 194-class signal for content pages in order to quiet
> tool pages. The discrimination must key on **the page's character, not the step** — either the
> tool pipeline marks its own pages, or the finding writer consults the page before filing.
> Until that is decided, the 41 correctly stay open: the pages genuinely hold no structured
> `content_data`, and an open finding naming a known class beats a suppressed one.

### B3 — accept and document

Cheapest: a comment at every §2.1 write site stating "no automated reader by design; human evidence
only; rows expire at 30 days". Honest, costs nothing, closes no door — B3 without B2 re-runs this
census in six months with more rows in the table.

## 5. Explicitly not in scope

- **Building readers for the sixteen codes.** Each needs domain judgement about what acting on it
  means; a blanket consumer would be a fourth way of feeling detected.
- **The two codes `content-loss-check` consumes** — owned by the `bugfix_238_regeneration_key_loss`
  lane, in flight now.
- **Changing retention.** The 14/30-day windows may well be right; what is wrong is codes whose
  entire output falls inside them unread. Rule on B1 first.

## 6. Acceptance

| candidate | test | control |
|---|---|---|
| B1 | every §2.1 code has a recorded ruling (consume/demote/keep) | a code with no ruling still listed = not done |
| B2 | a branch adding `ErrorCode: "TEST_UNREGISTERED_X"` fails the scan test | the same branch with a registry entry passes — mutation-proved both ways |
| B3 | each write site's comment names the retention window | grep for the window string at every §2.1 site |

## 7. If independent grading is wanted

Symptom for `090`, phrased to earn a verdict: *`agent_error_log` carries ~19 deliberately-written
finding codes (grep `ErrorCode:` literals and their carrying constants under `platform/`); determine
which have any automated consumer — Go, live `agent_definitions` workflow, cron, or script — that
selects rows by that code, and reconcile against the retention rule embedded in migration 466.*
Point it at `agent_error_log`, `platform/orchestration/agenterrors/`, and the four `FROM
agent_error_log` call sites in `platform/orchestration/actions/`. Assert no counts.

## 8. Traps for whoever picks this up

- **A zero over this table proves 30 days of nothing, never "never".** Retention truncates history;
  any all-history claim needs a source that outlives the window (`page_component_history`, git, the
  writers' own tests).
- **Grep the constant, not just the literal** (§3.2). The reader you miss is the one behind a const.
- **A resolved-count is not a readership test.** Zero resolved rows means the grader found nothing to
  heal, not that nothing grades them — a finding read and correctly left OPEN looks exactly like a
  finding nobody reads. This file's own author broke this rule an hour after filing it, on
  `CONTENT_DATA_REGRESSION`, having already run the correct test across sixteen codes (§2's second
  correction; `WRONG_CALLS.md` 2026-08-22). The test is always: grep the code, read the arm.
- **`resolved = true` is not an upgrade** — it halves the row's remaining life to 14 days. If B1
  builds a triage flow that marks rows resolved, it must extract what it needs first. **Live as of
  2026-08-22 10:40Z**: 48 rows now expire on the 14-day clock rather than the 30-day one. That is a
  deliberate, documented choice by `content-loss-check` (its durable records are the state census
  and the per-run heartbeat, not the finding rows) — but any *future* consumer must make the same
  choice knowingly, because the row it resolves is the row it shortens.
- **`error_code` is free text.** Uppercase and lowercase families coexist, and one writer emits
  colon-suffixed variants (`tool_crosslink_not_emitted:tool_page_will_not_go_live`). Any registry
  or GROUP BY must decide normalisation explicitly or it will double-count a family as compliance.
