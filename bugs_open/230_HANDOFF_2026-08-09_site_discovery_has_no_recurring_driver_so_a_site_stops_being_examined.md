# 230 — site discovery has NO recurring driver: every scheduled entry is a disabled one-off, so a site stops being examined the moment a human stops looking at it

**Filed 2026-08-09 by the `bugfix_201_page_content_writer_dispatch` lane. OPEN.**

> **TAKEN UP 2026-08-09 by the `bugfix_230_discovery_driver` thread** (standing five in
> `docs024_key_docs_latest/bugfix_230_discovery_driver/`). Every §2 figure re-verified
> live the same day — all hold; §6's worked case still undetected. Fix built on
> candidate 1 made fair (a `site_discovery_rotation` stamp table + three rotation
> tasks, observe-only, migration `346`) plus candidate 2 as a daily watchdog CronJob
> (`site-discovery-staleness-check`). §4's cost question was measured rather than
> deferred: one full discovery cycle = 2 LLM calls (~4.2k in / ~2.1k out tokens,
> dartsonline 08-09), steady state ≈ 9 site-examinations/day. Two findings §3 could
> not have seen: the designed driver was paused *on record* (register IMP-016,
> "intentionally paused during core build" — answering §4's first bullet: deliberate,
> but for a build phase that has moved on), and it could not simply be re-enabled
> anyway — its live pre_query's <50-open-items cap excludes webdesign.co.uk (85) and
> dartsonline.com (79), the two most-worked sites, and its `ORDER BY sites.updated_at`
> starves (IMP-010). Council `Council-Submitted: 2281fc48-f0c5-4842-88c7-8391d0098944`.
> The `bugs_open/083` drain question is deliberately untouched — this closes the
> *detection* half only. Inert until migration 346 is applied (post-verdict).

Found while chasing `bugs_open/201`'s handoff §4 — two live pages still serving empty
sections with no work item raised. The empty pages turned out to be the *symptom*; this is
the mechanism, and it is fleet-wide rather than site-specific.

---

## 1. The claim

`empty_sections` and its sibling discovery checks run **only when something dispatches them
at a named site**. There is no timer, no cron, and no enabled schedule. A site is therefore
examined exactly as often as a human happens to point a trigger at it — so a genuinely
broken page on a site nobody is currently working on is never re-detected, indefinitely.

## 2. Evidence (all first-hand, 2026-08-08/09, live `clients_db`)

**Every scheduled row targeting a site-discovery agent is a disabled one-off.** `[MEASURED]`

```sql
SELECT name, target_agent_type, enabled, last_triggered_at
FROM scheduled_tasks
WHERE target_agent_type IN ('quality-discovery-agent',
                            'completeness-discovery-agent','design-discovery-agent');
```

| name | target | enabled |
|---|---|---|
| `oneshot-completeness-discovery-fai-20260803` | completeness-discovery-agent | **f** |
| `oneshot-completeness-discovery-rh-20260730` | completeness-discovery-agent | **f** |
| `oneshot-design-discovery-rh-20260730` | design-discovery-agent | **f** |
| `oneshot-discovery-aao-20260726` | completeness-discovery-agent | **f** |
| `oneshot-quality-discovery-rh-20260730` | quality-discovery-agent | **f** |

With the negative control in the same query — `count(*) FILTER (WHERE enabled)` = **0** of
**5**. All five are hand-made `oneshot-*` rows, fired once and left disabled.

**No CronJob fires them either.** `kubectl get cronjobs -A` returns eight, all
cleanup/backup/drift-check (`agent-job-cleanup`, `database-backup`,
`component-fallback-check`, `component-render-check`, `concept-register-drift-check`,
`single-owner-carriers-check`, `shared-output-fields-check`, `bugs-open-staleness-sweep`).
None runs `run_discovery_checks`. `[MEASURED]`

**⚠ The obvious query gives the wrong answer.** Filtering `scheduled_tasks` on
`name ILIKE '%discovery%' OR target_agent_type ILIKE '%discovery%'` returns rows that ARE
enabled — `adoption-tracker-discovery`, `model-directory-discovery`, `protocol-tracker-discovery`.
Those target `adoption-researcher` / `directory-researcher`: the **model-directory research**
agents, a different subsystem that has nothing to do with site checks. Filter on
`target_agent_type IN (…)`, never on the word "discovery".

**Detection demonstrably follows attention, not a clock.** `[MEASURED]` Fleet-wide, by type:

| item_type | last filed |
|---|---|
| `hardcoded_section_colors` | 2026-08-08 18:11 |
| `literal_markdown` | 2026-08-08 18:09 |
| `voice_tells` | 2026-08-08 17:13 |
| `page_canonical_collision` | 2026-08-05 20:49 |
| `nav_drift` | 2026-08-05 14:10 |
| **`empty_section`** | **2026-08-04 19:36** |
| `truncated_component` | 2026-08-03 21:04 |

The three that filed on 08-08 landed on `webdesign.co.uk` and `leopardessconsulting.co.uk` —
precisely the two sites other lanes were hand-driving that day.

**The worked case, and it is a live customer site.** `finetuning.uk`
(`1368e337-dd1d-4799-bbb3-8221a1b79bcc`, `sites.status='deployed'`), pages `ai-guides`
(`69a50d5d-…`) and `insights` (`8867b4d5-…`), slot **`featured-content`**, component
`b3e0c2c0-…`, **334 bytes each**. Running `findEmptySections`' WHERE clause verbatim against
the live DB returns both pages today — `build_status='deployed'`, `locked_at IS NULL`,
`page_type='content'` (not `blog-index`), slot not in `suppressed_sections`. So the detector
is **not** blinded: nothing has run it. Every `empty_section` item this site has ever had was
created in two batches — seven sharing the timestamp `2026-08-03 10:15:22`, and an April
batch. `[MEASURED]`

## 3. Why this was not visible before

The two pages were "detected" on 08-03, under the slot name **`featured-article`**, and
those items were stamped `complete` the same afternoon while the pages stayed empty — the
handler satisfied them by *replacing* the component into a differently-named slot, the
verifier could not find its target, and (pre-`RFC_017`) errored **fail-OPEN**:

```
result->'_verification' = {"status":"error","item_type":"empty_section",
  "error":"cannot verify: component a390860e-… no longer exists
           (genuinely fixed or silently deleted — indistinguishable here)"}
status = complete
```

So the record said *fixed*, and no new detection has run to contradict it. Two separate
mechanisms produced one silence. (Those two rows are also, as far as the registry's history
shows, the only two occurrences of the absent-target case that `RFC_017` now fails closed
on — see `bugfix_201…/HANDOFF_2026-08-08` §3.) Full slot-rename trap is in `LANDMINES.md`.

## 4. What is NOT established

- ~~**Whether the absence of a driver is a defect or a deliberate cost decision.** Discovery
  runs cost LLM calls per site; someone may have disabled these on purpose. Nothing in the
  rows records *why* — they are just `enabled=false`. **This bug asserts the gap, not the
  remedy.** Whether discovery *should* be scheduled fleet-wide is an architecture question
  with real cost implications and it is not settled here.~~

  > **ANSWERED — OWNER RULING 2026-08-09:** *"The missing driver is probably a defect, I
  > haven't made any costs decisions lately."* So it is **a defect**, and there is no live
  > cost decision standing behind the disabled rows.
  >
  > **And the question was mis-framed by me, which matters more than the answer.** I posed it
  > as *defect vs deliberate COST decision*. The recorded rationale was never about cost:
  > `IMP-016` (found by the `bugfix_230_discovery_driver` thread) says the sweep was
  > *"intentionally paused during core build"* under an explicit operational policy that **a
  > discovery check should only be enabled once its handler agent actually exists — otherwise
  > findings accumulate unconsumed rather than clearing.** That is a **sequencing** gate, not
  > a budget one. So the pause was deliberate, for a reason that was never cost, taken for a
  > build phase that has since moved on, and never revisited. Both readings I offered were
  > wrong; the true one was in the register the whole time.
  >
  > **Consequence for the fix:** the real precondition is **handler-readiness**, not budget —
  > which is exactly why the owning thread's *observe-only* rotation is the right shape
  > (detection driven, triage still gated on `bugs_open/083`). Their cost measurement (2 LLM
  > calls per cycle, ≈9 site-examinations/day) stands and is useful for sizing, but it is
  > **not the gate** and nothing was ever waiting on it.
- **Whether the `oneshot-*` rows were ever meant to be permanent.** The naming suggests not.
- **How many other sites are in this state.** Only `finetuning.uk` was checked at the row
  level. The fleet-wide recency table above is suggestive, not a census.

## 5. Fix candidates, ordered by what closes the door

1. **A recurring `scheduled_tasks` row per discovery agent, with a `pre_query` that selects
   deployed sites on a rotation.** Config-only, live immediately, no roll. Closes the door
   properly: a site cannot fall out of examination by being un-interesting. ~~Needs a cost
   decision first (how often, how many sites per tick) — which is the architecture question
   in §4, so this is *not* a change to make unilaterally.~~
   > **CORRECTED 2026-08-09 (owner ruling, see §4): there was no cost decision to wait for,
   > and I should not have written that there was.** The real precondition is
   > **handler-readiness** (`IMP-016`): enable detection observe-only, keep triage gated. I
   > invented a budget gate from the *plausibility* that discovery costs money, and stated it
   > in the same voice as the measured facts in §2 — an `[ASSUMED]` wearing a finding's
   > clothes, in the one section headed "what is NOT established". If it delayed the owning
   > thread, that is the cost of it.
2. **Make the gap visible rather than closing it** — a check that reports sites whose last
   discovery run is older than N days. Cheaper, decides nothing, and turns an invisible
   silence into a number someone can act on. Good first move if (1) is contentious.
3. **Leave dispatch manual but make the trigger discoverable** — document it as the way to
   examine a site. Weakest: it still requires someone to think of it, and "operators must
   remember X" is the failure mode this platform keeps re-learning.

## 6. How to verify a fix

A site with no lane activity gets a fresh `empty_section` (or sibling) item without anyone
dispatching one. Concretely: `featured-content` on `finetuning.uk`'s two pages is an
outstanding, currently-undetected true positive — if a scheduled driver goes live, **an item
for that slot should appear on its own**. Until then its absence is this bug's live proof.

```sql
SELECT item_key, status, created_at FROM site_work_items
WHERE item_type='empty_section'
  AND page_id IN ('69a50d5d-3732-4efe-9a79-f887b072fa86',
                  '8867b4d5-12d1-4ecc-8956-109a80395a18')
ORDER BY created_at DESC;   -- today: nothing under 'featured-content'
```

## 7. Provenance of this filing — read before quoting it

**The `090` diagnosis loop WAS run and returned NO VERDICT.** Intake correlation
`c5778c3e-8cf9-41b3-b36f-6d1ad37b708a`, run correlation
`2ccc7551-76d3-40d2-ac2a-01d8120ea0fb`. The orchestration reached
`current_step='complete'`, `status='COMPLETED'`, the work item reached `status='complete'`,
and `diagnosis_artifacts` holds **5 `bundle` rows and no verdict artifact**; no `doc_notes`
row was written. `diagnose_route`'s `max_iterations` is **5**, so it exhausted its budget
while still re-scoping rather than converging.

**It is NOT the known ~60KB body-omission mode** — the final bundle is 72,310 bytes, carries
`findEmptySections` in full, and contains no truncation marker. The bundle *did* gather the
right evidence, including both live `featured-content` rows.

So, per the **owner ruling of 2026-07-31**, this file states plainly that it rests on
**first-hand verification** rather than on a loop verdict: every figure in §2 is a live query
recorded with its SQL, run by the filing session, with negative controls where one was
available. The loop was run as the ruling requires; it simply did not produce a verdict, and
a run that ends without one is **not** evidence for or against the claim.

## 8. Related

- `bugs_open/201` — the lane this came out of; its handoff §4 carried the (corrected) version
  of this finding, and `WRONG_CALLS.md` 2026-08-08 records why the original was wrong.
- `bugs_closed/032` — verifier reads a deleted target as a successful fix. Its own "stronger
  option" (ask whether the page still declares the slot) is the cheap fix for §3's half.
- `RFC_017` (`architecture_review/`) — the fail-closed flip; its option 3 (park instead of
  retry) is triggered by exactly the `empty_section` case in §3.
- `bugs_open/223` — the landmine-verifier's Go-only index. Unrelated mechanism, but it
  produced a false `STALE` on one of the two landmines filed alongside this bug.
- `LANDMINES.md` 2026-08-08 — the slot-rename entry, verified `STILL_VALID`.
