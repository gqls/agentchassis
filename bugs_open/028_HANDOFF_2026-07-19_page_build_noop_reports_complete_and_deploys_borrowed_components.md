# 028 — a page-build no-op reported `complete` and deployed a page built from other pages' components

**Filed 2026-07-19** (relojistas thread). **Status: OPEN — RE-SCOPED 2026-07-21.** Silent
success — the most expensive failure shape this platform has, because every status field says
the work happened.

> **RE-SCOPED 2026-07-21 (bugfix-028 session) — read `## INVESTIGATION 2026-07-21` at the
> bottom before acting on the fix candidates above.** In short, grounded against the live DB
> and the live page:
> - **Defect 1 (a no-op must not be `complete`) is FIXED & LIVE** via migration `149`
>   (`sql_for_agents/149_page_build_handler_noop_flags.sql`). Proven in prod: 22 no-op items
>   are parked at `needs_human_review`, **0 at `complete`**. This half is closed.
> - **Defect 2 (a page must not deploy without a genuine build)** is the shared family defect
>   now owned by **`/bugs_open/040-partial-build`** (bugfix_003 workstream); **`/bugs_open/041`**
>   (`section_lookup_never_normalises`) is the upstream "sections vanish" cause. Do not build a
>   competing deploy gate.
> - **The "borrowed components" is a site-generic HERO, not the homepage about-copy described
>   below, and it is CONTAINED to relojistas.com** (6 glosario entity-pages; no other site
>   affected). That residual belongs to the relojistas rebuild workstream, not to a fleet fix.

> Numbering note: a concurrent session also used `027` on 2026-07-19
> (`027_..._content_hero_unstyled_on_sites_without_a_style_guide.md`). Per
> `bugs_closed/README.md` numbers are never reassigned and a bare number is ambiguous —
> resolve by slug. This case is `028`.

## Symptom

`page-build-handler` recorded a no-op for `glosario-tourbillon` on relojistas.com:

```
site_work_items.error:
  page-build-handler no-op: no sections ready to build
  (empty spec sections, or all sections deferred for missing data)
```

and the item still ended at **`status='complete'`**, the page at
**`build_status='deployed'`**, live at `/glosario/tourbillon.html`.

The handler composed nothing. The page shipped anyway, carrying two components that were
not its own:

- `hero` holding the **site homepage's** headline — "Relojería en español: noticias, guías
  y glosario" — not a word about tourbillons;
- `content-block-about` holding generic about-the-company copy.

A page titled "Tourbillon" published to a live site saying nothing whatsoever about a
tourbillon, and nothing anywhere reported a problem.

## Why this is distinct from `bugs_open/015`

015 produces the *same error string* from a different cause (a mistyped `page_type`
orphaning the page from its machinery). Crucially, **in 015 the work item correctly went to
`needs_human_review`.** That is the right outcome for a no-op.

This case is the failure of that outcome: the identical no-op reached `complete` and
deployed. So 015 is "the page never built"; 028 is "the page didn't build and said it did".
Fixing 015 does not fix this.

## What is actually wrong — two separable defects

1. **A no-op must not be `complete`.** Whatever marks the work item complete is not
   consulting the handler's own no-op result. 015 shows the `needs_human_review` path
   exists and works, so this is a routing inconsistency, not a missing capability.
2. **A page with no composed sections must not deploy** — and more seriously, it must not
   deploy *borrowed* components. The provenance of the hero and about block on that page is
   the open question: they belong to other pages of the same site. Something is falling back
   to site-level or sibling components when a page has none of its own, and that fallback is
   invisible in the output. Whatever that path is, it is capable of publishing a page that
   is confidently, fluently about the wrong subject.

## How it was found (worth repeating)

Only by reading the deployed page's `content_data` instead of its status. Every status field
— work item `complete`, page `deployed`, no error surfaced to any dashboard — said success.
The `016b` invariant *trust the rendered artefact, not the status* is what caught it, and
this case is a clean argument for that invariant.

The originating mistake was mine and is worth stating, because it is how anyone else will
hit this: I populated `pages.sections` and assumed that composed the page. The build reads
its spec sections from **`site_plan_sections`**, a different table. Setting one without the
other yields exactly this no-op. That is a genuine trap, but a trap should produce a stuck
item, not a published page.

## Reproduction

1. Create a `pages` row with `build_status='planned'` and a non-empty `pages.sections`.
2. Do **not** create matching `site_plan_sections` rows for it.
3. Queue `needs_page` for it (`handler_agent='page-build-handler'`, `status='triaged'`) and
   run `build-dispatch-loop`.
4. Observe: no-op logged in `site_work_items.error`; item `complete`; page `deployed`;
   `page_components` populated from components that are not that page's.

## Fix candidates

1. **Make the no-op terminal-but-unsuccessful.** Route it to `needs_human_review` (as 015
   demonstrably does) and leave `build_status='planned'`. Smallest change, closes the
   silent-success half.
2. **Gate deploy on composed sections.** Refuse to deploy a page whose `page_components`
   count is zero *or* whose components were not composed during this build. Closes the
   published-wrong-page half, which fix 1 alone does not.
3. **Find and name the borrowing fallback.** Whatever selects a sibling/site-level component
   when a page has none should either be removed or made explicit and logged. Until it is
   identified, fixes 1 and 2 are guards around an unknown behaviour.

Recommend 1 + 2 together; 3 as the follow-up that explains the mechanism.

## How to verify a fix

Run the reproduction. The item must **not** be `complete`, the page must **not** be
`deployed`, and `page_components` must be empty. Check the DB rows and the served URL —
not the job status, which is the thing that lied.

## Related

- `bugs_open/015` — same error string, different cause, *correct* routing. Read both.
- `bugs_open/026` — a different silent-acceptance defect on the same site: a `required`
  input field rendered empty and still saved. Same family: something that should have
  refused, didn't.
- `bugs_open/040-partial-build` — the closest relative; now owns defect 2 (see below).
- `bugs_open/041` (`section_lookup_never_normalises`) — the upstream "no sections ready" cause.

---

## INVESTIGATION 2026-07-21 (bugfix-028 session)

Grounded against the **live config, the live DB, and the live page** — trusting the rendered
artefact over the status, as the case itself argues. The three original defects resolve to
three different owners, and one of them is already closed.

### Defect 1 — "a no-op must not be `complete`" — FIXED & LIVE (migration 149)

This is closed and was closed before this case was even filed. Migration
`docs/agent_docs/docs024_key_docs_latest/... /sql_for_agents/149_page_build_handler_noop_flags.sql`
(dated 2026-07-14, robot-hands false-completion investigation) wired the no-op exits to park
the work item instead of letting the dispatch loop stamp it complete. **It is live in the DB
config now** (config is live immediately; verified this session):

```
check_has_ready_sections.else_step  = mark_no_ready_sections
mark_no_ready_sections.status       = needs_human_review
mark_no_ready_sections.error_message= "page-build-handler no-op: no sections ready to build
                                        (empty spec sections, or all sections deferred for
                                        missing data) — the target section was NOT rebuilt"
mark_no_ready_sections.next_step    = complete_error
```

The chain that makes it stick: `UpdateWorkItemStatusAction` (v3_site_actions.go) sets
`needs_human_review` unconditionally once `input_data.work_item_id` resolves, and
`CompleteWorkItemAction`'s guard (load_work_item_actions.go:801-809,
`WHERE status NOT IN ('needs_human_review',…)`) then refuses to overwrite the flag when the
dispatch loop calls `complete_work_item` on the success-labelled `complete_error`.

**Live proof (the rendered outcome, not the mechanism claim):** every item whose `error`
begins `page-build-handler no-op:`, grouped by status:

```sql
SELECT status, count(*), max(updated_at)
FROM site_work_items WHERE error ILIKE 'page-build-handler no-op:%' GROUP BY status;
-- needs_human_review | 22 | 2026-07-21 02:39   <- correct outcome, still firing today
-- wont_fix           |  5                       <- human/triage dispositions of parked items
-- cancelled          |  1
-- rejected           |  1
-- complete           |  0   <-- the bug outcome NO LONGER EXISTS in prod
```

So the original observation ("the identical no-op reached `complete` and deployed") predates
149 being effective for that path — most likely the manually-queued item in the reproduction
carried no `input_data.work_item_id`, so under `skip_if_missing:true` the flag was skipped
before 149's wiring was in place. **The no-op → `complete` path is gone.**

### Defect 2 — "a page with no composed sections must not deploy" — OWNED BY 040-partial-build

This is the shared family defect ("page state not derived from build reality") and it is being
worked under `bugs_open/040-partial-build` (bugfix_003 workstream). 040's fix candidate 1
(*never stamp `built_from_plan_version`, and do not `deploy`, on a partial/failed build*) plus
its candidate 3 (*a plan-vs-artefact completeness check — it explicitly asks whether
`complete_work_item_verification.go` is the right home*) is the structural fix and it subsumes
028's fix candidate 2. **Do not build a competing deploy gate here** — contribute into 040.

Upstream of it, `bugs_open/041` (`section_lookup_never_normalises`) is the cause of most
"no sections ready" no-ops: a `snake_case` section (`call_to_action`) fails the un-normalised
lookup in `loadSectionComponents` (v3_site_actions.go:3374, :3446), is marked `deferred`, and
the page deploys short. 041 is unfixed as of this writing; its fix is a clean normalise-before-
both-passes change and is the highest-leverage fix in the family.

### Candidate 3 — "the borrowing fallback" — NAMED, and CONTAINED to relojistas

The original description ("`hero` holding the homepage's headline; `content-block-about`
holding generic about-the-company copy") is **half right and worth correcting**:

- **It is a HERO, and its content is site-generic, not the current homepage's.** The live page
  `https://relojistas.com/glosario/tourbillon.html` serves
  `<h1>Relojería en español: noticias, guías y glosario</h1>` — a generic site headline with a
  news CTA ("Ver las últimas noticias"). The *current* homepage headline is a different string
  ("Relojería en español, sin rodeos"). So the hero was not copied from today's homepage; it
  is a site-level/default hero applied where no page-specific hero was authored.
- **The page is NOT a pure no-op deploying pure borrowed content.** glosario-tourbillon also
  carries a genuine, correct `generic-text-block` about tourbillons. So exactly one slot (the
  hero) is wrong; the body was written for the page. A "page_components count is zero" deploy
  gate (028's fix candidate 2 as written) would therefore **not** catch this — the section is
  present, its content is just borrowed.
- **Blast radius, grounded on the live DB: relojistas.com only.** No other site has a hero
  headline shared across more than two deployed pages:

  ```sql
  SELECT s.domain, count(*) grp, sum(n) pages FROM (
    SELECT p.site_id, pc.content_data->>'headline' h, count(*) n
    FROM page_components pc JOIN pages p ON p.id=pc.page_id
    WHERE pc.slot_name='hero' AND p.build_status='deployed'
      AND COALESCE(pc.content_data->>'headline','')<>''
    GROUP BY 1,2 HAVING count(*)>2
  ) d JOIN sites s ON s.id=d.site_id GROUP BY 1;
  -- relojistas.com | 2 | 7    (only row)
  ```

  Six `glosario-*` entity-pages carry a borrowed hero — 3 the homepage's "sin rodeos", 3 the
  generic "noticias, guías y glosario" — all `updated_at` 2026-07-19 12:40–13:18, the window
  in which the originating thread hand-rebuilt them. `SavePageSectionsAction` (which clears and
  re-inserts `page_components`) only persists the `content_data` it is handed; the borrowed
  hero content is produced **upstream in content generation**, not in save/deploy.

  Because it is contained to one site's glossary entity-pages, this is a relojistas
  entity-page **hero-authoring** problem and a **data-repair** of ~6 live pages — it belongs to
  the relojistas rebuild workstream (which owns entity-page building there and filed this case),
  not to a fleet-wide fix loop. Naming the exact upstream code site (the entity-page hero
  path that falls back to a site/homepage hero when no page-specific hero exists) is the next
  step for whoever owns that rebuild; it did not warrant an expensive diagnosis run given the
  one-site blast radius.

### Net recommendation

028's headline framing — a *fleet-wide* silent-success mechanism — does not survive the live
data. Keep the case OPEN only for the **contained relojistas hero residual** (candidate 3);
mark defect 1 closed (149, live) and delegate defect 2 to `040-partial-build`. The remaining
live damage (~6 relojistas glosario pages serving a wrong-subject hero) is a site rebuild task,
not a chassis code defect — and per the robot-hands lesson, a hand-edit of `content_data` will
not hold because the hero is re-derived on render, so it needs a proper page rebuild once the
upstream hero-authoring path is fixed.
