# HANDOFF — bugs_open/450, continue here (2026-09-03, late)

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_450_tool_page_shells/`
Bug: `bugs_open/450_HANDOFF_2026-09-02_planned_tool_pages_are_built_as_prose_shells_by_the_link_repair_before_their_tools_exist.md`
Standing five in this directory: PLAN · RUNBOOK · NOTES (a)–(t) · README_where_we_are · SUMMARY_2026-09-03.

**Read §1 first.** It corrects a wrong conclusion of mine (the guard did NOT fail) and names the
real live bug: six pages serve working tools whose `component_id` is NULL, so they are one rerender
away from losing them. **That repair is time-sensitive and is the first job.**

---

## §1 ⚠ CORRECTED — THE GUARD DID NOT FAIL. A DIFFERENT AND WORSE BUG IS LIVE: `save_page_sections` KEEPS A TOOL'S HTML AND DROPS ITS `component_id`

> **This section said "THE GUARD DID NOT STOP A GENERIC WRITE" for about twenty minutes and that
> was WRONG.** Corrected 2026-09-03 ~13:5xZ after the `portfolio_positioning` lane measured the
> artefact and I verified it. The error was mine and it is the day's recurring one, in its most
> consequential form: **I measured POST-write state and inferred PRE-write state.** Left uncorrected
> it would have sent the next session hunting a guard bug that does not exist. Kept visible rather
> than deleted, because the shape is the lesson.

### What actually happened

The `portfolio_positioning` lane's repair wave attached real tool components to these six pages at
**09:34:09, 09:37:24, 09:40:35, 09:46:14, 09:49:05 and 09:54:14Z** — one library component each
(`tool-robots-txt-tester-seotools-co-uk` etc., `component_level='tool'`, `is_active=true`, still
present and untouched).

So at **13:05–13:24Z**, when `needs_content_page` → `page-build-handler` wrote to them, **these
pages were NOT shells. They carried live tools.** The guard therefore **correctly allowed** the
write, and its silence was right rather than a failure.

**`save_page_sections` then deleted all three rows and re-inserted them ~80 ms later, preserving
the tool slot's `rendered_html` (17,938–23,953 bytes) and setting its `component_id` to NULL.**
Verified directly:

```
tool-robots-txt-tester        hero-tool           cid=set    3,500 B   13:05:15
tool-robots-txt-tester        generic-text-block  cid=set    3,142 B   13:05:15
tool-robots-txt-tester        tool-robots-txt-…   cid=NULL  20,839 B   13:05:15   <-- reference gone
```

`page_component_history` for that slot shows six writes at 13:05:14, **every one already
`cid=NULL`**.

### Why my "0 tool rows ever" reading was an artefact

`toolShellPredicateFor` — and every census in this lane, and the query I used to declare the
falsification — joins `page_components` to `content_components` **on `component_id`** and filters
`component_level='tool'`. **With the id NULL the join drops the row**, so a page serving a real
20 KB tool reads as having *no tool component, ever*. The zero I measured was created by the very
write I was trying to explain.

**Consequences, and (2) is the one that bites this lane's own instruments:**

1. **Serving is correct today.** All six serve working tool controls with instance-scoped ids
   (`id="c-tool-robots-txt-tester-fetch-domain-input"`), 78–85 KB bodies, 6 scripts. A visitor sees
   a working tool. So the writes landed and the outcome was survivable.
2. **⚠ THE CENSUS AND THE GUARD NOW BOTH OVER-REPORT BY SIX.** These six count as shells in the
   66/67 population and are not. The guard will also now *refuse* future generic writes to them —
   which is accidentally protective, given (3), but for a reason that has nothing to do with what
   the predicate is meant to express. **This is the census-versus-predicate divergence written up
   in `016b` §9 this morning, reappearing inside my own instrument one field along.**
3. **THE REAL EXPOSURE IS THE NEXT REBUILD.** Anything that regenerates from `component_id` rather
   than from stored HTML has nothing to regenerate. **These six pages are one rerender away from
   losing their tools for real, and nothing about their current appearance would warn anyone.**

### What this means for `29b40e8bc` — the worry is RETIRED

The previous version of this section warned that removing the tool arm from `save_page_sections`
may have removed a needed backstop. **It did not, and could not have:** the pages were not shells
at write time, so neither the narrowed nor the un-narrowed guard would have refused. The narrowing
is irrelevant to this incident in both directions. (It is also not the cause: `29b40e8bc` was NOT
aboard `v1.0.1358`, the image running at 13:05Z, and it changes only the guard condition, not the
delete-and-reinsert.)

### The actual open bug — likely deserves its own `bugs_open/` file

**`save_page_sections`' delete-and-reinsert preserves the `rendered_html` of a slot it does not
recognise as one of its own planned sections, and drops the `component_id`.** The plan for these
pages names `hero-tool` and `generic-text-block`; the tool slot is not in it, so the save carries
the bytes forward without the reference.

### MECHANISM — verified in code 2026-09-03, so this is no longer a question

`save_page_sections_action.go:1036-1041` takes `componentIDPtr` from **`section.ComponentID` in the
payload it is handed**; the insert at `:1124-1127` writes whatever that is:

```go
var componentIDPtr *uuid.UUID
if section.ComponentID != "" { if parsed, err := uuid.Parse(section.ComponentID); err == nil { componentIDPtr = &parsed } }
...
INSERT INTO page_components (..., rendered_html_digest, slot_name, component_id, ...)
VALUES ($1,$2,$3, md5($3), $4, $5, ...)
```

**The reference is CARRIED, never resolved from the database at insert time.** So any preservation
path that keeps an unrecognised slot's HTML without also carrying its `ComponentID` nulls the
reference — which is why the bytes survived and the link did not. **The fix therefore probably
belongs in whatever assembles the section list, not in the insert.** (Found by the
`portfolio_positioning` lane; verified here at the lines.)

Note in passing: `rendered_html_digest` is `md5($3)` on this path, so it is always set here — a
NULL digest elsewhere comes from a different writer. That matters for §1's repair guard, below.

### FLEET SCOPE `[MEASURED 2026-09-03 ~14:0xZ]` — 20 orphaned rows across 7 sites, and it is NOT tool-specific

`page_components` rows with `component_id IS NULL` and `build_status <> 'removed'`:

| kind | rows | where |
|---|---|---|
| **tool slots** | **8** | seotools ×6 (**REPAIRED 14:01:10Z** by the portfolio lane), `idea.uk/tool-funding-fit`, `loanzy.uk/tool-loan-vs-savings` |
| blog-index listings | 8 | boxingonline `articles-index` ×6, finetuning `blog`, ai-agent-orchestration `blog` |
| content sections | 3 | finetuning `our-position-on-ai` |
| **a 25,901-byte `game` slot** | 1 | `gamesdesign.co.uk/game-jelly-invaders` |

**Both remaining tool pages are SERVING REAL TOOLS** — probed at the body: `idea.uk/tools/funding-fit/`
**2 forms / 40 inputs / 8 scripts**; `loanzy.uk/tools/loan-vs-savings/` **0 forms / 7 inputs /
1 select / 6 scripts**. So they are over-reported as shells by this lane's census exactly as the six
were, and they are one rerender from losing their tools. **They are the outstanding repair.**

⚠ **The `game` slot is the one to look at hardest.** 25,901 bytes with no reference, on a page type
this lane never considered — evidence the class is about *any* plan omitting a slot another
producer inserted, not about tools.

### THE REPAIR, and one trap already paid for

The portfolio lane's shape, which worked: a **guarded** `UPDATE` that refuses unless the expected
number of orphans exist and each maps to exactly one active component via **`cc.function =
pc.slot_name`** (the acceptance coupling), with each row's `md5(rendered_html)` and length
snapshotted to a temp table **before** and compared **after**. Result `UPDATE 6`, bytes untouched,
no rerender needed because the served bytes were already correct.

> ⚠ **DO NOT assert digest integrity across a POPULATION.** Their first verify block used
> `rendered_html_digest IS DISTINCT FROM md5(rendered_html)` and refused with "the repair altered
> bytes" — falsely. `IS DISTINCT FROM` convicts a NULL digest, and **a NULL digest is a normal
> state: 206 of 3,220 rows fleet-wide, of which only 34 genuinely mismatch.** Compare the rows you
> actually touch, before against after — not the population.

First moves:

```bash
# 1. Scope it — how many pages fleet-wide have a tool-level slot with a NULL component_id?
#    (This cannot use the component_level join — that is the whole point. Match on slot_name.)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT s.domain, p.name, pc.slot_name, length(pc.rendered_html) AS bytes
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE pc.component_id IS NULL AND pc.build_status<>'removed'
   AND pc.slot_name LIKE 'tool-%' ORDER BY 1,2;"

# 2. Read the re-insert in save_page_sections_action.go: which branch omits component_id for a
#    slot that is not in the plan's section list, and is the omission deliberate anywhere?

# 3. Decide the repair: re-point component_id from content_components by matching slot_name to
#    cc.function (the acceptance coupling — pages.name == cc.function, sanitiseFunction guarantees
#    the tool- prefix). Cheap and reversible. Do this BEFORE any rerender touches these six.
```

⚠ **`page_type='tool'` is not the only exposure.** Any page whose plan omits a slot that was
inserted by another producer can lose that slot's reference the same way.

## §2 What is live, what is committed, what is blocked

| thing | commit | state |
|---|---|---|
| Door half — derived refusal, 6 seams | `587666be8` | **live** (also in v1.0.1359) |
| Narrowing — tool arm off `save_page_sections` | `29b40e8bc` | **live** in v1.0.1359 (§1: the worry about it is RETIRED) |
| Receipt wording corrected (stated fact, not inference) | `196319707` | live |
| Plan-side gate `enforce_tool_sources` | `5e6fee47b` | live but **KEYLESS** — inert |
| Migration 729 arming the key | `681190083` | **committed, NOT APPLIED — BLOCKED** |
| Register PBP-053 / BLD-029, finding code, landmine, 016b ×2 | various | committed |

**Councils: both APPROVED.** Door `2b236e83-ffd1-4911-b73f-1c17249064c1`; gate
`4e7497ed-62ed-4426-a814-8361754c2352`. All mediums actioned (see NOTES (i), (k)).

**BLOCKED, needs the owner:** applying `729` was **refused by the session permission classifier**
(a live-DB write). Not worked around. Preconditions are otherwise met — verdict read, code live.
Recipe and the reason it waited: RUNBOOK §10. Until it applies, **the planner keeps naming tool
pages whose tools do not exist.**

**Owner lever, still available and now probably NOT wanted:** `DISABLE_TOOL_SHELL_REFUSAL` disarms
the tool arm fleet-wide with no build, scoped so it cannot touch migration 164's owned protection.
⚠ Per §1 it would also disarm the accidental protection those six `component_id`-NULL pages are
currently getting, so **do not pull it until their references are repaired.**

---

## §3 Numbers you can trust, and the ones you cannot

**Use the guard's own predicate — `toolShellPredicateFor` in `owned_page_guard.go` — as the
census. RUNBOOK §1 carries a copy and says to diff it against the function first.** Four
measurement errors in this lane came from paraphrasing it (016b §9 has the pattern).

- shell pages: **66–67 / 15–16 sites**, stable across independent readings by two lanes.
- of those, **~54 already serving** deployed components; only **~13** empty.
- genuinely NEW refusals versus the pre-existing owned population: **19** (48 of 67 were already
  `rebuild_policy='owned'`).
- ⚠ **`61 / 10` appears in older text and is a FLOOR twice over** (missing `deployed_at IS NULL`
  and missing `cc.is_active`). Superseded; the bug file's Verify block explains both.
- ⚠ the census is **repair-INITIATED, not repair-COMPLETED**: a page leaves it when a tool
  component attaches, while the public still sees prose until the rerender drains.
- ⚠ drain rate: **NOT established.** "39 repairs in 12h" over-counts — the predicate cannot
  separate a first tool arriving from a tool being *regenerated*. NOTES (q).

**The harm metric, and why its earlier zero meant nothing:** historical share of writes hitting
shell pages = **275 / 17,205 = 1.60%** of all `page_component_history` writes. Condition on fleet
activity, not wall-clock. §1's 36 writes are far above that share and are the falsification.

---

## §4 Peer lanes — live dependencies

- **`portfolio_positioning`** — owned the INSTANCE repair; **CLOSED on their side**: all 8 tools
  verified serving at 14:32:42Z by instance-scoped id, and the six orphaned `component_id`s
  repaired at 14:01:10Z. ⚠ **Their served-body reading RETRACTED two findings this lane had already
  adopted** (NOTES (w)): the tool and the prose do NOT compete — served order is stable
  `hero-tool → generic-text-block → tool` on all seven, and inferring served order from stored
  `position` was the error — and the prose is **not debris**, it is accurate copy about a tool that
  is now present. **So the 13:05–13:24Z writes IMPROVED those pages**, and the only casualty was
  the reference. The "sectionless heals cleaner" inversion recorded in NOTES (s) is therefore
  itself retracted; the case for holding empty-sectioned tool pages rests on the recurring HITL
  tax, not on post-repair cleanliness.
- **`bugs_open/427` / `454`** — reported the regression §2 row 2 fixes; wrote the measurement
  pattern into 016b (`80f74b23d`). Their `9831e9ab4` is live. ⚠ Their standing warning: until it
  rolled, every re-render served stored data back at itself.
- **`bugs_open/444`** — the gate frame this siblings. Told (bug file follow-up) that arming
  `enforce_tool_sources` **changes what their listing gate does on the same plan** (ordering).
- **`428`** — adding a record-only reconciliation block BELOW both gates and a page-type
  external-producer registry. **Warned** that the set of things that WRITE a tool page is wider
  than the set that PRODUCES its tool — §1 is the evidence.
- **`ai-agent-orchestration`** — left a CONTRIB **in this lane's own directory**
  (`CONTRIB_2026-09-03_from_the_aiao_lane_rule_4_of_the_tool_generator_prompt_now_has_appended_text.md`).
  ⚠ **It was staged-but-uncommitted at handoff time, so `git log` may not show it — read it off
  disk.** Its premise is WRONG and I have told them so: it believes 729 anchors on
  `tool-generator`'s prompt. 729 touches only `build-site-planner`, so **their 732 and my 729
  cannot collide** and neither needs to re-anchor. Their measurement is worth keeping though —
  148 of 261 tool components ink a `--color-primary` fill with the page ground, against 0 of 151
  non-tool, because the pairing rule lives in `component-creator`'s prompt and never reached its
  two siblings.
- **`apis.uk`** — owns 640's rule 17; confirmed the anchor 729 defends and added an EXTERNAL
  READERS note at their end.

---

## §5 Open, ordered

1. **§1 — repair the six NULL `component_id` references BEFORE any rerender reaches those pages**,
   then scope the class fleet-wide and read the re-insert branch that drops the reference. The
   guard question is CLOSED (it behaved correctly); this is a distinct and more urgent bug.
2. **Apply migration 729** once the permission question is settled (owner).
3. **`bug_historian`'s standing objection (council, low, accepted):** nothing PINS the §7
   assumption that nothing reads planned tool pages. It is a negative finding in a code comment.
   A periodic "has a reader appeared" check is the real answer. Named, not built.
4. **Residual, explicitly out of scope:** the 61+ existing shells (instance work), the
   `owned_page_review` hold still having no consumer, `rerender_single_page`'s re-assembly path
   (bug 210 family), and N-links-one-page churn (220's own candidate).

### §5a ⚠ "A tool is a FORM" is FALSIFIED as a universal — this lane's own probe is unsound

RUNBOOK §3 and `016b` §9 both say flatly *a tool is a FORM, never a size*, and the whole bug was
found with that probe. **The `portfolio_positioning` lane has now measured two working tools that
serve ZERO `<form>` elements** (their own watcher was passing them on `forms > 0 OR fields > 2`,
which a prose page with a search box would also satisfy — they volunteered that their test was
unsound even though it happened to be right).

**So the probe is asymmetric and must be stated that way:** a form is strong positive evidence
that a tool is present; **its absence is NOT proof of a shell.** The seven original seotools shells
read 0 forms / 0 inputs / 0 selects, which is a much stronger signal than 0 forms alone.

Not reworked here — flagged as an open item. Whoever touches it: the sound version probably keys on
interactive controls generally (`<input>`, `<select>`, `<button>` beyond the mobile-menu toggle,
and script count) with a known-real tool as an in-run control, and should carry the zero-form
tools as fixtures so the new probe is proved against the case that broke the old one.

**A named fixture, measured here 2026-09-03:** `loanzy.uk/tools/loan-vs-savings/` serves a working
tool at **0 forms / 7 inputs / 1 select / 6 scripts**. Independent second confirmation of the
portfolio lane's finding, from a different site. `idea.uk/tools/funding-fit/` is the positive
control at 2 forms / 40 inputs / 8 scripts.

## §6 Traps this lane paid for — read before touching anything

- **RUNBOOK §8b** — do not verify with a re-render.
- **RUNBOOK §10** — 729's apply preconditions; and while 729 is applied, `720_ROLLBACK` refuses by
  design (LANDMINES entry; unwind newest-first).
- **RUNBOOK §1** — copy the guard's predicate; do not paraphrase it.
- **016b §9 ×2** — the measurement pattern, and "a correct predicate wrapped in untested
  inferences" (a check can fire on the right rows and tell the operator something false; no test
  sees it because every test asserts the predicate).
- **`WRONG_CALLS.md`** — six entries from this lane today, all under my own name. The recurring
  one: *the predicate was right, and every sentence I wrapped around it was an untested inference.*
- ⚠ **`default_config::text LIKE '%…"…"…%'` WILL NOT FIND A LITERAL CONTAINING DOUBLE QUOTES**,
  and it fails as a clean `false` rather than an error. `::text` is the JSON *serialisation*, so an
  embedded `"` is stored escaped as `\"`. Measured 2026-09-03 on the very literal 729 defends:
  `position('may also carry a "subject"' in (default_config #>> '{workflow,steps,plan_site,config,prompt_template}'))`
  = **28860** (present), while
  `default_config::text LIKE '%may also carry a "subject"%'` = **false** on the same row, same second.
  **Extract the field with `#>>` first, then search it.** 729's verify block already does; a
  hand-run spot check very easily does not, and would report a defended sentence as missing.
- ⚠ **DISPATCH: "due" is not "will run". A site with ANY claimed item is invisible to
  `find_dispatchable_site` entirely**, so a newly-due high-priority item waits for the current
  batch on that site to drain rather than for its own `retry_after`. From the
  `portfolio_positioning` lane, which lost two predictions to it. **This bears on §1's repair:** if
  you fire work items on a busy site, do not predict timing from `retry_after`, and do not read a
  quiet period as "nothing tried" — it may be "the site is not selectable yet". It is also part of
  why this lane's demand control saw so little dispatch at shell pages.
- ⚠ **A CLAIM'S STATUS DIES IN THE COMPRESSION.** The rule this lane learned the hard way is that
  a peer's finding carries a status — *observed* or *reasoned*, *at the rows* or *at the artefact*
  — and you inherit that status when you adopt it. Their sharper version, worth having verbatim:
  *"my message said 'two rows share position 2 in the database', and the short version of that is
  'the tool and the prose collide', which is already a claim about a visitor. **The rewrite is
  where the qualifier dies.**"* So the dangerous moment is not receiving a finding, it is
  summarising it — including into a commit message or a handoff line.
  **The operational form, theirs, and the one to actually use:** *before restating someone's
  finding, ask what INSTRUMENT produced it and keep that instrument in the sentence — a database
  reading, a body probe, a log line. If the compression cannot carry the instrument, it is too
  short to send.* Worked both ways between two lanes in one afternoon: a database reading became
  "the tool and the prose collide" (a claim about a visitor), and it took a `curl` three hours
  later to retract it.
- ⚠ **TWO LIVE SESSIONS CARRIED THE NAME `bugs_open/450` on 2026-09-03**, and a peer's message to
  this lane was delivered to the other one and bounced. **A bare name is not an address when it is
  ambiguous** — `ListAgents` prints a `[ref]` per row; use it. More generally, addressing "the most
  recently active session with that name" **actively selects the wrong one when the right one is
  quiet**, which is the normal state for a lane mid-council or blocked on a permission decision.
  Full pattern (four instances, three lanes, one day): `016b` §9, *a recency-ordered lookup returns
  something plausible for the wrong identity*.
- Timestamps here are **UTC from the database clock**. `agent_error_log` has **no `created_at`** —
  `\d` it before querying.
