# 130 — aiservice HTTP clients have no timeout; one hung call wedges an agent's lane until a pod roll

> Filed as 129, renumbered to 130 within the hour: a concurrent session
> committed its own 129 (spawned-child handshake case) first. Committed wins;
> numbering is never reassigned.

**Filed:** 2026-07-28, by the gauntlet_dead_cta thread (this is `bugs_open/083`
gauntlet-engine-503's candidate 2, promoted to its own case on fleet grounds as
that file's §10 directed).
**Status:** OPEN — fix built, tests pass, submitted to the council 2026-07-28
~11:35 BST, `SUBMISSION_CORR = 1b7d802d-b416-4bcf-9b2f-0445e918ecda` (verdict
PENDING at commit time — read `decided_by` before believing any later trailer).
INERT until an image roll (chassis fleet AND island tools-api separately).

## Symptom

An agent stops consuming its Kafka lane and never recovers. Nothing is logged,
because nothing has failed — an LLM HTTP call is simply still waiting. The lane
stays frozen until the pod restarts. From outside it is indistinguishable from
the "wedged head orchestration" class (dispatch_queue_serialisation landmine).

## Mechanism

Two of the three `aiservice` providers construct their HTTP client with no
timeout of any kind:

- `platform/aiservice/anthropic.go:63` — `httpClient: &http.Client{}`
- `platform/aiservice/gemini.go:180` — `httpClient: &http.Client{}`

These are the **only two** naked `&http.Client{}` in the platform (grep over
`platform/ internal/ pkg/ cmd/`, 2026-07-28); every other client carries a
`Timeout`, including the third provider in the same package
(`ollama.go:54`, `Timeout: 120 * time.Second`). So this is an omission, not a
design choice.

Requests are context-aware (`http.NewRequestWithContext`, `anthropic.go:130`,
`gemini.go:355`, `gemini.go:507`) — so a call is bounded **only if the caller's
context carries a deadline.** The chassis's does not, at any link:

1. `cmd/agent-chassis/main.go:207` → `agentbase.New(ctx, …)`
2. `platform/agentbase/agent.go:190` — `agentCtx, cancel := context.WithCancel(ctx)` — cancel-only, pod-lifetime
3. `platform/agentbase/agent.go:1125` — `a.processor.ProcessMessage(a.ctx, msg)`
4. `platform/messaging/processor.go:1820` — `ExecuteWorkflow(ctx, …)`
5. `platform/orchestration/coordinator.go:898` — `executeStep(ctx, …)`
6. `platform/orchestration/actions/ai_actions.go:399` — `aiClient.GenerateText(ctx, …)`

And the consume loop is synchronous — `processMessage` runs an orchestration's
consecutive local steps inline before the loop can fetch again (the CS-2 design
comment at `agent.go:435-441` documents this). **One TCP connection that stalls
after connect therefore holds the agent's entire lane, forever.**

## Evidence it has already fired (not latent)

`llm_call_log`, queried 2026-07-28:

- **The incident.** 2026-04-28 15:05 UTC, `content-quality-auditor`:
  `latency_ms = 1,805,242` (30 min 5 s), `success = f`, `output_tokens` NULL,
  error `Post "https://api.anthropic.com/v1/messages": context canceled`.
  Nothing timed it out — **the call ended only because the agent's pod-lifetime
  context was cancelled.** The only thing that freed that lane was the pod dying.
- **The distribution.** 43,890 anthropic calls since 2026-03-25. Slowest
  *successful* call ever: `361,885 ms` (~6 min, a 32,000-output-token
  generation). Four successes exceeded 300 s; none exceeded 362 s. The only
  other two failures over 300 s were `max_tokens` truncations on ~355–360 s
  calls — completed transport-level, not hangs.
- **The positive control.** Ollama's April rows fail at exactly `600,001 ms` —
  its then-600 s client `Timeout` firing, logged diagnosably with the standard
  `Client.Timeout exceeded` shape. That is what the missing behaviour looks
  like when present.

Island exposure (tools-api gauntlet handlers) is milder: gin's request context
cancels on client disconnect, so a hang there is bounded by the visitor's / CF's
patience, not unbounded. The fleet chassis is the unbounded case.

## Fix candidates, ordered by what closes the door

1. **CHOSEN — a ceiling `Timeout` in the constructors themselves:**
   `Timeout: 600 * time.Second` at both construction sites. Makes the unbounded
   state unrepresentable: no caller can forget it, no config can disable it.
   600 s = 1.66× the slowest success in 44k calls over 4 months; it cuts nothing
   legitimate on record while converting "hangs until a pod roll" into "fails in
   10 minutes with a self-identifying log row". Per-call bounds remain the
   caller's ctx's job — this is a hang-killer, not a tuning knob.
   [ASSUMED] no future call legitimately exceeds 600 s; if output caps ever go
   above 32k tokens, revisit the ceiling alongside.
2. Transport-level timeouts (`ResponseHeaderTimeout`, dial, TLS) — more precise
   discrimination between connect-hang and slow-body, but more knobs than the
   evidence (one hang in 4 months) justifies.
3. Deadlines at every call site — 17 files reference `aiservice`; "every caller
   must remember X" is a defect in costume.

## How to verify (after a roll)

Go change ⇒ **inert until an image is rebuilt and rolled** — chassis fleet and
island `tools-api` separately (`tools-api` builds its own image; see
`infra/island/RUNBOOK_island.md`).

- Struct-literal fields leave no `strings`-greppable marker in the binary — do
  not invent a vacuous one (cta_link_integrity landmine). Verify by **image
  provenance** (image built from a commit ≥ the fix) plus the **behavioural
  signature**: any future transport hang now logs
  `… (Client.Timeout exceeded while awaiting headers)` at `latency_ms ≈ 600,0xx`
  — the exact shape ollama's April rows already demonstrate.
- Negative check that the ceiling never cuts real work: any *success* row with
  `latency_ms > 500,000` post-roll would mean the [ASSUMED] above is failing.

## Cross-references

- `bugs_open/083_HANDOFF_2026-07-26_gauntlet_engine_503_discards_the_error.md`
  §10 — where this was identified as candidate 2 and ruled "argue on fleet
  grounds with its own council round". That bug stays independently open (its
  armed log has caught nothing as of 2026-07-28 11:24 BST).
- `bugs_open/040` — Kafka dial timeouts: different transport, same family of
  "no bound on a network wait".
- dispatch_queue_serialisation workstream — the "wedged head orchestration
  freezes the lane until a pod roll" landmine this mechanism feeds.
