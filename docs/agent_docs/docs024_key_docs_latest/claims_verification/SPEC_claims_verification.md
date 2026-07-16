# SPEC — Claims verification: a fact-checking layer for generated content

**Purpose.** Start a fresh chat from exactly here to design-review and build this. It is
self-contained: the origin story, the evidence, the current-state map (verified against code
and the live DB on 2026-07-16), the proposed architecture, phases, landmines, and a benchmark.

**Status:** SPEC — approved for write-up by the owner 2026-07-16; not yet built.
**Origin workstream:** leopardessconsulting rebuild (`docs/leopardessconsulting/`), where the
absence of this layer was repeatedly load-bearing. Adjacent workstreams: fixloop
(`docs024_key_docs_latest/fixloop_eg_dartsonline/`) and the concept register
(`docs026_concept_register/`, content-governance bucket).

---

## 0. One-paragraph version

The platform generates site copy with an LLM and validates it before deploy — but no layer
anywhere compares a **claim** to **evidence**. Generation-time prompt rules say "never invent";
build-time validation checks form (placeholders, links, emails, meta-commentary); post-deploy
discovery checks structure (phantom links, empty sections). Truth is checked by a human reading
pages against a markdown file. This spec proposes making evidence machine-readable
(`site_specs.evidence_base` — the key already exists, unconsumed), then wiring it into the
three places the platform already has hooks: a build-time gate, a post-deploy discovery check,
and the writer prompt. Truth *decisions* stay human — the system's job is to notice an
unsupported claim and route it to a person, never to rewrite content on its own.

---

## 1. Why — the evidence from the leopardess rebuild

The leopardess site is the platform describing itself, rebuilt under one governing rule:
**no claim ships without a row in `docs/leopardessconsulting/AUDIT_verified_facts.md`**
(every row verified against code, live Postgres, or an HTTP response). During that rebuild:

1. **Fabrications shipped through every automated layer.** The generated use-cases page carried
   five invented client case studies with invented results. The About page carried invented
   stats ("2,767 Awards Won"). All passed generation, validation, and deploy.
2. **A removed fabrication came back.** "Eight departments" was audited out of the site in the
   L0 audit — and was found *weeks later*, alive, mid-paragraph on an orphan page
   (`for-engineering-teams`), and was nearly re-folded into a surviving page during a
   deliberate content merge. Same story for "a 70-plus running agent fleet" (audit: UNSUPPORTED;
   the orphan page claimed "more than 70 agents"). Nothing in the platform knew these claims
   were on the banned list.
3. **The one fabrication class with a deterministic checker is the one class that got caught.**
   `validate_page_content` checks emails against the site's real contact. It caught
   `jane@company.com` every single time (and, instructively, also produced a false positive —
   see §7 landmine 1). No other claim class has a checker, and no other class was ever caught
   by the platform.
4. **The evidence already has a home that nothing reads.** The rewritten leopardess identity
   spec has an `evidence_base` key. `grep -rn evidence_base --include=*.go platform/ internal/`
   → zero consumers. It's currently a prose string pointing at the markdown audit.

Conclusion: prevention-by-prompt is leaky, verification-by-human doesn't scale past one
operator's attention, and the check→work-item→handler pattern demonstrably works when a
deterministic checker exists. Build the checker.

**This is not a one-site problem.** vetcomparison.uk shipped **fabricated prices** — caught and
stripped 2026-07-15, with a legal record kept (`docs024_key_docs_latest/vetcomparison/`) and its
planned rebuild explicitly requiring "claim-licensing". That is a second site, a second
fabrication class (prices), and the first with legal exposure. It should be the second
evidence-base site after the leopardess pilot, and its price-claims class belongs in the V1
deterministic lane (a price asserted in copy must trace to a licensed source row).

## 2. Current state (verified 2026-07-16)

| Layer | What it is | What it checks | Fact-aware? |
|---|---|---|---|
| Generation | page-content-writer STRICT RULES (agent_definitions.default_config prompt) | "NEVER invent fake people / statistics / case studies / contact info" (rules 5, 13–15, 18) | Instructional only; probabilistic; demonstrated leaks |
| Build gate | `validate_page_content` (in page-build-handler, between writer and save; failure → `mark_needs_review`) | placeholders, unrendered templates, domain contamination, internal links, **emails vs site contact**, meta-commentary, text length (`platform/orchestration/actions/validate_page_content.go:405–681`) | Only emails |
| Post-deploy | 38 discovery checks (`platform/orchestration/actions/discovery_checks/`) | structure: phantom_internal_links, placeholder_contact, broken_nav_links, empty_sections, … | No |
| Post-deploy | `content-quality-auditor` agent | tone alignment, content gaps, CTA effectiveness, differentiation (one LLM call) | No — zero fact/evidence vocabulary in its config; never ran on leopardess |
| Manual | `AUDIT_verified_facts.md` + operator discipline | everything, by hand | Yes — the only layer that is |

## 3. Design principles

1. **Deterministic first.** The email check's track record is the argument. Every check that
   can be a string/number comparison should be one before any LLM gets involved.
2. **Truth decisions are human.** The system flags; a person rules. No auto-rewrite of factual
   content, ever. (This also honours the existing content-governance channel rules: auditors
   raise work items; they don't change direction.)
3. **Evidence is data, owned like a spec.** `site_specs` aspect, pinned, human-edited —
   exactly like the `direction` aspect ("only humans change" precedent).
4. **Reuse the grammar.** check → `site_work_items` → handler → HITL. No new pipeline shapes.
5. **History is the benchmark.** The leopardess audit's rejected claims become a regression
   corpus (the fixloop "bugs that dissolve become graded benchmarks" idea, applied to facts).

## 4. The evidence base as data

Formalize `site_specs` aspect `evidence_base` (per site, `is_current`, `pinned=true`,
`created_by` human or operator agent) as:

```jsonc
{
  "audit_doc": "docs/leopardessconsulting/AUDIT_verified_facts.md",   // provenance pointer
  "facts": [
    {
      "id": "C1-records-verified",
      "claim": "business records verified against Companies House",
      "value": 2767,                      // optional; for number-bearing facts
      "kind": "metric",                   // metric | capability | entity | attestation
      "source": {                         // exactly one of:
        "sql": "SELECT count(*) FROM business_intel.businesses WHERE verified", // live-verifiable
        "artifact": "platform/.../companies_house_actions.go",                  // code/URL evidence
        "attested_by": "owner, 2026-07-09"                                       // human word (e.g. '30 years')
      },
      "verified_at": "2026-07-09",
      "tolerance": "gte"                  // for metrics: exact | gte | approx_pct:5
    }
  ],
  "banned_claims": [                       // the audited-out fabrications — a per-site blacklist
    {"pattern": "eight departments",            "reason": "L0 audit: invented taxonomy"},
    {"pattern": "(70|seventy)\\+? (running )?agents", "reason": "L0 audit: UNSUPPORTED fleet claim"},
    {"pattern": "Awards Won",                   "reason": "turn-15: fabricated stat label"}
  ],
  "allowed_entities": ["Companies House", "New Media Age", "worldsoccernews.com", ...] // proper nouns the copy may assert relationships with
}
```

Notes:
- **Two tiers of facts.** `sql`-sourced facts are *live-verifiable* (and go stale — see
  freshness, §6 V4). `artifact`/`attested_by` facts are checked for *presence in the register*,
  not re-proved.
- **`banned_claims` is the cheapest, highest-yield piece.** It is a regression suite for this
  site's own history — it would have caught points (2) in §1 outright, deterministically.
- **Population for leopardess is a transcription job**, not research: the audit doc's C/U/D/P/X
  rows already contain claim + verdict + evidence. (~30 facts, ~10 banned patterns.)

## 5. Components to build

### V0 — evidence base formalized and populated (no behaviour change)
Schema above agreed; leopardess `evidence_base` rewritten from prose string to structured data;
transcribed from the audit doc. Deliverable: the spec row + a `jsonb_pretty` dump in the notes.

### V1 — deterministic checks (the core of this spec)
**V1a. Build-time, in `validate_page_content`:**
- `checkBannedClaims(html, evidenceBase)` — regex/substring scan of rendered text nodes against
  `banned_claims`. Severity **blocker** (these are *known* falsehoods for this site).
- `checkUnregisteredNumbers(html, evidenceBase)` — extract number-bearing claim candidates from
  text (see landmines §7 for the exclusion classes) and flag numbers asserted as *facts about
  the business* that match no `facts[].value`. Severity **error** (→ `mark_needs_review`),
  never blocker — extraction has false positives, and error already routes to a human.
- Both no-op when the site has no `evidence_base` (opt-in per site, like discovery checks).

**V1b. Post-deploy discovery check `check_unverified_claims`:**
Same two scans over *deployed* page_components (catches drift, hand-edits, and pages that
predate the gate — this is what would have caught "eight departments" sleeping on an orphan
page). Emits `claims_unverified` work items, `severity` from the scan class, routed to
**human review**, one item per page with the offending snippets in the spec (mirroring how
`check_phantom_internal_links` reports). Respect `locked_at` (precedent:
`check_placeholder_contact`).

### V2 — writer whitelist injection (generation-time)
In the page-content-writer prompt template, where research findings are injected today, add an
evidence block when `evidence_base` exists:

> ## Verified facts (the ONLY numbers and named entities you may assert)
> {facts list} — If a fact you want is not here, write the capability without the number, or
> mark it plainly as something we could do. Never approximate a listed number.

This flips "don't invent" (unbounded) to "use only these" (bounded) — the same fix that worked
for emails ("USE ONLY THESE — DO NOT INVENT" + a deterministic backstop).

### V3 — LLM claims auditor (judgement lane)
A `claims-auditor` agent (one LLM call per page, mirroring `content-quality-auditor`'s shape):
extracts *assertions* ("we have done X", "clients Y", "the system handles Z"), classifies each
as `supported | could-framed | unsupported`, citing the `facts[].id` that supports it. Findings
→ `claims_unverified` work items → HITL. This lane catches what regex can't: prose claims with
no number in them (the fake client case studies were this class). Run it from the discovery
agent's `checks` array like any group auditor; never inline in the build path (cost + latency +
false-positive tolerance are all wrong for a gate).

### V4 — live metrics + freshness
- A scheduled task re-runs `sql`-sourced facts and updates `value`/`verified_at`; a
  `stale_evidence` finding when live value drifts outside `tolerance` (e.g. the site says
  2,767 and the DB now says 3,104 — the copy is *underclaiming*; still worth a work item).
- This is the same query layer the planned L7 chart component wants (charts render real DB
  numbers; leopardess PLAN §5) — build once, feed both.

## 6. Integration points (exact)

| Piece | Where |
|---|---|
| Build gate additions | `platform/orchestration/actions/validate_page_content.go` (pattern: `validateEmails` at :571 — reads DB inside the action; severity plumbing at :220–279; failure log via `writeValidationFailureLog`) |
| Discovery check | new `platform/orchestration/actions/discovery_checks/check_unverified_claims.go`; register in `discovery_checks/registry.go`; enable via discovery agent `checks` array |
| Work item + routing | `site_work_items` `item_type='claims_unverified'`, handler → HITL surface (see concept register `hitl` bucket); dedup via `item_key='claims:<page>'` like `page_rerender:<page>` |
| Writer prompt | `page-content-writer` `agent_definitions.default_config` prompt template (the `execute_llm_prompt` step; research-findings injection is the model) |
| Evidence base | `site_specs` aspect `evidence_base` (unique-current index exists: `idx_site_specs_current`); read helper alongside the existing spec readers (`read_site_spec` action) |
| Auditor agent | new `agent_definitions` row `claims-auditor`, category `analyst`, modelled on `content-quality-auditor` |

## 7. Landmines (learned the hard way — read before building)

1. **DOM position matters.** The email validator flagged `placeholder="jane@company.com"` — an
   HTML *attribute example*, not a contact claim — and thereby blocked every build of every page
   using the shared contact-block (fixed 2026-07-14 by changing the fallback, but the checker
   was also wrong: an email in a `placeholder=` attribute is not an assertion). Claim scans
   must parse text nodes, not raw HTML, and must know assertion contexts (body text) from
   non-assertion contexts (attributes, `<code>`, tool UI defaults).
2. **Number false-positive classes** to exclude from `checkUnregisteredNumbers`: dates, years,
   times, prices in interactive tool examples, phone numbers (separately validated), CSS/px
   values that leak into text, ordinals, reading times, version numbers. Start with a
   high-precision extractor (numbers adjacent to business-claim nouns: clients, records,
   users, sites, years of experience, %, "verified", "processed") and widen from measurement.
3. **The audit itself has caveat semantics.** C1 is TRUE but "handles dissolved companies"
   would be FALSE (the matcher filters `company_status='active'`). A fact row supports a
   *specific* claim wording, not a topic. Keep `claim` text specific; the V3 auditor compares
   against claim text, not fact ids alone.
4. **Auditors must not rewrite.** Existing governance: auditor channel = work items;
   direction changes are human-only, `pinned`. `claims_unverified` items terminate at human
   review. The only automated *fix* ever permitted is reverting to a previously-human-approved
   value (and even that: phase-later, opt-in).
5. **Locked components** (`locked_at`) are skipped by precedent — don't flag content a human
   has explicitly pinned.
6. **Cost/scope.** V1 is pure Go + one spec read per page — negligible. V3 is one LLM call per
   page per audit pass; run it at discovery cadence (not per-build), respect the audit-pass
   counter conventions in the improvement loop.
7. **Fleet rollout.** Other sites have no audit doc. Everything is opt-in on `evidence_base`
   presence; `banned_claims` starts empty elsewhere. Do NOT block builds fleet-wide on a layer
   only one site has data for.

## 8. Benchmark & verification (verify by artifact, per house rules)

Build the regression corpus from leopardess history — every one of these previously shipped:

| # | Fabrication | Class | Must be caught by |
|---|---|---|---|
| B1 | "eight departments" | banned claim | V1a blocker + V1b on deployed HTML |
| B2 | "more than 70 agents" | banned claim (audit: UNSUPPORTED) | V1a/V1b |
| B3 | "2,767 Awards Won" | true number, false claim wording | V3 (number is registered; *claim* isn't) — hard case, document if missed |
| B4 | Five named fake client case studies ("Revenue Operations at a Growth-Stage SaaS Company", results "days to minutes") | unsupported prose assertions | V3 |
| B5 | `jane@company.com` in body text | email (existing check) | must STILL be caught post-refactor |
| B6 | `jane@…` in a `placeholder=` attribute | non-assertion context | must NOT be flagged (landmine 1) |
| B7 | "over 150 agent definitions" (156 in DB) | registered, live-verifiable | must pass V1; must go stale-checked by V4 if count drops below 150 |

Acceptance for each phase = run the corpus as fixture HTML through the check and show the
findings table in the notes. End-to-end: re-run V1b against the *live* leopardess site and
expect zero findings (the site is currently clean — that's the baseline).

## 9. Suggested build order

V0 (schema + transcription, no code risk) → V1a banned-claims blocker (smallest code, catches
the worst class) → V1b discovery check → V2 prompt injection → V3 auditor → V4 freshness.
V0+V1 are one sitting each; V3 is the only piece needing prompt design iteration.

## 10. Open questions for the owner

1. Severity of `checkUnregisteredNumbers` at build time: `error` (→ human review, proposed) or
   `warning` (log only) while precision is unproven?
2. Should `banned_claims` be fleet-shareable (some patterns are universal: "Awards Won",
   invented-client shapes) or strictly per-site? Proposal: per-site only until two sites have
   evidence bases.
3. V4 drift: when the live number *exceeds* the published one, is auto-updating the published
   number ever acceptable, or is that also human-gated? (Proposal: human-gated; it changes copy.)
4. Does the HITL surface for `claims_unverified` reuse the existing needs_human_review queue
   or warrant its own view? (Check concept register `hitl` bucket for what exists.)

## 11. Key files & references

- `docs/leopardessconsulting/AUDIT_verified_facts.md` — the evidence source and method
- `docs/leopardessconsulting/RUNNING_NOTES.md` turns 15–18 — every incident cited in §1
- `platform/orchestration/actions/validate_page_content.go` — the gate to extend
- `platform/orchestration/actions/discovery_checks/{registry.go, check_placeholder_contact.go, check_phantom_internal_links.go}` — patterns to copy
- `docs026_concept_register/.buckets/{content-governance.md, content-quality.md, hitl.md}` — governance rules that bind this
- `docs024_key_docs_latest/fixloop_eg_dartsonline/` — the benchmark-from-history convention
