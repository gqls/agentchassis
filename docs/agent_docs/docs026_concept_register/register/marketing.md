# Register — marketing

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 2 raw extractions across unit U01. (The cluster input file contained this category's raw block twice, back-to-back and byte-identical; merged into one entry below.)

### MKT-001 — Marketing as work items + OpenClaw adapter
- **status:** aspirational
- **status-evidence:** P1 marketing section entirely future (agents/adapter unbuilt).
- **what:** SEM campaigns, landing pages, email sequences, social content, schema markup, ad copy all as work items with dedicated handler agents; an openclaw-adapter (adapter service, self-hosted) translates structured campaign specs to external platforms (Google/Meta/LinkedIn) and returns metrics; marketing-discovery-agent finds gaps (GBP, schema, page-2 rankings, competitor ads); SEM setup is HITL-gated.
- **sources:** P1#Marketing: SEM, Outbound, and Growth
- **relations:** work-item system extensibility; SEO-001 SEO content agent
- **verify-later:** none built
