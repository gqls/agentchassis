# Geographic Area Sweep Discovery System

## What It Is

A two-agent system that systematically searches every UK postcode district for veterinary practices we don't already have in our database. We currently have around 3,500 practices and estimate there are roughly 5,000 in the UK. This system aims to find the missing ~1,500.

The UK is divided into 3,402 postcode districts (e.g. BT4 for part of Belfast, B29 for part of Birmingham). The system searches each district in turn, checks results against our existing data, and saves any new discoveries as candidates for later review and promotion.


## How It Works

You send one Kafka message. Two agents handle everything else.

```
You (shell script)
  │
  │  "orchestrate area-sweep-orchestrator, limit=50"
  ▼
┌──────────────────────────────┐
│  area-sweep-orchestrator     │
│                              │
│  1. load_unswept_areas       │  ← Queries search_areas table for 50
│     (from DB)                │    districts that haven't been swept yet
│                              │
│  2. dispatch_area_discoverers│  ← Sends 50 Kafka messages, one per
│     (via Producer)           │    district, to generic.requests
└──────────┬───────────────────┘
           │
           │  50 messages, each with a district_code
           ▼
┌──────────────────────────────┐
│  area-sweep-discoverer (×50) │  One instance per district
│                              │
│  1. web_search               │  ← "veterinary practice BT4 Belfast UK"
│     (10 results)             │    Uses firecrawl search adapter
│                              │
│  2. process_area_sweep       │  ← For each result:
│     (Go action)              │    - Skip directories (google, yell, rcvs)
│                              │    - Skip non-vet results
│                              │    - Check against businesses table
│                              │    - Check against existing candidates
│                              │    - Insert new → discovery_candidates
│                              │    - Update search_areas tracking
└──────────────────────────────┘
```


## Starting a Sweep

```bash
./start_area_sweep.sh            # sweep 50 districts (default)
./start_area_sweep.sh 10         # sweep 10 districts
./start_area_sweep.sh 100 BT     # sweep 100 districts in Belfast area only
```

The script sends a single message to `system.agent.generic.requests` with action `orchestrate` and agent type `area-sweep-orchestrator`. That's all it does.


## Database Tables

### search_areas (new)

Tracks every UK postcode district and its sweep status.

| Column | Purpose |
|--------|---------|
| district_code | e.g. "BT4", "B29" — unique per country |
| area_code | e.g. "BT", "B" — for filtering by region |
| area_name | e.g. "Belfast", "Birmingham" |
| sweep_count | How many times this district has been swept |
| last_swept_at | When it was last swept |
| last_result_count | How many search results came back |
| candidates_found | Running total of new candidates found |

Seeded with 3,402 UK postcode districts covering Aberdeen (AB) through Lerwick (ZE).

### discovery_candidates (existing)

Where new finds are stored. Each candidate has a status: `pending` → `matched` / `promoted` / `dismissed`. The process_area_sweep action inserts here with `ON CONFLICT` handling so duplicates are absorbed.


## The Three Go Actions

All three use `ActionInputSpec` and `ExtractActionInputs()` per the agent checklist.

### load_unswept_areas

Queries `search_areas` for districts ordered by sweep_count ascending (never-swept first), with optional filtering by area_code. Returns the list of areas plus summary stats (total, never_swept, candidates_found).

Inputs: `limit` (default 50), `country` (default "GB"), `area_code` (optional, e.g. "BT").

### dispatch_area_discoverers

Reads the areas list from collected_data (output of load_unswept_areas). For each area, it builds a Kafka message body with `{"action":"orchestrate","config":{"agent_type":"area-sweep-discoverer"},"input_data":{...}}` and produces it to `system.agent.generic.requests` using `params.Producer`. Adds 100ms delay between messages.

### process_area_sweep

The workhorse. For each search result it: skips directories and aggregator domains (reuses the `skipDomains` map from scan_discovery_candidates.go), filters non-vet results, checks the URL against `businesses.website_url`, checks against existing `discovery_candidates.source_url`, extracts a practice name from the title and a postcode from the snippet, and inserts new candidates. Finally updates the `search_areas` row with sweep tracking.


## Action Registration

Add to `action_registry.go`:

```go
"load_unswept_areas":         LoadUnsweptAreasAction,
"dispatch_area_discoverers":  DispatchAreaDiscoverersAction,
"process_area_sweep":         ProcessAreaSweepAction,
```


## Cost

Each district search uses 1 firecrawl credit. A full UK sweep costs 3,402 credits out of our 100k/month budget — roughly 3.4%. The sweep can be repeated monthly to catch new practices.


## Deployment Order

1. Run `migration_search_areas_with_seed.sql` — creates table and inserts 3,402 districts
2. Run `create_area_sweep_agents.sql` — creates both agent definitions
3. Add the three Go actions and register them
4. Build and deploy the chassis
5. Test: `./start_area_sweep.sh 5 BT` (5 Belfast districts)
6. Check results: `SELECT * FROM business_intel.discovery_candidates ORDER BY created_at DESC LIMIT 20;`
7. Scale up: `./start_area_sweep.sh 100`


## Files

| File | Purpose |
|------|---------|
| `migration_search_areas_with_seed.sql` | Schema + 3,402 district seed rows |
| `create_area_sweep_agents.sql` | Both agent definitions |
| `load_unswept_areas.go` | Load action for orchestrator |
| `dispatch_area_discoverers.go` | Dispatch action for orchestrator |
| `process_area_sweep.go` | Processing action for discoverer |
| `start_area_sweep.sh` | One-liner launcher script |