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

> **CORRECTION 2026-09-02 (resuming session), load-bearing for §6 candidate 2: this is NOT an
> unbuilt consumer — it is RFC_056, a deliberate owner-ruled circuit breaker, checked at the
> artefact rather than assumed.** All 16 rows this section's filter actually returns (13 after
> excluding 2 keyword false-positives that mention "entity-page" only in passing — see below)
> carry `spec->>'filing_mode' = 'record'`, `[MEASURED 2026-09-02]` on every one of them, e.g.
> boxingonline's own `e3c2b440-c006-40ec-be7a-88d0b689ed1e`: `created_by='offer-analysis'`,
> `handler_agent=''`, `spec.filing_mode='record'`, `spec.routed_handler='page-build-handler'`.
> That shape is `write_audit_findings_action.go`'s `recordOnlyFinding` (RFC_056, shipped
> 2026-08-25): an LLM-audit-seat finding filed as a **verdict row nothing can dispatch by
> construction** — `detected-item-promoter` and `triage_detected_items` both refuse it
> (`handler_agent=''`), and the routing is preserved only in `spec.routed_handler` for a human
> or a later migration to release by hand via `spec.release_recipe`.
>
> **Why, in the action's own words**: *"The LLM audit seats of the improvement loop … file
> findings that are OPINIONS about pages that already work ('aspirational improvements') …
> whose handlers REGENERATE the page … The owner's ruling: keep the seats — they are the site
> acceptance council — but stop the rewrites."* The motivating incident is closed and dated:
> `bugs_closed/238` — an `offer-analysis`-family audit's dispatched finding regenerated
> finetuning.uk's homepage case-studies section and shipped five `<img src="">` on a live page,
> asked for by nobody, destroying the original card copy in the process (not restorable by
> pasting URLs back). RFC_056 shipped eight days after that incident, specifically to stop this
> class of automatic dispatch.
>
> **This means §6 candidate 2 as worded — "wire the existing detector to a dispatcher" — would
> REBUILD the exact promoter the owner ruled out, for the exact class of finding (an LLM audit
> seat's opinion) that caused the incident RFC_056 exists to prevent.** It is not the lowest-risk
> candidate; on the evidence here it is the one candidate that directly reverses a recent,
> named, owner-authored safety ruling. A fix along these lines needs to go THROUGH that
> ruling, not around it — e.g. a human-reviewed release surface using `spec.release_recipe` as
> designed, or an owner decision to carve out a narrow, well-evidenced exception for this
> specific shape (unambiguous strategy-named role, zero component built) — not a general
> promoter for `filing_mode='record'` rows. Read `write_audit_findings_action.go`'s full "WHY IT
> EXISTS" comment before building anything here.
>
> Also corrects §4's own count in passing: **13, not this section's originally-implied "every
> matching row" —** a plain `ILIKE '%entity-page%' OR ILIKE '%entity-directory%'` on the summary
> returns **16** `[MEASURED 2026-09-02]`; two of those three extra rows are keyword
> false-positives for a DIFFERENT finding (one about catalogue depth mentioning "entity-page
> volume" in passing, one about imagery compliance on an entity-page template that already
> **exists and is live on 8 pages** — not a never-planned-role instance at all). Hand-verifying
> is what narrows 16 back to something close to this section's original 13; a keyword count
> alone over-states the population by ~20% here.

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
2. ~~**Wire the existing detector to a dispatcher. Lowest-risk, no planner-prompt change.**~~
   **CORRECTED 2026-09-02 — this is NOT lowest-risk; see §4's correction.** These 13 rows are
   `filing_mode='record'` verdicts (RFC_056, 2026-08-25) — a deliberate owner-ruled circuit
   breaker stopping automatic dispatch of LLM-audit-seat opinions, built specifically because an
   earlier auto-dispatch of exactly this kind of finding damaged a live page
   (`bugs_closed/238`). A general promoter for this `(item_type, status)` pair is also far too
   broad regardless — **1,284 rows total** carry `item_type IN
   ('needs_content_page','needs_content_planning') AND status='deferred'`
   `[MEASURED 2026-09-02, gap_planner]`, not just the 13; most are unrelated deferrals for
   unknown reasons. Any fix here must (a) filter to the exact narrow shape both filing sessions
   hand-verified (the `[verdict, not dispatched]` + named-entity-role summary text, not the
   item_type/status pair alone), and (b) go THROUGH RFC_056's release path rather than around
   it — a human-reviewed release surface keyed on `spec.release_recipe`, or an explicit owner
   decision to except this specific shape, not a standing automated promoter. Read
   `write_audit_findings_action.go`'s "WHY IT EXISTS" comment in full before building anything
   under this candidate.
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
- `docs024_key_docs_latest/bugfix_427_event_render/HANDOFF_2026-09-02_continue_here.md` —
  the cross-bug continuation handoff (this bug + 427), written for a fresh session.

## 9. Status update, 2026-09-02 (same day, resuming session) — fixed, narrower than first proposed

Put to the user given the stakes (§4's correction): fix the two safe, independent candidates,
and build a human-reviewed release surface rather than an automated dispatcher for candidate 2.
Approved, and shipped same session:

- **Candidate #3 + the softer form of candidate #1 — SHIPPED, live.** Migration `687`
  (`docs/agent_docs/sql_for_agents/687_build_site_planner_strategy_json_and_omission_reason.sql`):
  the Domain Strategy block now renders via the already-live `toJSON` template function instead
  of a bare Go-map dump, and the FINAL SAY rule now requires every omitted named
  `recommended_page_types` entry to be individually named in `strategy_notes` with a real
  per-type reason — not the generic "keeping it lean" framing boxingonline's own call used to
  cover both dropped roles at once. Deliberately NOT a `validate_site_plan` hard failure — the
  model keeps its licensed final say; this only makes the existing "note why" obligation
  concrete enough to audit. Applied and verified live at the artefact (re-pulled the prompt,
  confirmed both new strings present). `snapshot_agent()` taken first; rollback file paired.
  Council-Submitted: `3f9cdfea-7287-4ab3-afad-9c386fbb7365`.
- **Candidate #2 — NOT built as originally worded. Refused on the evidence, not merely
  descoped.** All 13 (16 by keyword match, 2 false positives) deferred verdicts this section
  names carry `spec.filing_mode='record'` — `write_audit_findings_action.go`'s RFC_056
  (2026-08-25), a deliberate owner-ruled circuit breaker stopping LLM-audit-seat findings from
  auto-dispatching as page rewrites, built specifically because an earlier auto-dispatch of
  exactly this finding class destroyed live content (`bugs_closed/238`, finetuning.uk). "Wire
  the existing detector to a dispatcher" would have rebuilt that same promoter for the same
  finding class RFC_056 exists to hold back. Instead: **a human-reviewed release surface** —
  `HandleReleaseRecordVerdict` (new admin endpoint), a `filing_mode` filter on the existing
  work-item list endpoint, and a "Review & Release" button in the admin dashboard (previously
  the `deferred` status wasn't even a dropdown option there). Releases exactly one row a person
  has reviewed via the same operation the row's own `spec.release_recipe` describes — never
  executes that stored string, and can structurally never touch a non-record-mode row. Also
  closes a real gap in the existing interim SQL-only release path (the loanzy RUNBOOK's raw
  `UPDATE`): a released row now stamps `filing_mode='released'`, so it can't be released twice
  and a future filing_mode-scoped census can't mistake it for still-parked. Registered as an
  addendum to IMP-056 (`docs026_concept_register/register/improvement-loop.md`) — that entry's
  own 2026-08-26 note had flagged this exact surface as "owed". Council-Submitted:
  `38be9226-d5b5-48b7-9b87-20efbaf3dec3`.
- **No code change actually dispatches any of the 13 rows.** That remains a human decision,
  one row at a time, through the new button — which is the point.
- Full writeup, decisions and reasoning: `docs024_key_docs_latest/bugfix_428_planner_deferral/`.

## 10. Status update, 2026-09-02 (later same day) — verdicts read, deploy gap found

**Council verdicts, checked at the artefact, not assumed:**
- Migration `687` (the prompt fix): **APPROVED**.
- The release-surface backend (`38be9226…`): **APPROVED**.

**Fresh chassis build confirmed live, both `agent-chassis` and `core-manager`**, commit
`ebf27c60377f984fd2847a1d5d88ff87ae01ebf7` — verified via `service_binary_capabilities`
(agent-chassis, every pod uniform) and `core-manager`'s own startup log directly (not
scrolled, far lower volume). `git merge-base --is-ancestor` confirms every commit this
bug made is an ancestor. **So `HandleReleaseRecordVerdict` and the `filing_mode` list
filter are live in `core-manager` right now** — an operator with API access (or `curl`)
can already release a verdict.

**Real gap found, not yet closed: the admin-dashboard FRONTEND has not been rebuilt or
redeployed.** `kubectl get deployment admin-dashboard` shows its pods are **170 days
old** — the "Deferred" status option, the "Record verdicts only" filter, and the "Review
& Release" button this bug committed to `frontends/admin-dashboard/src/App.tsx` are on
the branch but not in the running UI. Nobody can actually SEE or click the release
surface yet, even though the API underneath it is live. This was not caught earlier
because this session has no `node`/`npm` to build the frontend locally and defaulted to
"committed, backend confirmed live" without separately checking the frontend's own
deploy state — worth logging as a genuine miss, not a false claim (nothing was ever
asserted as deployed that wasn't checked), but a gap in what got checked.

## 11. What is left before this lane can close

- [x] **Deploy the admin-dashboard frontend — DONE, confirmed 2026-09-02 (`gap planner`
  session, owner said go-ahead).** `kubectl get deployment admin-dashboard`: both pods
  22 minutes old at check time, revision matching, image `v1.0.1355` (kustomize overlay
  and running pod agree, no drift). Checked at the artefact, not the rollout status: `kubectl
  exec` into both pods and grepped the served JS bundle directly —
  `grep -l 'Review Queue' /usr/share/nginx/html/assets/*.js` found it in both, same
  content-hashed filename (`index-D46-1-nI.js`) as this session's own pre-commit
  `vite build` verification, i.e. this is genuinely the code committed in `7c359649f`/
  `1babd6f63`, not a stale cached image under a reused tag. The "Deferred" filter,
  "Record verdicts only" checkbox, "Review & Release" button, and the "Review Queue" nav
  tab are all live now. This was NOT done by this session's own `make admin-dashboard`
  run — the deploy had already happened (22 min prior) by the time this was checked,
  most likely riding the same fresh-build event as agent-chassis/core-manager (§10)
  rather than a separate frontend-only release. Not independently confirmed which
  session/process triggered it.
- [ ] **A human actually uses the release surface** on at least one of the 13 verdicts
  (worked case: boxingonline's own `e3c2b440-c006-40ec-be7a-88d0b689ed1e`) — the frontend
  is live now, so this is unblocked. This is the operational decision named in
  `bugs_open/427` §10.2, not a code task.
- [ ] **Watch for the next council verdict** on anything resubmitted from this bug — none
  outstanding as of this update (both verdicts are in).
- [ ] **Not this lane's job, tracked elsewhere:** the follow-up sample bug 428 §4 itself
  named — checking, once real `plan_site` calls accumulate against the tightened prompt,
  whether `strategy_notes` actually names omitted roles individually now (no automated
  check for this; a manual `llm_call_log` sample in a future session). `bugs_open/206`'s
  data-coverage gap (3 verticals only). Candidate #2 as originally worded (an automated
  dispatcher) stays refused pending an actual owner ruling on RFC_056, not a future
  session's unilateral call.

## 12. Status update, 2026-09-03 — re-confirmed live, one more deploy cycle happened

Independently (before reading `gap planner`'s §11 entry above), this session ALSO built
`admin-dashboard` — `make build-dashboard IMAGE_TAG=v1.0.1355` — deliberately not editing
the shared makefile default (another session had an unrelated, unclear-provenance
`v1.0.1188→v1.0.1354` bump staged uncommitted on the kustomize overlay at the time; overriding
IMAGE_TAG at invocation avoided colliding with it). Verified the image directly before
attempting to ship it: `docker run --entrypoint sh … grep -c "Record verdicts only"
/usr/share/nginx/html/assets/*.js` → 1. `docker push` was refused by this session's own
auto-mode permission classifier (correct behaviour for a production push) — surfaced to the
user rather than worked around.

**By the time the user confirmed, the deployed image was `v1.0.1356`, not `v1.0.1355`** — one
tag higher than either this session's build or the `v1.0.1355` §11 already recorded as live.
`kubectl get pods -l app=admin-dashboard` showed both pods started 2026-09-03T08:57–08:58,
i.e. a THIRD build/deploy cycle landed after `gap planner`'s check, most likely part of the
same broader fresh-fleet build the user mentioned ("a fresh chassis build has been
deployed") rather than a dedicated admin-dashboard-only release. Re-verified at the artefact,
not assumed from the tag bump: `kubectl exec` into a live pod, same
`grep -c 'Record verdicts only' /usr/share/nginx/html/assets/*.js` → 1. **So the release
surface is confirmed live as of 2026-09-03, across three independent checks by two sessions
at two different tags** — this is now about as solid as "is it deployed" gets on this estate.
Nobody has used it on a real verdict yet (§11's other open item stands).

## CONTRIB 2026-09-03 ~11:00Z (gamedesign.uk lane) — the FIRST POST-687 instance: the obligation was MET, and the reason it produced is FALSE

Routed here by the `site-design-planner` session, which correctly pointed out I had first filed
this against the wrong agent (that lane is composition resolution — `resolve_composition_layout_action.go`
and siblings; this is `build-site-planner`/`plan_site`, no code overlap). The finding below was
originally appended to `bugs_open/444` (commit `7343ecb01`) as an articles-hub cause; it is
better read as **an instance of THIS bug in a page type §3 did not sample**, and as the first
audit of 687's output.

**Why this one matters more than "another site does it too": it is POST-fix, and 687 worked.**
§3's sample runs 2026-05-14 → 2026-08-31, i.e. entirely before migration 687. This call is
2026-09-03 10:40:15Z, and I confirmed 687's rule reached it by grepping the RENDERED prompt
(`llm_call_log.prompt_rendered`, call `7b3bffdd-64dc-4a97-bb00-7633aa7271f8`), not the agent row:

> "omitted named type with no per-type reason in `strategy_notes` is a gap, not a decision."

**The planner COMPLIED with that obligation.** It named the omitted type and gave a per-type
reason — which is precisely what 687 asked for, and it is now auditable because of 687. Verbatim
from its `strategy_notes` (not truncated: 4,072 output tokens against `max_tokens` 16,000):

> "The strategy recommends index, blog-index, blog-post, and content pages. **All four types are
> present.** … The blog-post type is **satisfied by the blog infrastructure**; individual posts
> are not planned as static pages here."

**Two defects live in that quote, and 687 catches neither:**

1. **The stated reason is factually false.** There is no blog infrastructure and no later
   editorial pass. `[MEASURED 2026-09-03 ~10:48Z]` active `page_type='blog-post'` pages with a
   non-empty `sections` array: webdesign.co.uk **52**, dartsonline.com **23**, finetuning.uk
   **22**, ai-agent-orchestration.com **18**, seotools.co.uk **14**, **gamesdesign.co.uk 13**
   (the omitting site's own sibling) — every one an ordinary planned page built by the normal
   page pipeline, and the planner plans them directly elsewhere (farmerinsurance.uk 13,
   loancalculator.co.uk 14, dartsonline.com 9 `blog-post`-role rows in the CURRENT plan).
   `needs_content_page` only BUILDS pages already planned. So the deferral names a producer that
   does not exist.
2. **It asserts presence and explains absence in the same paragraph.** "All four types are
   present" is false on its own output — the plan has zero `blog-post` pages. A checker keyed on
   the *presence of a reason* reads this call as compliant.

**So the residual after 687 is: the "note why" obligation is satisfiable with a hallucinated
justification, and nothing checks the reason against the estate.** 687 deliberately stopped short
of a `validate_site_plan` hard failure to preserve the model's licensed final say — that judgement
still looks right to me, and I am not arguing for reversing it. What this case adds is that the
audit surface 687 created has **no truth check on the artefact it points at**, and the falsehood
is mechanically checkable: "blog_posts resolves to zero" is already computed a few lines away by
444's gate, which filed `capability_gap` `builder_needed=blog_posts` on this very plan, at
10:40:18Z — three seconds after the planner said the type was satisfied. **Two mechanisms in the
same validation pass reached opposite conclusions about the same page type and neither saw the
other.**

**Not site-specific, and the fleet count is small but the sites are pointed.** `[MEASURED
2026-09-03 ~10:47Z]` **3 of 32** `plan_site` runs in the trailing 30 days carry this
blog-post-deferral reasoning: designblog.co.uk (2026-09-02 16:10:51Z, "no individual posts are
planned in this architecture pass — posts are created editorially"), seotools.co.uk (2026-09-02
16:13:24Z, "not planned as static pages … planning placeholder blog-posts with no verified
content would be dishonest"), gamedesign.uk (above). Note seotools.co.uk appears in BOTH lists —
it refused on 09-02 and still serves 14 posts — so the refusal only costs a site when there is
nothing already there, i.e. the remake/rebuild case, which is where §3's cluster lives too.

**The finding that constrains any fix: there is NO per-site lever.** gamedesign.uk's mission was
seeded 09:45:50Z, 55 minutes before this run, saying in plain words "The site launches with real
articles, not a description of what the articles will be like. A page that lists articles must
list articles." I verified those exact words reached the model by reading the rendered prompt
(line 110). **It planned zero anyway.** `site_plan_directives` is not an alternative either:
`[MEASURED 2026-09-03 ~10:49Z]` all 1,922 rows are written BY `build-site-planner`/`write_site_plan`,
and the string "directive" appears **0 times** in the rendered prompt — output, never input. So
this cannot be answered by briefing; it lands where 687 landed.

**One interaction worth checking before anyone words a follow-up**, offered as a hypothesis and
marked as such: `[INFERRED, not tested]` migration 720's rule 3 still opens "Pages with page_type
entity-page, tool, blog-index, **blog-post** may have empty sections arrays", while the
already-built-site preserve block separately tells the planner that a page shown with
`"sections": []` is "rendered by another part of the system". Those two sentences together are a
plausible route to exactly the inference above. I have not tested it and it may be coincidence —
but if a prompt edit is made, it is the pair I would read first.

**No `090` run behind this (CLAUDE.md 2026-07-31 — stating the substitution, as it permits).** The
claim rests on primary output: the planner's own recorded reasoning, the rendered prompt proving
both 687's rule and the contrary mission instruction reached it, and a fleet census of the
artefact it defers to. The one inferential step ("the producer does not exist") is that census,
framed so a non-plan source of `blog-post` pages would have falsified it; I also checked the
30-day `item_type` vocabulary for a producer and found only `needs_content_page`, which builds
pages already planned. Cheapest refutation if this lane reads it differently: name a mechanism
that creates `blog-post` page rows without the planner.

Cross-refs: `bugs_open/444` CONTRIB same day (the articles-hub symptom and 444's gate behaviour);
lane docs `docs/agent_docs/docs024_key_docs_latest/gamedesign_uk_rebuild/` (NOTES 2026-09-03
~10:40–10:55Z entry).
