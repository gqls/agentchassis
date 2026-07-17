# RUNNING NOTES — Claims verification (authenticitycheckers)

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
