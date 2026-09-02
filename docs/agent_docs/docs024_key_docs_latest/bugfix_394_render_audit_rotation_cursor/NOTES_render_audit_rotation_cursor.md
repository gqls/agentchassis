# NOTES — render audit rotation cursor (bugs_open/394)

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-26 ~16:20 — LANE CLAIMED, before any research

**This file exists first and is committed first, deliberately.** I lost `bugs_open/359`
an hour ago to a session that opened the same lane two minutes after me: we both ran
`who-owns.py`, we both ran the tree grep, and both instruments told both of us the bug was
unowned — because at the moment either of us looked, the other had written nothing. The
window is not a gap in the instruments; it is the interval they cannot see into. Full
account in `WRONG_CALLS.md`, 2026-08-26.

So: **claim in the commit log before doing the work**, not after. If you are reading this
and you also want 394, `git log --oneline -- <this directory>` will tell you who was first,
which is the thing neither of us could establish last time.

### Ownership as at 2026-08-26 16:20

| instrument | result |
|---|---|
| `scripts/who-owns.py 394` | `likely OWNING workstream(s): (none identified)` |
| commits touching the bug file | **one**, ever: `3cb6be421` (2026-08-25), the filing commit |
| lane directory named `bugfix_394*` | none existed before this one |
| cross-references | `bugfix_358_unread_finding_codes` cites it twice — the lane that FILED it, not one working it |

Its two siblings, filed in the same commit, ARE owned and I am not touching them:
`bugs_open/392` → `bugfix_392_link_context_unread`; `bugs_open/393` →
`bugfix_393_ungraded_completions`. 394 is the one of the three nobody took.

### What the bug asks for, in one line

`bugs_closed/242` made render-audit truncation **loud** — `pages_total`/`truncated` stamped
into the durable result plus an `agent_error_log` `RENDER_AUDIT_TRUNCATED` row. Nothing
reads it. Meanwhile the mitigation it shipped (raise the cap 25→60) has been outgrown:
webdesign.co.uk went 109 → 125 → 131 live pages in six days, and the writer's own message
says the unaudited tail is **the same pages every run**.

The owner commissioned a reader (ruling 2026-08-25, decision 4). The bug ranks three
candidates and candidate 1 — **persist a per-site cursor so the next run starts where the
cap cut off** — is the one that retires the signal's cause rather than reporting it. That
is the direction I intend to take, subject to what the evidence says next.

Next: re-validate the four `RENDER_AUDIT_TRUNCATED` rows against the live DB, and resolve
the `[UNEXPLAINED]` `5 of 26` row by reading the dispatching config rather than assuming
the cap is 60 everywhere.

---

## 2026-08-26 — the bug is VALID, and materially worse than filed

### The mechanism, from the code rather than from the message

`platform/orchestration/actions/request_render_audit_action.go` selects the site's live pages
with

```sql
ORDER BY COALESCE(nav_order, 999), name
```

and then, in the scan loop:

```go
total++
if len(urls) >= maxPages {
    continue // keep counting so the truncation is reportable
}
```

So it takes a deterministic **prefix** and counts the rest. The row it writes says *"the
unaudited tail is the SAME pages every run"* — that is not an inference by the author, it is
true by construction, and reading the loop is what establishes it.

### `[MEASURED 2026-08-26]` webdesign.co.uk: 146 live pages, 60 audited, 86 never

The bug quotes 60 of 131 on 2026-08-24. Two days later:

| live pages | audited | tail |
|---|---|---|
| **146** | 60 | **86** |

Fifteen pages added in two days, all of them into the tail.

### The finding that decides the fix: the tail is a whole CLASS of page, not a random 86

`nav_order` on webdesign.co.uk's active pages, `[MEASURED 2026-08-26]` — and it is never NULL
on this site, so `COALESCE(nav_order,999)` never fires here:

| nav_order | pages |
|---|---|
| 0, 10, 20, 30, 40, 90 | 1 each (6 nav pages) |
| **100** | **94** (the tools, then alphabetical by `name`) |
| **200** | **48** (all named `tool-*-guide`) |
| 201 | 1 |

A cap of 60 therefore covers the 6 nav pages plus the first **54 tools alphabetically**, and
cuts between:

```
rn 60  nav_order 100  tool-head-architect     <- last page ever audited
rn 61  nav_order 100  tool-html-minifier      <- first page never audited
```

**Every one of the 45 remaining `tool-*-guide` pages at `nav_order` 200 is structurally
invisible to the render audit and always has been.** No cap below 98 reaches them. That is the
argument against the bug's candidate 3 (raise the cap) stated as a measurement rather than as a
preference: a constant cannot chase a site that adds 15 pages in two days, and the specific
constant we would have to pick to reach the guides today is one the site would outgrow next
week.

### Two callers, and the cap is PER-DISPATCH, not per-agent

`[MEASURED 2026-08-26]` from `agent_definitions`, live rows only:

| agent | step | `max_pages` |
|---|---|---|
| `render-audit-agent` | `audit` | **60** |
| `design-critique-agent` | `audit` | **8** |

`design-critique-agent` was seeded yesterday (`sql_for_agents/645_design_critique_agent.sql`)
and **is already truncating from birth** — two rows today for leopardessconsulting.co.uk at
`8 of 37`, 14:22Z and 15:10Z.

### The bug's `[UNEXPLAINED]` "5 of 26" row — RESOLVED

The bug flagged the 2026-08-11 loancalculator row (`5 of 26`) as unexplained and warned against
assuming the cap is uniform. It was right to. The row's own context settles it:

```
2026-08-11 18:08:54Z | render-audit-agent | step audit
{"max_pages": 5, "pages_total": 26, "pages_audited": 5}
```

So: **the standing agent, running with a per-dispatch override of 5.** Not a different agent,
not a config regression in `agent_definitions`. The originating orchestration row has since
aged out of `orchestration_states` (rolling window), so where the override came from cannot now
be recovered — recorded as unrecoverable rather than left open.

**The conclusion that matters for the fix:** `max_pages` is a per-dispatch value, so any design
that reasons from "the cap is 60" is wrong.

### `[MEASURED 2026-08-26]` fleet exposure, at both live caps

25 sites have more than 8 live pages.

- **At cap 60** exactly ONE site truncates — webdesign.co.uk, tail 86.
- **At cap 8** — the design-critique caller — **25 sites truncate**, tails from 4
  (noted.co.uk) to 138 (webdesign.co.uk).

So this is two problems wearing one message: the render-audit caller has a **deep tail on one
site**, and the design-critique caller has a **shallow tail on the whole fleet**. A fix that
only serves the first leaves the second, which is now the larger population.

### The driver, for the record

`scheduled_tasks.site-render-audit-rotation` → `render-audit-agent`, `interval_seconds=3600`,
enabled; its `pre_query` picks ONE site whose `site_discovery_rotation.last_selected_at` for
that agent is older than **3 days**, stamps it, returns it. So a site is audited at most every
three days, and at that rate webdesign's 86-page tail would take a fortnight to cover even with
a perfect cursor — worth knowing before promising coverage in a given window.

⚠ `site_discovery_rotation` (`site_id`, `agent_type`, `last_selected_at`; PK on the first two)
is written **only from that SQL**, never from Go. A cursor written by the action would be the
first Go writer of that table — a fact any reuse-versus-new-table decision has to face rather
than assume away.

---

## 2026-08-26 — the two answers that most change the fix's shape

### 1. The retraction path is ALREADY scoped to what was measured — so a cursor is safe there

This was the risk I most expected to have to design around: `write_render_audit_findings`
RETRACTS standing findings, and a cursor makes every run look at a *different* subset. If
retraction inferred "resolved" from "not seen in this run", a cursor would silently close every
finding on every page the cursor happened to skip — fleet-wide, on the first run.

It does not. `write_render_audit_findings_action.go:219-222`, in the payload struct's own words:

> `PagesAudited` names the pages the adapter **SUCCESSFULLY MEASURED**, and is the **entire
> scope of what this action may retract**. Absent on an old-shape reply, which is why
> retraction is inert rather than wrong against an un-rolled adapter: an empty audited set
> retracts nothing.

and at :731, `if len(payload.Summary.PagesAudited) == 0 { … }` degrades to no retraction. The
adapter populates it per successfully-navigated URL
(`internal/adapters/browserrunner/render_audit_action.go:422`).

**So the most dangerous interaction in this change is already closed, by someone else, for a
different reason.** Recording it as a checked fact rather than an assumption, because the
opposite would have been a fleet-wide silent close and it is exactly the kind of thing a plan
asserts without opening the file.

⚠ The corollary is a constraint, not a freedom: **the cursor must never advance past a page the
adapter did not actually measure**, or coverage and retraction scope drift apart — the audit
would claim a page as "covered" that was never rendered. `len(PagesAudited) < Pages` exactly
when a navigation failed, and the adapter's own comment at :232-235 says so.

### 2. There is NO durable per-page audited list today — only counts

The bug's acceptance for candidate 1 is *"the union of audited pages reaches all 146, verified
by the audit's own durable page list, not the status"*. That list does not exist to verify
against:

```
grep -rn "PagesAudited|pages_audited" --include=*.go platform/ internal/ | grep -v _test
```
`[MEASURED 2026-08-26]` every persisted use is a **count**:
`request_render_audit_action.go:179` writes `"pages_audited": len(urls)` into the
`agent_error_log` context; `write_render_audit_findings_action.go:615` writes
`result["pages_audited"] = payload.Summary.Pages` and :654
`result["retraction_scope_pages"] = len(payload.Summary.PagesAudited)`. The **URL slice itself
is used and discarded**.

And the step's own result is not durable anyway — the file's header records that an awaiting
step's result never survives the park (RFC_012 addendum 2, owner-ruled option B), which is why
the truncation row is written to `agent_error_log` *before* the dispatch in the first place.

**So the acceptance arm the bug asks for is not runnable today.** Whatever the fix is, it has to
make coverage reconstructable — and the cheapest honest version is probably to record the cursor
window (`cursor_from` / `cursor_to` / `pages_audited`) on the durable row that is already being
written, so consecutive runs can be replayed into a union. An acceptance test nobody can run is
the same as no acceptance test, so this is part of the change, not a follow-up.

### 3. The second caller is MANUAL, which changes what a cursor means for it

`sql_for_agents/645_design_critique_agent.sql` states it plainly: **"NO cadence — manual trigger
only, per the recorded owner decision. Nothing emits work for this agent; nothing promotes into
it; the only path is a hand dispatch."** So `design-critique-agent`'s cap of 8 is a *sample for
a vision critique*, not a coverage sweep — its 25-site truncation footprint is real but it is
not accumulating a debt on a schedule the way the render-audit rotation is.

That file also records the co-dedup design: this agent is deliberately a **second producer** on
`contrast_failure` / `undeployed_asset` keys beside the rotation, and the shared key namespace
is "the point". Worth holding onto — it means the two callers' findings already interleave by
design, so a per-caller cursor must not be justified by "keeping their findings apart". If the
callers need separate cursors it is because their *coverage questions* differ, which is a
different and weaker argument, and the plan should say which one it is relying on.

### 4. The optional-key budget has headroom — checked, because it gates the design

`scripts/audit-optional-key-budget.sh --json` `[MEASURED 2026-08-26]`:

```json
{ "action": "request_render_audit", "optional_keys": 6, "over_budget": false,
  "optional": ["site_id_field","domain_field","max_pages","page_names","topic","capture_renders"],
  "consumers": 2, "agents": ["design-critique-agent","render-audit-agent"] }
```

Budget is **N = 10** (owner ruling 2026-08-14, register WFA-013), and **no action in the fleet
is over budget today** (0 of the audited set). So this action can take one or two new optional
keys without tripping the accumulated-surface review. Checked rather than assumed, because the
budget check is exactly the sort of thing a plan asserts in passing and gets wrong — and because
RFC_022's narrowing only exempts an opt-in field from *architecture* scope; it does not exempt it
from the count.

---

## 2026-08-26 — the plan's one real tension, and the measurement that dissolves it

The drafted plan (candidate 1, a keyset coverage cursor) surfaced a risk I did not expect and
which I think would have been wrong to accept as stated.

### The tension: a cursor re-creates the exact latency the owner ordered removed

`sql_for_agents/469_render_audit_rotation_three_day_window.sql` is an **owner instruction of
2026-08-18**, and its stated why is not "audits should be fresher":

> **WHY.** The render audit is the only thing that **GRADES a contrast repair**: it re-measures
> the page in a browser and withdraws the work item if the defect has gone (`bugs_open/296` §9 —
> 40 of the 226 parked rows drained this way on the first pass). Its eligibility window is
> therefore **the confirmation latency of the whole repair loop**: a fix that shipped today could
> wait up to **SEVEN DAYS** to be graded. That is the bottleneck, NOT the hourly tick.

A plain cursor makes each page's re-measurement latency `cadence × ceil(total / cap)` — on
webdesign, 3 days × ceil(146/60) = **9 days**. That is not merely slower than the 3 days 469
bought; it is **worse than the 7 days the owner issued an instruction to eliminate.** Shipping
that and recording it as an accepted cost would be trading away a thing the owner had already
ruled on, in a commit whose stated purpose is improving coverage.

### The measurement: the protected population is 3 pages, not 146

469's latency argument is about the pages that are **awaiting a grade**, not about the site. So
the question is how many pages those are.

```sql
SELECT s.domain, count(*) AS open_items, count(DISTINCT wi.page_id) AS distinct_pages
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE wi.item_type = 'contrast_failure'
  AND wi.status NOT IN ('complete','cancelled','rejected')
  AND s.domain = 'webdesign.co.uk'
GROUP BY 1;
```

`[MEASURED 2026-08-26]` → **webdesign.co.uk | 3 open items | 3 distinct pages.**

Three, out of 146. And fleet-wide the largest such set is robot-hands.com at 17 distinct
pages — on a site with 37 live pages, which never truncates at cap 60 and therefore never gets
a cursor at all.

**So the window should be a union, not a slice:**

```
window = (pages with an open regrade-awaiting finding)  ∪  (next N−k pages from the cursor)
```

The grading population rides in every run — latency stays at the 3-day cadence 469 paid for —
and the cursor covers everything else. Cost today: **3 of 60 slots, 5%.** The tension is
dissolved rather than traded, which is the better answer whenever it is available.

### And the priority set is `contrast_failure` ONLY — measured, not assumed

The obvious mistake here is to include `undeployed_asset`, since it shares the drain and the
key namespace. It cannot be page-targeted:

```sql
SELECT item_type, count(*) AS items, count(page_id) AS with_page_id
FROM site_work_items
WHERE item_type IN ('contrast_failure','undeployed_asset')
  AND status NOT IN ('complete','cancelled','rejected')
GROUP BY 1;
```

`[MEASURED 2026-08-26]`

| item_type | items | with page_id |
|---|---|---|
| `contrast_failure` | 111 | **111** |
| `undeployed_asset` | 190 | **0** |

Every open `undeployed_asset` row has a NULL `page_id` — it keys on the asset, not the page. So
there is no page to prioritise, and including the type would have produced an empty join that
looks like a working feature. Recording the zero because a silent empty set is precisely the
shape that reads as "implemented" for ever.

⚠ **Residual, stated rather than glossed:** this preserves *grading* latency, not *whole-site
re-measurement* latency. A NEW contrast defect on a page in the rotation tail can still take up
to a full cycle (~9 days on webdesign) to be discovered — versus 3 days today for the 60 pages
in the prefix, and versus **never** today for the 86 in the tail. So it is a strict improvement
over the status quo for 86 pages and a regression for at most 60, and the regression window is
discovery, not confirmation. That is the honest claim and it is the one that goes to the council.

---

## 2026-08-26 — baseline pinned, and the two build-enforced constraints the shipping commit must satisfy

**Green baseline pinned at `4ab76e5e0ea0f4f7d83aceceffb5d87a432370d4`.**
`go test ./platform/orchestration/actions/ -run 'RenderAudit|Truncat' -count=1` → `ok`.
Pinned as a SHA rather than as `HEAD`, per the recorded trap that committing kills a
`git show HEAD:` baseline — and I have committed six times since starting this lane.

### The finding-code registry will make me declare a reader, and it VERIFIES the claim

`RENDER_AUDIT_TRUNCATED`'s current entry in
`docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json`:

```json
{ "disposition": "human-evidence",
  "writer": "platform/orchestration/actions/request_render_audit_action.go:173",
  "why": "'A capped sweep reporting clean is a false green — the unaudited tail is the SAME
          pages every run.' A reader/rotation is COMMISSIONED (owner decision 4, 2026-08-25 —
          bugs_open/394 …); upgraded when it ships. Accepts the retention window it lives
          under (30 days unresolved, 14 resolved — migration 466; 365 days under 567)." }
```

`[MEASURED 2026-08-26]` the registry holds **62** codes: 33 `human-evidence`, 15 `operational`,
7 `consumed`, 7 `instrumented`. The shape a `consumed` entry must take, from a live exemplar:

```json
{ "disposition": "consumed",
  "writer": "…:636", "reader": "…:228", "reader_sink": "agent_error_log",
  "note": "…" }
```

⚠ **`reader` is not a string you can just type.** `cmd/config-key-audit/findingcodes.go`'s
`repoSourceReader` OPENS the named file and requires it to mention **both the code and its
sink** — the field exists because `consumed` was "silently ambiguous" and one entry would
otherwise have "passed the reader check and read as healthy for ever". So the reader file must
literally contain `RENDER_AUDIT_TRUNCATED` and `agent_error_log`. Satisfied naturally by the
planned `cmd/config-key-audit/rendertruncation.go`, but it is a constraint on where the reader
may live, not a formality — and it is checked at commit time by
`scripts/check-finding-code-registry.sh` via `.githooks/pre-commit`.

### The roster test is about distinctness, not disposition

`platform/orchestration/actions/finding_code_roster_test.go` reads the same JSON and asserts
code distinctness, with a vacuity guard that refuses to certify against fewer than 10 codes.
Keeping the same code string means it is unaffected — checked so I do not discover otherwise
at commit time.

---

## 2026-08-26 — CORRECTION to my own priority-set census: right number, wrong predicate

> **⚠ CORRECTED 2026-08-26.** The census two sections above filtered
> `status NOT IN ('complete','cancelled','rejected')`. **That is not the platform's closed
> set.** The grader uses `workItemClosedStatuses` (`work_items_common.go:85-91`):
>
> ```
> terminal (dedup / ON CONFLICT):  complete verified rejected wont_fix cancelled failed unresolved
> closed   (retraction):           complete verified rejected wont_fix cancelled
> ```
>
> — and the difference is deliberate (RFC_010, owner ruling 2026-08-02 "Decision 2:
> `unresolved` is OPEN"): `unresolved` and `failed` mean "we gave up" and "the handler
> errored", never "this stopped being a problem". My filter omitted `verified` and
> `wont_fix`, so it counted two settled statuses as open.

Caught by the planning agent reading `work_item_retraction.go:118-128` — the grader's actual
candidate query — rather than by me.

**Re-measured with the exact predicate**, `[MEASURED 2026-08-26]`, and keyed on the path
recovered from `item_key` (which is what the grader matches on — NOT `page_id`):

```sql
SELECT s.domain, count(*) AS open_items,
       count(DISTINCT split_part(replace(wi.item_key,'contrast_failure:',''),'#',1)) AS distinct_paths
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE wi.item_type = 'contrast_failure'
  AND wi.status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
GROUP BY 1 ORDER BY 2 DESC;
```

| domain | open items | distinct paths |
|---|---|---|
| robot-hands.com | 23 | **17** ← fleet max |
| idea.uk | 20 | 7 |
| loancash.co.uk | 10 | 10 |
| … | | |
| **webdesign.co.uk** | **3** | **3** |

**The number did not move: webdesign is still 3 and 3.** It happens to carry no `verified` or
`wont_fix` contrast rows, so the wrong predicate and the right one agree on the one site the
figure was about.

**That agreement is not a vindication and I am recording it as a near-miss rather than a
footnote.** A figure that comes out the same under a wrong predicate has not been checked by
that agreement — it has been checked by luck, and the luck is site-specific. Had I run the same
census on idea.uk or robot-hands to justify the design, the two predicates could have differed
and I would have had no way to notice, because the number would still have looked plausible.
The design conclusion stands; the evidence for it now rests on the predicate the grader actually
uses.

### Two things the planning agent established that I had assumed

1. **The page identity is the path inside `item_key`, not `page_id`.**
   `workItemKey("contrast_failure", path+"#"+selector)` (`work_items_common.go:510-512`), and
   the grader prefix-matches it against `urlPath(u)` of the measured URL. My census keyed on
   `page_id`, which happens to be populated on all 111 open contrast rows — but the grader has
   never read it, so a priority query keyed on `page_id` would have been a second, divergent
   notion of "which page" living beside the real one. It must key on `item_key`.
2. **The drain grades exactly ONE item type on this path**:
   `write_render_audit_findings_action.go:791` calls
   `loadAuditRetractionCandidates(ctx, tx, siteID, "contrast_failure")` and that is the only
   call on this payload's path. So "the priority set is what this reply can grade" has one
   member by construction — which is a stronger reason to exclude `undeployed_asset` than my
   NULL-`page_id` measurement was. Both hold; the first is the one that will not rot.

---

## 2026-08-26 — council submitted

`SUBMISSION_CORR = f67593f5-90cb-4a35-9cc0-926254645192`
(`council_submission_394_r1.json`, 8 edits, 16 `grounded_in` quotes.)

`DRY_RUN=1` passed admission first, so no credits were spent discovering a schema slip. Budget
~30 minutes, not ~2 — the council itself takes 2-5 but the dispatch queues behind the fleet.

Verdict:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='f67593f5-90cb-4a35-9cc0-926254645192' AND kind='council_report'
 ORDER BY created_at;
```

Committing code before the verdict lands uses **`Council-Submitted:`**, never
`Council-Reviewed:` — the latter on an unread verdict is what `098` buckets as MISMATCH, the
coverage report's dishonesty surface. `098` resolves the correlation at report time, so the
commit is credited automatically once it turns approved.

Ownership re-checked at this phase boundary (RUNBOOK §7): `who-owns.py` names this lane, and
`git status` shows no other session in `request_render_audit_action.go` or `work_items_common.go`.

---

## 2026-08-26 — round 1 REVISE, and it found a live defect I had shipped

`decision = revise`, gating objection from **editquality**, corr `f67593f5`.

> `pagePathFromContrastKey` assumes `contrast_failure` item_key is always
> `'contrast_failure:' + path + '#' + selector`. LANDMINES flags *"site_work_items.item_key for
> contrast_failure — The render-audit package now holds TWO selector-composition schemes"* — a
> landmine specifically about this exact key.

**I had not grepped that landmine.** My own RUNBOOK §7 tells the next thread to grep LANDMINES by
symbol and table; I greped it for `discovery_checks` on the 359 lane and did not repeat the habit
for `site_work_items.item_key` on this one.

### Both halves of the hazard are LIVE. `[MEASURED 2026-08-26]`

Whole population, not a sample:

```sql
SELECT count(*) AS all_rows,
       count(*) FILTER (WHERE item_key LIKE 'contrast_failure:%#%') AS has_hash,
       count(*) FILTER (WHERE item_key ~ '#.*#')                    AS more_than_one_hash
FROM site_work_items WHERE item_type='contrast_failure';
-- 469 | 469 | 1
```

- **A SELECTOR may contain `#`.** The one multi-hash row is
  `contrast_failure:/tools/sfi26-revenue-stacker/index.html#BUTTON#c-tool-…`, and the `describe`
  scheme emits `tag#id.classes` by construction. By scheme: **453** preserve-case (render-audit),
  **16** ancestor-anchored (`.wrap H3`), 0 lowercase.
- **A PAGE URL may contain `#`, and this is the one that decided it.**

```sql
SELECT count(*) FROM pages WHERE url LIKE '%#%';   -- 1
-- idea.uk | /tools.html#audience-check | active
```

`idea.uk` carries **both** `/tools.html#audience-check` **and** `/tools.html` as ACTIVE pages, and
has **35** open `contrast_failure` rows. So a first-`#` split maps a finding on the fragment page
onto **a different real page on the same site** — and because the target exists, nothing errors,
nothing is empty, and the wrong page is prioritised successfully.

### The reviewer's own check could not have found it, and that is the lesson

The seat ran a read-only query and got 25 live keys back, **every one conforming** to the shape I
assumed. A sample of conforming rows is not evidence about the non-conforming ones. What found it
was the LANDMINE — a prospective record written by a lane that had already been bitten — and the
seat's contribution was *knowing the landmine existed for this footprint*. That is precisely the
division of labour `LANDMINES.md` was created for, and it worked.

### The fix: stop parsing, match forward — which is the grader's own method

`write_render_audit_findings_action.go:740-750` never parses a key. It builds
`workItemKey("contrast_failure", p+"#")` per audited page and prefix-matches. I now do the same,
ordered **longest path first** so a shorter page cannot claim a longer one's rows. Ambiguity is
**unrepresentable** rather than guarded against, and drift with the grader is impossible because
both sides call the same composer.

`pagePathFromContrastKey` is deleted. A comment stands where it was, carrying the measurement, so
the next author's reasonable instinct — *the composer is right there, write the inverse* — meets
the reason not to.

**Mutation-proven, non-vacuously:** reinstating the split fails
`TestPriorityMatchIsNotFooledByAHashInThePageURLOrTheSelector` on idea.uk's real shape; restoring
passes.

**Round 2 resubmitted on the same correlation** (`RESUBMIT_CORR=f67593f5…`, run envelope
`bcba9847`), so the trail accumulates rather than starting a second one.

### The estate's own claim about REVISE rounds, tested once more

MEMORY says *"a REVISE round is cheaper than the defect it finds — 2 of 4 rounds found REAL
defects incl. a false claim; revise, don't defend."* This round makes it 3 of 5 by my count, and
the defect was not hypothetical: the code was **already committed to the shared branch** when the
verdict landed, which is the ordinary state on this tree and the reason the gate is advisory
rather than blocking. Had I defended the sample of 25 conforming keys, I would have been arguing
from the weaker evidence.

---

## 2026-08-26 22:2xZ — the fix is LIVE: deploy proven, 660 applied, and a prediction recorded before the first run

### 1. The binary carries the code — probed on BOTH replicas, with both controls

The provenance log line had already scrolled (`--tail=3000`, empty) — which per CLAUDE.md means
"not in range", never "unstamped". So: the capability probe, which has no shelf life.

`[MEASURED 2026-08-26 ~22:1xZ]`

| pod | `render_audit_page_cursor` (marker) | `content validation failed` (positive control) | `zzz_invented_marker_xq9` (negative control) |
|---|---|---|---|
| `agent-chassis-5864bf97c5-5l8xd` | **3** | 1 | 0 |
| `agent-chassis-5864bf97c5-68t5h` | **3** | 1 | 0 |

Both replicas, not one — `logs deploy/X` reads one pod of N, and a mixed fleet is the recorded
trap. The marker is the TABLE NAME, which appears in my three SQL literals and nowhere else in the
estate; the positive control proves the grep works; the negative control proves it discriminates.
`rotate_coverage` is present too. **Uniform fleet, no mix.**

### 2. Migration 660 applied — and it took three goes to get right

- **649 → 660.** My migration was renumbered TWICE. I took 646; another lane had committed 646 at
  16:27 (mine 16:51). I moved to 649; another lane took 649 as well. By then the sequence was at
  659, so this is 660. **`sql_for_agents` numbers are a shared sequence with no reservation
  mechanism, and on a busy afternoon a number you picked an hour ago is not yours.** Take
  `max+1` at the moment you apply, not at the moment you write.
- **`snapshot_agent('x')` aborted the whole transaction** — the function is overloaded and a bare
  literal is ambiguous. Safe failure (nothing written, verified), and now a LANDMINE.
- Applied clean at **22:20:40Z**: `snapshot_agent` NOTICE, `CREATE TABLE`, 4 `COMMENT`,
  `UPDATE 1`, both `DO` verify blocks, `COMMIT`.

`[MEASURED 2026-08-26 22:2xZ]` post-apply state:

| check | value |
|---|---|
| `render_audit_page_cursor` exists | **1** |
| `render-audit-agent` `rotate_coverage` | **true** |
| `design-critique-agent` carries the key | **0** ← the verify block's own assertion held |
| backup in `agent_definitions_backup` | present, `snapshot_taken_at` 22:20:40Z, **`backup_has_flag = f`** |

That last row is the one that matters: **a backup that already contains your change is not a
backup.** It is the pre-change state, so it is restorable.

> **⚠ CORRECTION to my own first reading, recorded because I nearly filed a bug on it.** I checked
> the snapshot with `SELECT count(*) FROM agent_definitions WHERE type='render-audit-agent' AND
> is_snapshot` and got **0**, which reads exactly like "the function reported success and wrote
> nothing" — the estate's most-repeated defect shape, so I believed it for a minute. It was **my
> instrument**: the two-arg overload writes to `agent_definitions_backup`, a different table, and
> RETURNS THE SOURCE ROW'S ID. The one-arg overload is the one that writes `is_snapshot` rows into
> `agent_definitions`. Full trap in LANDMINES.

### 3. ⚠ PREDICTION, recorded BEFORE the first cursor run

webdesign.co.uk is **151 live pages** now (146 two hours ago — it is still growing). 3 pages carry
an open `contrast_failure`, so the priority set takes 3 of 60 and the rotation gets **57**.

| field | predicted |
|---|---|
| `coverage_mode` | `cursor` |
| `pages_total` | 151 |
| `pages_audited` | 60 |
| `window_first` | `index` |
| `window_last` | **`tool-head-architect`** |
| cursor row `after_nav_order` / `after_name` | 100 / **`tool-head-architect`** |
| `priority_open_items` | 3 |
| `priority_paths` | `/tools/index.html`, `/learn/accessibility/touch-targets.html`, `/learn/data/cors-scraping.html` |

**`tool-head-architect` is not an arbitrary name.** It is the *exact* page the old prefix cut at —
the last page the audit had ever seen on this site, measured this afternoon at rank 60. Under the
cursor it becomes the last page of window 1 and the boundary the next run starts past. If the
cursor is working, run 2 begins at `tool-html-minifier`, the page that had never been audited and
never would have been.

This prediction could come out wrong in several ways — a different priority count, a window that
ignores the priority set, a cursor that lands on a priority page — and it is written down before
the dispatch so that it can.

---

## 2026-08-26 22:2x–22:3xZ — the cursor is PROVEN AT THE ARTEFACT, and the first live run found a defect no test could

### Run 1 — every one of the ten predictions hit

Dispatched `render-audit-agent` at webdesign.co.uk by hand (`orchestrate_safe.sh`, corr
`5e6a9e04`). The durable row, `[MEASURED 2026-08-26 22:25:20Z]`:

| field | predicted | observed |
|---|---|---|
| `coverage_mode` | `cursor` | `cursor` ✓ |
| `pages_total` | 151 | 151 ✓ |
| `pages_audited` | 60 | 60 ✓ |
| `window_first` | `index` | `index` ✓ |
| `window_last` | `tool-head-architect` | `tool-head-architect` ✓ |
| `priority_open_items` | 3 | 3 ✓ |
| `priority_paths` | the 3 named | exactly those 3 ✓ |
| `priority_dropped` / `priority_not_live` | 0 / 0 | 0 / 0 ✓ |
| `cursor_cleared` | false | false ✓ |
| cursor row | `(100, tool-head-architect)` | `(100, tool-head-architect)` ✓ |

`audited_paths` carries 60 entries with **the three priority pages FIRST**, then the rotation
from `/index.html` — the ordering that exists so a run the adapter abandons still measures the
latency-sensitive pages.

### Run 2 — THE ONE THAT MATTERS

`[MEASURED 2026-08-26 22:32:36Z]`, corr `0b2953f9`:

```
window_first = tool-html-minifier      window_last = tool-entropy-meter-guide
cursor:  (100, tool-head-architect)  ->  (200, tool-entropy-meter-guide)
```

**`tool-html-minifier` is the page that had never been audited and never would have been.** It sat
at rank 61 behind a cap of 60, and this afternoon's census named it as the first page of the
permanent tail. It is now the first page of window 2.

And `after_nav_order = 200` says something stronger: the window has reached the **guide band** —
the 45 `tool-*-guide` pages that were unreachable at *any* cap below 98. Two runs have carried the
audit from "the first 60, for ever" to page 120 of 151, through a class of page it had never seen.

⚠ **Say precisely what is proven.** The cursor SELECTS and SENDS these pages — that is what the
durable row records, and it is the half this change owns. It does **not** prove the browser
measured them: both runs ended at `complete_error` with `{"message": "Request timed out (code:
TIMEOUT)", "failed_step": "audit"}`. See the open item below. Coverage of the REQUEST is proven;
coverage of the MEASUREMENT is not.

### The defect the live run found, which no unit test could have

The cursor row came back keyed `agent_type = 'generic'` while the SAME run's durable truncation
row recorded `render-audit-agent`. **Two identities for one run.** The key read
`params.ExecutionContext.Sender.AgentType` — whoever put the message on the topic. A hand dispatch
goes via `system.agent.generic.requests`; the scheduled rotation uses
`system.agent.scheduled.requests`. Keyed on that, one logical caller keeps a **separate cursor per
dispatch path**, so a hand run's coverage is invisible to the scheduled run and each restarts from
the top.

Fixed to `runningStepProvenance(params)` — `ExecutionContext.ResolvedAgentType()` with a
`params.AgentType` fallback, the same resolver `LogActionFindings` uses to stamp the row. One
function, so the key and the row cannot disagree.

**Why no test caught it:** `renderAuditParams` sets `Sender.AgentType` to `"render-audit-agent"` by
hand, so both readings agreed in every fixture. It took the artefact. And my first attempt at the
regression test set only `Sender` — which models the *fallback*, not production, because
`ResolvedAgentType()` prefers `RunAgentType`. Corrected; mutation-proven by reverting to `Sender`.

### ⚠ OPEN, and NOT mine: the audit step times out before the adapter replies

Both runs: `complete_error`, `TIMEOUT` on the `audit` step, ~3 minutes after dispatch. The
render-audit adapter pod (`render-audit-adapter-5944c6458c-mw246`) is HEALTHY — it completed a
robot-hands.com render audit at **22:00:54Z** and sent its response — but logged **nothing** from
either of my dispatches.

So the request is produced and the adapter never sees it. That is upstream of the browser and
downstream of my change: page selection wrote its durable row and its cursor correctly on both
runs. **Do not read the timeouts as evidence against the cursor, and do not read the cursor's
success as evidence the audit works.**

⚠ Instrument note: `kubectl logs -l app=render-audit-adapter` returned another service's lines
(one image, every label — the recorded trap). Read the POD, not the label.

---

## 2026-09-02 — a week on: the cursor has been running unattended, and four of five open items closed themselves

Returning after six days. Every `[MEASURED 2026-08-26]` figure below has been re-taken.

### The mechanism ran on the SCHEDULED path, unattended, and it worked

`[MEASURED 2026-09-02]` — every `RENDER_AUDIT_TRUNCATED` row since my last session:

| when | agent | domain | mode | window |
|---|---|---|---|---|
| 08-27 10:49 | design-critique | leopardess | **prefix** | — |
| 08-27 12:01 | render-audit | webdesign | **cursor** | `index` → `tool-head-architect` |
| 08-27 21:06 | render-audit | **loanandmortgagecalculator** | **cursor** | `index` → `tool-overpayment-priority-guide` |
| 08-30 12:35 | render-audit | webdesign | **cursor** | `tool-html-minifier` → `tool-entropy-meter-guide` |
| 08-30 21:41 | render-audit | loanandmortgagecalculator | **cursor** | 2 pages, first == last |

Three things this settles that no hand-run could:

1. **The mode split is correct in production.** `design-critique-agent` writes `prefix` and
   `render-audit-agent` writes `cursor`, exactly as designed — and the `(absent)` rows on 08-26
   are the pre-roll binary, which is what `(absent)` is supposed to mean.
2. **A second site entered the population and cycled cleanly.** `loanandmortgagecalculator.co.uk`
   crossed the cap (61 live pages) and ran window 1 on 08-27, then a **2-page final window** on
   08-30 with `window_first == window_last` — that is the cycle-completion branch (`cursor_cleared`)
   firing on a real site, a path only production could exercise.
3. **The audits COMPLETE now.** `contrast_failure` rows created 08-27 **28**, 08-28 6, 08-29 2,
   08-30 1, 09-01 2. The `TIMEOUT` I handed over was transient. **Open item (b) is closed by
   evidence, not by assumption** — findings are being written again.

### ⚠ The identity bug's live damage, observed exactly as predicted

webdesign's windows read: `index→head-architect` (08-26 hand), `html-minifier→entropy-meter-guide`
(08-26 hand), **`index→head-architect` again (08-27 scheduled)**, `html-minifier→entropy-meter-guide`
(08-30 scheduled).

The scheduled run **restarted from the top** on 08-27 because my hand-runs' cursor was filed under
`agent_type='generic'` and it looked under a different key. That is precisely the split I described
in the handoff, now with a dated instance. The cursor table still carries both rows:

```
webdesign.co.uk | generic            | 200 | tool-entropy-meter-guide | 2026-08-26 22:32:38
webdesign.co.uk | render-audit-agent | 200 | tool-entropy-meter-guide | 2026-08-30 12:35:41
```

The `generic` row is orphaned history. The `render-audit-agent` row is the live one.

### The acceptance arm: 117 of 151, two runs in, third due today

```sql
-- union of audited_paths over the SCHEDULED cursor runs on webdesign
-- 2026-09-02: union_pages = 117, total_now = 151, runs = 2
```
webdesign is due again at **2026-09-02 12:35Z**. Window 3 should complete the cycle. This is the
bug's own acceptance test and it is one run from being answerable.

### ⚠ A NEW LANDMINE lands on the probe I used last week — and the direction matters

`LANDMINES.md` gained (2026-08-24, another lane): *"BusyBox `grep` over `/proc/1/exe` reports FALSE
ABSENCES — and your present/absent controls PASS while it does it."* The fleet's images are BusyBox
v1.37; its grep works line-by-line and a Go binary's "line" can be enormous, so a literal inside an
over-long line reads as absent with a clean exit code.

**That is the instrument I used on 2026-08-26.** So: does it undermine last week's conclusion?

**No, and the reason is the direction of the failure.** The claim "the code is live" rested on
three **PRESENCES** (`render_audit_page_cursor`, `rotate_coverage`, the positive control). The
described fault produces false **ABSENCES**; a false absence would have made me under-claim, never
over-claim. The negative control reading 0 *could* have been vacuous — that costs me the control,
not the conclusion — and the conclusion was independently confirmed at the artefact within minutes
when the cursor row appeared with `coverage_mode=cursor`.

Re-probed today with the prescribed instrument (`tr '\0' '\n' | grep -Fc`, both controls through
the SAME pipeline), on `agent-chassis-5bd89cf49-t4wdl`:

```
render_audit_page_cursor 3   rotate_coverage 2   runningStepProvenance 1
selectAuditWindow 2          zzz_invented_marker_xq9 0
```

### Is the identity fix live? Strong evidence, NOT decisive — and I will not claim more

Running image `v1.0.1351`, pods started **2026-09-01 21:00Z**; the fix committed **2026-08-26
23:28+01:00**, and `make build-*` builds from committed HEAD. Add the 08-30 scheduled run writing
under `render-audit-agent`, and the fix is almost certainly in.

**But that is not proof.** `faf4872ce` changed a call site and comments — it added **no new string
literal**, so no binary probe can discriminate it, and `runningStepProvenance` being present only
shows the pre-existing function is there. The 08-30 key is also consistent with the OLD code if
`Sender.AgentType` already equalled `render-audit-agent` on the scheduled topic.

**The decisive probe, for whoever picks this up:** hand-dispatch one audit (which goes via
`system.agent.generic.requests`) and read the cursor's `agent_type`. Old code → a `generic` row.
New code → `render-audit-agent`. **Do it AFTER window 3 has run**, or it consumes the window the
acceptance union needs.

---

## 2026-09-02 13:0x–13:2xZ — the acceptance arm PASSES, and the identity fix is confirmed live by a discriminating test

### 1. Window 3 ran on schedule and CLEARED the cursor

`[MEASURED 2026-09-02]` webdesign.co.uk, `render-audit-agent`, all cursor-mode rows:

| when | window | audited | cleared |
|---|---|---|---|
| 08-27 12:01 | `index` → `tool-head-architect` | 60 | false |
| 08-30 12:35 | `tool-html-minifier` → `tool-entropy-meter-guide` | 60 | false |
| **09-02 13:09** | `tool-fluid-typography-guide` → **`tool-llm-cost-calculator`** | **37** | **true** |

`tool-llm-cost-calculator` is `nav_order` 201 — the last page in the ordering. The final window is
the remainder (37), and `cursor_cleared = true`. Afterwards the `render-audit-agent` cursor row was
**gone from the table**: the `deleteAuditCursor` branch fired in production, on the scheduled path,
unattended.

### 2. THE ACCEPTANCE ARM PASSES — 151 of 151, zero missed

`bugs_open/394` §2 asks: *"over consecutive rotation runs on webdesign.co.uk, the union of audited
pages reaches all [of them] (verify by the audit's own durable page list, not the status)."*

```sql
WITH sched AS (SELECT context FROM agent_error_log
   WHERE error_code='RENDER_AUDIT_TRUNCATED' AND domain='webdesign.co.uk'
     AND agent_type='render-audit-agent' AND occurred_at >= '2026-08-27'
     AND context->>'coverage_mode'='cursor'),
u AS (SELECT DISTINCT p AS path FROM sched, jsonb_array_elements_text(context->'audited_paths') p)
SELECT (SELECT count(*) FROM u) AS union_pages, (SELECT count(*) FROM sched) AS runs;
```

`[MEASURED 2026-09-02]` **union_pages = 151, runs = 3, pages_total = 151.**

And graded the harder way — against the site rather than against itself:

| live now | missed by the union | missed AND born mid-cycle |
|---|---|---|
| **151** | **0** | 0 |

**Zero missed.** Not "151 sent" — 151 *distinct* pages, deduplicated, covering every live page on
the site. The three priority pages ride every run and are counted once. The second acceptance arm
is met too: the message no longer says "the SAME pages every run" in cursor mode.

Recall what this replaced: the same first 60 pages, for ever, with 91 pages — including all 45
`tool-*-guide` — that had never been audited and never would be.

### 3. The identity fix is LIVE — settled by a test whose two outcomes were opposite

The last handoff called this "strong evidence, NOT proof", because `faf4872ce` added no string
literal and no binary probe can discriminate it. The cycle-clear created the perfect conditions:
the `render-audit-agent` cursor was gone, and only my orphaned `generic` row survived at
`(200, tool-entropy-meter-guide)`.

**Prediction recorded before dispatch**, and the two hypotheses predicted *opposite* windows:

| | window_first | cursor effect |
|---|---|---|
| fix LIVE (keyed on running agent) | `index` — a fresh window 1 | NEW row under `render-audit-agent` |
| fix NOT live (keyed on `Sender`) | `tool-llm-cost-calculator` — the tail | reads and clears the `generic` row |

`[MEASURED 2026-09-02 13:21Z]` result: `window_first = index`, 60 pages, `cursor_cleared = false`,
and a **new row under `agent_type='render-audit-agent'`** at `(100, tool-head-architect)`. The
`generic` row was neither read nor updated — its timestamp stayed at 2026-08-26 22:32:38.

**The fix is live.** A hand dispatch over `system.agent.generic.requests` now keys the cursor on the
RUNNING agent, so hand-runs and scheduled runs share one position, which is the whole point.

That also made the `generic` row provably dead — nothing writes or reads that key any more — so I
deleted it (`DELETE 1`). Recorded here because deleting a row is the sort of thing a later reader
should be able to find the reason for.

### 4. What remains, and why 394 stays OPEN

The cursor half is **fixed, live and accepted**. The bug is not closeable, because the owner
commissioned TWO things (decision 4, 2026-08-25) and the second is not driven:
`cmd/config-key-audit --render-truncation` is built, tested, mutation-proven on four arms and
recorded in `finding_code_registry.json` as the `consumed` reader — but **no CronJob runs it**.
`kubectl get cronjob` shows no `render-truncation-check`. Until it is scheduled, the registry's
`consumed` claim is true about the code and false about the estate, which is exactly the state
`DBG-075` exists to prevent. The `/bugs_closed/` bar is "fixed AND live"; half of this is not yet
live, so it stays in `/bugs_open/`.

---

## 2026-09-02 16:1xZ — the CronJob is DEPLOYED and PROVEN AT THE POD, and the council approved

### Deployed by the release, not by hand

`[MEASURED 2026-09-02]`

```
NAME                      SCHEDULE     SUSPEND   LAST     IMAGE
render-truncation-check   50 7 * * *   false     <none>   docker.io/aqls/render-truncation-check:v1.0.1354
```

Chassis is on the same tag (`v1.0.1354`), so the release carried both. `LAST: <none>` — the
schedule has not fired yet; next is **07:50Z tomorrow**.

### Proven at the pod, which is exactly what the council asked for

Council `f49da30d` returned **APPROVED**, 2 advisories, none high-severity. The substantive one:

> **debug_historian, MEDIUM:** *"New CronJob shipping a binary with genuinely new logic (dormancy
> arm) has no specified post-deploy verification against the running pod. The plan relies entirely
> on pre-ship dry-runs…"*

Right, and answered by doing it rather than by promising it. A manual Job created **from the
CronJob** (`kubectl create job --from=cronjob/render-truncation-check`) — same image, same CMD, same
env, so it exercises the shipped artefact and not a hand-rolled approximation:

```
render-truncation: 19 RENDER_AUDIT_TRUNCATED row(s) across 4 site(s) and 2 caller(s)
  in an agent_error_log holding 61872 row(s); 0 finding(s); 1 dormant group(s);
  acks=/app/render_truncation_acks.json
  [dormant] loancalculator.co.uk / render-audit-agent: newest row is >14d behind …
  Every truncation row is accounted for …
```

Pod `Succeeded`, job `succeeded=1`. Three things that only an in-cluster run can establish, and all
three hold: the image pulls (no `ImagePullBackOff`, which this fleet reports as RUNNING); the acks
file really shipped **in-image** (`acks=/app/render_truncation_acks.json`, the container path, not
a repo path); and the DB env wiring works (61,872 rows read).

**And the durable half, which is the difference between "ran" and "ran and recorded":**

```
2026-09-02 16:17:57Z | render-truncation | render-truncation: 19 … 0 finding(s); 1 dormant group(s)
```

One `doc_notes` row per run, clean included — so a MISSING row means the job did not run and can
never read as "nothing is wrong". The manual Job was deleted afterwards; `successfulJobsHistoryLimit`
only reaps jobs the CronJob itself creates.

### The other two advisories, answered from the tree

- **editquality, MEDIUM — "no edit flips the registry to `consumed`".** Already done, and before
  this submission: `41b03241d` (2026-08-26) set `disposition: consumed`,
  `reader: cmd/config-key-audit/rendertruncation.go:1`, `reader_sink: agent_error_log`. Verified
  again just now. The seat was reading the plan, which is the right thing for it to do — the edit
  was absent from THIS submission because it had shipped in an earlier round of the same
  correlation trail. Nothing to change.
- **bug_historian, LOW — dormancy shares the structural risk of a silent drop.** Accepted and
  already mitigated in the shipped design rather than in prose: dormant groups are **counted and
  NAMED** in every run (the pod output above names loancalculator), and the blind spot is stated in
  the dockerfile header, the CronJob header and the handoff. The distinction the seat is drawing —
  named exception versus silent drop — is the one this implementation is on the right side of.
