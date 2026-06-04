# Development runbook — idea.uk + SFI26 tool

What we're building, in what order, with explicit acceptance criteria so we know
when each step is done. Designed to be checked off as we go. Each task says
*what's done* (the output), *how to test it*, and *what unblocks*.

References:
- Architecture: `idea_uk_architecture_and_deployment.md`
- Method: `idea_uk_method_v0.md`
- Liability + T&Cs: `LIABILITY_AND_TERMS.md`
- Method test runs: `idea_uk_testrun_v0.md`, `idea_uk_testrun_v2.md`
- Plan: `PLAN_idea_uk.md`
- Ongoing journal: `running_notes.md`

---

## Phase A — Pre-launch quality and liability hardening for idea.uk

Goal: bring the existing standalone Go service from "works in test" to "fit
to take real money from real people without exposing us."

### A1 — Engine upgrade: latest LLM features

**Output:** engine.go uses extended thinking on the cut step, the newer
`web_search_20260209` tool with code execution on verify, prompt caching on
the static system prompts (capability menu + system base), and Anthropic file
attachments where useful.

**Acceptance:**
- `go run . internal "agritec.uk" "..." "..."` succeeds end-to-end.
- Stderr log lines confirm: extended thinking on cut, new web_search tool on
  verify, cache reads on subsequent calls within a single run.
- Output report shows visibly higher verification depth (more sourced
  numbers, real computation on at least one quantitative claim).

**Unblocks:** A2, A3.

### A2 — Free audience-check taster endpoint

**Output:** new `/audience-check` endpoint on the service (public, no auth, no
billing). Accepts `business` + `audience` form fields. Returns the rendered
HTML of Step 1 output (carried audience + reasoning + 3 alternatives) in
~10 seconds. ~£0.02 per call.

**Acceptance:**
- Calling it returns valid HTML within 15s 95% of the time.
- The output is the audience-check shape we agreed (reframe + reasoning + 3
  alternatives + CTA to buy the full report).
- Rate-limited per IP (e.g. 3 tasters per hour) to prevent abuse.
- Cost per call ≤ £0.05.

**Unblocks:** page rewrite (A5) needs this endpoint.

### A3 — Streaming progress page after Stripe redirect ("real door")

**Output:** post-payment success page polls a new `/status/{order_id}`
endpoint and renders progress in real time. Engine writes status updates to
the store at each step. Report renders in-browser when done; copy also
emailed.

**Acceptance:**
- After completing a test Stripe payment, browser shows "preparing… generating
  candidates… cutting… verifying claim 1 of N… complete" with each step
  updating live.
- Report renders inline in the browser on completion.
- Customer also receives the report by email.
- If the engine fails mid-run, the page shows a clear "we hit a problem,
  you've been refunded" state.

**Unblocks:** decent launch UX (the 72h email model is honest but feels
flat compared to real-time).

### A4 — Programmatic refund endpoint

**Output:** new `/refund` endpoint on the service, operator-gated (same
X-Internal-Key as confirm/decline). Body: `{order_id, reason}`. Calls Stripe's
`POST /v1/refunds` with the payment_intent we already store; sets order
status to `refunded`; emails the customer to confirm.

**Acceptance:**
- Refunding a test payment via this endpoint refunds it in the Stripe test
  dashboard, customer gets confirmation email, order shows `refunded`.
- Unauthenticated calls return 401.
- Already-refunded orders return 409.

**Unblocks:** smooth ops on day 1; saves time vs the Stripe dashboard at
volume.

### A5 — Page rewrite: simpler language + integrated taster + disclaimers

**Output:** `idea_uk_fakedoor.html` rewritten with plain English copy, the
audience-check taster integrated as a working widget (calls `/audience-check`),
the T&Cs and refund policy clearly linked at checkout and at the footer, and a
short "what this is and isn't" plain-English honesty paragraph in the body.
Editorial design retained.

**Acceptance:**
- A non-technical reader can answer in 30 seconds: what is this, what do I
  get, what does it cost, what's free.
- Taster widget on the page works end-to-end (filled fields → audience-check
  result rendered → CTA to pay).
- T&Cs link visible at the page footer and at the order checkbox.
- Reading age (a Hemingway-app check, say grade 8 or below) for the hero +
  value-prop sections.

**Unblocks:** launch.

### A6 — Terms of service and refund policy pages

**Output:** `/terms` and `/refund-policy` pages on idea.uk. T&Cs based on the
draft in `LIABILITY_AND_TERMS.md` §3, *after solicitor review*. Refund policy
matches what the page promises ("14 days, no quibble").

**Acceptance:**
- Pages live, linked from footer + checkout.
- Solicitor's revisions incorporated.
- Checkout flow has a "I accept the terms" checkbox; orders without it are
  rejected.

**Unblocks:** legally cleaner launch. Also reduces real-world dispute risk.

### A7 — Insurance: PII in force

**Output:** confirmed PII policy covering "AI-assisted analysis services,
operator-reviewed," £100k+ cover, ~£200–500/year. Insurer told honestly what
we do. Policy documents on file.

**Acceptance:**
- Active policy, in date, with the cover summary on file.
- Insurer aware that scope will widen to SFI advice when that product
  launches (so they can quote on the extension).

**Unblocks:** launching to real customers responsibly. The £29 idea.uk
product is arguably borderline-OK without PII, but it's so cheap not to be
uninsured that there's no good reason to skip it.

### A8 — Stripe live mode

**Output:** Stripe account verified, live keys in service env, live webhook
endpoint pointed at production URL, signing secret rotated to live. One real
£0.30 test transaction (your own card → your own account) end-to-end to prove
live mode works.

**Acceptance:**
- Live mode test: real card, real payment, real webhook, real fulfilment,
  real refund via the new endpoint.
- AUTO_DELIVER stays false.

**Unblocks:** launch.

### Phase A acceptance summary

Phase A is done when all of A1–A8 are done. At that point idea.uk can take
real orders responsibly. Estimated effort: roughly 2–4 days of focused work
plus solicitor-review wait (~1 week elapsed for A6) and insurance turnaround
(~1–2 days for A7).

---

## Phase B — Deploy idea.uk

Goal: idea.uk live, taking orders, operator-reviewing each report manually,
producing real evidence of demand.

### B1 — Deploy the service

**Output:** container running on a small box (Hetzner / Fly / Railway / your
ai-persona-system K8s pod), behind HTTPS, with persistent storage for the
JSON store. Health endpoint reachable.

**Acceptance:**
- `https://idea.uk/health` returns OK from outside.
- Restart-safe: container restart preserves state.
- Stripe webhook endpoint reachable from Stripe's IPs.

### B2 — Deploy the page

**Output:** rewritten page (from A5) deployed through the normal git → GitHub
Actions → Backblaze pipeline. Page can call the service (same origin via path
proxy, or CORS configured).

**Acceptance:**
- Page loads cleanly at https://idea.uk.
- Taster widget works against live service.
- Order form posts to live service and creates a real order.

### B3 — First operator-reviewed orders

**Output:** the first 10 paid orders go through with operator review on
every one. Each review takes ~15–30 minutes (read the draft, check sources,
edit anything wrong, send). Refund anything we can't deliver well.

**Acceptance:**
- 10 orders delivered.
- Pattern data captured: which reports were good first time, which needed
  edits, which were refunded and why.
- At least one of: the method is sharpened based on real customer reactions;
  or the report template improved; or the prompt set adjusted.

### B4 — Throughput decision

**Output:** based on B3 data, decide:
- Stay operator-review (default) and keep going.
- Switch AUTO_DELIVER on (only if no errors in B3).
- Hire a part-time reviewer.
- Pause new orders and improve the engine.

**Acceptance:** decision documented in running notes.

### Phase B acceptance summary

Phase B is done when 10 real reports have been delivered and reviewed,
patterns documented, and B4 decision made.

---

## Phase C — SFI26 single-farm assessment tool

Goal: build the first vertical tool. Higher stakes, higher quality bar, real
liability work needed up front.

### C1 — Method specification for the SFI tool

**Output:** `SFI26_METHOD.md` — same shape as `idea_uk_method_v0.md` but
domain-specific. Inputs (farm details: SBI, parcels, current agreements,
tenancy type, etc.). Steps (eligibility check, action selection within cap,
window recommendation, conflict checks, citations). Output template
(structured report with the disclaimer at the top per `LIABILITY_AND_TERMS.md`
§5).

**Acceptance:**
- A human can follow the method by hand and produce a useful report.
- Method explicitly references which Defra/RPA documents each step relies on.
- Output template includes the "Read this first" disclaimer box and per-claim
  date stamps.

### C2 — Corpus curation

**Output:** scraped + versioned set of Defra/RPA SFI26 pages and PDFs. Each
document stored with a fetch date and version hash. Re-scraped weekly. Stored
in a way the engine can retrieve and cite.

**Acceptance:**
- ≥30 source documents in the corpus.
- Re-scrape job runs weekly without intervention.
- Engine can retrieve a specific document by URL and cite it accurately.

### C3 — SFI engine (variant of idea.uk engine)

**Output:** new Go binary or new endpoint on the service, using the same
engine *framework* (LLM calls, web search, structured outputs) but the SFI
method. Likely a new file `sfi_engine.go` alongside `engine.go`, sharing
helpers.

**Acceptance:**
- `go run . sfi <farm-details-file>` produces a draft SFI report.
- Every numeric claim is cited.
- Disclaimer block appears at the top.
- Report identifies any close eligibility calls and flags them for human
  review.

### C4 — Operational shape

**Output:**
- Static landing page on agritec.uk (or wherever) describing the product.
- Intake form (farm details).
- £49–99 price point (decide based on competitor benchmarking).
- Operator-review on indefinitely until ≥100 reports without material error.
- One named UK agricultural advisor on retainer for sanity checks.
- PII scope confirmed in force.

**Acceptance:**
- Page live, intake works, payment flow works, operator-review process
  documented for the SFI product specifically.
- First report produced and delivered without error.

### C5 — First 10 SFI reports

**Output:** as Phase B but for SFI. Operator-reviewed, customer follow-up
("did you apply? did the application go as expected?"), pattern data.

**Acceptance:**
- 10 reports delivered.
- No material errors that caused customer loss.
- At least 5 customers report they used the report in their actual
  application process and it helped.

### Phase C acceptance summary

C is done when 10 SFI reports have been delivered with operator review and no
material errors. At that point we decide whether to widen scope or layer
subscription on top.

---

## Phase D — Chassis-native version (deferred)

Goal: express the idea-generation method as an internal chassis agent +
workflow reusing existing actions, for running the method on our own domains
on demand or on a schedule.

**Not started yet.** Per the architecture doc, this needs a schema pass first.

### D1 — Schema and action contract pass

**Output:** brief document mapping each method step onto specific existing
actions (`execute_llm_prompt`, `web_search`, `request_human_input`, etc.) with
the exact action contracts (inputs/outputs) we need to honour. Confirm what
new actions, if any, need writing.

**Acceptance:**
- Mapping is complete and consistent with the action registry.
- Any new actions identified are listed with their proposed contracts.

### D2 — Agent definition + workflow SQL

**Output:** new agent definition `idea-orchestrator` + a workflow that
sequences the steps using the actions from D1. Spawns sub-agents per
convention (every agent is an orchestrator). Children reply on the parent's
responses topic.

**Acceptance:**
- Workflow runs end-to-end via the chassis orchestrator.
- Produces the same shape of output as the standalone Go service.
- Logs are clean.

### D3 — Wire into site_specs feasibility model

**Output:** a domain's site spec can declare it needs the idea-generation
capability; while blocked, `feasibility-recheck` waits; once the agent is
deployed, the page builds and is linked to the central service.

**Acceptance:**
- A new domain in the build pipeline can include "ideation" as a feature in
  its site_plan, status flips from blocked to planned to built once the agent
  exists.

### Phase D priority

Low for now. Phase A → B → C is the demand-validation path; D is the
"productionise inside the platform" path that pays off only once the standalone
versions have proven demand.

---

## Right-now-this-week shortlist

Concrete tasks ordered for the next session:

1. **A1** — engine upgrade to latest LLM features.
2. **A2** — audience-check taster endpoint.
3. **A5** — page rewrite (simpler English + integrated taster + disclaimers).
4. **A6** kickoff — send draft T&Cs from `LIABILITY_AND_TERMS.md` §3 to a
   solicitor.
5. **A7** kickoff — get a PII quote.

Numbers 4 and 5 run in the background while we build 1–3. Phase B starts after
A1–A8 are complete.
