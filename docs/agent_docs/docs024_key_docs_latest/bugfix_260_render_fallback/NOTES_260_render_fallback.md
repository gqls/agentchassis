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
