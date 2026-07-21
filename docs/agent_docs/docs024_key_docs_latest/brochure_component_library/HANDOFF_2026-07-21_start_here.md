# HANDOFF — brochure component library / fundamentallyai.com (2026-07-21)

**This is the cold-start entry point. Open a new thread from here.** Read this
first; the deeper detail is in the four standing docs in this directory
(PLAN / NOTES / README_where_we_are / the SUMMARY series).

---

## In one paragraph

Owner wants best-in-class consultancy brochure components (Bain/BCG/McKinsey
style — auto-advancing hero card carousels, hover-zoom cards, swipeable mobile
carousels, code-rendered stat bands) built **through the framework**, and a new
brand — **fundamentallyai.com** — that markets this platform's own **real,
verified** capabilities as service lines/case studies. Research is done, all
positioning decisions are made, and the site has been **onboarded and largely
built by the pipeline overnight** (v-build 2026-07-20). It is **not yet live**:
most content pages are blocked at a content-validation gate, and the
Cloudflare→origin serving path is not yet delivering anything. Two threads of
work remain: (A) get the built site unblocked and actually serving, and (B)
build the new interactive components (none exist in the framework yet).

## Resolved decisions (do not relitigate — all owner-confirmed 2026-07-20)

- **Domain**: fundamentallyai.com is owned; DNS has now **propagated to
  Cloudflare** (verified 2026-07-21: NS = `alexis`/`leah.ns.cloudflare.com`,
  same as the live fleet).
- **People imagery**: **line illustration only**, never photography, single
  consistent tint for cohesion. **This landed correctly** in the built site's
  `design_intent.imagery_direction` (verified — see NOTES).
- **Embeddings/private-search pitch**: frame as **buildable, not
  already-delivered** — real vector search exists, but tenant isolation does
  NOT exist on the shared store today (the platform's own audit rules it out).
  This is a near-hard constraint, not a wording preference.
- **Self-correction / claims-verification story**: **use it**, built strictly
  from `docs/leopardessconsulting/AUDIT_verified_facts.md`, not embellished.
- **Name leopardessconsulting.co.uk directly** as the worked example of that
  story (owner approved). The pipeline already created a
  `self-correction-leopardessconsulting` page for exactly this.
- **Hard exclusion list** (never resurface, even if an LLM regenerates copy):
  the specific past leopardess fabrications — invented founder, "70+ agents in
  8 departments", invented client case studies, "99.9% uptime", "2,767 Awards
  Won". Full list in NOTES under "The load-bearing exclusion list".

## What the pipeline actually built (all [VERIFIED] against live DB 2026-07-21)

Site row: `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`, domain fundamentallyai.com,
created 2026-07-20 20:36. The mission brief propagated end-to-end — the full
spec cascade completed (submission → mission_brief → identity → classification →
content_direction → design_intent → vertical_landscape → strategy → briefing →
resolved_composition), all imagery generated, and the pages the mission asked
for exist by name:

| page | build_status | note |
|---|---|---|
| index | needs_rebuild | homepage — blocked at validation |
| capabilities | needs_rebuild | blocked at validation |
| about | needs_rebuild | blocked at validation |
| multi-agent-review-council | needs_rebuild | flagship pillar; blocked at validation |
| model-fine-tuning | **deployed** (DB) | got through validation |
| contact | **deployed** (DB) | got through validation |
| self-correction-leopardessconsulting | planned | "no sections ready to build (empty spec sections)" |
| platform-log-index | planned | same empty-sections no-op |
| tool-decision-record | planned | tool page, not built |

Design intent (verified): dark navy/amber consultancy palette
(primary `#0E1B2E`, accent `#C8902A`), line-illustration imagery with a single
tint, code-generated charts from verified data. This is a strong,
on-brief starting point — the visual direction is right.

## The two blockers standing between "built" and "live"

**Blocker 1 — content validation gate (5 content pages).**
Every blocked page failed identically: `step validate_content failed: ...
content validation failed: **1 blockers, 0 errors**`. Consistent single blocker
per page; yet `contact` and `model-fine-tuning` passed, so it is content-
specific, not a universal outage. **The specific blocker reason is NOT
recoverable from the DB** (work-item `result` jsonb is empty) and the overnight
logs **rotated when the pod restarted onto v1.0.1144** — so the next thread must
**re-fire one blocked page on the fresh image and capture the blocker live**
(watch the chassis log during `validate_page_content`). Hypothesis worth
holding but NOT yet confirmed: this may be the same class as the leopardess
`contact-block` placeholder-email false-positive, or a claims/banned-language
blocker doing its job on the honesty-heavy copy — **diagnose, don't assume**.
This is the #1 next action.

**Blocker 2 — nothing is actually serving.**
DNS is on Cloudflare, but `https://fundamentallyai.com` returns a Cloudflare
**404** at root and even the DB-"deployed" pages don't serve (`/contact` →
connection timeout, `/model-fine-tuning` → 404). `github_repo`/bucket are empty
— but that is **normal** for the B2 fleet (robot-hands.com is live with them
empty too; the B2 fleet deploys per-domain directories via a **separate
portfolio repo** the git-adapter pushes to, synced by `deploy-to-b2.yml`, not
via `sites.github_repo`). So the serving gap is one of: (a) the git-adapter
hasn't pushed fundamentallyai.com's rendered pages to the portfolio deploy repo
yet, and/or (b) the new domain's **Cloudflare→B2 origin wiring** (a per-domain
infra step) isn't set up. **NOT yet diagnosed** — this is the #2 next action,
and part of it is likely an owner/infra step (as idea.uk's cutover was).

**Also outstanding, lower priority:** two `needs_section_data` human-review
items — `portfolio-showcase` on index needs real project data (titles/URLs/
descriptions — feed it our own 11 real sites, honestly labelled as ours, per
the exclusion rules), and `contact-info` needs a real business email.

## Recommended next actions, in order

1. **Diagnose Blocker 1 live**: re-fire ONE blocked page (e.g. `about`) on
   v1.0.1144 and capture the `validate_page_content` blocker from the chassis
   log. Pod is >300s old so the post-restart drop rule is clear; but the
   dispatch goes through the single-consumer queue (`bugs_open/030`) so expect
   a wait and **do not retry on an absent row** — check consumer lag.
2. **Diagnose Blocker 2**: determine whether rendered pages reached the
   portfolio deploy repo/B2 and whether the Cloudflare origin for the new zone
   points at the bucket. Likely needs an owner infra step — surface it plainly.
3. **Feed the two `needs_section_data` items** real, honest data.
4. **Then** the component build (Thread B) — start with `hero-card-carousel`
   (most-requested, exercises carousel-JS + hover-zoom-CSS + imagery-kind at
   once). Full acceptance checklist in PLAN ("Acceptance checklist for every
   new component type"). Council-review before commit per CLAUDE.md (touches
   `platform/`/`internal/`). NB: the pipeline chose existing components for
   this build — the fancy carousel/hover-zoom components **do not exist yet**,
   so today's site uses standard sections; the new components are a genuine
   from-scratch build that will then need re-planning selected pages to use
   them.

## Key identifiers

- Onboarding correlation `099ca178-92fc-41ac-bf6c-bc17c0aa6ec6`, orchestration
  `c6a53a35-...` (COMPLETED). Site id `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`.
- Chassis live on **v1.0.1144** (verified: image tag on running pod +
  deployment), pod `agent-chassis-59c675c4f-pxr9f`.
- Mission brief that produced this build:
  `MISSION_BRIEF_fundamentallyai_2026-07-20.md` (this dir).
- Trigger: `bash 082_submit_domain_unified.sh fundamentallyai.com --email
  fundamentallyai@contactforsales.com --mission-file <brief>`.

## Landmines carried from the research (full detail in NOTES/PLAN)

- Build components at `component_level='section'` (page path), not chrome —
  `bugs_open/041` silently drops chrome-only JS.
- A new component type is inert until it's named in the **build-site-planner /
  site-architect prompt** — registering the row is not enough.
- Any stat/number is **code-rendered from verified data**, never a generated
  picture of a chart (standing rule, and design_intent already encodes it).
- `bugs_open/001` — a full site re-plan can clobber built pages; prefer
  per-page/per-component re-fires over a whole-site re-plan on a site with
  hand-verified content.
- Verify by artifact, never by status: this build is the live example — the DB
  says two pages are "deployed" and none actually serve.
