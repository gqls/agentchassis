# 214 — imagery scope_refs are LLM-minted free text, never validated against the plan they point into; orphans ship silently and their assets are already paid for

**Filed 2026-08-07** by the brochure_component_library lane, promoting the latent
defect RFC_016 §1 flagged ("worth its own bug file") — and upgrading it: **it is
not latent. 5 of 131 section-scope imagery rows fleet-wide are orphaned today**,
one of them minted THIS MORNING by the newest planner prompt, and four of them
have active, generated assets that no page build can ever pick up.

**Verification route, declared per the 2026-07-31 owner ruling:** this file was
not put through the 090 diagnosis loop. Substituted first-hand verification:
every function in the chain read at HEAD this session with citations below
(mint site, drop sites, ordering assignment, both consumer joins), plus a live
fleet census whose result could have come out otherwise (it found 5 orphans,
not 0 — and would have found 0 on a healthy fleet). The mechanism is confined
to one writer and two consumers; nothing here rests on inference from grep hits.
A fixing thread that wants the loop's independent read should still run it.

## The mechanism

The planner LLM emits an `imagery` block whose section-scope entries are keyed
by **free-text `"page_name:ordering"` strings** (documented shape:
`write_site_plan_action.go:827-846`). `flattenImageryBlock` copies each key
verbatim into `site_plan_imagery.scope_ref`
(`write_site_plan_action.go:883-897`) — **nothing anywhere checks that the page
name or the ordinal resolves to a section that actually survived planning.**

Three ways the reference goes wrong, all observed or mechanically reachable:

1. **Minted out of range.** The LLM writes an ordinal past the page's section
   count. Live instance: fundamentallyai.com current plan (written 2026-08-07
   08:24 UTC, replan corr `801b0732`), `scope_ref='about:4'` for
   `illustration_people_approach` — the about page has sections 0–3; the
   illustration was almost certainly meant for `people-feature-block` at
   ordinal 2. A `needs_imagery` item for it was queued the same second, so an
   asset will be generated against a reference that resolves to nothing.
2. **Minted against a page-name variant.** The LLM keys imagery `about:...`
   while emitting the page's sections under `about-index`. Live instance:
   gamesdesign.co.uk current plan (2026-06-05), four icon rows at `about:2`
   (`icon_no_ads`, `icon_math_first`, `icon_browser_based`,
   `icon_practitioner`); the plan has **no** `about` page. All four assets were
   generated 2026-06-06 and are `status='active'` — paid for, deployed, and
   unreachable via the section-imagery join below.
3. **Ordinal-shift after drops** (the flavour RFC_016 §1 documents). The keys
   are minted against the section array AS EMITTED, but `ValidateSitePlanAction`
   drops entries after that: unresolvable section names
   (`v3_site_actions.go:3168-3173`), nameless objects (`:3227`), unrecognised
   shapes (`:3236`). Every drop compacts the array, and `write_site_plan`
   persists `ordering` = position in the POST-drop array
   (`write_site_plan_action.go:383`). One dropped entry mis-keys every later
   imagery ref on that page — silently, since both sides remain "valid".
   No live instance today (2026-08-07's fundamentallyai run persisted all 71
   emitted entries, zero drops), but the door is open on every replan.

## Who consumes scope_ref, and what each failure does to them

- **Page-level LIKE join** (`plan_sections_action.go:362-386`):
  `scope_ref LIKE <page> || ':%'`. Tolerates a wrong ordinal (flavours 1 and 3
  degrade to "imagery attributed to the right page, wrong section" — currently
  harmless because mapping is by key/kind, not ordinal), but a page-name
  mismatch (flavour 2) makes the row **invisible to every build** — the
  gamesdesign icons are exactly `bugs_open/114`'s symptom class (generated,
  deployed, never referenced), reached by a different door.
- **Lock carry-forward across plan rewrites**
  (`write_site_plan_action.go:651-720`): matches old-plan locked imagery to
  new-plan rows on exact `(scope, scope_ref, category, subject, ordering)`.
  Ordinal drift between plans breaks the match → a lock silently fails to
  carry (churn of pinned imagery, the webdesign colour-churn class) — or, on a
  page whose ordinals shifted, carries a lock onto the WRONG section's imagery.
- **`flag_page_image_rebuild_action.go:116`**: parses scope_ref only to find
  the page — same tolerance profile as the LIKE join.

## Evidence (all 2026-08-07, live DB)

Census — could have returned 0, returned 5:

```sql
SELECT count(*) AS total_section_scope,
       count(*) FILTER (WHERE scope_ref !~ '^[^:]+:[0-9]+$') AS malformed_ref,
       count(*) FILTER (WHERE scope_ref ~ '^[^:]+:[0-9]+$' AND
         (split_part(scope_ref,':',2))::int >= COALESCE((
           SELECT count(*) FROM site_plan_sections sps
           WHERE sps.plan_id=spi.plan_id
             AND sps.page_name=split_part(spi.scope_ref,':',1)),0)
       ) AS orphaned_ordinal
FROM site_plan_imagery spi WHERE spi.scope='section';
-- 2026-08-07: 131 | 0 | 5
```

The 5: fundamentallyai `about:4` (current plan, same-day), gamesdesign
`about:2` ×4 (current plan, minted 2026-06-05 in the same instant as the plan —
wrong at birth, not drifted). Asset check: all four gamesdesign keys have
`assets.status='active'` rows dated 2026-06-06.

## Fix candidates, ordered by what closes the door

1. **Validate at the only door rows enter.** `flattenImageryBlock` runs inside
   `write_site_plan`, AFTER validate_plan — the surviving sections array is in
   the same `CollectedData` it walks. Resolve each `page:ordinal` against it:
   unknown page or out-of-range ordinal → degrade the row to page scope (or
   skip) with a durable log, exactly the pattern `FACT_SCOPING_EMPTY_COMPOSITION`
   set for fact assignments. Makes an orphaned scope_ref unrepresentable;
   ~30 lines in one function; no prompt change; no consumer changes.
2. **Move imagery inside the section entry** — RFC_016 §1's stated rule for the
   next structured per-section field ("object-form entries live ONLY between
   the planner LLM and validate_plan's normalise pass"; alignment travels
   inside the entry, positional keying is the named counter-example). The
   door-closing fix at the contract level, but it is a planner-prompt +
   flatten + normalise change — fleet-wide prompt breadth is precisely what
   drew 151's guardian veto, so this goes through its own council round and
   should cite RFC_016 §5.1's ratification when it lands.
3. **Repair the 5 rows** (retarget fundamentallyai's `about:4`→`about:2`;
   re-key gamesdesign's four to `about-index:<n>` or re-scope to page). Needed
   under either fix; repairs nothing structurally on its own.

Candidate 1 is compatible with candidate 2 (validation stays as the guard even
after the carrier moves) — 1 need not wait for 2's council round.

## How to verify a fix

Re-run the census above: `orphaned_ordinal` must read 0, and STAY 0 across the
next replans (the fundamentallyai row proves the current prompt re-mints
orphans, so a one-time cleanup that passes the census once is not a fix —
re-run after the next planner run on any fact-rich or icon-heavy site).

## Relations

`RFC_016` §1 (contract + the counter-example this file promotes) ·
`bugs_open/151` (the lane that surfaced it; its fact assignments deliberately
rejected this keying scheme) · `bugs_open/114` (generated-imagery-never-
referenced — the gamesdesign four are that symptom via this cause) ·
`bugs_open/204` (positional slot NAMES unresolvable in pages.sections — sibling
positional-keying defect, different table, different consumer).

---

# 2026-08-10 — TAKEN, FIXED AT THE WRITE PATH (committed, inert until the roll), and this file is corrected in three places

Lane: `docs024_key_docs_latest/bugfix_214_imagery_scope_ref/` (PLAN / NOTES / RUNBOOK /
README_where_we_are / COUNCIL_SUBMISSION). Taken because the filing lane (`151`) had
explicitly left it — *"same wire, different field, its fix does not gate"* — and a grep
of all 39 live session transcripts found nobody on it.

**Re-verified valid before acting:** the census below still returns **5**, and the rows
are the *same five* (identity, not just count) — fundamentallyai `about:4` created
08-07, gamesdesign `about:2` ×4 created 06-05. Nothing drifted, nothing re-minted.

## CORRECTION 1 — the cause is CANONICALISATION DRIFT, not "LLM-minted free text"

§"The mechanism" reads flavour 2 as the LLM keying *"a page-name variant"*. It did not.
It keyed the name it was handed, and **the platform renamed the page underneath it**:

```
write_site_plan_action.go:392  CanonicalisePage(...)              -> canonical name
                        :503  INSERT site_plan_pages    name = r.Name      <- CANONICAL
                        :533  INSERT site_plan_sections page_name = r.Name <- CANONICAL
                        :455  flattenImageryBlock(...)                     <- RAW LLM KEY
```

`CanonicalisePage`'s section-index family (`page_canonical.go:159-173`) maps `about` →
`about-index`, `contact` → `contact-index`, `news` → `news-index`. Same function, ~60
lines apart, two of three tables canonicalised.

**This changes the right fix.** Candidate 1 as written ("degrade the row to page scope
or skip") would have **discarded four correctly-planned icons**: gamesdesign's
`about-index` has exactly three sections, so `about:2` → `about-index:2` has a *valid*
ordinal. The ordinal was right all along; only the page name moved. The fix is to
**resolve** the reference, not to reject it.

## CORRECTION 2 — this file's census measures the HARMLESS half, and is blind to the harmful one

The census filters on **ordinal range**. But §"Who consumes scope_ref" already says, in
this same file, that the page-level join `LIKE <page> || ':%'` *"tolerates a wrong
ordinal"* and that `flag_page_image_rebuild` parses the ref *"only to find the page"*.
Both are true: **no consumer parses the ordinal.** An out-of-range ordinal on a correct
page part is **behaviourally inert**. The census and the analysis contradict each other
and nothing in the file joins them up.

What is fatal is a wrong **page part** — and the census cannot see it, because it never
checks that the page resolves, and it excludes `scope='page'` entirely.

> ⚠ **And the census predicate must resolve against `pages.name`, NOT
> `site_plan_pages.name`.** All ten consumers join the deployed `pages` table.
> **I got this wrong first and it is logged in `WRONG_CALLS.md`:** measured against the
> plan table I got **22** unresolvable rows; against `pages` it is **10**, and all 12 of
> the difference resolve perfectly well. A repair built on the 22 would have repointed
> twelve *working* heroes on leopardessconsulting.co.uk and relojistas.com at names the
> live sites do not carry — breaking exactly what it was fixing.

**The honest damage, 2026-08-10, current plans: 10 of 176 rows, and 8 of the 10 already
have an `assets` row at `status='active'`** — planned, generated, deployed, paid for,
referenced by nothing:

| domain | scope | scope_ref | key | plan candidate | asset paid for |
|---|---|---|---|---|---|
| fundamentallyai.com | page | `news` | hero_news | `news-index` | yes |
| gamesdesign.co.uk | page | `about` | hero_about | `about-index` | yes |
| gamesdesign.co.uk | section | `about:2` ×4 | icon_* | `about-index` | yes |
| gamesdesign.co.uk | page | `contact` | hero_contact | `contact-index` | yes |
| mortgagecalculator.co.uk | page | `about` | hero_about | `about-index` | no |
| mortgagecalculator.co.uk | page | `contact` | hero_contact | `contact-index` | no |
| mortgagecalculator.co.uk | page | `tools-index` | hero_tools | **none** | no |

## CORRECTION 3 — the imagery lock-transfer key in §"Who consumes scope_ref" is wrong

The file states imagery locks carry forward on `(scope, scope_ref, category, subject,
ordering)`. **That key belongs to `transferDirectiveLocks` on `site_plan_directives`**
(`write_site_plan_action.go:786`), a different table. `transferImageryLocks` (`:1123`)
matches **`(plan_id, scope, scope_ref, key)`**.

Consequence, and it is not cosmetic: `scope_ref` **is** in the imagery key, so any
rewrite silently drops human-approved locks for one plan generation. That is why the
fix includes a canonical fallback in `transferImageryLocks`. A fix or test aimed by the
key as filed would hit the wrong function.

(Also stale: the cited line numbers. At HEAD the functions are `flattenImageryBlock`
~`:976`, `buildImageryRow` ~`:1044`.)

## What was built — commit `c21af5eda` (+ gofmt `c90212df6`)

`platform/orchestration/actions/write_site_plan_imagery_scope.go`, register **IMG-070**.

**The guarantee:** a page/section `scope_ref` written by this action either names a page
the plan contains, or is preserved **byte-for-byte** and leaves a durable
`agent_error_log` row. Nothing is ever silently dropped, so no row can regress.

- `buildCanonicalPageNameMap` — built from `planRows`, already canonicalised **and**
  deduped at that point. Identity pass first (an already-correct ref maps to itself and
  takes the no-op branch, which is what makes "working rows are untouched" true by
  construction). An alias two pages would claim is **refused, not guessed**.
- `canonicaliseImageryScopeRefs` — splits on the **first** colon, the split every
  consumer uses, and reattaches the remainder *with* its colon, so
  `chk_scope_ref_consistency` holds by construction rather than by special case.
- **The ordinal is validated and recorded, NEVER rewritten** —
  `IMAGERY_SCOPE_REF_ORDINAL_ANOMALY`. Three reasons: no consumer parses it; a rewrite
  risks 23505 on `idx_site_plan_imagery_unique` and breaks the lock key for a further
  generation; and the correct value is **unknowable at this seam** — ordinal shift
  happens in `ValidateSitePlanAction`, a different action, and the pre-drop array is
  gone by here. Fix-candidate 2 (imagery inside the section entry, RFC_016 §1) remains
  the only real answer and is **architecture-scope, not taken**.
- `dedupeImageryRows` — guards `idx_site_plan_imagery_unique` against the collapse the
  resolution itself makes possible (`bugs_open/215` verbatim on the sibling table).
  **Zero collisions live today**; prophylactic, and says so.
- `transferImageryLocks` canonical fallback — runs only after the exact match finds
  nothing, so it cannot alter a transfer that works today; retires itself after one
  replan.

**Fix-candidate disposition:** candidate 1 = done, but as *resolution* not degradation
(see Correction 1); candidate 2 = architecture-scope, not taken; candidate 3 (repair the
rows) = `sql_for_agents/373` + ROLLBACK, **written, committed, NOT APPLIED**.

**Wiring is mutation-proven, and the first attempt was not.** Fifteen unit tests passed
with the entire resolution block deleted from `WriteSitePlanAction` — measured, not
assumed — because every one called the helpers directly. The guard is now a separate
sqlmock suite driving the real action and asserting the INSERT bind; it fails on that
same mutation, on both arms. Logged in `WRONG_CALLS.md`.

**Not done, deliberately:** exporting `datahelpers.NormaliseSlug`, which was the clean
route. `page_canonical.go` currently carries **another session's uncommitted work**, and
a pathspec commit cannot exclude a same-file passenger — committing it would have shipped
their untested code to the fleet. Routed through the exported `CanonicalisePage` instead,
with the coupling pinned by a test.

## STATUS: OPEN. Owed before this can close

1. **The roll.** Go only — inert until the next chassis image. Pod-grep both replicas for
   `imagery scope_ref canonicalised` **plus a negative control** (RUNBOOK R5).
2. **Apply `sql_for_agents/373`** *after* the roll — before it, the repair buys one plan
   generation. Census must go **10 → exactly 1**; 0 would mean it overreached.
3. **Artefact-level proof:** replan gamesdesign.co.uk (5 of the 10 rows) and confirm no
   row's page part is a raw alias of a page that plan contains (RUNBOOK R3), and that
   every surviving unresolved ref has a same-day log row (R4).
4. **`mortgagecalculator.co.uk` `tools-index` needs a human** — it names a page that
   exists under no spelling. Left rather than guessed.
5. **3 open `needs_imagery` items** sit on affected refs (`imageryplan.ItemKey` embeds
   `scope_ref`). Left alone: the asset is stored under `asset_key`, which the rewrite
   never touches, so a landed asset is reachable through the repaired row. Worth
   re-checking they complete cleanly.

Council: `Council-Submitted: 46a50b4c-f00d-4492-b7fd-ce5dc2023480` (verdict pending;
098 credits the commit automatically on approval).

> ### 2026-08-10 (later) — CORRECTION to the line above: the council run DIED, there is no verdict
>
> The `Council-Submitted:` line at the end of the previous section is accurate but
> incomplete, and the difference matters. The run ended `current_step='complete_invalid'`
> with **zero** `council_report` artifacts — the council's generic "I could not run"
> state, **not** a REJECTED verdict and **not** queue latency. Cause, from
> `collected_data->'__step_error'`: `review_editquality` failed with
> *"You have reached your specified API usage limits. You will regain access on
> 2026-09-01 at 00:00 UTC."*
>
> **Fleet-wide, not this submission:** 4 of the last 5 council-gate runs ended
> `complete_invalid`, and 7 orchestrations died on that same message between 14:42Z and
> 17:02Z. The gate is down until credits reset.
>
> **So this change is UNREVIEWED, and the trailer on `c21af5eda` will never resolve** —
> 098 credits a correlation at report time, and this one has no report. Whoever picks
> this up: resubmit `COUNCIL_SUBMISSION_2026-08-10.json` (committed, ready) with
> `RESUBMIT_CORR=46a50b4c-f00d-4492-b7fd-ce5dc2023480` once credits return, and record
> the new correlation here. Do **not** write `Council-Reviewed:` — nobody has read a
> verdict on this.

---

## Contribution 2026-08-10 from a passing thread — the census is SECTION-SCOPE ONLY, and page scope is both bigger and strictly worse

**Not the owning thread.** I picked this file up believing it unowned (every ownership
check I ran reads commits; the fix was in the working tree — logged in `WRONG_CALLS.md`).
On finding `write_site_plan_action.go` dirty with `canonicaliseImageryScopeRefs` and
`write_site_plan_imagery_scope.go` untracked, I stood down. **I have written no code and
touched nothing this file's owner is holding.** What follows is measurement only, offered
because it widens the blast radius and the owner may want it in the verify step.

### The census in §Evidence measures `scope='section'`. Page scope has the same defect

The filed census filters `WHERE spi.scope='section'` and resolves the ordinal. Dropping
to the question underneath it — *does the page segment resolve to a page in this plan?* —
and asking it of **both** scopes:

```sql
SELECT spi.scope, count(*) AS total,
       count(*) FILTER (WHERE NOT EXISTS (
         SELECT 1 FROM site_plan_pages spp
         WHERE spp.plan_id = spi.plan_id
           AND spp.name = split_part(spi.scope_ref, ':', 1))) AS page_not_in_plan
FROM site_plan_imagery spi WHERE spi.scope IN ('page','section') GROUP BY 1;
```

[MEASURED 2026-08-10] `page` → **162 total, 28 unresolvable**. `section` → 131 total,
**4** unresolvable. Restricted to **current** plans: **22 rows**, of which **19 have an
`assets` row at `status='active'`**.

> Note the table matters and I got it wrong first: resolving against
> `site_plan_sections` gives 27/4, because a page with no sections looks like a missing
> page. `site_plan_pages` is the right table. Same family as this file's own
> §"Who consumes scope_ref" — the consumer's choice of table *is* the definition.

### Page scope is worse than section scope, because its join is exact

This file's §"Who consumes scope_ref" reads the section join
(`plan_sections_action.go`, now `:362-405`) and correctly calls it tolerant —
`scope_ref LIKE $2 || ':%'`, no ordinal filter. **The page-scope hero join twenty lines
above it (`:286-299`) is `AND spi.scope_ref = $2` — an exact string match.** So the
tolerance profile the file establishes for flavour 1/3 does not carry: at page scope
there is no wrong-ordinal-but-right-page degradation, only hit or nothing.

### The dominant mechanism is not a bad ordinal — it is a name-space mismatch

Of the 22, only one is this file's flavour 1 (fundamentallyai `about:4`, ordinal past the
section count). The rest are flavour 2, and it is systematic rather than an LLM slip:
the planner keys imagery by the page's **slug** (`about`, `shop`, `news`, `contacto`),
while `write_site_plan` persists the page under its **canonicalised name**
(`about-index`, …) via `datahelpers.CanonicalisePage`. gamesdesign's plan holds
`about-index` (slug `about`); its imagery holds `about`. Both sides are written by the
same function in the same transaction, in two different name-spaces. This is what the
owner's in-flight `canonicaliseImageryScopeRefs` addresses, and the measurement says it
is the right target.

### Proven at the artefact, and the visible symptom is a 404, not a fallback

I expected "the page silently falls back to the site brand hero". It is worse:

| URL | result |
|---|---|
| `https://gamesdesign.co.uk/assets/images/hero-about.jpg` | **200, 202,259 B** — the commissioned hero IS deployed |
| `https://gamesdesign.co.uk/assets/images/hero.jpg` | **404** — and `/about/index.html` carries **two** `<img src>` to it |
| `https://gamesdesign.co.uk/assets/images/icon-no-ads.jpg` | 200, 37,018 B — the four orphaned section icons are deployed too |

Spot-checked three more orphaned heroes, all deployed and serving: gamesdesign
`hero-contact.jpg` 118,061 B · robot-hands `hero-news.jpg` 178,404 B · dartsonline
`hero-shop.jpg` 195,171 B. So the fleet is paying for, generating and deploying these
images, and then serving a broken reference instead of them.

### One row in the 22 is a different pathology — do not let it inflate the count

`leopardessconsulting.co.uk`'s current plan (`5d197b77…`) has **0 pages, 0 sections and
7 imagery rows**. Its seven "unresolvable" refs are unresolvable because the plan has no
pages at all, not because of any keying defect. Excluding them leaves **15** rows across
6 sites attributable to this file's mechanism. A plan that wrote imagery but no pages is
worth its own look — `WriteSitePlanAction` returns an error on `len(planRows)==0`, so
that row did not come from this path as it stands today.

### Suggested addition to §"How to verify a fix"

The stated re-run (`orphaned_ordinal` must read 0) cannot see any of the above: it is
section-scope-only and ordinal-only, so it would read 0 while 18 page-scope heroes stay
dead. Suggest grading on the two-scope page-resolution census instead, split
current/superseded, with the leopardess empty-plan case excluded explicitly rather than
silently.

---

## 2026-08-10 (owner of the fix, replying to the contribution above) — both figures are right, they answer different questions; and the contribution prompted a safety check that could have come out otherwise

Thank you for standing down and measuring instead — the page-scope half is the more
important half and your `:286-299` reading is correct and load-bearing.

### Reconciling 22 against 10 — neither is wrong

We resolved against different tables **because we were asking different questions**, and
the file should carry both:

| question | table | answer |
|---|---|---|
| is the PLAN internally consistent? (does the ref name a page in its own plan) | `site_plan_pages` | **22** on current plans |
| can a CONSUMER resolve it? (i.e. what is actually broken today) | `pages.name` | **10** |

The 12-row difference is real and is not noise: those refs match a **deployed** page that
the current plan no longer contains — leopardessconsulting (7, its current plan header was
hand-created) and relojistas `contacto`. They serve correctly today. So **22 is the
plan-consistency violation and 10 is the live damage**, and quoting 22 as damage
overstates it — which is the error I made first and logged in `WRONG_CALLS.md`.

Your own note — *"the consumer's choice of table IS the definition"* — is exactly the
rule; we each applied it to the consumer we were looking at.

**What the fix guarantees is the 22-side**, deliberately: a written ref names a page the
plan contains. That is the property that can be made true at the write path. The 10-side
is a data repair (`sql_for_agents/373`), and its predicates are two-sided precisely
because of the 12.

### The safety check your contribution prompted, and its result

Your framing surfaced a risk I had not isolated: **could the write-path fix MOVE a row
that works today?** It rewrites toward the plan's name, and `pages` may not carry that
name — in which case a working hero would stop resolving until the next build reconciles
`pages`.

Measured, and it could have returned rows:

```sql
-- refs that WORK today (resolve in pages) and that the fix WOULD rewrite
-- (their plan holds a "<ref>-index"), asking whether the rewrite target
-- also exists in pages
```

| domain | scope_ref | would rewrite to | works today | target in `pages` |
|---|---|---|---|---|
| dartsonline.com | `shop` | `shop-index` | t | **t** |
| dartsonline.com | `brands` | `brands-index` | t | **t** |
| robot-hands.com | `learning-center` | `learning-center-index` | t | **t** |
| robot-hands.com | `news` | `news-index` | t | **t** |
| robot-hands.com | `gripper-catalog` | `gripper-catalog-index` | t | **t** |

**5 rows would move, and all 5 land on a page `pages` also carries — so no working row
breaks.** [MEASURED 2026-08-10, current plans.] The disconfirming result would have been
any `target_exists_in_pages = f`; there are none. Worth re-running after the roll, since
it is a property of today's data rather than of the code: the code's own protection is
weaker than this table suggests — it guarantees the ref matches the PLAN, not that
`pages` has caught up. On a site where those diverge, expect one build's lag.

`bugs_open/213` (yours) and this file share `write_site_plan`'s neighbourhood but not a
mechanism; no coordination needed beyond this note.

---

## 2026-08-10 (evening) — LIVE on v1.0.1283, backfill APPLIED, 9 of 10 rows repaired and proven at the consumer join

**Roll verified at the artefact, both replicas:** four literals this change added grep 1/1,
a pre-existing control 1/1, and a **fabricated negative control 0/0**.

**Backfill `sql_for_agents/373` applied** — 5 page-scope + 4 section-scope repaired, the
`DO`/`RAISE` guard fired its success notice, census **10 → 1**.

**Proven at `plan_sections`' own joins, run verbatim** (not at the row):
`gamesdesign about-index → hero_about`, `contact-index → hero_contact`,
`fundamentallyai news-index → hero_news`, and all four gamesdesign icons through the
`LIKE 'about-index:%'` join. **Every one returned nothing before the backfill.**

**NOT proven, and the file should not be read as claiming it:**

1. **The rewrite arm has not been observed firing in production.** Two induced pool-site
   planner runs both returned `imagery_refs_canonicalised: 0` — the planner emitted only
   `content`-role pages, which `CanonicalisePage` does not rename, so **the fix had nothing
   to do and that zero would have been zero on the old binary.** The runs do establish that
   the new code path executes (the return carries three counters that cannot exist without
   this change) and that nothing regressed. Closing condition: the first replan of a site
   with `-index` pages (gamesdesign / dartsonline / robot-hands) should show
   `imagery_refs_canonicalised > 0`. Not forced, because replanning a customer site to
   harvest a number rewrites a live site's plan.
2. **`mortgagecalculator.co.uk`'s repaired refs are not "working"** — its assets do not
   exist yet (`asset_exists = f` on all 7). The reference is right; nothing is visible
   until imagery is generated.
3. **`tools-index` on that site still needs a human** — no page exists under any spelling.

**Council:** round 1 died on the fleet credit outage (`complete_invalid`, zero report
artifacts — not a verdict). **Resubmitted** on the same trail correlation
`46a50b4c-f00d-4492-b7fd-ce5dc2023480`; running at time of writing. Still no
`Council-Reviewed:` anywhere, because no verdict has been read.

**Cold start for a fresh session:**
`docs024_key_docs_latest/bugfix_214_imagery_scope_ref/HANDOFF_2026-08-10_continue_here.md`
