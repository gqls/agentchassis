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
| both producers share one title/deck rule | **COMMITTED, INERT** until the next chassis roll — Go |
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

So the propagation is:

```
page-rerender   spec.reason   = 'section_data_resolved'      (re-runs the query — the Go fix)
                spec.page_name = '<page>'
```

**and it must be sequenced AFTER the chassis roll.** Firing it today re-resolves against the
OLD resolver in the running pod: the guarded slots would collapse (682 is live), but no deck
would arrive, and the render would have been spent. `bugs_open/384` / **PBP-048** built
`requestPageListReresolve` for producer-driven invalidation of exactly this shape; a code fix
is not one of its trigger events, so the post-roll pass is a deliberate, dispatched one.

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
which it corrected an hour later.

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
141**, and that distinction is the answer to the objection. `ListItemTitle` / `ListItemExcerpt`
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
