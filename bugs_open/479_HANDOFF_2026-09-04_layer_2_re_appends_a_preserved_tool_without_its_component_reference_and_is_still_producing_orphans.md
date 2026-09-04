# 479 — Layer 2 re-appends a preserved tool WITHOUT its component reference, and is still producing orphans

**Filed 2026-09-04**, by the `bugfix_450_tool_page_shells` lane, found while scoping bug 450's
§1 orphan repair fleet-wide. **The writer is `save_page_sections_action.go`'s Layer 2 re-append
arm — the same arm `bugs_open/385` diagnosed on 2026-08-25**, whose fix (`a799579fd`, live
2026-08-26) corrected the arm's MATCHER and left its identity drop untouched.

> **This is not a re-filing of 385.** 385's symptom was a byte-identical DUPLICATE beside a
> locked row, and its own discriminating query excludes every row in this file: `[MEASURED
> 2026-09-04]` all 17 current orphans have `byte_twins_on_page = 0` and `locked_same_slot = 0`.
> 385 §5c joint 5 *named* the identity drop in passing and correctly treated it as incidental to
> the duplication it was explaining. This file is about the drop on its own, which is a larger
> population, is still firing, and has a different fix site.

## 1. The defect, in one sentence

When a rebuild's composition does not name a slot at all, Layer 2 re-appends the stored
interactive section built wholly from the stored row — its bytes, its `content_data`, its slot
name, its provenance stamp — and drops the one remaining field, `component_id`, so a page serving
a 13–25 KB working tool carries no reference to what produced it.

## 2. Mechanism, at the lines

`save_page_sections_action.go`, Layer 2 "slot dropped entirely" arm:

```go
sections = append(sections, SectionData{
    ComponentName:      p.slot,
    HTML:               p.html,
    ContentData:        p.contentData,
    ComponentVersionID: p.componentVersionID,
    ComponentID:        carriedIdentity(carryStoredIdentity, p.componentID, p.componentFunction),
    Position:           len(sections) + 1,
})
```

`carriedIdentity` (`adopt_fragment_section.go:173-181`) returns `""` unless **both** the
`adopt_unidentified_fragments` flag is armed **and** the stored component's function is exactly
`adopted-fragment`:

```go
if !armed || storedComponentID == "" { return "" }
if storedComponentFunction != adoptedFragmentFunction { return "" }
```

**A real tool's function is never `adopted-fragment`, so this arm writes NULL unconditionally —
arming the flag would NOT fix it.** That is worth stating plainly because the flag's own doc
comment reads as though it governs this case.

`enrichSectionsWithComponentIDs`, the pass that resolves a component from a name, runs ~190 lines
EARLIER (at the `sections` assembly, before the Layer 2 block). It never sees the appended entry,
so nothing resolves it afterwards. The INSERT then writes `componentIDPtr = nil`.

**The arm's signature, and it is falsifiable:** a re-appended section is `Position: len(sections)+1`
and the insert loop numbers from `i+1`, so the orphan must be at its page's LAST position and must
be the LAST row inserted in its write burst. `[MEASURED 2026-09-04]` **5 of 5** tool orphans are
exactly that — last position, last `created_at` in a burst of rows milliseconds apart. The
prediction could have come out otherwise.

## 3. Population — dated, because a census goes stale by ADDITION

`[MEASURED 2026-09-04 ~09:30Z]` `page_components` rows with `component_id IS NULL` and
`build_status <> 'removed'`: **17 rows across 7 sites** (11 on 2026-08-24, per 385 §4).

| kind | rows | where |
|---|---|---|
| **tool slots** | **5** | `advertise.co.uk` ×2, `finetuning.uk/playground`, `idea.uk/tool-funding-fit`, `loanzy.uk/tool-loan-vs-savings` |
| blog/article listings | 8 | boxingonline `articles-index` ×6, finetuning `blog`, ai-agent-orchestration `blog` |
| content sections | 3 | finetuning `our-position-on-ai` |
| a 25,901-byte `game` slot | 1 | `gamesdesign.co.uk/game-jelly-invaders` |

**THE CLASS IS STILL PRODUCING — this is not a backlog.** Three of the five tool orphans were
created AFTER 385's matcher fix went live on 2026-08-26:

```
advertise.co.uk  tool-ad-budget-calculator          2026-09-03 16:23:18Z
advertise.co.uk  tool-cpm-cpc-benchmark-comparator  2026-09-03 16:57:03Z
finetuning.uk    tool-playground                    2026-09-04 04:12:34Z
```

The re-append itself is CORRECT in these cases — e.g. `advertise.co.uk/tool-ad-budget-calculator`'s
plan named `hero-tool · tool-guide-intro · tool-ab-test-calculator · tool-cta`, which genuinely
does not include the page's own tool, so 385's identity arms rightly found no match. **385 fixed
false-negative matching; it did not touch what the arm then writes.**

### 3a. ⚠ How large the exposed population actually is, and why 385's number misleads

385 §5c closes with `[MEASURED 2026-08-25] the armed set — locked AND build_status='deployed' AND
interactive — is 1 row fleet-wide`. That is true of **385's** symptom, which needed a locked row.
**The orphaning arm needs no lock.** Layer 2's preload is:

```sql
WHERE pc.page_id = $1 AND pc.build_status = 'deployed' AND <interactiveHTMLSQL(pc.rendered_html)>
```

`[MEASURED 2026-09-04]` that predicate selects **378 rows across 371 pages** fleet-wide. (The
boolean was printed from `interactiveHTMLSQL` itself via a throwaway test rather than transcribed —
RUNBOOK §1's rule; four measurement errors in the 450 lane came from paraphrasing a predicate.)
Being preloaded is necessary, not sufficient — the slot must also fail to match the incoming set —
so **371 pages is an upper bound on exposure, not a prediction.** The lower bound is the observed
rate: **3 new orphans in the 19 hours before filing.**

## 4. The harm, and it is NOT "the tool disappears"

All five tool pages serve working tools **today** `[MEASURED 2026-09-04, at the body, controls
held]`:

```
advertise.co.uk /tools/ad-budget-calculator/          200  2 forms  22 inputs  1 select  7 scripts
advertise.co.uk /tools/cpm-cpc-benchmark-comparator/  200  2 forms  11 inputs  2 selects 7 scripts
finetuning.uk   /playground.html                      200  1 form    1 input   0 selects 8 scripts
idea.uk         /tools/funding-fit/                   200  2 forms  40 inputs  0 selects 8 scripts
loanzy.uk       /tools/loan-vs-savings/               200  0 forms   7 inputs  1 select  6 scripts
```

(`scripts/probe-page-url.sh` for the recorded-URL half; invented-URL control 404, known-good
sibling 200.) The damage is **prospective, on the re-render path**, and it has two shapes:

1. **SILENT SUBSTITUTION — the one that bites.** With `component_id` NULL, `resolveComponent`
   (`rerender_page_sections_action.go`) falls through to the slot-NAME map, which
   `loadComponentSchemas` builds as `result[ci.Function] = ci` over a query with **no `ORDER BY`**,
   last write wins. `[MEASURED 2026-09-04]` three of the five orphaned tool slots have **2–3
   ACTIVE** `content_components` sharing their `function`:
   `tool-ad-budget-calculator` → 2, `tool-cpm-cpc-benchmark-comparator` → 3,
   `tool-loan-vs-savings` → 2 (several are cross-site forks, e.g.
   `…-advertise-co-uk-websitepromotion-co-uk`). **Which component renders the page is decided by
   row order.** That is exactly the silent substitution `bugs_open/182` exists to prevent,
   arriving one field along.
2. **A STUCK PAGE.** Where the name does not resolve to a self-contained tool, the orphan's NULL
   `content_data` (all five: NULL) trips the re-render content pre-check, which escalates the
   WHOLE page to the writer and returns without rendering. The bytes survive; the page can never
   be re-rendered. This is 385 §3's outcome reached by a different route.

**⚠ Do NOT repeat "one rerender away from losing their tools" (450's HANDOFF §1(3)).** It was a
reasonable inference and it is wrong in both directions: the pre-check will not blank a section
with no `content_data`, and the real risk is a *different tool* rendering, not an empty slot.

## 4a. ⚠ THE PREDICTED HARM HAPPENED, on one of these very pages, the same afternoon

§4(1) was written at ~11:00Z as a prediction: with `component_id` NULL the re-render falls through
to the slot-NAME map, which has no `ORDER BY`, so with 2-3 active same-function forks **which tool
renders is decided by row order**. At **14:05:08Z** — after that was written, before the fix rolled
at 16:01Z — it happened.

`advertise.co.uk/tool-ad-budget-calculator`, slot `tool-ad-budget-calculator`:

| | this morning 09:30Z | now 16:0xZ |
|---|---|---|
| `component_id` | NULL | `tool-ad-budget-calculator-advertise-co-uk-**websitepromotion-co-uk**` |
| `rendered_html` | 16,953 B | **17,238 B** |
| `name="industry"` on the select | absent | **present** |

**`name="industry"` is the fork's distinguishing attribute** — it is in the websitepromotion
template and is NOT in `tool-ad-budget-calculator-advertise-co-uk`. That is the discriminator §6
below establishes at the bytes, and it says the page was re-rendered from **another site's
component**. `[MEASURED 2026-09-04 16:0xZ]` advertise's own component now has **zero** users on
any site, and the fork serves **both** advertise.co.uk and websitepromotion.co.uk.

The page still serves a working tool (2 forms / 22 inputs / 8 scripts, 200, controls held), which
is exactly why this is worth writing down: **nothing about the artefact would tell you it changed
owner.**

- The re-bind to the fork and the byte change are `[MEASURED]`.
- That the writer was the re-render name-fallback specifically is `[INFERRED]` — consistent with
  every column, and with `adopt_fragment_section_test.go`'s own note that *"an unadopted row named
  `hero` resolves to the hero component there and the fresh-render entry re-emits hero's id —
  which re-binds it on the next save"*, but I did not catch the write in the act.

**Two consequences for whoever picks this up:**

1. **The repair population has CHANGED and shrunk for the wrong reason.** 5 tool orphans this
   morning → **3** now (`advertise/tool-cpm-cpc-benchmark-comparator`, `idea.uk/tool-funding-fit`,
   `loanzy.uk/tool-loan-vs-savings`). Two left the census: `finetuning.uk/playground` re-bound at
   11:57Z to `tool-playground-finetuning-uk`, which is **correct** (one candidate, so the name
   match was unambiguous); `advertise/tool-ad-budget-calculator` re-bound at 14:05Z to the **wrong
   fork**. A shrinking orphan count is not progress — **check what each departure bound to.**
   `advertise/tool-cpm-cpc-benchmark-comparator` was itself rewritten at 14:21:42Z and came back
   NULL, so it has been round the loop and is orphaned again.
2. **A WRONGLY-BOUND row is not repaired by this bug's fix or its repair script.** Both act on
   `component_id IS NULL`; a row bound to the wrong fork has a non-NULL id and is invisible to
   both. It needs correcting, not binding, and nothing currently detects it. **That is an open
   residual with no owner** — see §6.

## 5. Fix — BUILT, submitted, committed 2026-09-04

**Split the two carry arms on the identity axis.** The splice arm keeps the narrowed, opt-in
`carriedIdentity` the council required; the re-append arm keeps the stored row's own
`component_id`, when that component is still active.

```go
ComponentID: reappendedComponentID(p.componentID, p.componentActive),
```

**Why this is not the widening three council seats rejected**, stated up front because it looks
like it. `adopt_fragment_section_test.go:249` records the objection verbatim: *"Carrying identity
for every interactive section is broader than the diagnosed bug and would silently keep a
legitimately-typed component at its OLD identity when a plan intended to swap it — three council
seats made that point independently."* That is correct, and it is about the **SPLICE** arm, where
an incoming section from the plan is present and expresses an intent. The **RE-APPEND** arm fires
only when the plan named the slot **nothing**: there is no incoming section, no intent to
override, and the appended row is a verbatim copy of a stored row whose id the copy was dropping.
`TestLayer2_SplicedToolStillTakesThePlansComponent` pins the boundary so the rejected version
cannot arrive later by accident.

**The `active` guard is load-bearing, not tidiness:** a DANGLING id is worse than NULL, because
`loadComponentSchemasByID` drops a row it cannot load and `resolveComponent` then returns
`invalidTemplate` rather than falling through to the name map — which fails the whole page. NULL
at least falls through.

Both mutations were run and both go red: restoring `carriedIdentity` in the re-append arm, and
making the helper ignore the active flag.

- Council: `Council-Submitted: 567954f2-1b40-407a-b7b1-d3b299a0af9a` — **⚠ NO VERDICT, and it is
  not coming.** The run reached `council_decide` and died with *"no reviewer produced a readable
  opinion (6 abstained, 11 unreadable)"*, which reads like a submission defect and is not one: all
  11 seats failed at the API with `400 … "Your credit balance is too low to access the Anthropic
  API"`. `[MEASURED 2026-09-04 11:45Z]` the fleet's last successful LLM call was **11:20:49Z** and
  this submission landed 47 seconds later. **RESUBMIT once credits are restored** and record the
  new correlation here; `098` can never credit the current one. Do NOT rewrite the submission to
  make it smaller — it passed admission and all 17 seats dispatched.
- **RESUBMITTED 2026-09-04 16:07Z, unchanged: `Council-Submitted: 2ef3a34b-284b-46ce-82f4-160fbfb73f54`**
  (run correlation `4cb20d04`, orch `39eb7518`). Held until after the v1.0.1361 roll rather than
  fired on credit restoration alone, because **a roll kills an in-flight council** — dispatching
  into the roll window would have lost a second run to something outside the submission again.
- **LIVE, and proven rather than inferred.** Both `agent-chassis` pods (started 16:01:26Z /
  16:01:53Z) state `build provenance git_commit 06c0b18f233bc600918ef481d32b40f29535f78f`, and
  `git merge-base --is-ancestor 2fae8baa4 06c0b18f2` is true. **Verified by ancestry, NOT by
  probing for a literal:** this change adds no reachable string, and unreachable code is dropped
  by the linker, so a binary probe would read ABSENT with perfectly clean controls.
- `verify-head-builds.sh --with` (4 files) → `OK — HEAD b9cdfd5d1 builds`.
- ⚠ Two package tests fail at HEAD **before this change** and are the 440 lane's, not this one's:
  `TestFindingCodeScanEveryWriteIsRegistered` and `TestTemplateExecutorsAreDeclared`, both about
  `fail_work_item_message_template.go` (committed `83407cd37`).

### 5a. Council APPROVED (`2ef3a34b`) — and what the objections turned up

`decided_by: "approved with 1 advisory objection(s) — none high-severity"`. 10 seats approve, 1
objects; 5 objections (2 medium, 3 low). Actioned rather than noted:

**MEDIUM (guardian) — "SavePageSectionsAction has more than one caller; did you check whether any
OTHER consumer assumes a re-appended interactive section always carries `component_id` NULL?"**
Checked, and it found something better than a risk. There IS another consumer keying on that
column, and it is a **WRITER**: `discovery_checks/check_unlinked_components.go`
(`UnlinkedPageComponentsCheck`), a self-healing sweep that links orphaned `page_components` rows
automatically. **This bug's entire population is structurally invisible to it**, on two
independent conditions:

```sql
AND pc.rendered_html LIKE '%data-component="%'   -- requires a self-describing attribute
JOIN content_components cc ON ... AND cc.forked_from IS NULL   -- base components only, no forks
```

`[MEASURED 2026-09-04]` **0 of 15** current orphans carry a `data-component` attribute, and this
bug's tool components are per-site FORKS. So the estate has had an auto-repair for unlinked
components all along and it has never been able to see these rows — which is why the population
accumulates instead of self-healing, and it is a better answer to "why did nobody notice" than
§3a's narrowing alone. ⚠ **Note also what that check does when it CAN see a row:** an unguarded
`UPDATE` taking `DISTINCT ON (pc_id) … ORDER BY cc.created_at ASC` — the oldest matching active
base component wins, arbitrarily among same-function candidates. `forked_from IS NULL` limits the
damage; it does not make the choice evidenced. My repair script and that check are disjoint in
practice (it needs `data-component`, mine needs `slot_name LIKE 'tool-%'` and a non-fork-owned
candidate), but nobody should assume that stays true.

**MEDIUM (debug_historian) — "the name-fallback map has no `ORDER BY`; any other path leaving
`component_id` NULL still lands on the same non-deterministic fallback. Either add a tie-break or
scope this explicitly as future work."** **Scoping it explicitly, as invited — and it is no longer
hypothetical: §4a is an observed instance.** `loadComponentSchemas` builds `result[ci.Function]`
over a query with no ordering, last write wins. Adding a deterministic tie-break is NOT a small
change: that map is shared with `plan_sections`, so changing which component wins would change
composition fleet-wide and needs its own round with its own blast-radius measurement. **Recorded
as open, in §6, rather than smuggled in here.**

**LOW (guardian) — "a third builder of `layer2PreloadRows` would short-scan silently."**
Checked and refuted: `grep -n "layer2PreloadRows\|NewRows(\[\]string{\"slot_name\"" *_test.go`
shows one definition, one preload builder (`layer2PreloadWithFunction`, plus the new
`layer2PreloadWithInactiveComponent`). Every other `slot_name` fixture in the package is a
different, 2-column query.

**LOW (editquality) — the edit-2 sketch showed the local scan var but not the struct assignment.**
Fair; the shipped code does both. No change.

**LOW — "the 17 existing rows are left unrepaired ('truncation casualties never swept', 016b §9
case 46)."** Correct, and now has a committed remediation path rather than an intention:
`REPAIR_479_reattach_orphaned_tool_component_ids.sql`, rehearsed against the live DB
(`UPDATE 5`, both guards passed, bytes untouched, rolled back), held for the owner because it is a
live write across customer sites. ⚠ **Re-derive the mapping before running it — §4a changed the
population from 5 to 3.**

**Named in `missing`, and honestly outstanding:** no audit of whether producers OTHER than this arm
can also write `component_id` NULL. The 12 non-tool orphans (blog listings, `generic-text-block`,
`faq`, a `game` slot) do NOT match this arm's signature and are very likely a different producer.
Not chased. §6.

## 6. STILL OPEN — the repair, and why it is not a one-liner

**The code fix stops new orphans. It does not repair the 17 existing rows**, and they will not
self-heal: once a row's `component_id` is NULL, the Layer 2 preload reads `''` for it and the next
rebuild re-appends it with nothing to carry. **The state is self-perpetuating.**

⚠ **The obvious repair is UNSOUND.** 450's HANDOFF proposes matching `slot_name` to `cc.function`
and requiring exactly one active component (the shape the `portfolio_positioning` lane used
successfully on the seotools six). `[MEASURED 2026-09-04]` **that mapping is not unique for 3 of
the 5 tool rows** — 2, 3 and 2 active candidates respectively — so a repair written that way
either refuses (safe) or binds an arbitrary fork (**not** safe). It must not be run unguarded.

**What the rows do NOT carry**, checked so the next session does not re-check: no
`data-component` attribute, no `data-tool` attribute, no `component_version_id`, `content_data`
NULL. `page_component_history` cannot help either — its `component_id` is a FK to
**`page_components(id)`**, the *instance*, with `ON DELETE SET NULL`, so it goes NULL when the
instance is deleted and never held the library reference at all. (385 §1's reading of "every one
already `cid=NULL`" in history is measuring that column, not this one.)

**So identification needs evidence, not a naming convention.** The candidate that has not been
tried: render each candidate's `html_template` and compare against the stored bytes — the same
proof standard `adoptFragmentSection` holds itself to. Untried here; flagged rather than guessed.

## 7. Why there is no `090` verdict, stated rather than omitted

Per the owner ruling of 2026-07-31, the substitution is declared. 385 §6 records that a `090` on
this exact symbol returned **`UNVERIFIABLE — stopped: iteration-cap`** with zero non-bundle
artifacts, twice, because `save_page_sections_action.go` is far over the ~60 KB bar in the
landmine (*"a 090 on a symbol in a large file returns bundles and NO verdict"*). The file has not
shrunk. The declared substitute performed instead: the arm read at the lines; a falsifiable
positional prediction stated and checked (5/5); 385's own discriminating query re-run over today's
population (17/17 excluded); the preload predicate printed from the function rather than
transcribed and measured fleet-wide; the served artefacts probed with controls; both fix mutations
run.

## 8. Relations

- **`bugs_open/385`** — same arm, different symptom. Its matcher fix is live and correct; §4's
  narrowing (*"a bare `component_id IS NULL` count … over-reports by 10 out of 11"*) is true of
  duplication and is what has kept this class unread since. A forward pointer has been added there.
- **`bugs_open/450`** — the lane that found this, scoping its §1 repair. 450's HANDOFF §1
  attributes the seotools six to `save_page_sections`' delete-and-reinsert "carrying" a payload
  `ComponentID`; the arm identified here is the specific carrier.
- **`bugs_open/357` / RFC_046** — owns `carriedIdentity` and the `adopt_unidentified_fragments`
  flag. Unchanged by this.
- **`bugs_open/182`** — owns the slot-name-resolution silent substitution that §4(1) is an
  instance of.
