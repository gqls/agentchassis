# NOTES — bugfix 114 imagery wiring

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-22 — lane opened, bug re-validated, one mechanism filed to the loop

### Ownership check first

`scripts/who-owns.py 114` names three lanes that **cite** 114 —
`mortgagecalculator_couk_adoption`, `brochure_component_library`,
`bugfix_284_flag_only_items_promoted` — all contributors, none owning the fix. The only
in-flight `needs_diagnosis` item (`554ea3d2…`, status `diagnosing`) is the
OrphanPagesCheck lifecycle-axis bug, unrelated. `git log` on the bug file shows five
contributions 08-13 → 08-17 and no fix commit. Taking the fix.

### Validity

Re-measured everything the file asserts (queries in RUNBOOK). It is still valid and
bigger: 18 sites carry the poisoned default (was 10), 518/580 assets are entity-unlinked,
and only **23 of 94** content_hero assets fleet-wide are wired to a component.

The parked-queue half has changed shape since filing and the file does not say so: 40
`image_landed` items are now `complete` against 8 parked, where the file recorded 14
parked / 13 complete. The remaining 8 are pages with no built slots, not a queue nobody
drains — an emit-time guard (`flag_page_image_rebuild_action.go:147-160`) now skips
sectionless pages, and a revalidation pass stamped verdicts on 08-21.

### The finding that reframed the bug

`StoreAssetAction` writes `sites.content_data.<purpose>_url` on **every** asset store,
using `storage.BuildAssetPaths(purpose, ext)` — purpose-derived, so always
`/assets/images/hero.jpg`. The deployer commits under
`storage.DeployedAssetPath(assetKey, purpose)` — asset-key-derived. **Two derivation
rules for one artefact**, so every page-scoped hero generation re-stamps the site-wide
default with a path that may exist nowhere.

That explains something the bug file recorded as a mystery: fundamentallyai was repaired
on 07-29 and reads `/assets/images/hero.jpg` again today. It was not re-broken by a
person; the next hero store re-poisoned it. It also explains 10 sites → 18 sites.

And the workflow **already asks for this not to happen**: the imagery store step passes
`update_site_brand_assets: false`
(`docs/agent_docs/sql_for_agents/107_image_build_handler.sql:1163-1170`).
`grep -rn "update_site_brand_assets" --include="*.go" platform/ internal/` → **no hits**.
The key is dead. Somebody saw this coming and wrote a gate the action never read.

### The natural experiment, and what it rules out

Eight mcalc tool pages, one site, one day, one flow, all with an active
`content_hero_tool_*` asset. **Two wired, six on the fallback.** Full table in PLAN.

Ruled out by measurement:
- **race** — every asset was `active` 972–2650 s before its render, and `tool-affordability`
  rendered **2.2 days** later and still missed;
- **plan routes** — mcalc's current plan has no page-scope hero row for any tool page and
  no site-scope hero row at all (only `logo`), so routes 1 and 3 were empty for all eight;
- **asset shape** — all ten keys match the `ContentHeroKey` convention, purpose
  `content_hero`, status `active`; the Lane B predicate replays and matches **today**.

So Lane B works (23 wired components fleet-wide prove it) and something about *which
path each render took* differs. The 08-15 orchestration rows are purged.

**Filed to the diagnosis loop rather than guessed.** CLAUDE.md's rule is that a durable
mechanism claim goes through 090 before I assert it, and "it feels obvious" is exactly
the signal that is worthless here. Intake corr `23da0760-f2da-4095-967e-2bdd308aa7ea`,
run corr **`ea7dfeef-c11d-40c4-b24f-b8f42413b1ae`**.

### MISSTEPS

1. **Wasted a 090 intake on unescaped double quotes.** I wrote `assets["hero"]` in the
   symptom to name a map key. The trigger interpolates the symptom into a `$json$` SQL
   literal without escaping, so it aborted with
   `invalid input syntax for type json … Token "hero" is invalid`. Cheap check: write
   symptoms with **no double quotes at all**. Recorded in RUNBOOK with the error text so
   the next session recognises it. → `WRONG_CALLS.md`, and it is a LANDMINE candidate
   (a trap on a specific command where the failure is a confusing type error, not a
   "your quoting is wrong" message).
2. **Three round trips lost to guessed column names**: `site_work_items.attempts` (it is
   `attempt_count`), `orchestration_states.id` / `.agent_type` (they are
   `orchestration_id` / `owner_agent_type`), `site_discovery_rotation.last_run_at` (it is
   `last_selected_at`). CLAUDE.md says `\d <table>` before writing SQL and I did not.
3. **I nearly fired a live full page-build on another lane's active site as a "canary".**
   The approved plan called for it. On re-reading, an `image_landed` item routes to
   page-build-handler, which is the **full LLM build** — it would have rewritten
   mortgagecalculator's tool-repayment content while the mcalc adoption lane is active
   (92 commits/14d). The read-only census answered the same question better: it gave me
   eight cases instead of one, and the two positives are what made the experiment
   informative. **A destructive canary that yields one bit is worse than a census that
   yields eight.** Not run.

### Adjacent things found, deliberately not fixed here

- **The design-discovery rotation has been stalled since 08-11** (`site_discovery_rotation`:
  22 rows, newest `last_selected_at` 2026-08-11, while availability/completeness/quality/
  render-audit are all 08-21 or 08-22). That is the lane which runs
  `check_content_image_missing`, so the sweep-driven convergence this bug depends on has
  been dead for eleven days. The `site-discovery-staleness` CronJob (`bugs_open/230`)
  reports it **daily** — the design lane simply never appears in "stamps advanced last
  24h" — and nothing consumes the report. Belongs to 230's lane; contributing the
  measurement, not the fix. It is also the argument for making convergence event-driven
  rather than sweep-driven, which is this lane's Task B.
- `check_undeployed_assets.go:289-305` matches `rendered_html` against the **underscored
  purpose** (`content_hero.` / `content_hero-`) while deployed files carry the
  **hyphenated key** (`content-hero-tool-x.jpg`). `[UNVERIFIED]` — needs one query
  before filing; if it holds, that check cannot see any content hero as deployed.

### Stale claims in `bugs_open/114` itself (correction owed)

1. The ADDENDUM says the LLM-free rerender path "does not re-resolve fields", citing
   `flag_page_image_rebuild_action.go`. The header says that of the **terminal assemble
   leg**; `rerender_page_sections` **does** re-resolve (`:20-23`, `:459`).
2. Its `plan_sections_action.go:1608` citation is stale — that comment is now at
   `:2424-2432`.
3. Its merge-order claim (site-wide `hero_url` beats per-page `content_data.hero_url`) is
   contradicted by source: fresh `plan.ResolvedData` is merged **last** and wins
   (`rerender_page_sections_action.go:614-620`); the base only wins when the fresh data
   carries no hero.
4. "Why nothing caught any of it" quotes `check_image_url_404`'s old header. That check
   was fixed and closed (`bugs_closed/128`, live v1.0.1219) — it compares exact deployed
   paths now. The half of the claim that still stands is class (c): paths outside the
   `/assets/images/` prefix remain invisible to it.

---

## 2026-08-22 (later) — part 1 committed; two of my own hypotheses refuted; GAP 4 left narrowed, not solved

### Shipped

- `ebd1ce890` — the `store_asset` gate (IMG-072), council corr `3c0560f3-2873-439f-8311-61fde3903fc7`
  (`Council-Submitted:` trailer, verdict owed a read).
- `736108464` — the fallback branch of `ensureAssets` now logs that it fired and whether
  the page-scoped routes were eligible.

Both are Go, so **inert until the next chassis roll**. Nothing is proven live yet; the
roll is where the demand controls run (a page-scoped store that must NOT move
`content_data`, a canonical store that must).

### The register row travelled in someone else's commit

I named both register files on my `git commit` pathspec. The commit landed with the
**entry** (`imagery.md`) and without the **row** (`000_concept_index.md`) — because
another session had already committed my index edit as a same-file passenger in
`5fddba825` (the TL-049 lane) between my write and my commit. Verified at HEAD: row 1,
entry 1, the pair is intact and nothing was lost.

This is the hazard CLAUDE.md names and says no hook can prevent — *"if two sessions edit
one file, whoever commits takes both edits"*. Worth recording because the visible symptom
was the **commit-scope block listing 4 docs where I had named 5**, and the natural reading
of that is "my pathspec was wrong", not "my edit already shipped elsewhere". The check is
`git log -S '<your row text>' -- <file>`, which names the commit that actually carried it.

### GAP 4: what I refuted, and what survives

Full account in `WRONG_CALLS.md`; the short form, because the *sequence* is the lesson.

**Refuted 1 — "which flow ran".** The 090 loop (run corr `ea7dfeef-…`) came back
UNVERIFIABLE but narrowing: both the wired and the unwired page carried
`handler_agent='page-build-handler'`, and the failing page's hero write falls inside that
item's own run window after the only later rerender had completed. One run, and it killed
the explanation I would otherwise have written into this file.

**Refuted 2 — "a stored value is sticky".** `page_component_history` showed the failing
page carrying `/assets/images/hero.jpg` since 13:46 and the wired pages carrying no
`hero_url` in any archived version. A named mechanism (`carryStored`) even existed to hang
it on. Then I widened the population: **ten pages fleet-wide carry that value in history
and are wired to a content-hero today** (`idea.uk/tool-funding-fit`: 23 such versions).
Refuted in one `GROUP BY`, which I nearly skipped because the correlation inside my sample
was perfect.

**Also wrong along the way** — I queried `content_components.input_schema->'background_image'`
and got NULL for four rows, and read that as "the component declares neither field". The
fields live under `input_schema.fields`. `background_image` IS declared, with
`source: site_assets.hero` and `on_missing: use_fallback`. Four rows agreeing perfectly is
what made it convincing, and that uniformity was evidence about my PATH, not the data.
**`jsonb_pretty` the object once before writing predicates against its interior.**

**What survives, stated as narrowly as the evidence allows:** routes 1, 2 and 4 of
`ensureAssets` are all gated on `pageName != ""`; route 3 needed a `site_plan_imagery`
row mcalc does not have; route 5 is ungated. The observed outcome is exactly "every
pageName-gated route skipped". **This is NOT asserted as the cause** — the 08-15
orchestration rows are purged and the runtime trace the loop asked for cannot be
recovered. Hence the observability commit rather than a speculative code fix: the next
occurrence is a grep, not a dead end.

### Open, in order

1. Read the council verdict on `3c0560f3` and act on a REVISE (the code is already on the
   shared branch).
2. After the roll: pod-verify per service, then the held repair migration for the 18 sites.
3. Part 2 — the entity link at generation + event-driven DERIVE filing.
4. Part 3 — the flag-only detection check; queue hygiene for the 8 parked rows.

---

## 2026-08-22 (later still) — my "adjacent defect" in `check_undeployed_assets` is REFUTED, by the comment sitting on the function

I recorded, marked `[UNVERIFIED]`, that `check_undeployed_assets.go:289-305` matches
`rendered_html` against the **underscored purpose** (`content_hero.` / `content_hero-`)
while deployed files carry the **hyphenated key** (`content-hero-tool-x.jpg`), and that
the check therefore cannot see a content hero as deployed.

**It is wrong. In SQL `LIKE`, `_` is a single-character wildcard.** So
`'%/assets/images/content_hero-%'` matches `/assets/images/content-hero-tool-x.jpg` —
the underscore matches the hyphen. Measured, both predicates over deployed
`page_components`:

```
LIKE '%/assets/images/content-hero-%'                                        -> 31
LIKE '%/assets/images/content_hero.%' OR LIKE '%…/content_hero-%'  (the check) -> 31
```

Identical. The check sees them.

**And the code says so.** Immediately above the query:

> *"The `LIKE … || purpose || …` pattern is deliberately left unescaped. Read the
> LANDMINE section of the file header before changing it."*

I read that line, quoted the query underneath it, and did not read the header it points
at. The unescaped underscore is a deliberate design decision, not an oversight, and the
comment exists precisely to stop somebody "fixing" it.

**What saved this from becoming a filed bug** was the `[UNVERIFIED]` marker and the note
that it "needs one query before filing". That is the marker rule doing exactly its job:
the claim was written in a form that made it obvious it had not been checked, so checking
it was the natural next step rather than an afterthought. Had I written it unmarked, in
the same voice as the measured findings around it, it would have gone into the bug file
as a finding and cost the next reader a wasted investigation.

**The check:** when a comment on the code you are about to criticise tells you to read
something first, read it — that comment is usually the previous person who thought what
you are thinking. And for SQL specifically: `_` and `%` are wildcards in `LIKE`, so a
pattern built by concatenating an identifier that contains `_` is not doing string
equality on that identifier.

Corrected in the bug file too. → `WRONG_CALLS.md`.

---

## 2026-08-22 (part 2) — convergence moved to the event; a plan element DROPPED after reading; and the same testing mistake made twice

### Shipped

`d2b38b2ae` — `flag_page_image_rebuild` now files the `needs_content_image` DERIVE
itself, in the same transaction as the page re-render, when the landed page has a
content hero and no card. Register **IMG-073**. Go only, inert until the roll.

### A plan element I dropped, and why — read the readers before adding the writer

The approved plan's Task B step 1 was *"add optional `entity_type`/`entity_id` config
to `store_asset`"*. **I did not do it, because it would have fixed nothing.** Both
readers of the entity link require `purpose='card'`:

```
queryresolve.go:370-372   ca.entity_type='page' AND ca.entity_id=p.id AND ca.purpose='card'
check_content_image_missing.go:219-221   (identical predicate)
```

So an entity link on a `content_hero` asset has **no reader at all**. Adding the config
would have been an opt-in key with zero live consumers — exactly the accumulation
RFC_022's optional-key budget exists to prevent, and exactly the shape the owner ruling
calls out ("ten individually inert opt-in fields are a shared action nobody
understands"). The page's own hero already resolves **by key** through Lane B, needing no
link at all.

What actually needed fixing was only the convergence: the card is what the readers want,
`derive_card_asset` already writes the link correctly, and nothing filed the DERIVE
except a sweep that has been dead since 08-11. **The plan was wrong about the mechanism
and the code said so in two greps.** Recorded here because the plan had already been
approved — an approved plan is not evidence.

### MISSTEP — the same testing error, twice in one session, hours apart

Building the emitter's test I asserted
`discovery_checks.ContentImageItemKey(page) == "content_image:"+page`. Mutating the
emitter to use a hand-spelled `"content-image:"+page` — a hyphen where the contract has
an underscore, precisely the drift the shared helper exists to prevent — **passed**.

That is character-for-character the mistake I made this morning on `store_asset`'s URL
derivation (asserting `storage.DeployedWebPath` directly rather than the action's use of
it), which I had already written up in `WRONG_CALLS.md` and in IMG-072's register entry
before making it again.

**The fix is the same both times, and it is structural, not a matter of care:** extract
what the code under test must call into a pure function that **returns the value**, then
assert on that. `storeAssetContentDataUpdate` returns `(url, writeSiteWide)`;
`contentCardDeriveItem` returns the whole `workItem`. Both are now mutation-provable.

The generalisable rule: **a test that never touches the code under test passes every
mutation of it.** Writing the lesson down did not stop me repeating it four hours later
— only changing the shape of the code did. That asymmetry is the point worth carrying.

### Pattern-check note

The commit drew `unpaired-change: touches idx_swi_dedup but not workItemTerminalStatuses`.
It is a **comment mention**, not a contract change: the emitter files an existing
`item_type` at a non-terminal status through the existing `insertWorkItem`. The index and
the Go terminal-status list are untouched. Flagging it because it is the
"a source-scanning check makes your COMMENTS load-bearing" shape — the checker is right
to be twitchy about that identifier, and the honest answer is a note rather than removing
the word from a comment that earns its place.

### Council trail on corr `3c0560f3`

- Round 1 **REVISE** — `DeployedWebPath` vs `DeployedAssetPath` might differ for variant
  keys. Fair; nothing in the submission proved they do not. Answered with a test
  (`de1945f87`) rather than prose.
- Round 2 **REVISE** — the gate is per-STEP config, so a brand-update step with a dynamic
  `asset_key` could still write site-wide with a page-scoped key. **A real residual.**
  Answered in two parts: reachability IS per-invocation (the only dynamic-key true-step
  sits behind `input_data.spec.brand_update == true`, and the other two declare a literal
  purpose with no `asset_key_field`), and the remaining hole is now a WARN naming the key
  so a future producer that files `hero_about` as a brand update is greppable.
- Round 3 submitted with both answers.

Two of three rounds found something worth changing. That matches the lane memory
("a REVISE round is cheaper than the defect it finds") and is the argument against
treating the gate as a formality.

---

## 2026-08-22 — council APPROVED at round 3 (corr `3c0560f3-2873-439f-8311-61fde3903fc7`)

Verdict read before this was written, per the rule about never claiming a review you have
not read. **APPROVED, 2 advisory objections, neither high-severity, neither a code defect:**

1. *"Edit 4 and edit 7 are both 'add' operations targeting the same new file
   (`store_asset_site_brand_state_test.go`). If applied as literal sequential patches this
   is a conflict — the second should be 'modify'."* **Correct, and mine.** The second edit
   was the round-1 answer appended to an existing file, and I described it as an `add`.
   A submission-shape error, not a code one; worth remembering because the plan schema is
   read as a literal patch series.
2. *"Concept-register entry and its index row are documentation, not code that fixes the
   mechanism — fine as house convention but should not be counted toward 'the fix' when
   judging minimality."* Fair. They are in the commit because the platform-seams ruling
   requires the seam registered in the same commit that ships it, not because they fix
   anything.

The reviewer's own summary is worth keeping, because it states the fix better than my
rationale did:

> *"The round-2 reachability residual is honestly scoped (only one dynamic-key
> true-declaring step, gated behind a per-item conditional) and left visible via a WARN
> rather than silently over-restricted. `result[purpose+_url]` is updated in the same edit
> as the persisted write, so the in-run consumer doesn't retain the old derivation."*

**The round trail, and why the gate earned its cost here.** Round 1 REVISE — nothing in
my submission proved `DeployedWebPath` and `DeployedAssetPath` agree; answered with a
test. Round 2 REVISE — the per-step config gate cannot see a per-invocation `asset_key`;
a real residual I had not spotted, now a WARN. Round 3 APPROVED. **Two of three rounds
changed the work.** The commits carry `Council-Submitted:` and 098 credits them
automatically now the correlation is approved — no amend, which forward-only forbids
anyway.

### What is owed next, in order

1. **The roll.** Everything so far is Go and therefore inert. On the next chassis build:
   per-service build provenance, then the demand controls — a page-scoped store that must
   NOT move `content_data`, a canonical store that must.
2. **The held repair migration** for the 18 poisoned sites, only after (1) is proven at
   the artefact.
3. **The mcalc fixture end to end** — 10 heroes, 0 cards, 0 links since 08-15. First
   landing after the roll should file the derive, link the card, and put the content-hero
   path on the served page.
4. **Part 3** — the flag-only detection check, and hygiene for the 8 parked rows.

---

## 2026-08-22 (post-roll) — v1.0.1326 verified at the artefact, repair 562 APPLIED, and a landmine I had in my own memory index

### The roll, proven rather than assumed

Pods `agent-chassis-6bb7b67bd4-{l8lzf,pwtbf}`, image `v1.0.1326`, started 15:10Z. The
`build provenance` startup line had already scrolled out of `--tail=300`, which is the
documented shelf-life problem, so I probed the binary for the CAPABILITY instead of the
commit — three new literals, on BOTH replicas, with two controls in the same pass:

```
PRESENT  "Left site-wide content_data untouched"                 (part 1)
PRESENT  "queued card derivation at the landing event"           (part 2)
PRESENT  "hero resolved from the site-wide content_data fallback" (part 1b)
PRESENT  "StoreAssetAction: Asset stored"          <- control that MUST be present
ABSENT   "StoreAssetAction: XYZZY-not-a-real-string" <- control that MUST be absent
```

The absent-control is the half that makes the others mean anything: without it a
`grep` that matched everything would read identically.

### Repair 562 applied, with both controls INDUCED

`INSERT 0 12` (per-row backup), `UPDATE 11` (repointed), `UPDATE 1` (key removed),
`NOTICE repair OK: 11 point at hero-home.jpg, 7 correctly retain hero.jpg, 0 broken`.

- **Guard induced**: re-run WITHOUT the marker → `ERROR: REFUSING…`, nothing touched. A
  guard nobody has watched refuse is not a guard.
- **Idempotency**: re-run WITH the marker → `INSERT 0 0`, `UPDATE 0`, `UPDATE 0`.
- **Wire probe after**: `hero-home.jpg` = 200 on fundamentallyai, relojistas, vonc,
  dartsonline — every one of which served **404** on `hero.jpg` before.

### MISSTEP — a false claim in my own migration header, caught by previewing

The header's first draft said `noted.co.uk` and `webdesign.uk` were *"excluded by the
WHERE clause's own logic"*. **They are not.** Both hold a `hero_home` asset, so ARM 1
catches them. I found it by running the three arms as read-only `SELECT`s before
applying, rather than trusting what I had just written two minutes earlier.

Corrected in place, with the reason they are included anyway. The lesson is narrow and
useful: **a migration header describing a WHERE clause is a claim about code, and the
cheapest check is to run the clause as a SELECT and read the list of names it returns.**
That preview costs one query and it is the only thing that distinguishes a header from a
hope.

### MISSTEP — backticks in `git commit -m` executed, and this trap is in MY OWN memory index

The commit message for 562 contained ``an active canonical `hero` asset`` and
``an active `hero_home```. Bash ran them as command substitution inside the
double-quoted string:

```
/bin/bash: line 50: hero: command not found
/bin/bash: line 50: hero_home: command not found
```

Both words were **replaced with nothing** in the committed message, which now reads
*"holds an active canonical  asset"* and *"no canonical hero but an active  ->
repointed"*. The message is degraded, not false, and forward-only forbids an amend, so it
stands corrected here instead.

**What makes this worth writing down is that the trap is already documented**, in
`MEMORY.md`'s own index line: *"backticks in `-m` execute"*. I have read that line at the
start of this session. Knowing it did not stop me, which is the **third** instance today
of the same shape — the other two being the duplicated test-assertion mistake I had
written up hours before repeating it. **A documented trap is not a control.** The
mechanical control here is single quotes for the whole `-m` body, or a heredoc via
`git commit -F -`, and that is what I will use for the rest of this lane.

### Docs brought in line with today's owner ruling on counts

CLAUDE.md gained a ruling this session: *a count of things must carry the date it was
counted*, because a census does not go wrong, it goes **stale by addition**, and it reads
as current for ever. My figures sat under section headers dated 2026-08-22, which is
weaker than the rule asks. The most quotable ones now carry the date inline — `518 of 580
as of 2026-08-22`, `23 of 94 as of 2026-08-22`, and the `store_asset` caller census —
each with the re-derivation pointer, so `--since` is mechanically available to whoever
quotes them next.
