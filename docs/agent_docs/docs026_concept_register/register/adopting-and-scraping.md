# Register — adopting-and-scraping

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 2 raw extractions (1 unique block, appearing twice
due to exact whole-block duplication in the cluster input file) across unit
U22.

### SCR-001 — Polite-scraping throttle (REQUEST_THROTTLE_MS)
- **status:** aspirational
- **status-evidence:** "(Optional) Throttle adapters — if you want the 5s delays between requests, add the throttle code and set REQUEST_THROTTLE_MS=5000 on the webscrape and web-search adapter deployments."
- **what:** An optional per-adapter throttle env var adding fixed delays between outbound web-scrape/web-search requests, to keep bulk vet data collection polite and avoid rate-limit/blocking. Presented as opt-in infra config, not verified as deployed.
- **sources:** docs019_business/005_initial_messaging.md#before-running
- **relations:** area-sweep discovery (business-intel-collection.md BIC-001); vet-practice-verifier
- **verify-later:** REQUEST_THROTTLE_MS handling in webscrape/web-search adapters

### SCR-002 — Fetch-recorded provenance (`datahelpers.ExtractFetchProvenance`)
- **status:** built, inert until the chassis image rolls
- **status-evidence:** committed `2ebabf2ca` 2026-07-28 with tests; `platform/orchestration/datahelpers/fetch_provenance.go`. Not yet exercised by a live run — vet collection has been off since 2026-03-18.
- **what:** A shared reader that pulls `{SourceURL, SourceName, SourceType, CapturedAt}` out of a webscrape result, so a writer records where data came from **from the fetch record** rather than asking the LLM that produced the data. Every webscrape provider result already carries `url` + `captured_at`, set beside the HTTP call; this is the reader nobody had. Returns a `found` bool — an unreadable fetch record is a wiring fault to surface, never permission to store an observation unsourced. Accepts six response shapes because the adapter→coordinator chain has several unwrap points and a single-shape reader FAILS OPEN.
- **why it is callable by others:** any vertical that stores scraped observations (not just business-intel) can call it; the alternative each of them would otherwise write is "ask the model for its source", which is the rejected pattern.
- **sources:** bugs_open/100; docs024_key_docs_latest/bugfix_100_101_scrape_provenance/
- **relations:** `business_intel.data_observations` writer (business-intel-collection.md); the AI-asserted-claim remediation class (bugs_closed/043, bugs_closed/061); SQL 257 (the `NOT VALID` CHECK that makes the unsourced state unrepresentable)

### SCR-003 — Declared config-key contract + unknown-key detection (`ActionInputSpec.ConfigKeys`)
- **status:** built, inert until the chassis image rolls; 1 of 228 actions adopted
- **status-evidence:** committed `2ebabf2ca` 2026-07-28; `platform/orchestration/datahelpers/action_inputs.go` (`UnknownConfigKeys`, `IsStrictConfigAction`, `ListDeclaredConfigKeys`), wired in `platform/validation/workflow.go`. Live audit run reports 1 declared action, 208 undeclared, 726 undeclared (action,key) pairs.
- **what:** An action declares the step-config keys it actually reads; the workflow validator — the only place that sees an action name and its config together on every run — then reports keys the action does **not** read, instead of the runtime silently ignoring them. `StrictConfig` escalates the warning to a hard validation refusal once a contract is known complete. Opt-in per action by design: 811 distinct (action,key) pairs across 228 live actions make a fleet-wide allow-list a guess at scale, and an over-strict validator is a worse defect than the inert key it chases. `UnknownConfigKeys` returns a `checked` bool so "declared and clean" can never be read as "never examined". Extends the pre-existing spec registry (134 registrants, previously read by nothing but a parity test) rather than adding a second one.
- **why it is callable by others:** any action author can opt in with one field and get typo/staleness detection for their step config; this is the general answer to the recurring "config that lies" class (bugs_closed/042, bugs_closed/127, bugs_open/101).
- **sources:** bugs_open/101 fix candidate 1; docs024_key_docs_latest/bugfix_100_101_scrape_provenance/
- **relations:** `registry_parity_test.go`; 016b §9 *"A registry that everything registers with and nothing reads"*

### SCR-004 — Config-key coverage report (`scripts/audit-config-keys.sh`, `cmd/config-key-audit`)
- **status:** built and exercised (read-only; needs cluster access + `go run`)
- **status-evidence:** run against the live fleet 2026-07-28; found a fifth inert key (`add_protocol` on `domain-research-classifier/scrape_site`) that `bugs_open/101` had not identified.
- **what:** Offline half of SCR-003. Asks the **binary** what each action declares (the declarations are Go, registered by `init()`; a source grep would quietly disagree with the running code) and joins that against every live `agent_definitions` step config. Reports **UNKNOWN KEYS** (action declared its contract, key is not in it — a real inert key) separately from **UNDECLARED ACTIONS** (not opted in, so nothing is known and nothing there is evidence of a bug), because the fix differs and conflating them makes the report unactionable. `--json` for the full machine-readable list; the text listing's cap is labelled a display limit, not a filter.
- **why it is callable by others:** it is the adoption ratchet for SCR-003 and it answers "is this config key real?" for any action, for anyone, without reading Go.
- **sources:** docs024_key_docs_latest/bugfix_100_101_scrape_provenance/RUNBOOK_scrape_config_and_provenance.md
- **relations:** the `098` unreviewed-commits report (same coverage-ratchet shape); `102_coverage_ratchet.txt`
