# Continuation: Handler Testing & Asset Deployer

## What was completed in this chat

### Document consolidation (DONE)
20 overlapping architecture/development documents were consolidated into 3:
- `001_development_guide.md` — daily reference (checklist, API, messages, debugging)
- `002_system_architecture.md` — big picture (design system, agents, pipelines)
- `003_contracts_and_standards.md` — rules (naming, CSS, schemas, boundaries)
- `000_consolidation_map.md` — tracks where each old doc went

These were delivered as files. They should be added to project knowledge and the old documents removed per the consolidation map.

### Handler testing (IN PROGRESS)
We were working through `test_handlers.sh` — a bash script for testing individual handler agents via CLI/Kafka messages, independent of the full orchestrator.

---

## Where we stopped: asset-deployer agent

### The problem
The `test_handlers.sh` script passes `s3_uri` directly in the message body:
```json
{"domain": "...", "s3_uri": "s3://...", "purpose": "hero"}
```

The user's objection: **the initial message shouldn't include the S3 URI.** The assets are already stored in the `assets` table (put there by `store_asset` during pageflow-builder's run). Sending URLs manually is a crutch — in production, front-end agents or orchestrators won't know S3 URIs. The agent should be self-sufficient: receive a `domain` (or `site_id`), query the DB for undeployed assets, and deploy them.

### What exists

**`asset-deployer` agent definition** — EXISTS in the live database (was created in a previous chat session). It is NOT in the `agent_definitions_backup.sql` project file (that backup is older). The user explicitly stated they gave me the SQL for it a few turns before context was lost. **First step in next chat: ask the user to provide the asset-deployer SQL again, or query the live DB to retrieve it.**

**`deploy_image_asset` action** — EXISTS in Go code. Does:
1. `findStorageURI()` — looks for S3 URI in collected_data via config-defined paths
2. `findDomain()` — gets domain from collected_data
3. `s3Client.DownloadOptimizeAndPrepare()` — downloads from S3, optimizes image
4. `sendGitCommitRequest()` — commits optimized image to git repo
5. Returns: `image_url`, `output_path`, `size_bytes`, git commit result

It's designed to run mid-workflow where a previous step has already loaded the S3 URI into collected_data.

**`store_asset` action** — EXISTS. Used by pageflow-builder to store generated images. Writes to `assets` table with `site_id`, `purpose`, `url` (S3 URI), `asset_type`, etc.

### What the asset-deployer workflow should look like

The agent needs to be self-sufficient. Suggested workflow:

```
1. load_undeployed_assets  → query assets table WHERE site_id = X 
                              AND deploy_status != 'deployed' (or similar)
                              Returns array of {s3_uri, purpose, asset_id}

2. loop over assets:
     deploy_image_asset    → downloads from S3, optimizes, commits to git
     mark_asset_deployed   → updates assets table row

3. complete
```

The `load_undeployed_assets` action may or may not already exist — **this was debated in the previous chat** (the whole asset-deploy-agent vs asset-deployer incident). We determined we should NOT create a new `asset-deploy-agent` but should reuse `asset-deployer`. Whether the load action was actually created or just discussed needs verification.

### Key design point from the user
> "The assets have already been deployed to S3 but are not on the site. The agent or an agent should be able to dig out their url and deploy them as needed, I don't want to be sending local urls in messages as the initial messages are a proxy for messages that might be sent by agents from the front end in future."

So the initial message should be just:
```json
{"domain": "finetuning.uk"}
```

And the agent's workflow handles everything else.

---

## Next steps (in order)

### 1. Retrieve asset-deployer agent definition
Either ask user for the SQL again or query live DB:
```sql
SELECT type, default_config FROM agent_definitions WHERE type = 'asset-deployer';
```

### 2. Check what's in the assets table
```sql
SELECT id, site_id, purpose, url, asset_type, deploy_status, created_at
FROM assets
WHERE site_id = (SELECT id FROM sites WHERE domain = 'finetuning.uk')
ORDER BY created_at;
```

This shows what the load action needs to surface.

### 3. Verify the load action exists or create it
```bash
grep -rn "load_undeployed\|LoadUndeployed" platform/orchestration/actions/*.go
grep -n "undeployed" platform/orchestration/actions/registry.go
```

If it doesn't exist, create it. If it does, verify it returns what `deploy_image_asset` expects.

### 4. Update the asset-deployer workflow
Make it self-sufficient: `load → loop(deploy) → complete`

### 5. Update test_handlers.sh
The asset test should send just `domain`, not `s3_uri`:
```bash
test_asset_deployer() {
    local DOMAIN="$1"
    # ... send message with only domain in input_data
}
```

### 6. Test webdesign-agent (second handler in the script)
This one should already work — it loads its own context via `load_site_for_design` action and only needs `domain` in input_data. Verify and test.

### 7. Continue with remaining handlers
Each handler in the test script needs the same treatment: verify it's self-sufficient, test independently.

---

## Important context for next chat

- The `agent_definitions_backup.sql` in the project becomes old when agents are added or altered in subsequent chat and won't include these recently created agents. Don't rely on it to check what exists — check the chat history for more recent changes too and ask for updated context if you need it.
- The kubernetes namespace is `ai-persona-system` (e.g. `kubectl -n ai-persona-system get pods`)
- Kafka namespace: `kubectl -n kafka get pods`
- Deployment is to GitHub → GitHub Actions → Backblaze S3
- Every agent is an orchestrator
- Keep workflows simple, complexity in Go action code
- Don't create subworkflows — spawn sub-agents instead
- Check existing code before creating new (the pre-flight checklist in 001_development_guide.md)