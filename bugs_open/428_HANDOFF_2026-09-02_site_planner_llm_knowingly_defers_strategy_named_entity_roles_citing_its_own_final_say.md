# 428 — the site-planner LLM sees `entity-page`/`entity-directory` named in the strategy, understands them, and deliberately leaves them out — this is licensed behaviour, not a drop, and closes diagnosis `d6d350ec`

Filed 2026-09-02 by the `gap planner` session, resolving the open question left by an
already-run diagnosis (correlation `d6d350ec-e16b-4792-9282-ca5155369791`, `site_work_items`
row `74061968-9dfc-442d-be2e-9da3ceda7f08`) that `bugs_open/427` §6 pointed at and asked not to
be duplicated. That run reached its iteration cap at verdict `UNVERIFIABLE` / "NOT CONFIRMED":
it confirmed the *symptom* (boxingonline.com's strategy names `entity-page`/`entity-directory`
with full reasoning; the plan the site-planner wrote carries neither; other sites' plans do
carry both roles) but never obtained the one citation it still needed — whether
`recommended_page_types` is even wired into the site-planner LLM's own prompt. **It is, and the
LLM reads it correctly.** The gap is not a code drop.

## 0. First-hand verification, per CLAUDE.md's 2026-07-31 ruling

This supersedes further `090` iterations on `d6d350ec` rather than re-running the loop: every
citation below is a direct read of the live `agent_definitions` config, the live
`platform/orchestration` source, and the live `llm_call_log` table (both `prompt_rendered` and
`response_text` for the actual boxingonline.com call), done today, 2026-09-02. This clears the
loop's own confirm bar — a static citation of the mechanism AND a runtime citation of it
occurring — which the automated run never reached.

## 1. The mechanism, static half

`agent_definitions` type `build-site-planner`, `default_config.workflow.steps.plan_site` (an
`execute_llm_prompt` step). Its `prompt_template` includes:

```
## Domain Strategy
{{if .site_specs.specs.strategy}}{{.site_specs.specs.strategy}}{{else}}No strategy data available…{{end}}
```

`.site_specs.specs.strategy` is populated by the `read_specs` step
(`platform/orchestration/actions/site_spec_actions.go:490-534`, all-aspects mode of
`ReadSiteSpecAction`): `specs[aspect] = data`, i.e. the **entire** `site_specs.data` JSON object
for `aspect='strategy'` — which is exactly where `recommended_page_types` lives (confirmed by
`d6d350ec`'s own iteration 3: `data->'recommended_page_types'` returns the array with per-type
reasoning). So the field is in scope for the template.

Rendering is plain Go `text/template`
(`platform/orchestration/datahelpers/data_helpers.go:1129-1184`, `RenderPromptTemplate` —
`template.New("agent_prompt").Funcs(funcMap)`, `tmpl.Execute(&buf, data)`). The funcMap
registers a `toJSON` helper (`data_helpers.go:1132`), but the Domain Strategy block does **not**
call it — it interpolates the raw map with bare `{{.site_specs.specs.strategy}}`. Go's default
`%v` formatting of a `map[string]interface{}` containing nested `[]interface{}` of
`map[string]interface{}` produces an unquoted, unstructured literal dump — confirmed below, this
is a real quality defect, but **not** the root cause (§4 shows the model reads it anyway).

Two sentences later in the same prompt, under "## Strategy Guidance":

> "The recommended page_types should inform which pages you plan" (**advisory**, not a
> requirement)

and, further down, under RULES/general instructions:

> "You have FINAL SAY on architecture. If you disagree with the strategy, go with your judgment
> but note why in strategy_notes."

**This is the actual licensing mechanism.** The prompt does not merely fail to enforce
`recommended_page_types` — it explicitly authorises the model to diverge from it, on the sole
condition that it explains itself in `strategy_notes`.

## 2. The mechanism, runtime half — boxingonline.com's actual call

`llm_call_log` id `63d9b370-3f2d-41db-82b9-1dfed5204225`, `agent_type='build-site-planner'`,
`step_name='plan_site'`, `created_at=2026-08-31 12:36:44 UTC` — confirmed as boxingonline.com's
call (`prompt_rendered` contains "Plan a website for boxingonline.com").

`prompt_rendered`, the "## Domain Strategy" block, verbatim excerpt:

```
recommended_page_types:[map[page_type:index reasoning:…] map[page_type:blog-index reasoning:…]
map[page_type:blog-post reasoning:…] map[page_type:tool reasoning:…]
map[page_type:entity-page reasoning:Fighter profile pages — one per fighter mentioned across
articles — aggregate all site content about that boxer, display their current record, and link
to upcoming fights on the calendar. …]
map[page_type:entity-directory reasoning:An event directory — one page per major upcoming fight
— gives each bout a permanent URL with full details: fighters, date, venue, broadcast,
undercard, and a brief preview. …]]
```

Confirms: the data reached the model, ugly-but-legible, exactly as claimed in §1.

`response_text` for the **same call** — the model's own JSON output:

- `pages[].page_type`: `[index, blog-index, tool, content, content, blog-post]` — no
  `entity-page`, no `entity-directory`.
- `strategy_notes` (verbatim): *"…The strategy recommends fighter entity pages and event entity
  pages — these are best served by the blog-post and entity-page types respectively, but at
  launch I'm keeping the structure lean: homepage, articles listing, fight calendar, about, and
  contact. Fighter profiles and event pages will grow organically as content volume
  increases…"*

This closes the loop completely: the model **read the recommendation, correctly named the page
type (`entity-page`) it maps to, and consciously deferred it** — exercising precisely the
license §1 grants, and discharging precisely the obligation ("note why") that license attaches.
This is not a bug in the classic sense of a dropped value; it is designed, working-as-instructed
deferral with no downstream mechanism that ever revisits a deferral (§4).

## 3. This is fleet-wide, not boxingonline-specific

`[MEASURED 2026-09-02]` — sampled the 60 most recent successful `build-site-planner`/`plan_site`
`llm_call_log` rows (2026-05-14 through 2026-08-31; a recency-biased sample, not a full census —
mark any extrapolation beyond it `[UNMEASURED]`):

| | count |
|---|---|
| sampled calls | 60 |
| strategy names `entity-page` and/or `entity-directory` | 33 |
| — of those, `response_text` did not parse as clean JSON (excluded, `[UNMEASURED]` why — likely fenced/multi-part responses, not investigated) | 8 |
| — evaluable (parsed cleanly) | 25 |
| — evaluable calls where the output omitted at least one named role | **19 (76%)** |
| — evaluable calls where BOTH roles were named and BOTH were omitted | 3 |

Sites in the sample showing this pattern include (by their `strategy_notes`, which name the
domain or its business directly): `boxingonline.com`, `farmerinsurance.uk`, the `loanzy.uk` /
loan-calculator cluster, `dartsonline.com` (multiple historical calls), the `gamedesign.uk` /
`gamesdesign.co.uk` cluster, `mortgagecalculator.co.uk`. **This is the estate's normal planner
behaviour wherever a strategy names an entity role, not a boxingonline defect.**

## 4. Corroborating evidence — a detector for this already exists and already found it, independently

`[MEASURED 2026-09-02]`: `site_work_items` carries **13** rows, `item_type IN
('needs_content_page','needs_content_planning')`, `status='deferred'`, summary prefixed
`"[verdict, not dispatched]"`, each describing this exact shape — a strategy naming
`entity-directory`/`entity-page` (recipe libraries, brand directories, an FCA-broker directory,
and boxingonline's own event-directory/affiliate gap) with nothing built. Boxingonline's own row:
`e3c2b440-c006-40ec-be7a-88d0b689ed1e`, created 2026-09-01, names the missing "event-specific page
with broadcaster details embedded as affiliate entry points" and "no entity-directory page type"
directly. **A gap-detection mechanism already runs and already caught this** (independently of
`d6d350ec` and of this filing) — every one of the 13 rows is captured as a verdict and **none is
dispatched into a build action**. The detector is not the missing piece; the consumer is.

## 5. What this is NOT

- **Not `bugs_open/206`'s defect** — and a correction to this section, found 2026-09-02 while
  answering a peer's question about this bug: the line originally here cited 206's **pre-fix**
  2026-08-06 state ("the `directory-listing` component has never shipped on a live page
  anywhere") as if still current. It is not — 206 itself records closure evidence dated
  2026-08-08: `directory-build-handler` is live and proven (vetcomparison.uk's own
  `directory-index` page serves 61 real practices, sourced from `business_intel`, via
  `ensure_page_section_layout` + `queryresolve.resolveBusinessDirectory`). So a real builder for
  `page_type='entity-directory'` now exists — this bug is still upstream of 206 (the role is
  often never *planned*), but "no builder exists" is no longer the downstream blocker to name.
  **The real downstream constraint, `[MEASURED 2026-09-02]`, is DATA, not mechanism**:
  `business_intel.business_verticals` covers exactly three verticals (`veterinary`,
  `online-pharmacy`, `seaweed-farming`). None of the sites in §3's sample — boxingonline's
  fights, the recipe-library sites, the brand-directory sites, the FCA-broker-directory site —
  has a matching vertical. `directory-build-handler` fixed the *mechanism* for the shape 206 was
  filed against; it does not by itself give any of §3's sites something to populate a directory
  *from*. That gap is `bugs_open/427`'s territory (no writer turns real-world facts into
  structured, dated records), not 206's.
- **Not a downstream code drop.** `d6d350ec` iteration 2 already established
  `write_site_plan_action.go`/`page_role_validator.go` neither rewrite nor discard these role
  names once the LLM emits them — that finding stands and is now explained: the LLM usually
  doesn't emit them to begin with.
- **Not (solely) the ugly prompt formatting.** §2 shows the model reads the Go-map-dump
  correctly regardless — it named "entity-page" precisely. The formatting (§1) is a real,
  independent quality defect worth fixing, but it is not why the roles get dropped.

## 6. Fix candidates — named, not decided

1. **Tighten the license.** Currently "should inform" + "final say, note why" allows silent or
   soft-explained divergence. Options range from making unexplained omission a validation
   failure in `validate_site_plan` (the very next step — already validates component names, and
   has the recommended_page_types list from the same collected_data available) to simply
   requiring `strategy_notes` to address every named `page_type` in `recommended_page_types` by
   name, not just the ones the model chose to keep.
2. **Wire the existing detector to a dispatcher.** Lowest-risk, no planner-prompt change: the 13
   `needs_content_page`/`needs_content_planning` deferred verdicts (§4) are already correct,
   already scoped, and already sitting idle. Whatever currently produces them but stops at
   `deferred` needs a next step that turns the verdict into a `needs_page` (or similar) work
   item — this alone would fix every site already caught, boxingonline included, without
   touching the LLM's behaviour on new builds.
3. **Fix the prompt formatting** — `{{toJSON .site_specs.specs.strategy}}` instead of the bare
   interpolation (§1). Cheap, does not by itself fix the omission rate (§2 proves the model
   already parses the ugly form), but removes a real confound for the next thread that has to
   audit "did the model see X" — this filing needed a `llm_call_log` pull to answer that
   question; a clean prompt block would make it legible from the prompt text alone.
4. **Sequence with `bugs_open/427`, not `206`** (corrected §5, 2026-09-02): 206's builder
   mechanism is already fixed and live, but it only has data for three `business_intel`
   verticals — none of §3's affected sites. Dispatching a deferred verdict for boxingonline,
   a recipe site, or a brand-directory site (candidate 2) will hit an empty-data wall unless
   427's populator gap closes first, or the specific site's data genuinely lives somewhere
   `directory-build-handler` can already reach — check per-site before assuming dispatch alone
   is sufficient.

## 7. How to verify a fix

- For (1): re-run `plan_site` against a strategy naming `entity-page`/`entity-directory` and
  confirm the output either includes the role or `strategy_notes` names it explicitly per-type
  (not the current free-text "I'm keeping it lean" style that can omit a role silently).
- For (2): confirm a `needs_content_page`/`needs_content_planning` `deferred` row transitions to
  a dispatched status and produces a real page/work item, using boxingonline's own
  `e3c2b440-c006-40ec-be7a-88d0b689ed1e` as the worked case.
- For (3): pull a fresh `llm_call_log.prompt_rendered` for `plan_site` and confirm the "##
  Domain Strategy" block is valid JSON, not a Go map dump.

## 8. Cross-references

- `bugs_open/427` §6 — the dated-event-facts bug whose fix candidate #3 (entity-directory as
  render target) was explicitly blocked pending this diagnosis; cross-reference back there.
- `bugs_open/206` — the downstream builder gap; related, distinct mechanism, see §5.
- Diagnosis this closes: correlation `d6d350ec-e16b-4792-9282-ca5155369791`,
  `site_work_items` id `74061968-9dfc-442d-be2e-9da3ceda7f08`.
- `site_work_items`, `item_type IN ('needs_content_page','needs_content_planning')`,
  `status='deferred'`, summary `LIKE '[verdict, not dispatched]%'` — the 13 rows from §4,
  including boxingonline's own `e3c2b440-c006-40ec-be7a-88d0b689ed1e`.
- `llm_call_log` id `63d9b370-3f2d-41db-82b9-1dfed5204225` — the worked boxingonline.com case,
  both `prompt_rendered` and `response_text` cited in full in §2.
