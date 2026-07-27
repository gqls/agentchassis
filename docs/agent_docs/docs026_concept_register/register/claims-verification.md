# Register — claims-verification

12 concepts. **NOT from the 2026-07-13 extraction** — this whole subsystem shipped
after extraction froze (its first plan is dated 2026-07-16), so none of it was
ever in the register. Added 2026-07-27 from the oufe.com workstream, grounded in
live code and DB reads made that day; every citation below was read first-hand,
not carried from another document.

This is the **third** recorded instance of the same gap (after FIX-051/052/053 and
MDL-038/039), and by far the largest — an entire honesty subsystem, V0 through V5.
See `bugs_open/106` for why it keeps happening.

Status vocabulary per `README.md`. Where a thing is *built but never exercised*
that is said explicitly, because "deployed" would overstate it.

### CLM-001 — `evidence_base` is a `site_specs` aspect, not a table
- **status:** deployed
- **what:** The per-site fact register lives at `site_specs.aspect='evidence_base'`, one current row per site under a partial unique index (`site_id, aspect) WHERE is_current`. So it is **never UPDATEd in place** — every change is supersede-then-insert in one transaction, carrying `pinned` forward or the row silently loses human-owned status. Shape (`datahelpers/claims.go:97-103`): `audit_doc`, `governing_rule`, `facts[]`, `banned_claims[]`, `allowed_entities[]`, plus a `writer_block` string the parsed struct does not model.
- **sources:** `platform/orchestration/datahelpers/claims.go:97-130`; `docs024_key_docs_latest/claims_verification/RUNBOOK_claims_verification.md` §1-2
- **relations:** CLM-002, CLM-009, CLM-010; site-spec-and-classifier register
- **verify-later:** the `pinned` carry-forward on every supersede path

### CLM-002 — `ScanBannedClaims` is an unrestricted case-insensitive regex over prose
- **status:** deployed
- **status-evidence:** function body read 2026-07-27 at `claims.go:284-325` — blocks × patterns, `FindStringIndex`, nothing else.
- **what:** The banned-pattern scanner has **no numeric gating whatsoever**. It matches whatever patterns a site is given, about anyone, numeric or qualitative. Live registers already carry purely qualitative patterns (`leaderboard`, `live now`, `price target`, `years of experience`). Patterns compile as `(?i)`; an invalid regex degrades to a literal substring, so a typo never silently drops a ban. It scans *assertion blocks* extracted from HTML (`extractAssertions`, `claims.go:165-226`), which skips `script/style/head/code/pre` and cannot match across block boundaries.
- **why it matters:** this concept exists because its absence caused a live error. On 2026-07-26 a session concluded that qualitative claims were "invisible to every scanner", and wrote that into a live council seat, having read only the *numeric* scanner beside it. **The capability was never missing.** See CLM-003 for the limitation that was mistaken for a general one.
- **sources:** `claims.go:284-325`, `:121-128`, `:165-226`; `sql_for_agents/226`, `227`; `docs024_key_docs_latest/WRONG_CALLS.md` 2026-07-26
- **relations:** CLM-003 (the confusable neighbour), CLM-010, CLM-011

### CLM-003 — `ScanUnregisteredNumbers` is gated by an English business-noun window
- **status:** deployed, with a documented blind spot
- **what:** The unregistered-number scan inspects a number only when `businessClaimContextRe` (`claims.go:339`) matches its surrounding window — clients, customers, records, awards, uptime, years of experience and similar — and `isExcludedNumber` (`:408-494`) then drops years, dates, versions, ratios, measurements and **currency amounts**. Consequence: it is near-inert on non-English copy (recorded for relojistas.com) and on finance prose, where "£16bn of Class A debt" is never scanned at all. **Both gates apply to this function only, never to CLM-002.**
- **status-evidence:** call-site census 2026-07-27: `businessClaimContextRe` and `isExcludedNumber` are referenced only within `ScanUnregisteredNumbers` and the stat lane; zero references inside `ScanBannedClaims`.
- **sources:** `claims.go:329-349`, `:363-372`, `:408-494`; `traffic_probe/relojistas_evidence_base.sql` limitation header
- **relations:** CLM-002 — **do not generalise this entry to the sibling scanner; that is the exact error already made once**

### CLM-004 — Two enforcement surfaces: build gate (blocker) and post-deploy sweep (high)
- **status:** deployed; the post-deploy half has no cadence
- **what:** V1a runs inside `validate_page_content` check 8 over draft HTML before save — a banned-claim hit is severity **blocker** and fails the page build; an unregistered number is **error**, deliberately never a blocker. V1b (`discovery_checks/check_unverified_claims.go`) scans **stored `rendered_html` of live pages and site chrome**, skipping locked components, raising `claims_unverified` at **high**, terminating at `needs_human_review` with no handler by design. Both are opt-in by the mere presence of an `evidence_base` row: `loadEvidenceBase` returns nil and every lane silently no-ops.
- **the gap:** V1b is reachable only via `quality-discovery-agent` ← `improvement-loop` ← the `improvement-sweep` scheduled task, **disabled since 2026-05-02**. So the estate's only automatic detector of published-content drift effectively never runs.
- **sources:** `validate_page_content.go:274-282`, `:926-958`; `check_unverified_claims.go:135-150`, `:365-389`; `bugs_open/083` (detected-findings slug)
- **relations:** CLM-010; work-item-integrity register; improvement-loop register

### CLM-005 — V2: the writer whitelist bounds invention instead of forbidding it
- **status:** deployed
- **what:** When a site's register carries a `writer_block`, the page-content-writer prompt renders "## Verified Facts (the ONLY numbers and named entities you may assert…)", which **overrides** the writer's unbounded "never invent" rule with a bounded "state only these". A fact without a `writer_line` is omitted from a composed block entirely — never auto-phrased.
- **landmine:** a **`writer_block`-only register switches both deterministic checkers off silently**, because `ParseEvidenceBase` returns nil when `facts[]` and `banned_claims[]` are both empty.
- **sources:** `claims.go:110-130`; `refresh_evidence_base_action.go:485-489`; `robot-hands` 043 workstream notes
- **relations:** CLM-001, CLM-004

### CLM-006 — V3 `claims-auditor`: an LLM prose auditor, live, unscheduled, never demonstrated
- **status:** partial
- **status-evidence:** "LIVE 2026-07-18; not on a cadence (owner call)"; "the auditor's catch path isn't yet demonstrated — I wasn't going to plant fabrications on a live site to test it".
- **what:** Classifies every prose assertion as supported / could-be-framed / unsupported against the register, citing the supporting fact id — deliberately the lane that "catches what regex can't: prose claims with no number in them". Reports at most 12, worst first. Findings go to humans; nothing is auto-fixed.
- **provenance warning:** **this agent exists only as a DB row.** No seed file created it and none exists in git history; `agent_definitions.default_config` is the sole source of truth for its prompt. Any account of its behaviour in docs is second-hand.
- **sources:** `claims_verification/NOTES_claims_verification.md:266-279`; `SPEC_claims_verification.md:167-175`; `PLAN_2026-07-16:140-141`
- **relations:** CLM-011 — it is the lane most likely to catch overclaims about ourselves, and the least exercised

### CLM-007 — V4 freshness: the one part of the layer that sweeps the fleet daily
- **status:** deployed
- **what:** `refresh_evidence_base` re-runs sql-sourced facts, updates values, raises `stale_evidence` on drift **including under-claiming**, and writes back with compare-and-swap so a human edit is never lost. Driven by the `evidence-freshness` scheduled task, enabled, daily, and when given no `site_id` it selects **every** site with a current register.
- **why it matters here:** this is the only claims-layer component with a live fleet cadence, which makes the layer *look* actively swept when only its fact-refresh half is (see CLM-004).
- **sources:** `refresh_evidence_base_action.go:172-199`, `:341-349`, `:680`; `SEED_evidence_freshness_scheduled_task.sql:45-48`

### CLM-008 — V5: citations survive only if the quote is still literally there
- **status:** deployed; first end-to-end completion 2026-07-27
- **what:** `verify_and_register_citations` re-fetches every cited URL and registers a claim only when its **verbatim quote** is present. The model proposes, the fetcher disposes. Three outcomes are held apart deliberately: quote found → stands; 200 but quote absent → `citation_lost` ("the published claim now rests on nothing"); fetch failed / 403 / PDF → `fetch_error`, **UNKNOWN, never drift** — "a paywall going up is not evidence the fact is wrong". Quote matching is forgiving on presentation (entities, curly quotes, nbsp, thousands separators) and strict on content.
- **status-evidence:** activated on v1.0.1140 (2026-07-20) but its smoke run failed on `bugs_open/047` and was never repeated; **first successful completion 2026-07-27**, 19 citations registered for oufe.com, most from `legislation.gov.uk`.
- **sources:** `platform/orchestration/actions/evidence_citations.go:9-21`; `datahelpers/citations.go:35-49`; `docs024_key_docs_latest/oufe/NOTES_oufe.md`

### CLM-009 — `EvidenceFact.Kind` is declared, documented, and read nowhere
- **status:** unknown → **confirmed dead 2026-07-27**
- **what:** The struct declares `Kind string // metric | capability | entity | attestation` (`claims.go:73`) and the spec repeats the vocabulary, but no code anywhere reads it — every `.Kind` consumer in the tree belongs to a different struct. A fact declared `capability` behaves identically to one declared `metric` or misspelled. `EvidenceSource` beside it (`:60-64`) already models exactly one of `sql | artifact | attested_by`, and CLM-007 already re-checks the sql-backed ones on a fleet cadence.
- **why it matters:** that is the shape of a **promise register** — a claim, its kind, the mechanism behind it, when last checked, and a sweep that re-checks it — sitting unused. Its invisibility (a field that governs nothing has no call sites) is why a session nearly proposed building a separate promise register in 2026-07.
- **sources:** `claims.go:60-87`; `SPEC_claims_verification.md:104,124-126`; `bugs_open/105`
- **relations:** hitl register; the EXPERIENCE_PLAN promise ledger (`sql_for_agents/167:161`)

### CLM-010 — `banned_claims` is per-site only, and the fleet-share decision is a lapsed deferral
- **status:** convention (a decision, not an artifact) — **precondition expired**
- **what:** Every reader of the register is keyed on `site_id`; there is no global row, no inheritance, no merge of a base set. `SPEC_claims_verification.md:250-252`, under *"Open questions for the owner"*, asks whether `banned_claims` should be fleet-shareable "(some patterns are universal)" and proposes **"per-site only until two sites have evidence bases"**. Correct at n=1. Measured 2026-07-27: 8 sites have registers and **only 5 of 15 live sites carry a single pattern** — the ten without include vetcomparison.uk (the fabricated-prices site) and idea.uk (the only one taking money).
- **the transferable point:** **a deferral with a numeric trigger and no watcher becomes permanent policy.** Nothing re-read it; the precondition lapsed silently.
- **precedent for the fix:** `datahelpers/voicetells.go` already solves this shape one file away — `globalTellPhrases()` (`:121-137`) unioned with the per-site list at `:109`.
- **sources:** `SPEC_claims_verification.md:250-252`; `PLAN_2026-07-16:138-139`; `bugs_open/104`

### CLM-011 — The overclaimed-reliability class, and the line that separates it
- **status:** deployed (2026-07-26/27)
- **what:** A promise about **our own** accuracy is a claim like any other and had never had a pattern written for it on any site. Controls now: a fleet-wide writer rule on both content writers ("never promise accuracy you cannot guarantee"), an `OVERCLAIMED RELIABILITY` contract on the compliance council seat with judge clauses, and tested banned patterns on oufe. The governing line, which generalises: **a site may describe what it DOES; it may not claim what that GUARANTEES.** "We cite every figure and date it" is a keepable process commitment; "a claim without a source does not appear here" asserts a completeness nobody can verify.
- **origin:** oufe.com shipped four such promises live, hours after the owner struck a weaker version of the same slogan.
- **sources:** `sql_for_agents/223`, `226`, `227`; `docs024_key_docs_latest/oufe/`
- **relations:** CLM-002 (the scanner that always could have caught it), CLM-006

### CLM-012 — `grounded-explainer`: a high-attention content lane that cannot publish
- **status:** deployed 2026-07-26, exercised 2026-07-27
- **what:** For pages whose facts must not come from model memory (law, regulation, safety, clinical, tax): search → extract atomic claims with verbatim quotes → CLM-008 verification → compose from survivors only → an **independent** grounding audit listing every untraceable sentence → terminate at `needs_human_review`. **There is no config flag that makes it publish** — the one property that does not depend on a model behaving well. Reuses the V5 acquisition chain wholesale.
- **status-evidence:** first full run 2026-07-27 produced a 7,546-char draft citing 3 sources with 3 declared gaps; the audit returned `needs_revision` naming two plausible legal generalisations the draft could not support, and correctly declined to flag its closing disclaimer.
- **landmine found in its own first runs:** the verification action returns a **receipt of citation ids, not the claims** — a composer handed that receipt correctly refused to assert anything and reported honest gaps. **The failure was safe**, which is the design point.
- **sources:** `sql_for_agents/224`, `225`; `docs024_key_docs_latest/oufe/NOTES_oufe.md`
