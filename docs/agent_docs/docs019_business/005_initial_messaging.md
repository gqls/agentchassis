For the `vet-batch-processor`, the initial message is simpler — it doesn't need much input since it pulls work from the `collection_tasks` queue. Here's the script:Also useful — a script to test a single practice verification directly:**Two scripts:**

`start_vet_batch.sh` — runs the batch processor
```bash
./start_vet_batch.sh          # defaults: batch_size=5, task_type=initial_verification
./start_vet_batch.sh 10       # process 10 tasks
./start_vet_batch.sh 3 price_refresh veterinary   # different task type
```

`start_vet_single.sh` — tests a single practice (useful for debugging)
```bash
# First get a business_id:
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \
  "SELECT id, name, postcode FROM business_intel.businesses WHERE verification_status = 'seed_import' LIMIT 5"

# Then test one:
./start_vet_single.sh <business_id>
```

---

**Before running**, you'll need:

1. **Schema + seed data applied** (sounds like you've done this)

2. **Agent definitions in templates_db:**
   ```bash
   kubectl -n ai-persona-system exec -i postgres-templates-0 -- \
     psql -U templates_user -d templates_db < 003_agent_definitions_business_intel.sql
   ```

3. **Go actions registered** — add to `GlobalActionRegistry` and `LocalActions`:
   ```go
   "load_business_record":        LoadBusinessRecordAction,
   "store_business_verification": StoreBusinessVerificationAction,
   "load_business_batch":         LoadBusinessBatchAction,
   ```

4. **Rebuild + deploy agent-chassis** with the new actions

5. **(Optional) Throttle adapters** — if you want the 5s delays between requests, add the throttle code and set `REQUEST_THROTTLE_MS=5000` on the webscrape and web-search adapter deployments
