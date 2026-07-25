# 028 — a page-build no-op reported `complete` and deployed a page built from other pages' components

**Filed 2026-07-19** (relojistas thread). **Status: CLOSED 2026-07-25 — all three
defects resolved, fix live and verified on the live pages.** See
`## CLOSED 2026-07-25` at the foot for the closing evidence.

> **THE TITLE OF THIS CASE IS WRONG, and that is the most useful thing in it.**
> Nothing was ever "borrowed". No component is copied from a sibling or a
> site-level default. Each hero is **generated fresh**, and on these pages the
> writer was given no page subject to generate from. Two threads, ten weeks
> apart, described a fallback mechanism that does not exist — and the second one
> re-derived every *other* claim from the live DB while inheriting this one from
> the title. Logged in `WRONG_CALLS.md` (2026-07-25) and `016b` §9.

> Numbering note: a concurrent session also used `027` on 2026-07-19
> (`027_..._content_hero_unstyled_on_sites_without_a_style_guide.md`), and `028`
> is itself one of the documented ambiguous numbers — `bugs_closed/028` is a
> different case (`avoid_lists_are_inert_banana_discards_negative_prompts`).
> Per `bugs_closed/README.md` numbers are never reassigned and a bare number is
> ambiguous — resolve by slug. This case is `028-page-build-noop`.

## Symptom

`page-build-handler` recorded a no-op for `glosario-tourbillon` on relojistas.com:

```
site_work_items.error:
  page-build-handler no-op: no sections ready to build
  (empty spec sections, or all sections deferred for missing data)
```

and the item still ended at **`status='complete'`**, the page at
**`build_status='deployed'`**, live at `/glosario/tourbillon.html`.

The handler composed nothing. The page shipped anyway, carrying components that
were not its own:

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

> **CORRECTED 2026-07-25:** defect 2's second sentence is wrong. There is no
> fallback. See `## CLOSED 2026-07-25` below.

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
- `bugs_closed/040-partial-build` — the closest relative; owned defect 2. CLOSED 2026-07-24.
- `bugs_closed/041` (`section_lookup_never_normalises`) — the upstream "no sections ready"
  cause. CLOSED 2026-07-22.
- `bugs_closed/025` (`content_direction_column_documents_behaviour_that_does_not_exist`) —
  built the lever that closes candidate 3. CLOSED 2026-07-21.
- `bugs_open/078` (`null_handler_agent_silently_livelocks_the_build_dispatcher`) — filed
  from this session; it halted the fleet build queue while the last page here was being
  rebuilt. Not a cause of 028, but read it if a rebuild queued from this case's runbook
  appears to hang.

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

---

## CLOSED 2026-07-25 (bugfix-028 session)

All three defects resolved. Defects 1 and 2 were already closed by other work; candidate 3
— the residual this case was held open for — is fixed and verified on the live pages.

### Defect 1 — still closed, re-verified today

Re-ran the same grouping. **`complete` is still 0**, four days and 25 more no-ops later:

```
 needs_human_review | 25 | 2026-07-25 09:28:51   <- correct outcome, still firing
 cancelled          |  1
 rejected           |  1
 unresolved         |  1 | 2026-07-25 15:32:28
 complete           |  0   <-- still gone
```

The `unresolved` row is not a regression: it is a `[stale: triaged 48h+]` row parked by the
staleness reaper (`bugs_open/070`), and `unresolved` is terminal-but-unsuccessful — it is in
`idx_swi_dedup`'s excluded status set. The no-op still never reaches `complete`.

### Defect 2 — its owner closed it

`bugs_open/040-partial-build` → **`bugs_closed/040`, CLOSED 2026-07-24**, fixed at every
layer and live (guard v1.0.1146, skip persistence v1.0.1155), council `164058e6` APPROVED.
Its upstream, `bugs_closed/041` (`section_lookup_never_normalises`), **CLOSED 2026-07-22**,
live on v1.0.1146. Nothing is owed here.

### Candidate 3 — the mechanism, and why "borrowed" was the wrong word

**There is no borrowing fallback. There is no fallback at all.** The heroes were
*generated*, not copied. The evidence that settles it is a field-by-field diff of the rows —
which is the check neither earlier pass ran:

| field | glosario-tourbillon (broken) | index (the "source") |
|---|---|---|
| `headline` | Relojería en español: noticias, guías y glosario | **same** |
| `subheadline` | "Seguimos el sector de cerca…" | *different* |
| `cta_text` | Ver las últimas noticias | Leer las últimas noticias |
| `hero_url` | `/assets/images/hero.jpg` | `/assets/images/hero-home.jpg` |
| `background_image` | `/assets/images/hero.jpg` | `/assets/images/hero-home.jpg` |

A copy matches in every field. One field matching is *generation from shared context*.

**The actual mechanism.** `page-build-handler` → `page-content-writer`; the writer's prompt
(`process_sections_loop.prompt_template`) opens with

```
Write content for the {{.current_section.name}} section of {{.current_page.title}}.
```

and then supplies a large **Company Context** block (company name, industry, tagline,
`value_proposition` via site specs). On a glossary entity-page `title` is a single bare
term — "Tourbillon", "Calibre". These pages have `content_direction` NULL, `page_spec` NULL
(all 19 relojistas pages), no `meta_description`, and a `content_brief` of
`{"purpose":"","tone_direction":"","section_guidance":"hero section"}`. So the model is
handed a one-word subject and a page of company context, and writes a site-level hero
paraphrasing the site's `value_proposition`.

**It is non-deterministic, and that is the proof there is no branch to find.** All 8 glossary
pages have *identical* inputs — same empty brief, same NULL `content_direction`, same
2-section plan (`hero`, `generic-text-block`), same bare-term title shape. 6 came out wrong,
2 came out right. Identical inputs, divergent outputs: the variance is in the model, not in
code.

**Containment re-tested, not inherited.** A fleet sweep of every deployed leaf page
(`entity-page`, `guide`, `blog-post`, `tool`, `game`) whose hero headline omits its title's
first word flagged 8 sites. **Every one outside relojistas was a false positive** — those
headlines rephrase rather than repeat (gamesdesign's *"Your XP Curve Is Probably Wrong"* for
*"Understanding XP Curve Designer"*, leopardess's *"What LLM Providers Actually Cost in
Production"*). The real discriminator is a **bare one-word page title**, and relojistas'
glosario is the only such page set in the fleet. The 07-21 containment holds — but it now
holds because it was re-tested, not because it was repeated.

### The fix — migration `210`, config only

`docs/agent_docs/sql_for_agents/210_relojistas_glosario_content_direction.sql`
(commit `d3242b6d5`, applied and recorded 2026-07-25).

Sets per-page `pages.content_direction` on **all 8** glossary entity-pages — an
`instruction` naming the term, a one-sentence `format`, an `examples` entry, and an `avoid`
list that names both offending site-level headlines explicitly.

This uses the lever **`bugs_closed/025`** built and proved for exactly this purpose (wired
v1.0.1146; `load_page_record_action.go:171` and `get_pages_to_build_actions.go:104,126`
select the column; it reaches the writer as `.current_page.content_direction` and is consumed
by the prompt's *"## Page-Specific Content Direction (for THIS page - follow closely)"*
block). No Go change, no image roll, no fleet blast radius.

Applied to all 8, not just the 6 broken: **the two correct pages were luck, not design**, and
without steering they would regress on any future rebuild.

**Do NOT hand-set `page_components.content_brief` instead.** The writer does read it
(`enrichSectionComponentsWithBriefs`, v3_site_actions.go:3726), but
`save_page_sections_action.go:604-614` rewrites that column from `page_spec->>'purpose'` on
every save — so the very rebuild you trigger clobbers your brief. `content_direction` is read
straight off the `pages` row and is never rewritten by the build.

### Live verification — the served pages, not the job status

The 6 broken pages were re-queued (`item_key = 'bug028_hero_rebuild:<page>'`) and rebuilt.
Live `<h1>` on `https://relojistas.com/glosario/*.html`:

| page | before | after |
|---|---|---|
| tourbillon | Relojería en español: noticias, guías y glosario | **Tourbillon: qué es y cómo funciona esta complicación** |
| calibre | Relojería en español, sin rodeos | **Calibre: el motor de un reloj y cómo leerlo** |
| complicacion | Relojería en español, sin rodeos | **Complicación: cualquier función del reloj más allá de la hora** |
| horas-saltantes | Relojería en español, sin rodeos | **Horas saltantes: la hora aparece en una ventana, sin agujas** |
| movimiento-automatico | Relojería en español: noticias, guías y glosario | **Movimiento automático: cómo funciona el mecanismo que se carga…** |
| cronografo | Relojería en español: noticias, guías y glosario | **Cronógrafo: un reloj que también mide el tiempo transcurrido** |
| hermeticidad | *(was already correct)* | unchanged — not rebuilt |
| reserva-de-marcha | *(was already correct)* | unchanged — not rebuilt |

**Steering worked on every page it reached: 6/6, against an unsteered baseline of 2/8.** The
bodies stayed on-subject too — complicacion's `generic-text-block` still opens *"una
complicación es cualquier función de un reloj que va más allá de mostrar las horas y los
minutos"*.

**The original detector query now returns zero rows fleet-wide** — no site has a hero
headline shared across more than two deployed pages:

```sql
SELECT s.domain, count(*) grp, sum(n) pages FROM (
  SELECT p.site_id, pc.content_data->>'headline' h, count(*) n
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE pc.slot_name='hero' AND p.build_status='deployed'
    AND COALESCE(pc.content_data->>'headline','')<>''
  GROUP BY 1,2 HAVING count(*)>2
) d JOIN sites s ON s.id=d.site_id GROUP BY 1;
-- (0 rows)      <- was: relojistas.com | 2 | 7
```

All six work items are `complete`; `glosario-cronografo` was built via direct fire and
closed by hand (see below), the other five by the dispatcher.

### What it took to rebuild the last page — worth knowing before you re-run any of this

`glosario-cronografo` took ~90 minutes longer than the other five, for reasons that had
nothing to do with this fix and everything to do with the build queue:

- **The dispatcher picks ONE site per 120s tick, `ORDER BY wi.site_id … LIMIT 1`** — an
  arbitrary UUID ordering. relojistas (`ecf1…`) sorts *last* of the active sites, so it waits
  behind every other site holding work (webdesign was sitting on ~95 items). `priority` does
  not help; it only breaks ties *within* the winning site. Filed as a pattern in `016b` §9.
- **Then the queue livelocked entirely** on a single work item with `handler_agent IS NULL`,
  which the site-selection query counts and the item-loader silently refuses. No builds
  completed anywhere on the fleet for ~45 minutes. Filed as **`bugs_open/078`**, with the
  offending row repaired.
- The escape hatch is
  `docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/scripts/049c_build_single_page.sh <work_item_id>`
  — a full build (content writer included) for one item, unlike `049b` which is
  assemble-only. **Caveat, and it bit here:** a direct fire bypasses the loop's `claim` and
  `mark_complete`, so a *successful* build leaves the item `triaged` and it must be closed by
  hand or it will be rebuilt again later. It is also slow to land under chassis load — the
  orchestration row took several minutes to appear both times, which reads exactly like a
  dropped message. **Do not re-fire on that evidence.**

The case closes on the mechanism being fixed and proven on all six repaired pages plus two
protected ones, with the fleet-wide detector clean.

### Residual: a NEW glossary page is still exposed — say so plainly

Migration `210` sets `content_direction` on the **eight pages that exist today**. It is a
per-page column, so **a ninth glossary entry created tomorrow gets no steering and can come
out with a site-generic hero exactly as these did.** The fix repairs and protects the
current set; it does not immunise the page *type*.

That is a deliberate scope call, not an oversight, and the reasoning should be checkable:
the fleet survey found no other site exposed, so a fleet-wide change (strengthening
`page-content-writer`'s prompt for every page on every site) would carry real regression
risk across ten sites to defend one site's glossary. The proportionate fix for the *type*
belongs to whoever owns entity-page authoring on relojistas, and has two obvious shapes:

- set `content_direction` at creation time in the glossary entity-page authoring path (same
  jsonb shape as `210` — copy the `format`/`avoid` values from it); or
- give these pages a descriptive `title` rather than a bare term, which is what makes every
  other site's leaf pages immune (`"Understanding XP Curve Designer"` never produces this
  failure; `"Tourbillon"` does).

**Owner: relojistas rebuild workstream** (`HANDOFF_RESUME_relojistas_rebuild.md`). Filed
here rather than as a new bug because it is the same mechanism, already fully documented
above, and `/bugs_open/` is for what is biting production — this is not, until a ninth
glossary page is written.

### What this case is worth remembering for

- `016b` §9 — *"Borrowed content" is a claim about PROVENANCE, and a rendered page cannot
  show you provenance*, and *The build dispatcher picks ONE site per tick ordered by
  `site_id`, so cross-site priority does nothing*.
- `WRONG_CALLS.md` 2026-07-25 — *twice we called it "borrowed"; nothing was ever borrowed*.
  The transferable rule: **a claim in a case's title is the one nobody re-tests.**
