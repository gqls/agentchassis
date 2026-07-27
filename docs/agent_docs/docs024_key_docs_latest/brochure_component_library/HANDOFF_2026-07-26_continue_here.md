# HANDOFF — brochure component library / fundamentallyai.com — 2026-07-26

**This is the cold-start document. Start here, not at
`HANDOFF_2026-07-21_start_here.md` (superseded — it predates the site going live).**

Read in this order: this file → **`SUMMARY_2026-07-27_one_line_became_four.md`**
(newest state, plain prose — the series is the record, so the older summaries stay) →
`RUNBOOK_brochure_component_library.md` (every command, each with its gotcha) →
`NOTES_brochure_component_library.md` **from the bottom** (newest last; the
missteps are the point) → `README_where_we_are.md` (the owner's own log; append
only, never rewrite) → `PLAN_2026-07-20_*.md` (design + decisions log).
`MISSION_BRIEF_fundamentallyai_2026-07-20.md` is the originating brief and still
governs.

---

## 1. Constants you will need in the first five minutes

| thing | value |
|---|---|
| domain | `fundamentallyai.com` (LIVE, `.html` URLs — see landmine L1) |
| site_id | `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd` |
| current plan_id | `81741260-6447-492c-bf98-4b3c185f8e7b` |
| owner phone (live on site) | `+44 (0) 7934 524 911` |
| site email | `fundamentallyai@contactforsales.com` |
| DB | `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db` |

Chassis image **as of 2026-07-27 14:0x: `v1.0.1173`**, deployed 13:45 UTC (pod
`agent-chassis-5f85dff548-8d2tq`). It carries this workstream's first platform
change — 085's build-path fix — and several other threads' work. **Re-read this
from the pod, never from the tag or from this line:**
`kubectl -n ai-persona-system get pods -l app=agent-chassis -o custom-columns='NAME:.metadata.name,IMAGE:.spec.containers[0].image,START:.status.startTime'`
Anything Go-side committed after 13:18 UTC is still inert — including 085's
second fix (the scoped-re-render path).

## 2. State: what is done and verified

**The site is sound.** Verified 2026-07-25 ~18:35 by crawling the served pages —
not from the database:

- **43 unique internal link targets across all 7 pages, 0 broken.** (Was 21 of 22
  broken.) Command in the RUNBOOK's "Live crawl" section; re-run it before
  trusting this line, it is a snapshot.

  > **UPDATED 2026-07-27 — and the snapshot warning above earned its keep.** Two
  > pages were rebuilt on the 26th to add the chart, and the rebuilds **authored
  > 16 broken internal links** between them; the gate detected every one and
  > shipped them as warnings. All repaired. Current crawl, 9 deployed pages,
  > anchor-inclusive: **11 unique internal targets, 1 broken** — a favicon that
  > has never existed (see §3a). The counts are not comparable: 43 counted
  > anchored targets separately, 11 strips the fragment first.
  >
  > **The durable lesson is not the count.** A per-page link repair does not
  > survive a rebuild of that page, so "the site is link-sound" describes an
  > artefact and expires the next time anything is rebuilt. **Re-crawl after every
  > rebuild**, and never with a pattern that excludes `#` (landmine L2 — it caught
  > this workstream twice, the second time in this very document's own author).
- 10 pages `deployed`. Contact page complete: `hero-contact` + `contact-form` +
  `contact-info`, phone live as `tel:`.
- `self-correction-leopardessconsulting` LIVE (5 sections) — names
  leopardessconsulting.co.uk directly, does not repeat the invented details.
- **All 5 interactive components are placed at PLAN level and proven to survive
  rebuilds** — five independent pipeline rebuilds each restored its own component:
  index/`stat-band` p2, about/`people-feature-block` p3,
  capabilities/`hero-card-carousel` p2, council/`swipeable-insight-carousel` p4,
  fine-tuning/`image-hover-card-grid` p4.

Verify the whole lot in one go:
```sql
SELECT p.name, p.build_status,
       (SELECT string_agg(cc.function,' | ' ORDER BY pc.position)
        FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
        WHERE pc.page_id=p.id) AS components
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='fundamentallyai.com' ORDER BY p.name;
```

## 3. Owner decisions

### 3a. CHART COMPONENT — **BUILT AND LIVE 2026-07-26 (later session). Kept below for the terms it was built under.**

> **VERIFIED AGAIN 2026-07-27 after the queue drained.** Live crawl of all 9
> deployed pages, capturing `href="(/[^"]*)"` and stripping the fragment
> AFTERWARDS (never `[^"#?]` — see the correction below): **11 unique internal
> targets, 1 broken**, and the one is `/assets/images/favicon.png`, which is
> referenced by the shared `head` chrome and **has never existed at any path**
> (`/favicon.ico`, `/assets/images/favicon.ico` all 404 too). Pre-existing, not
> from this work; favicon/OG derivation is the imagery workstream's tracked gap.
>
> **The chart self-refreshed and the page followed.** Seeded at 108/37/9 on the
> 26th; `refresh_evidence_base` re-ran the queries overnight and the register and
> the live page now both read **110/38/10**. Nobody retyped a number. This is the
> first observation that "a SQL-sourced fact must carry no `display`" earns its
> keep — a hand-written display would now disagree with its own bar.
>
> **CORRECTION, and it is the important line here: the rebuilds broke SIXTEEN
> links, not six.** Capabilities carried ten more, all `/capabilities#…` —
> extension-less *with a fragment* — which my DB check could not see because it
> used `href="(/[^"#?]*)"`, the anchor-blind pattern landmine L2 names, in a
> handoff section I had written that morning. Repaired
> (`bak_pc_fai_cap_links_20260727`); the dead fragments are left to `bugs_open/071`,
> where the corrected count and the evidence now live. Logged in `WRONG_CALLS.md`
> ("independent witness" row now at 2).
>
> **STATUS: done.** `evidence-chart` is registered, seeded, placed on the index at
> plan position 2, and verified on the SERVED page — seven figures, each matching
> its fact row, geometry drawn from the same values. Config-only, so it needed no
> image roll and no council run.
>
> **Read `components/evidence-chart/README.md` first** — it holds the data
> contract, six reproduced traps, and the acceptance state. Commands are in the
> RUNBOOK under "The evidence-chart component".
>
> **What is NOT done:** the chart is on the index only. Per-page targeting is
> blocked by `bugs_open/085` (no page identity reaches a section template, so the
> `pages` key cannot be honoured and every chart renders on every page carrying
> the section). The data is already correct for the fix; restoring the
> capabilities placement is a one-line Go change plus a re-render.
>
> Leopardess needs **only** a `charts` key in its own evidence base — no code, no
> registration, no roll. Recorded in their `REPLICATION_in_chassis.md` and
> `RUNNING_NOTES.md`.

The original terms, unchanged:

The brief requires "numbers rendered as real, code-generated charts from true
figures, never an AI-generated picture of a chart"; the owner asked for "charts and
infographics"; **zero chart components are registered in the entire fleet**
(verified 2026-07-25 — `content_components` matching chart/graph/`semantic_tags`
chart, active: 0). `stat-band` renders verified numbers as text and is not a chart.

Build constraints already decided — do not relitigate:

- **ONE shared component in the chassis**, not a per-site build. It must work for
  fundamentallyai.com *and* leopardess.
- **Values sourced from the `evidence_base` aspect**, so a chart structurally
  cannot display an unverified figure. This is the whole point: the site's pitch is
  that claims are sourced, and a chart is the most persuasive place to put a number.
  Do NOT let the LLM supply values — the LLM may supply labels, framing and
  ordering only. (Same discipline `stat-band` already follows.)
- **Code-rendered** (inline SVG or CSS from the data). Never a generated image of a
  chart. No external chart library at runtime — components ship no CSS/JS unless the
  template inlines it, and JS must go via `js_snippets` (landmine L9).
- **Reuse the prior art, do not start a fourth parallel design:** leopardess **L7**
  is already scoped (`docs/leopardessconsulting/PLAN_leopardess_rebuild.md`, listed
  as `[gap]` in that dir's `REPLICATION_in_chassis.md`, and called "the one
  genuinely-new build"); `features_open/023` covers infographic figures from the
  evidence base. **Coordinate before building** — the leopardess workstream has its
  own docs and may have started. Read theirs first; a second chart component is the
  exact drift the council reviews for.
- Follow the PLAN's acceptance checklist, all 7 items — including item 2 (name it in
  the build-site-planner / site-architect prompt, or the planner will never select
  it, exactly as happened to the other five) and item 7 (links verified as served).

### 3b. DECISION-RECORD PAGE (`platform-log-index`) — still open, described below

Still `planned`, and it was **never in the site plan at all** (0
`site_plan_sections` rows) — that is why it never built; it is an absence, not a
failure. The self-correction page currently says the record is *"something you can
read"*, which overstates matters while no such page exists. Either build it or
soften that sentence.

**Deliberately not built by a thread: publishing our internal review records
outward is an owner call.** What it would contain, from data that genuinely exists
— **re-grounded 2026-07-27**, because these figures move daily:

- **177 council-gate decision notes** since 2026-07-17 (`doc_notes` where
  `categories ? 'council-gate'`); was 156 on 07-26.
- **42 commits** since 2026-07-01 carrying a `Council-Reviewed:` trailer, which is
  the exact commit↔verdict join.
- Verdicts across all 177 rounds: **42 APPROVED, 118 REVISE, 10 REJECTED**
  (7 unparsed). The failures are what would make the page credible rather than
  promotional — a review record showing only approvals is not evidence of review.
- **And the split that is the actual story**, because the decision rule itself was
  broken until 2026-07-22 (`bugs_closed/057` — objection severity was ignored, so
  approval was effectively unreachable):

  | period | approved | revise | rejected |
  |---|---|---|---|
  | before the 07-22 fix | **0** | 85 | 2 |
  | since the 07-22 fix | **42** | 33 | 8 |

  ```sql
  SELECT CASE WHEN created_at < '2026-07-22' THEN 'before' ELSE 'since' END AS period,
         substring(body from 'COUNCIL GATE — ([A-Z]+)') AS verdict, count(*)
    FROM doc_notes WHERE categories ? 'council-gate' GROUP BY 1,2 ORDER BY 1;
  ```

  Ninety-one consecutive rounds with **zero** approvals, because the reviewer of
  the reviewers had a bug. That is a better argument for the practice than any
  approval rate — but it is also the single most quotable line against us, and
  publishing it is exactly the decision being asked for.

The judgement the owner has to make is not technical: it is whether internal
review artefacts (objection text, seat names, what was rejected and why) go
outward, and in what redacted form. Same call applies to `tool-decision-record`
(also `planned`, never built). **The exact sentence the absence makes untrue** is
on the deployed self-correction page: *"Self-correction isn't a design principle we
claim — it's something you can read."* Nobody can read it. Softening that sentence
is a five-minute job and does not need the page.

## 4. Next actions, in the agreed order

**(0) ~~BUILD THE CHART COMPONENT~~ — DONE 2026-07-26, see §3a.** What it left
behind, in priority order:

- **`bugs_open/085` — HALF LIVE on v1.0.1173, and the live test FAILED.** Read the
  two dated sections at the foot of the bug file. The build path is fixed and in the
  running binary (pod-grep with controls); the **scoped section re-render path was a
  FOURTH drop point**, found by exercising the feature rather than trusting the
  deploy — `RerenderPageSectionsAction` never calls `BuildRenderContextAction`. That
  fix is written, tested, council round 3, and needs the **next** roll. Until it
  ships, the only verification route is a full page REBUILD, which regenerates copy.
  **OLD TEXT, kept as the correction:** The dated section at the foot of the bug file has the detail and the
  post-roll checklist. **This line used to say "one line in
  `BuildRenderContextAction`" and that was wrong** — the value is dropped at three
  points on one journey and fixing only the filed one-liner would have changed
  nothing visible. Split out at the council's request:
  **`bugs_open/109`** now owns the generic mechanism (four hand-maintained
  allowlists with nothing checking they agree; `theme_css`, `title` and
  `description` are dropped the same way today).
  Owed at the roll: pod-grep `resolveCurrentPageName` with a negative control,
  restore the `capabilities` placement, re-render **in scoped mode** (assemble
  redeploys stored HTML and would read as a false green), and induce the failing
  branch — a page matching no chart must render nothing, not everything.
- **`bugs_open/071` has a fresh instance and a sharper claim**: the index rebuild
  *authored* six broken links, the gate flagged all six as non-blocking warnings,
  and the 2026-07-25 repair did not survive the rebuild. "Link-sound" describes an
  artefact, not the site.
- **The dark-theme render is unverified** — leopardess has no `charts` key yet, so
  nothing dark has rendered. Do that check when it does.

**(a) ~~Measure the voice fix~~ — RE-MEASURED 2026-07-27, and the metric was
measuring the wrong thing.**

> **CORRECTED 2026-07-27.** This entry read: *"index 11 → 6, capabilities 6 → 6.
> Two components (`portfolio-showcase`, `hero-card-carousel`) hold 8 of the 12
> that remain, so a per-component fix now looks better."* Both halves are wrong in
> the same way. **`hero-card-carousel`'s four em-dashes are literals in its
> `html_template`, not the writer's output** (its `content_data` em-dash count is
> 0). So capabilities' "no improvement" is four characters no prompt could ever
> reach; the writer only wrote **two** there. The count in `rendered_html` sums two
> populations that no writer change can both address.

Site-wide today: **66 em-dashes — 23 baked into component templates, 43 written by
the content LLM.** The split query is in the RUNBOOK under *"Counting em-dashes so
the number means something"*. `from_words` is the only column a writer-prompt
change can move.

The 23 template-baked ones are in `tool-model-approach-selector` (12),
`tool-llm-cost-calculator` (5), `hero-card-carousel` (4), `image-hover-card-grid`
(1) and `evidence-chart` (1, mine — a shipped `<style>` comment). **The tool
components are generated**, so those are the tool-builder's model output frozen
into a template at generation time; the next generated tool reproduces them.

**Recommendation, given to the owner 2026-07-27: neither of the two options as
posed — do the per-component fix on the TEMPLATES and leave the writer alone.**
Reasons, in order: the writer's own rate roughly halved where it was re-run
(index 11 → 6 like-for-like) so the prompt is working; a mechanical post-pass over
`content_data` cannot touch the 23 template ones at all, and would be a
find-and-replace over copy that the claims and voice gates have already passed;
and `from_words` is spread across 14 sections with a long tail, which is the
profile a post-pass is worst at and a prompt is best at. Concentration is on the
template side — three components hold 21 of 23. Cheapest real win: strip the
literals from those three templates and fix the tool-builder's generation prompt
so new tools are born without them.

`That X matters` is 1, on `self-correction-leopardessconsulting`, unchanged and
pre-existing.

**Note while you are here:** `capabilities` flipped to `build_status =
'needs_rebuild'` at 2026-07-27 11:26 (not by this thread; the served page is 200
and carries the link repairs). A rebuild will re-author its copy — and, on this
site's record, is the moment to re-run the anchor-inclusive link crawl.

The original instruction, for the next measurement:

```sql
SELECT p.name, (length(string_agg(pc.rendered_html,'')) -
                length(replace(string_agg(pc.rendered_html,''),'—',''))) AS em_dashes
FROM pages p JOIN sites s ON s.id=p.site_id JOIN page_components pc ON pc.page_id=p.id
WHERE s.domain='fundamentallyai.com' GROUP BY p.name ORDER BY p.name;
```
Baseline 2026-07-25: index 6, about 8, capabilities 6, fine-tuning 5, council 2.
Also grep for `That X matters` (was 0 — keep it there). Full rationale and the
base64 apply/rollback procedure: `sql/README_writer_prompt_v3.md`. Backups
`bak_agent_definitions_pcw_20260724` / `_20260725` / `_20260725b`. **If the count
has not dropped, the alternative already offered to the owner is a mechanical
post-pass** — do not attempt a fourth prompt round without saying so.

**(b) `features_open/017` — component-adoption check.** The one genuinely
unfinished half of this workstream's own goal: the planner has **never chosen** any
of the five components. Every instance exists because we placed it. Acceptance item
2 in the PLAN (component named in the build-site-planner / site-architect prompt)
is still unmet. Spec is written; nothing built.

**(c) `features_open/018`** — specialist design critic (Gemini, rendered
screenshots). **(d) `features_open/019`** — sweep enrolment, LAST per owner ruling
(the improvement loop is off), though `bugs_open/071` sharpens why it matters:
enrolment is the *other* route to a durable record of findings, and this site has
**zero** discovery-check work items.

**(e) Smaller, unblocked:** imagery is still thin (10 of 43 components carry an
image, up from 2 of 27 — `bugs_open`-free, just work); the 3 empty
`planned` pages need type-specific builds; 24 of 25 anchored links fleet-wide
resolve to no `id` (recorded in `071`, needs components to emit ids or stop
emitting fragments).

## 5. Landmines — read before touching anything

**L1. The site serves `.html`. Every extension-less internal href 404s.**
`/capabilities` → 404, `/capabilities.html` → 200. This caused most of the 21
broken links.

**L2. A check that shares its logic with the fix cannot test the fix.** It bit me
twice in one day. My link census, my repair and my post-check all used
`href="(/[^"#?]*)"`, which excludes anchored hrefs — so 21 stayed broken while all
three agreed they were fixed. **Capture `href="(/[^"]*)"` then
`split_part(href,'#',1)`.** Verify against the SERVED page.

**L3. Dispatch latency is 7–9 MINUTES and a good dispatch is indistinguishable
from a dropped one for that whole window.** I declared four dispatches dead at +2
min, matched them to the real documented kcat/stdin race, and **published a false
landmine in two files**. Establish the baseline before calling anything dead; query
with a window starting BEFORE you dispatched;
`initial_request_data->'config'->'workflow'->>'start_step'` = `spawn_rerender`
identifies an 086-script row vs NULL for 049b/work-item.

**L4. No publish route beats a stalled dispatch lane.** 049b, the 086 envelope, a
`page_rerender` work item and `apply_section_edit` all share it. On 2026-07-25 it
stalled fleet-wide (99 `triaged`, **0** claimed, 0 claims in 15 min, trigger firing
on time) and nothing published until it recovered — then the **queued work item**
did the job. That is `bugs_open/030`, **OWNED** by another thread: contribute
measurements, do not fork a diagnosis. Run `scripts/who-owns.py <n>` before routing
work at any bug.

**L5. Never re-queue a historic work item — INSERT a fresh row.**
`stale-work-item-reaper` keys on `created_at`, so an old row is stale on arrival
(`bugs_open/070`). `created_by` is NOT NULL with no default. `unresolved` is in
`idx_swi_dedup`'s terminal set so the same `item_key` inserts cleanly. With fresh
rows, **batches are safe** — the reaper was the whole reason they used to park.

**L6. A page with no `site_plan_sections` rows cannot build** — it fails fast and
quietly at `mark_no_ready_sections` (~38s, no LLM spend), item → `needs_human_review`.
Check the plan before diagnosing a "failed" build.

**L7. `grep -rn` goes silent-binary** when another session leaves a NUL byte in the
tree — it returns nothing for a string that is present. **Use `git grep`.**

**L8. Rapid cache-busted probing throttles the origin** into `000`s and spurious
`404`s that read exactly like broken links. Retry 3× with a pause before condemning.

**L9. JS ships via `js_snippets`, NOT `content_components.js_content`** — the
latter publishes `/tools/assets/X.js` but injects no `<script>` tag (`bugs_open/041`
class, applies to section components too). Fire `site-asset-renderer` to rebundle,
and re-fire if a rebundle was queued before your `js_snippets` UPDATE landed.

**L11. `printf … | grep -q` under `set -o pipefail` reports FAILURE ON SUCCESS.**
`grep -q` exits at the first match, `printf` then takes SIGPIPE and exits 141, and
pipefail propagates 141. My live verifier reported FAIL for every marker it FOUND
and PASS for every negated check — a checker that inverts its own result. It had me
chasing a deploy bug for a page that was correct all along. **Use a here-string:**
`grep -q PATTERN <<< "$html"`.

**L12. Half the obvious CSS variables are defined by no theme.**
`--color-surface`, `--spacing-section` and `--container-max-width` do not exist in
any active `css_themes` row, so anything using them silently renders its fallback —
which is how a light card ships to a dark site. The real vocabulary is
`--color-background/-text/-text-muted/-primary/-secondary/-accent/-card-bg/-border`,
`--border-radius`, `--shadow`, `--spacing-xs…xl`. **Query `css_themes` before using
a variable**; a `var()` fallback is designed to hide exactly this.

**L13. A full rebuild undoes a per-page link repair, silently.** The index was
crawled clean on 2026-07-25; a rebuild on 07-26 reintroduced six broken links from
one component's card links, and the gate deployed them as warnings (`bugs_open/071`).
Re-verify links after ANY full rebuild — a previous clean crawl says nothing about
the page you just rebuilt.

**L14. Only 6 components can have their links repaired.** `resolve_internal_links`
acts on `ctaFieldNames` (`resolve_internal_links_action.go:98`) and its own comment
says a component missing from that map is "detectable but not repairable".
`info-card-grid` is missing from it.

**L10. Editing a `site_specs` aspect: supersede BEFORE inserting.**
`idx_site_specs_current` is UNIQUE `(site_id, aspect) WHERE is_current` — insert
first and you get 23505. Snapshot the row, `is_current=false` + `superseded_at`,
then insert.

## 6. Bugs and features this workstream filed (all OPEN)

| ref | one line | who should act |
|---|---|---|
| `bugs_open/070` | stale reaper keys on row age, kills every re-queue; DB-config fix, no roll needed | any thread; measurement in-file is inconclusive by design |
| `bugs_open/071` | the gate detects every broken link **then discards the finding**; warnings never persisted, fail-open justified by a loop that is off. Also: nothing validates the `#fragment` (24/25 fleet-wide dead) | platform thread; candidate 1 (persist warnings) is small and worth doing alone |
| `bugs_open/072` | `contact-info` reads flat `identity` keys the writer nests → 8 of 13 sites can never render a contact block; **new sites broken by default** | platform thread; pick ONE side of the contract, verify the resolver handles 3 levels first |
| `features_open/021` | operator bulk page rebuild — the road EXISTS and is dormant (`maintenance_queue` + active `maintenance-triage` + `PrepareRebuildDispatchesAction`, **no trigger, 2 rows ever, newest 2026-02-18**) | revive, do not rebuild |
| `features_open/016–019` | brief-fidelity audit (built+seeded), component adoption, design critic, sweep enrolment | this workstream, in that order |
| `bugs_open/085` | the render data advertises `current_page` and the build path always supplies it empty, so **no section component can know which page it is on**. One line in `BuildRenderContextAction`; inert until a roll | platform thread; the containment (index-only chart) is already applied |
| `features_open/023` | R4 now has a working instance; three findings from the build recorded into it, chiefly that **text inside `<svg>` is invisible to the claims gate** | whoever builds the R3 prompt layer |

Also contributed measurements into `bugs_open/030` (owned elsewhere) rather than
forking its diagnosis.

## 7. Standing rules that are NOT negotiable here

- **Never invent a person, client, case study or statistic.** Every claim needs a
  real source. The leopardess fabrications (invented founder, "70+ agents / 8
  departments", invented case studies, fake uptime and awards) must never resurface
  — they are in this site's `evidence_base` `banned_claims` by regex, plus
  `"5 agent roles"` which actually reached the live index before the gate existed
  for this site.
- Our own sites are never presented as a client roster.
- Copy goes through content-writer + `validate_page_content`, **never hand-authored
  into `content_data`** — that is how the claims and voice gates keep applying.
  (This earned its keep: the gate caught a real visible defect on the index
  rebuild.)
- Commit with an explicit pathspec, forward-only, never `git add -A`. Verify a
  deploy against the running pod, never git or the image tag.

## 8. Artefacts written by this workstream

`components/` — version-controlled source for all five components (template,
behaviour.js, input_schema, register.sql, snippet.sql) + `placements/` +
`README.md` with the acceptance checklist. `scripts/republish_page_086.sh` —
parameterised assemble-only republish. `sql/` — `055_seed_allowlist.sql` (+VERIFY
/+ROLLBACK), `evidence_base_fundamentallyai.json` (9 facts, 6 banned patterns),
`page_content_writer_prompt_v3_2026-07-25.txt` + `README_writer_prompt_v3.md`.
`agents/brief-fidelity-auditor.{config.json,seed.sql}`.

Added 2026-07-26 (later session): `components/evidence-chart/` — the shared chart
component (`template.html` + `input_schema.json` are the source; `register.sql` and
`update.sql` are GENERATED, never hand-edited), with `sample_data.json` exercising
the failing branches and a `README.md` holding the data contract and six traps.
`scripts/gen_component_register_sql.py` (regenerates both SQL files),
`scripts/render_component_template.go` (renders the real template against sample
data and asserts, per page case), `scripts/verify_evidence_chart_live.sh` (checks
the SERVED page against the register, sharing no logic with it).
`sql/evidence_base_charts_2026-07-26.sql` (+`_VERIFY`, 8 defect checks;
+`_ROLLBACK`), `sql/planner_prompt_evidence_chart_2026-07-26.sql`,
`components/placements/index_capabilities_evidence-chart.sql`.

Backup tables (do not drop without checking): `bak_pc_fai_links_20260725`,
`bak_site_specs_fai_evidence_20260726`, `bak_agent_definitions_sitearch_20260726`,
`bak_pc_fai_index_links_20260726`, `bak_cc_evidence_chart_pre_update`,
`bak_cc_portfolio_showcase_20260725`, `bak_site_specs_fai_identity_20260725`,
`bak_sps_fai_20260724`, `bak_agent_definitions_pcw_20260724/_20260725/_20260725b`.
