# Append to FOCUS_finetuning_flywheel_and_service(13).md, section 15 (Changelog)
#
# Paste these entries at the TOP of the changelog (most recent first per the
# existing convention).

- 2026-05-12 (thunder-adapter Phase 3.4 delivered, 3.5 pending) —
  decommission_instance action handler complete. Lookup by provisioning_id
  (DB UUID) or thunder_identifier (numeric). Atomic state transition to
  'decommissioning' is the idempotency anchor; rows already
  decommissioned/failed/reaped return cached cost without re-calling APIs.
  Cost computed from running_since × hourly_rate_usd. Both API delete and
  Secret delete are 404-tolerant. Status doc captured at
  STATUS_thunder_adapter_2026-05-12.md including the verification sequence
  for the first real provision-then-decommission cycle. Next: 3.5 reaper
  scheduled task (15-min sweep finding instances older than
  max_uptime_hours, dispatching decommission for each), then deploy +
  manual verify + replace gpu-provisioner stub (migration 022) with real
  Kafka dispatch to thunder-adapter.

- 2026-05-12 (thunder-adapter Phase 3.3 delivered) — provision_instance
  action handler complete. Pre-check via thunder_provision_check view →
  ed25519 keypair generated client-side → POST /v1/instances/create with
  public_key (private never leaves cluster) → k8s Secret persists keypair
  → WaitForRunning polls until RUNNING or terminal status → INSERT
  thunder_instances row with three-retry backoff (1s/3s/5s) → compensating
  cleanup (decommission instance + delete Secret) on any partial failure
  using fresh ctx so cleanup runs even on parent ctx timeout. Two small
  interfaces (thunderAPI, secretManager) defined in provision_action.go
  for test injectability — same interfaces reused by decommission.
  Provisioning ID is a pre-generated UUID (DB row PK) so SSH Secret name
  is deterministic before INSERT.

- 2026-05-12 (thunder-adapter Phase 3.0/3.1/3.2 delivered) — Three structural
  pieces in one session. 3.0: config URL was wrong
  (https://api.thundercompute.com/v1) — actual is
  https://api.thundercompute.com:8443/v1 per Thunder docs, including the
  non-standard port and /v1 prefix. 3.1: Thunder Compute API client
  (internal/adapters/thunder/api/) — Client with bearer auth, CreateInstance,
  ListInstances, GetInstance with 404-fallback-to-list, DeleteInstance
  idempotent, WaitForRunning polling helper with ctx-based deadline. APIError
  with IsAuth/IsNotFound/IsRateLimit/IsServer classifiers. Three TODOs
  flagged for verification on first real call (response wrapper shape, JSON
  field names, status casing). 3.2: ed25519 SSH keypair generation + k8s
  Secret CRUD (internal/adapters/thunder/ssh/). SecretManager wraps
  kubernetes.Interface for test injection; NewInClusterSecretManager for
  production. Plus rbac.yaml (Role + RoleBinding scoped to Secrets, bound
  to ai-persona-app SA — same SA used by other adapters). RBAC verbs limited
  to create/get/delete; no list/watch/update. Resource names not enumerated
  (no glob in k8s RBAC) — practical exposure limited by adapter code only
  creating Secrets with thunder-ssh- prefix.

- 2026-05-12 (thunder-adapter Phase 2 deployed, end-to-end verified) — After
  a long chain of deployment issues (image tag mismatch, private repo /
  missing imagePullSecrets, missing serviceAccountName, Dockerfile CMD vs
  deployment args trap, Strimzi auto-create-off missing topic), Phase 2
  skeleton came up clean as v1.0.1010. kcat round-trip verified: request
  consumed, parsed, not_implemented response produced with correct
  in_response_to_request_id, sender_agent_type, correlation_id, and error
  body shape. All Tier 1 / Tier 2 validator headers populated. Three docs
  updated from the deployment lessons: section 10 added to
  016_debugging_guide_v2.md (adapter deployment failure modes), FOCUS_adapter_design.md
  Deployment essentials rewritten with full annotated manifest + Kafka
  topic gotcha + RBAC pattern + makefile integration notes. The thunder-adapter
  is the first adapter in this codebase that needs k8s API access (for SSH
  Secret CRUD in Phase 3.2+), so it's also the first to need scoped RBAC
  beyond the default SA permissions.

# End of changelog additions
