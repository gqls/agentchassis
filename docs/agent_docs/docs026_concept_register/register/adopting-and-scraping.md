# Register — adopting-and-scraping

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
