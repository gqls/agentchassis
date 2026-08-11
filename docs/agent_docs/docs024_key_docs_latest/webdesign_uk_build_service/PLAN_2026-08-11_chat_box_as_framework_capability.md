# PLAN 2026-08-11 — turn the chat box into a framework capability, not a one-off

## Context

The chat-input-box on webdesign.uk's contact page is hand-built: one
`content_components` row, one `page_components` row spliced into one page,
one hand-written Go backend hand-deployed to one VM with the site's facts
hardcoded in Go source. It works, and it is now locked and lock-verified
against the automated rebuild pipeline (see NOTES 2026-08-11) — but "safely
rebuild it, improve it, or change it on any site with different parameters"
is a different bar. Today, doing that on a second site means repeating the
entire hand-build by hand. This plan decomposes it into pieces the framework
already knows how to run — component, tool, decision, deploy, maintain — and
names honestly the one piece the framework genuinely cannot do yet.

**This is scoped narrower than the earlier `ai_site_selling_automation`
research** (chat-driven automated site *selling*, `SAAS-001/002`,
`BIZ-009/014`). That work is about automating the *business* — payment,
provisioning, hosting a customer's finished site. This plan is about making
*this one capability* — a safe, parameterized chat box — something the
framework can place on any site the normal way, which that other research
already named as prior art worth reusing (`CHAT-001..009`). Read this plan
first if the immediate question is "how do we build this"; read the other
one if the question is "how do we sell sites via chat."

## What the framework already has, that this plan reuses rather than reinvents

Verified against live `agent_definitions` and `content_components` schema,
2026-08-11 — not assumed from memory:

| Piece | What it does today | Verified |
|---|---|---|
| `experience-planner` | Composes an `EXPERIENCE_PLAN` (journeys, promise ledger, data contracts, MVP cut, journey criteria) for a *named experience*, then runs a four-critic challenge council (journeys / feasibility / **honesty[hard-veto]** / mvp-referee) that converges or escalates. Writes travelling docs, not code. | `agent_definitions.description` |
| `tool-suggester` | Evaluates what interactive tools would benefit a site from its industry/services/audience/existing pages — LLM judgment, not limited to the catalogue. Creates `add_tool` work items. | same |
| `tool-deployer` | Forks a library tool to a site: component fork + tool page + `page_component` link, then the page flows through the *normal* render/deploy pipeline. Workflow confirmed: `ensure_site_record` → `deploy_tool_to_site` → `complete`. | workflow steps read directly |
| `tool-auditor` | LLM code review of *deployed* tools — reads full HTML/CSS/JS, finds bugs/mobile/UX/accessibility issues, creates `improve_tool` items. | same |
| `tool-improver` | Applies those fixes incrementally, re-renders. | same |
| `content_components.component_level` | Already has a real, used value `'tool'` distinct from `'section'` (`patent-check`, `funding-fit`, the loan/mortgage calculators). | live schema + rows |
| `content_components.semantic_tags` | `jsonb`, exists, inconsistently populated (`["tool","bridging","loan","calculator",...]` on some rows, empty on others). | live rows |
| `content_components.forked_from` | Real fork lineage column — `tool-ab-test-calculator_pre_037-idea-uk` is a live fork of `tool-ab-test-calculator_pre_037`. | live rows |
| CTS-044 pattern | The pattern chat-input-box already follows: external JS loader via `js_snippets`, no inline `<script>`, `data-runtime-fill="true"` marker, no server-rendered dynamic content baked into `rendered_html`. Proven safe — this is what let the loader survive being rewritten around it all session. | this session's own build |

**What every existing "tool" has in common, and where chat-input-box
diverges:** every deployed tool checked (`gripper-payload-calculator`,
`bridging-loan-calculator`, the `tool-ab-test-calculator` forks) is
**client-side only** — a calculator running entirely in the visitor's
browser. `agent_type`/`agent_workflow` columns exist on `content_components`
but are empty on every one of them. **No existing tool has a real backend.**
That is the one genuine gap this plan cannot paper over with reuse.

## Decomposition

### 1. Register it as a library tool, not a page-specific splice

`chat-input-box` becomes a normal `content_components` row with
`component_level='tool'`, a real `category` (`'chat'` or `'lead-capture'`),
and `semantic_tags` including `"tool"`, `"chat"`, `"lead-capture"`, and a
**new** tag named below. This alone is why "safely rebuild/improve/change on
any site" becomes possible at all: `tool-auditor`/`tool-improver` already
know how to review and incrementally fix a `component_level='tool'` row
without anyone teaching them chat specifically — they read HTML/CSS/JS and
reason about it generically. Today's one-off row doesn't get any of that for
free; a registered tool does.

### 2. A new semantic tag closes a real, already-identified gap: `requires-backend`

This session's own research found `VMB-010` (a `requires-backend` semantic
tag on the planner's eligibility check) was **designed, never built** — "the
column that tag would live in does not exist, and no active agent
configuration mentions it." `semantic_tags` already exists as a column; what
was missing was the *planner reading it*. Building this properly:

- Tag `chat-input-box` (and any future tool needing a live backend)
  `"requires-backend"` in `semantic_tags`.
- `tool-suggester`'s eligibility check gains one condition: a
  `requires-backend`-tagged tool is only suggested for a site whose
  `deploy_config.target` names a backend it can actually reach (see §4) —
  the same shape as the existing `capabilities:["backend"]` flag this
  session already saw used for VM-hosted sites (`PLAN_2026-08-04…`).
- This is a **platform seam** by CLAUDE.md's own test (§"Platform seams and
  the ordering exemption") — it changes what the tool-eligibility gate can
  express fleet-wide, not just for this one component. It ships with a
  concept-register entry in the same commit and goes through the council
  gate before or alongside that commit, same as this session's `LCO-008`
  cache seam. It does **not** need an RFC under the 2026-08-02 owner ruling:
  it's an opt-in field (a tag that does nothing until read, reachable by
  nothing until named) — additive-and-inert, not guarantee-changing.

### 3. The decision to place it: `experience-planner` once, `tool-suggester` per site

Two different questions, two different mechanisms, deliberately not
conflated:

- **"What is a safe, correct site chat experience, in general?"** — asked
  **once**, via `experience-planner`, producing an `EXPERIENCE_PLAN` for
  "site chat intake" as a named experience: journeys (visitor asks about
  price → gets a real answer → converts to a lead), a **promise ledger**
  (per-IP rate limit, turn cap, daily spend ceiling, fail-closed to real
  contact details — the four controls Phase 4 already built and
  mutation-tested, now stated as the *contract* rather than left implicit
  in one Go file), data contracts (exactly which per-site parameters a
  deployment needs — see §4), an MVP cut, and journey criteria. The
  four-critic council — especially **honesty (hard-veto)** — is the right
  gate for a component that can fabricate claims about price or process if
  its facts drift from `evidence_base`, which this session's own NOTES
  flagged as a live, unresolved coupling risk. This produces the reusable,
  approved spec every future per-site decision below points at, rather than
  re-litigating safety from scratch per site.
- **"Should *this* site get one?"** — asked **per site**, by
  `tool-suggester`, the same way it already decides a mortgage-advice site
  should get a repayment calculator: industry/services/audience/existing
  pages, now also checking `requires-backend` eligibility from §2 and citing
  the approved `EXPERIENCE_PLAN` rather than reasoning from first principles
  each time.

### 4. Deploy: extend `tool-deployer`, don't fork a parallel path

`deploy_tool_to_site` already does component-fork + tool-page +
`page_component`-link for the frontend shell — chat-input-box's HTML/CSS/JS
half needs nothing new there beyond being a registered tool at all (§1). What
`tool-deployer` cannot do today, for any tool, is provision a **backend**.
That capability needs to be added, and its shape is the one real design fork
in this plan:

| Option | What it is | Cost | Risk |
|---|---|---|---|
| **A — parameterize what's already proven** | The *same* Go binary as today's webdesign.uk chat-service, but reads a per-site **config file** (business facts, price, tone, contact fallback) the framework writes from `site_specs.evidence_base`, instead of Go constants. One binary can then serve several sites sharing a VM (`PLAN_2026-08-04…`'s own "our product sites" trust class already puts multiple sites on one box). | Smallest — mostly deleting the hardcoding this session's own landmine already flagged (`chat.go`'s `systemPromptFacts` has no code link to `evidence_base`), plus one small templated-config generator step. | Still one VM-hosted-sites-group per box; doesn't scale past that group without repeating deploy work. |
| **B — a shared multi-tenant chat relay** | One service, many sites, `site_id`-scoped config loaded per request — closer to `tools-api`'s proven shape (CORS-by-origin allowlist, per-caller rate limit, one Postgres) than to a bespoke binary. | Real new platform infrastructure — a genuine build, not a config change. | Directly collides with `SAAS-001`'s own warning, found in this session's earlier research: *"an anonymous, internet-triggered, token-spending pipeline must not run on core."* `tools-api` itself was deliberately kept **off** the k8s cluster for exactly this reason. Building B without the same isolation is repeating a mistake this codebase has already reasoned its way past once. |
| **C — the full satellite (`SAAS-001` "Y-copy")** | A second, cut-down chassis instance handling all chat traffic, isolated from core. | Largest — this is the option that research already flagged as *"kept open, not committed."* | Correct long-term shape if chat becomes a genuine multi-tenant product; wrong first step for "make one component reusable." |

**Recommendation: A now, B or C only if A's scale limit is actually hit.**
Option A is buildable immediately, fixes a real defect already on record
regardless of this plan, and doesn't require a platform-architecture
decision before any of the rest of this can proceed. B/C are named here so
the fork is visible and deliberate, not because either is being proposed
for this iteration — **this is the one open decision in this plan that is
the owner's, not mine, if a second site's chat box is ever needed off a
different box than webdesign.uk's.**

### 5. Maintain: `tool-auditor` / `tool-improver` already fit the frontend half

Once registered per §1, the chat-input-box HTML/CSS/JS on any site is a
normal `component_level='tool'` row — `tool-auditor` reads it and files
`improve_tool` items the same as any calculator; `tool-improver` applies
fixes and re-renders, the same pipeline that (correctly, this session
proved) leaves a *locked* row alone. **The backend half has no audit
mechanism today, under any option in §4** — `tool-auditor` reads
"HTML/CSS/JS source," not a running Go service's behaviour. Out of scope for
this plan; flagged so it isn't assumed to be covered by extension.

## Phasing

1. **Register `chat-input-box` properly** (§1) — reclassify the existing row
   as `component_level='tool'`, real `category`, `semantic_tags` including
   the new `requires-backend` tag. No behaviour change; makes the *existing*
   deployment visible to `tool-auditor`/`tool-improver` for the first time.
2. **Build the `requires-backend` gate** (§2) — the platform seam, its own
   commit, registered + council-submitted per CLAUDE.md, same discipline as
   this session's `LCO-008`.
3. **Run `experience-planner` once** for "site chat intake" (§3) — produces
   the approved `EXPERIENCE_PLAN` every later per-site decision cites.
4. **Fix the config-not-code landmine** (§4 Option A) — the config generator
   step, sourced from `evidence_base`, closing the gap this session's own
   NOTES already named.
5. **Extend `tool-deployer`** to call that generator when deploying a
   `requires-backend`-tagged tool, and prove it end to end on a **second**
   real site sharing webdesign.uk's box — the actual test of "any site,
   different parameters," not a hypothetical.
6. **Wire `tool-suggester`** to recommend chat-input-box using the approved
   `EXPERIENCE_PLAN`'s criteria — only after step 5 has proven deployment
   works, so the suggestion path doesn't outrun what deployment can deliver.

Steps 1–4 need no owner decision and no new platform infrastructure. Step 5
is where "any site" first gets tested for real. Option B/C from §4 is a
distinct, later, owner-level decision — not blocking anything above.

## Open decisions, named for the owner

1. **§4's fork** — is a second box-sharing site (Option A) enough, or is a
   shared multi-tenant relay (B) or full satellite (C) wanted sooner? Not
   urgent; nothing above depends on the answer.
2. **How wide does `tool-suggester` cast this?** Every framework site
   automatically gets evaluated for a chat box, or is this opt-in per
   product line (matching the existing "our product sites" vs "customer
   deliverables" trust-class split)?
3. **Does the honesty-critic promise ledger need Stripe-adjacent
   commitments** (spend caps, refund interaction) baked in now, given this
   session's own deposit-pricing work, or is that a later revision once
   billing is live?
