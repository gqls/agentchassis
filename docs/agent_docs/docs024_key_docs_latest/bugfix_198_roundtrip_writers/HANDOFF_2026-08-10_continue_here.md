# HANDOFF 2026-08-10 — continue here: the round-trip-writer inventory (the 012/198 class, fleet-wide)

Written at the owner's request at the end of the 2026-08-05→08-10 session that fixed
`bugs_open/198`. Everything DONE is committed and verified; this file exists so a fresh
chat can pick up the one substantial piece of owed work without re-deriving five days
of context. Re-run every liveness claim below before acting on it — this tree moves.

## Where things stand (all verified at the artefact, dates given)

- **bugs_open/198 (css-patch-agent clobbers stylesheet): BOTH fix candidates LIVE.**
  Candidate 1 (config, migration 318): applied+recorded 08-05, drift-guarded, proven
  on the real theme row. Candidate 3 (`commit_message_field` on GitCommitAction):
  binary verified 08-10 on chassis `v1.0.1277`, both replicas, pod-grep positive 2 /
  negative-control 0; config half re-probed intact same day. Council APPROVED r1
  (correlation `5249320e`), all 7 advisory objections dispositioned in the bug file.
  **The ONLY thing keeping 198 open: the witnessed end-to-end run** (next real
  contrast finding → promote → append → next audit stops re-filing) — that dispatch
  belongs to the vigilant_designer lane (session `137460cc…` was running it; they had
  B1+B2 offer work queue-jumped ahead by owner decision).
- **The A2 "blocker" is refuted** (owner caught it; census 08-09, 10/10 types match
  the gate). Spawned pods get full S3 iff `isStorageEnabledAgent(type)` (hardcoded
  12-type list, `spawn_actions.go:3039`) OR `agent_definitions.category` ∈
  {orchestrator, code-driven} (`spawn_actions.go:2556`). Unblock for the A2 critic:
  one Go line on the list, or seed the agent as `code-driven` if honest. Full
  evidence: `vigilant_designer_offer_analysis/CONTRIB_2026-08-09_spawned_pod_storage_gate_A2_unblocked.md`;
  correction sits in-place under the 08-08 warning; WRONG_CALLS row added.
- **DGH-007** register status = deployed (v1.0.1277); first live use still unobserved
  (css-patch-agent has had no dispatch since the fix).

## THE OWED WORK: inventory of round-trip writers (arch seat's ask, 198 council round)

The architecture seat's objection, verbatim: *"No inventory of other
agent_definitions workflows that round-trip a full artifact through an LLM and write
it straight to a DB column/file with no size or shrink guard — that inventory would
turn the class-cost claim from qualitative to quantified."* Nobody owns it yet. The
defect class is `bugs_open/012` + 198: a writer persisting an LLM's returned artefact
with no check against what it replaces.

**A floor measurement exists — do NOT mistake it for the answer.** This query (08-10)
returned exactly ONE row, css-patch-agent's own now-safe step:

```sql
WITH steps AS (
  SELECT ad.type, s.key AS step, s.value AS cfg
  FROM agent_definitions ad, jsonb_each(ad.default_config #> '{workflow,steps}') s
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL)
SELECT type, step, cfg->>'action',
       COALESCE(cfg#>>'{config,content_field}', left(cfg#>>'{config,query}',60))
FROM steps
WHERE (cfg->>'action'='git_commit' AND cfg#>>'{config,content_field}' LIKE '%.result.%')
   OR (cfg->>'action'='query_database' AND cfg#>>'{config,query}' ~* '^\s*(UPDATE|INSERT)'
       AND cfg#>>'{config,params}' LIKE '%.result.%');
```

It is blind three ways (a grep proves absence only for its spelling):
1. LLM outputs are not always under `.result.` — e.g. tool-recreation-handler saves
   from `validation_result.clean_html` (see migration 313's header).
2. Writers are not only `git_commit`/`query_database` — the save/deploy/store action
   family writes artefacts too (`save_page_sections`, `StoreGeneratedComponentAction`,
   `deploy_page`, multipage_actions' `content_field`, section_editor saves, …), some
   with real Go-side guards (claims floor on save_page_sections, F1 field-contract
   guard on component store), some without.
3. `files_field` (map-of-files) references are not matched by `content_field` LIKE.

**Method that actually answers it:**
1. Enumerate ALL steps with `action='execute_llm_prompt'` (and vision/generate
   variants) per active definition → their `output_field`s. That set is ground truth
   for "what is LLM output" — do not infer it from path spellings.
2. Enumerate ALL writer-shaped steps: `git_commit` (content_field/files_field/files),
   `query_database` with UPDATE/INSERT, and every registered save/deploy/store action
   (walk `registry.go` categories rather than guessing names).
3. Join per workflow: writer input references an LLM `output_field` (any path prefix,
   including `.response` unwraps — see datahelpers' auto-unwrap).
4. For each hit, read the PROMPT: whole-artefact contract ("return the complete X")
   vs fragment/patch contract. Whole-artefact + no guard = the 012/198 class. Note
   `max_tokens` vs realistic artefact size (the 198 file shows why that pairing is
   structural, not incidental).
5. Note the guard story per hit: truncation ALREADY fails loud on execute_llm_prompt
   paths without `tolerate_truncation` (`ai_actions.go:427`, `aiservice.IsTruncated`)
   and `output_format: json` parse-fails closed — so the live risk is **prompt
   non-compliance that still parses** (exactly how 198 happened), not max_tokens cuts.
6. Deliverable: a table in this dir (population, per-hit classification, guard
   present/absent), a `bugs_open/` filing if the population is non-trivial, and the
   198 bug file updated to point at it. The 090 diagnosis loop is NOT needed — this
   is a survey, not a root-cause claim — but the 2026-07-31 ruling applies the moment
   you assert a cross-cutting defect from it: file through 090 or state the
   substitute verification plainly.

## Smaller owed/candidate items (each independently pick-up-able)

- **remote-job-spawner has NO storage block** (`cmd/remote-job-spawner/main.go` env
  list, ~330-395) — the same two-spawner drift class `bugs_open/112` fixed for
  provider keys. A storage-enabled agent dispatched remotely fails like the 08-08
  pod. Candidate: file as a bug; fix mirrors the chassis spawner's gate (and consider
  extracting the gate to `platform/agentenv` where 112 put the provider keys).
- **`env_vars` `valueFrom` entries are silently dropped** by both spawners (structs
  carry Name/Value only) — html-developer / visual-designer / image-generator rows
  carry decorative secretKeyRef entries that do nothing; the types work via the
  allow-list. Candidate: either support valueFrom or strip the dead entries + a
  LANDMINES entry so nobody copies the pattern.
- **Storage credentials land as literal values in pod specs** (spawn_actions.go:2568+
  copies the spawner's env). `43c1801d6` stopped LOGGING them; pod-spec exposure
  remains. Owner-adjacent security question, not unilateral work.
- **First live use of `commit_message_field`** — when css-patch-agent next runs,
  check the vm-sites commit message reads `CSS fix: <category> (theme vN)` and not
  the fallback. One glance, closes the DGH-007 "unobserved" caveat.

## Ground rules that bit this session (so you don't re-learn them)

Pathspec commits, one per task, `Council-Submitted:` when committing pre-verdict
(`097_TRIGGER…` script; ~30min budget; find runs by payload not printed id). Platform
code → council gate before/alongside commit. Migration numbers: check
`sql_for_agents/` tail AND the untracked files other sessions leave (317 was
untracked when 318 was written). Config is live on apply; Go is inert till roll;
prove rolls at the pod with positive+negative greps. Read the owning lane's docs
before touching a bug (`scripts/who-owns.py`), and correct refuted claims IN PLACE
with a dated block naming what caught them.
