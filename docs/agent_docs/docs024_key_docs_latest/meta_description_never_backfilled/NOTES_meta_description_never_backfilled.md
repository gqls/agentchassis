# NOTES — `pages.meta_description` never backfilled

Append-only, newest at the bottom.

---

## 2026-08-19 — picking up `bugs_open/309`'s case repair, and finding it is not what it says

**Where I started.** Handed the `bugfix_284` lane's two handoffs. Both say the 284 lane
is CLOSED and hand on `bugs_open/309`, whose §9/§10 are the live text. §10's remaining
step: backfill `meta_description` on five fundamentallyai.com pages, *by the framework,
not by hand*, then re-dispatch the rerender.

**Re-measured before trusting anything** (the handoff says to, and it was right):

- served `platform-log/index.html`: HTTP 200, 32,594 bytes, **6 cards, 0 anchors each** —
  unchanged from 08-18.
- the five pages: all still `meta_description` length **0**.
- **Control, and it is what makes the zero readable**: the three `/guides/tool-*`
  siblings serve a real description (156 chars on
  `tool-automation-savings-estimator-guide`), and all six pages return 200 with 27–29 KB
  bodies. So these are not 404s and not a fetch artefact — the trap 309 §C already
  logged once ("a zero from a guessed URL is a false zero").

### Misstep 1 — I nearly accepted §10's routing without reading the check

§10 says the five have not self-healed because `head_essentials_missing` carries no
handler on all 606 rows, and routes that to `bugs_open/083`. That is a clean, plausible
story and I was one step from repeating it.

**It is wrong.** `HeadEssentialsMissingCheck` asserts exactly three things — non-empty
`<title>`, a skip-link, a `<footer>` (`check_site_structural_validity.go:991-1044`; the
`headEssentials()` helper is nine lines and reads `doc.Find("title")`, `doc.Find("footer")`
and an `href="#content"` match). **It never looks at a meta description.** So it would
not have fired for these pages whatever `083` does, and fixing `083` would not fill them.

The check that caught it was cheap and I should have run it first: **grep the check's own
source for the field before believing a claim about what it detects.** One command.

### Then the size turned out to be wrong too

`[MEASURED 2026-08-19]` `pages` where `status='active'`:

| | |
|---|---|
| active pages | **731** |
| empty `meta_description` | **407 (55.7%)** |
| sites with at least one | **26 of 27** |
| sites at 100% | `loancalculator.co.uk` 43/43, `adversecreditmortgage.co.uk` 19/19, `loanzy.uk` 13/13 |

So the five are ordinary members of a fleet-wide class, not a local content gap.

### The mechanism

Every writer of the column is a create-or-upsert path (`upsertPage`,
`create_tool_component_action`, `deploy_tool_action`, `create_report_page_action`,
`ApplyAdoptionPlanAction`, `adopt_verbatim`, `cmd/webdesignport/import`). A
multiline-aware scan for `UPDATE pages`/`INSERT INTO pages` blocks containing the column
returns **no UPDATE path at all**. Nothing repairs an existing page.

Nor can a writer agent reach it: `llm_fields` / `llm_field_specs` are **per-section**
(`plan_sections_action.go:864-897`), scoped to a component's `input_schema`, and
`meta_description` is a **`pages` column**. Exactly the bugfix-203 shape — *ask who owns
the field before asking the framework for it.*

And of the **58** checks under `discovery_checks/`, **none names the column.**

### The live proof, which is the part I would not have believed from code alone

`content_rewrite` items **are** filed about missing meta descriptions and routed to
`page-build-handler`. Two are `complete`:

| item | target | completed | `meta_description` today |
|---|---|---|---|
| `ec701bb3-…` *"neither page carries a meta description"* | gaswholesalers.com `about` | 2026-08-15 16:30Z | **0** |
| `13522562-…` *"the home page meta description … leads with a self-description of inventory"* | webdesign.co.uk `index` | 2026-08-15 19:15Z | **0** |

The gaswholesalers page's own `updated_at` is 19:46Z that day — **the handler ran and
touched the page, and the column stayed empty.** A `complete` work item is not a
repaired artefact. This is what rules out the obvious unblock: filing a `content_rewrite`
for the five would complete and do nothing.

### The finding that changes the fix ranking

I assumed this was historical debt to be backfilled. It is not.

`[MEASURED 2026-08-19]` active pages by creation month, % born with an empty description:
2026-02 **94.1%**, 03 **86.0%**, 04 **50.7%**, 05 **90.9%**, 06 **38.5%**, 07 **54.4%**,
**08 52.7%**. August, by page_type: `content` **60 of 64**, `landing` 16/17,
`blog-post` 18/68, `tool` 15/53.

**The tap is still running.** A backfill alone would be mopping under it, so any option
put to the owner has to say what it does about birth as well as about the 407.

### Misstep 2 — my first diagnosis run failed and reported `complete`

Fired `090` at 09:36Z, run correlation `d7a9ab39-4917-4479-b1d5-68aae982f79c`. It ran
`lookup_symbols` → `assemble_bundle` → **FAILED**, and the orchestration's last row then
reads `current_step=complete, status=COMPLETED`. Zero artifacts.

The real error is only in `collected_data->>'__step_error'` on **that last row** — the
two FAILED rows carry `(none)`:

```
diagnose_assemble_bundle: no scope (tried "route.scope.Symbols",
"input_data.seed_scope", then code_results)
```

Cause: my symptom named **files**, not **symbols**, and nothing seeded a scope. The
trigger takes `SEED_SCOPE="path:Symbol,path:Symbol"` (090 script, lines 117 / 282-346)
and I had not read that far. Re-fired 09:47Z with five `path:Symbol` entries —
run correlation **`7375631a-dfa1-4145-9a07-d13586f1a7cf`**.

⚠ **Two things worth carrying:** a failed diagnosis run's terminal row says
`COMPLETED`, so "did my diagnosis produce anything?" is `count(*) FROM
diagnosis_artifacts`, never the orchestration status. And a `090` symptom that names
only paths can be accepted, dispatched and then fail to scope — pass `SEED_SCOPE`.

---

## 2026-08-19, later — the diagnosis came back UNVERIFIABLE, and was more useful than a CONFIRMED

Run `7375631a-dfa1-4145-9a07-d13586f1a7cf` reached **`status: UNVERIFIABLE`**, stopped on
the **iteration cap** after 5 bundles. **It did not confirm my hypothesis and I am not
going to write it up as though it had.** What it did was better: it refused my framing
and handed back a sharper one plus the exact test that would settle it.

Its last hypothesis, in its own words:

> *"meta_description is NOT frozen at creation. upsertPage's ON CONFLICT DO UPDATE sets
> meta_description = EXCLUDED.meta_description unconditionally on every upsert of an
> existing page, with no COALESCE/NULLIF guard the way nav_label has … any resync/replan
> pass whose page map omits that field will overwrite an existing, previously-populated
> meta_description with an empty string — the opposite failure mode from 'set once at
> creation and never touched again'."*

And its discipline, which I should copy:

> *"The code proves the mechanism EXISTS … What is missing is proof the mechanism
> OCCURRED … consistent-with is not direct."*

It also recorded that its own join of `site_plan_pages` against `pages` **returned 0 rows
and that this was inconclusive**, not a refutation — a distinction I have got wrong
before. Then it named the settling evidence: an earlier `site_snapshots.pages_snapshot`
entry showing a non-empty description for a page now empty.

### I ran that test. The overwrite has FIRED.

`[MEASURED 2026-08-19]` `site_snapshots.pages_snapshot` does carry `meta_description`
(the key list per page element includes it):

| domain | page | snapshot | chars then | chars now | page updated |
|---|---|---|---|---|---|
| robot-hands.com | `index` | 2026-04-10 | **120** | **0** | 2026-08-19 |
| robot-hands.com | `product-detail` | 2026-04-10 | **97** | **0** | 2026-08-17 |
| robot-hands.com | `tool-gripper-payload-calculator-guide` | 2026-04-10 | **169** | **0** | 2026-08-17 |
| robot-hands.com | `tool-gripper-payload-calculator` | 2026-04-10 | **329** | **0** | 2026-08-17 |

**All four are `built_from_plan_version IS NOT NULL`** — plan-managed, i.e. reachable by
the one writer that overwrites unconditionally. `index` was updated at 02:40Z **today**,
so this site is actively being replanned.

### The denominator, because 4 on its own is not a rate

| | |
|---|---|
| pages covered by any snapshot | **139** (7 sites, most snapshots 2026-04-10) |
| of those, had a description then | **30** |
| **lost it** | **4** (13% of the 30) |
| gained one | 10 |

⚠ **This sample cannot be extrapolated to the 407.** It is 139 of 731 pages, 7 of 27
sites, and almost all of it is one day in April. What it establishes is **existence** —
the overwrite is not merely representable, it has happened to four named pages — and
nothing about scale. Anyone quoting "13%" as a fleet rate is quoting a 30-page
denominator from four months ago.

### So there are TWO mechanisms, not one, and I had only found one

- **M1 — born empty.** `build-site-planner`'s `Return JSON:` page object asks for `name`,
  `title`, `page_type`, `nav_label`, `nav_order`, `in_header`, `in_footer`, `sections`
  and **no description field**; `upsertPage` reads
  `GetStringField(page,"meta_description","")` and takes the empty default. Certain from
  the live agent row + the code. Explains the bulk (`content` 81.4%, `landing` 85.3%).
- **M2 — blanked later.** The same upsert's `ON CONFLICT … DO UPDATE SET
  meta_description = EXCLUDED.meta_description`, unguarded, against `nav_label`'s
  `COALESCE(NULLIF(pages.nav_label,''), EXCLUDED.nav_label)` two lines above. Proven to
  have fired on the four pages above. **Unsized.**

M2 is the one I would have missed, and it matters for the fix: an option that only teaches
the planner to emit a description leaves the unguarded overwrite in place, so any later
pass that omits the key still blanks the field.

### Misstep 3 — I was ready to call M2 "the likely explanation" off the code alone

I had spotted the missing `COALESCE` myself and written it up as a code-certain fact with
the firing marked `[UNMEASURED]`. That was the right marker, but I had stopped there and
would have shipped the file that way. The loop's "consistent-with is not direct" is what
sent me to `site_snapshots` — a table I had not thought to look for, and which settled it
in one query. **The check to carry: when you mark something `[UNMEASURED]`, spend one
minute asking what evidence WOULD measure it before accepting the marker as the answer.**

---

## 2026-08-19, later still — owner chose the FULL fix, and what got built

Owner ruling 2026-08-19, on the four options in `bugs_open/320` §8: **everything,
including a backfill producer.** So the lane stopped being a report and became a build.

### The root cause of M1, found after the bug was already filed

I had `320` written and committed saying "the planner is never asked". I had not yet
found *why*. It is one omission, and it is exact: `build-site-planner`'s `plan_site`
step gives the model a `Return JSON:` template whose page object is

```
name, title, page_type, nav_label, nav_order, in_header, in_footer, sections
```

**and no description field.** `upsertPage` then reads
`GetStringField(page,"meta_description","")` — asking the plan for a key the plan was
never told to produce, and taking the blank.

⚠ **The check that hides this:** `default_config::text ILIKE '%meta_description%'` on
that agent returns **TRUE**. The planner does contain the string — in
`load_existing_pages`, a `query_database` step that SELECTs the column. Matching the
string proves the agent *mentions* the field, not that it is *asked* for one. I ran
that grep early and it is why I nearly stopped at "the planner doesn't do it, somehow".
Read the output schema, not the census.

### What shipped

| piece | what | state |
|---|---|---|
| M2 guard | `upsertPage` ON CONFLICT → `COALESCE(NULLIF(EXCLUDED.meta_description,''), pages.meta_description)` | committed `aeccfc595`, **rides the next fleet roll** |
| M1 fix | migration `485` — the planner's page object gains the field + an authoring rule | **APPLIED + ledger-recorded 2026-08-19**, config is live on apply |
| the missing mechanism | `save_page_meta_description` (register **SEO-004**) — the persist half only | committed `aeccfc595`, **rides the roll** |
| the driver | migration `486` — `meta-description-backfiller` agent | **HELD**, see below |

**Why only the persist half.** Finding the pages is `query_database` and writing the
sentence is `execute_llm_prompt`; both already exist. Building a monolithic Go action
that did all three would have re-implemented two working things and taken authorship
away from the framework, which is the 2026-08-06 ruling's whole point.

### `486` is HELD, and the suffix is the control

CLAUDE.md: *image first, then seeds — a seed naming an unregistered action fails at
runtime.* `[MEASURED 2026-08-19]` the live chassis is `v1.0.1314`, revision
`d3590ca4638d…`, and `git merge-base --is-ancestor aeccfc595 d3590ca4` is **FALSE**
(145 commits unshipped). So the action does not exist in the running binary.

A banner would not have held it — **a migration's guard checks DRIFT, not ORDER**. The
file is named `486_..._HOLD.sql` so the runner's `SIDECAR_RE` excludes it from
`--apply` while still listing it. Verified rather than assumed: `--no-probe` shows it
under *"Sidecars (hand-run only, NOT applied by this runner)"* and not in Pending.

### Misstep 4 — a test I wrote CLAIMED a mutation would kill it, and the mutation passed

The first version of `save_page_meta_description_test.go` matched
`regexp.QuoteMeta("UPDATE pages")`, and its comment asserted that deleting the
`AND ($3::bool OR COALESCE(meta_description,'')='')` guard would fail it.

**I ran the mutation. All five tests stayed green.** With the clause deleted the action
still passed three arguments, so `WithArgs(..., false)` still matched; sqlmock does not
care that `$3` became unreferenced. The overwrite policy could have been moved out of
the SQL into nothing at all and the suite would have applauded.

Fixed by matching the clause itself (`metaDescGuardClause`). Re-ran: the mutation now
fails three tests, and restoring makes them green.

**The lesson is not "write better tests" — it is that the comment was the error.** I had
written a MUTATION THAT KILLS IT line as though stating it made it true. Running the
three named mutations took about two minutes and found that one of three was fiction.

⚠ Related, and worth its own note: **the upsert guard had NO test at all before this.**
The whole `actions` package passed both before and after the M2 fix. That is why
`upsert_page_meta_description_guard_test.go` exists — and its mutation (restore the
unguarded clause) was run too, and does fail.

### Provenance discipline on the diagnosis run

`320` and this file both say the `090` run came back **UNVERIFIABLE** on its iteration
cap. It is cited as a **contributor** — it supplied M2 and named the `site_snapshots`
test — and never as ratification. Every figure in `320` is first-hand and carries its
query or `file:line`.

---

## 2026-08-19 — council round 1: REVISE, and it found a real defect

Correlation `46734ae9-91c5-47d6-9a8a-4cd1fa213d21`. Decision **REVISE**, gated by
`bug_historian` (HIGH), seconded by `guardian` (HIGH). 11 of 16 seats approved.

### Misstep 5 — I fixed ONE of FOUR write paths and called M2 closed

The objection, near enough verbatim:

> *Landmine on record: 'There are THREE `pages` upsert helpers and they have OPPOSITE
> collision policies'. This plan patches ONE. This is exactly the recurring shape at
> 016b §9: 'one call site of a shared judgement gets the rigorous fix; the sibling stays
> heuristic.'*

I checked instead of arguing, and **there were three more**, each with the same
empty-default upstream:

| site | why its value can be `""` |
|---|---|
| `apply_adoption_plan_action.go:546` | `metaDesc, _ := pm["meta_description"].(string)` — a bare type assertion |
| `adopt_verbatim.go:470` | `extractHTMLMetaDescription` **explicitly `return ""`** when the source HTML has no meta tag |
| `cmd/webdesignport/import.go:182` | `p.MetaDescription` is `""` for an imported page carrying none |

`adopt_verbatim` is the one I would bet actually fired: **re-adopting a page whose source
HTML lacks a description blanked a good one.** All four now carry the guard; the sweep
`grep -rn 'meta_description = EXCLUDED' --include='*.go' . | grep -v COALESCE` returns
**0 live sites**.

**The check I skipped, and it is in MEMORY under my nose:** *grep LANDMINES for the
SYMBOL you are about to trust — the SessionStart hook only matches files already DIRTY,
so a shared helper is never shown.* `site_db_actions.go` was not dirty when I started, so
the hook showed me nothing, and there were `doc_notes` landmines keyed
`pages.meta_description` from **08-14, 08-15 and 08-18**. One grep before touching the
file would have handed me the sibling list in the first round.

### And the prior art I missed entirely

`idea_uk_vm_site/sql/2026-08-15_fix_head_title_meta.sql:40`, four days before I filed
`320`:

> *"pages is a materialised cache; site_db_actions.go:1173 re-upserts
> meta_description = EXCLUDED.meta_description unconditionally, so 6a alone regresses on
> the next plan sync."*

Another lane had the mechanism, in writing, in a committed file. They were fixing copy on
two sites and noted in passing that the cache would regress. I filed it as a new
discovery. **The finding stands, and it was not new.**

That prior art also gave me something round 1 lacked: `pages` is a CACHE, and
`site_plan_pages` is the SOURCE. Which lets me verify `485` end to end rather than
assert it — `write_site_plan_action.go:535` reads
`GetStringField(raw,"meta_description","")`, **exactly the key 485 adds to the planner's
template**, and `:631` inserts it into `site_plan_pages`, from which `upsertPage` binds
it into `pages`. That was round 1's very first reviewer check, and I could not have
answered it then.

### What I could answer with evidence

- **`UpsertPageForRole` is safe by construction** — it rewrites only columns the caller
  names in `Refresh`, and all five live callers name `url`/`title`/`sections` or `{}`.
  **Enumerated, because asserting it without the query is itself the objection.**
- **Sibling columns** (medium): `meta_description` was the **only** column in
  `upsertPage`'s clause with an empty default — `name`→`page-N`, `title`→`name`,
  `url`→computed, `page_type`→`"content"`, `nav_label`→`title` and already guarded. That
  asymmetry is *why* this column and no other produced the defect.
- **`103` is CLOSED**, checked at the path
  (`bugs_closed/103_HANDOFF_2026-07-27_…`). The objection's premise that the index lists
  it as open is mistaken — noted rather than accepted, since a wrong status accepted
  quietly is how the next reader inherits it.

Round 2 resubmitted on the SAME correlation so the trail accumulates.

**The tally that matters: two council rounds, and round 1 found a defect that would have
shipped as "M2 closed" while three paths still blanked descriptions.** That is the second
time this lane's confident conclusion needed an outside check — the first was the `090`
loop refusing my "frozen at creation" framing.
