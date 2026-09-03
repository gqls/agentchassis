# NOTES — Claims verification

*The workstream's running technical log (standing-five `NOTES` doc; renamed from
`RUNNING_NOTES_*` on 2026-07-19 to match the directive). Append-only, newest at the
bottom. Missteps are recorded deliberately — they are the part the next thread
cannot rederive.*

> **NOTE ON ORDER, 2026-07-19:** entries below up to this line were appended
> ABOVE the DECISIONS section rather than at the file's end, so within that block
> the newest sits just before DECISIONS, not last. Left as-is rather than
> reordered (rewriting a log to look tidier destroys its evidential value).
> Everything from here on is strictly append-at-bottom.

Chronological; decisions promoted to DECISIONS with rationale. Method carried
from the fixloop conventions: pre-registered criteria (the §8 benchmark corpus),
evidence by artifact never by report, thin slices, reuse before recreate.

Spec: `SPEC_claims_verification.md` (same directory). Origin: leopardess rebuild.

---

## 2026-07-16 — Build session 1: V0 + V1a + V1b built and verified

### Intake verification (current state re-checked before building)

- `site_specs` had NO `evidence_base` aspect row — the key existed only as a
  prose string inside the `identity` aspect, as the spec said. `is_current`
  unique index (`idx_site_specs_current`) confirmed present.
- SQL-sourced audit values re-verified live: businesses verified = **2,767**
  (exact match to audit), companies_house_data = 937, ch_vet_companies = 5,798,
  agent_definitions (not deleted) = **157** (B7's "156" has grown by one; "over
  150" still true), active = 60, sites deployed = 9 (audit said 8; one added),
  feed items = 6,262 / scored 5,228, orchestration_states = 90,790.
- **The spec's §8 baseline assumption ("the live site is currently clean —
  expect zero findings") was FALSE.** Pre-build text scan of live
  `page_components` found banned claims alive on the site; the finished
  checker then found more (see live run below).

### ★ Headline: the first live run caught 9 genuine findings on 4 pages

`cmd/claimscan` (same shared engine as gate + audit) against live unlocked
leopardess `page_components` + `site_components` (95 components), evidence base
row `f469e0aa`:

| Page | Slot | Finding | Class |
|---|---|---|---|
| for-engineering-teams | features | "more than 70 agents across **eight departments**" | U1 + U2 |
| for-engineering-teams | generic-text-block | "**70+ agents** covering **eight business functions**" | U1 + U2 |
| hierarchical-multi-agent-orchestration-explained | article-body | "runs **70+ agents** across **eight functional departments**" | U1 + U2 |
| technical-architecture | generic-text-block | "more than **seventy agents** operating in **eight functional areas**" | U1 + U2 (word-form!) |
| insights | generic-text-block | "**least-privilege IAM** in containerised environments" | U10 |

Zero false positives after one precision fix (see DECISIONS). Notable:

1. `for-engineering-teams` is THE orphan page from spec §1.2 — the fabrication
   is still/again sleeping there, in `content_data` (the render source), not
   just baked HTML.
2. The hierarchical guide's `article-body` was **written 2026-07-15** — five
   days AFTER the fabrication sweep declared content_data CLEAN. The writer
   leaked the banned claim again. Prevention-by-prompt is leaky, exactly as
   the spec argued; V2 (whitelist injection) earns its place.
3. The word-form variant ("seventy agents … eight functional areas") was
   caught only because the fleet-claim regex covers `seventy`; "areas" was
   then added to the departments pattern (evidence base rev 3). The banned
   list is a living per-site register — every resurfacing widens it.
4. All findings are in UNLOCKED components; fixes belong in `content_data`
   (source of truth), then re-render.

**Owner action queued: rule on the 9 findings.** The check will create the
`claims_unverified` work items itself once deployed and enabled (or claimscan
output can drive hand-fixes sooner). No content was auto-rewritten — truth
decisions are human.

### What was built

| Piece | File | Status |
|---|---|---|
| V0 evidence base (structured, live) | `site_specs` aspect `evidence_base`, site 4851f6fc…, row `f469e0aa` (rev 3), `pinned=true`, `is_current=true`, source `hitl` | LIVE in DB |
| Shared scan engine | `platform/orchestration/datahelpers/claims.go` | built + tested |
| V1a build-time gate (check 8) | `platform/orchestration/actions/validate_page_content.go` — `checkBannedClaims` (blocker) + `checkUnregisteredNumbers` (error) + email-scan refactor to assertion contexts | built + tested |
| V1b post-deploy check | `platform/orchestration/actions/discovery_checks/check_unverified_claims.go`, name `unverified_claims`, auto-registered | built + tested |
| Benchmark corpus | `platform/orchestration/datahelpers/claims_test.go` (B1–B7 + exclusion classes) + `actions/validate_page_content_claims_test.go` (severity contract) | ALL PASS |
| Operator CLI | `cmd/claimscan/main.go` — same engine over TSV exports; exit 1 on findings (scriptable acceptance gate) | built + used |

Evidence base contents (rev 3): 18 facts (7 sql-sourced live-verifiable, 6
capabilities by artifact, attestations incl. the worldsoccernews figure with
its hedge), 19 banned patterns (the 2026-07-10 sweep's 16 + widened variants),
25 allowed entities. Full JSON: query
`SELECT jsonb_pretty(data) FROM site_specs WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND aspect='evidence_base' AND is_current=true;`

### Benchmark corpus results (spec §8 — pre-registered criteria)

| # | Fixture | Expected | Result |
|---|---|---|---|
| B1 | "eight departments" (+ live variants) | V1a blocker | **CAUGHT** (banned lane) |
| B2 | "more than 70 agents" / "70+" / "Seventy-plus" / inline-split `<strong>70+</strong> agents` | V1a/V1b | **CAUGHT** (all forms) |
| B3 | "2,767 Awards Won" | spec graded V3-hard | **CAUGHT at V1** — the banned label pattern ("awards won") fires; the number lane correctly stays silent (2,767 is registered) |
| B4 | Named fake clients / unnamed prose case studies | named: V1; prose: V3 | named **CAUGHT**; prose **SILENT by design** (documented miss → V3's lane) |
| B5 | jane@company.com in body text | still caught post-refactor | **EXTRACTED** (text node) |
| B6 | jane@… in placeholder= attribute | must NOT flag | **NOT extracted** (attribute = non-assertion) |
| B7 | "over 150 agent definitions" (157 live) | must pass; V4 stale-checks | **PASSES** (gte fact); drift >157 flags |

Plus a false-positive suite (all silent): phone numbers, years, dates,
guide hypotheticals ("100,000 daily calls"), formulas ("× 100"), Band/Tier
labels, list ordinals, 24/7, versions, currency, reading times.

## 2026-07-16 — Ruling session: all 9 findings remediated; specs de-poisoned; scan now CLEAN

Owner rulings: **A** (fleet claim) → recast as catalogue capability ("can call on a
catalogue of more than 150 agent definitions"); **B** (departments taxonomy) → requote
as real functional areas (site build, design, content, research, quality audits, news,
interactive tools, business intelligence — the doc-002 agent-family grouping; areas
NAMED, never counted — a count is another rot-prone number); **C** + all
least-privilege claims → concept retired entirely, rewrite/remove everywhere.
Wording choice: "more than 150" over exact "157" on evergreen copy — the gte fact
(value 157) supports both, survives catalogue growth, and V4 will flag if the
catalogue ever drops below 150.

**Root cause of the resurfacings, found and fixed: the SPECS were poisoned.** The
2026-07-10 sweep cleaned pages but not the specs feeding the writer. Found and
replaced across TWO rounds (first pass revealed only first-match-per-spec — the
second, wide `regexp_matches(…,'g')` sweep found six more):

1. `content_direction.emphasis` + `.formatted` + `site_plan.content_direction`:
   "Secondary emphasis … 70+ agents across 8 departments …" → catalogue framing +
   explicit "never describe the catalogue as a running fleet / no departments
   taxonomy" guard.
2. `strategy.content_strategy`: "70+ specialised agents across 8 departments, named
   managing agents" → catalogue framing.
3. `briefing.about_us`-adjacent description: "coordinates over 70 specialised agents
   across 8 departments, each managed by a dedicated managing agent" → orchestrators
   decompose goals into work items, dispatch to specialists.
4. Blog-topic list (site_plan + content_direction ×2): least-privilege IAM topic removed.
5. `content_direction.terminology` whitelist: "least-privilege IAM" element removed.
6. Jargon rule: "…and least-privilege IAM do not need definition" → trimmed.
7. **The smoking gun** — a `writing_rules` entry literally instructed the writer:
   *"When referencing security, be specific about the mechanism — 'least-privilege
   IAM policies and encrypted inter-agent communication' not 'enterprise-grade
   security'."* Replaced with: name only mechanisms that actually exist. The writer
   was never hallucinating — it was obeying direction.

**Page fixes** (content_data AND rendered_html edited identically; no LLM):
faq (agent-contract framing replaces least-privilege access), for-engineering-teams
features card ("Depth Across the Catalogue") + closing text block, hierarchical
guide topology paragraph (real topology: orchestrators → work items → specialists;
department-supervisor layer removed), insights topics list (IAM → agent failure
modes and recovery), technical-architecture Kubernetes paragraph (short-lived jobs
with resource limits), features card ("Isolated agent workloads": own pod, own
topic, per-type resource limits — all true) and catalogue-breadth paragraph.

**Evidence base rev 4** (row `111b07af`): `least-privilege iam` widened to
`least[ -]?privilege` per owner ("the concept has long since gone").

**Deploy**: five `page_rerender` items (`triaged`, `item_key page_rerender:<page>`,
handler page-rerender, plain re-assemble — no `reason`, deliberately: copy already
corrected in both columns, keep the LLM away). Dispatch loop deploys via commit →
GitHub Action → S3.

**Acceptance artifact**: claimscan re-run over fresh export (95 components, rev-4
evidence base): **0 findings, exit 0**. content_data sweep across all pages: CLEAN.

**AI team personas — RULED 2026-07-16 (round 3): delete, NO ban.**
`briefing.leadership_team` held three named AI personas ("Archivist — Head of
Research (AI Managing Agent)", "Sentinel — Head of QA", "Quartermaster — Head of
Operations") — U7-class, dormant (never on a live page). Owner ruled delete without
banning (personas are not a banned-claim class for this site; a transparently-
labelled AI-persona page remains a possible future choice). Inspection of the full
array also found **entry 1 was "Peter Grenfell — Founder & Principal Consultant"** —
the U5 invented person, ordered deleted 2026-07-09, still alive in the briefing spec
wrapping the owner's real background (and the C4-overstated "each named agent is
itself an orchestrator managing teams" claim). Deleted under the standing U5 ruling.
Applied: `leadership_team` → `[]`; site_plan hero-image prompt dropped "with AI
managing agents extending capability"; team-page direction rewritten to "presents
the real founder profile only — no invented staff, and no AI personas presented as
team members"; the "AI managing agents as social proof" sentence removed from
content_direction (structured + formatted) and site_plan's nested copy. Wide sweep
across all current specs (archivist|sentinel|quartermaster|managing agent|peter
grenfell): CLEAN. No page re-render needed — nothing live referenced them. No
evidence_base change (per ruling).

## 2026-07-17 — DEPLOYED & PRODUCTION-VERIFIED (V1 live end to end)

Owner deployed the chassis build. Verification per house practice (against the
pod, never git):

1. **Binary**: `agent-chassis` pod (v1.0.1128 at verification time) —
   `strings /app/agent-chassis` contains `unverified_claims`,
   `checkBannedClaims`, `checkUnregisteredNumbers`, `ScanBannedClaims`,
   `ExtractAssertionText`, `claims:site_components`, "Failed to load
   evidence_base spec". The gate (V1a) is therefore live for any site with an
   evidence_base row — leopardess only, today.
2. **V1b enabled**: `"unverified_claims"` appended to
   `quality-discovery-agent`'s `run_checks` checks array (home chosen for the
   `placeholder_contact` precedent — the fabricated-content sibling). Workflows
   re-read per message → live immediately.
3. **First production runs** (2026-07-17, leopardess): improvement-loop run —
   discovery phase executed the check (40 log mentions in the
   quality-discovery pod during the run), **0 `claims_unverified` items** (the
   correct result on the cleaned site). The loop itself later FAILED at
   `call_site_review` ("timed out after 3 retries") — pre-existing
   site-review-agent timeout class, downstream of and unrelated to discovery.
   A second, single-agent quality-discovery run: orchestration **COMPLETED**,
   0 claims items, 0 discovery items of any type. Note: a clean check leaves
   no positive log/result trace by design (checks emit findings only) — clean
   runs are evidenced by COMPLETED + zero items.
4. **Commit hygiene note**: this workstream's code was committed by concurrent
   sessions (core under `87d13b864 "claims verification further development"`,
   remainder under `d076c3c8e`/`f51a7accc`) before the repo-root CLAUDE.md
   commit-per-task rule was followed here — exactly the swept-WIP mode that
   file documents. Forward-only; nothing lost; the deployed binary carries the
   final code including the plural-"businesses" precision fix.

**V1 is done.** Remaining phases: V2 writer whitelist injection, V3
claims-auditor, V4 freshness; vetcomparison as second site. (Concurrent WIP
observed in this directory: `SPEC_voice_tells_check.md` +
`datahelpers/voicetells.go` — another session's sibling check; not this
task's files.)

## 2026-07-17 (later) — V2 BUILT & PRODUCTION-VERIFIED; V1b's first organic catch

**V2 (writer whitelist injection) is live — both pieces are DB config, no image
dependency:**

1. Evidence base **rev 5** (`6f088a53`): new `writer_block` field — a
   writer-readable verified-facts whitelist (numbers with their meanings and
   dating rules; capabilities without numbers; allowed entities). Follows the
   `content_direction.formatted` precedent: formatted at write time, referenced
   by the prompt. Maintained by whoever edits the evidence base (V4 can automate).
2. `page-content-writer` prompt template: conditional block "## Verified Facts
   (the ONLY numbers and named entities you may assert about this business)"
   inserted before the rewrite-guidance section — renders only when
   `site_specs.specs.evidence_base.writer_block` exists (fleet-safe; leopardess
   only today). It explicitly overrides STRICT RULE 14 for listed numbers:
   bounded "use only these" replaces unbounded "don't invent".
   **Banned claims are deliberately NOT injected** — naming the phrases would
   seed them into generation context; the deterministic gate stays their backstop.

**End-to-end proof** (needs_page rebuild of for-engineering-teams, item
`7100be9a`, dispatched 19:46): all **5/5 writer LLM calls carried the whitelist**
in `llm_call_log.prompt_rendered`; **0 validation blockers**; item complete;
claimscan on the rebuilt page: **0 findings**. Best artifact — the writer used a
whitelist fact verbatim WITH honest dating, unprompted beyond the block:
*"Our own systems have produced more than 90,790 orchestration state records to
date (live count, 2026-07-16)."* That is exactly the bounded behaviour V2 exists
to produce.

**V1b's first organic catch (same day it went live).** A concurrent
tool-pipeline thread updated llm-cost-calculator's `tool-cta` at 16:42 with
"Directly maps to the Digital Transformation Strategy and Multi-Agent Systems
services" — the retired register (D4), AND no service by that name exists in the
briefing (verified). The scheduled discovery cycle caught it at 19:43 (~3h
drift-to-detection) and parked `claims:c67ed17b…` as needs_human_review.
**RULED & FIXED 2026-07-17**: owner approved the suggested rewrite; tool-cta now
maps to the two REAL services ("AI Strategy & Architecture Consulting",
"Hierarchical Multi-Agent Orchestration") in content_data + rendered_html; item
completed with the ruling in `result`; `page_rerender:llm-cost-calculator`
queued for deploy. Full drift→detection→human ruling→fix cycle closed on real
concurrent traffic. Read-out summary added:
`SUMMARY_where_we_are_claims_verification.md`.

Remaining phases: V3 claims-auditor (prose lane), V4 freshness; vetcomparison
second site.

## 2026-07-17/18 — V3 BUILT & FIRST RUNS: both paths proven; one platform bug found

**`claims-auditor` agent live** (analyst, modelled on content-quality-auditor):
ensure_site_record → load_evidence_facts (register + allowed entities formatted
in SQL) → conditional opt-in gate → load_page_text (tag/style/script-stripped
text of all deployed unlocked pages, 3,500 chars/page) → ONE LLM pass
(claude-sonnet-4-6, `ai_service` at STEP level only — the fixloop gotcha: a
root-level ai_service SHADOWS step config) → conditional on findings →
`create_work_item` (`claims_unverified`, `item_key claims_llm:<domain>`,
status `needs_human_review`, handler `human-review`) → complete.

Prompt encodes the audit's caveat semantics: a fact supports its WORDING not
its topic ("handles dissolved companies" fails though a CH fact exists);
could-framed offers are fine and unreported; true-number-under-false-label is
unsupported (B3); entity relationships must trace to the allowed list or a
fact. Reports at most 12 unsupported assertions, worst first; clean = literal `[]`.

**Platform bug found on first dispatch:** `checkpoint_for_review` — the
documented HITL checkpoint action — was NEVER registered in registry.go, so
any workflow referencing it fails validation with "requires a topic". Its own
file header shows the intended registration; nothing else ever used it (dead
since creation). Worked around with the registered, dedup-aware
`create_work_item` (same HITL-terminal shape: needs_human_review +
human-review pseudo-handler, the checkpoint action's own convention);
**registry entry added in Go** (inert until next image) so the documented
mechanism exists for future HITL workflows. Lesson re-learned: a header
comment's "Registration:" block is an intention, not a fact — grep the
registry (001's dated-claim discipline, against myself this time).

**First runs (verify-by-artifact):**
- leopardess (opted in): orchestration COMPLETED; one LLM call; response was
  the literal `[]` (4 output tokens) → conditional skipped item creation.
  Clean-pass-produces-zero-noise proven on the freshly-scrubbed site.
- robot-hands (no evidence base): COMPLETED with **zero LLM calls** — the
  opt-in gate short-circuits before any cost. Landmine 7 (never fleet-wide on
  data only one site has) holds by construction.

**Honest boundary:** the catch path (auditor actually flagging a fabrication)
is NOT yet demonstrated — I would not plant fabrications on a live site to
test it. It will be proven the way V1b was: by the first real drift. Precision
is likewise unmeasured until then; the prompt is the "design iteration"
surface the spec predicted.

**Not yet wired into any cadence** — the agent exists and is invocable
(spawn/call), but adding it to an audit schedule (cost: one LLM call per site
per pass) is an owner call. V4 (freshness) remains, plus vetcomparison.

## DECISIONS (with rationale)

1. **Shared engine in `datahelpers`** (`claims.go`), consumed by both the gate
   and the discovery check — the ExtractHrefs/PageURLSet precedent: gate and
   audit agree by one literal implementation.
2. **Assertion text nodes only, via `golang.org/x/net/html`** (landmine 1).
   Skipped subtrees: script, style, code, pre, noscript, template, svg,
   iframe, textarea, select, option, head; comments and attributes never
   scanned. Inline elements concatenate (catches `<strong>70+</strong>
   agents`); block elements delimit. **mailto: hrefs are the one attribute
   surface treated as an assertion** (published contact) — B5 keeps catching,
   B6 stops false-positiving. `validateEmails` now consumes
   `ExtractAssertionEmails`.
   *Recorded V1 boundary:* alt/title attribute text is user-visible but not
   scanned; scanning it is a V3-time decision.
3. **Schema addition: `context_terms` on facts (rev 2).** A `gte` fact with a
   large value would otherwise blanket-support every smaller number on the
   site ("we support 12 clients" must not pass because 12 ≤ the 90,790
   orchestration-records count). Non-exact tolerances require a window term
   match; a non-exact fact without terms degrades to exact, never to blanket
   support. Tested (`TestNumberScanContextTermScoping`).
   *Known precision boundary:* window-scoped gte still over-supports smaller
   numbers sharing the fact's context ("56 … active" supported by the
   catalogue fact). Claim-wording precision is V3's lane.
4. **Number-gate is high-precision by claim-noun window** (landmine 2),
   widened only from measurement. First live run found one FP class:
   singular "business" gating "22 for business hours" (calculator help text).
   Fixed: plural-only "businesses" (count-claims use the plural; "business
   records" still gates via "records"). Word-form numbers ("eight
   departments") are the banned lane's job, not the number lane's.
5. **Severities per spec:** banned = blocker (known falsehoods, human-ruled
   onto the list); unregistered number = error (extractor has FPs; error
   already routes to `mark_needs_review`). Answers spec open Q1: error, not
   warning — the one FP found was an engine fix, not review noise.
6. **HITL routing (answers spec open Q4):** `claims_unverified` items are
   created with `status='needs_human_review'`, `HandlerAgent=""` — the same
   HITL-terminal shape as `check_required_fields_missing` /
   `check_section_source_drift`. They reuse the existing needs_human_review
   queue; no new surface. Item keys `claims:<page_id>` /
   `claims:site_components` (dedup like `page_rerender:<page>`). Severity
   high + priority 25 when any banned claim (outranks placeholder_contact's
   30); medium/35 for numbers-only.
7. **Opt-in per site** on `evidence_base` presence (spec landmine 7): both the
   gate's check 8 and the discovery check no-op without a current row —
   nothing fleet-wide changes until a site gets an evidence base.
8. **Banned patterns are case-insensitive regex with literal fallback** — an
   invalid regex degrades to a QuoteMeta substring so a typo can never
   silently drop a banned claim (tested).
9. **Locked components skipped** (`locked_at IS NULL`) in V1b, per
   check_placeholder_contact precedent (landmine 5).
10. **created_by = 'operator-claude', source = 'hitl'** on the evidence-base
    rows: transcription of owner-approved audit data by the operator agent;
    each rev's `notes` records what changed.

## Deploy & enable (NOT yet done — deploy is owner-gated)

1. Build/deploy the chassis image per house practice (Makefile from local
   filesystem; verify against the pod, never git).
2. Gate (V1a) activates by itself for leopardess once the image carries the
   code — check 8 is opt-in on the evidence_base row, which is already live.
   Config kill-switch: `check_claims: false` on the validate step.
3. Enable V1b by adding `"unverified_claims"` to the relevant discovery
   agent's `checks` array (agent_definitions) — AFTER the image ships, so the
   registry knows the name.
4. First production run should reproduce the 9 findings above as
   `claims_unverified` work items (dedup keys `claims:<page_id>`).
5. Re-run acceptance: `go run ./cmd/claimscan -evidence <eb.json> -components
   <export.tsv>` (invocation in the CLI header) — exit 0 once the owner has
   ruled on and fixed the 9.

## Next phases (unchanged from spec §5/§9)

- **V2** — writer whitelist injection ("use ONLY these facts") in
  page-content-writer's prompt where research findings inject. The 2026-07-15
  leak is the motivating exhibit.
- **V3** — `claims-auditor` agent (LLM lane) for prose assertions (B4's
  unnamed case-study class), modelled on content-quality-auditor; also the
  lane for claim-wording vs fact-topic precision (B3-class, "56 active"
  boundary) and alt/title text.
- **V4** — freshness: re-run `sql` facts on a schedule, update
  value/verified_at, `stale_evidence` items on tolerance drift (underclaiming
  included). Shares the query layer the L7 chart component wants.
- **Second site:** vetcomparison.uk (price-claims class → licensed source
  rows), per spec §1.
- Spec open Q2 (fleet-shareable banned patterns) and Q3 (auto-update of
  exceeded numbers) remain owner decisions; nothing built assumes either.

---

## 2026-07-20 — V4 activated end-to-end; first freshness finding is REAL

**Gap in my own work, found by checking rather than remembering.** The live
evidence base was still rev 5: `writer_block_managed` unset, zero facts carrying
`writer_line`. I generated rev 6 on 07-18 and never applied it — diverted into the
council submission. V4's whitelist regeneration would have silently no-op'd
("unmanaged"). Applied as rev 6, spec row `00b01b38`, after a compare-and-swap
check that rev 5 was still current (it was).

**The image already carried V4.** `v1.0.1139`, 11 symbol hits in the pod — someone
built after `06376bcbf`. The "waiting on an image build" state I had been reporting
was already stale. Verified against the pod, per house rule, not against git.

**Dry run (orchestration `8ec0743e`).** 3 sites swept, 9 sql-facts, 0 errors,
`dry_run` wrote nothing (spec row unchanged, no work item). Whitelist regeneration
fired. Seven facts re-synced within tolerance; all the `gte` ones had grown.

**★ The finding: `C1-records-verified` drifted DOWN, 2,767 → 2,291.** Exact
tolerance, so it breaches. Sanity-checked the underlying data directly rather than
trusting the action: `business_intel.businesses` now reads verified 2291,
**dismissed 874**, pending 238, seed_import 16 — records were reclassified out of
verified since 07-16. **The live site is overclaiming by ~476.** Owner ruling
needed; recorded at the top of the handoff. This is the first thing V4 has caught
and it justifies the phase on its own — the number had been stale for days with
nothing to notice it.

**Seed applied** (`evidence-freshness`, daily, enabled) only AFTER the dry run
passed, per the gate I set myself.

**Two dispatch missteps worth recording.**
1. First dry-run attempt set `config.agent_type: "generic"` — that loads the
   **no-op generic agent definition** instead of running the inline workflow.
   Send `config.workflow` alone.
2. I then concluded the second attempt had also failed, because no orchestration
   row existed after 4 minutes. It landed later: the chassis was on ONE replica and
   saturated with other threads' council runs. Same lesson as the council and
   diagnosis runs — **"no row" means "not yet"**, and on a shared cluster your work
   queues behind everyone else's.

**Other sites are opting in.** The sweep found `evidence_base` specs on vonc.com
and relojistas.com (0 sql-facts each) — other threads have started using the layer.
No seed change needed; the sweep covers them.

---

## 2026-07-20 (later) — V5 BUILT: citation evidence, verified by verbatim quote

Owner directed: site numbers "verified from web deepsearch cited references, so
not manual but part of the chassis' capability". Spec'd
(`SPEC_V5_researched_citations.md`) then built the same day.

**What was built** (all tests pass; Go inert until an image ships):
- `datahelpers/citations.go` (+tests): the pure half. ParseCitation (url+quote+
  publisher mandatory — an unverifiable citation must never pose as evidence);
  NormalizeForQuoteMatch (entities, curly punctuation, nbsp, dashes, thousands
  separators, case — forgiving on PRESENTATION, strict on CONTENT: 411 matches
  "411&nbsp;", never 412); QuoteFoundInText over VisibleTextFromHTML (reuses
  ExtractAssertionText, so a quote living only in <script>/<pre> cannot verify).
- `actions/evidence_citations.go` (+tests): the network half. Fetch is http(s)-
  only, 25s timeout, 4MiB cap; PDFs REFUSED (half-reading a binary would fake
  verification → reverifiable:false + human attestation is the honest route).
  Failure classification is the safety property: 200-but-quote-gone =
  `citation_lost` (drift); 403/network/PDF = `fetch_error` (unknown — a paywall
  going up is not evidence the fact is wrong). Proven offline via httptest.
- V4 extension: the freshness pass re-verifies citation facts each run — fresh
  bumps accessed/verified_at; citation_lost or past `staleness_days` (aged from
  the PUBLICATION date, not the last check) = drifted → the existing
  stale_evidence HITL route; errors are errors.
- `verify_and_register_citations` action (registered): the acquisition
  endpoint. LLM-extracted candidates are verified LIVE before anything is
  written — **the model proposes, the string comparison disposes**. Verified →
  citation facts (register auto-created for a site without one; CAS supersede
  when one exists). Rejected → ONE citation_unverified item at human review,
  never a fact. IDs are fnv(url + normalised quote) so re-runs are idempotent.
- `SEED_evidence_researcher.sql` (NOT applied — names the new action, so image
  first): search → prepare_urls (primary publishers preferred) → scrape →
  extract ATOMIC claims with verbatim quotes (the prompt warns the model its
  quotes will be machine-checked and paraphrases fail) → verify_and_register.

**Design notes for the next reader:**
- The whole point is that verification is a string comparison, not a judgement —
  the same shape as the email check that founded the layer. Everything
  downstream (V1 flags unregistered numbers, V2 whitelists + attributes via
  writer_line, V4 re-checks) already existed, so V5 is acquisition only.
- Council submission f5ab4fb5 sent with a deliberately TERSE plan (bug 019
  punishes verbose rounds — reviewer output scales with submission size).

---

## 2026-07-20 (evening) — V5 ACTIVATED on v1.0.1140; smoke run found bug 047

- Fresh chassis (v1.0.1140) pod-verified to carry V5 (verify_and_register_citations,
  citation_lost classifier). The V5 council submission (f5ab4fb5) was LOST in the
  deploy restart after an hour queued — never reviewed; recorded here, not resubmitted.
- SEED_evidence_researcher.sql APPLIED (image first, then seed). Agent active.
- Smoke run (correlation f930dc2f, "global LNG trade volume 2024"): search and
  prepare worked — 4 primary sources chosen (IGU World LNG Report, EIA, IEEFA ×2)
  — then FAILED at scrape_pages: awaited request expired at retry=3, no step
  error. Adapter pod log had the truth in one line: "Empty URL in request".
- Root cause = bugs_open/047: the webscrape adapter validated top-level url
  BEFORE the action switch; batch_scrape has no top-level url by construction,
  so EVERY batch scrape ever sent was rejected at the door — including
  research-agent's scrape step (the writer's whole research lane). Fix
  committed (guard reordered); OPEN until a web-scrape-adapter image rolls.
  §9 pattern added to 016b.
- V5 smoke therefore PENDING the adapter image: re-run per RUNBOOK §7 with
  {site_id, domain, research_query}; expect the gaswholesalers register to be
  created by its first verified facts.

## 2026-07-26 — V4's freshness sweep is LIVE, and had never run once (bugs_closed/074, another thread)

Left by the bugs_open/074 session. **Your `evidence-freshness` task had never executed its
sweep.** It carried its workflow inline at `input_data.config.workflow` — a shape the scheduler
cannot deliver (it builds the message `config` from the row's columns, so the workflow landed a
level below the only reader), so `generic`'s one-step no-op ran instead and every fire stamped
both timestamps and did nothing. The check that shows it, and that no timestamp can fake:

```sql
SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_evidence_base%';
--  0 for the whole life of the task; 3 after the repair
```

**Repaired the way model_directory repaired the identical shape:** a new `evidence-freshness`
`agent_definitions` row now carries the workflow verbatim, the task points at it, `input_data` is
`{}`. `SEED_evidence_freshness_scheduled_task.sql` in this directory has been rewritten to match,
so a re-seed cannot reintroduce it, and migration `217` now refuses the shape with a CHECK
constraint.

**First real pass, staged behind a `dry_run: true` pass first** (it rewrites `writer_block` on
sites other threads are actively working): 8 sites swept, 24 sql-sourced facts, 0 errors, **3 new
`site_specs` revisions** as `evidence-refresher` with `pinned` preserved, and **3 `stale_evidence`
items** raised for human review — fundamentallyai (3 facts), ai-agent-orchestration (3),
leopardess (1). All genuine: council round counts +1, live sites 12→14, orchestrations/day down.

**Then an induced fault**, because a green pass proves deployment and not detection: one leopardess
fact set to 9,370 against a live 937 came back `drifted` with the right detail, and the number
re-synced itself.

**Two things for you:**

1. **`bugs_open/091`** — found by that induced fault. While an earlier `stale_evidence` item is
   open, a second, *different* drift is dropped by the work-item dedup (`ON CONFLICT … DO NOTHING`
   on `stale_evidence:<site_id>`) and the run still reports `work_item_created: true`, because
   `createStaleEvidenceItem` throws away the `inserted` bool it is handed. The contained half of
   the fix (propagate `inserted`) is yours; the shared-helper half belongs to
   work_item_completion_integrity. Not touched here.
2. **The sweep supersedes the spec** (`is_current=false` + INSERT, `refresh_evidence_base_action.go:669-693`)
   and now runs **daily**. Any code or session holding a `site_specs.id` for an evidence base must
   re-SELECT the current row before writing, or the write lands on a dead revision.

## 2026-09-02 — this session picks up as "the claims verification thread" (owner renamed it); RFC_060 arc, cross-session with `bugfix_414_planted_marker_as_claim` and three site lanes (lendzy, loanzy.uk, loancalculator.co.uk)

Owner asked, out of the poisoned-register question: *"shouldn't compliance be strong for sites that
require compliance strongly — finance, legal, insurance?"* — routed to `RFC_060` by the
`bugfix_414` lane (its own thread had closed as a bug 08-31 and continued past that into this design
question same day). This session read in, oriented (V0–V5/CLM-025–030 already live, cold-audit gap
already closed), then owned the RFC's live evolution for the rest of the day. Full technical detail
lives in `RFC_060_compliance_tier_the_claims_layer_is_weakest_where_the_sector_is_strictest.md`
(architecture_review/) — this entry is the log of HOW the day went, including what was wrong first.

**Owner ruled Q1/Q2/Q4 same-morning** (require registers; build order register-gate→fact-floor→
severity; a signed record not a boolean). **Q3 (vocabulary) approved after one discussion round** —
a three-rung posture ladder (`standard`/`sourced`/`relied_upon`), named for claim RELATIONSHIP not
sector, after the owner's own steer ("maybe it should have a semantic decision layer rather than
sector specific").

**Two structural gaps found by reading the code, not by assumption, both folded in as new
addenda (Q5, Q6):** the citation-code exclusion (`fad209b92`) is a single hardcoded FCA-only Go
constant, not per-site data like `banned_claims`/`allowed_entities` — confirmed `vetcomparison.uk`
is live, deployed, zero register, and would reproduce the finance false-positive on its first RCVS
citation. And a citation can be true, sourced, and still name the wrong rule (Q6) — traced with the
lendzy lane to a real structural cause (the FCA Handbook has no rule-level URL; a 54-rule chapter page
lets a quote genuinely belonging to one rule verify against a citation labelled a neighbouring one).

**Then the register-writing programme actually ran, same day, four site lanes** — and every lane
that read a cited source rather than trusting existing prose found an error: 5 wrong live claims
(lendzy's Q6 pair, loanzy's MaPS mis-classification, loancalculator's settlement-period and ERC
figures), all real, all now corrected or routed to the owner. That number is the base rate this
whole layer's design is being justified against, not a worst case.

**A four-round argument with the `bugfix_414` lane over one afternoon produced the banned_claims
compile-check** (`e5b1a0f01`, council-approved `bc3697a5` — read the verdict directly, not the
paraphrase: 3 advisory objections, none high-severity, `abstained:6`). They first argued a clean
census (239 patterns, 0 broken) didn't justify new surface; the loancalculator lane argued them out
of that position ("practice-as-remedy is prose-as-control", RFC_006's own precedent); I built it.
**Two of the council's objections are worth a line for the next reader:**
- `prior_art_librarian` asked whether anything already validates per-site banned_claims compile-
  ability before this — checked afterward (not just asserted): every consumer of `BannedClaims`
  (`discovery_checks/check_unverified_claims.go` included) goes through `ParseEvidenceBase`'s SAME
  silent fallback at `claims.go:348`. No other check exists. The absence claim holds.
- `tooling_provenance` objected that this landed with no doc_plans/doc_notes check beforehand and no
  NOTES entry planned after — fair, and this paragraph is that entry, late rather than absent.

**Two of my own wrong calls, both caught by the peer lane re-measuring rather than by my own
re-read — logged in `WRONG_CALLS.md` in full, summarised here so this NOTES file carries the
pattern too:** (1) told a peer session "39 citation facts, measured 2026-09-02" — real number, wrong
date; it was the 2026-08-09 SUMMARY's count, and my own RFC table from that SAME morning already
said ~192 (`622443862`). (2) told a peer lendzy's register was "built" when my own RFC table, three
sections above the sentence I wrote, already had the correct written-not-applied distinction
(`465d86951`). **Both times the error was restating my own prior work from memory instead of
re-reading it** — a document authored earlier in the SAME session is a source, not something safe to
paraphrase.

**One test bug worth a line on its own: my first cut of the banned_claims-compile-check's
regression test was VACUOUS.** `TestInvalidBannedClaimPatternItemsHonoursExistingOpenItem` asserted
`created == 0` on a conflict — true under BOTH the correct policy (`dropOnConflict`) and a wrongly
mutated one (`refreshOnConflict`), because I had not called `mock.ExpectationsWereMet()`, so the
extra unmocked UPDATE a wrong policy would issue went unchecked. Caught by mutating the code
(flip the policy, confirm the test fails) in an isolated `git worktree` — needed because another
session had unrelated uncommitted WIP (`cta_label_universe.go`, a signature change) breaking `go
test` for the whole `actions` package at the time; the worktree let me verify against clean HEAD
plus my own files without touching their tree.

**Deployment status, stated plainly because it is the gap my own summary to the user first left
out, caught by the peer verifying at the pod rather than taking my report:** `e5b1a0f01` is
COMMITTED, not RUNNING. `[MEASURED ~19:40]` both `agent-chassis` pods started 15:53Z/15:39Z,
before the 18:30Z commit — the new check is not in the deployed binary. It needs an image build +
roll (owner's call — releases are whole-fleet), and even once rolled it is inert until the NEXT
daily `evidence-freshness` tick (~09:09). Two gates, and "built" must not be read as "covering the
fleet" until both clear.

## 2026-09-03 — Q5/Q6/Q7 all built and council-submitted; a demand-control probe is planted and PENDING tomorrow's 09:10 UTC tick

Full detail lives in `RFC_060` (§3f onward) and this session's memory — this entry is so the
**pending action below survives a session boundary**, since it cannot resolve until a scheduled
tick that may fall after this session ends.

**Owner ruled Q5/Q6/Q7 directly** (held for direct confirmation rather than acting on an earlier
relay — see the RFC's own note on this). All three built same day: Q6 (`ac670badf`, council
`57a9939f` APPROVED, one follow-up fix `3c1e1b61c` after a real objection about marker letters),
Q7 facts half (`6ec879212`, council `17fb9105` APPROVED unanimously), Q5 (`939593e4c`, council
`9b11752c`, verdict pending). None of the three is deployed yet.

**PENDING ACTION — check this first if picking up after 2026-09-03 09:10 UTC (tomorrow's date):**
A demand-control probe is live on `buytoletcalculator.uk` (status='test', 0 pages, harmless):
`site_specs` row `623c1de8-6893-4700-b0c6-88f177cb955c`, one deliberately-broken `banned_claims`
pattern (`guaranteed(`). Planted 2026-09-03 09:34 UTC to prove the `e5b1a0f01` write path
(`createInvalidBannedClaimPatternItems`) fires outside mocks — the one arm still unproven in
production. As of 13:59 UTC the daily `evidence-freshness` tick has not run since 09:10:22 (before
the plant), so nothing has been tested yet.

**Check (per the `bugfix_414` lane's design, verified independently before trusting it):**
```sql
SELECT last_completed_at FROM scheduled_tasks WHERE name='evidence-freshness'; -- wait for > 2026-09-04 09:10 UTC
SELECT count(*) FROM site_work_items WHERE item_type='invalid_banned_claim_pattern'; -- expect >= 1 once ticked
```
Two arms in the next tick: **write** (`buytoletcalculator.uk`, expect `patterns_checked=1,
invalid=1` in the log + one work item — proves the write path) and **read** (`loancash.co.uk`,
freshly registered today with 19 facts/6 valid `banned_claims`, expect `patterns_checked=6,
invalid=0` — proves the check reads real data and correctly finds it clean, no plant needed). If
loancash logs 6 but the probe files nothing, the defect is isolated to the write function
specifically, not the check.

**THEN REVERT THE PROBE** — `UPDATE site_specs SET is_current=false WHERE id='623c1de8-6893-4700-b0c6-88f177cb955c'` (or delete the row; it has no history value). It is harmless today but is a live register inside the daily fleet sweep and will be re-read until removed. Do not leave it in place after the answer is known.
