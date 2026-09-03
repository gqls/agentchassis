# HANDOFF — theme kits lane, 2026-09-02 (session "themes")

**Read this first, then §1's verification block before believing anything about liveness.**

The owner asked for a system of themes: reusable, named bundles of design defaults
(CSS, components, page structure, nav, typography, copy style) that a site may
*optionally* start from and then freely diverge from, creatable from example sites too.
Phase 1 is built and committed. **Whether any of it is live is UNVERIFIED — see §1.**

---

## 0. RESOLVED AFTER THE HANDOFF WAS FIRST WRITTEN — IT IS ALL LIVE NOW

§1 below was written while the kubeconfig token was expired. The token was refreshed,
the owner said go ahead, and the state is now settled. **Read this section, not §1's
uncertainty.**

**Binary: LIVE.** `agent-chassis` is on `v1.0.1355` and carries this work. The
provenance log line had already scrolled (a startup line — expected), so this is a
capability probe of `/proc/1/exe` **with both controls**:

| needle | result |
|---|---|
| `fork_theme_from_site` (positive control — pre-existing action) | PRESENT |
| **`apply_theme_kit`** | **PRESENT** |
| **`page_archetypes`** | **PRESENT** |
| `zzz_not_a_real_action_zzz` (negative control) | absent — *the probe discriminates* |

**Schema: APPLIED, 2026-09-02.** Migrations `689` and `691` applied via a **scoped** run
(`MIGRATIONS_DIR` pointed at a temp dir holding only those two files) because
`--apply` takes EVERY pending file and would have swept a dozen other lanes' migrations.
Verified after: `to_regclass` returns both tables; **4 kits and 14 fleet archetypes**
seeded; `validate_resolved_composition_spec` accepts `layout_source='theme_kit_default'`.

**691's three live sites: verified appearance-neutral at the artefact.** Its premise was
re-read immediately before applying (served CSS still byte-identical to
`reference_values` for all three), and re-read after: `cv1.co.uk`, `finetuning.uk` and
`gaswholesalers.com` now hold their own `site_split` palette rows and serve exactly the
same colours as before.

**Ordering hazard closed correctly.** The Go half emits `layout_source='theme_kit_default'`,
which the old validator would have REFUSED. It could never have fired before the
migration (no `theme_kits` table → no adoption spec → the layout rung never returns that
source), and 689 widened the validator before any kit could exist. ⚠ **An image rollback
to before this build is safe; a DATABASE rollback past 689 while this binary runs would
break composition installs for themed sites.**

**Council gate: SUBMITTED**, correlation **`bed139b2-f512-436a-9ba8-ff2fbfade8ef`**.
Verdict not read (~30 min queue). Resolve it with:
```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = 'bed139b2-f512-436a-9ba8-ff2fbfade8ef';
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```
The submission deliberately names the two edits I most want challenged: the layout
short-circuit (it skips tag matching AND the `needs_new_layout_candidate` signal) and
the `generic_theme` suppression. **Do NOT write `Council-Reviewed:` until you have read
an approved verdict** — 098 buckets an unread claim as MISMATCH.

**What is NOT live and did not change:** nothing has adopted a kit yet
(`theme_kit_adoption` specs: 0). The four seeded kits exist and are selectable; no site
uses one. Findings §3(a)-(d) all still hold — in particular a kit still cannot deliver
colour to a site.

## 1. ⚠ (superseded by §0, kept for the method) FIRST ACTION: establish liveness. I could not.

A fresh chassis build was deployed near the end of this session. **I could not verify
what it carries** — the kubeconfig token expired (`error: You must be logged in to the
server (Unauthorized)`; the owner refreshes these, ~3-day expiry). Do not assume either
way. Run these before acting:

```bash
# 1. Does the running binary carry apply_theme_kit?
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=400 | grep -m1 'build provenance'
git merge-base --is-ancestor 0902039c0 <the stamped sha>   # my Phase 1 commit

# 2. Do the tables exist? (migrations do NOT ride a build — separate action)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -c "SELECT to_regclass('public.theme_kits'), to_regclass('public.page_archetypes');"
```

**As of my last successful check (~17:00Z) both were NULL — migrations 689 and 691 were
committed but NEVER APPLIED, and no roll had carried the code.** If the binary now has
the code but the tables do not exist, the system degrades safely by design:
`loadSiteThemeKitDefaults` errors are swallowed and every consumer falls through to
existing behaviour. Nothing breaks; nothing works either.

**Do not repeat my mistake:** I wrote "deployed … live since commit 0902039c0" into the
concept register when neither half was true, and each half took one command to check. A
Fable review caught it, not me. The migration runner's dry-run prints *"ran to its own
COMMIT without error (everything rolled back)"*, which reads like success and means the
opposite.

## 2. What exists, and what it is

**`theme_kits`** — a named registry bundling EXISTING library rows by FK (layout,
palette, typography_set, header/footer components) plus an `extras jsonb` slot. Named
`theme_kits`, deliberately **not** `themes`: `css_themes`/`theme_id`/
`needs_theme_review`/`forked_from_theme_id` already mean "one site's CSS composition
record" throughout the tree, and a second `themes` would make every `theme_id`
ambiguous. "Theme" stays the user-facing word.

**`page_archetypes`** — reusable, forkable per-page-type section lists, three-way scoped
(site row > theme-kit row > fleet row, CHECK-enforced). Replaces the hardcoded
`defaultSectionsForPage` Go switch, which is kept as a logged last-resort fallback.
**`UNIQUE NULLS NOT DISTINCT`** is load-bearing — with a plain UNIQUE the constraint
could never fire (every row has a NULL in the key) and the seed's `ON CONFLICT` was dead
code. Proven by induced fault.

**`apply_theme_kit`** — materialises a kit into a site's own rows and queues
`needs_composition`. **It never creates a live binding**; a themed site can diverge on
any field immediately, and nothing checks "is this site themed" on any edit path.

### The owner's ruling that governs everything here (2026-09-02)

> *"I think the classifier can be given the choice. I think by default it can start with
> a theme and change it if it wishes, but it must have full authority to ignore our set
> of themes if it chooses."*

So `apply_theme_kit`'s default mode is **`start`** — it WRITES the kit's palette/
typography, superseding what is there, because a theme that defers to whatever exists is
a no-op on any classified site. Values are marked `reference_source: "theme_kit:<name>"`
and `reference_is_default: true` so a later reader can tell a default from a decision.
`fill_gaps` remains as an explicit conservative mode; `reapply` also replaces an
installed composition. **The only thing no mode overwrites is `design_intent.<dim>.locked:
true`** — a human pin nothing sets automatically. An in-data key, NOT `site_specs.pinned`,
because `pinned` is lost on supersede (observed in production this session).

## 3. THE FINDINGS THAT MATTER MOST — several deflate this lane

Stated first because they should shape what the next session does, and two of them
reduce this lane's own value.

**(a) A kit cannot deliver COLOUR to a site. Measured at the artefact.** gamedesign.uk
resolved a deliberately hand-chosen palette at composition (`palette_source=mission_hint`,
the first time that rung has ever fired in fleet history) and served **none of its eight
core colours**. `render_css_from_spec` makes the 8 core slots spec-wins and
`analyze_design` reads `design_intent`, never the composed palette row. **The palette
cascade decides only SPECIALISED slots for any site that gets a design overlay.** The
lever on served colour is the BRIEF. This is the ruling working, not a defect — do not
file it as one, and do not plan colour differentiation through kits, palettes or pins.

**(b) `page_archetypes` governs at most 1 page in 18.** [MEASURED] 1,022 of 1,083 live
pages (**94.4%**) do not match any exact `defaultSectionsForPage` output; 5.6% is an
UPPER bound on fallback-fed since a planner can choose those lists unaided. The
components lane challenged this and was right. **The structure lever is the planner's
prompt.** The hints half of the design (`load_theme_structure_hints`) is unbuilt and
would be advisory anyway.

**(c) Chrome is the strongest available lever and it is UNPROVEN.** 36 of 37 sites render
`site-header`+`site-footer`; 10 chrome-eligible functions exist with 4 headers and 1
footer entirely unused; `ChromeSlotFunction()` hardcodes slot→function so every site asks
for the same one. The documented escape is a `style_collections.header_component_id` pin
— **but all 6 existing pins point at the same component the default picks**, so "pin
honoured" and "pin ignored, default coincides" are indistinguishable in current data. A
decisive one-site test is queued (§5).

**(d) Layout is real, working, and under-used.** 9 of 18 layouts in use, 73% of sites on
three, nine never used. Layout drives `layouts.css_template` and is NOT overwritten by
the overlay (structure tokens are layout-only). Best visible-difference-per-effort lever
available today, and it needs no theme kit — it is driven by classification tags and
`design_intent.style_direction` prose in the brief.

## 3a. KNOWN DEFECTS IN WHAT SHIPPED — found after apply, none yet fixed

**(i) The seed set narrows toward the layouts that are already dominant.** The four
seeded kits name `brochure-formal`, `docs-sidebar`, `soft-editorial`,
`tool-portal-light`. [MEASURED 2026-09-03] `tool-portal-light` is on 14 sites,
`brochure-formal` 6, and `magazine-grid` (8) sits outside the set; **six layouts have
zero sites and none is in a kit.** So if adoption rises, layout selection narrows from
"18, tag-matched" to "4, kit-selected" — *toward* the concentration the owner's sameness
directive is complaining about. **As seeded, the kits are a sameness risk wearing a
differentiation label.** Free to fix (adoption is 0, nothing depends on the seeds), but
*which* looks become kits is curation, not mechanics — it is an owner/design call, not
something to pick by taste. **`bugs_open/445` is producing the input that makes it a
real decision**: splitting the never-used layouts into "correctly unused, no site of
that shape exists" versus "reachable but losing". Only the second kind is worth a kit.
Wait for that list rather than guessing; seeding a kit for a layout nothing needs would
repeat this defect in the opposite direction.

**(ii) The layout rung records a candidate that was never scored.**
`resolve_composition_layout_action.go` returns `"candidates": []string{kitLayoutName}`,
and `install_site_composition_action.go:637` writes it through as
`lineage.layout_candidates`. A themed site would therefore record
`layout_candidates: ["soft-editorial"]` — reading as *"one candidate was considered and
won"* when **no candidate was scored at all**. This is the same false-structured-fact
class I fixed for `layout_source` in this very session, reintroduced one field over in
the same edit. The honest record is an empty list, with `source: "theme_kit_default"`
carrying the story. **Not yet emittable** (adoption 0), needs a one-line change plus a
roll; rides the next one. `bugs_open/445` has been told to design their fit-evidence
against this being empty/absent, not a one-element list.

**(iii) "A chrome pin is an available-now lever" was half wrong — RETRACTED to both
lanes that received it.** A pin **selects** a component; nothing **populates** it.
Measured: `header-with-categories` needs ~12 template variables including
`action="{{.search_action_url}}"`; `header-minimal-tool` needs tool vocabulary
(`tool_status_label`, `avatar_initials`); `header-with-cart-or-nav` needs cart
vocabulary — and designblog.co.uk's header `content_data` is **empty, zero keys**. An
unsupplied variable renders blank: a form action that posts to the current page, empty
aria-labels, missing nav. **So the truer explanation of "36 of 37 sites render identical
chrome" is not that nobody selected — it is that nobody ever supplied the data for
anything else.** That is a bigger job than a per-site UPDATE and the 18 remakes should
be sized on it. Correction sent to `portfolio_positioning` (whose RUNBOOK §5 carried the
incomplete version) and to `designblog.co.uk` (who were about to pin one on a live site
today).

## 4. Commits (all on `087_towards_multiple_domains`)

| commit | what |
|---|---|
| `0902039c0` | Phase 1: registry + page_archetypes + apply action + layout rung |
| `67433f907` | optional-key-budget literal for `apply_theme_kit` (RFC_022/WFA-013) |
| `18d877772` | RFC_059 filed (DRAFT) |
| `8a9a0c865`, `0bdd4575b` | concept register DES-085 + index row |
| `4d3616b78` | **migration 686→689 collision fix + retracted DES-085's false "deployed"** |
| `5efe434c9` | **eight defects from two Fable reviews** (schema fixed while unapplied) |
| `daf08b391` | RFC_059 review objections recorded in-file |
| `b995dfd7d` | owner ruling implemented (`start` mode) + RFC_059 WITHDRAWN + migration 691 |
| `3129f2e95` | register records the ruling |
| `a648ae166`…`ffb0cc41e` | `bugs_open/430` (JS fork drop) filed + 3 corrections |
| `8000652ba`…`32b20d7d8` | `bugs_open/438` filed + 5 corrections |
| `0bca5d510` | LANDMINES: reference_values is NOT a pin |
| `b6039c26b` | CONTRIB to portfolio_positioning (differentiation levers) |

## 5. OPEN — decisions and pending work

### Owner decisions
1. ~~**Apply 689 + 691, and roll?**~~ **DONE 2026-09-02 — owner said go ahead.** Applied
   scoped, verified at the artefact. See §0.
2. ~~**Council round**~~ **DONE — submitted**, correlation
   `bed139b2-f512-436a-9ba8-ff2fbfade8ef`, verdict unread. See §0.
3. **`bugs_open/438`: retire or build?** ← **STILL OPEN** Both lanes agree the capability does not exist
   and neither will choose. Retire the dead rung, or give `082` a structured preference
   input (`--palette`). **Note (a) above: even building it would not put a colour on a
   site**, because the overlay overwrites core slots. So "build" only makes sense
   alongside a render-merge change — which is RFC_059, withdrawn.
4. **The seam question**, raised by gamedesign.uk, filed by nobody: should
   `resolved_composition` — schema-validated, with an enforced lineage enum — describe
   core colours the public never sees?

### Pending experiment — the chrome question
`portfolio_positioning` will run it at **remake №5** (held behind `bugs_open/444`). Full
recipe with resolved UUIDs, pre-flight queries and served-diff commands is **§5 of
`docs024_key_docs_latest/portfolio_positioning/RUNBOOK_remake_release.md`** (their commit
`4a1aa5310`). **Read the three-way result, not pass/fail** — the four alternative headers
are `*_pre_037` legacy rows never rendered on any live site, so a failure could mean
"pin ignored" OR "component stale":
- header changes and looks right → mechanism real, recipe safe for the other 17
- header unchanged → **pin ignored**, chrome differentiation needs a code change
- header missing/broken → pin honoured, component stale; retry another before concluding

**I owe `portfolio_positioning` and `vetcomparison` a ping with the outcome either way.**

## 6. Cross-lane state — commitments made in my name

- **`gamedesign.uk`** — built fresh with a palette I specified; agreed to be the first
  consumer of a seeded `practice-editorial` kit when kits go live. Site is live. Its
  served colours are the overlay's, not the seed's (finding (a)).
- **`portfolio_positioning`** — has my CONTRIB and the chrome recipe in their runbook;
  layout + colour-by-referent go into the next 18 remake briefs verbatim.
- **`designblog.co.uk`** — ACKed the owner's sameness directive; has the measurements.
- **`vetcomparison`** — told a kit is NOT a fit (3 reasons). Gave them a diagnosis
  instead: their accent-contrast defect is a **consumption** problem —
  `--color-accent-ink: #0f172a` is computed and served and consumed ZERO times, while the
  raw accent is used as text 4×. Converged independently with the site design planner's
  read. Fix belongs to component templates, not the palette.
- **`site_design_planner`** — owns the composition resolvers; tracking 438; refuted my
  fix candidates.
- **`webdesign-tool-rebuild(s)`** — co-diagnosed `bugs_open/430`.

## 7. What I got WRONG this session — calibration for whoever picks this up

Recorded because the pattern matters more than the individual errors, and every one was
caught by someone else.

1. **DES-085 said "deployed"; nothing was.** Both halves false, one command each.
2. **Commit message attribution backwards** — I said I'd picked up another session's
   change; in fact their pathspec commit swept MY half-written rung, and HEAD did not
   compile for 51 seconds. Corrected in the plan, not amendable (forward-only).
3. **§6's tripwire over-stated** — I warned a fix would overwrite a hand-seeded row;
   `siteSpecDeepMerge` means it would not. The real hazard is narrower (a writer that
   bypasses `write_site_spec`).
4. **Both my leading fix candidates for 438 were refuted** — and candidate 2's
   disqualification was a parenthetical I wrote *in the same sentence* and ranked anyway.
5. **I told gamedesign.uk `mission.preferred_palette` was the reliable lever.** It was —
   for the composition — and had no effect on the site.

The through-line: **a confirmed diagnosis is not a working fix, and a successful test
makes the wrong fix look safer.** Both lessons are in
`memory/measurement-discipline-index.md`. The single most valuable act of the day was
another lane tracing my leading fix before applying it.

## 8. Standing docs owed

This lane has only this handoff. Per CLAUDE.md a workstream keeps five living docs
(PLAN / RUNBOOK / NOTES / README_where_we_are / SUMMARY) created at the START. They were
not — the work ran through the approved plan file at
`/home/ant/.claude/plans/please-think-hard-about-starry-locket.md`, which carries the
full design plus corrections **C1–C10** and is the best single companion to this handoff.
Whoever continues should create the standing five here and migrate the plan's content in.
