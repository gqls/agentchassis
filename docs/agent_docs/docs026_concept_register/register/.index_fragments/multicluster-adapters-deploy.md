| MCL-001 | Multi-cluster dispatch contract (dispatch_agent + remote-job-spawner) | partial | Parent action publishes to Kafka; remote spawner creates the K8s Job on the target cluster | multicluster.md |
| MCL-002 | Adjacent-cluster Phase 4a rollout: va001 second cluster | deployed | Second Rackspace Spot cluster shares primary's Kafka/Postgres for trusted dispatch | multicluster.md |
| MCL-003 | Cluster-filter gap in remote-job-spawner (Gap A) | partial | target_cluster filter exists but logs at Debug not Info on the skip path | multicluster.md |
| MCL-004 | Dispatch confirmation observability gap (Gap B / agent_dispatch_log) | aspirational | Fire-and-forget dispatch has no near-real-time failure signal; table proposed, not built | multicluster.md |
| MCL-005 | Same-cluster loopback test requirement (Gap C) | aspirational | Prove dispatch round-trip locally before adding a real second cluster | multicluster.md |
| MCL-006 | Cross-cluster Kafka external listener (nodeport→loadbalancer) | aspirational | Third Strimzi listener lets a second cluster reach primary Kafka, no MirrorMaker | multicluster.md |
| MCL-007 | Cross-cluster KafkaUser + secret replication pattern | aspirational | Manual kubectl-copied KafkaUser secret authenticates a remote cluster's agents | multicluster.md |
| MCL-008 | Kafka cluster-wide authorization gap (ACLs decorative) | partial | No spec.kafka.authorization block means declared ACLs are unenforced | multicluster.md |
| MCL-009 | Cross-cluster Postgres reachability strategy (Option C) | aspirational | Local PgBouncer per remote cluster tunnels back to primary Postgres | multicluster.md |
| MCL-010 | RTT-based agent-type viability classification | partial | Classifies which agent types tolerate cross-cluster DB latency | multicluster.md |
| MCL-011 | Cross-cloud cluster expansion (Phase 4: AWS EKS / GCP GKE) | aspirational | Extends adjacent-cluster pattern to a genuinely remote cloud provider | multicluster.md |
| MCL-012 | Multi-cluster environment re-discovery handoff practice | unknown | Treat prior FOCUS docs' IPs/names as illustrative; re-derive live facts each session | multicluster.md |
| MCL-013 | Multi-cluster scaling tiers (10K/100K/1M agents) | aspirational | Maps each agent-count tier to its bottleneck and the one architectural fix | multicluster.md |
| MCL-014 | Shared topic pools (replace per-agent topics) | aspirational | Fixed partitioned pool topics route by agent-ID header instead of per-agent topics | multicluster.md |
| MCL-015 | Worker pool architecture (replace per-agent Jobs) | aspirational | Long-running pods run many agent workflows as goroutines instead of per-agent Jobs | multicluster.md |
| ADP-001 | Adapter/response message envelope contract (normative) | deployed | in_response_to_request_id + typed bool headers + ProduceWithValidation, or replies silently vanish | adapters.md |
| ADP-002 | Adapter response-header tier taxonomy and validator-coverage gap | partial | Tier-2 routing fields aren't validator-enforced; unfiled tracking issue | adapters.md |
| ADP-003 | Adapter design pattern / guide (adapter vs agent vs inline) | deployed | Canonical decision rule and Kafka/HTTP microservice structure for adapters | adapters.md |
| ADP-004 | Adapter deployment essentials & troubleshooting checklist | deployed | command vs args, KafkaTopic CRDs, RBAC globs — real thunder-adapter deploy lessons | adapters.md |
| ADP-005 | Tier-4 browser-runner adapter (headless Chromium over Kafka) | deployed | Playwright/Chromium adapter runs page/selector/console checks, P0 deployed and smoke-passed | adapters.md |
| ADP-006 | git-adapter as sole write credential holder | deployed | Fix-implementer never holds a GitHub write token; git-adapter does all writes | adapters.md |
| ADP-007 | git-adapter new actions (create_branch, create_pull_request) | deployed | Idempotent branch creation and PR-as-human-review-terminal actions | adapters.md |
| ADP-008 | Agent-to-adapter capability maturation path (fast/slow/job lanes) | aspirational | Capabilities prove out as spawned agents, then promote to warm adapters | adapters.md |
| ADP-009 | Analyser adapter (build + migration path) | deployed | Polyglot code-parsing adapter, deployed to production 2026-06-12 | adapters.md |
| ADP-010 | GitHub read-token scoping / least-privilege adapter secrets | deployed | Read-only repo-scoped PAT injected only for isRepoCloningAgent types | adapters.md |
| ADP-011 | thunder-adapter — GPU provisioning adapter | deployed | Holds Thunder/B2 creds, provisions ephemeral GPU VMs, verified end-to-end 2026-05-22 | adapters.md |
| ADP-012 | Thunder adapter schema and provisioning gates | deployed | thunder_instances/thunder_config/thunder_provision_check enforce cost + concurrency caps | adapters.md |
| ADP-013 | Thunder consecutive-unreachable probe streak | deployed | Counter-based durability so one transient SSH blip doesn't kill a training run | adapters.md |
| ADP-014 | Thunder Compute API specifics (field/casing/template traps) | deployed | snake_case create vs camelCase status, real template names, ubuntu login user | adapters.md |
| ADP-015 | Firecrawl scraping adapter and actions | deployed | Kafka adapter exposing scrape/crawl/extract; v2 owns screenshot/image S3 copies | adapters.md |
| ADP-016 | vmhost/service-deployer adapter — persistent-VM provisioning | aspirational | Proposed adapter to automate what idea.uk and traffic_probe both did by hand | adapters.md |
| STG-001 | Storage: per-call S3 client construction is canonical | deployed | params.StorageClient is unreliable (nil at startup); construct per action instead | storage-architecture.md |
| STG-002 | Hostile-VM threat model for the training data plane | deployed | GPU box holds no B2 key/DB access, only time-limited presigned URLs | storage-architecture.md |
| STG-003 | Presigned-URL data plane / storage boundary | deployed | Adapter mints URLs; only URLs cross Kafka, bytes go direct VM↔B2 over HTTPS | storage-architecture.md |
| STG-004 | Storage credential architecture decision (no storage-adapter) | partial | Rejected a storage-adapter service since multi-MB blobs would wreck Kafka brokers | storage-architecture.md |
| STG-005 | asset-deployer (S3 → optimize-by-purpose → git) | deployed | Downloads S3 asset, optimizes by purpose (logo/hero), commits to git | storage-architecture.md |
| STG-006 | Checkpoint & final-adapter upload to B2 | partial | Save-index-keyed checkpoint uploads with a hard-gate final-adapter upload | storage-architecture.md |
| STG-007 | JSON store scaling evolution (whole-file → daily JSONL) | deployed | site-engine's store evolved from write-cliff whole-file to bounded daily JSONL | storage-architecture.md |
| STG-008 | Persistence design: tiered one-way data flow (box → B2 → chassis) | partial | Exposed box writes B2 dead-drop; chassis pulls and ingests, never the reverse | storage-architecture.md |
| STG-009 | Result storage split (DB paper-trail + S3 artefacts) | deployed | Postgres holds the record of what happened; S3 holds the actual product | storage-architecture.md |
| DGH-001 | Commit-is-deploy: git → Actions → Backblaze B2 (+ chassis image-tag deploys) | deployed | Standing deploy mechanism: commit triggers B2 sync; chassis code ships via image tag | deployment-github.md |
| DGH-002 | Git-adapter non-fast-forward commit race (shared sites repo) | aspirational | force:false + no retry on updateRef can silently lose a concurrent multi-site commit | deployment-github.md |
| DGH-003 | sites.github_repo deploy-target selector / resolveGitRepoName patch | partial | Dormant deploy-target column; 3-touch patch to actually wire it up | deployment-github.md |
| DGH-004 | Site manifest + external-edit desynchronisation detection | aspirational | manifest.json + git webhook would flag human edits and halt agent overwrites | deployment-github.md |
| DGH-005 | Chassis build/deploy practice (local Makefile builds) | deployed | Images build from local tree, decoupled from commits; verify against the running pod | deployment-github.md |
| SNAP-001 | Site snapshots: point-in-time capture and revert | deployed | JSONB full-site-state snapshot/revert functions, migration 085, used in production | site-snapshots-and-revert.md |
| SNAP-002 | component_versions population and change_source provenance | deployed | Three best-effort writers populate version history; lesson on silent best-effort | site-snapshots-and-revert.md |
| SNAP-003 | Milestone-tagged site-spec history with inline git-snapshot function | superseded | Original unbounded milestone history replaced by pruned specs + snapshot-agent | site-snapshots-and-revert.md |
| SNAP-004 | Snapshots and revert for agent definitions (snapshot_agent/revert_agent) | deployed | Convention to snapshot before patching default_config; audit trail table | site-snapshots-and-revert.md |
| VMB-001 | VM-hosted backend sites — a new infrastructure class | partial | Persistent non-reaped internet-facing VM class outside k8s, generalised from idea.uk | vm-backend-sites.md |
| VMB-002 | site-engine — API-only capture backend for the class | deployed | Stdlib-only Go binary forked from idea.uk, live for relojistas.com | vm-backend-sites.md |
| VMB-003 | Probe as Layer 4 build + thin Layer 5 VM deploy (D1-D4) | deployed | A probe is a normal chassis site whose only difference is the deploy target | vm-backend-sites.md |
| VMB-004 | vm-sites content repo and deploy-to-vm Action | deployed | Private repo + rsync-over-SSH Action mirrors the B2 action for VM targets | vm-backend-sites.md |
| VMB-005 | site-engine deploy Action and narrow-sudo privilege model | deployed | Push-to-deploy engine binary via a scoped sudo hook, no root key in CI | vm-backend-sites.md |
| VMB-006 | setup.sh — idempotent multi-vhost box provisioning | deployed | Adapted from idea.uk's script; idempotent nginx+TLS+hardening provisioner | vm-backend-sites.md |
| VMB-007 | Multi-domain single-binary hosting and domain onboarding/relocation | partial | One engine binary behind many domains; THANKS_PATH is engine-wide per box | vm-backend-sites.md |
| VMB-008 | Dedicated vs shared box policy and VM sizing | deployed | Unknown-traffic experiments get their own box; low-traffic domains share one | vm-backend-sites.md |
| VMB-009 | Pull-not-push off-cluster data return | partial | Cluster pulls over key-gated HTTPS on a schedule; box never holds credentials | vm-backend-sites.md |
| VMB-010 | requires-backend capability gate (Decision 5) | partial | Backend-requiring components gate on deploy-target capability, not site type | vm-backend-sites.md |
| VMB-011 | Cloudflare-proxied-in-front option | deployed | Orange-cloud CF in front of the VM origin; real-IP nginx config required | vm-backend-sites.md |
| VMB-012 | idea.uk VM deployment: Path A manual setup.sh, systemd binary | deployed | Abandoned Docker/S3 plan for a single embedded-page Go binary on a Hetzner VM | vm-backend-sites.md |
| VMB-013 | VM launch plan (idea.uk): dedicated hardened box | deployed | Dedicated VM chosen over reusing the shared OVH multi-domain reverse proxy | vm-backend-sites.md |
| VMB-014 | VM cutover: nginx front door with reserved tool paths | aspirational | Staging-in-place cutover plan for a chassis site sharing a domain with a live tool | vm-backend-sites.md |
