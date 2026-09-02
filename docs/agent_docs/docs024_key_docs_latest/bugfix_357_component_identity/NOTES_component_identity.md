# NOTES — `bugs_open/357`, component identity vs stored bytes

Append-only, newest at the bottom. Technical log: evidence, commands, what the system said,
and every misstep.

---

## 2026-08-22 — session start: 277 was the ask, 357 is the work

Asked to take up `bugs_open/277`. **It is not open** — it moved to `bugs_closed/` earlier the same
day (`4a193b390`, owner ruling). Verified the close rather than trusting the file:

```sql
SELECT COALESCE(result->>'route','(UNROUTED)'), status, count(*)
FROM site_work_items WHERE item_type='required_fields_missing' GROUP BY 1,2;
--  no_content_data/needs_human_review 27 · asset_sourced 2 · no_plan_owned 1 · (UNROUTED)/complete 37
```
Matches 277's stated close exactly, and its served-bytes proof still passes today
(cubic-bezier `<code>ease-in-out</code>` = 1; gas-unit-converter `<tr` = 6).
Owner picked the named successor: **`bugs_open/357`**.

## The population is 22, not 9 — and 357's own query says so

357 states nine rows [MEASURED 2026-08-22 ~10:00Z] and prints a "re-runnable" population query.
**Run today, that query returns 22.** The nine were the subset that also had a parked
`required_fields_missing` work item (all of them one-component pages); the query as written is
unrestricted and picks up multi-component pages too. Newest row: **`vetcomparison.uk` `index`,
created 2026-08-22** — a site homepage, born today. So this is a live producer, not a June artefact.

⚠ **The file's stated population and the file's own query disagree.** Anyone re-measuring 357 from
its query and comparing to its table will think 13 rows appeared overnight. They did not.

## Which writer — closed by fingerprint, which the 090 run never tried [MEASURED 2026-08-22]

`63d4d1a7` came back UNVERIFIABLE with "which writer assigns the component_id" open. The writers
leave distinguishable marks on the row, so the row itself answers it:

| writer | position | content_brief | build_status |
|---|---|---|---|
| `save_page_sections_action.go` | `i+1` | **written** (`{slot} section`) | deployed |
| `deploy_tool_action.go` | 2 | never | deployed |
| `create_tool_component_action.go` | 2 | never | deployed |
| `adopt_verbatim.go` | 0 | never | approved |

**All 22 rows: `position=1`, `content_brief.section_guidance='hero section'`.** Only
`save_page_sections_action.go` writes `content_brief` at all. The writer is settled.

⚠ **A near-misstep, recorded because I nearly filed it as a second discriminator.** The nine older
rows have `rendered_html_digest` NULL and the thirteen newer ones have it set — which reads like two
writers. It is not: the INSERT writes `md5($3)` unconditionally, and the column postdates the older
rows (`bugs_open/229` / IMP-052). **The split is by DATE, not by writer.** A digest-based
"two writers" claim would have been confidently wrong.

## Root cause, in committed code, cited

1. `saveSectionsExtractFromHTML` (`save_page_sections_action.go`) — tool/game pages emit
   `<div class="tool-page">…</div>` with no `<section>` element, so the section regex matches
   nothing. The documented fallback stores the whole fragment as ONE section named `"section"` —
   the sentinel for *identity unknown*. Its own comment says it stays `"section"` "unless a planned
   name exists".
2. `enrichSectionsWithPlannedNames` — treats `"section"` as needing enrichment and assigns
   **`planned[Position-1]`** from `pages.sections`. Purely positional.
3. `enrichSectionsWithComponentIDs` — resolves that name to the shared component's UUID.
4. The single INSERT persists `slot_name='hero'`, `component_id=<hero>`,
   `rendered_html=<the whole tool>`.

**`hero` is planned FIRST on all 22 pages** (`["hero","generic-text-block"]`,
`["hero","tool-list"]`, `["hero","info-card-grid","latest-news","call-to-action"]`), so
`planned[0]` is always `hero`. The tool is called a hero because hero is first in the list —
**the sentinel meaning "I do not know what this is" is converted into a confident wrong answer.**

## What 357 does not record, and what makes it worse than filed

- **13 of the 22 carry a complete hero `content_data`** (`headline`, `subheadline`,
  `background_image`, `hero_url`, `cta_text`, `secondary_cta`) alongside tool HTML. 357's nine all
  have `content_data` NULL, which is why its "the parked state is the protection" argument holds
  for them — **it does not hold for these thirteen.** `ContentDataCanFillTemplate` is already true
  on them, so the regeneration 357 warns a backfill would arm **is already armed**, with no backfill.
- **16 of the 22 are `rebuild_policy='generic'`** — rebuildable, not owned.
- The rows were **born** wrong: `updated_at = created_at` on all 22. No later overwrite.
- `vetcomparison.uk/index.html` serves `<div class="tool-page">` at the hero position and
  `data-component="hero"` **zero** times [MEASURED 2026-08-22, http 200, 44,496 bytes]. The hero
  copy its writer produced is stored and never rendered.

## The discriminator for a guard — measured, not guessed

357's fix candidate 2 warns a born-wrong guard "must be measured against the whole live table
before arming" because the static-template-prefix test also fires on legitimately drifted templates.
Correct: that test flags **158** rows fleet-wide. So test the component's own self-declaration
instead — `data-component="…"`, which `saveSectionsExtractFromHTML` already trusts as the identity
carrier in this same file:

```sql
tmpl_attr := substring(cc.html_template from 'data-component="([^"{]+)"')
html_attr := substring(pc.rendered_html from 'data-component="([^"]+)"')
-- flag when tmpl_attr IS NOT NULL AND html_attr IS DISTINCT FROM tmpl_attr
```

| | rows |
|---|---|
| components declaring the attribute | 149 of 339 |
| template and HTML **agree** | **1,550** |
| template and HTML **disagree** (both declared) | **0** |
| template declares, stored HTML silent | **27** |
| static-prefix mismatches (the noisy test) | 158 |

**Zero false positives fleet-wide, and the 1,550 agreements are the demand control** — the
predicate could have come out otherwise, 1,550 times, and did. The 27 are a strict subset of the
158: the 22 hero rows, 3 `loancash.co.uk` Ported Page rows, one `leopardessconsulting.co.uk`
`blog-listing_pre_037`, and one `idea.uk` `tool-list` row whose `rendered_html` is **zero bytes**.
The ~131 drift-only rows the noisy test flags are **not** in it.

⚠ **Stated as a limit, not glossed:** this predicate is silent for the 190 components whose
template declares no `data-component` at all, and 97 of the 100 `Ported Page` prefix-mismatches are
NOT flagged. It is a **refuse-on-certainty** test, not a census of everything wrong. That is the
right trade at a write seam and the wrong one for a sweep.

## [UNMEASURED] — whether the destruction has already fired

A page born this way, then rebuilt from its hero `content_data`, would serve a title band and
**leave the population query**, so the survivors cannot tell me about the casualties. There is no
systematic `page_components` history to check (only ad-hoc `_backup_*` tables). Candidates exist —
`mortgagecalculator.co.uk` `game-fact-finder` is `page_type='game'` with 2 components, 4,363 bytes
and no `<script>` — but "never had a tool" and "had one and lost it" are indistinguishable from
current state alone. **Not claimed in either direction.**

## Diagnosis loop

Re-filed sharper (357 itself says the first run stalled because the symptom asked one question
while pointing at three files): intake `f7aedef7-0bee-4c68-8cde-c86ac552e3e2`,
**`RUN_CORRELATION_ID=e580b34a-d284-4f80-ac96-81af1c4adaba`**. It is asked the one thing the row
fingerprints cannot settle: which leg pairs a genuine hero `content_data` with the tool's bytes,
since the `saveSectionsExtractFromHTML` fallback sets no `ContentData` at all.

---

## 2026-08-22 (later) — ⚠ CORRECTION: I published "already armed" and it is REFUTED. The rerender is the AUTHOR, not the threat.

Recorded here in full because the correction is more useful than the finding it replaces.

**What I wrote** (in `bugs_open/357` §3, and said to the owner): the 13 rows with a complete hero
`content_data` are one rebuild away from having the tool replaced by a title band.

**What refutes it.** `vetcomparison.uk` `index` has **six completed `page_rerender_index_…` items
between 08-19 and 08-22** and the tool still serves. Then the timestamps:

```
rerender item   created 2026-08-22 08:44:51 → completed 08:50:19
all 4 page_components rows on the page       created 08:50:12
```

The rows were written **inside** the rerender window. `RerenderPageSectionsAction` DELETEs and
re-INSERTs; it carried the tool HTML forward and re-attached the hero identity and hero
`content_data`. **The rerender re-mints the mismatch on every pass.**

**Three failures, worth separating:**

1. **Imported a mechanism from an adjacent bug without testing its reach.** `ContentDataCanFillTemplate`
   is 277's arming mechanism. Its only non-test caller is
   `discovery_checks/check_literal_markdown.go:429` — a **detector's classifier**, not a rebuild
   gate. I ran that grep, saw "one caller", and went looking for the rebuild path anyway with the
   conclusion already formed.
2. **Predicted a codepath's behaviour instead of looking for its output.** The prediction was
   defensible from source. The system had already run the experiment six times and written the
   answer into `site_work_items`.
3. **Read `updated_at = created_at` as history.** I had it in these NOTES as "born wrong, never
   touched since". It means the opposite — this writer re-inserts rather than updates, so every row
   is *newly* born, repeatedly. **A column equal to its sibling is evidence about the WRITER's
   method, not about stillness.**

**What survives unchanged:** the identity/bytes disagreement on 22 rows; `save_page_sections` as the
writer; hero planned first on all 22; the positional enrichment in `enrichSectionsWithPlannedNames`;
the `data-component` predicate and its 1,550/0/27 census.

**What this changes about the fix:** the population is **self-renewing**, so stock repair without a
flow fix is wasted — the next rerender reproduces it. Flow before stock, and any repair needs a
re-check *after* a subsequent rerender, not just after the repair.

**STILL OPEN, assumed in neither direction:** whether a full page rebuild *through the writer*
(rather than a rerender) renders the hero template over the tool. Rerender demonstrably does not.

**Also established while checking this** (`rerender_page_sections_action.go`): the content pre-check
at ~380–445 exempts self-contained tools via `isSelfContainedSection` (~1270), keyed on
`component_level=='tool'` **and** an empty `input_schema`. `hero` is `component_level='section'`
**with** a schema [MEASURED], so the exemption never fires for these rows — they are processed as
ordinary sections end to end. That is the seam where a tool-shaped payload could be recognised and
is not.

## 2026-08-22 (later still) — the persistence mechanism, and the general form of the defect

`carryStoredSection` (`rerender_page_sections_action.go:1150-1169`) forwards the stored row as one
bundle: `rendered_html`, `stored_slot_name`, `component_id` and `content_data`. On the receiving
side `extractSectionsFromMetadata` prefers `stored_slot_name` over everything else when naming the
section. **Identity, bytes and content are propagated together, verbatim, and nothing compares
them.**

That the vetcomparison row was CARRIED rather than rendered is not an assumption: the hero template
opens `<section class="hero" data-component="hero"`, and the stored 11,326 bytes contain no
`data-component` at all, so they cannot be that template's output. Only a carry branch emits stored
HTML. **Which** carry branch (component unresolvable / section not ready / empty template) is not
established — three reach `carryStoredSection` and the logs for that run have rotated.

### The general form, which is the part worth fixing

> A component row's **identity** (`component_id`, `slot_name`), its **bytes** (`rendered_html`) and
> its **content** (`content_data`) are written, carried and re-written as an atomic bundle by every
> path in the page-composition system — and **no seam anywhere asserts that the three agree**.

Two independent consequences, and the second is why this is not a data-repair job:

1. **A pairing that is wrong once is wrong for ever.** The carry preserves it faithfully.
2. **It is refreshed on every pass**, so the population is self-renewing and cannot be repaired
   ahead of the producer.

Where identity is *invented* rather than carried, it is invented **positionally**
(`enrichSectionsWithPlannedNames`: `planned[Position-1]`), from a plan, without looking at the
bytes — which is how a tool comes to be called a hero in the first place. Origin and persistence are
therefore two different code paths, and a fix that addresses only one leaves the class alive:
fixing the origin leaves 22 rows re-minting for ever; fixing the carry leaves the next tool page
mislabelled at birth.

**The one place both converge** is `save_page_sections_action.go`'s single
`INSERT INTO page_components` (~line 999) — documented in-code as *"the single INSERT every
page-composition path flows through"*, and already carrying a precedent guard
(`sectionIsUnresolvableStub`, `bugs_open/039`) that refuses a bad row, raises a typed work item and
continues. That is the seam, and the precedent is the shape.

---

## 2026-08-22 — COUNCIL ROUND 1: **REVISE**, and checking the gating objection found a defect in my plan worse than the objection

Trail `62aac6c2-996f-4b5d-8f8f-72e3daf4c82e`. 13 reviewers, 4 abstained, gated by `bug_historian`.

### The finding that resized the submission — mine, not theirs, but theirs is why I looked

`bug_historian` (HIGH, edit 1) asked whether the honest-unknown degrade path had been tested against
the Layer 2 carry-forward, reasoning that if carry-forward decides protection *by inspecting the
component/template*, a re-typed row might stop qualifying.

**Its stated mechanism is wrong**, and that is checkable in one read: Layer 2's PROTECTION keys
purely on the bytes — `interactiveHTMLSQL(col)` is `ILIKE` over markers on `rendered_html` (:1698),
`sectionHTMLIsInteractive(html)` is `strings.Contains` over the same markers (:1677). Neither reads
`component_id`, `slot_name` or a template. Re-typing cannot remove Layer 2 protection.

**But the question was the right one**, because Layer 2's **MATCHING** is:

```go
for i := range sections {
    if sections[i].ComponentName == p.slot { matchedIdx = i; break }   // :517 — slot NAME, and nothing else
}
```

Rename the stored slot to `adopted-fragment` while an incoming plan-driven section set still names
`hero`, and the match fails. Control falls to `default:` — *"Slot dropped entirely — re-append the
tool so it survives"* — which **appends** the tool as an extra section while the incoming hero title
band is **also** saved. **The page gains a hero band and the tool moves position.**

> **So my "byte-preserving re-type" would have changed what 22 live pages serve**, on four sites,
> as a side effect of a change I described in the submission as preserving bytes exactly. It needs
> `pages.sections` updated in the same transaction, or identity-agnostic matching — and either is
> its own round with its own measurement.

That is the second time today the estate's protective machinery has been the thing I mis-modelled
(the first is the correction above). **The pattern to carry: a mechanism that preserves BYTES may
still key its decisions on IDENTITY, and those are different questions.**

### Round 2: withdrawn, kept, and answered

**Withdrawn** — Layer A (fallback typing), the `adopted-fragment` seed, the positional-enrichment
narrowing. All three change what is persisted, and all three are blocked on the slot-matching
question above. This also answers `guardian`'s three objections (unconditional changes, three
simultaneous change vectors on the busiest save path) without argument.

**Kept** — the seam guard (record-only; refusal behind an opt-in seeded on nobody), the detector, the
register entry, the tests. None changes any persisted output.

**Objections answered by checking:**

| seat | objection | what the check said |
|---|---|---|
| `prior_art` | is `adopted-fragment` dormant-machinery duplication? | **No.** `SELECT name FROM content_components WHERE regexp_replace(html_template,'\s','','g') ~ '^\{\{\.?[A-Za-z_]+\}\}$'` → **0 rows**. No byte-preserving passthrough exists |
| `prior_art`, `reuse_agent` | is the claimed inline duplication real? | **Yes** — `dataComponentRe := regexp.MustCompile(...)` at **:1392 AND :1498**, byte-identical, used at :1427/:1463/:1509. Both deleted and repointed; `grep -c 'dataComponentRe :='` must read 0 after |
| `editquality` | does `ON CONFLICT` protect idempotency? | **Yes** — `content_components_name_key` UNIQUE on `(name)` exists |
| `architecture` | state the optional-key count before adding the Nth key | **It cannot be read, and that IS the finding.** The audit covers 123 actions against N=10 and lists **99 more as `uncounted`** — `save_page_sections` is one of them (no `RegisterActionInputSpec`), so a new key enters counted as ZERO. The `retract_asset_files`/`publish_site` trap exactly |
| `tooling_provenance` | do not hand-edit the concept index count | **Correct.** Stored counts are RETIRED; the checker reads only the first 4,000 bytes and its own line says *"any that reappears is the finding"*. CLAUDE.md's "update the index count" is the anti-pattern here. Row added, headline untouched, two-way `comm` parity run instead |
| `bug_historian` | Layer D is detection-only at five other writers | **Accepted as residual exposure, not closed.** Named in the submission |

Round 2 dispatched on the same trail (`RESUBMIT_CORR`), run envelope `3ad2cff2`.

---

## 2026-08-22 — COUNCIL ROUND 2: **REVISE** again, and it is right that the resized plan fixes nothing

Gated by `editquality` (HIGH), seconded by `bug_historian` (HIGH). Both say the same thing:

> the mistyping mechanism and the Layer 2 carry-forward *"both continue to fire exactly as before.
> The diagnosed corruption is not stopped by this plan"* — and the 22 pages *"get a work item filed
> and nothing else, indefinitely. This is exactly the documented 'detected but never blocked'
> recurrence — `bugs_open/079` (phantom links detected, never blocked) and `bugs_open/083`."*

**That is a fair hit and I accept it.** Round 1 was too broad (it would have changed what four live
sites serve as a side effect); round 2 over-corrected into pure paperwork. Six seats approved
cleanly, and the three that objected converge on the same gap.

**The `architecture` seat named the actual answer**, and it is an RFC rather than a third round:

> *"This is the fourth distinct heuristic-identity mechanism layered into
> `save_page_sections_action.go` (stub detection, Layer 2 slot-name carry-forward, shrink/floor
> guards, now data-component matching). Recommend an RFC scoping a durable component-instance
> identity (e.g. a stamped identity token independent of `slot_name`/`position`) rather than a fifth
> heuristic layer next time this class recurs."*

Every failure this lane has found is a symptom of one thing: **identity is inferred — from position,
from a plan, from a slot-name string match, now from an HTML attribute — and never stamped.** A
fifth inference is not the fix. Filed as `RFC_046`.

`bug_historian`'s second objection is the same point from the other end: pinning Layer 2's
name-only matching with a test *"leaves the mechanism itself as exploitable as before for any future
author who renames a slot without reading this test"*.

## The diagnosis run (`e580b34a`) — UNVERIFIABLE again, and worth more than its verdict

Stopped at iteration-cap, no fix proposed. But it did something I had not: it read
**`page_component_history`** and found that on `tool-portfolio` the hero `content_data` was already
present at the `2026-08-15T13:52:39` archive event and unchanged through the `19:53:52` write that
produced the live row — *"content_data looks carried-forward, not freshly paired, which cuts against
the 'second path pairs identity+content_data+tool bytes' half of the hypothesis."*

**That supports the Layer 2 carry account and weakens my "second path" hypothesis** — which is what
I filed the run to test. A run that refutes half your hypothesis for the price of one is a good run.

## ⚠ CORRECTION — "there is no systematic `page_components` history" is FALSE

I wrote, and published in `bugs_open/357` §5, that the casualty question was unanswerable because
*"there is no systematic `page_components` history to check (only ad-hoc `_backup_*` tables)"*.
**`page_component_history` exists** — 26,965 rows over 558 pages, 2026-03-16 → 2026-08-22, with
`rendered_html`, `slot_name`, `op` and `created_at`. It is `bugs_closed/229`'s page-side artefact
archive. I missed it because my table listing was `| head -30` and the alphabetical `_backup_*` and
`page_components_bak_*` tables filled every line.

## The casualty census — and the honest answer is NO, not the one I announced

With the history table, the question I marked `[UNMEASURED]` is answerable. Slots that were
interactive in history and are not interactive now:

| | slots |
|---|---|
| still interactive (**control**) | **182** |
| no longer interactive | 17 |
| slot gone entirely | 39 |

**I said "so the destruction HAS happened" on those counts. Then I opened the 17, and it does not
hold:** 15 are `ported-page` slots on `loancash.co.uk` guides whose byte count **GREW**
(e.g. 6,929 → 9,300) — a normal rebuild replacing ported HTML that happened to contain a `<script>`,
not a loss. One (`fundamentallyai.com` `platform-log-index`, 14,256 → 11,470) shrank modestly. One
is worth a look on its own merits: **`webdesign.co.uk` `learn-ai-builders-content-first`, 8,855 →
3,781, a 57% shrink on 2026-08-15** — recorded as a follow-up, not as a 357 casualty.

**None of the 17 is a page in the 357 population.** So: no evidence that this mechanism has
destroyed a tool. The correct statement is *"searched with a working control and found none"*, which
is a much stronger claim than the `[UNMEASURED]` it replaces — and the opposite of what I said
sixty seconds earlier. The 39 vanished slots have not been opened and are NOT claimed either way.

---

## 2026-08-22 — note from the editorial-design-uplift / 035 lane (not this lane's author)

Appended under the coordination rules (2026-07-29 §3: a shared mechanism's other consumers must be
told). `features_open/035_FEATURE_component_hierarchy.md` (component composition, written today)
originally planned to read `page_components.component_version_id` as a **render-input PIN**
("NULL = follow the library's current template"). **Your RFC_046 stamping makes that read wrong by
design** — once renders stamp provenance, every row is non-NULL and a pin read would silently
freeze each instance at whatever last rendered it. 035 D6 is corrected in place: the pin becomes a
**separate opt-in column** (`pinned_component_version_id`, default NULL), so your stamp stays pure
record and — a property worth having — *"was the pin honoured"* becomes the mechanical equality
`stamp == pin` wherever a pin is set. Nothing for you to do now and nothing of yours is blocked or
edited by us; our P1 (composition read-path in `assemble_from_library.go` +
`rerender_page_sections_action.go`) is **deliberately deferred while your uncommitted work is in
those files** — we will re-read the seams (incl. `RenderedTemplateSHA` on `RenderTemplate`) after
your commit rather than build against a moving surface. Contact doc:
`docs024_key_docs_latest/editorial_design_uplift/NOTES_editorial_design_uplift.md` (08-22 tail).

---

## 2026-08-22 — OWNER RULED `RFC_046` OPTION 1, and phase 1 is BUILT (commit `bbe178309`)

> *"Option 1 please, we can change the existing pages once option 1 has been built."*

Two separate things settled: the estate will **stamp** identity rather than infer it, and the 22 rows
are authorised in principle **after** the stamp exists. Recorded in the RFC itself, including what
the ruling does NOT decide (it is not an implementation approval, not a licence to repair now, and
not a decision on the stamp's shape).

### The finding that made phase 1 cheap

`page_components.component_version_id` **already existed and had never been wired** — 0 of 1,930
rows populated, no Go writer, no Go reader [MEASURED 2026-08-22], against a control
(`rendered_html_digest`: 1,623 of 1,930, 65 code references) proving the same method sees a live
column. `component_versions` is live and holds the **full template text** (369 of 369 rows). So
option 1 is reuse, not addition: no new column, no new table, no new concept.

> ⚠ `component_version_id` is a column on **two** tables. `site_plan_sections`'s is the live one and
> accounts for nearly every grep hit. A bare `grep -rn component_version_id` makes the dormant one
> look busy — qualify by table. This is why it stayed dormant in plain sight.

### What shipped

`RenderTemplate` (the ONE render spelling, 15 non-test call sites) sets
`RenderContext.RenderedTemplateSHA` — an out-field on `AbsentRequiredFields`' precedent, no I/O.
`RenderComponentAction` emits it; `extractSectionsFromMetadata` carries it; `resolveComponentVersionID`
resolves it to a `component_versions` row at the single INSERT, in the same statement as the bytes;
`carryStoredSection` forwards an existing stamp so a rerender cannot downgrade a known row to
unknown. **Inert: 0 readers, so it cannot change what any page serves — and it does not fix 357.**

Verified at committed HEAD, not the working tree: `git archive HEAD` extracted and built clean, and
`go test ./platform/orchestration/actions/...` passes there (`dddd4b6c5`).

### Two things the work itself taught

**The obvious spelling of the drift test does not work, and it took writing it to see why.** Written
as *"expect no `component_versions` query"* it passes against code that DOES infer: sqlmock refuses
the unexpected call, the resolver sees a query error, and it returns no-stamp **for the wrong
reason** — a guard in series standing in for the one under test. Rewritten with a **decoy** row made
available and attractive, plus a paired control asserting the decoy is genuinely on offer, because
"it didn't take the decoy" is meaningless if the decoy was never served.

**11 pre-existing tests broke, and one of them was pinning the wrong thing.** The `bugs_open/229`
digest guard's regex ended at `'deployed')`, which froze the INSERT's **column list** rather than the
property it protects. Adding a column reddened it without weakening anything — and a pin that
reddens on a safe change teaches the next author to widen it carelessly. Tail moved to `'deployed'`,
and mutation-proven that it still bites (anchor 247→242 chars, `contains md5($3)` false).

### ⚠ MISSTEP: my first mutation proof mutated the wrong occurrence

I ran `src.replace("md5($3)", "$9", 1)` and the guard still passed — which I read aloud as *"the
mutation passed the guard, that's a red flag"*. The guard was fine. **The file has two `md5($3)`,
at lines 994 and 1016, and `replace(..., 1)` took the first — which is OUTSIDE this anchor.** The
correct mutation is inside the matched span, and it kills the test. This is the "first occurrence
wins" trap on source-scanning tests, which is in my own memory index, hit while proving a
source-scanning test. Logged in `WRONG_CALLS.md`.

### Council round 3 submitted (`364a9fd4` envelope, trail `62aac6c2`)

Submitted as **step 1 of the owner-ruled two-step**, not as a substitute for fixing — and saying
plainly that round 2's gating objection is still literally true of this code. What changed is the
ruling and the named step 2, not the rhetoric. Also flags for the council that if they think a
write-half should not be reviewed apart from its readers, the code is already on the shared branch,
so that changes the review unit rather than what is on main.

---

## 2026-08-22 (evening) — phase 1 is LIVE, unexercised; council round 3 REVISE with 6 clean approvals; the writer census was wrong

### The deploy, verified at the artefact

Chassis rolled **2026-08-22 15:10:31Z**. The `build provenance` startup line had already scrolled out
of `--tail=3000` (the known landmine: an empty result there means *not in range*, not *unstamped*),
so I probed the running binary for the **capability** with four arms:

| probe | result | expected |
|---|---|---|
| `observed at render` (new SQL literal) | PRESENT | PRESENT |
| `component version: template changed` (new log line) | PRESENT | PRESENT |
| `self-contained tool section` (pre-existing) — **positive control** | PRESENT | PRESENT |
| `zzz_never_shipped_literal_357_qqq` — **negative control** | ABSENT | ABSENT |

**The stamp is NOT yet exercised, and I am not claiming it works.** [MEASURED 2026-08-22 17:50Z]
`page_components` rows created since the roll: **0**. So `since_roll_stamped` is 0 of 0 — *pending*,
not passing. The control half does hold: **0 of 1,940** pre-roll rows are stamped, which is the
disconfirmable half (a backfill would have moved it, and nothing backfills). `component_versions`
holds 372 rows, **0** with `change_source='render_stamp'`. A capability with no live caller has an
untested dependency on its environment; this one is waiting for the first page save.

### Council round 3 — REVISE, but the shape changed

**Six clean approvals** (guidelines, tooling_provenance, render_guardian, constitution, mission,
prior_art_librarian), and **editquality and architecture moved to APPROVE** with low-severity notes —
both had objected in earlier rounds. Gated by `debug_historian` (HIGH) on a livespec/landmine
concern.

**The gating objection does not apply, and that is checkable rather than arguable:** this change
ships no migration and alters no live DB object. `git show --name-only bbe178309` has no `.sql`,
nothing under `sql_for_agents/`, no livespec file; `grep -c component_version_id
platform/livespec/livespec.go` = **0**; `go test ./platform/livespec/...` passes. The landmine fires
on migrations that move a guarded live object — the edit here widened a *source*-scanning regex over
a Go file, which has no live-object sibling.

### ⚠ The `reuse_agent` seat was RIGHT, and I had published the error three times

I claimed `component_versions` was *"written by one path"*. It has **five INSERT sites across four
files as of 2026-08-22**. Corrected visibly in the PLAN §10.2, in register CLC-026, and logged in
`WRONG_CALLS.md` (8th). The live table would have told me for free — **26 distinct `change_source`
values**; a single-writer table does not have 26 provenance labels.

**What survives, stronger than what it replaced:** 277's nine components had no versions not because
one Go writer missed them, but because a template edited by a **hand-written SQL migration** passes
through **no Go writer at all** — which is exactly why the stamp must be taken at RENDER time rather
than at edit time. And the error hid a real defect: every writer allocates `version_number` by its
own `MAX+1`, so the version-number race a seat raised against my get-or-create is a property of the
**table**, predating this change by four writers.

### New finding no seat raised

`page_component_history` has 15 columns and **none** is `component_version_id`. The archive **drops
the stamp**, so a row's provenance is lost the moment it is archived — precisely the forensic case
`bugs_closed/277` needed and could not have. Named as a phase-2 item.

### Round 4 dispatched (`81a38cc4` envelope)

Answers each round-3 objection with the check that settles it. **The dispatch receipt I filed as a
landmine two hours ago worked in use**: `fix_plan` artifacts on the trail went 3 → 4 within 75
seconds, which is how a landed dispatch is now told apart from the silently dropped one that cost
this lane 95 minutes earlier today.

---

## 2026-08-23 — the stamp was severed at a hop nobody measured; phase 0 fixes the CONTRACT, not the key

Picked the lane up after ~23h idle. First question asked, deliberately not "what shall I build next"
but **"did yesterday's approved, rolled capability actually do anything?"**

### The answer, with its control

[MEASURED 2026-08-23] **820** `page_components` rows born since the 08-22 15:10:31Z roll, **0**
stamped. `component_versions` with `change_source='render_stamp'`: **0**. `component version:` log
lines in 24h: **0** — which is *consistent with* the resolver's silent early return
(`src.renderedSHA == ""` returns with no log), not with it refusing. Pre-roll control: 0 of 1,146,
correct, nothing backfills.

The decisive measurement was of the **middle** of the chain, not either end:

```sql
WITH r AS (SELECT collected_data FROM orchestration_states
           WHERE created_at > '2026-08-22 15:10:31Z' AND collected_data::text LIKE '%rendered_template_sha%'),
     sm AS (SELECT jsonb_path_query(r.collected_data,'strict $.**.sections_metadata[*]') AS elem FROM r)
SELECT count(*) AS elems,
       count(*) FILTER (WHERE elem ? 'rendered_template_sha') AS with_sha,
       count(*) FILTER (WHERE elem ? 'component_id')          AS with_control
FROM sm;
-- 546 | 0 | 546
```

**546 / 0 / 546.** The copy runs; the key is dropped. `extractSectionFromMap` rebuilt the metadata
map from a hand-written six-key allow-list and `rendered_template_sha` was not on it. Both ends were
verified yesterday and the middle was assumed.

### Why this is a contract fix and not a two-line fix

This hop ate `stored_slot_name` in `bugs_open/189`. The remedy then was to add that key and pin it
with `TestExtractSectionFromMap_ForwardsStoredSlotName` — which asserts **its own key** and was green
the entire time our stamp was being dropped beside it. **No test file in this repo mentions
`rendered_template_sha`.** Three keys were being dropped at once.

### Two findings the lane's own plan did not have

**F1.** `RerenderPageSectionsAction`'s *fresh-render* entry (`:746`) never emitted the digest, though
`:662` calls `RenderTemplate` and holds it 85 lines earlier. `bbe178309` wired only `carryStoredSection`.
The rerender is the repair vehicle *and* the re-minting path.

**F2, and this one inverts the obvious fix.** Layer 2's splice swaps the section's HTML for the
stored tool but kept the incoming render's digest. Inert while the digest never arrived — deliver it
and it resolves against `hero` and stamps the hero version onto the tool. **The "just add the two
keys" fix would have turned no-stamp into FALSE-stamp on exactly the 357 population.**

### ⚠ MISSTEP: my mutation PASSED and I nearly wrote that down as the test working

Full-action sqlmock test asserting the spliced row's stamp bind is nil: green. Mutation removing
`RenderedTemplateSHA = ""`: **also green.** The resolver swallows its own query errors and returns
"no stamp", so a stale digest and no digest write the same `nil`. A guard in series standing in for
the property under test. Fixed by extracting `adoptCarriedProvenance` and pinning it directly, plus
an AST test that the splice still calls it — both mutations now kill their tests. Logged in
`WRONG_CALLS.md` along with two others (a `git apply` that exited 0 while skipping the patch, because
the scratchpad is itself a git repo; and a concept-register id another session already owned).

### ⚠ The suite is red in the working tree and green at HEAD, and that is not my code

`refused_link_targets.go` / `_test.go` are **untracked** files from another lane, mid-write (the test
references a function that does not exist yet). Every test result in this section was taken in an
isolated `git archive HEAD` tree with my changes applied, never in the working tree.

### ⚠ Committing `v3_site_actions.go` needed care, and a plain pathspec commit would have broken HEAD

Another lane has uncommitted WIP in that file passing an **8th** argument to
`applyWorkItemFailureLadder`, whose *committed* signature takes **7**. `git commit <path>` takes the
file from the working tree, so it would have shipped a HEAD that does not compile — and every
`make build-*` builds from HEAD. Committed a HEAD+my-hunk version of the file by pathspec, then
restored their hunk; verified afterwards that the working-tree diff is exactly their 5/1 hunk and
that `git archive HEAD` builds and passes.

### Shipped

`a2e2fbac2`, `Council-Submitted: 73a638c7-f2a0-4a69-8145-96fc9a89c7bb`. Register CLC-028; CLC-026
corrected in place (it claimed "not yet rolled" and it had been rolled for a day, writing nothing).
Landmine appended and dispatched. **Phase 0 does not fix 357** — the mint continues at ~12/day.

---

## 2026-08-23 (later) — phase 2 built and shipped OFF, phase 3 written and held

### Phase 2, and the cut that makes safe and effective the same plan

`slot_name` and `component_id` are different facts read by different consumers.
`component_id` joins to `content_components` (`check_required_fields_missing.go:80`
reads `cc.input_schema` through it) and carries 100% of this bug's damage;
`slot_name` is Layer 2's match key, **420** Go references as of 2026-08-23, and
carries 100% of the landmine. So: correct the component, never the name.
`enrichSectionsWithPlannedNames` is untouched, `pages.sections` is untouched — the
landmine is not managed, it is never armed.

Constructive adoption, not a sixth inference: bind to `adopted-fragment`
(`{{.body}}`) with the fragment as `content_data.body`, **only after rendering that
template and comparing bytes**. `RenderTemplate` uses `text/template` (checked at
the import), so nothing escapes — but the comparison is what makes it a fact, and
it means a later edit to the seeded template stops adoption rather than corrupting
rows. Adoption is a real render, so the row earns a genuine stamp through the
resolver phase 0 repaired this morning.

Opt-in `adopt_unidentified_fragments`, default OFF. It governs BOTH halves —
adoption, and Layer 2 carrying the stored `component_id` with the stored bytes —
because adoption alone does not survive a rebuild: the incoming section carries the
plan's identity and the next rebuild would re-mint `hero` over an adopted row.

### ⚠ MISSTEP: a fixture one column short made the splice silently not run

Adding `component_id` to the Layer 2 preload widened the query to five columns
while `layer2PreloadWith` still supplied four. `rows.Scan` then fails, the loop
logs and `continue`s, **and the splice never runs** — at which point the provenance
assertions pass while testing nothing. Only the re-append case noticed, because it
asserts a row COUNT that a skipped splice cannot satisfy. The spliced test now pins
the INSERT's `rendered_html` bind to the TOOL bytes rather than `AnyArg`, and that
vacuity mutation is proven to fail it. Second time today a green test turned out to
be worthless; both are in `WRONG_CALLS.md`.

### ⚠ MISSTEP: 578's first predicate silently dropped the six rows the bug is about

Written as `ILIKE '%tool-page%'` AND "not owned", it selected **16 of 22**. Running
it read-only against the live database — rather than trusting it — showed the six
missing were the original gamesdesign tool pages, `tool-ttk-calculator` among them,
which is the worked example in the bug file's own opening paragraph.

Two separate faults in one clause. The interactivity test was a **second, narrower
spelling** of a definition the estate already has, so it now mirrors
`interactiveStructuralMarkers`/`interactiveControlMarkers` (the markers
`interactiveHTMLSQL` renders). And the six are `rebuild_policy='owned'` — which is
also why they have been stable since June while the other 16 re-mint — so skipping
them is an **owner decision**, not a technical exclusion, and it is now a per-row
`RAISE NOTICE` plus a count line instead of a `WHERE` clause nobody would read.

### State

- `a2e2fbac2` phase 0 — **council APPROVED** (`73a638c7`), 3 advisory objections, two answered in `c2edcd6aa` (producer DISCOVERY rather than a hand-maintained list; the later-substep recovery case).
- phase 2 — committed, `Council-Submitted: 74e4c1fd`, verdict pending at time of writing.
- `b702e9d04` phase 3 — written, HELD, not applied.
- Diagnosis `1ca712e3` — **5 bundles, no verdict**; the same scope-not-narrowing shape as this lane's earlier `090`. The root cause rests on first-hand measurement with controls, declared as the owner ruling of 2026-07-31 allows, not on the loop.

---

## 2026-08-24 — the stamp is LIVE AND WRITING, verified with its controls; phase 2 approved on round 2

Phase 0 rolled overnight (~09:00Z, visible as a step change: hours before it are
0-stamped, hours after are 46/48, 117/119, 24/24, **58/58**).

Controls, because a non-zero count is not a working mechanism:
- **pre-roll cohort 0 of 987** — nothing backfills, as designed;
- **churn 1.00 versions per component** (39 rows / 39 components) — it settles;
- **stamp truth**: 239 of 245 name the component's current template. The 6 that do
  not are all `Illustrated Text Block`, whose component was edited at **11:15**
  against rows written at **10:55** from a version born **10:17**. The rows are
  right and the component drifted — a STALE ROW, which PLAN §4 predicted and which
  is the first live demonstration of the stamp's value.

### ⚠ The F2 guard's zero is currently VACUOUS, and I nearly recorded it as a pass

0 of the 22 population rows are stamped — the required answer. But the demand
control says **0 of them were born since stamping started**, so the guard has had
no opportunity to fail. PENDING, not passed. This is the
[[a-post-fix-zero-needs-a-demand-control]] shape and I caught it only by putting
the demand count in the same query as the result — which is now the recorded
re-run.

### Council

Phase 2 round 1 REVISE → both findings real (identity carry too broad; the
"serves identically either way" claim unchecked against the rerender path). Round 2
**APPROVED**. Phase 0 APPROVED. Both trails: `73a638c7`, `74e4c1fd`.

### ⚠ MISSTEP (housekeeping, but it cost real disk): I hand-rolled `git archive HEAD | tar` six times

Six ~450 MB trees, **2.7 GB**, left behind in one session — `head357`, `t357`,
`headcheck`, `headcheck2`, `finalhead`, `closehead`. Each `rm -rf` in my pasted
recipe cleared the tree that run was about to USE, never the previous one, so the
cleanup half never fired even though it was in every command. CLAUDE.md gained
`scripts/verify-head-builds.sh` and an explicit prohibition on the hand-rolled
recipe on 2026-08-24 — the same day, from the same class of waste. Reaped; use the
script.

---

## 2026-08-24 — ⚠ CORRECTION: F1's line shipped in ANOTHER lane's commit, twelve minutes before mine

Filed to this lane by the `283`/RFC_032 session as a CONTRIB, checked by me rather
than accepted:

| commit | time | contains |
|---|---|---|
| `024303681` (283 lane, "retire the ComponentID bindings") | **08-23 18:56:55** | **my** `entry["rendered_template_sha"]` block |
| `a2e2fbac2` (mine, "357 phase 0") | 08-23 19:08:12 | everything else in phase 0 |

`git log -S'entry["rendered_template_sha"] = rc.RenderedTemplateSHA' -- rerender_page_sections_action.go`
names **theirs**. So F1 reached HEAD as a same-file passenger in their commit and
went live on **v1.0.1332 (rolled 08-24 09:39)** — which is precisely the step change
I measured at ~09:00Z, so the two records agree.

Nothing is lost, forward-only holds, and the code in HEAD is exactly what I wrote.
But **my phase-0 commit message claims F1 and the commit does not contain it** — the
file was already at its final content when my pathspec commit took it.

### ⚠ MISSTEP: "my files show clean" is not "my commit carries my change"

My end-of-session check was `git status --porcelain` over my files. It read clean —
**for the wrong reason.** The file was clean because *someone else had committed my
change*, not because I had. On a shared tree those are different facts and the
status output cannot tell them apart.

The check that discriminates is the one the other lane used:
`git log -S'<a literal from your change>' -- <file>` — it names the commit that
actually introduced the line. Logged in `WRONG_CALLS.md` (8).

### The four HIGH objections in council `e8c7414c` are answered, and were already covered

Because my hunk appeared context-free inside their reviewed diff, five seats read it
as an undisclosed mechanism and the round was vetoed. Four HIGH objections were about
**my** seam and addressed to nobody:

- *"why a third provenance field, when `content_hash` and `rendered_html_digest` exist"* —
  it is not a third field. `rendered_template_sha` is an in-flight out-field resolved
  into the pre-existing, dormant `component_version_id` (0 of 1,930 rows, no reader,
  no writer, measured 08-22). The siblings answer *"are these bytes reproducible from
  this content"*; neither answers *which template text produced them*, which is the
  fact that vanishes when a template is edited (`bugs_closed/277`: 15 rows permanently
  unrepairable for exactly that absence).
- *"targets a column that does not exist on `page_components`, so the write fails or is
  silently dropped"* ×3 — correct that no such column exists, and nothing writes one.

**Already reviewed:** trail `62aac6c2` (phase 1) put this exact design — "REUSE, NOT
ADDITION" — and was **APPROVED 08-22 18:02Z**, a day BEFORE `e8c7414c`. Plus `73a638c7`
(phase 0) and `74e4c1fd` (phase 2), both approved. Three approved rounds; that round's
seats saw a bare hunk and objected to the only reading available to them. Disposition
sent to the 283 lane for their resubmission.

---

## 2026-08-25 (afternoon) — driving the phase-2 route on purpose: two site adoptions

**Why this session exists.** The 08-25 handoff's §4 named the lane's top priority: phase 2
is armed on all six carriers and has **never fired**, `adoption_candidates = 0`, so
`adopted = 0` says nothing about correctness. Phase 3 (`578`) refuses to run until one
organically adopted, stamped row exists. The owner's instruction was "run an adoption",
naming `cv1.co.uk` and `lampenkap.com` as candidates.

### The two meanings of "adoption" turn out to be the same event

I first read "adoption" as this lane's fragment adoption and looked for a page to drive.
Neither candidate domain was in `sites` at all, which is the wrong shape for that reading
and the right shape for the OTHER one — the site-adoption pipeline. Both readings converge,
and the convergence is the finding: **adopting a site that has an interactive page is the
supported way to make phase 2 act.** Traced in code, not inferred:

```
082_submit_domain_unified.sh <domain> --from <url>
  -> site-adoption-orchestrator -> site-adoption-agent
  -> apply_adoption_plan_action.go:719   if len(page.Features) > 0   (page is INTERACTIVE)
       -> work item needs_tool_recreation, handler tool-recreation-handler
  -> tool-recreation-handler: recreate_tool -> validate_tool -> save_page_sections
       (declares expects_no_sections_metadata -> the metadata path yields nothing)
  -> save_page_sections_action.go:344    len(sections)==0 -> HTML fallback on html_field
  -> saveSectionsExtractFromHTML:1561    no <section>, no <html>/<!doctype>
       -> whole fragment stored as ONE section, FallbackAdopted = true
  -> enrichSectionsWithComponentIDs:1631 armed + FallbackAdopted + no data-component
       -> adoptFragmentSection -> binds to `adopted-fragment`, content_data.body, stamp
```

The comment at `save_page_sections_action.go:10` said so all along — *"Fallback path: regex
parsing of assembled HTML (for adopted sites or older pipelines)"* — and this lane had never
read it as an instruction for how to GENERATE demand.

### THE FLAG THAT WOULD HAVE MADE THE WHOLE EXERCISE SILENTLY POINTLESS

`--fidelity locked` is the natural-looking choice for "adopt this site" — it is the
byte-preserving path, and it is the one this estate reaches for when adopting a site it
means to KEEP. **It cannot fire phase 2, by construction.**
`apply_adoption_plan_action.go:486` returns early on `fidelityLocked` into
`adopt_verbatim.go`; the tool-recreation routing lives at `:708`, two hundred lines below
the return. A locked adoption would have run, succeeded, produced a live site, and left
`adopted = 0` — and every number on the watch script would have looked exactly as it does
now. **The demand control and the mechanism would both have read correct while nothing was
ever tested.** Recorded as a landmine.

The recreate path (default `high` for adopt mode) is therefore mandatory here, and it is
not a preference: it is the only path that reaches the code under test.

### Candidates, measured 2026-08-25 (both fetched, not assumed)

| | cv1.co.uk | lampenkap.com |
|---|---|---|
| index | 19,545 B, **0 `<script>`, 0 controls** — static | 12,065 B, 1 `<script>`, 7 controls — a lux calculator (`runSim()`) |
| second page | `example.html`, 18,839 B, 28 checkboxes + `localStorage` JS | none (single-pager) |
| language | English/UK | Dutch |
| live copy dated | 2026-08-02 | 2026-04-20 |

Both are already served from the estate's own bucket (Cloudflare + `x-amz-id-2`, identical
header shape to `garden-tools.uk` and `mortgagecalculator.co.uk`), i.e. they are hand-built
pages on our own hosting. **Neither appears in
`portfolio_positioning/HOSTED_domains_for_owner_decision.md`**, so neither was covered by the
2026-08-20 "22 free, 3 protected" ruling — the owner naming them today is the authorisation.

cv1 is the better experiment because it exercises BOTH routes in one run (static page ->
page-build-handler, interactive page -> tool-recreation-handler), which is exactly the
comparison the handoff's §4 asks for. Owner chose both, cv1 first.

### Baseline, immediately before dispatch [MEASURED 2026-08-25 11:28:43Z]

`adopted=0 population=22 population_stamped=0 saves_since_arming=605`
(571 at 10:36 -> 605 at 11:28, so the seam is alive and the zero is not a dead pipeline).
Armed check re-run in the same breath: six rows, all `true`, via the recursive
`jsonb_path_query` form — a top-level `jsonb_each` sees three of six.
Seed present and correct: `9d4b922b-a548-4ca2-987c-ecacc7904b1f`, `is_active`,
`btrim(html_template) = '{{.body}}'` true.

### Dispatched

| domain | correlation | orchestration | landed |
|---|---|---|---|
| cv1.co.uk | `468cb727-d2c7-4299-b332-3fc36c0996c6` | `8c83d368-3c47-4b74-a3c5-e440add32e1a` | `EXECUTING_STEP|spawn_adopter` |
| lampenkap.com | `a3e1a948-0979-4b0f-8592-cfbd979d9899` | `e8909b4a-adfe-4984-9ca5-0e77b6555b05` | `EXECUTING_STEP|spawn_adopter` |

Both **LANDED**, not merely published: `082` calls `kafka_verify_landing` after
`kafka_publish_checked`, so a row proves something picked the bytes up. The `kcat -P`
silent-drop trap does not apply to this entry point.

`082_submit_domain_unified.sh` is **not executable** in the tree (`-rw-rw-r--`). Run it as
`bash <path>`; do not chmod it, that is a tree change nobody asked for.

cv1's site row was created within ~40s: `8c3e9118-2455-4f0d-b01a-5dcde13dcf99`, status
`active`, build_status `pending`.

### MISSTEP, same session: an append that silently did nothing

I ran `cd <lane dir> && cat >> NOTES... <<'EOF'` from a shell whose cwd was ALREADY the lane
directory. The relative `cd` failed, `&&` short-circuited, and **the heredoc was consumed by
the shell without `cat` ever running** — no error I would notice, because the failure message
was about `cd`, and the `wc -l` on the next line printed a plausible line count for a file
that had not changed. It read like a successful append.

The check that discriminates is not a line count, it is the CONTENT:
`grep -c "<a phrase from what you just wrote>" <file>` — 0 means it did not land. Cheap, and
it cannot be fooled by a file that was already long. Same shape as this lane's earlier
`git status` misstep: an instrument that reads plausible for the wrong reason.

### THE RESULT: the route ran twice, with perfect inputs, and a DIFFERENT guard refused the save

Both cv1 pages were classified interactive and both were dispatched to
`tool-recreation-handler`. Both recreations completed. **Neither wrote a row, and
`adopted` is still 0** — but the reason is not phase 2, and this is the session's real finding.

**The fragments were exactly what phase 2 wants.** Read out of the orchestration record
(`validation_result.clean_html`) before the pods were reaped [MEASURED 2026-08-25]:

| precondition | `index` | `tool-example` |
|---|---|---|
| bytes | 26,271 | 21,265 |
| no `<section>` → the fallback arm fires | ✔ | ✔ |
| no `data-component=` → adoption not declined | ✔ | ✔ |
| no `<html`/`<!doctype` → fallback guard passes | ✔ | ✔ |
| carries `tool-page` wrapper | ✔ | ✔ |

So `saveSectionsExtractFromHTML` produced exactly ONE `FallbackAdopted` section on each —
which is independently confirmed by the floor's own arithmetic below, whose numerator is 1.

**Then `save_page_sections` refused the whole save.** Two `save_refused_incomplete` items,
both `needs_human_review`:

```
index        11:40:12Z   planned sections 25% (1 of 4)   prune_floor_ratio=0.50
tool-example 11:42:41Z   planned sections 33% (1 of 3)   prune_floor_ratio=0.50
```

`pages.sections` for the two pages, written by `apply_adoption_plan` itself:
`index` = `["hero","features","call-to-action","contact-form"]` (4);
`tool-example` = `["generic-text-block","features","call-to-action"]` (3). Both pages hold
**0** `page_components` rows.

### The mechanism, and why it is a defect rather than a guard doing its job

`apply_adoption_plan_action.go:719` routes a page to `tool-recreation-handler` when it has
interactive features, **and the same action writes that page's multi-entry `pages.sections`
plan.** `tool-recreation-handler` declares `expects_no_sections_metadata`, so its save can
only ever reach the HTML fallback, which by construction emits **exactly one** section.
`measurePageSectionCompleteness` (`save_sections_prune_floor.go:148`) then divides that 1 by
the planned count and `prune_floor_ratio=0.50` refuses.

**The route chooser and the floor's denominator are written by the same action, in the same
transaction, and they disagree about what the page is.** Any adopted interactive page planned
with **3 or more** sections is therefore unsaveable: 1/3 = 33% and 1/4 = 25% are both below
0.50. A page planned with 1 or 2 clears it (1/1 = 100%, 1/2 = 50%, and the floor trips on
`ratio < floor`, so exactly 0.50 passes).

### The cross-check that turns this from a theory into the explanation

If that is the mechanism, the 22 existing mislabelled rows should live almost entirely on
pages planned with **≤2** sections — because those are the only ones whose one-section save
could ever have completed. [MEASURED 2026-08-25]:

| planned sections on the page | rows in the 357 population |
|---|---|
| 1 | 1 |
| 2 | **20** |
| 4 | 1 |

**21 of 22.** The floor has been silently *selecting* which tool pages get a row at all:
≤2 planned → saved (and mislabelled, which is 357); ≥3 planned → refused entirely, page left
empty, item parked for a human. The single `planned=4` row predates or bypassed the floor
(the floor's own cohorts were measured 2026-07-31).

### And it is not a cv1 curiosity

`save_refused_incomplete` items sitting in `needs_human_review`, all history [MEASURED
2026-08-25]: **32 rows, 2026-07-31 → 2026-08-25, across ~14 domains.** The one-of-N shape
recurs on named tool pages — `webdesign.co.uk/tool-llm-cost-calculator` (1 of 4),
`fundamentallyai.com/tool-model-approach-selector` (1 of 3),
`mortgagecalculator.co.uk/game-fact-finder` (1 of 4), `finetuning.uk/blog` (1 of 3).
⚠ Several older rows have EMPTY cohort captures; the `planned sections` cohort postdates them,
so a blank is a reason-string format difference and **must not** be read as a different cause.

⚠ **`site_work_items` is a rolling window** — joining `needs_tool_recreation` to `pages` finds
only the two cv1 rows, because the historical items have been archived out. Do NOT read that
as "only two pages were ever routed to tool recreation" ([[a-closer-census-cannot-see-what-it-succeeded-at]]).

### So the handoff's §4 question now has a better answer than expected

*"By what route does a NEW mislabelled row appear, and does phase 2 intercept it?"* — the
route is reachable, its inputs are perfect, and **phase 2 never gets the chance to record its
answer because the save is refused after the binding is decided.** `adopted = 0` was never
evidence about phase 2's correctness; it was evidence about a guard two hundred lines further
down. Filed through the diagnosis loop (intake `f2fa4b9e-28b6-4f45-9ffa-2627c2031af0`,
**RUN_CORRELATION_ID `fbdaca97-a97e-41e6-b422-2475521e6a6c`**) rather than asserted here,
because it is a structural claim about a mechanism outside the symptom (owner ruling
2026-07-31).

### Phase 3 is still correctly blocked

`578`'s precondition 2 counts adopted rows carrying a stamp. It is still 0, so the migration
will RAISE and abort. **That is the check working**, and it must not be weakened — the shape
it would mint has still never been demonstrated in production.

### Evidence trail, for whoever picks this up

- spawned agent pods are **ephemeral** (`agent-<type>-<hash>`), so the `adopt fragment:` log
  lines for these two runs are **gone** — the pods were reaped within minutes. Anything you
  want from a run's logs must be captured live; the DB record (`orchestration_states`,
  `site_work_items`) is what survives. My log monitor was watching `-l app=agent-chassis`,
  which is the WRONG pod set for spawned agents, and its silence meant nothing until I ran a
  control that must have matched and got 0.

### Two same-file passengers in one commit, in both directions — neither lost anything

The commit-scope report on `b0cf6e501` flagged `shared-ledger-not-appended`: **1 line removed
from `LANDMINES.md`**, a fleet-wide append-only ledger. I only ever appended, so this was
worth chasing rather than waving through.

**It was the `bugs_open/386` lane's in-place correction, riding in my commit.** They replaced
one line about hand-written claim floors with a three-part dated correction block (a/b/c,
re-attributing which facts sit inside their floors). That is the prescribed way to correct an
append-only ledger, their content is at HEAD intact, and **nothing was lost** — my pathspec
commit simply took the file as the working tree held it.

**And the reverse happened the same hour:** my `WRONG_CALLS.md` entry was committed by
*another* session, in `1bdcc929a` ("WRONG_CALLS: I predicted 17 failures from a page ROLE…"),
under their message. Also fine, also nothing lost.

Both are the documented shared-tree behaviour — a pathspec commit protects you from another
session's *staged index*, never from a *same-file* edit in the working tree.

⚠ **And chasing it walked me into `HEAD~1`.** My first three attempts to read the removal
diffed `HEAD~1..HEAD` — and between my commit and my investigation another session committed
`e273fcb2f`, so `HEAD~1` had become THEIR commit and the diff came back **empty**. An empty
diff reads exactly like "there was no removal after all", which is the comfortable answer and
was false. Pin the sha: `git diff b0cf6e501^ b0cf6e501 -- <file>`. This is
[[relative-git-refs-are-not-evidence]] firing in the ten minutes after a commit, on a tree
where HEAD moves every few minutes.

### The diagnosis loop returned UNVERIFIABLE — and the two things it wanted, I hold

Run `fbdaca97-a97e-41e6-b422-2475521e6a6c` stopped at `scope-not-narrowing` with status
**UNVERIFIABLE** — **not REFUTED**. It agreed the numeric signature was present ("1 projected
section against a 3-entry `pages.sections` plan → 33%") but would not conclude, naming exactly
two gaps. Both are closed first-hand below, so this claim rests on verification I did myself
and say so plainly (owner ruling 2026-07-31), not on a loop verdict it did not give.

**Gap 1 — "nothing ties THIS page's tool-recreation route and THIS sections write to a single
`apply_adoption_plan` transaction."** The work item for page
`f763ca0e-a5ad-4d25-9e6c-37d158c13493` [MEASURED 2026-08-25]:

```
item_type    needs_tool_recreation      handler_agent  tool-recreation-handler
source       adoption                   created_by     site-adoption-agent
batch_id     04cb939f-3afe-4083-942e-445dfabbb4ee
item_key     needs_page:tool-example    spec.mode      recreate
```

`item_key` is `fmt.Sprintf("needs_page:%s", page.Name)` — the literal at
`apply_adoption_plan_action.go:751` — and `spec.mode='recreate'` plus a populated
`interactive_features` is the `pageSpec` built at :710-724. The page row itself is written by
`applyAdoptionPlanPagesUpsertSQL`, the same action. Same action, same batch, same run.

**And the sharpest single piece of evidence is inside that spec.** The adoption's own analysis
recorded this page as:

```json
{"name":"Interactive Job Prep Checklist","type":"tool","self_contained": true, …}
```

**`self_contained: true`** — and the very same action then wrote that page a **three-entry**
`pages.sections` plan. The contradiction is not an inference across two subsystems; it is two
fields written by one action in one transaction, and one of them is the floor's denominator.

**Gap 2 — "the actual body of `measurePageSectionCompleteness` showing the denominator is read
from `pages.sections`."** Read directly (`save_sections_prune_floor.go`):

```go
CASE WHEN jsonb_typeof(p.sections) = 'array'
     THEN jsonb_array_length(p.sections) ELSE 0 END,   -- ...FROM pages p WHERE p.id = $1
m.Planned = planned - suppressed - m.LockedRows
{Label: "planned sections", Confirmed: projected, Stored: m.Planned},
```

The denominator is `pages.sections`, less suppressed, less locked. Confirmed.

⚠ **Read the loop's stop correctly.** UNVERIFIABLE is not a refutation and it is not a pass —
it is the loop declining to conclude on the bundle it could assemble, and its "still needed"
list is the useful output. It could not fetch a function body its own symbol scope had already
named (`measurePageSectionCompleteness` was in the scope list), which is worth knowing about
the instrument: **being in scope is not being retrieved.**

## 2026-08-25 12:24Z — PHASE 2 HAS FIRED IN PRODUCTION, twice

After the plan correction (`OPERATION_2026-08-25_correct_cv1_tool_page_plans.sql`, owner
decision), both recreations were re-queued at priority 5 and claimed within a minute. Both
saves cleared the floor at 1 of 1 and **both adopted**.

```
page                    slot_name            component         regenerable  stamped  bytes
cv1.co.uk/index         hero                 adopted-fragment  t            t        17,595
cv1.co.uk/tool-example  generic-text-block   adopted-fragment  t            t        20,076
```

**`cv1.co.uk/index` is the whole of `bugs_open/357` in one row, and it is now correct.** A
17,595-byte self-contained interactive tool, sitting in a slot named **`hero`** — the exact
configuration the bug was filed about — and the row says `adopted-fragment` with
`content_data.body` reproducing `rendered_html` byte for byte, carrying a real provenance
stamp. Before phase 2 this row would have declared itself the shared `hero` component and
joined the population.

`slot_name` untouched on both, `position` 1, `rendered_html_digest = md5(rendered_html)`
(the digest is honest), `build_status = deployed`.

**Both point at the SAME `component_versions` row** (`3301ef65-4d83-4ea5-aa7c-65cb38e83653`,
template `{{.body}}`) — the "1.00 version rows per component, not a log" property holding
across two independent adoptions rather than being asserted from one.

### The three STOP conditions, checked rather than assumed

| condition | reading |
|---|---|
| a 357 population row acquires a stamp | **0** — splice hygiene holds, as it has since 08-25 09:08 |
| a page's row count goes UP by one | `cv1/index` = 1 row (was 0; the page was empty, so this is a birth, not the carry-forward landmine) |
| a new fragment lands with `component_id` NULL | **1 — and it is NOT ours, and it is NOT a phase-2 failure** |

The NULL row is `loanandmortgagecalculator.co.uk/tool-overpayment-priority`, created
**10:58:19Z — thirty-one minutes BEFORE my first dispatch**, on another lane's site. It has
no `<section>` and no `data-component`, which looks like an adoption candidate until you
count the rows: **5 on that page.** The fallback arm emits exactly ONE section, so a 5-row
save came through the metadata path, `FallbackAdopted` was never set, and adoption was never
offered the row. It is the ordinary "no component matched this slot name" outcome, which
predates phase 2 entirely. ⚠ **Worth stating because the surface reading — no section, no
data-component, no component_id — is exactly what a failed adoption would look like.** The
row count is what discriminates.

**The population is still 22** and neither adopted row is in it, which is the point: they are
bound to `adopted-fragment`, not to `hero`.

### What this settles, and what it does not

**Settled:** phase 2's central claim is no longer unproven in production. The mechanism binds
a fragment to a component that provably reproduces its bytes, earns a genuine stamp through
the existing resolver, and leaves the slot axis alone. 578's precondition 2 is met on its own
terms, with no check weakened.

**Not settled by this alone:** whether an adopted row SURVIVES a rebuild — 578's precondition
4, which the file states in prose and does not enforce. Canary fired 12:29Z on `index` only
(correlation `e0c2d505-9875-4347-a718-a852f32ec6b7`), with **`tool-example` deliberately left
untouched as a control**, so a change can be attributed to the rebuild rather than to time.
Baseline pinned in `scratchpad/canary_before.txt`: `index` position 1, slot `hero`, md5
`26f484f2744ab3e9cd19e50f600a52b8`, component `9d4b922b`, version `3301ef65`, 17,595 bytes.

### NEAR-MISS: I almost read a PRE-SAVE state as "the adopted row survived a rebuild"

At 12:57 the canary's numbers looked like a clean pass on every axis — `index` still 1 row,
slot `hero`, md5 `26f484f2…` unchanged, component still `adopted-fragment`, still stamped,
17,595 bytes. Against the pinned baseline that is a perfect match, and it would have gone
into a summary as *"precondition 4 satisfied."*

**It was the state BEFORE the save had run.** The rebuild's own record said so:
`collected_data->'save_sections'` was **null**, and no save step appeared among its keys —
only `plan_sections`, `write_page_content` and the review/link steps.

Reading `page-rebuild`'s sub-workflow settles the order, and it is not the intuitive one:

```
plan_sections -> write_page_content -> review_page_content -> check_review_approved
   -> assemble_page -> save_sections -> update_page_status -> deploy_page
```

**`save_sections` runs AFTER `assemble_page`**, and the orchestration was sitting at
`assemble_page`. So every "unchanged" reading I had was taken before the only step that could
have changed anything.

**The shape, and it is this lane's own recurring one:** an instrument that reads exactly the
same whether the mechanism is working or has not yet been asked the question. Nothing about
the numbers themselves discriminates — 1 row, right md5, right component is what a passing
canary looks like AND what a canary that has not started looks like. The demand control is
`save_result IS NOT NULL`, i.e. did the step that matters actually execute, and it costs one
query.

**The check that generalises:** before comparing an after-state to a baseline, prove the
operation under test RAN. For any step-based rebuild that means the step's own `output_field`
(here `save_result`) being present in `collected_data` — not the orchestration being alive,
not the page being flagged, and not elapsed time. Nine of this lane's twelve `WRONG_CALLS`
entries are versions of this, and I walked into it while writing up the one that fired.

### Canary #1 died to the chassis roll, and the state survived it

`e0c2d505-9875-4347-a718-a852f32ec6b7` ended **FAILED** with
`"reaper: stale EXECUTING_STEP for >4h; step=build_pages_loop_iter_0_assemble_page"`. The
chassis rolled underneath it (`agent-chassis-67fd9c76f5-*` → `669b45fdb4-*`, pods started
19:07Z). `save_result` was never set, so **the canary tested nothing** — it is not evidence
about the carry, in either direction.

Re-measured **19:29Z, after the roll**: `adopted_rows=2 · population=22 ·
population_stamped=0 · armed_carriers=6`. **The roll changed none of it**, and the reason is
worth stating rather than assuming: arming lives in `agent_definitions`, so it is DB config,
live immediately and untouched by an image change. A roll can kill work in flight; it cannot
disarm phase 2.

Canary #2 fired 19:30Z on the fresh build: `5a0cad41-fe0c-4636-9b2d-9c942486019c`, `index`
only, `tool-example` still held as the untouched control. Pods were 22 minutes old, past the
~300s post-restart window in which a spawn is silently dropped.

### CORRECTION + the likely reason both canaries stalled: there is a `git_commit` between assemble and save

I wrote in the handoff and above that the order is
`assemble_page -> save_sections -> update_page_status -> deploy_page`. **That was inferred,
not traced**, from the orchestration sitting at `assemble_page` with `save_result` absent —
which is consistent with the true order and does not establish it. Traced from each step's own
`next_step` in `page-rebuild`'s `build_pages_loop` sub-workflow:

```
plan_sections -> write_page_content -> review_page_content -> check_review_approved
   -> assemble_page -> deploy_page (action: git_commit) -> save_sections
   -> update_page_status -> complete_page
```

**`deploy_page` is a `git_commit`, and it sits BETWEEN assemble and save.**

The load-bearing claim survives — the save runs after assemble, so `collected_data ?
'save_result'` is still the only thing separating a passing canary from one that has not
started. But the correction matters for a different reason: **both canaries stalled at
`assemble_page`, and the step immediately after it commits to git.** Canary #1 sat there until
the >4h reaper took it; canary #2 sat there 8+ minutes with no update. A git operation is a
far more plausible stall than page assembly, and it is not something you would look at while
believing the save came next.

**[UNVERIFIED]** that the git step is the cause — the evidence is two stalls at the same step
and the identity of the step that follows it, not an observation of git blocking. Naming it as
a hypothesis so the next session checks it rather than inheriting it as fact. The cheap check
is whether OTHER sites' `page-rebuild` runs are also stalling at `assemble_page` right now: if
they are, this is fleet-wide plumbing and nothing to do with 357 or with adopted rows.

**The discriminating check, run rather than deferred — and it CANNOT discriminate yet.**
[MEASURED 2026-08-25 19:40Z] `orchestration_states` holds **2** `page-rebuild` runs in total:
`0` reached `save_result`, `2` ended at `assemble_page`, `1` FAILED (the reaped one).

**Both are mine, on the same page.** So the zero has a denominator of two and says nothing
about `page-rebuild` in general — I nearly wrote "no page-rebuild run has ever reached the
save", which is true of the window and worthless as a claim. The table is a rolling window and
the fleet simply is not running page rebuilds today.

**What would discriminate, and it is cheap:** flag a NON-adopted cv1 page (`request-index` or
`how-it-works-index`) as `needs_rebuild` and fire the same rebuild. If it also stalls at
`assemble_page`, the stall is the path (or the git step after it) and has nothing to do with
adopted rows or with 357. If it sails through to `save_result`, the stall is specific to the
page carrying a 17.5KB adopted fragment — which WOULD be a 357 finding and would matter to
phase 3. **Not run: canary #2 is still EXECUTING and a third concurrent rebuild on the same
site would muddy both.** Named as the next session's first control.

### Canary #2 failed identically — and the control is now running

`5a0cad41-fe0c-4636-9b2d-9c942486019c`: **FAILED**, `save_ran = f`, same reaper message, same
step — `"reaper: stale EXECUTING_STEP for >4h; step=build_pages_loop_iter_0_assemble_page"`.

**Two rebuilds, two identical deaths, and the second one was on a FRESH chassis** (pods
`669b45fdb4-*`, 22 minutes old at dispatch, past the ~300s window). So the first failure was
not "the roll killed it" after all — the roll explains the *timing* of canary #1's reaping,
not the stall itself. Both stalled at `assemble_page`, whose `next_step` is a `git_commit`.

**Precondition 4 remains UNTESTED.** Not failed — untested. Neither run reached
`save_page_sections`, so neither says anything about whether an adopted row survives a
rebuild. The adopted rows are byte-identical throughout (`26f484f2…` / `291b88d8…`, both still
`adopted-fragment`, both still stamped) — because nothing touched them, which is not evidence.

**The control, fired 2026-08-26 on `request-index`** (correlation
`8d002375-1524-4abd-b04c-91a2e6a74277`): 2 rows, components `hero` and `contact-form`, **no
adopted fragment**. `index`'s leftover `needs_rebuild` flag was cleared back to `deployed`
first, so the rebuild targets the control page alone and the result is attributable.

- control also stalls at `assemble_page` → the stall belongs to the rebuild path or the git
  step after it. **Nothing to do with 357**, and precondition 4 needs a different vehicle
  (e.g. drive the page through `page-rerender`, or wait for a natural rebuild).
- control reaches `save_result` → the stall is specific to a page carrying a ~17.5KB adopted
  fragment. **That IS a 357 finding** and it bears directly on whether phase 3 is safe: a
  re-typed row that cannot be rebuilt afterwards is a worse state than the mislabelling.

### The control failed for a THIRD reason, and it is fleet-wide: the Anthropic credit balance is exhausted

`8d002375-1524-4abd-b04c-91a2e6a74277` FAILED at `build_pages_loop_iter_0_write_page_content`
— **before** it ever reached `assemble_page`, so **it did not discriminate.** The error:

```
AI endpoint unavailable: provider=anthropic model=claude-sonnet-5
status 400: {"type":"invalid_request_error","message":"Your credit balance is too low to
access the Anthropic API. Please go to Plans & Billing to upgrade or purchase credits."}
```

[MEASURED 2026-08-26 00:0xZ] **22 orchestrations across 10 distinct agent types** carry this
error, first at **2026-08-25 23:46:18Z**, latest **00:06:01Z**: `page-content-writer` (9),
`tool-improver` (3), `reader-experience-auditor` (2), `content-quality-auditor` (2),
`site-review-agent`, `brief-fidelity-auditor`, `visual-design-auditor`, `generic`. Anything on
the estate that calls an LLM is failing. Owner notified.

⚠ **DO NOT CONFLATE THIS WITH THE CANARY STALLS — the timeline refutes it.** The stalls began
**12:57Z** (canary #1) and **~19:41Z** (canary #2); the credit failures begin **23:46Z**, six
to eleven hours later. The stalls are NOT explained by credit exhaustion, and a session
arriving tomorrow will find both symptoms live at once and be tempted to collapse them into
one cause. They are three separate things:

1. **the prune-floor contradiction** (§ above) — diagnosed, real, estate-wide, unfixed;
2. **`page-rebuild` stalling at `assemble_page`** — twice, on both a stale and a fresh
   chassis, cause **[UNVERIFIED]**, and the step after it is a `git_commit`;
3. **the credit exhaustion** — from 23:46Z, fleet-wide, an account matter, not a code one.

**So precondition 4 is still untested and cannot be tested while (3) holds**, because every
route to it runs a content writer first. When credit is restored, re-run the control
(`request-index`, non-adopted) before re-running the canary — the control is what says whether
(2) is a 357 problem at all.

⚠ When checking the account: MEMORY [[the-fleet-key-is-not-on-the-default-console-org]] —
**capped while billing reads 0% used means the WRONG ACCOUNT is being looked at**; check the
keys' `Last used`. And per the owner's 2026-08-23 ruling, **never read a key into a session** —
probe from the pod.

## 2026-08-26 08:17Z — state check, and a correction I made within two minutes of making the error

**MISSTEP: I read `llm_call_log` row counts as evidence the fleet was partly working.** Seeing
602 credit failures but "35 LLM calls in the last 30 minutes, latest seconds ago", I inferred
some provider or key must still be serving. **Wrong.** `llm_call_log` has a `success boolean
NOT NULL DEFAULT true` column and logs the ATTEMPT, not the outcome:

```
success | calls | with_output | latest
   f    |  126  |      0      | 2026-08-26 08:17:25Z     <- ALL of them, both models
```

**Zero successful calls in two hours.** Not one. The fleet is completely down for LLM work,
worse than last night rather than recovering (63 credit failures in the last hour alone).

The shape is this estate's own [[a-receipt-nobody-asserts-on-is-a-log-line]]: a row in a table
named `*_call_log` reads as "a call happened", and the column that says whether it *worked* is
one I had to go and look for. **The discriminating query is `GROUP BY success` with
`count(*) FILTER (WHERE output_tokens > 0)` beside it** — a row count alone cannot tell a
working fleet from a flat-lining one, and it reported the more comfortable of the two.

### Lane state, re-measured 2026-08-26 08:17Z

`adopted=2 · population=22 · population_stamped=0 · armed_carriers=6`

Chassis pods `agent-chassis-6dd68888dc-*`, started 2026-08-25 23:11:52Z (a newer build than
the `669b45fdb4` that ran the adoptions). **The two adopted rows and the arming both survived
a second roll** — again because arming is `agent_definitions` config, not code.

**Everything blocking this lane now needs an LLM.** Precondition 4 requires a rebuild; every
rebuild runs a content writer first. So the control cannot be re-run, the canary cannot be
re-run, and 578 must not run in a window where no page on the estate can be rebuilt to verify
it afterwards.

### The fresh build is `v1.0.1341`, it IS rolled, and phase 2 is in it

Deployment spec and BOTH running pods read `docker.io/aqls/agent-chassis:v1.0.1341`
(`agent-chassis-6dd68888dc-*`, started 2026-08-25 23:11:5xZ). Spec and pods agree, so this is
not the same-tag-rebuild trap where a deploy ships the node's cached binary.

Capability probed at the running binary with both control arms (never `strings`, never a
discovery grep):

```
PRESENT: adopt fragment: bound an unidentified fragment   <- the phase-2 capability
PRESENT: self-contained tool section                      <- positive control
ABSENT : zzz_never_shipped_literal_357_qqq                <- negative control
```

All three correct, so the probe discriminates and phase 2 is genuinely in `v1.0.1341`. The
adoptions themselves ran on the previous build; the two adopted rows and the six armed
carriers have now survived **two** rolls.

### `bugs_open/406` filed — and its blast radius corrected within the hour

Filed the prune-floor contradiction as **`bugs_open/406`** (LLM-free work, so it could proceed
while the fleet is down): full evidence, the closed-form arithmetic, four fix candidates
ordered by what makes the bad state unrepresentable, verification queries, and the declared
090-substitution. Pattern into **016b §9**, row into **016b §10**.

**Then I corrected my own headline.** I had quoted the parked `save_refused_incomplete` queue
— 34 items, 16 domains — as 406's blast radius, to the owner twice and into two files, without
classifying a single item. Classified by the `(N of M)` in each item's own reason string:

| shape | items | domains |
|---|---|---|
| the 406 shape, `1 of ≥3` | **6** | 5 |
| no cohort captured (older reason format) | 26 | 15 |
| other shrinkage (`2 of 5`, `7 of 20`) | 2 | 2 |

**Two of the six are cv1's, which this lane created** — so four pre-existing victims, against a
published thirty-two. A queue is not a cause: those items share a *symptom*, and the 26 are
unattributed rather than attributed elsewhere. Corrected in `bugs_open/406`, in the 016b index
row and to the owner; logged in `WRONG_CALLS.md`. What survives unchanged: the arithmetic, the
21-of-22 cross-check with 357, and that nobody reads that queue.

## 2026-08-26 09:0xZ — credit restored, and the CONTROL PASSED

**Credit is genuinely back, checked at the instrument rather than taken on trust.** Last failed
call **08:57:45Z**; every call from **08:58:28Z** onward succeeded — 7 calls, 7 with output,
6,317 output tokens. `success` and `output_tokens > 0`, not a row count.

### The control passed — the rebuild path is NOT broken

`1adaac43-58f0-41bc-b368-44274d54ca58`, `request-index` only (2 rows, components `hero` +
`contact-form`, **no adopted fragment**), with the two adopted pages cleared to `deployed`
first so the result would be attributable. Result: **`COMPLETED`, `save_ran = t`** — it went
through `assemble_page`, through the `git_commit`, through `save_page_sections` and out the
other side. `request-index` came back with its 2 rows correctly typed; both adopted pages
untouched.

**So the earlier stalls were NOT "page-rebuild is broken".** That reading is dead. What remains
open is which of two explanations holds:

- **(a) the adopted page specifically** — something about a single 17.5KB `adopted-fragment`
  row makes `assemble_page` hang;
- **(b) transient** — both canaries happened to run in a window where something unrelated was
  wedged, and it has since cleared.

⚠ **The timeline does NOT settle it, and it is worth saying why, because the tempting inference
is wrong.** Canary #1 stalled 12:57Z and canary #2 ~19:41Z on 08-25; the credit failures began
**23:46Z**, hours LATER. So the stalls cannot be blamed on the outage — but the control ran
after credit was restored, so "the outage" and "the fix" are confounded with "before/after" in
a way one control cannot separate. **Canary #3 on `index` is the discriminator**: same
conditions as the control, same hour, differing only in whether the page carries an adopted
row.

Canary #3: `bf29ec85-8ef9-457a-9366-1ca121a95810`, `index` only. Baseline re-pinned live in
`scratchpad/canary3_before.txt` (the pages had grown since the first pin — always re-pin):

```
index  pos 1  slot hero  md5 26f484f2744ab3e9cd19e50f600a52b8  component 9d4b922b  version 3301ef65  17,595 B
```

**Pass = `save_result` present AND rows still 1 AND md5 still `26f484f2` AND still
`adopted-fragment`.** Rows going to 2 is the carry-forward landmine: STOP, do not run 578.

## 2026-08-26 09:35Z — PRECONDITION 4 HAS FAILED, and 578 MUST NOT RUN

Canary #3 (`bf29ec85-8ef9-457a-9366-1ca121a95810`) stalled at `assemble_page` like the other
two — **while the control passed in the same hour**. Three stalls on the adopted page, one
clean pass on the non-adopted one, same conditions. The page was the variable, and chasing it
found the mechanism end to end.

### The chain, all of it measured

1. **`plan_sections` planned the page's one section onto `adopted-fragment`.** From the run's
   own `section_plan_0`: `ready_count: 1`, `ready_names: ["hero"]`, `function:
   "adopted-fragment"`, `component_id: 9d4b922b`. **The component's own seeded description
   says: *"Not for authoring: nothing should ever plan a page section onto this component."***
   Nothing enforces that sentence, and the planner selected it.
2. **The content writer then returned nothing usable:** `skipped: true`,
   `reason: "no sections defined for page"`, `section_count: 0`, `page_body` length **0** —
   despite the plan reporting one ready section. So
   `page_content_0.response.page_html` never existed.
3. **`assemble_page` went looking for it and never came back.** `extractFieldValue`
   (`multipage_actions.go:1185`) has two fallbacks that are exact inverses — one strips
   `.response.` and recurses, the other adds it back and recurses — with no depth bound. The
   path ping-ponged until the goroutine stack passed 1 GB:
   `fatal error: stack overflow`, container `exitCode: 2`. **12,654 of 12,654 log lines in the
   dead container are the same warning.** Filed as **`bugs_open/408`**.

### What this means for phase 3, and it is decisive

**An adopted row does not survive a rebuild. The rebuild crashes the pod.** That is not
"precondition 4 untested" any more — it is **precondition 4 FAILED**, on its own terms
("bytes identical, component still adopted-fragment, row count unchanged" was never reachable
because the save never runs).

**578 would bind 22 more live pages to `adopted-fragment`.** On present evidence every one of
them would then wedge its next rebuild and crash the agent. **Re-typing them now would convert
22 mislabelled-but-rebuildable pages into 22 unrebuildable ones — strictly worse than the
defect 357 describes.** Do not run it. This is exactly the outcome precondition 4 exists to
prevent, and it earned its place today.

### The honest boundary of the claim

⚠ **What is proven is the CHAIN, not that `adopted-fragment` is the only way to start it.**
Step 3 (`bugs_open/408`) is fully general — any unresolvable `content_field` crashes the pod.
Step 2's skip is what produced the missing field here, and step 1 is what produced step 2. So:

- **[MEASURED]** planning a section onto `adopted-fragment` → writer skips → assemble crashes.
- **[UNVERIFIED]** whether a page bound to `adopted-fragment` can EVER be rebuilt — i.e.
  whether the writer's skip is caused by the component's `category: internal` /
  "not for authoring" nature, or by something else about this page. The discriminating check
  is to read why the writer counted 0 sections from a plan reporting 1 ready.

**That distinction decides whether phase 3 needs a fix to phase 2 as well.** If the writer will
always skip an adopted-fragment section, then adoption creates pages that can never be rebuilt,
and that is a defect in phase 2 itself — not merely in 408 — and it must be fixed before ANY
repair re-types rows onto that component.

### Cross-lane: analytics_gtm is tagging cv1 — rerender is safe, rebuild is not

The `analytics_gtm` lane (`bugs_open/397`) is applying GTM-PQ3WCTBD to cv1.co.uk, which files
one `stale_chrome → needs_rerender` and re-renders chrome plus the four pages. **Answered: go
ahead**, and the answer is measured rather than reasoned.

`page_rerender` has already completed on cv1's adopted pages **six times** [MEASURED
2026-08-26 from `site_work_items`] — `index` and `tool-example` at 12:03Z and again at 13:47Z
on 08-25, plus the other two pages. **So the rerender path handles an `adopted-fragment` page
fine.** That is a genuinely useful fact for this lane too: it means the crash is specific to
the *rebuild* path, not to adopted rows in general.

⚠ **What I warned them off: anything that sets `build_status='needs_rebuild'` on cv1.** That is
the path that crashes the pod (`bugs_open/408`). I have cleared the flag I set on `index`.

**I did NOT answer their "is cv1 a throwaway that should not report into GA4?" question.** It
is a portfolio decision and the owner's to make — I only used the site as a vehicle to prove a
mechanism. Told them to put it to him and flagged it in my own reply rather than writing a
guess into `docs024_key_docs_latest/analytics_gtm/`.

Also noted for whoever picks this up: two `page_rerender` items (index, tool-example) have sat
`triaged` since **07:27Z today**, stranded by the credit outage (23:46Z → 08:58Z). Credit is
back; they should drain.

### The regeneration question, part-answered — and it WEAKENS my own framing above

I said the next session's first question was whether an `adopted-fragment` section can ever be
regenerated. Chasing it found the canary run's content review, attributed by correlation
(`bf29ec85-…`) rather than by time:

```json
{"approved": false, "overall_score": 0.1,
 "issues": [{"section": "company_brief", "severity": "error",
             "issue": "Company name is missing - field contains '<no value>'"}]}
```

**`<no value>` is Go `text/template`'s output for a missing key**, so the rebuild's content
really did render against absent data and was rejected outright (score 0.1). That is the
missing link between "plan says 1 ready section" and "compile counted 0".

⚠ **But the failing section it names is `company_brief`, NOT the adopted fragment** — and that
matters, because it opens a second explanation I had not been giving weight to:

- **(a) an `adopted-fragment` section cannot be regenerated** — `{{.body}}` has no `body` on a
  fresh write, renders empty, and `extractSectionFromMap` drops any section whose html is `""`;
- **(b) cv1's own site specs are incomplete**, so content generation produced `<no value>`
  placeholders and the reviewer rejected the page. cv1 was adopted yesterday and **its build
  cascade never finished** — `needs_briefing` was still queued when I fired the canaries, so
  the identity/brief specs the writer reads may genuinely be thin.

**The control is what still favours (a), and it is worth stating explicitly:**
`request-index` rebuilt successfully **on the same site, with the same specs, in the same
hour**. If (b) were the whole story, the control should have failed too. So incomplete specs
cannot be sufficient — but the `company_brief` error shows they are not innocent either, and
the two may compound.

**[UNVERIFIED] which of (a) or (b) dominates.** What would settle it, cheaply and without
touching a live population page: rebuild `how-it-works-index` (cv1, `planned`, **0 rows**, 3
planned sections, **no adopted fragment**). Same thin specs, no adopted row.

- it fails the same way → the specs are the cause, `adopted-fragment` is exonerated, and
  **578's risk drops sharply**;
- it rebuilds cleanly → adopted-fragment is the discriminator, and **578 must not run** until
  phase 2 can regenerate its own rows.

⚠ **Do not run that test by flagging an ADOPTED page** — that is the crash path
(`bugs_open/408`), and it costs a pod and a 4-hour wedged orchestration each time.

**Correction to my own handoff banner:** it says the chain is measured and only the `{{.body}}`
step inferred. That is still true as far as it goes, but the banner implies adopted-fragment is
the established cause, and on this evidence it is the better-supported of two live candidates,
not a settled one. The 578 recommendation is unchanged either way — **do not run it while the
cause is open**, because the downside is 22 unrebuildable live pages.

### The discriminator was INCONCLUSIVE — three rebuilds, three different failure points

`how-it-works-index` (0 rows, 3 planned sections, no adopted fragment, same site and specs)
**FAILED at `write_page_content`, before reaching the step in question**:

```
step process_sections_loop_iter_0_render_section failed: render_component:
component "mechanism-flow": content does not match the declared field type(s) —
steps[0].branches: declared array (items: object), got string; steps[1]…
```

An LLM schema mismatch on an unrelated component. **It does not answer (a) vs (b)** and I am
not going to read it as if it does. Flag cleared back to `planned`.

**Three rebuilds on one site, three different failure points** — and that is itself the useful
observation:

| page | rows | failure |
|---|---|---|
| `request-index` | 2, normal components | **none — COMPLETED, `save_ran = t`** |
| `index` | 1, `adopted-fragment` | compile found 0 sections → no `page_html` → `assemble_page` stack overflow |
| `how-it-works-index` | 0, normal components | `render_component` type mismatch on `mechanism-flow` |

**What this DOES support, at the strength the evidence allows.** Candidate (b) — "cv1's specs
are too thin to generate content" — is now **weak**: `request-index` generated, rendered,
reviewed, assembled and saved cleanly, and `how-it-works-index` got far enough to generate
content and attempt a render (its failure is a schema mismatch, i.e. content EXISTED). Thin
specs do not stop this site building. That leaves **(a) — the `adopted-fragment` binding — as
the better-supported explanation**, by elimination rather than by direct observation.

⚠ **[INFERRED], and it is inference from three runs with three outcomes, not a controlled
result.** I have not observed an `adopted-fragment` section rendering empty. The direct
evidence would be the per-section render output for the adopted page, and the writer
orchestration that held it has been reaped.

**What I would do next, and why I am not doing it now:** re-run the canary on `index` and
capture the page-content-writer's per-section output LIVE (the child orchestration, by
correlation, before it is reaped) — that shows the adopted section's rendered html directly and
settles it. It costs one more pod crash and a 4-hour wedged orchestration
(`bugs_open/408`), which is a real price to pay for a confirmation, and the 578 recommendation
does not change either way: **do not run it while the cause is open.**

### 2026-08-26 evening — review pass: the canary's pass condition is VACUOUS once 408 is fixed, and cv1 cannot test precondition 4 at all

Context: owner asked for a second-opinion review of the lane state before carrying on. Checks
re-run first-hand at HEAD and the live DB. Two state changes, three findings.

**State change 1 — the credit blocker CLEARED ~09:00Z and held all day.** `llm_call_log`
grouped by success: 200–300 successful calls/hr WITH output from 09:00Z onward (read at
~15:00Z and again ~21:00Z). The morning banner's "ZERO working LLM calls" is stale.

**State change 2 — v1.0.1345 rolled ~20:36Z (pods `agent-chassis-5864bf97c5-*`, both agree on
the tag). It does NOT carry a 408 fix** — not by stamp-reading but structurally: no commit in
this repo's history touches `multipage_actions.go` after `c4baa53e7` (328, pre-408), and the
recursion is present at HEAD (`:1213`, `:1223`) — so every buildable commit carries the
defect, whichever one 1345 was cut from. **Canary and 578 remain forbidden.**

**Finding 1 — post-408-fix, the morning handoff's canary pass condition passes VACUOUSLY.**
Traced at HEAD:
1. writer skips (0 sections) → `assemble_page` gets `""` from a fixed `extractFieldValue` →
   returns the skip shape (`multipage_actions.go:108-120`) — no crash;
2. `git_commit` skips via `checkUpstreamSkipped` (`git_deployer_actions.go:673`, keys on
   `assembled_page.skipped`);
3. `save_sections` RUNS — its early exit is keyed to the OWNED-page marker only, deliberately
   (`save_page_sections_action.go:71-90`) — and exits at `len(sections)==0` with
   `success:true, skipped:true` (`:344`/`:401`), BEFORE the DELETE-and-reinsert and BEFORE the
   Layer 2 carry-forward (`:555-600`);
4. the run reaches COMPLETED.

So `save_result` present + rows still 1 + md5 unchanged + still `adopted-fragment` — every
clause of the stated pass condition — while the conservation machinery never executed. The row
"survives" because nothing touched it.

**Finding 2 — cv1 is structurally the WRONG vehicle for precondition 4.** What 578 depends on
is the ARMED identity-carry inside a real Layer 2 splice: an incoming, normally-generated
section matching the adopted row's slot, and the stored bytes keeping their own identity
instead of the plan's (`save_page_sections_action.go:562-577` — "without this … the very next
rebuild re-mints `hero` over an adopted row and the population renews itself"). That path only
runs when the save carries >0 incoming sections for that slot. cv1's adopted pages can never
produce one: their `pages.sections` plans name `adopted-fragment` itself (index planned=1,
tool-example planned=1), so the writer always skips. [INFERRED at the code level that the
planner consumes `pages.sections`; measured at the behaviour level — it planned exactly the
plan's one section onto exactly the plan's component.]

**Finding 3 — the 22 population pages are the RIGHT shape, and one of them is the natural
pilot.** Measured 2026-08-26: all **22** plans name `hero` (`plan_names_hero = t` on every
row); **16** are `rebuild_policy='generic'`, **6** `'owned'` (the gamesdesign `tool-*` six —
matching 578's "six owned pages"). Post-578 a rebuild of a generic one generates a fresh hero
per its plan, the save carries it, Layer 2 splices (match key is the SLOT, which 578 preserves
— it RAISEs if `slot_name` moved), and the armed carry must keep `adopted-fragment`. That IS
precondition 4. Minimal pilot: retype ONE row by 578's own procedure scoped to one page —
`mortgagecalculator.co.uk/tool-simple` is the minimum (planned=1, generic, single row) — then
rebuild that page and assert the REAL pass condition: `save_result` present AND
`sections_saved > 0` AND rows still 1 AND bytes preserved AND component STILL
`adopted-fragment`, not re-minted `hero`.

**Open question the pilot must answer before anyone widens it — what does the SERVED page hold
afterwards?** The chain commits the assembled (prose-hero) page to the sites repo BEFORE the
save's preservation runs (the LANDMINES `assemble → git_commit → save` entry). This is a
property of TODAY's rebuild of any generic population page, not something 578 introduces — but
the pilot will exercise it, so read the served page after, not just the DB row. The 6 owned
pages are refused at assemble and are not exposed.

**Also measured:** `extractFieldValue` has exactly **1** caller as of 2026-08-26
(`multipage_actions.go:106`) — 408's candidate 1 cannot break another call site. And no 406
confound on cv1: index planned=1 → ratio 1.0, the prune floor cannot fire. Note the pilot does
not structurally require the 408 fix (its plan generates real content), but any no-content
outcome on the way still crashes the pod until 408 is fixed — fix 408 first regardless.

### 2026-09-02 — 408 FIXED IN CODE (`6e2d4a039`, Council-Submitted `3918db52`); the lane's blocker is one roll away

Resumed per the owner. Thread check: 408 owned by this lane, nobody else on it (who-owns +
git log + zero queued work items). Validity re-proven at HEAD before touching anything
(recursion at `:1214`/`:1224`; it had shifted one line under 420's commit).

**Cross-session coordination that worked:** `multipage_actions.go` was dirty with
`bugs_open/423`'s uncommitted `UpperFirst` hunks calling a helper that did NOT exist at HEAD —
committing my fix by pathspec would have swept them and broken HEAD's build. Messaged the 423
session; they committed `3edb30476` (definition + all 8 call sites together) within minutes and
confirmed. Passenger check before my commit: exactly one hunk in the file, mine.

**The fix** (plan prepared by a fable agent, reviewed and grounded first-hand): bounded
ordered-candidate loop over a new pure three-valued walk helper (`walkFieldPath`), matching
the family convention of the three sibling resolvers. The plan's key catch — the bug file's
flat 3-candidate sketch would have NARROWED resolution on exotic shapes; the builder
reproduces the old recursion's exact terminating tried-set (`WRONG_CALLS.md` 2026-09-02,
correction noted in the 26b handoff item 1 and bug §9). 15 tests under `go test -timeout`;
**mutation control run**: crash input vs the old function verbatim = FAIL by stack overflow in
3.7s. `verify-head-builds --with --test` green at `38db61b28`. Council submission
`3918db52-4d94-4b65-8065-6be4cdef42eb` (dry-run admission first), commit carries
`Council-Submitted:`; verdict to be read when it lands (~30 min queue).

**Diagnosis loop NOT run, stated per the 2026-07-31 ruling:** the root cause is directly
observed (crash captured at the pod, cycle read at source), not hypothesised — substitution
statement now in bug §9.

**What this does and does not unblock for 357:** once an image ≥ `6e2d4a039` rolls, the cv1
canary becomes a clean no-op instead of a pod crash — but per Finding 1 (26b handoff) a green
canary is VACUOUS for precondition 4. The honest precondition-4 vehicle remains the one-row
pilot on `mortgagecalculator.co.uk/tool-simple` (Finding 2), which needs the owner's nod.

**Separate track surfaced, not taken:** four `.response.`-fallback resolvers with three
different candidate orders + the ~10–14 near-clone walkers — recommendation is a
concept-register census entry first; consolidation is RFC-scope if ever needed. Recorded in
bug §9; no lane owns it.

### 2026-09-02 (later) — 408 APPROVED r1; all three advisories acted on; two cross-lane loops closed

Council: APPROVED round 1, corr `3918db52`, 9 reviewed / 7 abstained / none high. Advisories
all real and all closed in `b8bf40694` (see bug §10): the loud→quiet trade is now countable
(`ASSEMBLE_CONTENT_FIELD_UNRESOLVED` when content was expected), the walker census re-ran WITH
a positive control (20 bodies, one bounded recursion, class closed), the test carries its own
watchdog. Original commit `6e2d4a039` carries `Council-Submitted`; the follow-up carries
`Council-Reviewed` — 098 will credit both.

Cross-lane: webdesign-tool-rebuilds supplied the roll-verification instrument (bug §9);
webdesign-tool-rebuild's "affects us" turned out FALSE for their chain (2,124/2,124
page-rerender — the other assembly function) and my notification's framing seeded their
vacuous discriminator — my WRONG_CALLS entry today; their reply surfaced a genuine
pre-existing rerender-path skip population (36/1,146 in 7d) now recorded in §9/§10, and a
skipped-key-absent-on-normal-runs trap (test for 'true', never 'false').

408 remains OPEN: fixed, approved, NOT live. Owed: an image ≥ `b8bf40694` rolled, the
`paths_tried` capability probe PRESENT, then §6's end-to-end checks. Then move to
bugs_closed and clear the 357 lane's blocker note.

### 2026-09-02 (evening) — PRIOR ART OFFERED for phase 3: the lendzy lane's adoption shape (migration 693) — decision input for the pilot, recorded, not adopted

The lendzy lane found the NULL-id arm of this class (3 pages, component_id NULL, rerender
fatal for ever while the 08-02 artefact serves — the OPPOSITE failure profile: deployment
impossible, regeneration harmless) and measured the arms DISJOINT fleet-wide: the only active
pages with no component_id are their three; our 22 carry a WRONG id, no intersection.

**Their fix shape (693, council corr a1b691e8, r3 in flight): per-tool adoption** — a
content_components row per tool whose html_template IS the stored rendered_html byte-for-byte
(component_level='tool', CLC-020 naming), repoint the row, rerenders filed in-transaction.

**Why it bears on OUR impasse, precisely:** 578 retypes onto the SHARED adopted-fragment and
its rebuild-safety rests on the armed Layer 2 identity-carry — the exact machinery
precondition 4 exists to prove and cannot yet (Finding 2, 26b handoff). Template-equals-bytes
removes that dependency: regeneration reproduces the page BY CONSTRUCTION, no carry needed —
the estate's own fix-ranking rule (close the door / make the bad state unrepresentable)
favours that property. [INFERRED, unverified] it may also dissolve the writer-skip arm: a
placeholder-free template should render to its own bytes, where adopted-fragment's `{{.body}}`
renders empty (the cv1 crash trigger). OPEN QUESTION put to lendzy: does 693 also rewrite
`pages.sections` plans? Our 22 rebuild via the PLAN (all name `hero`), so without the plan
repointed the Layer 2 carry is still in play and the no-op property doesn't attach on the
rebuild path.

**Transferable gotchas they measured (for whoever runs any adoption on our 22):**
1. `toolTemplateValid` (plan_sections_action.go) must pass per body or
   `loadComponentSchemasByID` silently drops the component and rerender goes fatal —
   ⚠ their probe trap: a must-fail control under 100 chars is VACUOUS (short bodies admitted
   as stubs).
2. Hand-filed rerenders need `site_work_items.page_id` (first-class column) AND `created_by`
   (NOT NULL, no default) — the live producer sets both on all 7,481 of its rows.
3. Their Guard 2 (abort unless the row still looks exactly as censused) transfers to our
   wrong-id repoint.

**Status: recorded as decision input for the owner at the pilot decision point (26b handoff
item 3), alongside 578 as designed. Not a pivot — 578 is reviewed, backed up, and phase 2's
birth-path (shared adopted-fragment) is live and proven; switching the REPAIR shape is the
owner's call with both options on the table.** Their docs: `lendzy_co_uk/` NOTES (c)/(i),
693 header round-3 block. Correction sent to them: this lane has no pending 090 on 357/408
(first-hand substitution, stated in 408 §9); their `63d4d1a7` reference doesn't match any run
of ours.

### 2026-09-02 (later still) — lendzy ANSWERED: 693 does NOT rewrite plans; the transferable shape for our 22 is TWO repoints in one transaction

Their measurement: lendzy's pages carry `sections = []` (8 of the site's 9 tool pages), so
their rebuild path has no plan to regenerate from — repointing `page_components` alone closed
their chain. **Ours is the opposite on exactly that axis (all 22 plans name `hero`), so
adoption alone does NOT close our chain.** The transferable version, per lendzy:

- ONE transaction, TWO repoints per page: `page_components.component_id` → the per-tool
  bytes-template component, AND the `pages.sections` hero element → the same component's
  identity; both guarded on their exact censused current values, abort-on-drift (their
  Guard 2 shape).
- **[UNMEASURED, named by lendzy]** whether a plan EDIT interacts with the drift reconciler /
  `built_from_plan_version` comparisons — 693 leaves `built_from_plan_version` alone and lets
  the rerender stamp it via the COALESCE in `buildPageDeployStampQuery`; a plan edit on our
  side may behave differently. **Must be measured before any adoption-shaped pilot.**
- Writer-skip dissolution: consistent one level down (their adopted bodies carry ZERO `{{`;
  a binding-free template renders to its literal bytes) — but our writer path is unexercised
  by their measurement. Still [INFERRED] for us.

**The owner's decision is now fully specified:**
| | 578 as designed | adoption-transferred (693 shape) |
|---|---|---|
| mechanism | retype onto shared adopted-fragment | per-tool component, template = stored bytes |
| rebuild safety | armed Layer 2 carry — UNTESTED (precondition 4) | by construction — plan repointed, nothing regenerates hero |
| status | reviewed, backed up, ready | needs a new migration + council round |
| preconditions | 4 unmet (blocked on test vehicle) | toolTemplateValid ×22, drift-reconciler interaction UNMEASURED |
| crib | 578 + _ROLLBACK | lendzy_co_uk/ docs + sql_for_agents/693_*.sql |

### 2026-09-02 (night) — OPTION B RULED by the owner; measurement session complete; migration 700 drafting

Owner ruled **Option B**. Decision + reasons + all pre-design measurements recorded in PLAN
(2026-09-02 decision block) — drift reconciler safe (compares plan IDs, not sections), the
THIRD repoint leg found (`site_plan_sections.component_name` — `pages.sections` is a derived
copy sync overwrites), one function collision (tool-equity-release → fork per RFC_036 §9.3),
22/22 bodies binding-free AND 22/22 pass the real `toolTemplateValid` (both-way controls;
probe file written, run, deleted in one step). Full pinned census (pc_id/md5/bytes ×22) is in
the migration's guards; bodies exported to session scratch `bodies_22.json`.

**Site-owner notification duty identified:** mortgagecalculator.co.uk has an ADOPTION lane
owning these exact tools as product (dormant session; site UNLOCKED, improvement loops live —
their rebuild mid-apply would drift the census and our Guard 2 aborts, which is correct);
vetcomparison.uk lane is live in a peer session; gamesdesign.co.uk has NO lane in the
workstreams index (checked 2026-09-02). Notify at council-submission time with the concrete
file: vetcomparison by message, mortgagecalculator by CONTRIB file, gamesdesign noted unowned
in the submission.

Migration pair `700_retype_357_population_by_adoption_HOLD[.../_ROLLBACK].sql` drafting via
fable with the full dossier; review → place → council → owner-sequenced apply (pilot
tool-simple first). 408 fix approved and committed but NOT yet rolled — the roll is still
wanted before any rebuild-dependent verification.

### 2026-09-02 (later) — migration is **701**, not 700 (number collision); drafted, reviewed, SUBMITTED corr `df6c1b41`

> **CORRECTED 2026-09-02: every "700" in today's earlier PLAN/NOTES entries reads 701.** The
> 396 lane committed its own `700_park_provenance_covers_the_handler_repoint.sql` twenty
> minutes before my placement (`1f0cd8ae2`) — the known migration-number-collision landmine.
> Renumbered throughout (files, GUC `m701.scope`, abort/verify strings); both files re-parse
> clean (20 + 10 statements) after the rename.

The fable draft came back with TWO findings that corrected my own brief, both re-verified
first-hand: **tool-ttk-calculator has ZERO plan rows** (my "exactly one per page" was 21/22 —
WRONG_CALLS today; its repair is sections-only, guard pins the zero) and **a second unplaced
fork already sits on function tool-equity-release** (`befacff0`, disclosed to council as the
RFC_036 §11-addendum-2 tail). Also from the draft: vetcomparison names its component
`tool-vet-comparison` rather than claiming the fleet-wide function `index`; the self-caught
Guard-1 defect (first cut refused its own remainder run) is recorded in the DESIGN notes with
the reviewer walk-table. Census in the SQL mechanically diffed against my pinned baseline:
22/22 pc_ids + md5s exact. Submission `df6c1b41-b600-41d1-8f7e-3e96fe422b31`; scope=pilot is
the DEFAULT apply mode; the pilot rides the page-rerender path which does NOT contain 408
(measured), so it need not wait for the roll.

### 2026-09-02 (close of day) — 701 APPROVED r3; bugs_open/408 CLOSED with the crash input exercised in production; the pilot awaits the owner's hand

**701: APPROVED round 3** (corr `df6c1b41`, 2 advisories none-high, both dispositioned in the
DESIGN notes — tool-health scope is intentional with `toolTemplateValid` as the 012-class
guard; the 044 heuristic is a named pilot watch item; the gamesdesign sign-off line added to
the header). Three rounds total; round 2's report was initially misread mid-run (a
byte-identical round-1 copy surfaced before completion — corrected in-session before acting).

**408: CLOSED.** v1.0.1354 carries the fix (capability probe vs the proven baseline), and the
§6 end-to-end ran through the real pipeline: cv1 flagged, receipt-asserted dispatch (corr
`6e84a4e3`), iteration 0 hit the exact crash input and produced the fixed code's own skip
message, chain COMPLETED in minutes, pods restartCount 0, both adopted rows byte-identical
after; flags reset to `deployed`. File moved to `bugs_closed/` with the evidence (§11); 016b
§10 row updated; LANDMINES gains the defusal entry (the crash arm is dead; the
completed-but-skipped arm is the live trap now) and the verifier armed via
`landmines-verify-dispatch.sh`. Honestly stated in §11: the one-vs-12,654 log count was not
captured (ephemeral pod, empty stream) — the count claim rests on structure (a flood ends in
a dead pod; the pods lived).

**Ready for the owner:** the pilot apply —
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -v scope=pilot < docs/agent_docs/sql_for_agents/701_retype_357_population_by_adoption_HOLD.sql`
(defaults to pilot even without the flag; pre-apply `ls | grep '^701'` per the header; then
the header's post-conditions, then `scope=remainder`).

### 2026-09-02 ~19:10Z — PILOT GREEN on every clause; remainder unblocked, owner's hand next

Owner applied `scope=pilot` by hand ~19:03Z: guards passed against the full 22-row census,
COMMIT clean, population 22 → 21. Post-conditions, all MET:

- three legs verified repointed (row slot=cc.name aligned, `pages.sections`,
  `site_plan_sections`), md5 unchanged at apply;
- the queued rerender (`triaged` 19:03:53) went `complete` by 19:07 — a REAL render and
  deploy (commit `c23f3cf0`, file `tools/simple/index.html`, success true), NOT a skip — so
  the 044 empty-schema-deferral watch item did NOT fire;
- **rows still 1** (carry-forward landmine did not fire) and **md5 STILL
  `7873509b8087a15cc3b32120e746f9e5` after the rerender** — the by-construction property
  (template = bytes ⇒ regeneration is a no-op) proven in a live run, which is the thing
  Option B was chosen for;
- served artefact (browser Accept, past the 97s lag): HTTP 200, 5 interactive controls,
  5 script tags, **zero `data-component="hero"`**; invented-URL control 404 — the 200
  discriminates.

Remainder: `-v scope=remainder` (21 rows; its own guard re-verifies the pilot state first).
Per the header's sign-off design, the owner's hand applies it — explicitly covering
ownerless gamesdesign.
