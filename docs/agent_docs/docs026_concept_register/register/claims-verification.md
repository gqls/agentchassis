# Register — claims-verification

> **covers-through: 2026-07-28** · written 2026-07-27 from first-hand code/DB reads, never part of the extraction.
> Everything else dates from the 2026-07-13 extraction freeze — absence
> here is not evidence of absence in the platform. See `bugs_open/106`.

15 concepts. **NOT from the 2026-07-13 extraction** — this whole subsystem shipped
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
- **UPDATED 2026-07-28 (CLM-015):** "both are opt-in by the mere presence of an `evidence_base` row" is **no longer true of the banned-claim half**. It now runs on every site, register or not, on both surfaces. The unregistered-number half is unchanged and still opt-in, so the two halves of check 8 have **different** opt-in rules by design — do not simplify that back.
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
- **UPDATED 2026-07-28 — the figure above has drifted and the fix now has a measured cost.** Re-run of the same query: **7 of 15** carry patterns (was 5; ai-agent-orchestration gained 10 and finetuning 3 by hand), **9** current `evidence_base` rows. vetcomparison.uk and idea.uk are **still at zero**. And the count conflates two states that decide the fix's reach: **6 sites have no row at all**, while **2 have a row with 0 patterns but non-empty `facts[]`** (robot-hands 5, gamesdesign 4) — those two parse **non-nil**, so the "union a global set into `ParseEvidenceBase`" fix reaches **9 of 15** sites, not 7.
- **the landmine, measured before the fix shipped:** dry-running the tested oufe set (10 patterns) over the stored `rendered_html` of all 15 sites (908 components, via CLM-014) produced **7 findings on 3 sites, of which 4 are FALSE POSITIVES** — every one firing on a *negated* sentence ("has **not** been independently verified", "**cannot** be independently verified", "**not** independently verified"). At severity `blocker` that fails a page build for making the honest disclosure this layer exists to encourage. One pattern (`(fully|independently|externally|properly) (verified|audited|fact.?checked)`) causes 6 of 7 hits and all 4 false positives; with it dropped the other 9 fire **exactly once** fleet-wide, on a true positive. **CLM-002's "no gating whatsoever" is safe only because each pattern was human-audited for one site — that review IS the false-positive apparatus, and going fleet-wide removes it while keeping blocker severity.** Nothing is biting today: every armed site scores 0 against its own register. No negation-guard prior art exists in the estate, and RE2 has no lookbehind, so a guard must be code.
- **sources:** `SPEC_claims_verification.md:250-252`; `PLAN_2026-07-16:138-139`; `bugs_open/104` § "Dry run 2026-07-28"; `docs024_key_docs_latest/bugfix_104_fleetwide_claim_patterns/`
- **relations:** CLM-002 (the premise that the audit was the guard), CLM-011 (the pattern family proposed for sharing), CLM-014 (how it was measured)

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

### CLM-013 — series facts: many dated observations, each independently sourced
- **status:** deployed in `v1.0.1185` (pod-verified 2026-07-28) — **no live site holds a series fact yet**
- **status-evidence:** three distinctive string literals from `claims_series.go` found in the running binary, with a positive control. Added 2026-07-28, after this register file was first written.
- **what:** A fact could hold one `Value` and three dates — `accessed`, `published`, `verified_at` — **all provenance**, none of them the date the value *applies to*. `Observation{as_of, value, source, verified_at}` adds that, under two rules: **every observation carries its own source and never inherits the parent fact's**, and `as_of` is distinct from `verified_at`. `ValidateSeries` fails closed (≥2 observations, valid and unique `as_of`, a resolvable source per point); `numberSupported` gained a series branch, matching exactly even when the fact carries a `gte` tolerance.
- **why it matters:** without a series shape the only options were one fact per year (no series identity, so nothing can plot it and nothing knows a point is missing) or numbers embedded in a claim string, which is invisible to every scanner — the `bugs_open/043` failure mode. And the never-inherit rule is what stops interpolation entering as data.
- **the lesson attached to it:** round 1 of the council review found that `ValidateSeries` enforced the source rule while `numberSupported` never called it, so an unsourced observation still registered its value. **A rule enforced only in a validator is not enforced** — it must hold at the gate that decides. Both now share `observationHasResolvableSource`.
- **sources:** `platform/orchestration/datahelpers/claims_series.go`; `claims.go` (`Observations`, `numberSupported`); council correlation `da40ddf0` round 1

### CLM-015 — the fleet-wide banned-claim set: nine patterns no site may assert about itself
- **status:** committed 2026-07-28 — **in code, therefore INERT until the next chassis image is built and rolled.** Council review pending at commit time (advisory).
- **status-evidence:** verified before commit against the real corpus, not in the abstract — the shipped set over the stored `rendered_html` of all **15 live sites / 908 components** with **no register supplied**, i.e. exactly as an unarmed site is scanned: **0 findings**. Positive control the same run: **6 of 6** overclaim shapes still blocked with no register at all, and the four live sentences that the *candidate* set falsely flagged all pass. 8 unit tests in `claims_global_test.go`.
- **what:** `globalBannedClaims()` in `datahelpers/claims_global.go` — the claims-layer answer to CLM-010. `ScanAllBannedClaims(blocks, eb)` scans the fleet-wide set **plus** the site's own, deduping a pattern present in both, and is **nil-safe**: a site with no `evidence_base` row is still protected, which is the point — otherwise site sixteen is born unarmed exactly as site fifteen was. Wired at both enforcement surfaces (CLM-004) so they cannot drift.
- **the two design constraints that shaped it, both of which rule out the obvious mirror:**
  1. **It is NOT unioned into `EvidenceBase` at parse time**, though `voicetells.go` does exactly that for the voice layer. `EvidenceBase` is marshalled **back** to `site_specs` by `refresh_evidence_base_action.go` and `evidence_citations.go`, so seeding `eb.BannedClaims` would silently persist the fleet-wide set into every site's stored register through write paths that never intended to touch it — the trap `claims.go` already documents for `EvidenceFact.Kind`. The set is held outside any parsed register and joined only at scan time; `globalEvidence` is unexported so it cannot reach a writer.
  2. **`ParseEvidenceBase`'s nil contract is unchanged.** That nil is a load-bearing opt-in signal read deliberately by several lanes. Only the banned half goes fleet-wide; **the numeric scan stays strictly opt-in**, because its false-positive rate is why it is never a blocker.
- **the pattern that is deliberately ABSENT, and the landmine:** `(fully|independently|externally|properly) (verified|audited|fact.?checked)` is **excluded**. It is the strongest of the ten — it catches the shape oufe actually shipped — and it caused **4 of 7** findings in the fleet dry run, every one on a *negated* sentence ("has **not** been independently verified"). **Do not re-add it without a code-level negation guard**; RE2 has no lookbehind, so it cannot be done in the pattern, and there is no negation-guard prior art in the estate. `claims_global_test.go` holds the four real sentences as regression fixtures, so re-adding it fails the suite rather than a live build.
- **the other landmine:** pattern 2 (`(does not|doesn't|do not|don't) appear here`) is **itself a negative construction**, so "prices do not appear here because they change daily" would match. Zero hits fleet-wide today; it is the most likely source of the next false positive. Flagged in the code beside it.
- **the rule for changing it:** dry-run fleet-wide first (`claimscan -no-global`), and **put the NEGATION of every pattern in the pass-list.** The original set was tested 10-for-10 on fabrications and 13-for-13 on legitimate sentences and still shipped this class, because no test sentence contradicted a banned phrase.
- **sources:** `platform/orchestration/datahelpers/claims_global.go`, `claims_global_test.go`; `validate_page_content.go` check 8; `discovery_checks/check_unverified_claims.go` `scanComponentClaims`; `bugs_open/104` § "Dry run 2026-07-28"; `docs024_key_docs_latest/bugfix_104_fleetwide_claim_patterns/`
- **relations:** CLM-010 (the decision it discharges), CLM-002 (whose "no gating whatsoever" was safe only because a human audited per site), CLM-004 (both surfaces), CLM-011 (the class), CLM-014 (how it was measured); `voicetells.go` `globalTellPhrases()` is the shape it mirrors but could not copy

### CLM-014 — `cmd/claimscan`: run the live gate's own engine over exported page HTML, offline
- **status:** deployed (committed 2026-07-16), exercised fleet-wide 2026-07-28
- **status-evidence:** `go build ./cmd/claimscan` clean against the shared tree 2026-07-28; used that day for a 15-site / 908-component dry run whose positive control blocked 6 of 6 overclaim shapes and passed 3 of 3 legitimate sentences.
- **what:** An operator CLI that calls **the same shared scan functions as the deploy gate and the post-deploy sweep** (`ParseEvidenceBase` → `ExtractAssertionText` → `ScanBannedClaims` / `ScanUnregisteredNumbers`), against a TSV export of component HTML plus an `evidence_base` JSON file. So you can ask "what would this pattern set flag, on real copy, across any number of sites" **without deploying, migrating, or touching a site**. Prints `BANNED` / `NUMBER` lines and exits 1 when findings exist, so it works as a scripted acceptance gate.
- **why it earns an entry:** this is the tool that answers "measure the blast radius before you submit" for anything touching the claims layer, and it is the only way to test a candidate pattern set against copy other than the site it was written for. **A session on 2026-07-28 was about to build a second one** — it was found only by reading `sql_for_agents/226`'s verify footer to the end, not by grep. That is `bugs_open/106`'s failure mode landing on a tool that already existed.
- **landmines:** (1) the CLI prints the prefixes `BANNED`/`NUMBER`; `banned_claim` is the JSON `check` value and appears nowhere in the output, so grepping for it returns 0 on every site — a false all-clear. (2) An evidence file with `banned_claims` but no `facts[]` makes **every number** an unregistered-number finding; filter to `^BANNED`. (3) Some site copy is non-UTF-8, and plain `grep -c` returns **empty with no error** on those outputs — use `LC_ALL=C grep -ac`. (4) The documented export query uses `kubectl exec -i`, which **eats a `while read` loop's stdin** and silently scans only the first site.
- **sources:** `cmd/claimscan/main.go` (usage block carries the export SQL); `sql_for_agents/226` verify footer; `docs024_key_docs_latest/bugfix_104_fleetwide_claim_patterns/RUNBOOK_fleetwide_claim_patterns.md`
- **relations:** CLM-002, CLM-003 (the engines it calls), CLM-004 (the two surfaces it reproduces), CLM-010 (the decision it costed); `cmd/voicescan` is the sibling for the voice layer
- **relations:** CLM-001, CLM-003; VIZ-002, VIZ-003, VIZ-004 (visualisation-and-charts)
- **verify-later:** first live series on a real site; whether `ParseCitation` per observation per scan is acceptable at fleet scale

### CLM-016 — `ClaimSurface`: the page's structural type gates the prose number heuristic
- **status:** committed 2026-07-28 (`3ddb4ed2d`) — **in code, therefore INERT until the next chassis image is built and rolled.** Council review submitted before commit (advisory), correlation `de4a19f5`.
- **status-evidence:** measured with CLM-014 against **each of the nine opted-in sites' own live registers** over live `rendered_html`, before and after, over an identical export: **124 unregistered-number findings → 63**. The 61 suppressed are exactly the editorial-page ones (blog-post 46, tool 7, game 4, section-index 2, news-index 1, guide 1) and **`comm` reports zero findings newly appearing**. All 61 were read individually and all 61 are false positives.
- **what:** `datahelpers.ClaimSurface{PageType}` + `ProseNumbersAreClaims()`. `ScanUnregisteredNumbers` now **requires** a surface, and returns nothing on an editorial page type (`guide`, `blog-post`, `blog-index`, `news-index`, `section-index`, `tool`, `game`). `businessClaimContextRe` is a **lexical** gate — it asks whether the words near a number sound like business — and on teaching content that question cannot be answered, because an explainer's worked example is lexically identical to a sales claim ("10,000 active players farming that item"). `pages.page_type` is the structural signal that separates them, and it was already in the row both call sites read.
- **the boundary, which is the part worth copying:** **only the heuristic is surface-gated.** `ScanBannedClaims` runs on every page type — a human-authored pattern for a known falsehood has no false-positive problem to protect against, and this layer's motivating catch (CLM-004, first live run 2026-07-16) was a banned claim **on a guide**. `ScanStatClaims` likewise, per its own structural-position argument. **A precision fix belongs to the mechanism whose precision is the problem, not to the surface it was noticed on.**
- **why a required parameter and not an optional variant:** a `...OnPage` sibling that defaults to the old behaviour lets the next caller silently inherit the blind scan — `bugs_open/093`'s shape exactly (one guarded call site, one nobody remembered). Requiring it makes the compiler visit every caller; it found three production call sites and four in tests.
- **the fail direction, stated deliberately:** the **zero value is UNKNOWN and is SCANNED**. Site chrome belongs to no page; a component reviewer holds no page record; a page may not exist yet. A scanner that has gone quiet and one that is broken look identical from outside, so an absent or unrecognised page type stays noisy. A page type nobody has invented yet is scanned too.
- **landmines:** (1) **The gate only knows the page type if the workflow loaded it.** `page-build-handler` and `tool-recreation-handler` run `load_page_record` (which selects `page_type`) before `validate_content`; `content-reviewer` does not, so it resolves UNKNOWN and behaves exactly as before. A fix like this is inert wherever the signal does not arrive — check the live `agent_definitions` workflow, do not assume. (2) `report` is deliberately **not** editorial although it carries 14 false positives: those are model numbers inside product names ("Schunk EGP 40-N-S-B — manufacturer specification") tripping on `verified`, a different mechanism that would have been fixed by coincidence.
- **sources:** `platform/orchestration/datahelpers/claims.go` (`ClaimSurface`, `editorialPageTypes`, `ScanUnregisteredNumbers`), `claims_surface_test.go`; `validate_page_content.go` (`resolvePageType`); `discovery_checks/check_unverified_claims.go` (`scanComponentClaims`); `cmd/claimscan/main.go` (4th TSV field); `bugs_open/102`; `docs024_key_docs_latest/bugfix_102_page_type_claims/`
- **relations:** CLM-003 (the scan it gates), CLM-002 / CLM-015 (the banned half, deliberately NOT gated), CLM-014 (how it was measured, both directions), CLM-004 (the two surfaces); `bugs_open/093` (the one-guarded-call-site shape the signature change avoids)
- **verify-later:** pod-grep after the next roll (marker + negative control in the workstream RUNBOOK § 2); whether any site files marketing copy under an editorial `page_type`, which would go prose-unscanned
