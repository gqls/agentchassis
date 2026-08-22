# 357 — a whole tool page is stored in a slot that claims to be the shared `hero` component, so every check about it reports the wrong thing and no repair route can act

**Filed 2026-08-22 by the `bugfix_277_required_fields_repair` lane**, while scoping the
`no_content_data` backfill the owner asked for. These 9 rows were the population I had to REFUSE,
and the refusal is the finding.

> **Root cause is UNDIAGNOSED and deliberately not asserted here.** The symptom, the population and
> the blast radius below are first-hand measurements against the live database. Which writer assigns
> the `component_id` is **not** established, so it is in the diagnosis loop rather than guessed:
> `RUN_CORRELATION_ID=63d4d1a7-ffec-4570-866b-8a0a41e3c69d` (filed 2026-08-22). Do not repeat a cause
> from this file until that verdict is read — there isn't one in it.

## The mechanism, plainly

A tool page is built with **one** `page_components` row. That row's `component_id` points at the
shared **`hero`** component — whose template renders a title band with a headline, an optional
subheadline and up to two buttons — but its `rendered_html` holds **the entire interactive tool**,
9.5–21.8KB of markup, controls and JavaScript. The declared component never produced the stored
bytes.

Nothing errors, because nothing compares the two. What happens instead is that every mechanism
keyed on the component reasons about a hero that is not there:

1. **The schema check is right and useless.** `hero` declares `headline` as required;
   `content_data` is NULL; so `required_fields_missing` files *"Component 'hero' on page
   tool-ttk-calculator is missing 1 schema-required value field(s): headline"* — while the page
   serves `<h1>Time-To-Kill (TTK) Calculator</h1>` perfectly well.
2. **The router then classifies it `no_content_data` and parks it**, correctly, because the only
   repair it has would regenerate from `content_data`.
3. **And that park is load-bearing, not tidiness.** Any repair that gives this row a `content_data`
   makes `datahelpers.ContentDataCanFillTemplate` true, and the next regeneration renders the
   **hero template** — swapping a working 16KB tool for a 2KB title band. The parked state is the
   only thing standing between these pages and that.

## The population [MEASURED 2026-08-22 ~10:00Z]

| site | pages | born | components on the page |
|---|---|---|---|
| gamesdesign.co.uk | `tool-jump-physics`, `tool-ehp-calculator`, `tool-drop-rate-simulator`, `tool-lanchester-sim`, `tool-ttk-calculator`, `tool-progression-architect` | 2026-06-05 | 1 each |
| gamesdesign.co.uk | `game-p2p-networking` | 2026-06-06 | 1 |
| gamesdesign.co.uk | `game-pathfinding` | 2026-06-26 | 1 |
| mortgagecalculator.co.uk | `tool-simple` | **2026-08-08** | 1 |

**It is not purely historical.** Eight are from June, but one recurred on **2026-08-08**, two weeks
before this filing — so whatever does this was still reachable a fortnight ago. Every one of the
nine has exactly **one** component on the page, and none has ever been re-written
(`updated_at` = `created_at` on all nine).

```sql
-- the population, re-runnable
SELECT s.domain, p.name AS page, pc.created_at::date,
       (SELECT count(*) FROM page_components x WHERE x.page_id = pc.page_id) AS components_on_page,
       left(regexp_replace(pc.rendered_html, '\s+', ' ', 'g'), 60) AS html_starts_with
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
WHERE cc.name = 'hero'
  AND position(left(cc.html_template, position('{{' in cc.html_template) - 1) in pc.rendered_html) = 0;
```

The predicate is the honest test and is worth reusing: **does the component's own static template
prefix appear anywhere in the HTML it supposedly produced?** For a genuine hero it appears at byte 1.

## Why this is worth its own file rather than a line in 277

`bugs_open/277` is about findings with no repair handler. This is a level below it: the finding
itself is **about the wrong thing**. A repair route built for `no_content_data` — including the one
this lane just shipped (migration `540`, `cmd/content-data-recover`) — must *exclude* these rows, and
the only reason it does today is that the recovery tool refuses anything whose re-render is not
byte-identical to the stored HTML. **A looser tool would have written these nine and armed a
regeneration that destroys nine working tools.** That near miss is the argument for filing.

## Fix candidates, ordered by what closes the door

1. **Re-type the row** — point `component_id` at a component whose template is `{{.body}}`-shaped (or
   a bespoke per-page one) and move the tool HTML into `content_data.body`. Makes the page rebuildable
   AND makes every downstream check correct, because the component would then describe what is
   stored. Needs the diagnosis first, or the producer will mint more.
2. **A born-wrong guard at the write seam**: refuse (or flag) a `page_components` write whose
   `rendered_html` does not contain its component's static template prefix. This is the check that
   found them, it is one string comparison, and it would have caught all nine at birth. ⚠ It must be
   measured against the whole live table before arming — the same predicate that finds these nine
   also fires on legitimately drifted templates (15 more rows in this lane's census), so a naive
   version would be noisy. Threshold and exclusions need the census, not a guess.
3. **Leave parked** — today's state. Costs nothing visible and keeps the tools safe, but every future
   repair mechanism has to independently rediscover that these rows are poison.

## How to verify a fix

Not "the item closed". The page must still serve its tool: `curl` it and assert the tool's own markup
is present (`class="tool-page"`, its controls, its `<script>`), then re-run the population query above
and expect the row to have left it. A `complete` work item proves nothing here — `bugs_closed/287`.

## Relations

`bugs_open/277` §8 (the census that found these, and the three-way split of the 27 parked rows) ·
`cmd/content-data-recover` + migration `540` (the repair that deliberately refuses them) ·
`datahelpers.ContentDataCanFillTemplate` (why a backfill would be destructive here) ·
`bugs_open/149` (checker-layer defect queue) · LANDMINES *"A writer of `page_components.rendered_html`
that does not repair its links…"* — the same two writers (`create_tool_component_action.go`,
`deploy_tool_action.go`) are named there as a known un-allow-listed gap, which is where the diagnosis
was pointed.

---

## ADDENDUM 2026-08-22, same day — there is a LIVE, PROVEN route for this shape already, found via a council objection

The `reuse_agent` seat, reviewing this lane's recovery tool (council `cd8e555d`), objected that
adjacent tooling existed and had not been evaluated. It was right, and what it pointed at matters
more to **this** file than to the one under review.

**`docs024_key_docs_latest/loancalculator_couk/decompose/`** (the `loancalculator_couk` lane) exists
because that lane hit this exact shape — a page whose content is one stored blob — and solved it:

- **`load_decomposition.py`** replaces a page's single verbatim row with properly decomposed component
  rows, in one transaction per page, backing up **every** affected page's rows first (*"a restore path
  that only covers the page you thought you were changing is not a restore path"*), and writing a
  predicted assembly so the real output can be diffed against it afterwards.
- It also documents the rule this file needs and did not know: **a page ships VERBATIM when
  `rebuild_policy='owned'` ∧ it has EXACTLY ONE component row ∧ that row carries
  `content_data.deploy_mode='verbatim'`.** The flip between verbatim and assembled **is the row count,
  not a flag** — so *adding* a row beside a verbatim one silently switches the page to assembly with
  the old full document still in the mix, producing a document nested inside a document.

**Why this changes fix candidate 1.** These nine pages have exactly one row and **NULL**
`content_data`, so they do **not** carry `deploy_mode='verbatim'` — they are assembled, and they work
only because assembly emits the single row's stored HTML. Two consequences:

1. **A one-row page is one edit away from either outcome**, and the safe target should be chosen
   deliberately: either make it genuinely verbatim (`deploy_mode='verbatim'` in a recovered
   `content_data`), or decompose it into real components as that lane did. Both are established; the
   thing to avoid is the accidental middle where a second row appears.
2. **Whoever fixes this should read that lane's scripts before writing new ones.** They already
   carry the backup convention, the predicted-assembly diff, and the restore path — and a second
   hand-rolled decomposer is how the estate ends up with two.

Not adopted here, and deliberately: the producer is still undiagnosed (`63d4d1a7`), and decomposing
nine pages before knowing what mints them would repair the stock while the flow ran.

---

## DIAGNOSIS RESULT 2026-08-22 — **UNVERIFIABLE**, not confirmed and not refuted. Read this before citing the loop.

`63d4d1a7-ffec-4570-866b-8a0a41e3c69d` completed with
`status: UNVERIFIABLE — "Diagnosis NOT confirmed (stopped: scope-not-narrowing). Best-effort trail
attached for a human; no fix proposed."` It did locate the population (nine page ids), and it named
precisely what it lacked. **So the header's "root cause undiagnosed" still stands — the loop did not
change it, in either direction.**

**Two of its three stated gaps are already closed by this file's own first-hand measurements**, and
they are recorded here so the next session does not re-run them:

1. *"whether the check has ever fired against THESE pages is unestablished"* — **it has. That is how
   the nine were found.** Every one of them is a `required_fields_missing` row at
   `needs_human_review` with `result->>'route' = 'no_content_data'`; the population query at the top
   of this file selects them from exactly that set.
2. *"content_data joined to input_schema … only rendered_html length was queried"* — **`content_data`
   is NULL on all nine** (measured; it is why the router classified them `no_content_data` in the
   first place), and the `hero` component declares `headline` required, which is what the finding
   reports.

**The third gap is the real question and remains open:** *which writer assigns the `component_id`.*
The loop never fetched the files. Narrowed here, first-hand, without asserting a cause:

- **Six writers insert `page_components` rows** (`grep -lE "INSERT INTO page_components"`,
  non-test): `create_report_page_action.go`, `deploy_tool_action.go`, `save_page_sections_action.go`,
  `create_tool_component_action.go`, `rebuild_blog_listing_action.go`, `adopt_verbatim.go`.
- **A LEAD, explicitly not a conclusion:** both tool actions set the PAGE's section list to
  `["hero", "article-body", "call-to-action"]` (`deploy_tool_action.go:614`,
  `create_tool_component_action.go:606`), while `create_tool_component_action.go:496` says it sets the
  component row's *"slot_name to the function (component naming contract)"* — i.e. **not** `hero`. So
  the `hero` slot on these pages plausibly comes from the page's `sections` list being written as a
  standard three-slot page, with a later writer filling that first slot with the tool. **That is a
  hypothesis about two writers interacting, and it is exactly what the next run should test** — it is
  not established, and nobody should repeat it as cause.

**Why the loop stalled is worth knowing for the re-run:** the symptom I filed asked one question
(*"which writer assigns the component_id"*) while pointing at three files and a join, and the loop
spent its narrowing budget on the population instead. A sharper re-file would state the mechanism as
the interaction — *"the page's `sections` list is written as a standard three-slot page while the
tool's HTML is written into the first slot"* — and point at the two writers plus
`save_page_sections_action.go`, which is the one that fills slots from a page's section list.

---

## CONTRIB 2026-08-22 (from the `bugfix_357_component_identity` lane) — the population is **22**, the writer is **settled**, and 13 rows are **already armed**

Taken up after `bugs_closed/277` closed and its lane's final handoff routed here
(*"there is no work left in this lane; go to `bugs_open/357`"*). Working docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_357_component_identity/`.

### 1. ⚠ This file's stated population (9) and this file's own query (22) disagree

The "re-runnable" population query above, run unmodified today, returns **22 rows**. The nine in the
table are the subset that also carried a parked `required_fields_missing` work item — every one of
them a single-component page. The other thirteen are multi-component pages the same predicate
matches. **Newest row: `vetcomparison.uk` `index`, created 2026-08-22** — a site homepage, born the
same day this file was filed. Not "one recurrence a fortnight ago": a live producer.

### 2. The third gap is CLOSED — `save_page_sections_action.go`, by fingerprint

The file leaves *"which writer assigns the `component_id`"* open and `63d4d1a7` returned
UNVERIFIABLE on it. The row answers it directly, because the writers leave different marks:
`save_page_sections_action.go` is the **only** one that writes `content_brief` (as
`"{slot} section"`) and the only one using `position = i+1`. **All 22 rows carry
`content_brief.section_guidance = 'hero section'` and `position = 1`** [MEASURED 2026-08-22].

The mechanism, all in committed code:

1. `saveSectionsExtractFromHTML` — a tool/game page is `<div class="tool-page">…</div>` with no
   `<section>`, so the regex matches nothing and the documented fallback stores the whole fragment
   as ONE section named `"section"`, the sentinel for *identity unknown*.
2. `enrichSectionsWithPlannedNames` — treats `"section"` as needing enrichment and assigns
   **`planned[Position-1]`** from `pages.sections`.
3. `enrichSectionsWithComponentIDs` — resolves that name to the shared component's UUID.

**`hero` is planned first on all 22 pages.** The tool is called a hero because hero is first in the
list. The lead in this file's diagnosis section — that the three-slot list written by the tool
actions is implicated — is **correct in outline**; the assignment is positional, and the sentinel
meaning "I do not know what this is" is converted into a confident wrong answer.

### 3. ⚠ The "parked state is the protection" argument does NOT hold for 13 of the 22

This file's nine all have `content_data` NULL, which is what makes the argument work for them.
**Thirteen of the 22 carry a complete hero `content_data`** — `headline`, `subheadline`,
`background_image`, `hero_url`, `cta_text`, `secondary_cta` — alongside 11–22KB of tool HTML.
`ContentDataCanFillTemplate` is therefore **already true** on them: the regeneration this file warns
a backfill would arm **is armed now, with no backfill**, and **16 of the 22 are
`rebuild_policy='generic'`**. Nothing has to be written for a rebuild to swap a working calculator
for a title band.

`vetcomparison.uk/index.html` [MEASURED 2026-08-22, http 200, 44,496 B] serves `class="tool-page"`
once and `data-component="hero"` **zero** times — the hero copy its writer produced is stored and
never rendered.

### 4. Fix candidate 2's blocker is measured, and the predicate it needs is a different one

This file rightly says the guard "must be measured against the whole live table before arming"
because the static-template-prefix test also fires on drifted templates. It does: **158 rows**
fleet-wide. Test the component's **own self-declaration** instead — `data-component="…"`, which
`saveSectionsExtractFromHTML` already trusts as the identity carrier in this same file:

| | rows |
|---|---|
| template and stored HTML **agree** | **1,550** |
| both declared and **disagree** | **0** |
| template declares, stored HTML **silent** | **27** ← the pathological set |
| static-prefix mismatches (the noisy test) | 158 |

**Zero false positives fleet-wide**, the 1,550 agreements are the demand control, and the 27 are a
strict subset of the 158 — the ~131 drift-only rows are not in it. The 27 are the 22 plus three
`loancash.co.uk` Ported Page rows, one `leopardessconsulting.co.uk` `blog-listing_pre_037`, and one
`idea.uk` `tool-list` row whose `rendered_html` is **zero bytes**.

**Limit, stated:** silent for the 190 of 339 components declaring no `data-component`, and 97 of the
100 `Ported Page` prefix-mismatches are not flagged. A refuse-on-certainty test for a write seam —
not a census.

### 5. [UNMEASURED] — whether this has already destroyed a tool

A page born this way and then rebuilt would serve a title band and **leave the population query**,
so the survivors cannot report the casualties. There is no systematic `page_components` history
(only ad-hoc `_backup_*` tables). Candidates exist — `mortgagecalculator.co.uk` `game-fact-finder`,
`page_type='game'`, 2 components, 4,363 B, no `<script>` — but "never had a tool" and "had one and
lost it" are indistinguishable from current state. **Not claimed in either direction.**

### 6. Re-filed diagnosis

`RUN_CORRELATION_ID=e580b34a-d284-4f80-ac96-81af1c4adaba` (intake `f7aedef7`), asked the one thing
row fingerprints cannot settle: which leg pairs a genuine hero `content_data` with the tool's bytes,
given the `saveSectionsExtractFromHTML` fallback sets no `ContentData` at all.

### 7. ⚠ CORRECTION to §3 above, same day, ~2h later — "already armed" is REFUTED, and what replaces it is worse

**§3 claims the thirteen rows with a full hero `content_data` are one rebuild away from having their
tool replaced by a title band. That is not true, and the check that refutes it is one I should have
run before writing it.**

`vetcomparison.uk` `index` — the newest row of the 22 — has **six completed
`page_rerender_index_…` work items between 2026-08-19 and 2026-08-22**, and the tool still serves
today (`class="tool-page"` present, http 200, 44,496 B). If a rerender rendered the hero template
over it, that could not be so.

**The timestamps say something stronger: the rerender WROTE these rows.**

| | |
|---|---|
| rerender item | created `2026-08-22 08:44:51`, completed `08:50:19` |
| all four `page_components` rows on the page | created `2026-08-22 08:50:12` |

The rows were written **inside** that window, seven seconds before the item completed, and the hero
row it wrote holds 11,326 bytes of tool HTML with the hero `content_data` beside it.

**So `RerenderPageSectionsAction` does not destroy the tool — it RE-MINTS THE MISMATCH.** It carries
the stored HTML forward, keeps the hero identity and the hero `content_data`, and writes a fresh
row. `updated_at = created_at` across the population never meant "never touched since birth"; it
means **every row is newly born, repeatedly**, because this path deletes and re-inserts rather than
updating.

**What this changes for anyone fixing it:**

1. **The population is self-renewing.** Repairing the 22 rows without fixing the producing path is
   wasted work — the next rerender of those pages reproduces the defect. Flow before stock.
2. `bugs_open/277`'s "the parked state is the protection" argument is about a *backfill* arming
   `ContentDataCanFillTemplate`. That function has exactly **one** non-test caller —
   `discovery_checks/check_literal_markdown.go:429` — so it is a **detector's classifier, not a
   gate on the rebuild path**. Do not carry it into this file as a safety mechanism; I did, and it
   is the root of the wrong claim in §3.
3. The rerender's own content pre-check (`rerender_page_sections_action.go` ~380–445) exempts
   self-contained tools via `isSelfContainedSection` — keyed on `component_level=='tool'` AND an
   empty `input_schema`. `hero` is `component_level='section'` **with** a schema [MEASURED], so the
   exemption never fires for these rows. They are handled as ordinary sections throughout.
4. **STILL OPEN, and not to be assumed either way:** whether a full page rebuild *through the
   writer* (as opposed to a rerender) renders the hero template over the tool. Rerender demonstrably
   does not. Nothing here establishes what the writer path does.

Logged in `WRONG_CALLS.md`. The cheap check that would have caught it: **before claiming a stored
row is one operation away from destruction, look for that operation having ALREADY RUN on it** —
six times, in this case, with the outcome recorded in `site_work_items`.

### 8. CORRECTION to §4 — "zero false positives" was wrong: 3 of the 27 are LEGITIMATE, and the exemption that removes them is already documented in this file

I called all 27 flagged rows pathological without opening five of them. Three are correct by design:

| rows | what they are |
|---|---|
| `loancash.co.uk` `tool-price-cap-checker`, `tool-true-cost-calculator`, `tool-complaint-deadline-calculator` | **legitimate verbatim pages** — each is `rebuild_policy='owned'`, **exactly one** component row, and `content_data->>'deploy_mode' = 'verbatim'` [MEASURED]. They store a full `<!DOCTYPE html>` document *because that is what a verbatim page is.* |

That is precisely the rule this file's own ADDENDUM records from the `loancalculator_couk/decompose`
lane — *"a page ships VERBATIM when `rebuild_policy='owned'` ∧ exactly one component row ∧ that row
carries `content_data.deploy_mode='verbatim'`"* — and I flagged them anyway.

**So the guard must carry a verbatim exemption**, and with it the numbers are:

| | rows |
|---|---|
| flagged by the raw predicate | 27 |
| **legitimate verbatim, exempt** | **3** |
| **genuine defects** | **24** |

The other two of the five non-`hero` rows are real, and both are defects this bug would not otherwise
have surfaced:

- `leopardessconsulting.co.uk` `blog` — component **`blog-listing_pre_037`**, a superseded component
  still bound to a live page.
- `idea.uk` `index` — the `tool-list` row's `rendered_html` is **zero bytes** on a live homepage,
  and the served page contains no `tool-list` at all [MEASURED: http 200, 51,689 B, 0 matches]. A
  hollow slot shipping live. `bugs_open/039`'s `sectionIsUnresolvableStub` does not catch it because
  that guard requires `component_id` to be NULL, and this row has one.

**The transferable point for whoever arms this:** the exemption is not a tuning threshold, it is a
statement about what the platform legitimately does. Any version of this guard that lacks it will
refuse three correct pages on `loancash.co.uk` the first time they are rebuilt.
