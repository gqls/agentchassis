# HANDOFF — Phase C PROVEN AT THE ARTEFACT: the directory page is deployed and rendering cited lenders — 2026-08-18, continue here

Supersedes `HANDOFF_2026-08-17_continue_here.md` (accurate on its own history; this file
carries everything a fresh chat needs). Owner rulings unchanged: P9 six decisions, pilot =
remortgagecalculator.uk (M4), build order M→B→I, B8/B9/I10 HOLD, bug 270 hands-off,
copy-voice work lives in session "copy quality two stage".

## 1. WHERE THIS GOT TO — the whole objective is met

**The chain is closed, verified at the RENDERED HTML rather than at a status.** The
`mortgage-lenders` page is `build_status='deployed'` with `deployed_at` set, and its
`mortgage-lender-directory-listing` component (4,882 chars) renders:
- heading *"UK mortgage lenders, listed"*;
- the owner\'s NON-PRICE ruling surviving into the copy: *"It does not list rates, fees or
  APRs, because those change daily and depend on your circumstances"*;
- real cited entries — **Mansfield Building Society**, **Family Building Society** — each
  claim carrying `lender_type` / `product_types` and a **`source`** link.

researcher → quote-verified claim → register → kind-aware publish → `evaluate_directory_features`
flag → planner rule → page → deploy → **a live page naming regulated firms with citations and
no prices.** That was the Phase B/C question and it is answered.

**All proof points passed** (432 flag ✅ · 433/441 page name+type+composition ✅ · zero
`PLAN_SECTION_NAME_DROPPED` ✅ · directory checks silent for the right reason ✅).
**Cost baseline (CORRECTED 2026-08-18): 73 LLM calls · 663,759 in · 184,596 out · 11 assets
= ~$3.81/domain today, ~$4.83 from 2026-09-01** when Sonnet 5's introductory rate ends
(~$534 vs ~$677 for 140 domains). **The 43-call figure first published here was ~70% low** —
it sampled a set still being written (`collected_data` fills in as a run progresses). Excludes
images (different provider, unmeasured) and assumes no cache discount. See NOTES 2026-08-18.

## 1a. URLs — what you can actually look at

**LIVE AND VIEWABLE (verified by reading the served BODY, not the status code):**
| directory | URL |
|---|---|
| AI Model Directory | `https://ai-agent-orchestration.com/model-directory.html` |
| Enterprise AI Agent Adoption Tracker | `https://ai-agent-orchestration.com/adoption-tracker.html` |
| Agent Communication Protocol Tracker | `https://ai-agent-orchestration.com/protocol-tracker.html` |

**THE PILOT IS NOT VIEWABLE — and a `curl` status code will LIE to you about it.**
`https://remortgagecalculator.uk/mortgage-lenders.html` returns **200**, and the body is
`<script>window.onload=function(){window.location.href="/lander"}</script>` — the registrar\'s
**parking page**, which answers EVERY path with 200. An `%{http_code}` probe against a parked
domain is a check that cannot fail. **Always read the body.**

**So: BUILT and DEPLOYED TO THE BUCKET, but NOT SERVED.** `pages.build_status='deployed'` +
`deployed_at` is truthful about what the pipeline did (it wrote the tree to
`b2://portfolio-sites/remortgagecalculator.uk/`); **nothing points the domain at the serving
worker**, so no visitor can reach it. **`deployed` and `reachable` are different facts** and
this lane had been conflating them.

- Serving mechanism: `scripts/cloudflare/worker.js`, objectKey = `<hostname><path>` in bucket
  `portfolio-sites`. It works where DNS points at it — that is why aao.com serves and the
  pilot does not.
- The bucket cannot be probed from here: `s3.us-east-005.backblazeb2.com/portfolio-sites/…`
  returns **401 for the known-good site as well as the pilot**, so that 401 means "private
  bucket", NOT "missing object", and is evidence of nothing.
- The pilot\'s directory content IS verified — in `page_components.rendered_html`: two cited
  building societies, per-claim `source` links, and the non-price copy.
- **A DNS/domain step is missing from the fleet-build path.** Phase E hits this on all ~140
  domains. Decide whether the pipeline owns it or it is deliberately manual — this is an
  owner question, not a bug to fix quietly.

## 2. Pilot state — built and PARTLY live

| | |
|---|---|
| deployed | `mortgage-lenders`, `next-steps`, `about` |
| `needs_rebuild` | `index`, `what-your-number-means`, `six-month-checklist` |
| `sites.build_status` | still `pending` |

**Remaining work (none of it directory-related):**
- `needs_new_component` ×3 FAILED at `store_generated_component` (3/3 attempts gone)
- `needs_rerender` ×1 FAILED (timeout, 3/3), `needs_imagery` ×2 FAILED
- 2 pages blocked by **20 × `unrendered_template` `{{end}}`** — a component leaking raw Go
  template syntax. **Checked: NOT the seeded `banned_claims`**; the claims guard blocked nothing.
- `component_validation_rejected` ×6 for `mortgages-repayment`
- **HITL queue: 10 × `unresolved_cta`, 4 × `needs_page`, 1 × `needs_section_data`** — work it,
  do not debug it.

## 3. Two corrections of mine that a fresh reader must not inherit

1. **"Does the `assets` count rise above 11?" was the WRONG decisive measurement** — I wrote it
   into the last handoff. It stayed at 11 and that is CORRECT: the 8 successful retries
   reference 8 asset ids that PREDATE the outage and created zero new rows, because the step
   that failed was the **deployer**, not the generator — a good retry ships an existing asset
   and must not duplicate it. **Before calling a measurement decisive, name the step it
   observes and check that it is the step that failed.**
2. **"A retry failed again → the outage is back" was too crude** and misfired within minutes:
   the failure was an asset-deployer *timeout* while base-tree 404s stayed at 0. **Only a
   renewed `%base tree%` error is evidence about that outage.**

Earlier in the same session I also read a pre-existing `complete:1` as a retry success. A count
already non-zero at t=0 evidences nothing at t+n — take the baseline, or watch a delta.

## 4. Infrastructure facts worth carrying

- **v1.0.1308 is live** (1305 → 1306 → 1307 → 1308). The same-tag reuse that made three days of
  commits inert is fixed; another lane measured it at *"24 code commits across ~10 lanes"*.
- **`bugs_closed/292`** — the random directory recommendation — is **FIXED AND LIVE**.
- **⚠ Never verify a fix by grepping the binary for its commit sha.** The binary carries ONE
  stamp (the commit it was built FROM), so ABSENT is the normal reading on a healthy build; a
  discovery grep for `[0-9a-f]{40}` is worse (20 hits, none a real commit). Use the
  `build provenance` log line, or `git merge-base --is-ancestor <fix> <tag-bump commit>`.
  All three of my failed probes: `WRONG_CALLS.md`, 2026-08-17.
- **The deploy outage is CLOSED** (~832 `base tree` 404s on 08-17, 13:31→16:14Z; zero since).
  **The 090 `75220928…` was overtaken and never returned a verdict — what it would have
  answered is still unknown: which component routed a no-repo site to git, and what changed at
  13:31.** If it recurs, that is the question; do not retrofit the roll as the cause.
- **B3f structural checks, volume MEASURED**: `head_essentials_missing` **247 / 8 sites**,
  `dead_internal_link_live` 6, `canonical_mismatch` 4; `structured_data_invalid` and
  `sitemap_entry_dead_live` at **0 — deliberately NOT interpreted** (clean vs never-exercised
  look identical; separating them is open work). All flag-only: a backlog to triage.

## 5. What is next in this lane — ordered

**Blocked on the owner (nothing else moves past these):**
1. **Sign-off on Phase C** — the plan makes this the gate before any Phase E wave. The number
   to sign off against is the cost baseline: **43 LLM calls · 389,406 in · 120,822 out · 11
   assets** for one site, read as a FLOOR.
2. **Decide who owns DNS.** The pilot is built and unreachable because its domain still points
   at a registrar parking page. Phase E meets this on ~140 domains. Either the build pipeline
   acquires a domain-pointing step, or it is deliberately manual and the fleet plan budgets
   for it. **Do not fix this quietly** — it changes what "the pipeline builds a site" means.
3. **Phase D decisions**, unchanged and still outstanding: `loanzy.uk` (L9) conflict with the
   webdesign lane, the B8/B9/I10 holds, and build order across the remaining domains.

**This lane can do without waiting:**
4. **Work the pilot\'s HITL queue** — 10 × `unresolved_cta`, 4 × `needs_page`, 1 ×
   `needs_section_data`. Ordinary new-site work; it is what stands between 3 deployed pages
   and 6.
5. **The three genuine build failures**: `needs_new_component` ×3 (`store_generated_component`,
   3/3 attempts gone), `needs_rerender` ×1 (timeout), `needs_imagery` ×2. Retry budget is
   exhausted on all of them, so they need diagnosis, not a re-queue.
6. **File the `{{end}}` template leak as a platform bug.** 20 blockers across 2 pages, a
   component emitting raw Go template syntax into rendered output. **Confirmed NOT the seeded
   `banned_claims`.** It is not site-specific and will recur on every build — grep
   `bugs_open/` first, it may already exist.
7. **Write the Phase-C-complete SUMMARY.** The 2026-08-17 one predates both the artefact proof
   and the reachability finding, so the series\' next entry is genuinely a new inflection.
8. **Triage the `head_essentials_missing` backlog** — 247 findings across 8 sites, flag-only.
   Intelligence, not an incident, but nobody has looked at it.

**Open questions this lane should not close by assumption:**
- The **090 on deploy routing (`75220928…`) never returned a verdict.** Which component sent a
  no-repo site to the git adapter, and what changed at 13:31 on 08-17, are both still unknown.
  The outage stopped when the roll landed; that is a coincidence in time, not an explanation.
- **`structured_data_invalid` and `sitemap_entry_dead_live` sit at ZERO.** Clean and
  never-exercised look identical and have not been separated.
- **RFC_031\'s trigger stands**: the THIRD hand-spliced `content_features` enricher must build
  the shared ordered list instead of copying the splice a third time.

## 6. Files of record

`PLAN_2026-08-12_fleet_buildout.md` · `SUMMARY_2026-08-17_pilot_built_and_the_machinery_proved_itself.md`
· `NOTES_portfolio_positioning.md` (newest at bottom — the 2026-08-18 entry has the artefact
proof and both corrections) · `README_where_we_are.md` · `MISSION_2026-08-17_…md` ·
`SEED_2026-08-17{,b}_…sql`. Migrations `432/433/434/441` (+ROLLBACKs), all applied and
council-approved. Register: `docs026_concept_register/register/directory-pipeline.md` (DIR-001).
RFC: `architecture_review/RFC_031_…md`. Closed: `bugs_closed/292_…md`.
