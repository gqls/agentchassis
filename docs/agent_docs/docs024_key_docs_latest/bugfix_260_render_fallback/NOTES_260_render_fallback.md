# NOTES — bugs_open/260, the silent regex render fallback

Append-only, newest at the bottom. Technical log: what was tried, what the system actually
said, and every misstep.

Lane opened 2026-08-19 to fix **260's renderer half** (the silent fallback). The writer-output
half was handed to `copy_quality_two_stage` on 2026-08-12 and stays theirs.

---

## 2026-08-19 — validity re-confirmed, and the file's own figures are stale in BOTH directions

Picked this up cold on owner instruction. First job was deciding whether the bug is still real
before designing anything, because the file's newest measurement was three days old and four
lanes had contributed to it since.

### Ownership, checked before starting

`scripts/who-owns.py 260` names four lanes with commits against the file —
`mortgagecalculator_couk_adoption`, `copy_quality_two_stage`, `portfolio_positioning`,
`brochure_component_library` (the filer), plus `loanzy_uk_example_site`. **Every one of them
states in its own contribution that it is NOT fixing this.** The filing lane's owner log
(`brochure_component_library/README_where_we_are.md`, 08-12 late afternoon) parks the fallback
removal as *"a decision I would like eventually, not now"*. So the fix was genuinely unclaimed.
`component_library.go` was clean in the working tree at the time of writing (`git status
--porcelain` on the five relevant paths returned empty).

⚠ **Ownership checks are LAGGING** — they read commits, so a session mid-fix is invisible.
Re-checked at each phase boundary rather than once.

### The census has moved a long way past §10b

`[MEASURED 2026-08-19]` Isolating this defect from the other eight issue types sharing
`error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'` (a bare count of that code is ~177 and would
badly overstate the bug):

```sql
WITH tmpl AS (
  SELECT DISTINCT e.id, e.domain, e.occurred_at, e.work_item_id
    FROM agent_error_log e, jsonb_array_elements(e.context->'issues') i
   WHERE e.error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'
     AND i->>'type' IN ('unrendered_template','unrendered_template_block'))
SELECT count(*), count(DISTINCT domain), count(DISTINCT work_item_id),
       min(occurred_at), max(occurred_at) FROM tmpl;
```

**26 events · 7 domains · 25 work items · 08-11 15:39Z → 08-18 23:36Z.** §4 recorded 6/4;
§10b recorded 11/4/10 on 08-16. It is accelerating, not settling: 08-18 alone was 9 events
across three domains (mortgagecalculator 5, loanzy 3, webdesign 1).

**24 of the 25 work items sit at `needs_human_review`**, and the type mix matters —
`needs_page`, `unbuilt_internal_link`, `content_rewrite`, `needs_content_page`. See the
misstep below for what I got wrong about that mix.

### The leaked-token discriminator, and its ceiling

`[MEASURED 2026-08-19]` Distinct leaked values per occurrence:

| occurrences | distinct tokens |
|---|---|
| **25 of 26** | `{{end}}` · `{{if ` · `{{.label}}` · `{{range ` |
| **1 of 26** | `{{ variable }}` — webdesign.co.uk, 08-18, item `4d1094c0` |

The `portfolio_positioning` lane traced `{{.label}}` to `mechanism-flow` being the only
planned component whose template emits it, on both of their blocked pages. That inference now
survives contact with six further domains it was never built on. The `loanzy_uk_example_site`
lane's case is a **greenfield** build — zero prior components — carrying the same four-token
set, so an aged or hand-edited component is not a precondition.

⚠ **`[CEILING, NOT COUNT]`** — these value lists inherit `checkUnrenderedTemplates`'s
`FindAllString(html, 10)` cap at `validate_page_content.go:793,804`. Any token past position 10
per detector is invisible. So read this as *consistent with* `mechanism-flow` on every
occurrence, **not** as `mechanism-flow` proven on every occurrence. This is the exact trap §4
already fell into once and it applies to my table as much as to anyone's.

**The webdesign row is the one that decides the fix shape.** Different token, and note the
spaces inside the braces — `{{ variable }}`, not `{{.field}}`. That is not a Go range block
failing to iterate; it is content *about* templates (very likely §4's known-benign
prompt-library copy carrying `{{TONE}}`/`{{COLOR}}`). **The population is not homogeneous, so a
fix aimed at `mechanism-flow`'s schema would leave a live occurrence class untouched.**

### The two measurements the bug file never had, and they de-risk candidate 1 completely

Both harnesses are in the scratchpad (`probe260/parseprobe.go`, `probe260/execprobe.go`) and
both carry controls that could have come out otherwise.

**1. Do any live templates fail to PARSE?** Parse is data-independent, so this needs no
`RenderContext` replica. The only thing a replica can get wrong is the FuncMap NAME SET — an
undefined function is itself a parse error — so the seven names were extracted mechanically
from `executeGoTemplate` rather than typed from memory:

```bash
sed -n '/func executeGoTemplate/,/}).Parse(templateStr)/p' platform/orchestration/actions/call_agent.go \
  | grep -oE '^\t\t\t"[a-zA-Z]+":' | tr -d '\t":' | sort
# default eq isset lower ne safe upper  (7)
```

`[MEASURED]` **0 parse failures out of 304 components (251 active at the time of the run).**
Controls, both fired: an unclosed `{{if}}` MUST fail to parse (it did — the probe panics
otherwise), and a valid nested `{{if}}`/`{{range}}` MUST parse (it did).

**2. Would any STORED section fail to EXECUTE on a rerender?** Faithful without a replica
because `contextToInterfaceMap` merges `ContentData` at the **top level** of the data map
(`component_library.go:1266-1268`), and `missingkey=zero` makes every absent site-level field
safe — which is §2's own finding. So a failure here is caused by `content_data` and nothing
else. It is conservative rather than inflated: it cannot manufacture a failure from a missing
site field.

`[MEASURED]` **0 execute failures out of 1,778 stored sections.** Controls, both fired: §2's
A/B pair — a string where an array is ranged MUST error, and the same field coerced to the
declared array-of-objects MUST render.

**Together these say: deleting the fallback changes the behaviour of nothing that currently
works.** Nothing parses through it, nothing executes into it, nothing is written in its
dialect, and no stored artefact depends on it.

### §4's zeros, re-verified on the grown population

§4's constituency measurement was taken at 255 components; there are 253 active today of 306
total, so it was worth re-running rather than quoting.

`[MEASURED 2026-08-19]`

| | 08-12 (§4) | today | note |
|---|---|---|---|
| components using `{{#` handlebars blocks | 0 of 255 | **0 of 253 active** | the fallback's own dialect |
| using `{{nav_items_html}}` | 0 | **0** | fallback-only placeholder |
| using `{{quick_links_html}}` | 0 | **0** | fallback-only placeholder |
| stored `page_components` leaking control directives | 0 of 1,452 | **0 of 1,789** | positive-controlled regex |
| stored rows containing any `{{` | 1 | **1** | the same known-benign prompt-library row |

**New, not in the file: chrome is clean too.** 72 stored `site_components`, **0** leaking
control directives and **0** containing braces at all. This matters more than the page numbers
because the chrome paths have no `validate_content` downstream — a chrome template failure
would ship mangled markup to a live site silently (LANDMINES has this as its own entry). Today
nothing is in that state.

**Exposed population: 110 active components use Go control syntax** (`{{range|if|end|with|else`).
§4's "33 components with a `{{range}}`" was the narrower cut.

### ⚠ CORRECTION to the bug file — §5 candidate 2's blocker has GONE, and §9b's defect with it

§5's boxed correction says a type gate would *"cover 4 components and report a clean sweep over
the other 251"*, because 4 used the legacy JSON-Schema `properties` dialect, 164 the house
`fields` dialect and 87 neither. §9b filed the 4 as an adjacent defect — a supposedly extinct
dialect reintroduced four times, most recently 08-10.

`[MEASURED 2026-08-19]` **`properties` is extinct again: 0 components, active or inactive.**
Of the four §9b named, `report-dossier` and `loans-consolidation` no longer exist under those
names; `mechanism-flow` and `evidence-timeseries` are still active and **both now carry the
house `fields` dialect.** Someone converted them in the intervening week.

Active schema shapes today: **175 `fields` · 75 NULL · 2 empty `{}` · 1 other · 0 `properties`.**

And the number that actually decides candidate 2's feasibility — coverage over the **exposed**
population rather than over all components:

> Of the **110** active components whose template uses Go control syntax, **107 carry a
> `fields` schema** and 2 have no schema at all.

So the gate is **97% covering where it matters**, not 4-of-255 armed-but-inert. §5's warning was
correct when written and is now obsolete; the expiry date its own addendum predicted has
arrived, in the favourable direction. **Do not design around a dialect split that no longer
exists.**

`[MEASURED]` The acute set is now **14 llm-authored `array` fields across 14 components**
(§9a recorded 13), and **all 14 declare `items`** — so the array-of-objects shape is
expressible with what `SchemaContentFields` carries forward. No `list`-typed field is
`source: llm`.

### A design constraint that came from another lane, not from the code

The owner has ruled that all sites should be capable of having tools. Tool pages legitimately
contain `{{ }}` literals in their copy — a prompt library, a syntax gallery, anything
documenting template syntax. My 26th event is exactly that shape. **So any fix must
distinguish "the renderer failed to execute" from "this content contains braces", and the
positive control has to be a tool page whose copy contains braces and which must still PASS.**
A fix tested only against failing pages cannot detect that it has started refusing good ones.

Worth stating explicitly because it cuts one way: with the render seam failing loud, no leaked
HTML reaches `validate_content` at all, so this hazard only bites if someone tightens that
detector's regex. It is an argument **for** the seam fix and **against** touching the regex.

### MISSTEP — I told two lanes their `unbuilt_internal_link` rows were this bug counted twice

Logged in full at `docs024_key_docs_latest/WRONG_CALLS.md` (2026-08-19). Short version: I
generalised from the item TYPE NAME while the `summary` column and the join I had already run
were on screen. An `unbuilt_internal_link` item is filed because a link points at a missing
page, is then **dispatched to build that page**, and that build is what the leak refused — so
the type records why the build was requested, not a second sighting. **Every one of those rows
is a genuine occurrence; the census of 26 is not inflated and must not be reduced.** The
receiving lane said it would have carried my framing into its own notes as fact.

The real double-count lives one level up and is now owned elsewhere: the `loanzy_uk_example_site`
lane filed **`bugs_open/328`** for the class *"any failed build leaves a live dead link"*, which
survives 260 being fixed because truncation, a missing component and 307's terminal kills all
produce it too. **260 owns the render seam; 328 owns the link consequence.** Point at it, do
not annex it.

### Cross-lane coordination this session

Conferred with `portfolio_positioning` (remortgagecalculator.uk — a locked, stable, still-failing
specimen; they ran the component-identity lookup and reported the blocker rows carry NO
`location` and no class names, so CSS fingerprinting is unavailable from `agent_error_log`) and
with `loanzy_uk_example_site` (greenfield case; they corrected their own §11 fingerprint after I
checked it, and logged that in WRONG_CALLS themselves). Both offered reproductions; both were
asked to **hold** — what this lane lacks is an AFTER, not another BEFORE. loanzy will run a
clean greenfield build once the fix rolls and report the result either way.

**Fix is Go, so it is inert until an image is rebuilt and rolled.** Told both lanes so, since
one of them is sequencing an end-to-end re-run around it.

---

## 2026-08-19 (later) — the detection half re-verified, and the code comment disagrees with the live roster

§7 says *"the registered `unrendered_templates` discovery check is configured in no agent and has
never run"*. Confirmed today, and the evidence is sharper than the file's:

- **The check exists in code**: `UnrenderedTemplatesCheck` at
  `platform/orchestration/actions/discovery_checks/check_integrity.go:676`, scanning both
  `site_components` and `page_components`.
- **Its own header comment claims an owner it does not have.** `check_integrity.go:9` reads
  *"completeness-discovery-agent: cross_site_contamination, unrendered_templates"*. The live
  `completeness-discovery-agent` configures **44 checks and `unrendered_templates` is not among
  them** (read from `agent_definitions.default_config`, active non-snapshot rows only). No other
  discovery agent configures it either — `design`, `quality` and `availability` rosters were all
  read in the same query.
- The only live config mentioning the string is `page-build-handler`, and that is the
  **blocker type** `validate_content` emits, not the discovery check.
- **`[MEASURED]` 0 `site_work_items` of type `unrendered_template` have ever been created**, all
  history.

⚠ **Demand control, stated honestly:** that zero does NOT prove the check would have caught
anything. It scans STORED components, and stored components are clean (0 of 1,789 — §13c). So
"never configured" and "configured but nothing to find" are indistinguishable from the item count
alone. What the absence does mean is that **if the seam fix ever regresses, or a chrome template
starts failing, nothing sweeps for it** — and chrome is the path with no `validate_content`
downstream. This is a gap in the safety net rather than a current miss.

**Not fixing it in this lane** — it is `bugs_open/149` §B1's item and wiring a discovery check to
an agent roster is that lane's seam, not mine. Recorded here because "the detector exists" is the
kind of claim a future reader will take from the code comment, and the comment is wrong.

## 2026-08-19 (later) — chrome call sites already carry an error return

Reading the three chrome renders before judging any plan: `RenderHeader`
(`component_library.go:1969`), `RenderFooter` (`:2038`) and `RenderHead` (`:2311`) **already
return `(string, error)`**, so propagating a render failure needs no signature change at those
sites. Each also already has a *structural* fallback — `RenderFallbackHeader`/`Footer`/`Head` —
used when no eligible component resolves, which is a different and much better-behaved thing than
the regex render fallback: it produces a real working header rather than mangled markup.

That opens a per-site design question worth deciding deliberately rather than by default: on a
template execution failure, does chrome (a) return the error and fail the step, or (b) fall back
to the known-good structural header? (b) keeps the site up but silently swaps a designed header
for a generic one — a degradation of the same family as the bug being fixed, just less visible.
Flagging it rather than assuming; it is exactly the "order fix candidates by what closes the
door" question.

## 2026-08-19 (later still) — ⚠ THE SECTION-EDITOR PATH HAS NO `validate_content` GATE, and it writes to LIVE pages

This is the finding of the day and the bug file does not state it. §4 says *"the gate is why"* —
`validate_content` refuses before persisting. **That is true of the page-BUILD path only.**

`applyContentEdit` (`section_editor_actions.go:886`) and `applyComponentSwap` (`:996`) render
through the same `RenderTemplate`, and their output is written by `updatePageComponentAfterEdit`
(`:1233`) with a plain `UPDATE page_components SET rendered_html = $2` (`:1251-1252`). Grepping
the whole file for `validate`/`unrendered` returns **one comment about a review-queue sweep and
nothing else**. There is no content validation between that render and that write.

So on this path the mangled fallback output would be **stored and served on an already-live
page**, with no gate to refuse it. The copy lane flagged the same seam from the other direction
in §9c (their stage-2 executor supplies agent-written `content_data` and re-renders here).

**Both editor sites already guard `if rendered == ""` — and that guard CANNOT catch this bug**,
because the fallback never returns empty; it returns well-formed HTML with the directives left
in. An existing check written for roughly this class, blind to the actual failure. (Same family
as `[VERIFIED]` off an echo: the guard exists, reads as protection, and tests the wrong thing.)

`[MEASURED 2026-08-19]` The path is live and busy: **271 `content_rewrite`/`content_edit` work
items, 117 complete, 2026-04-08 → today.** (`orchestration_states` has no `agent_type` column and
prunes ~24h anyway, so the durable count comes from work items; §9d's "132 section-editor
orchestrations" was a same-day read of the pruned table and cannot be re-derived now.)

**So the correct risk statement is not "no live damage is possible" — it is "the ungated path has
not yet been unlucky."** 117 completed live edits and 0 of 1,789 stored components carry the
leak, which is a real demand control (the path is exercised, not idle) but is not a guarantee of
anything about the 118th. Chrome is the same shape: no gate downstream, currently clean.

**This changes the fix's justification.** It is not only "convert a confusing failure into a
clear one on a path that already refuses" — on the editor and chrome paths it is "close a route
by which mangled markup reaches a live page with nothing to stop it." The build path is where it
FIRES today; the editor path is where it would COST the most.

---

## 2026-08-19 late → 2026-08-20 — the fix is BUILT and COMMITTED, one council round found three real defects, and I broke HEAD once in the process

### Council round 1 (`a44d9eb8`) came back REVISE in 14 minutes and was right on every gated count

Submitted 20:50Z, verdict 21:00:50Z. Two councils were already executing ahead of it and it still
landed in a quarter of an hour, so **the "budget ~30 minutes" line in CLAUDE.md is a ceiling, not
an estimate.** 13 seats ran, 3 abstained, `decided_by: gating objection from bug_historian`.

Verdicts: `editquality` object · `bug_historian` object (the gate, one HIGH) · `reuse_agent`
object · `guardian` object (one HIGH) · `prior_art_librarian` object · `architecture` object
(`needs_rfc`) · approve from `guidelines`, `diagnosis_guardian`, `improvement_guardian`,
`compliance`, `render_guardian`, `debug_historian`, `constitution`, `mission`.

**What it found that I had actually got wrong** — not stylistic, not process:

1. **`editquality` + `guardian`: three call sites named in the plan with no edit.** My submission
   said they were "in this change" and cited the 8-edit schema cap. The seat's reply is the one
   that matters: *"the plan's own risk section states all 15 call sites must land in one commit or
   HEAD does not compile — this is not a stylistic gap, it is an admitted incompleteness."* Right.
   All three are implemented (`assemble_from_library`, `GateConvertedTemplate`, the audit binary).
2. **`bug_historian` (GATING, high): I treated `missingkey=zero` as a safety property.** My own
   rationale used it to argue the execute-probe was conservative, while the estate's history says
   that default is the mechanism behind fleet-wide silent blanking. The absent-field sibling is
   NOT closed by this change, and saying "absent is never a violation" without saying what DOES
   cover absence reads as declaring it safe. Now stated as a known-open gap in three places, with
   the measurement: the presence gate covers **2 of the 15** render call sites and only fields
   marked BOTH `source:"llm"` AND `required`.
3. **`reuse_agent`: I cited `resolvedValueSatisfiesDeclaredType` as the live precedent and then
   wrote a parallel copy of it.** That is the fork-path pattern with the evidence in my own
   `grounded_in`. Moved to `datahelpers.DeclaredTypeSatisfied`; both questions now share it.
4. **`prior_art_librarian`: two of my claims did not hold.** Below, because they are the
   expensive kind.
5. **`bug_historian` + `guardian` + `render_guardian`, converging from three directions on the
   same shape**: chrome "not stored, logged" and rerender "carried, named in the output" are both
   *reports-complete-but-degraded*, which is `bugs_closed/028`/`040`. `render_guardian` put the
   sharpest version of it: a mistyped field that breaks the render is *the same class* as a
   missing required field, which this very action already escalates. Answered by using the same
   escalation, not by arguing the difference.

**The cheap lesson: three seats independently finding the same shape is not three objections, it
is one design error.** I had written the chrome and rerender arms in different sittings and never
compared them against each other.

### The two prior-art claims I withdrew

**(a) "A 13th render seam nobody had named."** FALSE, and one `grep` would have shown it:

- the concept register already names it — `page-build-pipeline.md`: *"`RenderTemplateWithMap` …
  is deliberately EXEMPT, named rather than silently skipped"*;
- `bugs_open/238`'s council round enumerated it as one of **eight** unguarded call sites;
- `idea_uk_vm_site`'s own `bug_historian` seat FOUND it and routed it through the `<no value>`
  detector — *and that lane's notes say it measured the symbol as absent from the binary.*

What survives is narrower and still worth the edit: its caller `ReplaceAllString`s the live
contact block with the result, so an error **deletes** it; and its language **diverges** (no
FuncMap, so `{{safe}}` is a parse error there while it works on the component seam).

**(b) "Latent, one ordinary edit away."** Also wrong, and I found this one by chasing the seat's
question rather than the seat finding it: the chain ends at `RerenderSitePagesAction`, which is
in **no entry of `GlobalActionRegistry`** (320 handlers). The path cannot be dispatched. My DB
query that "found 3 live agents naming `rerender_pages`" was measuring a STEP NAME
(`rerender_pages.pages` is `get_pages_for_rerender`'s output field), not an action — the exact
"your measurement answers the question you encoded" trap. So the edit is a trap disarmed before
revival, not damage stopped today, and the honest framing is in the code, the RFC and the bug
file.

**(c) A third, smaller one, mine and unprompted:** I wrote in a code comment that declaring
`refuse_dead_url_controls` in `ConfigKeys` had cost the RFC_022 budget its visibility. It had
not. `cmd/config-key-audit/optionalbudget.go:14-21` counts `spec.Optional` only and skips
`ConfigKeys` **on purpose** ("settings rather than input references"). What the declaration buys
is the unknown-config-key report. Corrected in the comment before commit.

**And I checked the dead-URL precedent claim I had repeated from `dead_url_guard.go`'s header
("three seats independently demanded default-OFF").** The corpus says **two** — `guardian` and
`architecture` on council `98852baa`; `render_guardian`'s objection that round was that the
rerender path records without refusing, which is adjacent, not the same. The design is unchanged;
the citation is corrected. This is the second time in two days on this tree that a *code comment's
account of a council round* turned out to be looser than the round — see `WRONG_CALLS.md`.

### The defect my own census found in my own checker — the best evidence this round produced

Writing the arming migration's "known population" section meant asking: **what would the gate
refuse today?** Two queries, and the second one came back non-zero:

- top-level declared-array fields holding a non-array, across every stored `page_components`
  row: **0**;
- the nested case: **5 elements on ONE page** — `fundamentallyai.com`
  `/production-backend-engineering`, `mechanism-flow`, `steps[].branches`.

That page is **deployed, serving, 8,824 bytes, no braces in its HTML.** Every one of the five
`branches` is the **empty string**, and the template gates them (`{{if $s.branches}}` precedes
`{{range $s.branches}}`). My first checker reported `""` as "a string where an array is declared"
— so an armed gate would have **refused a rebuild of a healthy live page**, and it is the only
such row on the estate, so no test I would have thought to write would have caught it.

Fixed by making the checker share ONE emptiness predicate with the presence gate:
`isEmptyContentValue` → `datahelpers.IsEmptyContentValue`. My round-1 doc comment had *asserted*
that the two gates "must not disagree"; it took the census to notice they already did.

**The transferable bit: the question "what population would this refuse today" is not a
formality for the migration header. It is the only test that runs against the whole estate, and
it found a false positive that nine hand-written unit tests did not.**

### The mutation proof, including the part that did NOT fail

Re-added a one-line fallback (`return templateStr, nil, nil, nil`) and ran the suite:
**5 tests fail** — `TestRenderFailsOnAMistypedNestedField`, `TestParseErrorIsRefusedNotDegraded`,
`TestRetiredHandlebarsDialectHardFails`, `TestSecondRenderPathNoLongerExists`,
`TestFooterComplianceWrongTypeIsRefusedAtTheSeamAndFallsBackWhole`. The control
(`TestRenderSucceedsOnCorrectlyShapedContent`) still passed, as it must.

**`TestConversionGateRefusesATemplateTheRendererCannotExecute` PASSED under the mutation**, and
that is worth writing down rather than quietly enjoying the 5: `GateConvertedTemplate` has a
guard **in series** — with the fallback restored, the mangled `{{.InstanceID}}` survives the
render and its own placeholder check catches it. So that test does not prove the seam; it proves
the gate's second line of defence. A mutation that passes has usually hit a guard in series.

### Two tests whose PREMISE this change inverts, reworked rather than deleted

- `TestFooterComplianceWrongTypeDoesNotDestroyFooter` said, in its own comment, *"degraded output
  is acceptable"* — it was pinning the fallback as a FEATURE. Rewritten as
  `…IsRefusedAtTheSeamAndFallsBackWhole`: same promise to the page (a real footer, never a
  half-rendered one), kept by `Inject*`'s existing ladder instead of by the regex renderer.
- three `…OnRegexFallbackPath` form-action tests tested a branch that no longer exists. Replaced
  by one test of the property that actually mattered — *no second, weaker renderer can ship a
  page* — with the deleted tests' reasoning preserved above it.

### Then HEAD broke, twice, and the second one was me

1. **08:06Z, `ae7a8d739`** (an unrelated section-editor claims-guard commit) swept two lines of
   this work into HEAD as a **same-file passenger**. HEAD stopped compiling:
   `section_editor_actions.go` called `RenderTemplate` with two returns while the seam still
   returned one. `make build-*` builds from committed HEAD, so **every session's image build was
   broken** and nothing in `git status` would tell them why.
2. I committed my half (`80b9c6235`), which fixed that — and **broke it again in a new way**,
   because my pathspec included `v3_site_actions.go`, which in the shared tree already carried
   another lane's uncommitted call to `buildPageDeployStampQuery`. The call shipped; the
   definition (in `deploy_evidence.go`, still uncommitted) did not.
3. Fixed forward with `a0bb2d867`, a declared `sweep:` of that one file, saying whose it is and
   why deleting their line was the worse option.

**The tell, and the only reliable one: a clean `git worktree` at HEAD.** My own tree built fine
in all three states, because it holds everyone's uncommitted work. `git worktree add <scratch>
HEAD` then `go build ./...` is a 90-second check that answers a question no amount of local
building can.

### The test failure that was not mine, proven rather than assumed

While all this ran, `v3_site_reconcile_identity_test.go` failed in my tree on **a different
reconcile test each run** and passed in isolation. Two checks settled it: it touches no render
code, and at clean HEAD in the worktree the package passes **3/3**. The cause is another lane's
**uncommitted** edits to that very test file (plus an untracked `plan_sections_item_fields_
dialect_test.go`), which compile into the package in the shared tree and not at HEAD. Left as
found, named in the commit message.

**Practice worth keeping: on this tree, "the package fails locally" is not evidence about your
change until you have run it at HEAD in a worktree.**

### State at 2026-08-20 10:10Z

- **Committed:** `80b9c6235` (the change, 27 files, with `STY-057` + its index row and `RFC_041`
  in the same commit — ordering-exemption condition 2) and `a0bb2d867` (the sweep).
- **Council round 2** submitted on the same trail `a44d9eb8`; run correlation
  `efd19ef7-79cf-4603-b97e-e905ad0e3094`, orchestration `10acf41b-55f0-4bb5-ab88-2809b435f4a9`.
- **NOT LIVE.** Go, so inert until a chassis built from `80b9c6235` rolls. Verify per SERVICE at
  the binary's provenance stamp, never at git or the tag.
- **The arming migration is `502_bugfix_260_arm_mistyped_llm_fields_HOLD.sql`** (+ ROLLBACK
  sidecar). ⚠ **I first numbered it 498 and three other sessions had already used 498 today** —
  the sequence is contended, so `ls` the directory immediately before naming a file, not before
  writing it.

### Council round 2 (same trail `a44d9eb8`): **APPROVED** 2026-08-20 08:34:52Z, ~10 minutes after submission

`decided_by: approved with 5 advisory objection(s) — none high-severity`. 16 seats, 3 abstained,
not truncated. Round 1's three gating findings are gone: the three call sites are implemented, the
absent-field gap is registered rather than implied safe, and the prior-art claims are withdrawn.

**Advisories acted on, with the evidence the seats asked for:**

- **`reuse_agent` (medium) — "`SchemaContentFields` already normalises dialect differences; was it
  checked before writing your own `items` handling?"** Fair, and now answered *in the code* rather
  than in a submission: `SchemaContentFields` normalises the **field-set** dialect (v2 `fields`
  vs legacy top-level `properties`) and `ContentTypeViolations` calls it for exactly that; it
  copies `items` through **verbatim** (with `source`, `on_missing`, `fallback`, `missing_reason`,
  `min_items`), so there was no element-shape normalisation to extend. Widening it would change
  what all three of its other callers read for the benefit of the one that walks into items.
- **`prior_art_librarian` (medium) — the registry-absence claim is asserted, not quoted.** Right,
  and it is load-bearing (it is what makes the contact-info edit a disarmed trap rather than a
  live save). The evidence, run 2026-08-19:
  ```
  $ grep -c "Handler:" platform/orchestration/actions/registry.go
  320
  $ grep -rn "RerenderSitePagesAction\b" --include=*.go . | grep -v rerender_pages_actions.go
  (no output)
  $ grep -n "rerender" platform/orchestration/actions/registry.go
  873: "get_pages_for_rerender": {    879: "rerender_single_page": {
  885: "rerender_page_sections": {    891: "create_rerender_items": {
  ```
  320 registered handlers, no entry for it, and no non-test caller outside its own file. Now a
  LANDMINE entry with the same commands.
- **`debug_historian` (low) — "no needle-gate count of live templates relying on any of the
  deleted substitutions beyond `{{#`, `{{nav_items_html}}`, `{{quick_links_html}}`."** The seat is
  right that I named three markers, and the better answer is that I do not need an enumeration:
  **every dialect only the fallback could render is, by definition, a Go-template PARSE error** —
  `{{#each}}`, `{{nav_items_html}}` and bare `{{field}}` (no dot, which `renderHandlebarsSubstitutions`
  handled) all fail `template.Parse` with "function not defined" or a bad-character error. The
  measured **0 of 251 active templates fail to Parse** therefore covers the whole class, not just
  the three I greped. Recorded here because that argument is stronger than the greps and I had not
  stated it.
- **`architecture` (low) — cite the optional-key budget for these actions post-change.**
  `./scripts/audit-optional-key-budget.sh`: *"122 actions declare optional keys; 22 of them are
  SHARED (>=2 live carriers) — budget: 10 — **0 shared action(s) over it**"*, and
  `render_component` does not appear in the counted set at all, because the counter reads
  `spec.Optional` and both new flags are `ConfigKeys` (settings). The cron parity test passes 6/6.
- **`guardian` (low) — "enumerate every caller migrated, or the two gates can drift again."**
  Checked: `grep -rn "resolvedValueSatisfiesDeclaredType\|func isEmptyContentValue"` returns the
  MOVED comment, the one-line alias, and nothing else — **zero surviving copies of either logic.**
- **`guardian` (low) — "confirm nothing downstream reads `unanalysed`."** Checked across Go, SQL,
  Python and shell: **no consumer outside `rendercheck.go` itself.** The bucket's printed label
  already says "NOT cleared", which is precisely the meaning the new members carry.

**One advisory is WRONG and is recorded as refuted rather than quietly dropped.** `guidelines`
(low): *"the platform's WORK-ITEM DEDUP rule requires DELETE+INSERT against idx_swi_dedup, not ON
CONFLICT"*. `insertWorkItem` (`load_work_item_actions.go:1509-1518`) is an
`INSERT INTO site_work_items … ON CONFLICT (site_id, item_key) …`, and the surrounding comments
say so explicitly — the file's own header warns against a caller writing its **own**
`INSERT … ON CONFLICT DO NOTHING` and thereby *inheriting none of* the shared helper's anti-churn
behaviour, which is the opposite of the seat's reading. `emitChromeRenderFailedItem` uses the
shared helper, so it inherits whatever that helper does; my sketch described it correctly. **Second
time this week a medium/low advisory on this estate has been wrong on a checkable point — read
them, then check them.**

**Two advisories accepted as fair criticism of the SUBMISSION rather than the code**, recorded
because the next resubmission on any lane can avoid them: `editquality` objected that edit 6
described changes to two files not named in its `file` field, and that the register entry and the
RFC were claimed in prose with no edit entry tracking them. Both true. The 8-edit cap is the cause
and the fix is to spend a slot on the registration artefacts rather than mention them — a
reviewer cannot check what the plan does not list.

**Trailer:** the code commit `80b9c6235` carries `Council-Submitted: a44d9eb8…`, which 098 credits
automatically now approval has landed — no amend, and forward-only forbids one anyway. The docs
commit carries `Council-Reviewed:` because by then the approved verdict had been read.
