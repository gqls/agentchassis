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

### 9. CORRECTION to §5, and the casualty question is now ANSWERED: no tool has been destroyed

§5 said the question was unanswerable because *"there is no systematic `page_components` history
(only ad-hoc `_backup_*` tables)"*. **That is false.** `page_component_history` exists — **26,965
rows over 558 pages, 2026-03-16 → 2026-08-22**, with `rendered_html`, `slot_name`, `op` and
`created_at`. It is `bugs_closed/229`'s page-side artefact archive. I missed it because my table
listing was piped through `head -30` and this database has 71 matching tables whose alphabetical
head is all `_backup_*` / `page_components_bak_*`. Logged in `WRONG_CALLS.md`; found by the
diagnosis run, which read the table I had declared absent.

**With it, the answer is NO**, and it is a searched-with-a-control no rather than an absence:

| slots that were interactive in history | |
|---|---|
| still interactive today (**control**) | **182** |
| no longer interactive | 17 |
| slot gone entirely (not opened; not claimed either way) | 39 |

Opening all 17: **fifteen GREW** (`loancash.co.uk` guide pages, `ported-page` slot, e.g. 6,929 →
9,300 bytes) — an ordinary rebuild replacing ported HTML that happened to contain a `<script>`, not
a loss. One shrank modestly (`fundamentallyai.com` `platform-log-index`, 14,256 → 11,470). One is a
real 57% shrink worth its own investigation: **`webdesign.co.uk` `learn-ai-builders-content-first`,
8,855 → 3,781, 2026-08-15** — recorded here as a lead for whoever wants it, NOT as a 357 casualty.
**None of the 17 is a page in this file's population.**

### 10. Council trail `62aac6c2` — TWO revise rounds, and the second one is why this needs an RFC

- **Round 1 REVISE** (gated by `bug_historian`). Checking its objection found a worse defect in my
  own plan: the Layer 2 carry-forward matches stored rows to incoming ones on **slot-name equality
  and nothing else** (`save_page_sections_action.go:517`). So *correcting* a row's identity makes
  the next plan-driven rebuild miss the match, take the `default:` re-append branch, and save the
  incoming hero band **beside** the tool. **A byte-preserving re-type would change what four live
  sites serve.** Fix candidate 1 in this file inherits that hazard and must not be attempted without
  it.
- **Round 2 REVISE** (gated by `editquality`, seconded by `bug_historian`): the resized,
  record-only plan *"changes nothing but writes a log/work-item … the diagnosed corruption is not
  stopped"*, and is the estate's own `079`/`083` "detected but never blocked" recurrence. Fair, and
  accepted.

**The `architecture` seat named the real answer and it is now `architecture_review/RFC_046`:**
identity here is **inferred five different ways and stamped none** (attribute; the `"section"`
sentinel; position; fuzzy name matching; slot-name equality), and a sixth inference is not the fix.
Round 1 was unsafe, round 2 was ineffective, and that oscillation is the evidence that no
inference-based fix is both.

**So the state of this bug is: cause established and cited, population corrected to 22 and still
minting, no casualties found, both proposed repairs blocked on a mechanism question that is now an
open RFC.** Nothing has been changed in code or on any site.

---

## 2026-08-23 — THE STAMP WAS LIVE FOR A DAY AND WROTE NOTHING. Phase 0 shipped; the mint continues.

Picked up from the `bugfix_357_component_identity` lane, which stopped on 08-22 evening after
dispatching council round 4. **Two things it never learned:**

### a. Round 4 was APPROVED, and nothing recorded it

`doc_notes`, 2026-08-22 **18:02:26Z**, on submission correlation `62aac6c2-996f-4b5d-8f8f-72e3daf4c82e`:
*"COUNCIL GATE — APPROVED — approved with 2 advisory objection(s) — none high-severity."* The lane's
last entry says "round 4 dispatched". The verdict landed 9 minutes later and no document says so.

### b. The approved, rolled stamp writes nothing — measured, with the control

Phase 1's own §10.5 named the disconfirming result: *"still 0 after the roll ⇒ the seam is not
reporting or the writer is not persisting."* It has fired.

| check [MEASURED 2026-08-23] | result |
|---|---|
| `page_components` born since the 08-22 15:10:31Z roll | **820, of which stamped: 0** |
| pre-roll rows (**the control** — nothing backfills, by design) | 1,146, stamped 0 |
| `component_versions` with `change_source='render_stamp'` | **0** |
| chassis logs, 24h, `component version:` | **0** — consistent with the *silent* early return |

**Cause, with a control that could have come out otherwise.** `extractSectionFromMap`
(`v3_site_actions.go:2835`) rebuilds every section's metadata into a fresh map from a hand-written
allow-list of six keys. `rendered_template_sha` is not on it. Of **546** live `sections_metadata`
elements since the roll, **0** carry the digest and **546** carry `component_id`. The copy runs; the
key is dropped. Three keys were being dropped, not one (`stripped_markdown_fields` and
`copy_gate_findings` too).

> **This hop had already eaten a key.** `bugs_open/189` lost `stored_slot_name` here and was fixed by
> adding that one key plus a test asserting *its own key* is forwarded. That test was green
> throughout. **No test file in the repo mentions `rendered_template_sha`.**

### c. The bug is still valid and is minting at roughly a dozen a day

The population query in this file returns **22 rows as of 2026-08-23**, of which **12 were born
2026-08-23** — after the phase-1 roll. All: `position=1`, `slot_name='hero'`, component `hero`,
`content_brief` present, `content_data` present, `build_status='deployed'`, `component_version_id`
NULL, no `data-component` in the HTML. `pages.sections` on an affected page reads
`["hero","generic-text-block"]` with `rebuild_policy='generic'`, so the plan really does put `hero`
first and these pages really are rebuilt automatically.

**Every row in the fleet whose HTML holds a tool fragment is typed `hero` — 22 of 22.** There is no
correctly-typed exemplar anywhere, and no passthrough/identity-function component exists to point
one at (`content_components` searched for a `{{.body}}`-shaped template and for
`adopted-fragment`/`passthrough`/`verbatim`: **zero rows**). The estate has no representation for
*"these bytes are the content"*, which is why one gets invented per page.

**And the tools are one Layer-2 miss from being replaced by a title band:** today's rows carry a full
hero `content_data` (headline, subheadline, cta_text, hero_url, …) beside 14.5KB of tool markup, so
`ContentDataCanFillTemplate(hero_template, …)` is **true**. What prevents the swap is Layer 2
splicing the stored bytes back on every rebuild — accidental protection, keyed on slot-name
equality, and the same mechanism that re-mints the false identity.

### d. Two findings that change the fix, one of which inverts the obvious one

**F1 — a SECOND severed hop.** `RerenderPageSectionsAction`'s *fresh-render* entry
(`rerender_page_sections_action.go:746`) never emitted the digest either, though `:662` calls
`RenderTemplate` and holds it 85 lines earlier. `bbe178309` wired only that file's *carry* path. The
rerender is the fleet's repair vehicle **and** the path that re-mints this population — so a page
mended by a rerender came out less well-provenanced than one left alone.

**F2 — adding the key alone would have been a REGRESSION.** Layer 2's splice (`:532`) replaces a
section's HTML with the stored tool but left the *incoming* section's `RenderedTemplateSHA` — the
digest of the hero band it just discarded — in place. Harmless only while the digest never arrives.
Deliver it and `resolveComponentVersionID` matches it against `hero` and stamps **the hero version
onto a whole interactive tool**: by this lane's own standard, worse than no stamp. The "just add two
keys to the allow-list" fix would have converted *no stamp* into *false stamp* on precisely these
rows.

### e. What shipped today — phase 0 (`a2e2fbac2`, `Council-Submitted: 73a638c7-f2a0-4a69-8145-96fc9a89c7bb`)

Fixed as the **class**, not the key: one declared carry list + a deny list naming each deliberate
drop with its reason + an AST parity test that fails when a producer sets an undeclared key **or**
when the save READS a key the carrier does not carry (the half that would have caught this on day
one). Plus F1's emission, and `adoptCarriedProvenance` for F2 at both Layer 2 carry arms. Register:
**CLC-028**, with CLC-026 corrected in place. Four mutations proven to kill their tests; full actions
suite green at committed HEAD.

**It does not fix this bug.** Identity is still inferred at birth and the mint continues. It makes
the fact the fix is made of actually reach the database, and stops that fact being false.

> ⚠ **Phase 0 is INERT UNTIL THE NEXT CHASSIS ROLL, and it must be verified at the artefact then** —
> `sections_metadata` carrying the digest on **both** producer paths, post-roll rows stamped while
> the pre-roll cohort stays exactly 0, and **the F2 guard holding**: this population is rewritten
> roughly daily, so within a day its re-minted rows must still read `component_version_id IS NULL`.
> A population row appearing *with* a hero stamp means the splice hygiene failed — stop there.

### f. What remains

- **Phase 2 — stop the mint.** Design settled (see the lane's PLAN): the damage is on the
  **component** axis (`component_id` → hero's schema, hero's template) while the landmine is entirely
  on the **slot** axis (Layer 2's match key, 420 Go references). So: **no slot name changes, at birth
  or at repair, and `pages.sections` is never touched** — the landmine is not managed, it is never
  armed. A fallback-adopted fragment that declares no `data-component` is *constructively adopted*
  onto a seeded `adopted-fragment` component (`{{.body}}`, verified byte-identical by an actual
  `RenderTemplate` round trip — `text/template`, so no escaping), which earns a real stamp through
  existing machinery; if adoption cannot complete, `component_id` stays NULL and the row is honestly
  unknown. Shipped behind an opt-in key defaulting **OFF** (owner ruling 2026-08-02 §2).
- **Phase 3 — the 22.** Owner-authorised in principle *after* the stamp is live and readable.
  Byte-preserving re-type (`component_id`, `content_data.body`, `component_version_id`), leaving
  `slot_name`, `position`, `rendered_html`, `rendered_html_digest` and `pages.sections` untouched.
  **Re-census on the day** — the population mints daily, so the target is that day's query result,
  not the number 22 — and exclude the three loancash verbatim pages and the two unrelated defects.

---

## 2026-08-24 — VERIFIED AT THE ARTEFACT: the stamp fires. Both phases council-APPROVED. The mint is not yet stopped.

Phase 0 rolled overnight. Every check from the plan, with the disconfirming result
it was written against:

| check [MEASURED 2026-08-24 ~13:30Z] | result | reads as |
|---|---|---|
| rows born since the 08-22 roll, stamped | **239 / 1051**, rising to near-100% per hour after 09:00Z (46/48, 117/119, 24/24, **58/58**) | the carriage is repaired |
| **control:** pre-roll cohort stamped | **0 / 987** | nothing backfills, exactly as designed |
| `component_versions` with `change_source='render_stamp'` | 39 rows across **39 distinct components — 1.00 each** | it settles; it has NOT become a log |
| the stamp is TRUE: stamped version's template = component's current text | 239 of 245 match | see below — the 6 are the mechanism WORKING |
| **the F2 guard:** 357 population rows stamped | **0 of 22** | ⚠ PENDING, not passed — see below |

**The step change is the evidence, not the count.** Rows born 02:00–08:00Z carry no
stamp; from ~09:00Z every hour is at or near 100%. That is the roll landing, and it
is not a number a backfill or a coincidence produces.

### The 6 "mismatches" are the stamp doing its job

All six are one component, `Illustrated Text Block`, and the timestamps settle it:
the version was born 10:17:06, the rows at 10:55:21, and **the component was edited
at 11:15:21 — after the rows existed**. So the rows correctly name the template
that produced them, and the component has drifted since. Before this change those
six were indistinguishable from rows rendered with the new template. **That is the
first live demonstration of what the stamp is FOR**, and a "mismatch" of this shape
is a stale row, not a detector fault (PLAN §4 said so in advance).

### ⚠ The F2 guard reads ZERO, and that zero is currently VACUOUS

0 of the 22 population rows carry a stamp — the required answer. But the demand
control says why that is not yet evidence: **0 of them were born after stamping
started.** No population row has been minted since ~09:00Z, so the guard has had
no opportunity to fail. It is **PENDING**, and the check to re-run is:

```sql
SELECT count(*) FILTER (WHERE pc.created_at > '2026-08-24 09:00:00Z') AS demand,
       count(pc.component_version_id) AS stamped_must_be_zero
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
 WHERE cc.name='hero'
   AND position(left(cc.html_template, position('{{' in cc.html_template)-1) in pc.rendered_html)=0;
```
Non-zero `demand` with `stamped = 0` is the pass. **Any population row appearing
WITH a stamp means the splice hygiene failed — stop and fix before arming phase 2.**

### Council

- Phase 0 (`73a638c7`) — **APPROVED**, 3 advisory objections, two answered in code.
- Phase 2 (`74e4c1fd`) — round 1 REVISE (both findings real: the identity carry was
  broader than the bug; and I had asserted "the page serves identically either way"
  without checking the rerender path). Round 2 **APPROVED**, 3 advisory, none high.

### State: the bug is NOT fixed

Population **22** as of 2026-08-24, one born today (12 were born on 08-23 — a rate,
not a total). Phase 2 is committed and approved but ships **default OFF and is not
armed**, so nothing is stopping the mint yet. Phase 3 is written and unapplied.

**What is now true that was not: the evidence the fix is made of is real and
checkable.** A stamped row means its component provably produced its bytes; every
one of the 22 is unstamped, and structurally cannot become stamped.

---

## 2026-08-24 — ARMED. The producer is fixed; the 22 existing rows are not.

Owner instruction: *"arm it"*. Applied by hand in order — `577` (seed the
`adopted-fragment` component, verified `{{.body}}` exactly) then `579` (arm). Both
verify blocks RAISE rather than SELECT, so a wrong result aborts instead of
committing; `579` additionally refuses unless the seed exists AND all six steps come
back armed.

**Independently re-read after applying** (recursive walk, not the migration's own
output): six live `save_page_sections` steps, all `adopt_unidentified_fragments =
true`.

### ⚠ Armed WIDER than this lane's runbook proposed, and the reason is a measurement

The runbook said arm one canary — `tool-recreation-handler`, the obvious producer,
the one declaring `expects_no_sections_metadata`. **Enumerating the surface first
overturned that.** The mint occurs on the HTML-parsing path (where the no-`<section>`
fallback lives), and the save falls through to HTML parsing whenever the metadata
path yields nothing — so **five of the six carriers have an `html_field` and can
mint**, not one:

| carrier | can mint | armed |
|---|---|---|
| page-build-handler, pageflow-builder, page-rebuild, site-work-orchestrator, tool-recreation-handler | yes | ✔ |
| page-rerender (no `html_field`) | no | ✔ — carry half only |

Arming the obvious one would have left four minting: the *"one call site of a shared
mechanism gets the rigorous fix while the mechanism stays generic elsewhere"* shape
that migration `575`'s own `bug_historian` objection names, and that gated this
lane's phase 2 round 1.

**Why arming wide is safe, in one checkable sentence:** adoption fires only on a
section produced by the no-`<section>` fallback (`SectionData.FallbackAdopted`, set
nowhere else) that *also* declares no `data-component`. An ordinary page has
`<section>` blocks, so the branch is unreachable for it — arming a whole-site builder
cannot change what a normal page stores.

### State at arming [MEASURED 2026-08-24 16:15Z]

population **22** · adopted rows **0** · population rows stamped **0** · per-page row
counts and slot lists for all 22 pages captured to
`scratchpad/baseline_before_arming.txt`, because the landmine's tell is a row COUNT
and that is only checkable against a before.

### What has NOT happened yet, and must not be assumed

**No adoption has been observed.** Arming is a config change; the first real proof is
a page coming through the path and landing correctly typed. Until then this is
"armed", not "working" — the same distinction that cost this lane a day when a
council-approved, rolled mechanism turned out to be writing nothing.

The three STOP conditions, each invisible to *"is the tool still there?"*:
- a page's row count goes **UP by one** (the landmine fired);
- a population row acquires a stamp (the splice hygiene failed);
- a new fragment lands with `component_id` NULL (adoption is refusing — read the
  `adopt fragment:` log lines, which name the arm that refused).

**The 22 existing rows are untouched and still wrong.** `578` repairs them, is
committed, and remains unapplied — it enforces its own preconditions rather than
trusting a runbook.

### 2026-08-24 17:00Z — the six `owned` pages are INCLUDED, and the seam has not run since arming

**Scope corrected on the owner's instruction.** `578` now targets all **22**
(dry-checked read-only: `targets_now=22`). The six `rebuild_policy='owned'` pages are
in, and the reason they were ever out was a misreading of mine:

> **`owned` does NOT mean "a human claimed this page".** The guard's own comment:
> such a page *"belongs to a tool/widget or is a runtime-fill shell"*
> (`save_page_sections_action.go:172`), and `create_report_page_action.go:176` writes
> the value in code. **172 of 704** pages estate-wide carry it. I read the column's
> name, inferred its meaning, and escalated that inference into a decision for the
> owner. `WRONG_CALLS.md` (9).

**And the misreading inverted the conclusion.** These six are the ONLY rows phase 2
can never heal — the owned-page guard returns at `:186`, adoption runs at `:397`, so
the save is refused two hundred lines before the new code is reached. Every other bad
row will now be repaired by an ordinary rebuild; these never will. A migration is
their sole route. They were the rows I was least willing to touch and they are the
only ones that cannot fix themselves. They are also the most durable targets: because
the pipeline refuses these pages, a row repaired here stays repaired.

**Checked before including them, because one condition would have made it unsafe.**
A page ships VERBATIM only when THREE things hold — `owned` ∧ exactly one component
row ∧ `content_data->>'deploy_mode'='verbatim'`. All six are owned with one row (two
of three) and **every one reads `deploy_mode` = NONE** [MEASURED 2026-08-24]: they are
assembled pages that work because assembly emits the single row's stored HTML, exactly
as §"Why this changes fix candidate 1" already recorded. Had the third condition held,
touching them would have risked the verbatim↔assembled flip, which **is the row count,
not a flag**. This migration adds no rows and sets no `deploy_mode`.
The three genuinely verbatim `loancash.co.uk` pages are excluded **structurally** —
none is bound to `hero`, so the predicate's first clause rules them out. Verified, not
inherited from the earlier lane's exemption note.

### ⚠ NOTHING IS PROVEN IN PRODUCTION YET, and the demand control says why

| check [17:00Z, ~45 min after arming] | value |
|---|---|
| adopted rows | **0** |
| population | 22 (unchanged) |
| population rows stamped | 0 (the F2 guard — still no demand) |
| **`SavePageSectionsAction` invocations, last 6h** | **0** |

**The seam has not executed since arming.** The 20 `page_components` rows written this
afternoon came from other writers (there are seven), and all of them were stamped —
6 of 6 in the 16:00 hour, 11 of 11 in 15:00 — so phase 0 continues to work. But the
adoption path has had no traffic at all, which means **the zero adoptions carry no
information about correctness.** Not a failure; no qualifying page was built.

That distinction is the whole reason the watch now reports
`seam_invocations_5m` alongside the counts: v1 could not tell *"armed, working, and
nothing qualified"* from *"seam never ran"*, and those demand identical follow-ups
from a reader and opposite ones from an operator.

**What to do with a zero next time:** read the demand control first. All-zero
invocations ⇒ keep waiting. Non-zero invocations with zero adoptions ⇒ that is a real
signal, and the `adopt fragment:` log lines name which arm refused.

### ⚠ CORRECTION 2026-08-24 19:00Z — "the seam has not executed since arming" was WRONG

The entry immediately above states `SavePageSectionsAction` invocations in the last
6h as **0**, and concludes the seam had no traffic. **Both are false.** That figure
came from a `kubectl logs … | grep` and the logs did not know.

The database says otherwise, and it is the authority here:

| [MEASURED 2026-08-24 19:00Z] | |
|---|---|
| rows saved **through `save_page_sections`** since arming (16:15Z) | **209** |
| of those, stamped | **209** |
| latest row | 18:49Z — nine minutes before the check |

Every row written this afternoon carries `content_brief`, which is that action's own
fingerprint. The seam is not idle; it is one of the busiest things on the estate.

**Why the log check lied is not worth chasing** (two pods, a fresh roll 24 minutes
earlier, retention, level) — the point is that it was the wrong instrument. A
`kubectl logs` grep is not a census of anything, and this file already carries that
lesson for build provenance: *an empty result means "not in range", not "absent".*
I applied it there this morning and not here.

### The conclusion survives, on a better control

Zero adoptions is still fully explained, but by the RIGHT measurement:

| the demand control that actually bears on adoption | |
|---|---|
| adoption **candidates** since arming — pages saved with exactly one component row, no `<section`, no `data-component` | **0** |
| rows saved since arming on single-row pages, at all | **0** |

**No qualifying page has been built since arming.** Adoption fires only on the
no-`<section>` fallback; 209 ordinary multi-section pages went through the seam and
correctly produced no adoptions. That is the expected reading, and it is now measured
rather than inferred.

**The watcher was rebuilt twice for this** and lives at
`bugfix_357_component_identity/watch_357_adoption.sh`: its demand control is now
DB-based and asks *"did a qualifying page arrive?"* rather than *"did the seam run?"*
— the seam runs constantly, so the latter answers a question nobody asked. A tick now
reads:

```
adopted=0 population=22 population_stamped=0 adoption_candidates=0 saves_since_arming=209
```

Every number needed to interpret the zero is on the line.

### 2026-08-24 19:35Z — the first flagged candidate was NOT a defect, and it sharpens the diagnosis

The watch reported `adoption_candidates=1`. Opened it rather than reacting to it, and
it is a page the platform got **right**:

| `agritec.uk/tool-sfi26-revenue-stacker`, born 19:17Z | |
|---|---|
| `slot_name` | `tool-sfi26-revenue-stacker` |
| component | `tool-sfi26-revenue-stacker` — **its own bespoke component, not `hero`** |
| `content_brief` | NULL ⇒ **did not come through `save_page_sections`** |
| stamped | no |

**The tool pipeline types its rows correctly.** `create_tool_component` /
`deploy_tool` mint a bespoke component per tool and bind the row to it, so a page
built that way is never mislabelled. It also never reaches adoption, which lives
inside `save_page_sections` — so this row could not have been adopted and should not
have been counted as an opportunity.

**That narrows the defect usefully.** The mislabelling is not "tool pages" in
general; it is specifically a tool fragment reaching **`save_page_sections`' HTML
fallback** without the tool pipeline having created a component for it. Two routes
produce a tool page and only one of them is broken.

**Control corrected (watcher v5):** a candidate must now also carry `content_brief`
— the `save_page_sections` fingerprint. Without that clause the control reports
opportunities that never existed, which would eventually have been read as "adoption
is failing" when nothing had been asked of it. Re-reads **0** candidates, 321 saves
since arming.

⚠ **Noted, not chased:** that correctly-typed row is **unstamped**, because
`create_tool_component`/`deploy_tool` are among the five `page_components` writers
that do not write `component_version_id`. Phase 0 covers the `save_page_sections`
path only. It is the already-named follow-up — five other writers, no watch on them —
and this is the first live instance of it.

---

## 2026-08-25 — the F2 guard PASSED with real demand; the bug itself is unchanged

**Cold-start doc for this lane:**
`docs/agent_docs/docs024_key_docs_latest/bugfix_357_component_identity/HANDOFF_2026-08-25_continue_here.md`

### The guard is settled — it was vacuous yesterday and is proven today

At **09:08:20** `vetcomparison.uk/index` was re-minted through `save_page_sections`:
Layer 2 spliced its stored 11,326-byte tool back into a freshly generated `hero`
section, and the row was written. **It received NO stamp — and it is the ONLY
unstamped save of 571 since arming.**

| | |
|---|---|
| saves through the seam since arming | **571** |
| stamped | **570** |
| unstamped | **1 — this row** |

That is the discrimination the check needed. A stamp here would have named the *hero*
template as the producer of a whole interactive tool — the "worse than no stamp" case
phase 0 exists to prevent. Yesterday's zero had no demand and the bug file said so;
today's has demand and holds. **Treat F2 as settled.**

### The bug is unchanged, and one row was re-minted

Population **22**. `born_since_arming = 1` — the row above. It kept its `hero`
binding because `carriedIdentity` only carries identity for `adopted-fragment` rows
(the council's round-2 narrowing), and this row has never been adopted. **For an
existing mislabelled row that is BY DESIGN: phase 3 is its remedy, not phase 2.**

### ⚠ THE TOP OPEN QUESTION: phase 2's adoption path has never fired

`adopted = 0`, `adoption_candidates = 0` since arming. 571 ordinary saves went through
the armed seam and correctly produced none. **So phase 2's central claim is unproven
in production**, and this lane has already lost a day to a mechanism that was
approved, rolled and doing nothing — do not assume it works because it is armed.

And there is positive evidence of a route it may not cover: **every multi-row affected
page has rows from DIFFERENT saves** (2 rows = 2 distinct `created_at`;
`vetcomparison/index` = 4 rows, 4 distinct times) [MEASURED 2026-08-25]. The
no-`<section>` fallback emits exactly ONE section, so those hero rows did not come
from a single fallback save. Rows accumulate across saves rather than being replaced.

**What must be established next:** by what route does a NEW mislabelled row appear,
and does phase 2 intercept it? Suggested start — `page_component_history` for
`vetcomparison.uk/index` and one `mortgagecalculator` tool page, to find which save
introduces the `hero` binding.

### Phase 3 will refuse to run today, correctly

`578` requires an organically adopted row carrying a stamp before it will touch
anything. `adopted = 0`, so it raises and aborts. **Do not weaken that check** — it is
the owner's "once option 1 has been built" expressed as code. Order:
**make adoption fire once → verify that row → then run 578.**

### Fresh build 2026-08-25

`agent-chassis-67fd9c76f5` carries the capability, probed at the running binary with a
must-be-present and a must-be-absent control, both correct.

---

## CONTRIB 2026-09-02 (mortgagecalculator_couk_adoption lane) — 701 changes the acceptance ladder's subject key, and will orphan eight criteria fences on this site including its own pilot's

Answering your CONTRIB of 2026-09-02 into our directory. **We have no objection to 701 — we think
it should go ahead, and it fixes something for us that your design notes do not claim.** One
consequence needs a line of SQL alongside it.

### The good half you may not have costed: 701 makes nine invisible tools verifiable

The acceptance ladder's eligibility predicate (`discovery_checks/tool_eligibility.go`,
`toolEligibilityWhere`) admits a component when **either** it is `component_level='tool'`, **or**
it is the *sole* component on a `page_type='tool'` page that has no tool-level component. Run
against mortgagecalculator.co.uk today, it returns **9 of 18** tool pages. [MEASURED 2026-09-02]

The nine it cannot see are exactly your adopted population minus `tool-simple`: affordability,
bridging-loan, equity-release, fee-analyser, overpayment, portfolio, rate-forecaster, repayment,
stamp-duty. Each is multi-component (the calculator under the `hero` identity, plus a
`generic-text-block`) with no tool-level component — so it satisfies neither clause. **Neither Tier
2 nor Tier 4 has ever looked at them.** 701 gives each a `component_level='tool'` row, which admits
all nine under clause (a). That is a real gain and worth stating in your design notes.

### The half that needs SQL: the subject key moves, and every existing fence is keyed the old way

The ladder keys its subject on `toolSubjectKeyExpr`:

```sql
CASE WHEN cc.component_level = 'tool' THEN cc.function
     ELSE regexp_replace(p.name, '^tool-', '') END
```

Today these pages resolve through the `ELSE` arm. Post-701 they resolve through the `THEN` arm, and
701's own census sets `new_function = 'tool-<slug>'` (line 204ff: `tool-affordability`,
`tool-bridging-loan`, … `tool-simple`, `tool-stamp-duty`). So the key moves **`<slug>` →
`tool-<slug>`** — and all eight current `doc_plans` rows for this site are keyed the old way:

```
bridging-loan  equity-release  fee-analyser  overpayment
rate-forecaster  repayment  simple  stamp-duty      (subject_type='tool', is_current)
```

**Every one of them is orphaned by 701**, and the failure is silent in the direction that looks
like success: Tier 2 goes on writing `needs_criteria` notes, Tier 4 emits nothing, and nobody gets
an error. That is the RUNBOOK §14 trap ("a PLAN under the wrong key produces no error") arriving
through a route §14 does not describe — the key is not mistyped, **the page's shape changes under
it**.

⚠ **`tool-simple` is your designated pilot (`scope=pilot`), and it is the one page here whose
fence works today** — it is eligible right now under the sole-component clause with key `simple`.
So the pilot is the case where the regression is most visible, and "the pilot looked fine" will
not catch it: the page renders identically, the tool behaves identically, and the only symptom is
a verification that quietly stops happening.

### What we suggest, and what we are not doing

Re-key the eight rows in the same transaction as 701, guarded the way the rest of 701 is:

```sql
UPDATE doc_plans SET subject_key = 'tool-' || subject_key
 WHERE subject_type = 'tool' AND is_current
   AND subject_key IN ('bridging-loan','equity-release','fee-analyser','overpayment',
                       'rate-forecaster','repayment','simple','stamp-duty');
```

Two cautions on that statement, neither of which we have resolved for you:

1. **`idx_doc_plans_current` is UNIQUE on `(subject_type, subject_key) WHERE is_current`.** If a
   `tool-<slug>` current plan already exists for any of the eight — from another site sharing the
   function name — this raises rather than silently merging. That is the right failure, but it
   wants checking before apply, not during.
2. **The eight keys are not site-scoped.** `doc_plans` has no `site_id`; the subject key is
   fleet-wide. `stamp-duty` and `equity-release` in particular are plausible keys on
   `loanandmortgagecalculator.co.uk` and `remortgagecalculator.uk`, which carry same-named tools.
   **Enumerate which pages each of the eight currently resolves to before re-keying** — if a fence
   is shared with another site's page, re-keying it here moves it there too.

We are **not** applying this ourselves: it is your migration, your pinned census, and your pilot,
and a hand-edit from us between your census and your apply is exactly what your md5 guards exist to
refuse. Ping us and we will do the enumeration in (2) for our three domains.

*Evidence and full working:
`docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/NOTES_mortgagecalculator_couk.md`
`## 2026-09-02 (b)`. Separately relevant: the ladder is currently misreporting on instance-scoped
tools fleet-wide (criteria name pre-`bugs_closed/283` bare ids, both checkers match ids exactly) —
`090` run correlation `7177c2d6-fe22-40c4-b9bc-b53f93ec59c9`. That is upstream of this and does not
change the advice above.*

### CONTRIB addendum, same day — the pilot case is now concrete, not hypothetical

Measured after the note above, cluster-free, against the live pages: **`tool-simple` is the only
tool on mortgagecalculator.co.uk that can be verified today**, and it is your designated pilot.

The site splits on one property. All 8 tool pages carrying a `component_level='tool'` component are
instance-scoped (`id="c-tool-…"`) — ladder-eligible, but their fences name pre-conversion bare ids
and so are unsatisfiable (`bugs_open/441`). All 10 adopted pages have bare ids and **valid** fences —
all 48 selectors present on the live pages, checked — but 9 of the 10 are ladder-ineligible.
`tool-simple` is the single page that is bare-id **and** eligible: valid fence, and the ladder can
see it.

**701 moves its key from `simple` to `tool-simple` and the fence is orphaned.** So the pilot is the
one case where the regression is both real and maximally invisible: the page renders identically,
the calculator behaves identically, `rendered_html` is byte-identical by your own guards, and the
only thing that changes is that the site's sole working verification stops running. **Nothing in a
pilot's verification-at-the-DB-row-and-the-served-page will show it** — the check that would is
"does `simple` still appear in the eligibility query's output under a key that has a current
`doc_plans` row", and that is the one to add to the pilot's acceptance.

This does not change our advice — ship 701, it makes nine invisible tools verifiable — but the
re-key `UPDATE` is no longer a tidy-up to schedule afterwards. It belongs in the pilot.

---

## CLOSED 2026-09-02 — population 0, verified at the served artefacts

**The complaint:** 22 live `page_components` rows declared themselves the shared `hero`
component while storing whole interactive tools. **The state at close:** the bug's own
predicate returns **0**; all 22 rows carry per-tool adopted identities.

**The repair (owner decision "Option B", migration 701, council APPROVED r3 corr `df6c1b41`):**
per row, a `content_components` entry whose `html_template` IS the stored bytes
(`created_from='adopted'`, `component_level='tool'`, one RFC_036 §9.3 fork), with the plan
repointed in the same transaction at both levels (`site_plan_sections` + `pages.sections`)
and slot = plan element = cc.name aligned. Applied by the owner's hand: pilot
(`tool-simple`) 19:03Z green on every clause, remainder 21 rows ~19:35Z, COMMIT clean, all
guards passed. Backup: `page_components_backup_357b_20260902` (kept).

**Verified at the artefacts (2026-09-02 ~21:10–21:5xZ):**
- 22/22 served pages: HTTP 200 with browser Accept, interactive controls present, scripts
  present, **zero `data-component="hero"`**; per-domain invented-URL controls all 404.
- 21/22 stored rows byte-identical to the pinned census through the day's rebuild waves;
  the 22nd (vetcomparison/index) was REBUILT ORGANICALLY at 21:15Z by the news wave — and
  that rebuild is the close's strongest evidence: the repointed plan resolved the adopted
  component, **rendered the template**, reproduced the tool to within one trailing newline
  (`post + '\n' = pre`, proven by equality), kept `created_from='adopted'` through a full
  delete-and-reinsert save, and the served page still carries the working tool. The
  regeneration-is-a-no-op property — the reason Option B was chosen over 578's untestable
  carry dependency — held in vivo, unprompted.
- The migration's own 16 filed rerenders shipped correct artefacts but took `render_page`
  (the §10 row-404 reason-vocabulary trap — prose in a parsed field); corrected in the HOLD
  file for any re-run. The organic rebuild above is what exercised the real route.

**What this lane leaves behind:** phase 2 (stop the mislabel at birth) live since 08-25 with
the growth-guard as its regression tripwire; `bugs_closed/408` (found and fixed en route);
`bugs_open/406` (the sibling refusal defect — DIAGNOSED, fix still owed, shared-seam,
council-gated); migration 578 superseded-for-this-population (its file untouched, per its
own header note in 701); the walker-family census; and three parked vetcomparison contrast
items whose natural home is now the adopted component's template (their lane's design pass
is queued). Full trail: `docs/agent_docs/docs024_key_docs_latest/bugfix_357_component_identity/`.

---

## CONTRIB 2026-09-03 (mortgagecalculator_couk_adoption) — post-close, informational: the fence re-key we flagged was not applied, and a second effect nobody predicted has appeared

**Not a reopen.** 701 did what it said: all 11 of our rows repointed, bytes preserved at apply time,
tools working. This records two downstream facts for the next reader of this file, both measured
2026-09-03 on mortgagecalculator.co.uk.

**1. The subject-key re-key we raised in our 09-02 CONTRIB was not done, and the predicted effect
happened.** The ladder keys on `cc.function` once a component is `component_level='tool'`, so all 18
of our tool pages are now eligible (**up from 9 — a real gain from 701**) under keys `tool-<slug>`.
But our 8 `doc_plans` fences are still keyed `<slug>`, so **all 8 are orphaned**: nothing loads them,
Tier 2 will write `needs_criteria`, Tier 4 emits nothing, and no error appears anywhere. This is our
lane's to fix and we are fixing it — recorded here only so the file does not read as if the
consequence was hypothetical. Re-checked before acting: no `tool-<slug>` current plan collides, and
none of the 8 keys serves any site but ours.

**2. The effect we did NOT predict, and which the closing note's "bytes unchanged" does not
cover.** 701 created the adopted components with **instance-scoped templates** (`{{.InstanceID}}-`)
while preserving the pre-existing rendered bytes. Those two states are consistent only until
something re-renders. This morning's rebuild wave re-rendered **5 of the 10** (08:46–08:49) and
their element ids changed on the served page — `amt` → `c-tool-simple-amt`. So the site is now
**half-converted: 5 tools scoped, 5 bare, all under scoped templates**, and the remaining 5 will
convert whenever they next render.

**Nothing is broken by it** — we checked all ten for dangling JS bindings and found **0**, so the
converter rewrote bindings alongside ids, exactly as designed. **But each of those five re-renders
silently invalidated that tool's acceptance fence** (`bugs_open/441`). The transferable point, and
the reason it is worth a paragraph in a closed bug: **"bytes unchanged, md5-verified" was true at
apply time and stopped being true at the next render, because the migration changed the TEMPLATE the
next render would use.** A byte-equality guard proves the migration did not move anything; it cannot
promise the next render will not. Worth stating in any future adoption migration's verification
section.

No action requested. → `bugs_open/441`, and this lane's NOTES 2026-09-03.
