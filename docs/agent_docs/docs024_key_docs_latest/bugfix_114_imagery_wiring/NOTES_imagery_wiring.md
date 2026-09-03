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

---

## 2026-08-22 18:03Z — THE ACCEPTANCE TEST PASSED, and it nearly read as a failure

Filed one `needs_content_image` item by hand, in the exact shape `emitContentCardDerive`
produces (`triaged`, `asset-deployer`, item key `content_image:tool-repayment`, the shared
spec), for the designated fixture page. This tests the CONSUMER half — the half that has
never run — without generating an image or rewriting any page content.

**Result, at the artefact and not the status:**

```
work item          -> complete, no error
assets row         -> card_tool_repayment, active, entity_type='page',
                      entity_id=1f59a7d6-…, origin lineage set
the reader join    -> queryresolve.pageImageJoins resolves it for that page
the file           -> 200
```

mortgagecalculator had **0** card assets and **0** entity links across its ten 08-15
heroes. It has one of each now, and the listing-card reader can see it — which is the
precise thing `bugs_open/114` says never happens.

### ⚠ The near-miss worth carrying

**My first probe of the card file returned 404.** I had already started asking whether
`derive_card_asset` deploys at all, and was one step from filing "the emitter will produce
linked cards whose files 404" as a finding. The re-probe ~4 minutes later returned **200**:
the file landed between the two.

That is the trap **this bug's own file already records** — 41 images reported broken, 35
served 200 on an unhurried retry — and I walked toward it anyway. What stopped it was
comparing against sibling rows on other sites first (2 of 3 served), which made "mine is
uniquely broken" implausible and prompted the re-probe rather than the write-up.

**The check: never conclude a deploy failed from ONE probe.** Re-probe on an unhurried
retry, and where possible compare against a sibling of the same shape — if the siblings
serve, the difference is probably time, not mechanism.

**Applied consistently, it also found a real one:** `gamesdesign.co.uk`'s
`card_tool_gacha_pity_designer` still 404s on re-probe, and its asset was created
**2026-08-17** — five days is not a lag. A derived card whose file never deployed. Not
114's class, not this lane's to fix, recorded in the handoff for whoever owns the derive
path.

### What this does and does not prove

- **Proven:** the consumer path end to end — a `triaged` `needs_content_image` in the
  emitter's exact shape is claimed, derives a card, writes the entity link, and serves.
- **NOT proven:** that `flag_page_image_rebuild` actually emits one on a real landing. The
  emitter is unit-tested with five mutation proofs and its consumer is now proven live, but
  the two have not yet met in production. **First natural imagery landing is the thing to
  watch** — grep the chassis logs for `emit_content_card_derive`, which logs every
  disposition including each skip.

---

## 2026-09-02 — LANE RESUMED after 11 days idle. Verification first, and the world moved a long way on its own

Session `bugs_open/114`. Ownership re-checked before resuming: no dirty files touch this
lane, last commit 2026-08-22 (`1d0b9f407`), `who-owns` names this dir as owner. Two fresh
CONTRIBs landed on the bug file today from `inline_guide_imagery` and
`editorial_design_uplift` — contributions, not competing fixes. Chassis is `v1.0.1354`
(deployment image tag; was `v1.0.1326` at handoff).

### The handoff's closing bar, re-measured — items 1–3 are MET

**1. The emitter HAS fired naturally — 193 times.** `[MEASURED 2026-09-02]`
```sql
SELECT created_by, spec->>'check', count(*) FROM site_work_items
WHERE item_type='needs_content_image' AND created_at > '2026-08-22' GROUP BY 1,2;
-- image-build-handler | flag_page_image_rebuild | 193   (08-26 .. 09-01 19:30)
-- design-discovery-agent | content_image_missing |  85   (sweep revived 08-25)
```
The `created_by`/`spec.check` split is what distinguishes emitter from sweep — the
item_key shape alone cannot (both use `ContentImageItemKey`, by design).

**2. 193 of 193 emitter-filed items produced an entity-linked active card.**
```sql
SELECT count(*), count(*) FILTER (WHERE EXISTS (SELECT 1 FROM assets c
  WHERE c.site_id=w.site_id AND c.entity_type='page'
    AND c.entity_id::text=w.spec->>'entity_id' AND c.purpose='card' AND c.status='active'))
FROM site_work_items w WHERE w.item_type='needs_content_image'
  AND w.created_by='image-build-handler' AND w.status='complete';   -- 193 | 193
```
File probes: `dartsonline.com/assets/images/card-tool-checkout-calculator.jpg` **200**,
`leopardessconsulting.co.uk/assets/images/card-tool-agent-complexity-estimator.jpg` **200**.
(`boxingonline.com` returns 000 on its own HOMEPAGE — site-level unreachability, not a
card failure; its card rows are linked. Left to that site's lane.)

**3. No post-roll poisoning.** `hero_url` now on 33 sites, 2 distinct values — 19 ×
`hero.jpg`, and **all 19 have an active canonical `hero` asset** (legitimate under the
gate); 14 × `hero-home.jpg` (deployer-derived, legitimate). The one ambiguous case:
**apis.uk** carries `illustration_url='/assets/images/illustration.jpg'` and was created
08-22 (roll day). **The value itself is the discriminator**: apis.uk's illustration stores
are page-scoped (`illustration_waggle_dance`, …, four MORE on 08-23, post-roll) — old code
wrote the purpose-derived literal, the gated code would have written the key-derived path
(`illustration-beetle-hole.jpg`). The stored value is the OLD derivation and the 08-23
stores did not rewrite it. **The gate held; the write predates the roll.**

**4. The detection check — still not built.** That is this resumption's main task, and
the design changed today (below).

### MISSTEP (recorded in WRONG_CALLS too): I asserted "0 of 193 linked" off a JSON path that does not exist

My first join tested `c.entity_id::text = w.spec->>'page_id'`. The spec key is
`entity_id`, not `page_id` — 193 rows "agreed" because the probe returned NULL for all of
them. **This lane's own 08-22 handoff lists this exact trap** ("a JSON path probe cannot
distinguish not-declared from not-there. Four rows agreeing perfectly is evidence about
your PATH") and I hit it anyway, in the same lane, on the same table. Caught within
minutes by `jsonb_pretty(spec)` on one row — which is the check that should precede any
`spec->>'k'` join, not follow it. Third recorded instance of "a documented trap is not a
control" in this lane.

### MISSTEP AVERTED: nearly filed a 090 for a mechanism that is already diagnosed (bugs_open/357)

mcalc's fresh 09-02 handoff (their lane, cold-start doc) found: stored
`content_data.background_image` never reaches `rendered_html` on `page_type='tool'`,
while the identical hero component renders on guide pages. Verified at the artefact
myself: mcalc tool heroes render NO `background-image` and their `rendered_html` opens
`<div class="tool-page">` — **the whole tool shell, not hero-template output**. I had a
090 symptom drafted. One grep first (`tool-page` in Go) landed on
`adopt_fragment_section.go`'s header: this is **RFC_046 / bugs_open/357** — a tool page
arrives as one fragment, is stored as a single section, and the identity sentinel is
replaced by `planned[Position-1]`, so the row DECLARES itself `hero` while storing the
tool. Already diagnosed, already owned, fix (constructive adoption, opt-in default OFF)
already built by that lane. **Grep before you file** is the whole lesson; a duplicate
run was one grep away.

**Consequence for GAP 4:** the 08-15 mcalc cohort (the natural experiment in the PLAN)
was ALL tool pages. Its wired/fallback contrast was measured at `content_data` — on rows
whose `rendered_html` is a tool fragment either way. **GAP 4 largely dissolves into 357
for tool pages.** For non-tool pages the wiring gap is `bugs_open/412`'s (§8: full build
ran 9×, wired 1 — their [INFERRED] cause unconfirmed; their fix candidate 1, deploy-time
wiring, is the structural sibling of our IMG-073). 412 is OWNED (finetuning lane, active
today) — coordinated, not taken.

### NEW FINDING — the platform already has a check aimed near GAP 5, and it is making the bug worse, not better

`check_undeployed_assets` half 1 files `undeployed_asset` items for assets no deployed
component references. Three defects compound `[MEASURED 2026-09-02]`:
- **evidence is purpose-prefix, site-wide** (`rendered_html LIKE '%/assets/images/'||purpose||'[.-]%'`):
  one wired sibling vouches for every asset of that purpose (dartsonline's 17 unwired
  content heroes invisible), zero wired flags all of them (webdesign.co.uk's 66);
- **the remedy is a deploy** (`handler_agent='asset-deployer'`), which re-commits the file
  and cannot wire a page;
- **the recurrence brake then parks the refiled items at birth**: 1,651 `undeployed_asset`
  rows sit `unresolved` with `created_at = updated_at` and `result={}` (1,086 of the icon
  ones born since 08-23). The platform has already noticed the non-convergence — by
  parking it where nobody looks. This is 114's "queue nobody drains", industrialised.

### NEW MEASUREMENT — the render-capability census that resizes the whole bug

Per page of the two GENERATE surfaces, does ANY component's template carry an image
branch (`hero_url`/`background_image` — bug 412 §7's test), and is that component's
rendered row actually a tool fragment (interactiveStructuralMarkers)?

| page_type | pages | no image slot | slot is a 357 fragment | genuinely capable |
|---|---|---|---|---|
| tool | 335 | **231** | 16 | 88 |
| blog-post | 319 | 7 | 0 | 312 |

`[MEASURED 2026-09-02]` So **~74% of tool pages cannot render a content hero at all**.
Generation there is NOT pure waste — the hero seeds the card derivation and cards DO
render on listings (mcalc's 6 tool cards serve today) — which is why I considered and
**REJECTED** a capability gate on the GENERATE arm: it would silently trade away the
card benefit. The decision goes to the owner with these numbers instead, via the check
below.

### Residual poisoned keys — analysis complete, migration is safe

`icon_url` (16 sites) / `content_hero_url` (6) / `illustration_url` (4) /
`sprite_sheet_url` (1): **zero sites have a canonical asset under any of those keys**
(`EXISTS(assets a WHERE a.asset_key = replace(k,'_url',''))` = 0 for all), probes 404
(`idea.uk/assets/images/icon.jpg`, `gamesdesign.co.uk/assets/images/content_hero.jpg`).
The one flagged consumer, `brief-explanation.html_template`'s `{{.illustration_url}}`, is
a FIELD sourced `site_assets.illustration` — resolved from plan/asset tables, NOT from
`sites.content_data`: the resolver's content_data fallback handles only `hero_url` and
`logo_url` (`plan_sections_action.go:653-694`). `logo_url` (26 sites, 1 value) is
**26/26 canonical-backed — legitimate, do not touch.** So deleting the four dead keys is
behaviour-neutral hygiene; migration follows 562's pattern (backup, DO/RAISE, idempotent).

### Council debt: migration 562's verdict was never produced

`orchestration_states` has **no row** for corr `4145fcdc-9ffe-42e0-a547-49e07bda04db` —
11 days on, that is a dropped dispatch, not latency (the ~30-min rule covers queueing, not
absence). Resubmit with `RESUBMIT_CORR` so the trail accumulates.

### 2026-09-02 evening — submissions dispatched, a peer's correction folded in, queue hygiene recorded

**Council: both submissions dispatched and EXECUTING within ~3 minutes** (no 29-minute
queue today): the detector under corr `3b568104-566d-43d9-8d73-d30fbdf6e9df` (commit
`a87746b77`, `Council-Submitted` trailer), and 562's RESUBMISSION on its original trail
`4145fcdc-…` (the 08-22 dispatch was dropped — 0 orchestration rows in 11 days). Both at
`review_reuse_agent` when checked. Verdicts owed a read.

**boxingonline "000" CORRECTED by the editorial_design_uplift session (cross-session
reply, ~18:20Z): not an outage — the customer domain has NO DNS record yet (pre-cutover,
site 2 days old); the site serves at its publish target `boxingonline.ugg2.com`
(`sites.publish_target`/`publish_project` name it).** Re-probed there:
`card-womens-boxing-having-a-moment.jpg` **200**, invented-filename control **404**. So
closing-bar item 2 is proven on a third site, and the general rule is worth its words:
**probe the PUBLISH TARGET, not the customer domain** — a 000 on the customer domain is
what an un-pointed name looks like (and the parked-domain inverse, which 200s every
path, is already in LANDMINES).

**Their second correction, folded into the code pre-verdict** (one-line commit after
`a87746b77`): the `no_image_slot` remedy text now states the composition question AS a
question — the census established that these pages lack an image-capable component, not
WHY the planner composed them that way; asserting a cause would overstate what was
measured, and `bugs_open/412` (unverified from here) may hold the actual answer. My
CONTRIB into their lane said their reframe was "quoted in the state's remedy text" —
it was PARAPHRASED, and the correction below the CONTRIB says so.

**Queue hygiene (PLAN C4), disposition recorded rather than acted on:**
- 3 × `still_holds` (`loanzy.uk your-rights`, `remortgagecalculator.uk
  six-month-checklist` + `what-your-number-means`, all `page_type='content'`): the
  revalidator says the premise holds; they are their sites' lanes' rows. LEFT PARKED.
- 4 × robot-hands `page_rerender:tool-*` (reval `unknown`, all `page_type='tool'`):
  357-shaped — the hero row on those pages is a tool fragment, so a rerender through the
  current binary would neither harm nor wire. LEFT PARKED; when IMG-077 rolls they
  surface as `fragment_slot` members with the mechanism named, which is strictly more
  informative than a cancelled row. Cancelling would erase the only queue record of the
  original landing.
- The 7 `failed` `page_rerender` rows (08-27..31, dartsonline/farmerinsurance/
  loanandmortgagecalculator/leopardess): their sites' lanes; result rows show
  `mark_item_failed` with `completion_skipped` — not this lane's mechanism. Reported
  here, not touched.

### 2026-09-02 late evening — the 412 handover executed, three council rounds in flight, and one procedural misstep the council caught

**The 412 handover.** The owning finetuning lane answered the build-or-keep question in
writing (`bugs_open/412` §10): THIS lane builds candidate 1 — their three stated reasons:
their lane redirected off platform work; our 193/193 event-seam evidence beats their
1-of-9; IMG-077's rollups give an acceptance population they structurally cannot. Their
refinement folded into the design before building: finetuning.uk is excluded from
acceptance NOT because it is "masked" but because 664 changed the JOIN and 649 the
SCHEMA — two defects overlap there and a before/after cannot attribute a delta.

**Built: `wirePageHeroOnLanding` (IMG-078, commit `8aa51f599`, council corr `bd78490d`).**
At the landing event, same tx as the re-render emit: DeployedWebPath(ContentHeroKey)
into hero-family rows' `content_data.hero_url`, only where BOTH image fields hold
empty/legacy/site-fallback, never on a 357-fragment row. Opt-in default OFF
(`wire_hero_on_landing`); armed post-roll by `710_…_HOLD.sql` for the ONE live carrier.
Four mutation proofs in a clean worktree (`git worktree`, because another session's
in-flight datahelpers WIP — `CTALabelCandidateRow` signature change on DIRTY files —
broke the shared tree's actions test build; their plan, not our bug; routed around,
never touched).

**Council round 1 results (detector corr `3b568104` + 562's `4145fcdc`):**
- **562: APPROVED**, round 1, 2 low advisories. The dropped-dispatch debt is discharged;
  the 08-22 commit's `Council-Submitted` trailer resolves automatically via 098.
- **Detector: REVISE**, gating objection from prior_art_librarian — and it was RIGHT:
  I cited `storage.DeployedWebPath` as the single source of truth **without grepping
  LANDMINES for the symbol**, the exact reflex the memory index demands
  (grep-landmines-for-your-symbols). The landmine, now read in full: **FIXED AND LIVE
  v1.0.1229** ("THE TRAP BELOW IS HISTORY"), historically brand-head-only, and its own
  guidance prescribes page_components as the comparison surface for this population —
  so the check was RIGHT by luck and is now right on the record (header carries the
  landmine back; commit `0ea807252`). **Misstep logged here rather than WRONG_CALLS: no
  false claim was recorded — the omission was an unread landmine, cost one round.**
- Other round-1 objections, each acted on: 709 SPLIT into its own submission (guardian)
  and hardened (debug_historian: pre-mutation counted gate 16/6/4/1-or-clean with abort
  on partial, per-arm ROW_COUNT assertions, ROLLBACK sibling; bug_historian +
  prior_art_librarian: the no-reader claim now VERIFIED — template sweep 0/0/0/1 with
  the 1 analysed, field-source sweep 0, and the council's own read-only check agreed at
  0 rows). **Dry-run of the full 709 against LIVE data under ROLLBACK: INSERT 18,
  removed 16/6/4/1, every assertion green** (commit `ae8029c31`). reuse seat's overlap
  question answered by premise in the header (the two neighbour checks trigger on
  ABSENT assets; this one only on PRESENT ones). guardian's lockstep-ownership note
  added. architecture seat's follow-up recorded in PLAN (below).
- The council's own re-measure of the parked backlog: **1,662** (+11 on my morning
  1,651 — stale by addition within hours; both figures dated).

**In flight at time of writing:** detector round 2 (same trail `3b568104`), 709
standalone (`151a51db`), the wiring (`bd78490d`). All three verdicts owed a read by
whoever is next here if this session ends first — query in the RUNBOOK.

### 2026-09-02 night — the wiring's REVISE is RIGHT, round 2 is parked behind a GAP-4 canary, and a shared-tree race cost five predicates

**The wiring (corr `bd78490d`) drew a REVISE and the gating objection improves the
design.** Full triage + the round-2 protocol: `bugs_open/412` §11a. The short form: the
content_data floor loses to the resolver's own fresh-merge on exactly the GAP-4 pages
that need it (bug_historian, twice, high); the raw UPDATE bypassed the lock guard stack
(render_guardian); and the seats converge on the resolver's own vocabulary — a
`site_plan_imagery` page-scope row at the landing (route 1) instead of a fifth
content_data writer. **But route 1 shares route 2's `pageName != ""` gate, so the
diagnosis comes first**: one controlled rerender on a quiet unwired page with the
disposition logs watched. The committed code is opt-in OFF and 710 is HELD — nothing
regresses while parked. Four of the round's objections were misapprehensions (symbols
the seats could not see exist — they were committed this week); round 2 cites shas.

**Shared-tree race, cost five predicates:** commits `ed8480a25`/`8aa51f599` hand-spelled
`build_status IS DISTINCT FROM 'removed'` while `def8126e3` (another session, same
hours) single-sourced it as `datahelpers.NotRemoved` with a source-scan test — so HEAD's
`TestNoHandSpelledTombstonePredicate` went red for every session until `d1cf3aac3`
adopted the helper at all five sites. Flagged cross-session by `bugs_open/436` within
the hour, which is the coordination working. (Their message also attributed a
third session's broken untracked test file to this lane — corrected in the reply;
worth the reminder that "dirty in the tree" carries no authorship.)

**Also mooted by the design shift:** the finetuning lane's name-vs-capability matcher
warning (4 hero-named components with no image field, all zero-instance) — the plan-row
design matches no components at all, and IMG-077's predicate was already
capability-based.

### 2026-09-02 close — the detector is APPROVED (round 3, corr `3b568104`, 19:09Z)

**"approved with 3 advisory objection(s) — none high-severity."** Ten advisories across
seats in total; adjudication:
- **guardian [medium], duplicate-active-rows preflight for 708**: ACTED ON — the apply
  runbook now opens with the preflight query and the landmine's name (the needle-gate
  already refuses in-transaction; the preflight saves the apply attempt).
- **bug_historian/reuse [medium] ×2, the 033 queue + the undeployed_assets coexistence**:
  already tracked — the queue's working surface is 033's, and the narrow-or-retire
  follow-up is in this PLAN's addendum and the bug file. The reuse seat is right that a
  sequenced unification has no forcing function; the forcing function is now the rollups
  themselves — once live they make the older check's false positives countable, which is
  the evidence the retirement decision needs.
- **prior_art_librarian [medium/low] ×4, verify-the-assertions**: the assertions (file
  lines, 709's APPROVED verdict, the CONTRIB file, the rolling-window caveat on the
  1,651 count) are all evidenced in this NOTES file and the diagnosis_artifacts rows;
  the rolling-window point is fair — that count carries its date everywhere it appears,
  and its LANDMINES entry says the born-parked query is the re-derivation.
- Remaining lows: the two mirrors (lockstep-fenced, single-sourcing offered to 357),
  the 24→25 apply-time re-derive (already in the header).

**FINAL LEDGER for this resumption, all corrs:** 562 repair **APPROVED** (`4145fcdc`,
trail completed after the dropped dispatch) · 709 dead-keys **APPROVED** (`151a51db`) ·
IMG-077 detector **APPROVED round 3** (`3b568104` — commits a87746b77, 0ea807252,
ec2b3353d, d1cf3aac3 resolve via their Council-Submitted trailers) · IMG-078 wiring
**REVISE, deliberately parked** behind the GAP-4 canary protocol (`bd78490d`,
412 §11a/§11b — plan-less domain limit and canary site recorded). Closing bar: items
1–3 MET and measured; item 4 built + approved, inert until a chassis roll carries
`d1cf3aac3`+ and migration 708 is applied per its runbook.

### 2026-09-02 postscript — RFC_063 filed by the 443 lane; this lane's consumer input is in it

The plan-less tier boundary (412 §11b) is now an owner decision:
`architecture_review/RFC_063_plan_tables_are_becoming_the_capability_tier…`. This lane
contributed (commit `b587f116e`): the two-row minimal materialisation route 1 needs, the
`check_unfulfilled_imagery_plan` spend trap for a naive option-B seed (seed imagery rows
from the ASSETS table, never from page enumeration; cookly.uk proves plan creation but
holds zero content-hero assets), and the honest down-ranking — imagery is B's weakest
argument because route 2 is plan-independent. The 443 lane anchored the qualifications
into the RFC's top-down path. **For whoever runs the wiring's round 2: the plan-less arm
now depends on the RFC_063 ruling — option B gives it for free, option A means route 2
stays the plan-less path and the round-2 submission says so as scope.**

---

## 2026-09-03 morning — the roll landed, both migrations discharged, and the detector FIRED on its first sweep

**Roll verified at the artefact** (never the tag): chassis `v1.0.1356`, pods born
08:57Z. Capability probe with controls: `unrendered_page_imagery` on **79 pods =
undeployed_assets on 79** (positive control), `_NOTREAL` 0 (negative control). Wiring
literal in `/proc/1/exe`: present=1, absent-control=0 — IMG-078 is aboard and inert
(opt-in OFF, 710 still deliberately HELD pending the GAP-4 canary).

**709 APPLIED** 09:1xZ: `removed icon=16 content_hero=6 illustration=4 sprite_sheet=1`,
18 rows in `bak_site_dead_purpose_urls_20260902` — byte-identical to the dry run. The
last poisoned `content_data` surface is gone.

**708 APPLIED** per its own runbook: preflight n=1 (duplicate-active-rows), UPDATE 1,
array 24→25, post-apply what-did-I-break clean. `_HOLD` suffixes discharged on both
files (git mv, both paths named on the commit, verified at HEAD — exactly one path
each); `liveConfiguredChecks` gained the name in the SAME commit (`8c4a5789e`,
`Council-Reviewed: 3b568104`).

**FIRST NATURAL FIRING, ~10 minutes after the apply:** idea.uk swept 09:25:32 → ONE
rollup, `unrendered_page_imagery:no_image_slot`, count **6**, `measured_at 2026-09-03`
— plausible against idea.uk's census (16 content-hero assets, 9 wired on 08-22) — and
the design-discovery orchestration **COMPLETED**, so the check ran alongside its 24
siblings without failing the step. Closing-bar item 4 is now exercised, not merely
built.

**What closure still WAITS ON, and it is observation, not work:** one sweep is one
site. Before moving 114 to `bugs_closed/`, let the rotation cover the fleet (~a day at
the >4h-stale floor) and confirm: rollups appear on the sites the census predicts
(webdesign.co.uk's unwired population, gamesdesign, the 16 fragment_slot tool pages),
at plausible counts, with the orchestrations completing. A fleet-wide pattern of
plausible rollups = bar met; a fleet-wide silence after full rotation = an unexercised
detector and a bug, not a clean fleet.

### 2026-09-03 afternoon — second roll (`v1.0.1358`), check survived it, coverage on track

Re-probed after the 12:06Z roll: `unrendered_page_imagery` on **73/73** reporting pods,
controls clean. Coverage since the 09:16Z enablement: **3 sites swept → 3 rollups, all
plausible** (idea.uk 6 / garden-tools.uk 3 / remortgagecalculator.uk 1, all
`no_image_slot`, all dated), **5/5 design orchestrations COMPLETED**. The closure
criterion still waits on the census-predicted `unwired` (webdesign.co.uk et al) and
`fragment_slot` (the 16 tool rows) sites entering the rotation — expected within ~a day.
`flag_page_image_rebuild_action.go` gained the 440 lane's routing_reason conversion
(`d44644635`) — coexists with step 2b; noted in the handoff so round 2 re-reads the file.
