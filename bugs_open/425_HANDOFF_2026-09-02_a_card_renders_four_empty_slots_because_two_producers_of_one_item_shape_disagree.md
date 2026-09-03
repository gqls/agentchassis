# 425 — a card renders four empty slots because two producers of one item shape disagree

**Filed 2026-09-02** (components lane, from the boxingonline.com owner review).
**Status: root cause FIXED; the component half is LIVE, the Go half is INERT until the next roll.**
Severity: low as a crash, high as a class — it is the shape in which a data gap gets
reported as a design fault, and it has been mis-reported at least once already.

## What the owner saw

> "The 'Latest from the Ring' section on the home page has cards that need better designs.
> They have images which is better but the cards could look a lot better."
> — boxingonline.com, 2026-09-02

The cards do not have a design problem. Served markup, first of six on `/index.html`,
verbatim:

```html
<article class="article-card hover-lift">
  <div class="article-card__image">
    <img src="/assets/images/card-cruiserweight-boxings-best-kept-secret.jpg"
         alt="Cruiserweight Is Boxing's Best-Kept Secret — And It Won't Stay That Way | Boxing Online">
    <span class="article-card__category"></span>            <!-- EMPTY -->
  </div>
  <div class="article-card__content">
    <h3 class="article-card__title"><a href="…">Cruiserweight … | Boxing Online</a></h3>
    <p class="article-card__excerpt"></p>                   <!-- EMPTY -->
    <div class="article-card__meta">
      <span class="article-card__date"></span>              <!-- EMPTY -->
      <span class="article-card__read-time"></span>         <!-- EMPTY -->
    </div>
  </div>
</article>
```

Four of six content slots empty, on all six cards, plus a headline that is the page's raw
document `<title>` — site-name suffix and all — on a page whose header already says Boxing
Online. So the card is an image, one over-long headline, and a run of empty elements that
still take up layout. That reads as unfinished, and "needs a better design" is a reasonable
thing for a person to say about it.

## Root cause — one sentence

**Two producers write the same standard list-item shape for the same shared component and
disagree about it**, and the component renders every per-item slot unguarded, so whichever
keys the producer did not write become empty elements that still occupy layout.

| | `scanBlogArticles` (`rebuild_blog_listing_action.go`) | `resolvePagesWhereType` (`queryresolve.go`) |
|---|---|---|
| feeds | the blog-index listing | `query.blog_posts` → the home page's `content-listing` |
| `title` | site-name suffix **stripped** | `COALESCE(p.title, p.name)` — **raw** |
| deck | `excerpt`, from `meta_description` | *(none)* — writes `meta_description` under that name |
| `date` / `read_time` / `category` | written | *(none)* |

The home page is fed by the right-hand column. Its stored `content_data`, verbatim
(`page_components` `7f3b6cae-f118-4d84-80a1-97b75357b9c7`):

```json
{ "url": "/blog/cruiserweight-boxings-best-kept-secret.html",
  "name": "cruiserweight-boxings-best-kept-secret",
  "image": "/assets/images/card-cruiserweight-boxings-best-kept-secret.jpg",
  "title": "Cruiserweight Is Boxing's Best-Kept Secret — And It Won't Stay That Way | Boxing Online",
  "nav_label": "Cruiserweight's Rise",
  "meta_description": "Discover why cruiserweight boxing sits between light heavyweight and heavyweight, and why the division's best fighters are finally getting the spotlight." }
```

**The deck the empty `<p>` wants is PRESENT in that row, under a different key.** The
template reads `.excerpt`; the resolver writes `meta_description`; nothing joins them up.
That is the whole of defect 2 for the excerpt slot — not missing data, a missing agreement.

And `title` was never meant to carry the suffix: `resolvePagesWhereType`'s own doc-comment
has always documented the shape as `"title": "Jump Physics"`, unsuffixed. The code returned
the document title raw. This is the projection failing the contract it publishes.

## Why nothing caught it

`bugs_closed/054` ("unguarded `{{range .items}}` in list templates, no empty state") left a
standing lint, `scripts/check_list_empty_states.py`. Its predicate:

```sql
WHERE is_active AND html_template LIKE '%{{range .items}}%'
```

A **literal**. Components range over whatever the schema field is called, and
[MEASURED 2026-09-02] `.items` is not even the commonest spelling — `.entries` 13, `.items`
8, `.cards` / `.features` / `.products` 3 each, then `.articles`, `.categories`,
`.testimonials`, `.periods`, `.rows`, … So a component was checked or not **according to
what its author named a variable**, and `content-listing` had never been looked at once.

The check was not reporting green, which is worse than it sounds: it reported **1 unguarded
of the 8 it could see** and exited 1, so it looked like a working check with a small known
backlog. Widened to derive the collection name per template, it reports **29 unguarded
`{{range}}` blocks of 72, across 55 active components**. Nothing regressed — the other 28
were always there.

And even for the 8 it did see, 054's check asks whether the **array** can be empty. This
bug is about a **field inside a present item**. No detector existed for that class at all.

## What is fixed, and what is not

| half | state |
|---|---|
| `content-listing` renders no empty slot | **LIVE** — migration `682`, applied + ledgered 2026-09-02, config so live on apply |
| both producers share one title/deck rule | **IN THE BINARY, NOT EFFECTIVE ON THE RERENDER PATH** — see the negative result below. "Live" is true and misleading |
| a detector for this class exists | **LIVE** — `scripts/check_card_slot_guards.py` |
| 054's lint can see the whole library | **LIVE** — `scripts/check_list_empty_states.py` widened |

> ⚠ **THE SERVED PAGES HAVE NOT CHANGED YET, and neither half changes them on its own.**
> `page_components.rendered_html` holds what was rendered when it was rendered. Migration
> 682 changes what the NEXT render produces; the producer fix needs an image roll before a
> re-render can put a deck in the row. **Verify at the served page after both, not before,
> and do not read a still-empty card as the fix having failed.**

## Propagation — the step this bug OWES, and the exact shape it has to take

> Added 2026-09-02 after the council returned **REVISE** on correlation `84b51f16`, with
> matching **high-severity** objections from `render_guardian` and `debug_historian`. Both are
> right, and the second names a landmine I had not grepped for before writing the migration:
> *"A template edited by SQL ships NOTHING — the `template_changed` fan-out lives in
> `component-template-fixer`."* Saying "the next render picks it up" in the prose is not a
> propagation step, and a bug whose stated symptom is a served page is not fixed until the
> served page changes.

**The rule, from `LANDMINES.md`:** `page-rerender` branches on `spec.reason` alone
(`check_rerender_mode`). Only `image_landed | section_data_resolved | cta_links_stale |
template_changed | literal_markdown` route to `rerender_page_sections`, which **re-runs the
`query.*` resolvers**. *Anything else, including no reason at all*, routes to
`rerender_single_page` — "simple concatenation", which re-ships the stored `articles` array
byte for byte. So a re-render can complete, stamp a new `deployed_at`, and change nothing.

> **REWRITTEN 2026-09-02 after council round 2.** The first version of this section hand-rolled
> a fan-out. Three seats (`reuse_agent`, `prior_art_librarian`, `guardian`) objected — correctly —
> that I had *quoted the landmine naming the owner and then bypassed it*. The owner is
> `component-template-fixer`, step `create_rerender`, and **copying its SQL exposed three real
> defects in mine**: it excludes OWNED pages (`rebuild_policy IS DISTINCT FROM 'owned'`,
> `bugs_open/301`) where mine would have filed rerenders the save path refuses; it dedups with
> `NOT EXISTS` on an open item rather than `ON CONFLICT DO NOTHING`, which is the documented
> convention for `site_work_items`; and it spells the filter `p.status = 'active'`.
>
> **That last one was a dead literal in my version and `LANDMINES.md` already says so — three
> separate entries.** `[MEASURED 2026-09-02]` `pages.status` holds exactly two values fleet-wide,
> `active` 1,106 and `archived` 90; **`'deployed'` never occurs**, because it belongs to
> `pages.build_status` (1,011) and to `sites.status` (33). The two tables have **inverted** live-
> state spellings — on `pages` the live value is `active`, on `sites` it is `deployed`, where
> `active` is 1 row of 57 — and `IN ('active','deployed')` is the defensive union that is
> accidentally right on both, which is exactly why it spreads. It is carried by ~20 call sites
> today, including `resolvePagesWhereType` itself. Harmless as a union; the danger is the reverse
> direction, where a single-value predicate silently selects nothing.

So the propagation is a **hand-run instance of `component-template-fixer`'s own step** — same
SQL, same reason, same exclusions:

```
page-rerender   spec.reason = 'template_changed'   +  page_id, page_name, domain
                WHERE  rebuild_policy IS DISTINCT FROM 'owned'
                  AND  p.status = 'active'
                  AND  NOT EXISTS (an open template_changed rerender for this page)
```

**Using the owner's reason string is load-bearing, not cosmetic:** its dedup filters on
`spec->>'reason' = 'template_changed'`, so a row filed under any other reason is *invisible to
it*, and the next real template fix would file a second, duplicate rerender for the same page.

**Why a file at all, then:** that step is triggered by `fix_result.component_id` — a template fix
made *by the agent*. A template edited by SQL never reaches it, which is precisely what the
landmine says. If the fixer ever gains a trigger for SQL-applied edits, **delete `683` rather
than maintaining both.**

**And the scoped path really does re-resolve — cited now, not assumed.** `render_guardian`
objected that the plan leaned on a landmine about *assemble* mode and then assumed the opposite
of scoped mode with no citation; if scoped mode also rendered from stored `content_data`, `683`
would ship the new template against the OLD articles array on all 14 pages. It does not:
`rerender_page_sections_action.go`'s header states *"newSourceResolver + planSection … rebuild
each section's resolved_data (**queryresolve for query.\***…)"* and *"resolved_data merged last
so it wins"*, and the call is `newSourceResolver(siteID, params.DB, logger, pageName)`.
`articles` is `source: query.blog_posts` — a resolved field — so it is rebuilt fresh and wins.

**It must be sequenced AFTER the chassis roll.** Firing it today re-resolves against the
OLD resolver in the running pod: the guarded slots would collapse (682 is live), but no deck
would arrive, and the render would have been spent. `bugs_open/384` / **PBP-048** built
`requestPageListReresolve` for producer-driven invalidation of exactly this shape; a code fix
is not one of its trigger events, so the post-roll pass is a deliberate, dispatched one.

**And the verify block's counts are computed by DIFFERENT predicates.** The first version derived
`targets` from the INSERT's own `WHERE`, so a wrong predicate made `targets` and `filed` agree
*at zero* and the guard could never fire — a check that cannot detect its own predicate being
wrong (`debug_historian`). It now counts `carrying` with no status/ownership filter beside
`eligible` with them, and asserts `filed = eligible - already_queued` **exactly**, because
`ON CONFLICT`/`NOT EXISTS` can drop a *subset* and an un-filed page keeps the stale deck for ever
with nothing surfaced (`bug_historian`; 016b §9's *"`err == nil` is not work queued"*). All three
induced in rolled-back transactions: the dead status literal aborts with *"14 page(s) carry the
component but 0 are eligible"*; a dropped subset with *"filed 11 but expected 14"*; an
assemble-mode reason aborts too.

**Two checks before dispatching, both already run:**

1. **Will the section-shrink guard refuse it?** `LANDMINES.md` records this component family as
   the guard's classic trigger — moving a list onto the collection dialect loses the flat
   dialect's per-card LLM fields, visible text drops below the 50% floor, and the refusal's own
   error text invites you to lower a fleet-wide guard to land one page. **It will not fire
   here.** `[MEASURED 2026-09-02]` across all six sites carrying `content-listing`, pages with
   an empty `meta_description`: boxingonline **0 of 7** (mean deck 123 chars), dartsonline
   **0 of 23**, garden-tools **0 of 5**, homegarden **0 of 4**, idea.uk **0 of 9**, robot-hands
   **0 of 8**. Every card can be filled on every site, and the change **adds** visible text
   (a ~123-character deck per card) while removing only the suffix. The direction is the
   opposite of a shrink.
2. **Will the section branch escalate the whole page to the writer?** It does so if any section
   lacks a required `source:"llm"` field (STY-048). Check per page before dispatching.

**And when verifying: a COMPLETED `page_rerender` row is not evidence.** Read
`spec->>'reason'` on it — a row with no reason took assemble mode and structurally cannot have
picked this up, however fresh the page looks. That is `bugs_open/384`'s own filing error,

> **⚠⚠ AND THE REASON CHECK IS NOT SUFFICIENT EITHER — corrected 2026-09-02, after this guidance
> had been repeated to two lanes all day.** Raised by the boxingonline lane, and it is the sharper
> half of the rule. `[MEASURED 2026-09-02]` boxingonline `/index.html` completed **twice**
> (17:26:34 and 17:32:46), **both carrying `reason='template_changed'`** — the correct,
> non-assemble reason — and **both produced the old item shape**. So a reader can do exactly what
> this paragraph tells them, find the right reason on a completed row, and conclude the change
> landed when it did not.
>
> **A correct reason is NECESSARY and demonstrably NOT SUFFICIENT.** It proves which branch was
> *taken*; it says nothing about what that branch *produced*. The check that discriminates is on
> the **artefact**: does `content_data->'articles'->0` carry the `excerpt` key that the fixed
> projection writes unconditionally? Verify there, and use the reason only to explain a negative.

which it corrected an hour later.

### The Go half went LIVE 2026-09-02 12:28 UTC — verified at the artefact, with controls

Both `agent-chassis` pods restarted at 12:28:03 and 12:28:24 UTC (a fleet roll, 0 restarts on
the new ReplicaSet). The startup `build provenance` line had already scrolled out of range by the
time I looked — **an absent grep there means "not in range", not "unstamped"** — so the binary
probe is the honest instrument, and it needs controls in both directions or it proves nothing:

| probed symbol | result | why it is in the probe |
|---|---|---|
| `ListItemExcerpt` | **PRESENT** | created by this fix |
| `ListItemTitle` | **PRESENT** | created by this fix |
| `resolvePagesWhereType` | PRESENT | pre-existing **positive control** — a probe that found nothing would be indistinguishable from a broken probe |
| `ListItemTitleXYZNOTREAL` | absent | invented **negative control** — and it *contains* `ListItemTitle` as a substring, so its absence also proves the grep is not matching loosely |

So the producer fix is running. `[MEASURED 2026-09-02]` — and note this probes the **capability**,
not the commit, which is the stronger question: it answers "is the code there" without depending
on a stamp that may be out of log range.

⚠ **THE ROLL ALSO KILLED THE COUNCIL ROUND MID-FLIGHT**, and that is a documented trap rather than
bad luck: a deploy restarts the chassis and every in-flight orchestration dies where it stood.
Three runs froze within 40 seconds of each other at 12:27–12:28 — mine at `review_render_guardian`,
two other lanes' at `review_prior_art` — and all three sat in `EXECUTING_STEP` for over an hour
looking exactly like a slow seat. **The tell is the correlation between runs, not the duration:**
one stuck run is latency, three stuck at the same minute is an event. `kubectl get pods -l
app=agent-chassis -o custom-columns=...startTime` settles it in one command.

### ⚠ THE PROPAGATION RAN AND DID NOT DELIVER THE PRODUCER FIX — verified negative, 2026-09-02

`683` was applied across all six sites on the owner's explicit instruction. The batch drained:
**10 complete, 4 cancelled.** And the result is a finding, not a pass.

**What DID work:** the empty slots are gone. On every completed page the four unguarded card
slots now collapse instead of rendering blank elements that occupy layout. Migration `682` does
exactly what it was written to do.

**What did NOT work:** the headlines still carry the site-name suffix and no deck arrived. The
freshly-written `content_data` — written *by* the 13:59 rerender, not left over from before —
carries the OLD projection shape: no `excerpt` key, `title` with the suffix. So the guard is
correctly collapsing a slot whose data never arrived.

**Each alternative was eliminated at an artefact** (the `site_delivery_and_editor` lane verified
boxingonline independently; I verified garden-tools and homegarden):

| candidate | eliminated by |
|---|---|
| mirror lag / stale object | served object last-modified **14:05**, newer than `deployed_at` 13:59 — a fresh publish |
| the template never updated | `content_components` row updated **10:43:57**, guard present |
| the page pins an old component version | both pages pin `component_version_id` `1454705a` created **11:11:55**, *after* the update, and that version's template carries the guard |
| the pinned version declares a different source | **[MEASURED 2026-09-02]** pinned version and live component both declare `articles` ← `query.blog_posts` on all four pages checked |
| wrong rerender mode | `spec.reason = 'template_changed'` on every completed item — the scoped branch, not assemble |
| the binary lacks the fix | `ListItemTitle` / `ListItemExcerpt` **PRESENT** in the running binary, with positive and negative controls |

**So the code is in the binary and the executing path does not produce its output.** That is the
purest form of the estate's own lesson: *probe the CAPABILITY, not the commit*. Symbols present
verified the roll; it did not verify the path.

The structural reading only deepens it: `rerender_page_sections_action.go:1438` calls
`planSection(…, comp, resolver, …)`, `plan_sections_action.go:2695-2709` resolves any `query.*`
source through `queryresolve.Resolve` **unconditionally** (no "only if missing" gate), and
`:1617-1618` merges `plan.ResolvedData` into the render context **last, so it wins**. Every step
of that says the fresh array should reach the page. It does not.

> **CORRECTED 2026-09-02, before it cost a diagnosis round.** I wrote above that garden-tools'
> stored array holds four guides "while `query.blog_posts` would return five blog posts for that
> site" — and the `site_delivery_and_editor` lane rightly proposed reading the rendered cards as a
> discriminating test, since that would be the only page where stored and resolved differ.
> **The premise was mine and it was wrong.** My count of five came from a query that did not
> replicate the floor the resolver actually applies. `blog_posts` calls `resolvePagesWhereType`
> with `listedOnly=true` → `ListedPageEligibilitySQL`: `deployed_at IS NOT NULL` **and**
> `jsonb_array_length(p.sections) > 0`. The fifth page, `/blog/buying-guide-post.html`, has
> `deployed_at` NULL, zero sections and `build_status='planned'` — it fails both clauses. The
> resolver returns **the same four**. So garden-tools does not discriminate either, and the
> rendered output showing four guides is consistent with both hypotheses.
>
> A census that does not replicate the predicate the code uses answers a different question, and
> both of us had quoted the looser number.

**THE DISCRIMINATOR THAT DOES WORK, and it is about a KEY rather than a value.** The fixed
projection writes `"excerpt": ListItemExcerpt(metaDesc)` into the item map **unconditionally** —
the key is present whatever the value, including empty string. The old projection never wrote it
at all. `[MEASURED 2026-09-02]` the items written by these rerenders have **no `excerpt` key**
(`content_data->'articles'->0 ? 'excerpt'` is false on every page checked).

So this is not "the deck resolved empty". **Any item produced by the fixed code carries the key;
these carry none, therefore they were not produced by the fixed code** — while the binary running
them demonstrably contains the symbols. That narrows it to the resolution not executing on this
path, or executing through something that is not `resolvePagesWhereType`, and it rules out "it ran
and the data was thin".

**Filed for diagnosis rather than guessed at: correlation `c19a975d-b32c-4ed8-825a-e8d6100bbec7`.**
The cheap eliminations above are done; what remains is inside the execution, which is exactly
what the loop is for.

### ANSWERED: the producer fix does NOT execute on the RERENDER path — a controlled A/B on one binary

`[MEASURED 2026-09-02]` The two paths were run against **the same site, the same component, the
same binary (v1.0.1354), three minutes apart**:

| path | dispatched as | result on `content_data.articles[0]` |
|---|---|---|
| **build** — `needs_page` rebuild (`7f1f4993`, 17:23:02Z) | delivery lane's own fix | **`excerpt` key PRESENT, title suffix STRIPPED** |
| **rerender** — `page_rerender`, `reason='template_changed'` (`684`, 17:26:52Z) | this lane, on boxingonline `/index.html` | **no `excerpt` key, title suffix INTACT** |

**Controls, all clean, because this is the claim the whole day has been failing to establish:**

- the rerender **wrote** — `page_components.updated_at` 17:26:19 **and 8 `page_component_history`
  rows at 17:26:19** for that page (keyed on `page_id`, the `NOT NULL` column). Not a refused save.
- **zero** `SECTION COMPONENT FLOOR REFUSED` entries fleet-wide in the surrounding 20 minutes.
- the slot is still bound to the **canonical** `content-listing`, not a per-site fork.
- the rendered markup carries **no** `article-card__category` — the post-682 template ran.

**So the stale-pod hypothesis is dead**, and with it my own one-pod-of-two hole: the same binary
produces the fixed shape via one path and the old shape via the other, minutes apart. The
difference is the **path**, not the image, not the timing, not the template.

**This is exactly what the council's `render_guardian` seat objected to in round 2** — that the
plan cited reason-branching and never showed the scoped path re-invoking the projection. The
citation I answered with was real and describes `planSection` calling `queryresolve.Resolve`
unconditionally; the behaviour disagrees with it, and the behaviour is what ships.

**The open question is now small, concrete, and well-posed:** the rerender path builds its
`articles` array through something that is not the fixed `resolvePagesWhereType`. The delivery
lane's hypothesis — the fix landed in the wrong producer *for this path* — is the surviving one.

### Reproducible, and four branches eliminated by reading them

`[MEASURED 2026-09-02]` **Three** rerenders of boxingonline `/index.html` — 13:59, 17:26, 17:32 —
all wrote (fresh `updated_at`, archive rows), all produced **6 items in the old shape**: no
`excerpt` key, title suffixed. Two of the three ran on the post-roll binary that demonstrably
carries the fix. It reproduces on demand.

Because `mergedContent` is `stored ⊕ plan.ResolvedData` with **ResolvedData last**, an unchanged
stored array means the key was **missing from `ResolvedData`** — not that it lost a merge. Four
branches that would produce that absence, each eliminated at the code or the artefact:

| branch | why it is not this |
|---|---|
| the literal-markdown **strip** (`rerender_page_sections_action.go` ~:1600) | gated on `spec.reason == 'literal_markdown'`; the dispatch used `template_changed`. It also mutates values, never deletes keys |
| **`queryListBelowContract`** → `handleMissingField()` → key never written | the base returns **6** items for this site under its own `listedOnly` floor, and the field declares **no `min_items`** |
| **whole-page escalation** to the content writer (STY-048) | **0** `needs_page` items created |
| **`save_page_sections` refusing** | the row was written *and* archived at the rerender minute; **0** floor refusals fleet-wide in the window |

**An instrumented run did not settle it**, and the reason is worth recording: live tails on both
pods during a rerender captured no `queryresolve:` or `plan_sections:` message at all — but the
pods serve every agent, so the surrounding lines cannot be attributed to this run without
correlation filtering, and **I have no positive control that the resolver's line would appear if it
did run.** An absence with no control is the day's recurring mistake, so it is recorded as
inconclusive rather than as evidence the resolver never fired.

**Filed with these eliminations as `c755b0be-8035-4108-bf24-5b216ca327a5`** — a far better-posed
symptom than the two runs before it, both of which failed partly because the symptom was vague
about which branches were already ruled out.

### SETTLED AT THE MARKUP — read this section first; the three readings below it are superseded

`[MEASURED 2026-09-02]` The pre-682 and post-682 templates leave **different fingerprints**, and
testing for them splits the batch exactly along its status line:

| | `article-card__category` | `article-card__meta` | `article-card__excerpt` | reading |
|---|---|---|---|---|
| **4 cancelled** | present | present | present | **pre-682 markup, untouched** |
| **10 complete** | absent | absent | absent | **post-682 markup — the slots collapsed** |

**So migration 682 IS delivered, on 10 of the 14 pages.** The empty card slots are gone. The four
cancelled pages kept their pre-682 markup because the component floor refused the save outright,
exactly as diagnosis `c19a975d` confirmed.

**And the producer half is the part that did not run**: no `excerpt` key on any of the 14, and
every title still suffixed. So the honest split is:

- **template half → LANDED** (10 of 14; the other 4 blocked by the floor, which the same flattening
  triggers)
- **producer half → DID NOT EXECUTE** (categorical, by the key-presence test)

> **AND THE INSTRUMENT WAS NEVER BROKEN — my QUERY was. Fifth revision, same habit.**
> `[MEASURED 2026-09-02]` `page_component_history.component_id` is **NULL on 44,555 of 45,285
> rows (98.4%)**. My "zero rows for this component, ever" filtered on that column, so it returned
> zero for a reason that has nothing to do with archiving. Keyed on `page_id` instead, the archive
> holds rows for **all 10 completed pages at exactly the rerender minutes** (13:51–14:14) and
> **none for the 4 cancelled** — which is precisely the split the markup fingerprint shows, from
> an independent instrument.
>
> So the archive corroborates rather than contradicts: content changed on the ten, did not on the
> four. **Two instruments now agree, and I had declared the working one broken and told a peer to
> strike it.** Before filtering on a column, check the column is populated — a filter on a
> mostly-NULL column returns zero and reads exactly like "no such rows".
>
> **The paragraph below is kept as written, because its reasoning was sound and its premise was
> not.** The rule it states is right; it simply was not the rule that applied here.

> **WHY I GOT THIS WRONG TWICE, and the instrument is the lesson.** I reported "nothing wrote
> content on any of the 14" on the strength of `page_component_history` holding zero rows for this
> component. I ran a positive control and it passed — 389 rows across 5 other components in the
> same window — and **the control was the wrong one.** It proved the TABLE works; it did not prove
> the table works **for this component**. `[MEASURED 2026-09-02]` content-listing has **zero
> history rows EVER**, against 45,285 in the table overall. Its writes bypass the archive
> triggers, so its zero carries no information at all, and I built two corrections on it.
>
> **A positive control must exercise the same row population as the claim.** A control drawn from
> the neighbours answers a question about the neighbours. This is the day's fourth turned-over
> claim on this bug and the sharpest instance of the same habit: I keep reaching for a column
> rather than the artefact, and the artefact — the stored markup — settled it in one query.

### The "it rendered before the roll" explanation — checked and it does not hold

Offered by the `site_delivery_and_editor` lane, and worth checking rather than adopting because it
would have dissolved the whole thing. **There were TWO rolls today**, and the proposal reads the
second as the first:

| | |
|---|---|
| `f57f5ad1f` committed | **10:51:55 UTC** (`2026-09-02T11:51:55+01:00` — the `+01:00` matters) |
| roll 1, ReplicaSet `96c48f448` | **12:28:03 / 12:28:24 UTC** — after the commit |
| binary probed on a roll-1 pod | **~13:39 UTC**: `ListItemTitle` / `ListItemExcerpt` present, `resolvePagesWhereType` positive control, invented symbol absent |
| the 14 rerenders | **13:51–14:14 UTC**, on roll-1 pods |
| roll 2, ReplicaSet `744cfb4bf` | **15:39 / 15:53 UTC** — after everything in this batch |

So the fix was in the running binary before the rerenders, and the 15:39 roll is not the one that
shipped it. Both current pods re-probed with the same controls: fix present.

> **⚠ THE ONE REAL HOLE, and it is mine.** I probed **one pod of two** at 13:39, and this estate's
> own landmine says a same-tag rebuild can serve a node's **cached** binary — so two pods of one
> ReplicaSet are not *guaranteed* identical, and the un-probed pod is now gone. If the rerenders
> were handled by that one, my evidence does not cover them. Recording it as a hole rather than
> arguing past it.

**Follow-up diagnosis `fe4b8537` returned NOT CONFIRMED** (iteration cap, no fix proposed), with
both its data requests returning 0 rows.

> **CORRECTED — I explained that with a guess and the timestamps refute it.** I wrote that the run
> "was chasing a target that moved under it", because a target page had been repointed to a
> per-site fork. `[MEASURED 2026-09-02]` the diagnosis item completed **14:36:43 UTC**; the fork
> `content-listing-guides-boxingonline-com` was created **17:14:24 UTC** — **two hours and
> thirty-eight minutes later.** The repoint cannot have touched the run. Caught by the delivery
> lane, and it took two `created_at` reads to check something I had asserted from plausibility.
>
> **The 0-row data requests remain unexplained.** One was behind an `EXPLAIN` that errored with
> *"syntax error at or near $"* — a parameterised query handed to `EXPLAIN` unbound — which is a
> concrete lead and not the same as the story I invented. Still not worth re-filing in that shape.

**What supersedes both:** a post-roll rebuild dispatched by the delivery lane (item `7f1f4993`) on
the current binary. It answers the producer question at the artefact rather than by inference about
images, which is the only kind of answer this bug has responded to.

### An instrument that is NOT available here: the chassis log

Worth stating so the next reader does not spend the attempt. `queryresolve` logs
`"queryresolve: resolved pages_where_type"` with a count whenever it runs, so "did the resolver
execute during the rerender?" looks like a log question. It is not answerable retrospectively on
this service.

`[MEASURED 2026-09-02]` `kubectl logs --since=3h` over both `agent-chassis` pods returned **318
lines spanning 14:28:27 to 14:28:47 — twenty seconds.** The rerenders ran 13:51–14:14, entirely
outside the retained window. A grep for the resolver's line returns **0**, and that zero says
nothing whatever about whether it ran.

**The control is what saved this from being a sixth wrong claim**, and it is two lines: print the
earliest and latest timestamp actually captured, and probe for a line that must be present
(`orchestration` → 291, `save_page_sections` → 26 in the same corpus, so the capture is real and
merely narrow). CLAUDE.md warns that the `build provenance` startup line "scrolls" on a busy
service; on this one the readable window is **seconds**, not hours.

### DIAGNOSIS `c19a975d` CAME BACK **CONFIRMED** — and it explains 4 of the 14, not all of them

**The confirmed mechanism:** `save_page_sections`'s SECTION COMPONENT FLOOR aborts the *whole
save* — its own message says **"Nothing was written"** — when the freshly rendered slot carries
fewer than half the layout-class-bearing elements the stored version had. Quoted from
`agent_error_log`: *"content-listing 114→54 class attributes (47% kept, floor 50%)"*. So the
resolved array never reaches the merge step at all: it is **computed with the fix and discarded**,
and the row on disk stays whatever the last successful save produced.

That is a better mechanism than the one I hypothesised. I was looking for a merge that stored data
won; there is no merge, because there is no write.

> **CORRECTED 2026-09-02, third revision of this section — the "10 wrote, 4 did not" split below
> is WRONG, and the DB's own archive says so.** `page_components` carries archive triggers that
> fire when `content_data` changes (`trg_page_component_content_archive_upd`) or `rendered_html`
> changes (`trg_page_component_artefact_archive_upd`), writing to `page_component_history`.
> `[MEASURED 2026-09-02]` for component `aa3e4b68` (content-listing) in the last four hours:
> **ZERO archive rows.** Positive control, same window, same table: **389 rows across 5 other
> components**, including at 14:13, 14:14 and 14:15 — the exact minutes two of these very pages'
> items completed. So the archive was working and content-listing simply never changed.
>
> **Neither `content_data` nor `rendered_html` changed on ANY of the 14.** The fresh
> `page_components.updated_at` on the ten reflects a write that touched neither content column —
> metadata only. I read a moving `updated_at` as "something wrote content", and it is not that.
> There is no `updated_at` trigger on this table (checked), so the column is set by whatever
> `UPDATE` the code issues, which need not touch content at all.
>
> **And that sharpens the open question rather than closing it.** Other components on the *same
> pages* DID archive in that window (the `delete`/insert pairs at 14:13–14:15), so the page-level
> save was not wholly refused there — only content-listing's slot produced no change. A slot that
> re-renders to a **byte-identical** result writes nothing, because the trigger's condition is
> `IS DISTINCT FROM`. So the live question is now much narrower than "what wrote pre-fix data":
> **why does this slot re-render byte-identically on a binary whose projection would change its
> title?**

> **QUALIFIED, same day, by the `site_delivery_and_editor` lane — and the qualification is right.**
> A zero in the archive cannot by itself distinguish *"nothing rebuilt the array"* from *"the old
> projection rebuilt it identically"*, because the trigger tests `IS DISTINCT FROM`: the same code
> over the same inputs produces a byte-identical result and writes nothing. "Byte-identical" and
> "the old projection ran" are mutually consistent, so the archive narrows the question without
> settling it.
>
> **But the four refusals are a lever on exactly that**, and they point away from "the old template
> ran". The floor refusal quoted by the diagnosis — *"content-listing 114→54 class attributes"* —
> is proof that on those pages the render produced a **materially flattened** output: collapsing
> the guarded slots is precisely what removes class-bearing elements. So the post-682 template,
> when it renders, demonstrably yields a different result.
>
> On the ten, that did not happen: no flattening, no refusal, no archive row. Had the post-682
> template rendered there, it would have collapsed the same empty slots, changed the class count,
> and either archived or tripped the same floor. **So on the ten, either the pre-682 template
> rendered, or nothing rendered that slot at all** — and the pinned component version is not an
> explanation, since `1454705a` was created at 11:11:55, after the 10:43:57 template update, and
> carries the guard.
>
> That is the shape the follow-up diagnosis (`fe4b8537-4833-4de2-9d2e-2141619a911c`) is pointed at.

> **The superseded reading is kept below deliberately, because the correction is the point.**
> `[MEASURED 2026-09-02]` across all 14 target pages: **every one** has no `excerpt` key and a
> suffixed title. The **4 cancelled** pages carry stale `page_components.updated_at` (03:10, 22:36,
> 22:37 — from before this work), which is exactly the "nothing was written" signature the
> diagnosis describes. The **10 completed** pages carry *fresh* `updated_at` (13:51–14:14) and the
> **old shape anyway**. Something wrote, and what it wrote was pre-fix.
>
> `agent_error_log` holds floor refusals for the four cancelled pages and **none** for the ten. So
> the confirmed mechanism is real and is not the whole story: it explains the pages that never
> wrote, and the pages that *did* write with stale content remain unexplained.

**What the key-presence discriminator still tells us, and it holds across all 14:** the fixed
projection writes `"excerpt"` unconditionally, so its absence is categorical — an absent map key
cannot be thin data, only un-executed code (or, per the diagnosis, discarded code). For the four,
"discarded" is now proven. For the ten, neither is yet.

**Open, and it is the live question:** what wrote fresh `content_data` on ten pages between 13:51
and 14:14 carrying the pre-fix item shape, on a binary that contains the fix and via a path whose
own header says the resolver rebuilds `query.*` fields and merges them last?

### ⚠ A SECOND GUARD REFUSED 4 OF 14 — and its error names an escape hatch that must not be used

Four items (`idea.uk` guides-index, `dartsonline.com` guides-index, `robot-hands.com`
learning-center-hub, and one more) were refused by the **section COMPONENT floor**
(`bugs_open/253`) — a sibling of the text-based shrink guard, counting **class attributes** rather
than visible text:

> *"SECTION COMPONENT FLOOR REFUSED for page "guides-index" — content-listing 69→34 class
> attributes (49% kept, floor 50%)."*

**This fix is what trips it, by design.** Every empty element it collapses carries layout classes,
so a page whose cards are mostly unfed flattens hardest. A pass at 51% and a refusal at 49% are
the same change — the ratio is a property of how much data the site happens to have.

> **NARROWED 2026-09-03 — "fleet-wide" was too strong, and a peer proved it by doing it safely.**
> The floor keys are read from **STEP config**, and each AGENT has its own step. `[MEASURED
> 2026-09-03]` `page-build-handler.save_sections.config.section_shrink_floor = 0.1` (a scoped
> override, owner-ruled, applied by migration 725) while `page-rerender.save_sections` has it
> **unset** — the two are separate rows and setting one does not touch the other.
> **So a scoped override IS possible: per AGENT-step, not global.** What stands is the narrower
> caution: setting it on **`page-rerender`'s** step reaches every rerender in the fleet's
> highest-volume pipeline, and even a per-agent override is not per-PAGE — it is
> per-pipeline-for-the-duration, which is why the peer paired theirs with a monitored ROLLBACK at
> the item's terminal state. **Scope it to the agent that needs it, time-box it, and roll it back
> — do not reach for it as a way past a refusal you have not read.**

**Do NOT set `section_component_floor` casually, which is what the error invites.** Checked at source:
`save_sections_component_floor.go:158` reads it via
`pruneFloorFromConfig(params.StepConfig.Config, …)` — **step** config, which lives in the
`page-rerender` agent definition. "Set it for this page" does not exist; setting it lowers a
flattening guard for every rerender in the fleet's highest-volume pipeline to land one page. Same
shape as its sibling `section_shrink_floor`, which `LANDMINES.md` already documents with the same
warning. The four were **cancelled** instead, with the reason written into each row.

**My pre-flight checked the wrong guard.** I measured the text floor — correctly, and it would not
have fired — and did not know the component floor existed. It was found by the `idea.uk` lane
reading its own refusal, not by me.

### What is now unblocked

Migration `683`'s stated precondition is **met**: the Go is live, so a re-render will re-resolve
`articles` through the fixed projection and produce a clean headline and a real deck. `683` is
still deliberately unapplied — **applying it files rerenders across 14 pages on 6 sites owned by
other lanes, which is a dispatch decision and not a schema change.** The header carries the
one-line narrowing to a single domain for whoever takes it.

### Verify

```bash
python3 scripts/check_card_slot_guards.py --component content-listing   # exit 0 since 682
python3 scripts/check_card_slot_guards.py --self-test                   # fires on pre-682, quiet on post
python3 scripts/check_list_empty_states.py                              # now sees 72 blocks, not 8
```

```sql
-- after a roll AND a re-render of a query.blog_posts consumer:
SELECT content_data->'articles'->0->>'title',           -- no " | <site>" suffix
       content_data->'articles'->0->>'excerpt'          -- non-empty
  FROM page_components WHERE id = '7f3b6cae-f118-4d84-80a1-97b75357b9c7';
```

## Evidence that the fix is discriminating, not assumed

- **Render proof with a control, run BEFORE applying 682.** The new template executed under
  *both* live executors (the component seam's `missingkey=zero` and `renderBlogTemplate`'s
  bare `text/template`) against a populated card, the live sparse boxingonline card, and a
  card with no image: 6 of 6 with **zero** empty elements and no `<no value>` leak. Control:
  the pre-682 template on the same sparse data produces **2** empty elements. A proof that
  cannot come out the other way is not a proof.
- **The migration's verify block was INDUCED before the real apply** — run against the
  unguarded template it raised, naming all six slots (`category, excerpt, date, read_time,
  meta-row, section-header`), and the transaction rolled back with the template untouched.
  It is a `DO`/`RAISE`, not a block of `SELECT`s: `ON_ERROR_STOP` does not fire on a
  non-empty result set, so `SELECT`s cannot stop the `COMMIT`.
- **The detector's `--self-test` proves it can go quiet**, not only that it can fire: it
  reports the four real slots on the pre-682 template and nothing on the post-682 one.

## Two things found by looking that the report did not mention

**1. The section header had the same defect one level up, and the render proof found it.**
`section__header`'s two children were already guarded; the wrapper was not. A section with
neither title nor subtitle rendered an empty `<div>` carrying its own margin. Guarding only
the cards would have left it. Fixed in 682.

**2. The same root cause points the OTHER way, and my first fix missed it.** The `090` run
(correlation `afbf8544-feaf-4742-be03-76c87607744f`) hit its iteration cap and returned
UNVERIFIABLE — no verdict — but its last refuted hypothesis named a second component, and
the claim checks out at the DB:

```
agritec.uk                 /guides/index.html       blog-listing_pre_037  image,meta_description,name,nav_label,title,url
fundamentallyai.com        /platform-log/index.html blog-listing_pre_037  image,meta_description,name,nav_label,title,url
leopardessconsulting.co.uk /blog.html               blog-listing_pre_037  category,date,excerpt,image,read_time,title,url
```

One shared component, **two incompatible item shapes across three sites**, decided by which
producer happened to run. And `blog-listing_pre_037` renders `{{.meta_description}}` where
`content-listing` renders `{{.excerpt}}` — so emitting only `excerpt` starves it exactly as
emitting only `meta_description` starved the other. Both producers now emit **both** keys.

> `[MEASURED 2026-09-02]` that second direction is **latent, not live**: `rebuild_blog_listing`
> loads the `content-listing` template by lookup rather than following the page_component's
> own `component_id`, so `blog-listing_pre_037`'s own template is not reached on that path
> today. Leopardess's `rendered_html` carries `article-card__excerpt`, not this component's
> `bl-card-excerpt`, which is how you can see it. **That is a door held shut by an unrelated
> mechanism, not a guarantee** — and it is a loose thread worth pulling (see below).

## The third value in the same row — `nav_label` — and why it is not the drop-in it looks like

Raised by the boxingonline lane after confirming the crux, 2026-09-02. The same item carries a
**third** differently-named text value that nothing renders:

```
"title":            "Cruiserweight Is Boxing's Best-Kept Secret — And It Won't Stay That Way | Boxing Online"
"meta_description": "Discover why cruiserweight boxing sits between light heavyweight and heavyweight, …"
"nav_label":        "Cruiserweight's Rise"
```

So **three differently-named values feed two visible slots on one item**, and that is the
sharper statement of fix-candidate 4 below: the missing per-item vocabulary is not only why
slots go unguarded, it is why **nothing in the system can say which of `title` and `nav_label`
is the headline.** Both are just keys a resolver happened to write.

Whether a card should show the short label or the full headline is a design judgement and it
sits with the visual-designer seat, not here. **But it must be made against this measurement,
because `nav_label` is not populated enough to be a headline source:**

> **CORRECTED 2026-09-02, same day, before anyone designed against it.** The first version of
> this section read the RAW `pages.nav_label` column and concluded that rendering it "would
> blank most cards fleet-wide". **That is wrong, and the error was mine: the projection
> coalesces.** `resolvePagesWhereType` selects `COALESCE(p.nav_label, p.title, p.name)`, so a
> NULL label never reaches the card — it silently becomes the full title. What caught it was
> asking the next question instead of stopping at a striking number: *does `COALESCE` actually
> save this?* It saves NULL and **not** the empty string, and those are different populations.

`[MEASURED 2026-09-02]` over the **303** active/deployed `blog-post` pages, split by what the
projection can and cannot rescue:

| | count | what the card actually shows |
|---|---|---|
| `nav_label` **NULL** | **251** (82.8%) | the full `title` — `COALESCE` rescues it, so **no shortening at all** |
| `nav_label` **empty string** | **5** (1.7%) | **blank** — `COALESCE` does NOT catch `''` |
| a real, distinct label | ~47 (15.5%) | genuinely short |

So the honest statement is not "it blanks most cards". It is that **rendering `nav_label`
does nothing for 83% of them** — they get the same long title back, via a fallback invisible
in the template — while blanking 5 and shortening ~47. That is a *worse* outcome to reason
about than a uniform one, because the mixed grid it produces has no visible cause: a reader
comparing two cards cannot tell that one fell through a `COALESCE`.

And on boxingonline, 1 of its 7 posts carries `"Article | Boxing Online"` as its `nav_label` —
both a placeholder and a carrier of the very suffix this bug strips.

If the shorter headline is wanted, the route is a **populated** display-headline field, not a
fallback chain over a column that is unset on 258 of 303 rows.

Recorded here so the next reader of that row does not rediscover `nav_label` and reach for it.

## Fix candidates, ordered by what closes the door

1. **DONE — share the rules.** `queryresolve.ListItemTitle` / `ListItemExcerpt`, one
   spelling, both producers. The estate already shares this shape's SQL
   (`ListedPageEligibilitySQL`, `PageImageProjectionSQL`) for exactly this reason; the TEXT
   half was the part nobody had shared. `ListItemExcerpt` also slices by **rune**, retiring
   a byte-slice truncation of a meta description — one of the three flagged in
   `bugs_open/423`'s addenda.
2. **DONE — guard the slots** (migration 682), so a component can tell a missing input from
   an intentional blank.
3. **DONE — detect the class** (`check_card_slot_guards.py`) and un-blind its sibling.
4. **OPEN, and the only one that makes the bad state unrepresentable: give `input_schema` a
   per-item field vocabulary.** *(Its cost is now visible in two directions: unguarded slots,
   AND three unranked names for two slots — see the `nav_label` section above.)* Today `required` / `on_missing` / `fallback` exist per
   TOP-LEVEL field and the resolver honours them — but there is no vocabulary for the fields
   INSIDE an array item. `articles {required: true, on_missing: skip_section}` governs
   whether the listing renders at all and says **nothing whatever** about
   `articles[].excerpt`. So there is no declaration to join against, which is precisely why
   `check_card_slot_guards.py` cannot tell a required headline slot from an optional
   read-time one and must report candidates. Until this exists, every listing component's
   per-item slots are ungoverned: they render whatever the resolver happened to name, and
   blank otherwise.
5. **OPEN — `rebuild_blog_listing` should render the component the row points at.** It looks
   its template up by name instead, which is what makes item 2's second direction latent
   rather than live. Not fixed here: it changes which template a live listing renders, which
   is a bigger blast radius than this bug, and it deserves its own measurement.

## How wide is this really — the triage the council asked for

> `bug_historian` objected that this fixes ONE call site of a generic unpatched mechanism, and
> asked (as MISSING) whether other components share the standard list-item shape and could be
> showing the identical symptom right now. Fair, and answerable. `[MEASURED 2026-09-02]`, every
> ACTIVE component whose `input_schema` declares a page-list `query.*` source, with its live
> instance count:

| component | source | live instances | sites |
|---|---|---|---|
| `tool-cta` | `query.pages_where_type:tool` | **96** | 19 |
| `tool-list` | `query.pages_where_type:tool` | 14 | 10 |
| `content-listing` | `query.blog_posts` | **14** | 6 |
| `guide-list_pre_037` | `query.pages_where_type:guide` | 10 | 7 |
| `blog-listing_pre_037` | `query.blog_posts` | 3 | 3 |
| `game-list_pre_037` | `query.pages_where_type:game` | 2 | 1 |
| `archetype-grid` | `query.pages_where_type:entity-page` | 2 | 2 |

**141 live instances of one shape. Migration 682 fixes 14 of them — but the Go half fixes all
141**, and that distinction is the answer to the objection. `bug_historian` rightly called the
breadth an *unverified* claim in round 2, since `tool-cta` alone is 96 of the 141; it is
demonstrated now rather than asserted — `queryHandlers["pages_where_type"]` is a closure calling
`resolvePagesWhereType`, `tool-cta` declares `query.pages_where_type:tool`, and all three bases
are closures over the two functions the Go edit changes. `ListItemTitle` / `ListItemExcerpt`
sit inside `resolvePagesWhereType`, which serves `pages_where_type:*` **and** `blog_posts`, so
the suffixed-headline defect is corrected fleet-wide by one edit. Only the *template guarding*
is per-component, and it has to be: each template guards different slots.

**The remaining exposure, named rather than waved at.** The other six render `.title`,
`.meta_description` and `.nav_label` unguarded inside their ranges. All three are **always
written** by the projection, so an unguarded slot is only a live blank where the underlying
column is empty. Measured over the 753 pages of the five feeding page types:

- `meta_description` empty: **7 fleet-wide** (3 blog-post, 2 entity-page, 2 tool) — so seven
  cards somewhere carry an empty deck today.
- `nav_label`: NULL on 472, which the projection's `COALESCE` rescues; **empty string on 5**,
  which it does not.

So the class is wide and the live damage is small — which is the honest shape of it, and the
reason those six are a **triage item rather than a second migration**. `check_card_slot_guards.py`
now reports them on every run, so they are visible rather than remembered.


## Why this is a class and not a site bug

An empty-slot-tolerant component **silently converts a data gap into a design flaw, and the
design flaw is what gets reported.** Nobody looks for a missing excerpt; they look at the
card and say it needs a better design. The routing then goes to a designer, who cannot fix
it, because there is nothing wrong with the design.

The peer lane that raised this counted it as the third instance in one week on this site of
a data gap presenting as something else — an index with nothing to index wrote a manifesto;
a calendar tool with no events wrote its own editorial policy. That is the same shape: a
mechanism given nothing to say, saying something anyway.

## Related

- `bugs_closed/054_HANDOFF_2026-07-21_unguarded_range_items_in_list_templates_no_empty_state.md`
  — the array-level sibling. Its lint's blind spot is fixed here. **Resolve 054 by slug: the
  number collides with two other closed cases.**
> **CORRECTED 2026-09-02 — the consumer count was 13 when I measured it and 14 hours later.**
> The council's own verification query returned **14** where the submission claimed 13, and it
> was right: **boxingonline.com gained a SECOND `content-listing`** (`/guides/index.html`,
> alongside `/index.html`) the same day, after my census and before their check. Current tally:
> homegarden.uk 6, boxingonline.com 2, dartsonline.com 2, garden-tools.uk 2, idea.uk 1,
> robot-hands.com 1 = **14 across 6 sites, as of 2026-09-02**. Nothing was wrong with the
> measurement; it went stale **by addition, within hours, on the very lane that took it** —
> which is exactly why CLAUDE.md requires a count to carry the date it was counted, and why
> `--since <that date>` is the re-check.

- `bugs_open/384` — the shared image projection, the same "two writers of one listing field"
  class, already fixed for `image`.
- `bugs_open/423` — the byte-slice truncation family; one of its three is retired here.
- `bugs_open/309` — cards rendered unclickable. The reason `check_card_slot_guards.py`
  deliberately does **not** report `<a href="{{.url}}">{{.title}}</a>`: a missing `url` makes
  a card broken, not blank, and that is 309's class with its own detector.
