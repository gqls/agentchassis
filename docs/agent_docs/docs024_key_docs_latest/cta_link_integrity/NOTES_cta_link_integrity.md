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
