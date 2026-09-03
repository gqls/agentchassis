# HANDOFF — experience_loop, continue here (2026-09-03)

**Full path of this file:**
`docs/agent_docs/docs024_key_docs_latest/experience_loop/HANDOFF_2026-09-03_experience_loop_continue_here.md`

**This file SUPERSEDES the two 09-02 handoffs and is self-contained.** Read it alone. The older
files stay for the trail, not for instruction:
`HANDOFF_2026-09-02_content_quality_auditor_into_the_new_build_path.md` (the original task, plus
the GATE RESULT block explaining why it turned out to be three jobs) and
`HANDOFF_2026-09-02b_after_694_…md` (superseded — **its `orchestration_states` checks have since
expired, see §4**).

---

## 1. Where this stands in one paragraph

The owner asked for `content-quality-auditor` to be put into the new build path. Reading what it
actually produced first — which the previous handoff required — showed it could see **four
hardcoded page names, 7.7% of fleet pages**, and none of the pages he had complained about. So the
job became three: **(1) give it sight — DONE, live, council-approved, verified across two chassis
rolls; (2) teach it the promise-keeping questions — DONE, and they demonstrably fire; (3) route it
into the build path — STILL OWED**, deliberately held. Two side-findings became work of their own:
my own detectors are blind to an empty listing page, and one of them reports a false control
failure on every scoped run.

**Chassis is `v1.0.1356`** (pods up 2026-09-03 08:57:46/08:58:07Z). Verified against it — see §2.

---

## 2. VERIFICATION STATUS against the current build — run 2026-09-03 09:27Z

| check | result |
|---|---|
| 694 markers on the live row | **PASS** — allow-list gone, non-greedy strip, `ORDER BY pc.position`, all four dimensions, `filing_mode` still `record` |
| behaviour, 18h to the roll | **PASS** — 36 auditor LLM calls, **0 failures**, avg **~7,000** input tokens (range 3,612–11,129) vs a pre-694 average of **1,744** |
| behaviour since `v1.0.1356` | **not yet observable** — 0 auditor calls in 30 min, but **45 LLM calls across 5 agent types** (fleet is live, control passes) and the sweep last ticked **09:24:43**, 3.5 min before the check. It selects one site per 900s and only some need a content audit. **Not a failure — re-check in an hour.** |
| boxingonline post-694 | **UNVERIFIABLE by the old method, see §4** |

**These checks are PER-ROLL, not once.** 694 is DB config so a roll has no mechanism to revert it,
but that is an argument, not a check. Note the seat's `updated_at` moves at roughly every roll
(14:36 mine → 15:38 → 20:55:58 → **08:56:53**), each time with **no snapshot taken**, and every
marker has survived every time. Cheap to re-run; do it.

```sql
-- (a) the markers
SELECT (default_config->'workflow'->'steps'->'load_page_content'->'config'->>'query' LIKE '%p.name IN (%') AS allowlist_back,   -- f
       (default_config->'workflow'->'steps'->'load_page_content'->'config'->>'query' LIKE '%<style[^>]*?>%')  AS nongreedy,     -- t
       (default_config->'workflow'->'steps'->'load_page_content'->'config'->>'query' LIKE '%ORDER BY pc.position%') AS ordered, -- t
       (default_config->'workflow'->'steps'->'run_content_llm_audit'->'config'->>'prompt' LIKE '%9. GUIDE PROMINENCE:%') AS dims, -- t
       (default_config->'workflow'->'steps'->'write_findings'->'config'->>'filing_mode') AS filing_mode,                        -- record
       updated_at
FROM agent_definitions WHERE type='content-quality-auditor' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- (b) the behaviour, from the DURABLE source (see §4 — do NOT use orchestration_states)
SELECT count(*) AS calls, round(avg(input_tokens)) AS avg_in, count(*) FILTER (WHERE NOT success) AS failed
FROM llm_call_log WHERE agent_type='content-quality-auditor' AND created_at > '<the roll time>';
-- pre-694 avg was 1,744. Post-694 it runs ~7,000. A drop back to ~1,700 means 694 is gone.

-- (c) ALWAYS pair (b) with a demand control, or a quiet fleet reads as a broken seat
SELECT count(*), count(DISTINCT agent_type) FROM llm_call_log WHERE created_at > '<the roll time>';
```

---

## 3. What is DONE

### Migration 694 — `content-quality-auditor` can see the site

| | |
|---|---|
| files | `docs/agent_docs/sql_for_agents/694_content_quality_auditor_can_see_the_site.sql` (+ `_ROLLBACK`) |
| council | **APPROVED, round 2**, `SUBMISSION_CORR=d52a0e45-5c64-4d32-a1ab-f73532684d37` |
| commits | `5ff171327` r1 · `0fa679a28` r1-REVISE actioned · `2967111d5` post-approval advisory fixes |
| applied | **2026-09-02 14:36:08Z** by hand, then `run-migrations.sh --record-only` so no other session's `--apply` re-runs it |
| owner ruling carried | `filing_mode` stays **`record`** — verdict rows for human approval, nothing auto-regenerates |

**What changed.** `load_page_content` hardcoded `p.name IN ('index','about','services','contact')`
— 3 of 22 pages on boxingonline, **92 of 1,196 fleet-wide (7.7%)** across 36 sites, avg 2.56 pages
seen per site against an avg site size of 33.2. It now samples **every `page_type`, 4 per type,
1,200 chars each**, ordered by `pc.position`, with `<style>`/`<script>` stripped **before**
truncation, exposing `page_type` and `url`. Three compounding defects fixed at once: index pages
average 28,180 chars so the old 1,000-char cap sampled **4.5%**; `rendered_html` is **42.8% CSS**
fleet-wide and on boxingonline's index `<style>` started at **character 1**, so 999 of 1,000 chars
reaching the model were stylesheet; and `string_agg` had **no `ORDER BY`**, so the window drifted
across runs on a byte-identical page. Prompt gained `audience` to its enum (10 of 210 stored
findings already emitted it out-of-vocabulary), TOP 5 → TOP 8, and four dimensions:
**6 PROMISE · 7 EMPTY INDEX · 8 TOOL DATA · 9 GUIDE PROMINENCE**.

### The dimensions are not decoration — they fire, and they reproduce the owner's own complaints

Unprompted, on **gamesdesign.co.uk** (21 pages / 6 page_types / 9,482 input tokens), a site nobody
aimed them at:

- **[HIGH] promise** — `/guides/index.html` hero CTA reads *"Launch Cooldown & Resource Cost
  Analyser"* and links to `/games/auto-battler/index.html`. Names one tool, opens another.
- **[HIGH] promise** — `/guides/economy-basics/` promotes the Sink & Faucet Modeller in body copy;
  both CTAs go to two *other* tools.
- **[HIGH] dimension 8** — *"asks the reader to supply a Host ID obtained from another user in a
  separate session … A tool that requires external coordination to produce any output is a form
  stub, not a working tool."*
- **[MED] dimension 9** — *"The explainer leads the usable thing it explains, inverting the
  priority the site's own value proposition demands."*

The last two are the owner's boxingonline complaints — *"the comparison tool requires the user to
input all the details"* and *"the guide is more prominent than the tool"* — **restated
independently by the seat on a different site.** That is the capability he asked for, working.

---

## 4. ⚠ THE TRAP THAT WILL BITE YOU FIRST — `orchestration_states` is a 24-HOUR ROLLING WINDOW

`[MEASURED 2026-09-03 08:53Z]` oldest row **2026-09-02 08:35**, newest now, window
**1 day 00:17**, 9,312 rows.

The 09-02b handoff told you to verify by counting `page_samples` rows in `orchestration_states`.
**That check has already expired.** Boxing Online's pre-694 audits (06:23Z on 09-02, 3 pages) that
it quotes as the baseline are **gone**: the same query that returned 3 rows on the evening of 09-02
returns **0** the next morning.

**A zero there means "reaped", not "never audited", and the two are indistinguishable.** Do not
build any durable claim on that table.

**Use `llm_call_log`** — it is the estate's training corpus, not reaped, rows back to April, and it
carries 694's effect as an input-token step change (1,744 → ~7,000) that no reaper can erase.

**Consequence for the one loose end:** whether boxingonline specifically has been audited post-694
can now only be caught *inside* a 24-hour window. Either catch it live, or accept the fleet-level
evidence in §2/§3 and judge the seat on a site you can observe. **Do not go hunting for the 06:23
row — it no longer exists.**

---

## 5. WHAT IS LEFT — in the order I would do it

### 5a. The empty-index rule — funded by a second instance, and PROMISED to a peer

`designblog.co.uk` serves four listing pages with **zero** items, each carrying prose about its own
brief instead: `/glossary.html`, `/inspiration/`, `/the-design-feed/`, `/uk-studios-directory/`.
Owner's words via that lane: *"the glossary has text about the brief and is not a glossary… the
directory is empty."* **Both my detectors report ZERO on it, run by hand.**

Rule C's corpus, `scripts/audit-experience-promises.py` (`INDEX_SQL`, ~line 253):
```sql
AND jsonb_typeof(COALESCE(pc.content_data->'articles', pc.content_data->'items')) = 'array'
), sized AS MATERIALIZED (SELECT * FROM inst WHERE jsonb_array_length(arr) > 0)
```
Two independent filters drop an empty index **before the rule runs**. Measured: all **11** component
rows under those four pages have `arr_type` **NULL** — no `articles`/`items` key at all — so they
die at the FIRST filter, never reaching the `> 0` one. **Rule C asks "does this index list things
from OUTSIDE its own directory?", which presupposes it lists something. An index that lists NOTHING
is invisible to it.**

Constraints you must not lose:
- **`/glossary.html` is `page_type='content'`** — outside Rule C's selector entirely. The fix needs
  a **content-class trigger, not a page_type widening**. This is what makes it more than one line.
- **It will fire on pages that CANNOT fill themselves, and that is correct.** Per `bugs_open/444`:
  the four remakes' feeds have **0 `content_sources` rows**, glossary and inspiration have **no item
  producer anywhere in the estate**, and the directory kind does not exist. 444's fix candidate —
  plan validation refusing a listing page whose item source resolves to zero — is the upstream
  door-closer this detector then holds shut. **Say so in the docstring** so nobody "fixes" the
  detector for over-reporting.
- This is the CONTRIB's open **ask 2 listing half**; boxingonline then designblog is what funded it.

> ## ✅ 5b AND 5b-bis ARE DONE AND LIVE — 2026-09-03 10:55Z (commits `ce8d5096a`, `e535fc4f0`)
> Both shipped and verified in-cluster; the sections below are kept for the reasoning, not as work.
> - **SQ-004 control:** an out-of-scope control now reports `n/a (control case not in --site scope)`
>   instead of FAIL. The NEGATIVE control got the same treatment — out of scope it passed
>   **vacuously**. Fleet run unchanged (187 instances / 2 mismatches / both PASS).
> - **SQ-005 rule B:** `repair_not_served` bucket added; such pages no longer count as `ok`.
>   **Narrowed twice by ground truth** — the first cut flagged **38** and 5 of 5 random curls from
>   other sites were serving fine. Discriminators: `build_status='needs_rebuild'` is honestly
>   labelled (13–44 day gaps, previous copy serves), and sub-second gaps are same-transaction noise
>   (0.047s). Narrowed to `build_status='deployed' AND gap > 1 minute` → **38 → 11**.
> - **It ships as a TRIAGE list stating its own precision (7/11)**, because ground-truthing the 11
>   showed the four non-seotools survivors all serve fine and **no DB column separates them** —
>   the false positives' gaps (2h17m–21h33m) *bracket* the true ones (9h25m–9h45m). Only served
>   bytes can tell. The receipt says so, and says curl before acting.
> - **Live:** ConfigMap `experience-promise-check-script-4t95f4hmm7`, `cronjob configured`,
>   triggered job **exitCode=0**. Fleet: 354 tool pages, **338 fine**, rule B 1, rule C 1,
>   never-built 4, triage 11.
> - **Remaining debt from this:** the vetcomparison promise-ledger pass of 09-02 still needs
>   re-checking against served bytes on the post-701 re-run (§6).
> - **The lesson, if you build a detector arm:** the first cut "worked" on the motivating case —
>   the one input guaranteed not to disconfirm it. **Sample your ground truth away from the case
>   that inspired the rule.** Five random curls cost a minute and changed what shipped.

### 5b. SQ-004's `--site` control bug — cheap, and it poisons every scoped run until fixed

`scripts/audit-listing-class-promise.py --site <anything-but-leopardess>` prints
`positive_leopardess: FAIL — classifier drifted or site changed`. **Not a regression, not a
finding.** The positive control is a leopardess page; `--site` filters the control's own case out of
the corpus, so it reports FAIL where it means **N/A**. Proof: `--site leopardessconsulting.co.uk`
→ control **PASS**, still finding the real mismatch (7/13 off-class on `/blog.html`).

It matters because that line is the only one telling a reader whether a zero is trustworthy, and it
currently cries wolf on every scoped run. Fix: report `N/A (control case not in scope)`, or exempt
the control from the site filter. **Until then every `--site` clean result is UNTESTED, not clean**
— I said exactly that to two peer lanes, so the caveat is on their record too.

### 5b-bis. ⚠ RULE B IS FALSE-CLEAN WHEN A REPAIR IS STORED BUT NOT SERVED — added 2026-09-03 10:15Z, do this with 5b

**Found live, on the case the owner was complaining about.** SQ-005 fired on schedule for the
first time ever at **07:40 on 2026-09-03** (its `lastScheduleTime` had been empty — CronJob created
09-02 10:30Z, after its own slot) and rule B correctly reported **8** findings, **seven of them
seotools.co.uk tool pages** including `/tools/serp-snippet-previewer/index.html`, ~2.5h before the
owner's critique arrived.

Then somebody repaired those seven in the database between **09:34 and 09:54**, and **none of the
repairs is being served**. Re-running rule B at 10:09Z returned **0 findings, "14 interactive"** —
a **FALSE CLEAN** that I nearly relayed as good news.

`[MEASURED 2026-09-03 ~10:1xZ]` all 14 seotools tool pages:

| component last written | stored has control | SERVED control count |
|---|---|---|
| 09-02 22:34–23:00 (7 pages) | yes | 2, 11, 14, 17, 4, 1, 2 |
| **09-03 09:34–09:54 (7 pages — the flagged ones)** | yes | **0 × 7** |

```
/tools/serp-snippet-previewer/  build_status=deployed  deployed_at=09-03 00:08:16  newest_component=09:43:18  -> component NEWER
/tools/seo-schema/              build_status=deployed  deployed_at=09-03 00:19:57  newest_component=09-02 22:48 -> deploy after build, serving fine
```

**`build_status='deployed'` records that a deploy once happened, not that the CURRENT components
have been deployed.** Rule B reads STORED `rendered_html`, so it goes quiet exactly when stored and
served diverge — i.e. **it certifies a page at the moment that page is most misleading**: the fix
is in, everyone believes it, the visitor still gets a description page.

**The fix, one join, no HTTP:** refuse to report a tool page clean when
`max(page_components.updated_at) > pages.deployed_at` — report **"unknown — repair not yet served"**
instead of clean. Comparing served bytes is better still, but the predicate is free and catches the
same class. **Do this at the same time as 5b**; both are "my clean results are not trustworthy" bugs
and they should land together.

**Debt it creates:** every clean rule B result already given to another lane is valid only where
stored and served agree — including **the vetcomparison promise-ledger pass of 09-02**, which must
be re-checked against served bytes on the post-701 re-run and corrected to that lane if it moves.
Both peer lanes have been told.

### 5c. The ORIGINAL task: route the seat into the new-build path

Deliberately held until a post-694 audit is observed on a real build (§2). Then:

- Edit **`site-work-orchestrator`**. Live tail was
  `fix_items_loop -> apply_site_design -> update_site_status -> complete`; **insert between
  `apply_site_design` and `update_site_status`** — site fully built and designed, nothing marked
  delivered.
- **Use `call_agent`, not `spawn_agent`** — the spawn→call handshake fails ~half the time
  fleet-wide, and `apply_site_design` in that same workflow is already a `call_agent` to copy
  verbatim (`action`, `config.agent_type`, `config.input_mapping`, `timeout_seconds: 300`,
  `output_field`).
- The auditor needs **no storage client** (DB queries + one LLM call), so the inline-chassis
  `params.StorageClient` nil landmine does not bite. Say so — a reviewer will ask.
- **Not redundant with `content-reviewer`**: that is per-PAGE inside `build_items_loop`; this is a
  GROUP/site auditor. Boxingonline is the argument — every page passed individually and the SITE
  was wrong.
- Council gate applies (a migration under `sql_for_agents/` is in scope). `DRY_RUN=1` tests
  admission free; `Council-Submitted:` lets you commit before the verdict.

### 5d. CONTRIB ask 3, planner half — still open
Refusing to SELECT a tool at planning time when we hold no data to put in it. The mechanical half
shipped as SQ-005 rule B; this is the planner-side door-closer.

---

## 6. Live promises to other sessions — these are commitments, not notes

| to | owed | state |
|---|---|---|
| `designblog.co.uk` | build the empty-index rule, then **re-run over their four pages and report** | outstanding (§5a) |
| `vetcomparison` | **re-run both detectors after migration 701 applies**, and fire the widened auditor (record mode) at vetcomparison post-701 | they will ping when 701's remainder batch lands |

Both lanes hold my two caveats (leopardess-control N/A-as-FAIL; empty-index blindness) so their
quoted zeros carry the right weight. Do not let those go stale.

**Findings already given to vetcomparison:** promise ledger clean; one finding —
`tool-compliance-deadline-calculator` is `status='active'`, `build_status='planned'`, **0
components**, serves 404, **`in_header=false` with 0 references on the served homepage**, yet
carries `nav_label='Deadline Calculator'`. Half-wired, not a live dead link — it becomes one the
moment anyone turns that nav on. Their lane says it is standing owner item `e30dc7b9` (07-17),
recommendation BUILD with relative periods only. They also correct the vigilant-designer's claim
that the parked contrast items concern an unrendered colour — **the three grey ones do render**,
inside the tool markup.

**Owner design directive** (nav/hero sameness; *"We are trying to make these sites best in class"*)
is **NOT this lane's work** — routed to and ACKed by `components`, `theme kits`,
`site design planner`, `vigilant-designer`, `editorial design uplift`. This lane is
**context-holder only**; do not silently adopt it.

---

## 7. Traps — read before touching anything here

- **`orchestration_states` is a 24h rolling window** — §4. The single most likely way to get a
  wrong answer in this area.
- **Count auditor runs by `collected_data ? 'content_audit'`**, never by text-matching the agent
  name in `workflow_plan` — that matches the PARENT's spawn record.
- **`content-quality-auditor`'s `default_config.workflow.steps` is an OBJECT keyed by step name,
  not an array.** `jsonb_path_query_array(…,'$.workflow.steps[*] ? (…)')` returns `[]` and reads as
  "no such step". Use `jsonb_object_keys` or `$.**`.
- **PostgreSQL takes the greediness of the FIRST quantifier.** `'<style[^>]*>.*?</style>'` is
  GREEDY as a whole and deletes prose between two style blocks — measured, 7 of 2,871 components
  affected, avg 2,528 chars destroyed, worst 9,076. Use migration 601's
  `'<style[^>]*?>.*?</style>'` with `'gi'` and a **space** replacement (`''` joins the words either
  side). **This is PostgreSQL-only — it does NOT transfer to the Go call sites**, RE2 is
  per-quantifier lazy (proven by running it). Landmine filed.
- **A migration verify of bare `IF position(...) = 0` FAILS OPEN on NULL** — `position(x in NULL)`
  is NULL, `NULL = 0` is NULL, plpgsql treats `IF NULL` as false. 694 refuses on NULL first; copy
  that shape. And **EXECUTE the SQL your migration embeds** — substring checks prove the text was
  written, never that it parses.
- **The prompt's category enum is TEXT NO GO CODE READS** (`grep -rn 'tone|gap|cta' --include=*.go`
  → nothing). `audience` matches no router rule and lands in the `capability_gap` fallback either
  way. The older belief that the fallback mints `item_type='audit_finding_'+category` is **stale**
  — removed 2026-08-15 (`write_audit_findings_action.go:724`).
- **The record-mode reader already exists; do not build one.** Admin dashboard `recordVerdictsOnly`
  filter (`frontends/admin-dashboard/src/App.tsx:474` → `GET /admin/work-items?filing_mode=record`),
  per-finding Release button (`App.tsx:737` → `POST /admin/work-items/:item_id/release`), route
  `internal/core-manager/api/server.go:455`, handler `site_admin_handlers.go:1049`.
- **`experience-promise-check` had NEVER fired on schedule** as of 09-02 19:15Z (`lastScheduleTime`
  empty — CronJob created 10:30:27Z, after its own `40 7 * * *` slot). **Check `lastScheduleTime`
  before reading a clean cron history as coverage.**
- **The seat's lifetime 65% LLM failure rate is NOT a code defect and is closed** — 5,013 of 5,114
  failures are API spending-cap 400s spanning April → 2026-08-31. Since 09-01 it has run clean
  (0 failures in 36 post-694 calls). This answers the CONTRIB's parting "34% failure rate" question.
- **Migration numbers collide.** I took 692, found it claimed mid-work, landed on 694; 691 was
  double-claimed by two lanes the same day. `ls docs/agent_docs/sql_for_agents/` immediately before
  writing the file.
- **`scratchpad/fire_cqa.sh` IS NOT A WORKING DISPATCHER** — see §8.

---

## 8. Two of my own missteps, kept because the check is the useful part

**A publish receipt proves the message LEFT, and nothing about anything consuming it.** I fired a
hand-rolled dispatcher at boxingonline twice (`30b5a2c5`, `34cb071c`). Both produced **zero
`orchestration_states` rows AND zero `llm_call_log` rows** — nothing ran, twice — while
`kafka_publish_checked` printed a genuine PUBLISHED receipt both times. I then misread it three
ways: latency, then "generic-topic runs are stateless so no row persists", then **"the audit ran"**
on the strength of two `content-quality-auditor` LLM calls in the same time window. Those calls
carried correlation `0f0dc48e` — **the sweep's run on gamesdesign.co.uk, not mine.**
**The one query that kills all three errors is to join on YOUR correlation, never a time window:**
```sql
SELECT count(*) FROM llm_call_log        WHERE correlation_id = '<your corr>';
SELECT count(*) FROM orchestration_states WHERE correlation_id = '<your corr>';
```
A busy fleet always has something in a time window.

**An apply advances "live"; it does not advance "committed".** The two safety fixes I made in
response to the council's approving advisories landed *after* my commit and *before* the apply, and
I committed nothing in between — so for ~7 hours the live database ran a migration git did not
have. Nothing was at risk (the rollback file was unaffected) but the audit trail was wrong. The
check costs nothing: **`git status <the file>` immediately after applying**, not at session end.

---

## 9. Commands

```bash
# detectors (NB the SQ-004 control reads FAIL when scoped away — §5b)
python3 scripts/audit-experience-promises.py --site <domain>
python3 scripts/audit-listing-class-promise.py --site <domain>
python3 scripts/audit-experience-promises.py --self-test        # fixtures, no cluster

# migration: dry-run WITHOUT persisting (how 694 was proven)
sed -e "s/^SELECT snapshot_agent(.*$/-- dry/" -e "s/^COMMIT;/ROLLBACK;/" <file>.sql \
  | kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1

# council: free admission test, submit, resubmit on the SAME correlation
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh sub.json
RESUBMIT_CORR=<corr> ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh sub.json

# out-of-band apply, then tell the ledger so nobody re-runs it
./scripts/migration/run-migrations.sh --record-only <file>.sql --note "<why>"
```

**Council lesson worth more than the commands: reviewers judge the SKETCH — it is the only view of
your code they get.** Three of round 1's objections were about a `CREATE TABLE` and a
`snapshot_agent()` call that were *always in the file* and that I had trimmed out of the sketch.
Round 2 sent the full executable bodies and those objections vanished. **Send the whole body.**
And a REVISE is cheaper than the defect it finds: round 1's HIGH objection found a real,
silent, content-destroying bug in my own fix.

---

## 10. The five standing docs

| doc | path |
|---|---|
| PLAN | `…/experience_loop/PLAN_experience_loop.md` |
| RUNBOOK | `…/experience_loop/RUNBOOK_experience_loop.md` |
| NOTES | `…/experience_loop/RUNNING_NOTES_experience_loop.md` — four entries dated 09-02/09-03 |
| README (owner's) | `…/experience_loop/README_where_we_are.md` — **append only, never rewrite; it is the owner's document** |
| SUMMARY | `…/experience_loop/SUMMARY_where_experience_loop_is_now.md` — **not rewritten this session, deliberately**: the milestone is the routing going live, and the five headings would repeat the last one. Write a NEW dated file when 5c lands. |

Register: **SQ-004** `listing-class-promise-check` (`25 7 * * *`) · **SQ-005**
`experience-promise-check` (`40 7 * * *`).

**Refused deliberately, do not re-attempt as a regex:** *"does this page contain the thing its title
asserts?"* — the disproof is in SQ-005's docstring. The empty-index rule (§5a) is decidable because
it counts items; the title-assertion one is not.
