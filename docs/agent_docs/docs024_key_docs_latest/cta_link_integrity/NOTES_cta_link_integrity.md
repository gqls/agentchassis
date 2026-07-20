# NOTES — CTA / link integrity

Append-only, newest at the bottom. One entry per session.

---

## 2026-07-19 — session 1 (leopardess3): the four broken buttons, traced to root

**Trigger.** Owner reported four buttons on leopardessconsulting.co.uk that "I don't
understand what these buttons are and what they do, I think they are broken":
*Start Ranking Free*, *See How It Works*, *Start the Guide*, *Visit the Tool*.

All four live on the two tool pages, `/tools/llm-cost-calculator.html` and
`/tools/ai-agent-roi-estimator.html`, in two components. **All four are genuinely
broken, in four different ways, and each way defeats a different check.**

### What actually renders (curl'd from the live site, 2026-07-19)

```
/tools/llm-cost-calculator.html
  <a href="/contact.html"                                   class="brht-btn-primary">Start Ranking Free
  <a href=""                                                class="brht-btn-secondary">See How It Works
  <a href="#guide-start"                                    class="tgi-btn-primary">Start the Guide
  <a href="https://leopardessconsulting.com/tools/llm-cost-calculator"
                                                            class="tgi-btn-secondary">Visit the Tool
/tools/ai-agent-roi-estimator.html
  <a href="/contact.html"                                   class="brht-btn-primary">Start Ranking Free
  <a href=""                                                class="brht-btn-secondary">See How It Works
  <a href="#guide-start"                                    class="tgi-btn-primary">Start the Guide
  <a href="https://leopardess.contactforsales.com"          class="tgi-btn-secondary">Visit the Tool
```

Verified independently:
- `grep -c 'id="guide-start"'` on both assembled pages → **0**. The fragment target
  does not exist. `#guide-start` scrolls nowhere.
- `leopardess.contactforsales.com` → **NXDOMAIN**. **Owner-confirmed 2026-07-19: he owns
  `contactforsales.com` and `leopardessconsulting.com`, but never created the
  `leopardess.` subdomain.** So this is not a third-party leak.

> **CORRECTED 2026-07-19 (same session, before hand-off).** I first wrote that the model
> "assembled a hostname from two real domains in the owner's own estate", which implies it
> had knowledge of what the owner owns. **That was wrong, and the owner challenged it** —
> he suggested his ownership of `leopardessconsulting.com` was probably coincidence. He was
> right, and checking it produced a sharper mechanism for both hostnames:
>
> - **`leopardessconsulting.com`** — simply the obvious `.com` variant of the site's own
>   name. The model needs no knowledge of the owner's estate to guess it. That the owner
>   happens to own it **is** coincidence (a common one — people who buy a `.co.uk` often
>   buy the `.com`).
> - **`leopardess.contactforsales.com`** — *not* coincidence, and *not* estate knowledge.
>   The site's own identity spec carries the real contact email
>   **`leopardess@contactforsales.com`** (`docs/leopardessconsulting/specs/identity.json`,
>   and the live `site_specs` identity row). The model saw a real email in its context and
>   **transformed it into a hostname by swapping `@` for `.`**.
>
> So the parts were true and in-context; only the *recombination* was invented. That is the
> classic fabrication shape, and it is far more checkable than "plausible string" — see the
> deterministic rule below. Caught by grepping the repo and DB for `contactforsales` instead
> of accepting my own first framing.

**Fleet exposure of the email→hostname transform.** `contactforsales.com` is the owner's
shared contact domain across **six sites**, each with a `<label>@contactforsales.com` address
in its *current* identity spec:

```
ai-agent-orchestration.com | agents@contactforsales.com
finetuning.uk              | finetuning@contactforsales.com
gaswholesalers.com         | gas@contactforsales.com
idea.uk                    | idea.uk@contactforsales.com
leopardessconsulting.co.uk | leopardess@contactforsales.com
dartsonline.com            | sales@darts.com
```

Any of them can produce the same fabrication. **This yields a cheap deterministic check:
a hostname equal to a known contact email with `@` → `.` is fabricated by construction.**
No HTTP call needed, no heuristic — it is a string identity against data the platform
already holds. Added to the plan as P1.5.
- `leopardessconsulting.com` (owner's, but not the live site) returns **200 with a
  114-byte body and no `<title>`/`<h1>` for every path tested**, including the root and
  the fabricated tool path. A visitor clicking *Visit the Tool* leaves the real site for
  a blank page. Not investigated further — owner interrupted the whois lookup; domain
  disposition is an owner question, not a platform one.
- The ROI page **does** contain a real ROI calculator (`id="roi-payback"`,
  `roi-roi-percent`, `roi-monthly-savings`…). My initial suspicion that the page had no
  tool at all was **wrong** — corrected before it reached the plan. The Bayesian widget
  sits *on top of* a working tool.

### Root cause, component by component

**`bayesian-ranking-hero-tool`** — *Start Ranking Free*, *See How It Works*.

The slot name resolves by `function`, not `name`. The row actually serving it is:

```
name                                | function                   | is_active
bayesian-ranking-hero-tool_pre_037  | bayesian-ranking-hero-tool | t
```

**A `_pre_037` backup snapshot, never cleaned up, still `is_active`, and it is the live
component on three tool pages across two sites** (leopardess ×2, finetuning.uk ×1).

Its CTA schema:

```
cta_primary_label   | static                       | fallback: Start Ranking Free
cta_primary_url     | site_specs.cta.primary_url   | (no fallback)
cta_secondary_label | static                       | fallback: See How It Works
cta_secondary_url   | site_specs.cta.secondary_url | (no fallback)
rank_btn_label      | static                       | fallback: Calculate Rankings
result_label        | static                       | fallback: Bayesian Rankings
tool_panel_title    | static                       | fallback: Try the Bayesian Ranker
badge_label         | static                       | fallback: Free Ranking Tool
```

Template, **ungated**:
```
<a href="{{.cta_primary_url}}"   class="brht-btn-primary">{{.cta_primary_label}}
<a href="{{.cta_secondary_url}}" class="brht-btn-secondary">{{.cta_secondary_label}}
```

So:
- Every *label* is `source:static`. Per `plan_sections_action.go:1210-1218`, static and
  renderer sources take a branch that writes the fallback and `continue`s — **bypassing
  `required` and `on_missing` entirely**, and re-applying on every render. `content_data`
  cannot override them. The Bayesian ranker's vocabulary is therefore *frozen onto every
  page that uses this component*, whatever the page is actually about.
- `cta_secondary_url` resolves from `site_specs.cta.secondary_url`, which is unset on
  this site → empty string → the ungated template emits `href=""`. That is *See How It
  Works*.
- `cta_primary_url` resolves to `/contact.html` — the exact residue of the fleet-wide
  "143 of 144 CTA buttons point at /contact.html" bug that migration 091 fixed. **091/098
  only freed the six components in `ctaFieldNames`; this component is not one of them**,
  so it was never unlocked.

**`tool-guide-intro`** — *Start the Guide*, *Visit the Tool*. Used on **4 pages / 3 sites**
(leopardess ×2, finetuning.uk, robot-hands.com).

```
cta_primary_label   | static | fallback: Start the Guide
cta_secondary_label | static | fallback: Visit the Tool
cta_secondary_url   | llm    | required: TRUE   ← no fallback, no source of truth
(there is no cta_primary_url field in the schema at all)
```

Template:
```
<a href="#guide-start"           class="tgi-btn-primary">{{.cta_primary_label}}
<a href="{{.cta_secondary_url}}" class="tgi-btn-secondary">{{.cta_secondary_label}}
```

So:
- *Start the Guide* has **no URL field anywhere in the schema**. Its destination is a
  hardcoded `#guide-start` in the template, and no component on any of those 4 pages
  defines that id. Dead by construction, fleet-wide, on every site that uses it.
- *Visit the Tool* asks an **LLM to author a URL, and marks it `required:true`**. There
  is no page-set resolution, no validation, and a required field the model cannot look
  up. It invents. It invented two different hostnames on two adjacent pages of the same
  site. This is the cleanest example in the repo of *a schema that manufactures
  fabrication*.

### Why nothing caught it — the detection map

Cited from a full read of the check machinery (see RUNBOOK for the file:line map):

| Button | Failure | Why every check missed it |
|---|---|---|
| Start Ranking Free | wrong tool's label; lands in excluded area | **DETECTED** — see below |
| See How It Works | `href=""` | `validate_page_content.go:551-560` files `empty_internal_href` as a **warning**; `valid := blockerCount==0 && errorCount==0` (`:257`) — warnings never block. `check_dead_controls.go` explicitly **cedes** `href=""` to the phantom check (`links.go:107`). |
| Start the Guide | `#guide-start`, no such id | `ClassifyLinkScope` returns `LinkScopeAnchor`; `validate_page_content.go:577-578` skips it, `check_phantom_internal_links.go:279` skips it, `check_misdirected_cta.go:198-201` skips it. **No check anywhere resolves a fragment against the page's ids.** |
| Visit the Tool | fabricated / dead external host | `LinkScopeExternal` is skipped by every consumer. **No discovery check performs HTTP.** `checkDomainContamination` (`validate_page_content.go:481-534`) matches a hardcoded 5-domain list by substring — it cannot compare a wrong-TLD variant against the site's own expected domain. |

**The one that WAS detected is the most damning.** On **2026-07-17**, two work items were
filed:

```
cta_names_unknown_destination | CTA "Start Ranking Free" on ai-agent-roi-estimator
                                (bayesian-ranking-hero-tool): lands in an excluded area (contact)
cta_names_unknown_destination | CTA "Start Ranking Free" on llm-cost-calculator …
```

Correct diagnosis, correct component named, two days before the owner clicked it. It was
filed `status='needs_human_review'` — which `TriageDetectedItemsAction` never promotes,
which no `handler_agent` consumes, and which
`load_work_item_actions.go:804` explicitly **excludes** from re-open queries. A grep of
the whole `platform/` tree for `unresolved_cta`, `cta_names_unknown_destination` and
`dead_control` returns **only their emission sites — zero consumers.**

Current leopardess queue: **21 `unresolved_cta` + 13 `cta_names_unknown_destination`, all
`needs_human_review`**, oldest 2026-07-13.

> The system found the bug, wrote it down correctly, and filed it somewhere nobody reads.
> That is not a detection gap. It is a **delivery** gap, and it is a different fix.

### Fleet-wide sizing (all active pages, all sites)

```
1. empty href=""      |  30 anchors |  7 sites
2. bare #             |  17 anchors |  6 sites
3. fragment #anchor   |   4 anchors |  3 sites
5. external           |  17 anchors |  5 sites
7. internal path      | 543 anchors | 11 sites
```

**51 dead or suspect controls across 7 of 11 sites.** This is not a leopardess problem.

### Corrections to my own earlier claims in this session

- > **CORRECTED:** I first read `component_id IS NULL` across many rows as "orphaned
  components" and nearly filed it. It is not — `component_id` is simply unpopulated for
  most component types fleet-wide, and resolution happens by `slot_name → function`.
  The real orphan finding is narrower and survives: `bayesian-ranking-hero-tool` has no
  row by `name`, only a `_pre_037` backup row by `function`. Caught by re-running the
  query against `content_components.function` instead of `.name`.
- > **CORRECTED:** I suspected `/tools/ai-agent-roi-estimator.html` had no ROI calculator
  at all. It does. Caught by grepping the rendered page for `roi-*` ids before writing it
  into the plan.
- The HANDOFF's claim that 4 of 5 tools "work" is defensible for the *calculators*, but
  every one of those pages carries a broken hero CTA pair. "The tool works" and "the page
  works" are different claims.

### 2026-07-19 — session 2 (experience loop 2): class G is wider than `needs_human_review`

*Added by a different thread. I was verifying a precondition for the Experience Loop's
build phase, not working this bug — but what I found is class G, so it belongs here
rather than in a new file.*

Class G is currently stated as findings dying at `needs_human_review`, which nothing
consumes. **The `detected` status leaks the same way, and that case is worse, because
those items DO have a handler.**

Evidence, vonc.com (`9ec3b9ee-…`), live 2026-07-19:

| Item | State |
|---|---|
| 3 × `deactivated_component` (slots `header`/`footer`/`head`) | `detected` since **2026-07-11**, `updated_at` never moved — 8 days |
| all vonc work items at `detected` | **11**, oldest 07-11, newest 07-15 |
| fleet-wide `deactivated_component` | 14 `unresolved`, 6 `detected`, 21 `complete` |

The check (`discovery_checks/check_integrity.go:158` `DeactivatedSiteComponentsCheck`) is
**enabled** (on `design-discovery-agent`) and fires correctly, and it names a real handler
(`HandlerAgent: "rerender-pages"`). It still never runs, because emission is at
`Status: "detected"` and `detected` is not dispatchable — it must first be promoted to
`triaged` by `TriageDetectedItemsAction` (`triage_detect_items_action.go:91-104`, the same
file the plan already cites for class G). That promotion is per-site and promotes
everything at once; for vonc it has evidently not run since 07-11.

So class G has two distinct leaks, and the plan's P3.3 should cover both:
1. **no handler** — `needs_human_review`, the case already documented;
2. **handler exists, never reached** — `detected` with no triage run. Adding a handler
   fixes nothing here; the item never gets far enough to reach it.

**Not currently a live breakage, and I want to be precise about that.** All three vonc
slots have `build_status='rendered'` with real baked HTML (3,638 / 3,903 / 8,605 B), so the
chrome serves fine today. The exposure is latent: the library components behind them are
`is_active=false`, so a re-render of those slots may skip them. This is fleet-wide — `head`
is inactive on **11 of 11** sites, `site-footer` on 9, `site-header` on 8.

> Correcting a claim from the experience-loop thread's own handoff, which recorded the
> council critic's report as "site components **deactivated across 16 pages**". That
> conflated two tables. Every one of vonc's 49 `page_components` attachments is active;
> the deactivation is in `site_components` (3 slot rows), which is not per-page at all.
> The critic named the right three components — `header-bold-gradient`, `footer-4-column`,
> `Document Head` — and those names match the three stuck work items exactly.

### Prior art found (so nobody re-walks it)

The link-integrity loop (`docs/social001_vonc_tiktok_social/minilobby_task/`) closed the
big one on 2026-07-16: migration 091 removed the `pages.contact` lock, 092 enabled
`phantom_internal_links` / `misdirected_cta` / `incomplete_page_group`, 098 broadened
`ctaFieldNames` to six components. Its own documented residue is exactly what bit us:

- **Repair scope < detection scope.** `resolve_internal_links_action.go:83-86` says a
  component outside `ctaFieldNames` is *"detectable but not repairable — its findings can
  only escalate to human review."* Both of our components are outside it.
- **`*_label` fields left static** — deferred to a content pass, never done.
- LNK-005, the stated invariant, is *"an unresolvable destination renders nothing (no
  button) rather than a broken/empty link"* — **the ungated templates here violate it
  outright**, and no test enforces it.
- idea.uk hit the identical symptom (`README_where_we_are.md:646-651`, "Buttons do
  nothing"), root-caused to *no discovery check had ever run against the site*.
- The `{{if .link_url}}` gating fix was applied to `info-card-grid` **on this very site**
  (leopardess HANDOFF §3) and never generalised to any other component.

Full citation set in the plan's appendix.

---

## 2026-07-19 — session 1 (cont): diagnosis loop CONFIRMS the mechanism and CORRECTS my attribution

Fired the diagnosis loop on the source-authority question (owner's call — the "same migration
written four times" worried him enough to want a cited direction rather than my preference).
Correlation `c84940fb-6a43-47eb-86ea-eb1690b9f305`, item_key `needs_diagnosis:cta-source-authority`,
3 iterations, bundles 67k → 97k → 131k, verdict **CONFIRMED**.

### ⚠️ MISSTEP OF MINE, caught by the loop — `static` does NOT overwrite the resolver

I wrote, in the plan and in chat, that *"only a field whose source is `renderer` lets the
resolver win"*, listing `static` among the sources that beat it. **That is wrong**, and the
loop grounded the correction in the code I had already read:

> `planSection`'s per-field loop groups `static` **with** `renderer` under one shared guard
> clause (`plan_sections_action.go:1211-1218`). Both skip resolver re-resolution entirely and
> only apply a fallback **if one is declared**. `sourceResolver.resolve` returns `(nil, true)`
> for both without ever querying.

The genuinely vulnerable sources are the **query-resolved** ones — `site_specs.*`, `pages.*`,
`site_assets.*`, `config.*` — which re-run their DB lookups (`ensurePages`/`ensureSpecs`)
fresh on **every** `plan_sections` call, with nothing cached from the prior build.

**Where my error came from, because the distinction is worth keeping.** I generalised from
*label* behaviour to *url* behaviour. For a **label**, `static` + a declared fallback genuinely
does re-apply every render — that is exactly why "Start Ranking Free" is frozen, and that part
of the original diagnosis stands. For a **url with no fallback declared**, the same `static`
source writes nothing and the resolver's value survives untouched. **Same source keyword,
opposite behaviour, decided entirely by whether a `fallback` key is present.** I had the
mechanism right for one field type and carried it to the other without checking.

The loop also flagged, correctly, that it could not establish which sources the *real* CTA
fields carry from its bundle. I could — measured live, against the derived CTA set:

| class | fields | components |
|---|---|---|
| **VULNERABLE** — `site_specs.*` / `pages.*` / `site_assets.*` / `config.*` | **83** | **18** |
| immune — `llm` (writer-authored, not resolver-overwritten) | 21 | 7 |
| immune — `renderer` / `static` guard | 13 | 6 |
| `query.section_index_for:{game,tool}` — separate `queryresolve` path, almost certainly also re-resolved | 2 | 2 |

> **CORRECTED figure.** The plan said *"only 6 of the 119 are `source:renderer`; the rest get
> re-resolved and beat the resolver."* Wrong on both halves — **13** are immune by guard, and
> the vulnerable set is **83, not ~113**. The target is meaningfully smaller and sharper than
> I claimed: 18 components, not 33.

The loop independently found a live example confirming `pages.*` as a real overriding source:
`header-bold-gradient` / `site-header` carries
`"cta_url": {"source": "pages.contact", "fallback": "/contact.html", ...}` — which is the
literal fossil of the "143 of 144 buttons point at /contact.html" bug, still in the schema.

### What the loop did NOT do

`is_fix: false`, `stopped_by: confirmed`. It diagnosed the cause; it did not choose between
the two remedies (migrate sources vs. teach `plan_sections` to yield to the resolver). That is
the expected division of labour — the diagnosis loop tells you the cause, the **council gate**
reviews a proposed fix. The direction question is council-shaped and still open.

### Trap for the next thread: the trigger's default `REF` is dangerous here

`090_TRIGGER_needs_diagnosis_v1.sh` defaults to `REF=main`. **`origin/main` carries a
2-entry `ctaFieldNames` map; the working branch carries 6, and is 345 commits ahead.**
Diagnosing on the default would have shown the loop a state in which the bug is barely
present and the "amended repeatedly" evidence largely absent. Pinned to
`085_debug_and_feature_loops` after verifying the remote branch carries both the current map
and the commit that produced it (`e10c656f3`). **Check what your REF actually contains before
firing — a confident diagnosis of the wrong tree is worse than no diagnosis.**

Also: omitting `SUBJECT_TYPE`/`SUBJECT_KEY` means `persist_note` runs but writes no
`doc_notes` row (it is a subject gate). I omitted them deliberately rather than guess a
routing vocabulary — 016b §9 has a pattern about mistyped routing keys producing silence in
every gate at once. Consequence: the verdict lives only in
`orchestration_states.collected_data->'diagnosis'`, not in `doc_notes`. Retrieve it with the
query in RUNBOOK R11.

---

## 2026-07-19 — council gate: REJECTED (guardian veto), and it caught a real error in my staging

Submission trail `2525f980-3fde-4b62-aff3-225de8454000`. Round 1 voided by `bugs_open/019`
(reviewer truncation — third instance, evidence appended to that case file). Round 2 ran the
full panel: **8 seats reported, 5 abstained, decision `rejected`, decided_by "hard veto from
guardian".**

### The direction — and it is neither of the two options I posed

Near-unanimous across seats, including both that objected and most that approved:

> **Ship Option A (the fifth migration) NOW as the safe stopgap, AND land Option B's
> consolidation with the `plan_sections` edit staged observe-only first — then retire A's
> pattern.** Sequence, not choice.

`guidelines` put it most compactly: *"ship A now as the safety valve, land B behind the
observe-only stage, then retire A's pattern."* The guardian's veto is narrower than
"Option B is wrong": *"ship Option A now; keep edits 1, 2 (observe-only) and 4 as prep, drop
edit 3 pending a real architecture review of resolution ownership."*

### ⚠️ MY ERROR: I staged the wrong edit. Exactly backwards.

Flagged independently by `bug_historian` (medium) and `debug_historian` (**high**):

> *"THE CONTESTED EDIT ships live and unstaged in the single hottest render path on the
> platform, while the sibling edit (resolve_internal_links) gets an observe-only logging round
> before any write changes. That asymmetry is exactly backwards."*

I applied careful staging to the **low**-risk call site and none to the **high**-risk one —
having myself labelled the latter "THE CONTESTED EDIT". `debug_historian` names the
aggravating factor: *"the plan itself demonstrates it knows how to de-risk this via staging,
and simply didn't apply it to the riskiest edit."* That is not a judgement call I got wrong;
it is an inconsistency I should have caught by reading my own plan.

### Three further defects in the plan, all real

1. **`required:true` + nil fallback → the field is left unset** (`guidelines`, medium). My
   sketch does a bare `continue` when `ctaOwnedURLFields[fieldName]` is true and no fallback
   is declared. *"if that field's schema marks required:true, this is exactly the
   skip_field/empty default the rule forbids, now applied to up to 83 fields."* A genuine bug
   in the proposed code, not a stylistic note.
2. **The sibling-label rule relocates the map's failure mode rather than removing it**
   (`bug_historian`). A component whose CTA url does not follow the convention silently drops
   out of both detection and repair — *"the map's exact failure mode, relocated into schema
   shape."* Needs a **loud-fail guard**: log any `*_url` field with no sibling label AND a
   query-resolved source, so an uncovered CTA is visible rather than silently reverting to
   overwrite-prone behaviour. My risk #2 admitted the problem and proposed no remedy.
3. **No pod-binary deploy verification** (`debug_historian`, medium) — Go changes are inert
   until an image roll, so the new binary must be **grepped from the running pod**, never
   inferred from a git commit or image tag. That is CLAUDE.md's own standing rule and my plan
   omitted it.

Minor: `editquality` caught an internal inconsistency — I wrote "only six could ever be
repaired" while my own citation says the map covers **5 of 33** functions.

### How the four questions came back

| Q | Answer |
|---|---|
| 1. Is the `plan_sections` blast radius justified? | **Not as submitted** — because the staging is on the wrong edit. `editquality` alone said justified outright. |
| 2. Is the sibling-label rule a durable contract? | **Unanimous: no** — today's naming convention. Keep `ctaFieldNames`' override slot **permanently**, not transitionally. |
| 3. Does a fifth migration count as pragmatic? | **Split, and this is the real disagreement.** Five seats: the repetition *is* the defect; a fifth fixes 83 fields and leaves the mechanism live for the next component. Guardian: *"contained toil — each migration touches only registry rows, never a hot Go path; four migrations is cheaper than one live-hot-path incident."* |
| 4. Does the inert-until-image-roll asymmetry decide it? | Guardian: **yes, strong tiebreaker.** `editquality`: no, liveness shouldn't override correctness — *and* Option A "isn't as purely mechanical as advertised" because it still needs the map extended to cover 28 more components before a migration reaches them. |

### Worth keeping

- `render_guardian` flagged a **silent-blank risk** nobody else did: the skip-with-fallback
  branch could leave a CTA url unset *between build and repair*. Outside its mandate, routed
  rather than judged — the relevance-gated panel working as designed.
- **Seat edit-indexing is inconsistent** between reviewers (some 0-indexed, some 1-indexed),
  so "edit 3" means `resolve_internal_links` to `editquality` and `plan_sections` to
  `guardian`/`bug_historian`/`guidelines`. The substance is unambiguous, but do not machine-
  read the `edit` field across seats.
- `council_report.metadata->>'source_agent'` is **empty**, not the seat name — the known
  fleet-wide landmine. Partition by the `reviews[].reviewer` field inside the body instead.

**Status: the plan is NOT to be shipped as submitted.** REJECTED is a veto, not a score.

---

## 2026-07-20 — P4 leopardess page fix APPLIED AND VERIFIED LIVE

All four owner-reported buttons are gone from the live site. Verified against the rendered
artefact, not the orchestration status.

**What changed.** Both tool pages lost their two misplaced sections; the genuine tool and
`tool-cta` are untouched.

```
llm-cost-calculator     hero-tool · tool-guide-intro · tool-llm-cost-calculator · tool-cta
                     -> tool-llm-cost-calculator · tool-cta
ai-agent-roi-estimator  hero-tool · tool-guide-intro · tool-ai-agent-roi-estimator · tool-cta
                     -> tool-ai-agent-roi-estimator · tool-cta
```

Applied in one transaction (`scratchpad/leo_toolpages_fix.sql`): backups →
`page_components` delete (4 rows) → position renumber → `pages.sections` align → in-transaction
verify → commit. Backups: `bak_leo_toolpages_20260720_pc` (8 rows),
`bak_leo_toolpages_20260720_pages` (2 rows). Deployed via `reassemble_pages.sh` (ASSEMBLE
mode — a structural change must not go through `section_data_resolved`, which skips
content-unchanged pages).

**Live verification, both pages:** 0 occurrences of all four button strings, 0 `href=""`,
0 `#guide-start`, 0 fabricated hosts, 0 Bayesian residue; `calc-btn` and
`roi-payback`/`roi-roi-percent` still present. Page sizes 62,627→44,373 and 41,632→23,243
bytes — the removed weight is the two wrong components.

**Fleet census moved:** empty hrefs **30 → 25**, fragments **4 → 2**. Bare `#` unchanged at 17
(a different class, untouched by this fix).

### The planner-selection cause, and it is smaller than I assumed

`pages.sections` asked for **`hero-tool`** — a generic, entirely sensible section name. It
resolved to `bayesian-ranking-hero-tool`. So the planner did **not** propose a Bayesian ranker;
the *selection* did, because:

```sql
SELECT name, function FROM content_components
WHERE is_active AND (function LIKE '%hero%tool%' OR name LIKE '%hero%tool%');
--  bayesian-ranking-hero-tool_pre_037 | bayesian-ranking-hero-tool   (the ONLY row)
```

> **The component library contains exactly one active component that can serve a `hero-tool`
> section, and it is hard-coded to a different tool.** That is a *missing component* problem,
> not a planner defect. Any site whose plan asks for a tool hero gets the Bayesian ranker's
> frozen vocabulary — which is precisely what happened on finetuning.uk too.

This is materially cheaper to fix than "diagnose the planner": build one generic `hero-tool`
component (no product-specific static fallbacks, gated CTA anchors) and the selection resolves
correctly everywhere. Added to the plan; **not** done here.

### Checks that changed the fix, worth repeating

1. **Section authority is the ASPECT on this site**, flagged by the bugfix thread's handoff
   note: the current `site_plans` row has zero `site_plan_sections`, so an
   aspect/`pages.sections` mismatch silently reverts. Verified the aspect lists 13 pages and
   **neither tool page is among them** → `pages.sections` governs, no aspect edit needed. Had I
   assumed either way I would have been wrong half the time.
2. **Removal is not lossy.** Checked before deleting: the real guide pages carry hero + an
   8.9KB `article-body` + CTA. The tool-page `tool-guide-intro` duplicated content that already
   lives where it belongs — a "read time" and a "Start the Guide" button do not belong on the
   tool itself.
3. **bugs_open/001 is fixed and ACTIVE in v1.0.1138**, which is what made a `page_components`
   fix durable enough to be worth doing. Verified both halves: Go symbols
   (`realised_sections`, `snapped_sections`) in the running pod AND `build_status` present in
   the live `build-site-planner` query — the migration's own header states the Go silently
   degrades to the old behaviour without it. Both pages are `build_status='deployed'`, so they
   are in the preserved set.
   > **Ledger trap:** migration **173 is absent from `schema_migrations`** (172, 174, 175 are
   > all recorded) yet its effect IS live in `agent_definitions`. Applied-but-unrecorded —
   > `bugs_open/007`'s exact shape. **Had I trusted the ledger I would have concluded 001 was
   > still broken and steered away from `page_components` for no reason.** I did not add a
   > ledger row: it is another thread's migration and attributing an application I did not
   > perform is the half of 007 that causes damage. Flagged to the owner instead.

### Still broken fleet-wide — this fixed the site, not the components

`tool-guide-intro` remains live on **finetuning.uk** (`llm-cost-calculator`) and
**robot-hands.com** (`gripper-cycle-time-estimator`) with the same dead `#guide-start` and the
same `source:llm, required:true` URL field. `bayesian-ranking-hero-tool` remains on
finetuning.uk's `llm-cost-calculator`. Those are P2.2/P2.3 and unaffected by today's change.
