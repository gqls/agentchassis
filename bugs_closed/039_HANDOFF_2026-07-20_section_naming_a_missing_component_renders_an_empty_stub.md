# Handoff — a section naming a component that does not exist renders an empty stub, and the build calls it success

**Filed 2026-07-20**, found while verifying `/bugs_open/001` live. Two related things: a **naming
convention that is easy to get wrong** (and that anyone comparing plans to pages must know), and a
**defect** — a section name that resolves to nothing produces a hollow `<section>` on a deployed
page instead of failing.

---

 ## CLOSED 2026-07-22 → `/bugs_closed/` — mechanism FIXED & LIVE on `v1.0.1146` (`bd4dc30a0`)
>
> > **CLOSE DECISION 2026-07-22 (owner: bugfix-039 thread).** The defect this case is about — *an
> > unresolvable section name renders a hollow stub and the build reports success* — is fixed and
> > live, so it is no longer reproducible for new builds. What was verified, and the one honest
> > caveat, spelled out so the close is not an overclaim:
> >
> > - **Deployed:** pod on `v1.0.1146`, `strings /app/agent-chassis | grep -c isEmptyGenericStub` = 2.
> >   (Committed 12:44; rode another session's HEAD-build sweep `fe2ba5e52` at 13:04 — the bugs_open/047
> >   lesson.)
> > - **Predicate proven against the real fault, not a happy path:** the 7 live stub rows each satisfy
> >   *both* guard conditions (`component_id IS NULL` AND an empty `section--generic` body) at their
> >   actual save time — so the discriminator provably matches genuine production fault input. Unit
> >   test `TestIsEmptyGenericStub` locks the logic (6 cases).
> > - **Production-stable since the guard went live:** stub count holds at exactly **7** (not growing);
> >   **63 `page_components` rows were built across 6+ hours of real builds with 0 new stubs.**
> > - **Layered prevention is why the backstop is quiet:** **0** `needs_new_component` items have been
> >   raised since deploy, i.e. no unresolvable section reached the save path at all — plan-time
> >   `validate_components` (live in both planners) strips them first, and the render-time
> >   `missingRequiredLLMFields` guard catches the content-writer path. The save guard is the third,
> >   innermost layer.
> >
> > **Honest caveat (why this is a close, not a claim of "behaviourally verified"):** the guard's
> > *skip+raise* branch has **not** been induced by live fault-injection — precisely because the two
> > upstream layers strip the fault before it reaches the save INSERT, so forcing one past them would
> > need a contrived direct-`pages.sections` build. The 6-line skip+raise wiring is therefore proven
> > by review + the fault-data match, not by execution. Given the strength of that evidence and the
> > disproportionate cost/risk of a contrived induction, the mechanism is closed.
> >
> > **Residual, tracked separately (NOT this defect):** the **7 legacy hollow stub rows** on 3 live
> > sites (finetuning.uk ai-guides/insights, gaswholesalers.com fuel-industry-insights) are pre-guard
> > *data damage*. Cleaning them is **content work, not a mechanical fix** — those slots
> > (`featured-article`, `article-grid`, `category-section`) genuinely need real components (the
> > `needs_new_component` route) or an owner content decision; stripping them degrades the pages, and
> > rebuilding them risks the `/bugs_open/029` fabrication path. **Routed to the
> > `empty_sections_loop_integrity` workstream.** Do NOT force-rebuild the 3 live sites to clear them.
>
> **Candidate 1 applied structurally, at the single chokepoint.** Every page-composition path
> flows through one INSERT — `SavePageSectionsAction` (`save_page_sections_action.go:543`). The fix
> adds a guard there: when a section is about to be written with **no component link
> (`componentIDPtr == nil`) AND its HTML is an empty `section--generic` stub** (new helper
> `isEmptyGenericStub` — `section--generic` marker present AND stripping tags leaves no visible
> text), it **skips the row** and raises a **deduped `needs_new_component` work item** (via the
> existing `CreateNeedsNewComponentItem`, routed to `component-creator`, `item_key` per section-type
> per site) naming the page + section. So the hollow section is never persisted as a `deployed`
> row, the build no longer reports a clean success while shipping it, and the gap becomes an
> **actionable, consumer-routed** item instead of a rotting `empty_section` finding.
>
> **Why the chokepoint and not `GetComponentWithFallback`:** the generic fallback is *also* how the
> **22** live sections that resolve to nothing but DID receive real content are rendered — those are
> legitimate and must not break (the handoff's "35 orphans, different situation"). The discriminator
> is therefore **empty-vs-content**, not fallback-vs-not. Placing it at the save INSERT catches the
> stub regardless of upstream path (content-writer HTML fallback, deprecated
> `assemble_from_library`, direct spec-load) in one place.
>
> **Grounded against live data (2026-07-21).** Mirroring the Go guard in SQL (strip tags, remove all
> whitespace, test empty) over the 29 null-component `section--generic` rows on deployed pages:
> **7 GUARD-MATCH = exactly the 7 known stubs; 22 has-text = the content-bearing renders, untouched.**
> The 14 non-generic orphan rows never carry the `section--generic` marker, so they are untouched too.
> Unit test: `save_page_sections_stub_guard_test.go` (`TestIsEmptyGenericStub`, 6 cases, passing).
>
> **Interaction with the existing render guard.** The `generic-text-block` seed already marks
> `heading`+`content` as `required` source:llm, and `RenderComponentAction`'s
> `missingRequiredLLMFields` guard (live since v1.0.1126) refuses an empty required-field render — so
> the **primary content-writer path already prevents new stubs** (all 7 live stubs are legacy:
> gaswholesalers 2026-04-08, finetuning 2026-04-10; none created since). This save-time guard is the
> **structural backstop** that (a) is independent of whether a fallback component happens to declare
> required fields, and (b) closes the dormant non-content-writer leak paths.
>
> **Candidate 2/3 were already live** and are necessary-but-insufficient: `validate_site_plan` runs
> with `validate_components: true` in both active planners (`site-planner`, `build-site-planner`),
> and its `componentNameResolver` resolves by function, normalised function, display name AND
> component `name` (so it already fixes the 8 name-not-function latent entries and drops unresolvable
> names) — but only for section names that pass through the planner, and not for the already-stored
> rows or non-planner rebuild paths. That gap is exactly what this save-time guard covers.
>
> **NOT done here — the cleanup (candidate 4), deliberately deferred:** removing/rebuilding the 7
> legacy stubs is gated on (i) this guard being **live** (a rebuild before then just re-stubs) and
> (ii) real components existing for `featured-article` / `article-grid` / `category-section` — which
> the `needs_new_component` items this guard raises will drive. It is content-sensitive
> (re-rendering guide/insight bodies risks the `/bugs_open/029` fabrication path), so it belongs to
> the `empty_sections_loop_integrity` / component-creator workstream, not this mechanical fix.
>
> **How to finish closing this bug** (after the next chassis roll): (1) confirm the guard is in the
> pod (`strings /app/agent-chassis | grep -c isEmptyGenericStub`); (2) rebuild a page with a bogus
> section name and assert **no** `component_id IS NULL` stub row appears and a `needs_new_component`
> item was raised; (3) re-run the fleet `stub` query and confirm it trends to 0 as pages rebuild;
> (4) confirm the 22 content-bearing generic renders and the 8 name-not-function entries still render.

> ### Council review (advisory) — corr `74cdb054`, 2026-07-21
>
> The run **crashed on a transient Anthropic 529 "Overloaded"** at `review_tooling_provenance`
> after all earlier seats had run, so there is **no aggregate `council_report` verdict**. But 8 seat
> reviews were captured in `collected_data`: **6 approve** (constitution, editquality, guidelines,
> mission, reuse_agent — plus prior_art's substance was positive), **2 logged `object`**
> (bug_historian, prior_art). None objected to correctness; all objections are grounding/completeness.
> Addressed here:
>
> - **"Is the `needs_new_component` consumer live, or is this a silent work-item pileup?"**
>   (prior_art, bug_historian). Checked: `component-creator` **is** an active agent
>   (`agent_definitions.is_active=t`), and `needs_new_component` items **reach a terminal status**
>   (2 complete, 11 failed) — i.e. they are *claimed*, unlike `empty_section` findings which rot
>   `unresolved` forever. So routing to it is strictly better than the rotting-detection path this
>   bug is about. **Caveat, recorded honestly:** the consumer's recent record is poor (11 failed),
>   so a raised item is not a *guarantee* the component gets built. The fix's core win — **not
>   persisting a hollow stub and not reporting false success** — does not depend on it.
> - **"'Single INSERT' claim not code-quoted"** (prior_art, editquality). Verified: the other three
>   `page_components` INSERTs are tool/blog-only — `deploy_tool_action.go:333` and
>   `create_tool_component_action.go:242` always set a real `component_id`;
>   `rebuild_blog_listing_action.go:255` omits the column (blog-listing, position hardcoded 3).
>   `save_page_sections_action.go:543` is the sole general page-composition writer.
> - **"Root mechanism `GetComponentWithFallback` left unpatched — instance-#7 shape"**
>   (bug_historian). Fair architectural note, **considered and deferred**: the fallback is *also*
>   how the 22 legitimate content-bearing generic renders are produced, so making it hard-fail is
>   not safe; the render path is separately guarded (`missingRequiredLLMFields`, v1.0.1126). Making
>   the *substitution point* merely observable (a loud signal without failing) is a reasonable
>   follow-on so a future non-persist consumer inherits the signal — noted, not done, because the
>   persistence choke point covers every path that produces a *deployed* hollow row today.
> - `[UNVERIFIED]` residual (editquality): I have not code-proven that *every* leak path populates
>   `section.HTML` with the `section--generic` fragment identically; the observed 7/7 match is from
>   already-produced rows. In practice the non-content-writer paths either feed
>   `saveSectionsExtractFromHTML` (which preserves the fragment → caught) or are guarded upstream at
>   render, but a code-level confirmation across all three paths is still owed.

## Part 1 — the convention (read this before comparing a plan to a page)

`pages.sections` stores the component's **`function`**. `page_components` reference the component
**row**, whose `name` is usually different:

```sql
SELECT name, function FROM content_components
WHERE name IN ('about-hero','differentiators-section');
--  about-hero              | hero-about
--  differentiators-section | differentiators
```

So a page whose `sections` read `["hero-about", …, "differentiators", …]` renders
`page_components` `about-hero, …, differentiators-section, …`. **That is correct and matching**, not
drift — but compared naively it looks like the page rendered the wrong components. This cost real
time on 2026-07-20: `about` on dartsonline.com was briefly read as a regression when it was fine.

Names are also normalised before lookup — `NormalizeComponentFunction`
(`platform/orchestration/actions/component_validation.go:78`) converts `call_to_action` →
`call-to-action` and `SocialProof` → `social-proof`. So a literal string comparison against
`content_components` is wrong in *two* ways; normalise, then match on `function`.

Fleet-wide, after normalising (743 section entries on active pages):

| | count |
|---|---|
| resolve by `function` (the convention) | 724 |
| resolve only by `name` (namespace mixed) | 8 |
| **resolve to nothing** | **11** |

The 8 name-not-function entries (`differentiators-section`, `services-hero`, `contact-hero`,
`case-studies-hero`, and one literal `faq section` **with a space**) are the LLM writing a component
*name* where a *function* belongs. They happen to work only when a component's name and function
coincide, so they are latent, not currently broken.

## Part 2 — the defect: an unresolvable section renders a hollow stub

The 11 that resolve to nothing do not fail. They render an empty section, and the page deploys.

**`finetuning.uk` `/ai-guides`, `build_status='deployed'`:**

```sql
SELECT sections FROM pages WHERE name='ai-guides' AND site_id=(SELECT id FROM sites WHERE domain='finetuning.uk');
-- ["hero", "featured_article", "category_section", "article_grid", "testimonials", "call_to_action"]
```

```
 comp           | function       | position | rendered_html len
 hero           | hero           | 1        | 2368
 (null)         | (null)         | 2        | 208
 (null)         | (null)         | 3        | 208
 (null)         | (null)         | 4        | 208
 testimonials   | testimonials   | 5        | 3245
 call-to-action | call-to-action | 6        | 2078
```

Positions 2–4 are `featured_article`, `category_section`, `article_grid` — none of which exist as a
component under any normalisation. Each produced a `page_components` row with **`component_id IS
NULL`** and this 208-byte body:

```html
<section id="…" class="section section--generic">
  <div class="container">
    <h2 class="section__title"></h2>
    <div class="section__content"></div>
  </div>
</section>
```

An empty heading and an empty content div. On a guides index page, the three missing slots are the
featured article, the category section and the article grid — i.e. **the entire point of the page**.

**Scale**: 7 stub slots (all exactly 208 bytes) across deployed pages on 3 sites; `finetuning.uk`
and `ai-agent-orchestration.com` carry most of it. A separate 35 orphan slots have
`component_id IS NULL` but real content (842–12,592 bytes) — that is a **different** situation
(content written without a live component link) and is NOT this bug; do not conflate them.

```sql
SELECT CASE WHEN length(pc.rendered_html) < 500 THEN 'stub' ELSE 'has content' END, count(*)
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE pc.component_id IS NULL AND p.build_status='deployed' GROUP BY 1;
--  has content | 35
--  stub        | 7
```

## Why it survives

It IS detected. `check_empty_sections.go:136` has an `empty_heading` rule
(`<(h[1-6])[^>]*>\s*</\1>`) that these stubs match exactly, and `empty_section` items exist —
including one literally named `hero-tool`, one of the unresolvable names:

```
finetuning.uk | empty_section | unresolved | [stale: triaged 48h+] Empty section 'hero-tool' on page llm-cost-calculator
```

**Every one of them is `unresolved`**, most stamped `[stale: triaged 48h+]`, newest 2026-05-01. So
this is not a detection gap — it is the same delivery gap as `/bugs_open/023` and `/bugs_open/033`:
the finding is raised, nothing consumes it, it ages out, and the hollow section stays live for
months. Fixing the emitter below without fixing that routing just produces more unresolved items.

## Fix candidates

1. **Fail loudly at resolution.** When a section name resolves to no component, the page build
   should refuse that page (or emit a `needs_new_component` / `needs_human_review` item naming the
   unresolvable section) rather than writing a `component_id IS NULL` row and reporting success.
   This is the platform's recurring shape — CLAUDE.md's "trust the rendered artefact, not the
   status" — a build that produced three hollow sections should not be `complete`.
2. **Validate at plan time.** `validate_site_plan` already normalises and strips section names; have
   it check each proposed section resolves to a real component function and drop/flag the ones that
   do not, so an unbuildable name never reaches `pages.sections`. Cheaper and earlier, but will not
   help the 11 already stored.
3. **Normalise the namespace.** Accept a component `name` as well as a `function` at resolution
   time (which would fix the 8 latent ones), or stop the LLM proposing names by giving it only
   functions in the prompt's component list. Check what `load_components` currently passes — if it
   offers `name`, the LLM is being invited to make exactly this mistake.
4. **Clean up the 7 existing stubs** — but only after 1 or 2, or they will come back on the next
   rebuild.

## How to verify a fix

1. Plan a page with a deliberately bogus section name. Assert the build does **not** report
   `complete` with a hollow section — it refuses or raises an actionable item.
2. Re-run the fleet query above: `stub` count should reach 0 and stay there after a rebuild.
3. Assert the 8 name-not-function entries still render (do not fix this by breaking them).
4. Check the artefact, not the row: fetch the live page and confirm no empty `<section>` remains.

## Related

- `/bugs_open/023`, `/bugs_open/033` — the delivery gap that lets the detected items rot.
- `/bugs_open/038` — a fix there needs to compare plan sections to `pages.sections` correctly; Part 1
  is the reason a naive comparison would false-negative.
- `/bugs_open/001` — where this was found; its "VERIFIED LIVE" section records the near-miss.
- `docs024_key_docs_latest/empty_sections_loop_integrity/` — the existing workstream on this class.
