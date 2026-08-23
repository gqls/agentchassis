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
